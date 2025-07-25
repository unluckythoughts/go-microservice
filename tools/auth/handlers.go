package auth

import (
	"fmt"
	"net/http"

	"github.com/unluckythoughts/go-microservice/tools/web"
)

func (a *auth) LoginHandler(r web.Request) (any, error) {
	details := loginRequest{}
	err := r.GetValidatedBody(&details)
	if err != nil {
		return "", err
	}

	if details.Email == "" && details.Mobile == "" {
		return "", web.NewError(http.StatusBadRequest, fmt.Errorf("email or mobile is required"))
	}

	if details.Mobile != "" {
		user, ok, err := a.VerifyUserPasswordByMobile(details.Mobile, details.Password)
		if err != nil {
			return "", err
		}

		if !ok {
			return "", fmt.Errorf("invalid mobile or password")
		}
		return a.getAuthResponse(r.GetContext(), user)
	}

	user, ok, err := a.VerifyUserPasswordByEmail(details.Email, details.Password)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", fmt.Errorf("invalid email or password")
	}

	return a.getAuthResponse(r.GetContext(), user)
}

func (a *auth) LogoutHandler(r web.Request) (any, error) {
	// Implement logout logic here
	// This could involve clearing session data or tokens
	return "logout successful", nil
}
