package handlers

import (
	"net/http"

	"sso/models"
	"sso/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// OAuth2Handler handles OAuth2-related requests
type OAuth2Handler struct {
	oauth2Service *services.OAuth2Service
}

// NewOAuth2Handler creates a new OAuth2Handler
func NewOAuth2Handler(oauth2Service *services.OAuth2Service) *OAuth2Handler {
	return &OAuth2Handler{
		oauth2Service: oauth2Service,
	}
}

// CreateClient creates a new OAuth2 client
// POST /api/v1/oauth2/clients
func (h *OAuth2Handler) CreateClient(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	uid, err := uuid.Parse(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req models.CreateOAuth2ClientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.oauth2Service.CreateClient(c.Request.Context(), &req, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// ListClients lists all OAuth2 clients owned by the current user
// GET /api/v1/oauth2/clients
func (h *OAuth2Handler) ListClients(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	uid, err := uuid.Parse(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	clients, err := h.oauth2Service.ListClientsByOwner(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"clients": clients})
}

// Authorize handles OAuth2 authorization endpoint
// GET /api/v1/oauth2/authorize
func (h *OAuth2Handler) Authorize(c *gin.Context) {
	var req models.AuthorizeRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if user is authenticated
	userID, exists := c.Get("userID")
	if !exists {
		// Redirect to login with return URL
		c.Redirect(http.StatusFound, "/login?return_to="+c.Request.URL.String())
		return
	}

	uid, err := uuid.Parse(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Generate authorization code
	response, err := h.oauth2Service.Authorize(c.Request.Context(), &req, uid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Redirect back to client with authorization code
	redirectURL := req.RedirectURI + "?code=" + response.Code
	if response.State != "" {
		redirectURL += "&state=" + response.State
	}

	c.Redirect(http.StatusFound, redirectURL)
}

// Token handles OAuth2 token endpoint
// POST /api/v1/oauth2/token
func (h *OAuth2Handler) Token(c *gin.Context) {
	var req models.TokenRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":             "invalid_request",
			"error_description": err.Error(),
		})
		return
	}

	response, err := h.oauth2Service.ExchangeToken(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":             "invalid_grant",
			"error_description": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// Introspect handles OAuth2 token introspection
// POST /api/v1/oauth2/introspect
func (h *OAuth2Handler) Introspect(c *gin.Context) {
	token := c.PostForm("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token required"})
		return
	}

	tokenInfo, err := h.oauth2Service.ValidateAccessToken(c.Request.Context(), token)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"active": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"active":    true,
		"client_id": tokenInfo.ClientID,
		"user_id":   tokenInfo.UserID,
		"scopes":    tokenInfo.Scopes,
		"exp":       tokenInfo.ExpiresAt.Unix(),
	})
}

// Revoke handles OAuth2 token revocation
// POST /api/v1/oauth2/revoke
func (h *OAuth2Handler) Revoke(c *gin.Context) {
	token := c.PostForm("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token required"})
		return
	}

	// Get token info
	tokenInfo, err := h.oauth2Service.ValidateAccessToken(c.Request.Context(), token)
	if err != nil {
		// Token doesn't exist or already invalid
		c.JSON(http.StatusOK, gin.H{"message": "Token revoked"})
		return
	}

	// Revoke the token
	if err := h.oauth2Service.RevokeToken(c.Request.Context(), tokenInfo.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Token revoked"})
}

// GetUserInfo returns user information (OpenID Connect UserInfo endpoint)
// GET /api/v1/oauth2/userinfo
func (h *OAuth2Handler) GetUserInfo(c *gin.Context) {
	// Extract token from Authorization header
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Remove "Bearer " prefix
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	// Validate token
	tokenInfo, err := h.oauth2Service.ValidateAccessToken(c.Request.Context(), token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Return user info based on scopes
	userInfo := gin.H{
		"sub": tokenInfo.UserID.String(),
	}

	// Add additional claims based on requested scopes
	hasScope := func(scope string) bool {
		for _, s := range tokenInfo.Scopes {
			if s == scope {
				return true
			}
		}
		return false
	}

	if hasScope("profile") {
		// Add profile information
		userInfo["name"] = "" // Fetch from user service
		userInfo["given_name"] = ""
		userInfo["family_name"] = ""
	}

	if hasScope("email") {
		// Add email information
		userInfo["email"] = "" // Fetch from user service
		userInfo["email_verified"] = false
	}

	c.JSON(http.StatusOK, userInfo)
}

// ConsentPage renders the OAuth2 consent page
// GET /api/v1/oauth2/consent
func (h *OAuth2Handler) ConsentPage(c *gin.Context) {
	// This would typically render an HTML page
	// For now, we'll return JSON with the consent details
	var req models.AuthorizeRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"client_id":    req.ClientID,
		"scopes":       req.Scope,
		"redirect_uri": req.RedirectURI,
		"state":        req.State,
	})
}
