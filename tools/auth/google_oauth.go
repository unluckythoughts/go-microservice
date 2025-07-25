package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/unluckythoughts/go-microservice/tools/web"
)

// getGoogleUserInfo fetches user information from Google OAuth
func (a *Service) getGoogleUserInfo(accessToken string) (*googleUserInfo, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get user info: %s", resp.Status)
	}

	var userInfo googleUserInfo
	err = json.NewDecoder(resp.Body).Decode(&userInfo)
	if err != nil {
		return nil, err
	}

	return &userInfo, nil
}

// GoogleOAuthLogin handles Google OAuth authentication
func (a *Service) GoogleOAuthLogin(r web.Request) (any, error) {
	// Extract the OAuth request from the request body
	oauthReq := &googleOAuthRequest{}
	err := r.GetValidatedBody(oauthReq)
	if err != nil {
		return nil, web.NewError(http.StatusBadRequest, err)
	}

	// Exchange authorization code for access token
	token, err := a.googleOAuthConfig.Exchange(context.Background(), oauthReq.Code)
	if err != nil {
		return nil, web.NewError(http.StatusBadRequest, fmt.Errorf("failed to exchange code for token: %w", err))
	}

	// Get user info from Google
	userInfo, err := a.getGoogleUserInfo(token.AccessToken)
	if err != nil {
		return nil, web.NewError(http.StatusInternalServerError, fmt.Errorf("failed to get user info: %w", err))
	}

	// Check if user exists by Google ID
	user, err := a.GetUserByGoogleID(userInfo.ID)
	if err != nil {
		// User doesn't exist, create new user
		user = &User{
			Name:          userInfo.Name,
			Email:         userInfo.Email,
			GoogleID:      userInfo.ID,
			GoogleAvatar:  userInfo.Picture,
			Role:          a.defaultUserRole, // default role for new users
			EmailVerified: true,
		}

		err = a.CreateUser(user)
		if err != nil {
			return nil, web.NewError(http.StatusInternalServerError, fmt.Errorf("failed to create user: %w", err))
		}
	}

	return a.getAuthResponse(r.GetContext(), user)
}
