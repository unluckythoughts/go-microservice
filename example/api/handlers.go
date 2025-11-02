package api

import (
	"errors"
	"net/http"
	"strings"

	"example/models"
	"example/service"

	"github.com/golang-jwt/jwt/v5"
	"github.com/unluckythoughts/go-microservice/tools/web"
	"go.uber.org/zap"
)

type Handlers struct {
	service   *service.Service
	jwtSecret string
}

func NewHandlers(s *service.Service, jwtSecret string) *Handlers {
	return &Handlers{
		service:   s,
		jwtSecret: jwtSecret,
	}
}

func getUserIDFromAuthHeader(headerValue string, secret string, l *zap.SugaredLogger) uint {
	if headerValue != "" {
		headerValue = strings.TrimPrefix(headerValue, "Bearer ")
		token, err := web.ParseJWT(secret, headerValue)
		if err == nil && token.Valid {
			claims, ok := token.Claims.(jwt.MapClaims)
			if ok {
				userIDFloat, exists := claims["user_id"]
				if exists {
					if userID, ok := userIDFloat.(float64); ok {
						return uint(userID)
					}
				}
			}
		}
	}
	return 0
}

// Authorized middleware
func (h *Handlers) Authorized(req web.MiddlewareRequest) error {
	l := req.GetContext().Logger()
	authHeader := req.GetHeader("Authorization")

	userID := getUserIDFromAuthHeader(authHeader, h.jwtSecret, l)
	if userID == 0 {
		return web.NewError(http.StatusUnauthorized, errors.New("unauthorized"))
	}

	// Get user from service layer
	user, err := h.service.GetUserByID(userID)
	if err != nil {
		return web.NewError(http.StatusUnauthorized, errors.New("unauthorized"))
	}

	// Store user in session
	err = req.GetContext().PutSessionValue("user", user)
	if err != nil {
		return web.NewError(http.StatusInternalServerError, errors.New("failed to store user in session"))
	}

	return nil
}

// Login handler
func (h *Handlers) Login(r web.Request) (any, error) {
	// Extract the login credentials from the request
	cred := &models.LoginRequest{}
	err := r.GetValidatedBody(cred)
	if err != nil {
		return nil, web.NewError(http.StatusBadRequest, err)
	}

	var user *models.User
	// Try to find user by email first, then by mobile number
	if cred.Email != "" {
		user, err = h.service.GetUserByEmail(cred.Email)
		if err != nil {
			// If email login fails, try mobile number if provided
			if cred.Mobile != "" {
				user, err = h.service.GetUserByMobileNumber(cred.Mobile)
				if err != nil {
					return nil, web.NewError(http.StatusUnauthorized, errors.New("invalid email/mobile or password"))
				}
			} else {
				return nil, web.NewError(http.StatusUnauthorized, errors.New("invalid email or password"))
			}
		}
	} else if cred.Mobile != "" {
		// Try mobile number login
		user, err = h.service.GetUserByMobileNumber(cred.Mobile)
		if err != nil {
			return nil, web.NewError(http.StatusUnauthorized, errors.New("invalid mobile number or password"))
		}
	} else {
		return nil, web.NewError(http.StatusBadRequest, errors.New("email or mobile number is required"))
	}

	if !models.CheckPasswordHash(cred.Password, user.Password) {
		return nil, web.NewError(http.StatusUnauthorized, errors.New("invalid email/mobile or password"))
	}

	return h.service.CreateAuthResponse(r, user)
}

func (h *Handlers) Signup(r web.Request) (any, error) {
	// Extract the user details from the request
	body := &models.SignupRequest{}
	err := r.GetValidatedBody(body)
	if err != nil {
		return nil, web.NewError(http.StatusBadRequest, err)
	}

	user := &models.User{
		Name:         body.Name,
		Email:        body.Email,
		MobileNumber: body.MobileNumber,
		Password:     body.Password,
		Role:         models.UserRoleUser,
		IsVerified:   false,
	}

	if user.Email == "" && user.MobileNumber == "" {
		return nil, web.NewError(http.StatusBadRequest, errors.New("email or mobile number is required"))
	}

	// Create the user in the database
	err = h.service.CreateUser(user)
	if err != nil {
		return nil, err
	}

	// Create default user settings
	settings := &models.UserSettings{
		UserID:               user.ID,
		Theme:                "light",
		Language:             "en",
		NotificationsEnabled: true,
		TimeZone:             "UTC",
	}
	h.service.CreateUserSettings(settings)

	return "User signup successful, please log in", nil
}

func (h *Handlers) Logout(r web.Request) (any, error) {
	r.GetContext().PutSessionValue("user_id", 0)
	r.GetContext().PutSessionValue("user", nil)
	return "Logged out successfully", nil
}

// User handlers
func (h *Handlers) GetUser(r web.Request) (any, error) {
	anyUser, err := r.GetContext().GetSessionValue("user")
	if err != nil {
		return nil, web.NewError(http.StatusUnauthorized, err)
	}

	user, ok := anyUser.(*models.User)
	if !ok {
		return nil, web.NewError(http.StatusUnauthorized, errors.New("unauthorized: Please log in to access this resource"))
	}

	// Clear password before returning
	user.Password = ""
	return user, nil
}

func (h *Handlers) UpdateUser(r web.Request) (any, error) {
	anyUser, err := r.GetContext().GetSessionValue("user")
	if err != nil {
		return nil, web.NewError(http.StatusUnauthorized, err)
	}

	currentUser, ok := anyUser.(*models.User)
	if !ok {
		return nil, web.NewError(http.StatusUnauthorized, errors.New("unauthorized"))
	}

	// Get update data from request
	updateData := &models.User{}
	err = r.GetValidatedBody(updateData)
	if err != nil {
		return nil, web.NewError(http.StatusBadRequest, err)
	}

	// Update allowed fields
	currentUser.Name = updateData.Name
	currentUser.MobileNumber = updateData.MobileNumber

	if err := h.service.UpdateUser(currentUser); err != nil {
		return nil, web.NewError(http.StatusInternalServerError, err)
	}

	// Clear password before returning
	currentUser.Password = ""
	return map[string]interface{}{
		"message": "user updated successfully",
		"user":    currentUser,
	}, nil
}

func (h *Handlers) DeleteUser(r web.Request) (any, error) {
	anyUser, err := r.GetContext().GetSessionValue("user")
	if err != nil {
		return nil, web.NewError(http.StatusUnauthorized, err)
	}

	currentUser, ok := anyUser.(*models.User)
	if !ok {
		return nil, web.NewError(http.StatusUnauthorized, errors.New("unauthorized"))
	}

	if err := h.service.DeleteUser(currentUser.ID); err != nil {
		return nil, web.NewError(http.StatusInternalServerError, err)
	}

	return "User deleted successfully", nil
}

// UserSettings handlers
func (h *Handlers) GetUserSettings(r web.Request) (any, error) {
	anyUser, err := r.GetContext().GetSessionValue("user")
	if err != nil {
		return nil, web.NewError(http.StatusUnauthorized, err)
	}

	currentUser, ok := anyUser.(*models.User)
	if !ok {
		return nil, web.NewError(http.StatusUnauthorized, errors.New("unauthorized"))
	}

	settings, err := h.service.GetUserSettings(currentUser.ID)
	if err != nil {
		return nil, web.NewError(http.StatusInternalServerError, err)
	}

	return map[string]interface{}{
		"settings": settings,
	}, nil
}

func (h *Handlers) UpdateUserSettings(r web.Request) (any, error) {
	anyUser, err := r.GetContext().GetSessionValue("user")
	if err != nil {
		return nil, web.NewError(http.StatusUnauthorized, err)
	}

	currentUser, ok := anyUser.(*models.User)
	if !ok {
		return nil, web.NewError(http.StatusUnauthorized, errors.New("unauthorized"))
	}

	var req models.UserSettings
	if err := r.GetValidatedBody(&req); err != nil {
		return nil, web.NewError(http.StatusBadRequest, err)
	}

	settings, err := h.service.UpdateUserSettings(currentUser.ID, req.Theme, req.Language, req.NotificationsEnabled, req.TimeZone)
	if err != nil {
		return nil, web.NewError(http.StatusInternalServerError, err)
	}

	return map[string]interface{}{
		"message":  "user settings updated successfully",
		"settings": settings,
	}, nil
}
