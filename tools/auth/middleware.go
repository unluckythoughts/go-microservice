package auth

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/unluckythoughts/go-microservice/v2/tools/web"
)

func (s *Service) getAuthResponse(ctx web.Context, user *User) (LoginResponse, error) {
	resp := LoginResponse{}

	err := ctx.PutSessionValue("user_id", user.ID)
	if err != nil {
		return resp, err
	}

	// Generate a JWT token for the user
	strToken, err := web.CreateJWT(s.jwtKey, jwt.MapClaims{
		"sub": strconv.Itoa(int(user.ID)),
		"iss": "table-app",
		"aud": []string{user.Role.Value()},
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(s.tokenValid).Unix(),
	})
	if err != nil {
		return resp, err
	}

	resp.Token = strToken

	// Generate a CSRF token for this session so the client can include it
	// in the X-CSRF-Token header on subsequent state-changing requests.
	csrfToken, err := web.GenerateCSRFToken(ctx)
	if err != nil {
		return resp, fmt.Errorf("failed to generate CSRF token: %w", err)
	}
	resp.CSRFToken = csrfToken

	return resp, nil
}

func (s *Service) isRouteIgnored(path string) bool {
	for _, route := range s.ignoreRoutes {
		if strings.HasPrefix(path, route) {
			return true
		}
	}
	return false
}

func getUserDataFromAuthHeader(headerValue string, secret string) (uint, error) {
	if headerValue == "" {
		return 0, fmt.Errorf("authorization header is empty")
	}

	headerValue = strings.TrimPrefix(headerValue, "Bearer ")
	token, err := web.ParseJWT(secret, headerValue)
	if err != nil {
		return 0, err
	}

	if !token.Valid {
		return 0, fmt.Errorf("invalid bearer token")
	}

	// Get the claims from the token
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, fmt.Errorf("error getting claims from JWT token")
	}

	// Get the user ID from the claims
	strUserID, err := claims.GetSubject()
	if err != nil {
		return 0, fmt.Errorf("error getting user ID from JWT token: %w", err)
	}

	// Convert the user ID to an integer
	intUserID, err := strconv.Atoi(strUserID)
	if err != nil {
		return 0, fmt.Errorf("error converting user ID to integer: %w", err)
	}

	return uint(intUserID), nil
}
func (s *Service) getUserFromRequest(r web.MiddlewareRequest) (*User, error) {
	authHeader := r.GetHeader("Authorization")
	if authHeader != "" {
		userID, err := getUserDataFromAuthHeader(authHeader, s.jwtKey)
		if err != nil {
			return nil, web.NewError(http.StatusUnauthorized, fmt.Errorf("unauthorized: %w", err))
		}
		user, err := s.GetUserByID(userID)
		if err != nil {
			return nil, web.NewError(http.StatusUnauthorized, fmt.Errorf("unauthorized: Please log in to access this link; %w", err))
		}
		return user, nil
	}

	// Session-based auth: validate CSRF on state-changing requests before trusting the session.
	switch r.GetMethod() {
	case "POST", "PUT", "PATCH", "DELETE":
		if err := web.ValidateCSRFToken(r); err != nil {
			return nil, err
		}
	}

	strUserID, err := r.GetContext().GetSessionValue("user_id")
	if err != nil {
		return nil, web.NewError(http.StatusUnauthorized, errors.New("unauthorized: Please log in to access this link"))
	}

	userID, ok := strUserID.(uint)
	if !ok || userID <= 0 {
		return nil, web.NewError(http.StatusUnauthorized, errors.New("unauthorized: Please log in to access this link"))
	}

	user, err := s.GetUserByID(userID)
	if err != nil {
		return nil, web.NewError(http.StatusUnauthorized, fmt.Errorf("unauthorized: Please log in to access this link; %w", err))
	}

	return user, nil
}

func (s *Service) GetAuthMiddleware() web.Middleware {
	return func(r web.MiddlewareRequest) error {
		if s.isRouteIgnored(r.GetPath()) {
			return nil
		}

		user, err := s.getUserFromRequest(r)
		if err != nil {
			return err
		}

		r.GetContext().PutSessionValue("user", user)
		return nil
	}
}

func GetAuthenticatedUser(r web.Request) (*User, error) {
	user, err := r.GetContext().GetSessionValue("user")
	if err != nil {
		return nil, web.NewError(http.StatusUnauthorized, fmt.Errorf("unauthorized: %w", err))
	}

	if user == nil {
		return nil, web.NewError(http.StatusUnauthorized, fmt.Errorf("unauthorized: Please log in to access this link"))
	}

	authUser, ok := user.(*User)
	if !ok {
		return nil, web.NewError(http.StatusInternalServerError, fmt.Errorf("internal server error: user data is not valid"))
	}

	return authUser, nil
}

func (s *Service) EnsureRole(role Role) web.Middleware {
	return func(r web.MiddlewareRequest) error {
		user, err := s.getUserFromRequest(r)
		if err != nil {
			return err
		}

		if user.Role < role {
			return web.NewError(http.StatusForbidden, fmt.Errorf("forbidden: You do not have permission to access this resource"))
		}

		r.GetContext().PutSessionValue("user", user)

		return nil
	}
}
