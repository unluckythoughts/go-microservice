package auth

import (
	"fmt"
	"net/http"

	"github.com/unluckythoughts/go-microservice/tools/web"
	"github.com/unluckythoughts/go-microservice/utils"
)

func (a *Service) LoginHandler(r web.Request) (any, error) {
	details := Credentials{}
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

func (a *Service) RegisterHandler(r web.Request) (any, error) {
	details := RegisterRequest{}
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

func (a *Service) LogoutHandler(r web.Request) (any, error) {
	// Implement logout logic here
	// This could involve clearing session data or tokens
	return "logout successful", nil
}

func (a *Service) GetUser(r web.Request) (*User, error) {
	return GetAuthenticatedUser(r)
}

func (a *Service) UpdateUserHandler(r web.Request) (any, error) {
	user, err := GetAuthenticatedUser(r)
	if err != nil {
		return nil, err
	}

	body := UpdateUserRequest{}
	err = r.GetValidatedBody(&body)
	if err != nil {
		return nil, err
	}

	err = a.UpdateUserPartial(user.ID, body)
	if err != nil {
		return nil, err
	}

	return "user updated successfully", nil
}

func (a *Service) ChangePasswordHandler(r web.Request) (any, error) {
	user, err := GetAuthenticatedUser(r)
	if err != nil {
		return nil, err
	}

	body := ChangePasswordRequest{}
	err = r.GetValidatedBody(&body)
	if err != nil {
		return nil, err
	}

	err = a.ChangeUserPassword(user.ID, body.OldPassword, body.NewPassword)
	if err != nil {
		return nil, err
	}

	return "password changed successfully", nil
}

func (a *Service) VerifyHandler(r web.Request) (any, error) {
	token := r.GetRouteParam("token")

	if token == "" {
		return nil, web.NewError(http.StatusBadRequest, fmt.Errorf("verification token is required"))
	}

	user, err := a.VerifyUserToken(token)
	if err != nil {
		return nil, web.NewError(http.StatusBadRequest, fmt.Errorf("invalid verification token: %w", err))
	}

	if !user.EmailVerified && user.Email != "" {
		err := a.UpdateEmailVerified(user.ID, true)
		if err != nil {
			return nil, web.NewError(http.StatusInternalServerError, fmt.Errorf("failed to verify email: %w", err))
		}
	} else if !user.MobileVerified && user.Mobile != "" {
		err := a.UpdateMobileVerified(user.ID, true)
		if err != nil {
			return nil, web.NewError(http.StatusInternalServerError, fmt.Errorf("failed to verify mobile: %w", err))
		}
	}

	return "verification successful", nil
}

func (a *Service) UpdatePasswordHandler(r web.Request) (any, error) {
	body := UpdatePasswordRequest{}
	err := r.GetValidatedBody(&body)
	if err != nil {
		return nil, err
	}
	if body.VerifyToken == "" || body.NewPassword == "" {
		return nil, web.NewError(http.StatusBadRequest, fmt.Errorf("verify token and new password are required"))
	}

	_, err = a.VerifyUserToken(body.VerifyToken)
	if err != nil {
		return nil, err
	}

	return "password reset successful", nil
}

func (a *Service) SendVerificationHandler(r web.Request) (any, error) {
	body := SendVerificationRequest{}
	err := r.GetValidatedBody(&body)
	if err != nil {
		return nil, err
	}

	var user *User
	if body.Email != "" {
		user, err = a.GetUserByEmail(body.Email)
	} else if body.Mobile != "" {
		user, err = a.GetUserByMobile(body.Mobile)
	}

	if err != nil {
		return nil, web.NewError(http.StatusBadRequest, fmt.Errorf("failed to find user: %w", err))
	} else if user == nil {
		return nil, web.NewError(http.StatusBadRequest, fmt.Errorf("user not found"))
	}

	token, err := utils.GenerateRandomString(16)
	if err != nil {
		return nil, web.NewError(http.StatusInternalServerError, fmt.Errorf("failed to generate verification token: %w", err))
	}

	err = a.UpdateUserVerifyToken(user.ID, token)
	if err != nil {
		return nil, web.NewError(http.StatusInternalServerError, fmt.Errorf("failed to update verification token: %w", err))
	}

	// TODO: Send verification email or SMS
	// This could involve using an email service or SMS gateway to send the token to the user

	return "verification sent successfully", nil
}
