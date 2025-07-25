package auth

import (
	"fmt"
	"net/http"

	"github.com/unluckythoughts/go-microservice/tools/web"
)

func (a *auth) LoginHandler(r web.Request) (any, error) {
	details := credentials{}
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

func (a *auth) RegisterHandler(r web.Request) (any, error) {
	details := registerRequest{}
	err := r.GetValidatedBody(&details)
	if err != nil {
		return "", err
	}

	if details.Email == "" && details.Mobile == "" {
		return "", web.NewError(http.StatusBadRequest, fmt.Errorf("email or mobile is required"))
	}

	user := User{
		Name:     details.Name,
		Email:    details.Email,
		Role:     a.defaultUserRole,
		Password: details.Password,
	}

	if details.Mobile != "" {
		var mobile Mobile
		err := mobile.Set(details.Mobile)
		if err != nil {
			return "", web.NewError(http.StatusBadRequest, fmt.Errorf("invalid mobile number: %w", err))
		}
		user.Mobile = mobile
	}

	err = a.CreateUser(&user)
	if err != nil {
		return "", err
	}

	return "user registered successfully", nil
}

func (a *auth) LogoutHandler(r web.Request) (any, error) {
	// Implement logout logic here
	// This could involve clearing session data or tokens
	return "logout successful", nil
}
