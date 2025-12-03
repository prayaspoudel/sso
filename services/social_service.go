package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"sso/models"
	"sso/repository"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/linkedin"
)

// SocialService handles social login operations
type SocialService struct {
	socialRepo     *repository.SocialRepository
	userRepo       *repository.UserRepository
	googleConfig   *oauth2.Config
	githubConfig   *oauth2.Config
	linkedinConfig *oauth2.Config
}

// SocialServiceConfig represents social service configuration
type SocialServiceConfig struct {
	AppURL               string
	GoogleClientID       string
	GoogleClientSecret   string
	GitHubClientID       string
	GitHubClientSecret   string
	LinkedInClientID     string
	LinkedInClientSecret string
}

// NewSocialService creates a new SocialService
func NewSocialService(
	socialRepo *repository.SocialRepository,
	userRepo *repository.UserRepository,
	config SocialServiceConfig,
) *SocialService {
	service := &SocialService{
		socialRepo: socialRepo,
		userRepo:   userRepo,
	}

	// Google OAuth2 config
	if config.GoogleClientID != "" {
		service.googleConfig = &oauth2.Config{
			ClientID:     config.GoogleClientID,
			ClientSecret: config.GoogleClientSecret,
			RedirectURL:  fmt.Sprintf("%s/auth/google/callback", config.AppURL),
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/userinfo.profile",
			},
			Endpoint: google.Endpoint,
		}
	}

	// GitHub OAuth2 config
	if config.GitHubClientID != "" {
		service.githubConfig = &oauth2.Config{
			ClientID:     config.GitHubClientID,
			ClientSecret: config.GitHubClientSecret,
			RedirectURL:  fmt.Sprintf("%s/auth/github/callback", config.AppURL),
			Scopes:       []string{"user:email"},
			Endpoint:     github.Endpoint,
		}
	}

	// LinkedIn OAuth2 config
	if config.LinkedInClientID != "" {
		service.linkedinConfig = &oauth2.Config{
			ClientID:     config.LinkedInClientID,
			ClientSecret: config.LinkedInClientSecret,
			RedirectURL:  fmt.Sprintf("%s/auth/linkedin/callback", config.AppURL),
			Scopes:       []string{"r_emailaddress", "r_liteprofile"},
			Endpoint:     linkedin.Endpoint,
		}
	}

	return service
}

// GetAuthURL generates OAuth2 authorization URL
func (s *SocialService) GetAuthURL(ctx context.Context, provider models.SocialProvider, redirectURI string) (string, string, error) {
	// Generate state token
	state := uuid.New().String()

	// Store state
	stateRecord := &models.SocialLoginState{
		ID:          uuid.New(),
		State:       state,
		Provider:    provider,
		RedirectURI: redirectURI,
		ExpiresAt:   time.Now().Add(10 * time.Minute),
		CreatedAt:   time.Now(),
	}
	if err := s.socialRepo.CreateSocialLoginState(ctx, stateRecord); err != nil {
		return "", "", err
	}

	// Get auth URL based on provider
	var authURL string
	switch provider {
	case models.SocialProviderGoogle:
		authURL = s.googleConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
	case models.SocialProviderGitHub:
		authURL = s.githubConfig.AuthCodeURL(state)
	case models.SocialProviderLinkedIn:
		authURL = s.linkedinConfig.AuthCodeURL(state)
	default:
		return "", "", fmt.Errorf("unsupported provider: %s", provider)
	}

	return authURL, state, nil
}

// HandleCallback handles OAuth2 callback
func (s *SocialService) HandleCallback(ctx context.Context, provider models.SocialProvider, code, state string) (*models.SocialUserInfo, error) {
	// Verify state
	stateRecord, err := s.socialRepo.GetSocialLoginState(ctx, state)
	if err != nil {
		return nil, err
	}
	if stateRecord == nil {
		return nil, errors.New("invalid state")
	}
	if time.Now().After(stateRecord.ExpiresAt) {
		return nil, errors.New("state expired")
	}
	if stateRecord.Provider != provider {
		return nil, errors.New("provider mismatch")
	}

	// Delete state
	s.socialRepo.DeleteSocialLoginState(ctx, state)

	// Exchange code for token
	var config *oauth2.Config
	switch provider {
	case models.SocialProviderGoogle:
		config = s.googleConfig
	case models.SocialProviderGitHub:
		config = s.githubConfig
	case models.SocialProviderLinkedIn:
		config = s.linkedinConfig
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}

	token, err := config.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}

	// Get user info
	userInfo, err := s.getUserInfo(ctx, provider, token)
	if err != nil {
		return nil, err
	}

	return userInfo, nil
}

// getUserInfo fetches user info from social provider
func (s *SocialService) getUserInfo(ctx context.Context, provider models.SocialProvider, token *oauth2.Token) (*models.SocialUserInfo, error) {
	switch provider {
	case models.SocialProviderGoogle:
		return s.getGoogleUserInfo(ctx, token)
	case models.SocialProviderGitHub:
		return s.getGitHubUserInfo(ctx, token)
	case models.SocialProviderLinkedIn:
		return s.getLinkedInUserInfo(ctx, token)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}

// getGoogleUserInfo fetches user info from Google
func (s *SocialService) getGoogleUserInfo(ctx context.Context, token *oauth2.Token) (*models.SocialUserInfo, error) {
	client := s.googleConfig.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var googleUser struct {
		ID         string `json:"id"`
		Email      string `json:"email"`
		Name       string `json:"name"`
		GivenName  string `json:"given_name"`
		FamilyName string `json:"family_name"`
		Picture    string `json:"picture"`
	}

	if err := json.Unmarshal(data, &googleUser); err != nil {
		return nil, err
	}

	userInfo := &models.SocialUserInfo{
		Provider:     models.SocialProviderGoogle,
		ProviderID:   googleUser.ID,
		Email:        googleUser.Email,
		Name:         googleUser.Name,
		FirstName:    googleUser.GivenName,
		LastName:     googleUser.FamilyName,
		Avatar:       googleUser.Picture,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
	}

	if !token.Expiry.IsZero() {
		userInfo.ExpiresAt = &token.Expiry
	}

	return userInfo, nil
}

// getGitHubUserInfo fetches user info from GitHub
func (s *SocialService) getGitHubUserInfo(ctx context.Context, token *oauth2.Token) (*models.SocialUserInfo, error) {
	client := s.githubConfig.Client(ctx, token)

	// Get user profile
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var githubUser struct {
		ID        int64  `json:"id"`
		Login     string `json:"login"`
		Name      string `json:"name"`
		Email     string `json:"email"`
		AvatarURL string `json:"avatar_url"`
	}

	if err := json.Unmarshal(data, &githubUser); err != nil {
		return nil, err
	}

	// Get primary email if not in profile
	email := githubUser.Email
	if email == "" {
		emailResp, err := client.Get("https://api.github.com/user/emails")
		if err == nil {
			defer emailResp.Body.Close()
			emailData, err := io.ReadAll(emailResp.Body)
			if err == nil {
				var emails []struct {
					Email   string `json:"email"`
					Primary bool   `json:"primary"`
				}
				if json.Unmarshal(emailData, &emails) == nil {
					for _, e := range emails {
						if e.Primary {
							email = e.Email
							break
						}
					}
				}
			}
		}
	}

	userInfo := &models.SocialUserInfo{
		Provider:     models.SocialProviderGitHub,
		ProviderID:   fmt.Sprintf("%d", githubUser.ID),
		Email:        email,
		Name:         githubUser.Name,
		Avatar:       githubUser.AvatarURL,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
	}

	if !token.Expiry.IsZero() {
		userInfo.ExpiresAt = &token.Expiry
	}

	return userInfo, nil
}

// getLinkedInUserInfo fetches user info from LinkedIn
func (s *SocialService) getLinkedInUserInfo(ctx context.Context, token *oauth2.Token) (*models.SocialUserInfo, error) {
	client := s.linkedinConfig.Client(ctx, token)

	// Get user profile
	resp, err := client.Get("https://api.linkedin.com/v2/me")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var linkedinUser struct {
		ID        string `json:"id"`
		FirstName struct {
			Localized map[string]string `json:"localized"`
		} `json:"firstName"`
		LastName struct {
			Localized map[string]string `json:"localized"`
		} `json:"lastName"`
	}

	if err := json.Unmarshal(data, &linkedinUser); err != nil {
		return nil, err
	}

	// Get email
	emailResp, err := client.Get("https://api.linkedin.com/v2/emailAddress?q=members&projection=(elements*(handle~))")
	if err != nil {
		return nil, err
	}
	defer emailResp.Body.Close()

	emailData, err := io.ReadAll(emailResp.Body)
	if err != nil {
		return nil, err
	}

	var emailResult struct {
		Elements []struct {
			Handle struct {
				EmailAddress string `json:"emailAddress"`
			} `json:"handle~"`
		} `json:"elements"`
	}

	email := ""
	if json.Unmarshal(emailData, &emailResult) == nil && len(emailResult.Elements) > 0 {
		email = emailResult.Elements[0].Handle.EmailAddress
	}

	firstName := ""
	lastName := ""
	for _, v := range linkedinUser.FirstName.Localized {
		firstName = v
		break
	}
	for _, v := range linkedinUser.LastName.Localized {
		lastName = v
		break
	}

	userInfo := &models.SocialUserInfo{
		Provider:     models.SocialProviderLinkedIn,
		ProviderID:   linkedinUser.ID,
		Email:        email,
		Name:         fmt.Sprintf("%s %s", firstName, lastName),
		FirstName:    firstName,
		LastName:     lastName,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
	}

	if !token.Expiry.IsZero() {
		userInfo.ExpiresAt = &token.Expiry
	}

	return userInfo, nil
}

// LinkAccount links a social account to a user
func (s *SocialService) LinkAccount(ctx context.Context, userID uuid.UUID, userInfo *models.SocialUserInfo) error {
	// Check if account already linked to another user
	existing, err := s.socialRepo.GetSocialAccount(ctx, userInfo.Provider, userInfo.ProviderID)
	if err != nil {
		return err
	}
	if existing != nil && existing.UserID != userID {
		return errors.New("social account already linked to another user")
	}

	// Create or update social account
	now := time.Now()
	account := &models.SocialAccount{
		ID:          uuid.New(),
		UserID:      userID,
		Provider:    userInfo.Provider,
		ProviderID:  userInfo.ProviderID,
		Email:       userInfo.Email,
		Name:        userInfo.Name,
		AccessToken: userInfo.AccessToken,
		ExpiresAt:   userInfo.ExpiresAt,
		LinkedAt:    now,
		CreatedAt:   now,
	}

	if userInfo.Avatar != "" {
		account.Avatar = &userInfo.Avatar
	}
	if userInfo.RefreshToken != "" {
		account.RefreshToken = &userInfo.RefreshToken
	}

	return s.socialRepo.CreateSocialAccount(ctx, account)
}

// UnlinkAccount unlinks a social account from a user
func (s *SocialService) UnlinkAccount(ctx context.Context, userID uuid.UUID, provider models.SocialProvider) error {
	return s.socialRepo.DeleteSocialAccount(ctx, userID, provider)
}

// GetLinkedAccounts gets all linked social accounts for a user
func (s *SocialService) GetLinkedAccounts(ctx context.Context, userID uuid.UUID) ([]models.SocialAccount, error) {
	accounts, err := s.socialRepo.GetSocialAccountsByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Remove sensitive data
	for i := range accounts {
		accounts[i].AccessToken = ""
		accounts[i].RefreshToken = nil
	}

	return accounts, nil
}

// GetOrCreateUserFromSocial gets or creates a user from social login
func (s *SocialService) GetOrCreateUserFromSocial(ctx context.Context, userInfo *models.SocialUserInfo) (*models.User, bool, error) {
	// Check if social account exists
	socialAccount, err := s.socialRepo.GetSocialAccount(ctx, userInfo.Provider, userInfo.ProviderID)
	if err != nil {
		return nil, false, err
	}

	if socialAccount != nil {
		// Update last used
		s.socialRepo.UpdateSocialAccountLastUsed(ctx, socialAccount.ID)

		// Get user
		user, err := s.userRepo.GetByID(ctx, socialAccount.UserID)
		if err != nil {
			return nil, false, err
		}
		return user, false, nil
	}

	// Check if user exists with this email
	user, err := s.userRepo.GetByEmail(ctx, userInfo.Email)
	if err != nil {
		return nil, false, err
	}

	isNewUser := false
	if user == nil {
		// Create new user
		user = &models.User{
			ID:        uuid.New(),
			Email:     userInfo.Email,
			FirstName: userInfo.FirstName,
			LastName:  userInfo.LastName,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		// This would need to be implemented in your user repository
		// err = s.userRepo.Create(ctx, user)
		if err != nil {
			return nil, false, err
		}
		isNewUser = true
	}

	// Link social account
	if err := s.LinkAccount(ctx, user.ID, userInfo); err != nil {
		return nil, false, err
	}

	return user, isNewUser, nil
}
