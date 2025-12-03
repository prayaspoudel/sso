package handlers

import (
	"net/http"

	"sso/models"
	"sso/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// TwoFactorHandler handles 2FA-related requests
type TwoFactorHandler struct {
	twoFactorService *services.TwoFactorService
}

// NewTwoFactorHandler creates a new TwoFactorHandler
func NewTwoFactorHandler(twoFactorService *services.TwoFactorService) *TwoFactorHandler {
	return &TwoFactorHandler{
		twoFactorService: twoFactorService,
	}
}

// SetupTOTP initiates TOTP setup for a user
// POST /api/v1/auth/2fa/setup
func (h *TwoFactorHandler) SetupTOTP(c *gin.Context) {
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

	// Generate TOTP secret and backup codes
	response, err := h.twoFactorService.GenerateTOTPSecret(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// EnableTwoFactor enables 2FA after verification
// POST /api/v1/auth/2fa/enable
func (h *TwoFactorHandler) EnableTwoFactor(c *gin.Context) {
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

	var req models.Enable2FARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Enable 2FA
	if err := h.twoFactorService.EnableTwoFactor(c.Request.Context(), uid, req.Code); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "2FA enabled successfully"})
}

// DisableTwoFactor disables 2FA for a user
// POST /api/v1/auth/2fa/disable
func (h *TwoFactorHandler) DisableTwoFactor(c *gin.Context) {
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

	var req models.VerifyTOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify TOTP code before disabling
	valid, err := h.twoFactorService.VerifyTOTP(c.Request.Context(), uid, req.Code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid verification code"})
		return
	}

	// Disable 2FA
	if err := h.twoFactorService.DisableTwoFactor(c.Request.Context(), uid); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "2FA disabled successfully"})
}

// GetTwoFactorStatus gets the 2FA status for the current user
// GET /api/v1/auth/2fa/status
func (h *TwoFactorHandler) GetTwoFactorStatus(c *gin.Context) {
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

	status, err := h.twoFactorService.GetTwoFactorStatus(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if status == nil {
		c.JSON(http.StatusOK, gin.H{
			"enabled": false,
			"method":  nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"enabled":          status.Status == models.TwoFactorStatusEnabled,
		"method":           status.Method,
		"backupCodesCount": status.BackupCodesCount,
		"verifiedAt":       status.VerifiedAt,
	})
}

// VerifyTOTP verifies a TOTP code (used during login)
// POST /api/v1/auth/2fa/verify
func (h *TwoFactorHandler) VerifyTOTP(c *gin.Context) {
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

	var req models.VerifyTOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	valid, err := h.twoFactorService.VerifyTOTP(c.Request.Context(), uid, req.Code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid verification code"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Verification successful"})
}

// RegenerateBackupCodes regenerates backup codes for a user
// POST /api/v1/auth/2fa/backup-codes/regenerate
func (h *TwoFactorHandler) RegenerateBackupCodes(c *gin.Context) {
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

	// Verify TOTP code before regenerating
	var req models.VerifyTOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	valid, err := h.twoFactorService.VerifyTOTP(c.Request.Context(), uid, req.Code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid verification code"})
		return
	}

	// Regenerate backup codes
	codes, err := h.twoFactorService.RegenerateBackupCodes(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"backupCodes": codes,
		"message":     "Backup codes regenerated successfully",
	})
}

// GetQRCode generates and returns QR code image for TOTP
// GET /api/v1/auth/2fa/qr
func (h *TwoFactorHandler) GetQRCode(c *gin.Context) {
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

	// Set content type
	c.Header("Content-Type", "image/png")

	// Generate and write QR code
	if err := h.twoFactorService.GenerateQRCode(c.Request.Context(), uid, c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
}
