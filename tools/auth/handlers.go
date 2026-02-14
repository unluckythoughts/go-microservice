package auth

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/unluckythoughts/go-microservice/v2/tools/web"
	"gorm.io/gorm"
)

// LogoutHandler handles user logout requests
// example path: GET .../logout
func (a *Auth) LogoutHandler(r web.Request) (any, error) {
	r.GetContext().PutSessionValue("user", nil)
	r.GetContext().PutSessionValue("user_id", 0)
	return "logout successful", nil
}

// LoginHandler handles user login requests
// example path: POST .../login
func (a *Auth) LoginHandler(r web.Request) (any, error) {
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

// SendTokenHandler handles the creation of a verification token for a given target (email or mobile)
// example path: PATCH .../verify/:target?type=(email or mobile)
func (a *Auth) SendTokenHandler(r web.Request) (any, error) {
	targetType := r.GetURLParam("type")
	target := r.GetRouteParam("target")

	if target == "" {
		return nil, web.NewError(http.StatusBadRequest, fmt.Errorf("target is required"))
	}

	token, err := a.CreateVerifyToken(target)
	if err != nil {
		return nil, web.NewError(http.StatusInternalServerError, err)
	}

	switch targetType {
	case "email", "":
		r.GetContext().Logger().Info("verification token created for email", "email", target, "token", token)
		// TODO: send verification email to the user with the token
	case "mobile":
		r.GetContext().Logger().Info("verification token created for mobile", "mobile", target, "token", token)
		// TODO: send verification SMS to the user with the token
	default:
		return nil, web.NewError(http.StatusBadRequest, fmt.Errorf("invalid target type: must be 'email' or 'mobile'"))
	}

	return "verification token created successfully", nil

}

// VerifyTokenHandler handles the verification of a token for a given target (email or mobile)
// example path: GET .../verify/:target/:token
func (a *Auth) VerifyTokenHandler(r web.Request) (any, error) {
	target := r.GetRouteParam("target")
	token := r.GetRouteParam("token")

	if token == "" {
		return nil, web.NewError(http.StatusBadRequest, fmt.Errorf("verification token is required"))
	}

	ok, err := a.VerifyToken(target, token)
	if err != nil {
		if errors.Is(err, ErrExpiredToken) {
			return nil, web.NewError(http.StatusBadRequest, err)
		} else if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, web.NewError(http.StatusBadRequest, fmt.Errorf("invalid verification token"))
		}
		return nil, web.NewError(http.StatusInternalServerError, err)
	}

	return ok, nil
}

// GetRegisterHandlerForUserRole returns a handler for user registration with a specific role
// example path: POST .../register
func (a *Auth) GetRegisterHandlerForUserRole(role Role) web.Handler {
	return func(r web.Request) (any, error) {
		details := RegisterRequest{}
		err := r.GetValidatedBody(&details)
		if err != nil {
			return nil, err
		}

		if details.Email == "" && details.Mobile == "" {
			return nil, web.NewError(http.StatusBadRequest, fmt.Errorf("email or mobile is required"))
		}

		user := User{
			Name:     details.Name,
			Email:    details.Email,
			Role:     role,
			Password: details.Password,
		}

		if details.Mobile != "" {
			err := user.Mobile.Set(details.Mobile)
			if err != nil {
				return nil, web.NewError(http.StatusBadRequest, fmt.Errorf("invalid mobile number: %w", err))
			}
		}

		if user.Email != "" && a.IsVerified(user.Email) {
			user.EmailVerified = true
		}

		if user.Mobile.String() != "" && a.IsVerified(user.Mobile.String()) {
			user.MobileVerified = true
		}

		err = a.CreateUser(&user)
		if err != nil {
			return nil, web.NewError(http.StatusInternalServerError, err)
		}

		return "user registered successfully", nil
	}
}

// GetUser returns the currently authenticated user
// example path: GET .../user
func (a *Auth) GetUserHandler(r web.Request) (any, error) {
	user, err := GetAuthenticatedUser(r)
	if err != nil {
		return nil, err
	}

	return a.getAuthResponse(r.GetContext(), user)
}

// UpdateUserHandler handles user profile update requests
// example path: PUT .../user
func (a *Auth) UpdateUserHandler(r web.Request) (any, error) {
	user, err := GetAuthenticatedUser(r)
	if err != nil {
		return nil, err
	}

	body := UpdateUserRequest{}
	err = r.GetValidatedBody(&body)
	if err != nil {
		return nil, err
	}

	new_user := User{
		Name:  body.Name,
		Email: body.Email,
	}
	new_user.Mobile.Set(body.Mobile)

	if new_user.Email != "" && new_user.Email != user.Email {
		new_user.EmailVerified = false
	}

	if new_user.Mobile != "" && new_user.Mobile != user.Mobile {
		new_user.MobileVerified = false
	}

	err = a.UpdateUserPartial(user.ID, new_user)
	if err != nil {
		return nil, err
	}

	return "user updated successfully", nil
}

// ChangePasswordHandler handles password change requests for authenticated users
// example path: POST .../user/change-password
func (a *Auth) ChangePasswordHandler(r web.Request) (any, error) {
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

func getResetTarget(user *User, targetType string) (string, string, error) {
	if targetType == "email" && user.Email != "" {
		return user.Email + ":email-reset-password", "email", nil
	}

	if targetType == "mobile" && user.Mobile.String() != "" {
		return user.Mobile.String() + ":mobile-reset-password", "mobile", nil
	}

	if user.Email != "" {
		return user.Email + ":email-reset-password", "email", nil
	}

	if user.Mobile.String() != "" {
		return user.Mobile.String() + ":mobile-reset-password", "mobile", nil
	}

	return "", "", fmt.Errorf("user has neither email nor mobile")
}

func (a *Auth) getTarget(target string) (string, bool) {
	if target == "" {
		return "", false
	}

	vals := strings.Split(target, ":")
	if len(vals) < 2 {
		return "", false
	}

	switch vals[len(vals)-1] {
	case "email-reset-password", "mobile-reset-password":
		return strings.Join(vals[:len(vals)-1], ":"), true
	}

	return "", false
}

// ResetPasswordHandler handles password reset requests
// example path: GET .../user/reset-password/:target?type=(email or mobile)
func (a *Auth) ResetPasswordHandler(r web.Request) (any, error) {
	targetType := r.GetURLParam("type")
	target := r.GetRouteParam("target")

	if target == "" {
		return nil, web.NewError(http.StatusBadRequest, fmt.Errorf("target is required"))
	}

	var user User
	err := a.db.Where("email = ? OR mobile = ?", target, target).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, web.NewError(http.StatusBadRequest, fmt.Errorf("user not found for the given target"))
		}
		return nil, web.NewError(http.StatusInternalServerError, err)
	}

	target, targetType, err = getResetTarget(&user, targetType)
	if err != nil {
		return nil, web.NewError(http.StatusInternalServerError, err)
	}

	token, err := a.CreateVerifyToken(target)
	if err != nil {
		return nil, web.NewError(http.StatusInternalServerError, err)
	}

	switch targetType {
	case "email", "":
		r.GetContext().Logger().Info("verification token created for email", "email", target, "token", token)
		// TODO: send verification email to the user with the token
	case "mobile":
		r.GetContext().Logger().Info("verification token created for mobile", "mobile", target, "token", token)
		// TODO: send verification SMS to the user with the token
	default:
		return nil, web.NewError(http.StatusBadRequest, fmt.Errorf("invalid target type: must be 'email' or 'mobile'"))
	}

	return "verification token created successfully", nil

}

// UpdatePasswordHandler handles password reset requests using a verification token
// example path: POST .../user/update-password
func (a *Auth) UpdatePasswordHandler(r web.Request) (any, error) {
	body := UpdatePasswordRequest{}
	err := r.GetValidatedBody(&body)
	if err != nil {
		return nil, err
	}

	v, err := a.GetVerification(body.VerifyToken)
	if err != nil {
		return nil, err
	}

	target, ok := a.getTarget(v.Target)
	if !ok {
		return nil, web.NewError(http.StatusInternalServerError, fmt.Errorf("invalid verification target"))
	}

	var user User
	err = a.db.Where("email = ? OR mobile = ?", target, target).First(&user).Error
	if err != nil {
		return nil, web.NewError(http.StatusInternalServerError, fmt.Errorf("user not found for the given verification token"))
	}

	err = a.UpdateUserPassword(user.ID, body.NewPassword)
	if err != nil {
		return nil, err
	}

	return "password reset successful", nil
}
