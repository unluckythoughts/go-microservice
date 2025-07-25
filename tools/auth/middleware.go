package auth

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/unluckythoughts/go-microservice/tools/web"
)

func (a *Service) getAuthResponse(ctx web.Context, user *User) (LoginResponse, error) {
	resp := LoginResponse{}

	err := ctx.PutSessionValue("user_id", user.ID)
	if err != nil {
		return resp, err
	}

	// Generate a JWT token for the user
	strToken, err := web.CreateJWT(a.jwtKey, jwt.MapClaims{
		"sub": strconv.Itoa(int(user.ID)),
		"iss": "table-app",
		"aud": []string{user.Role.Value()},
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(a.tokenValid).Unix(),
	})
	if err != nil {
		return resp, err
	}

	resp.Token = strToken
	return resp, nil
}

func (a *Service) isRouteIgnored(path string) bool {
	for _, route := range a.ignoreRoutes {
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
func (a *Service) getUserFromRequest(r web.MiddlewareRequest) (*User, error) {
	userID, err := getUserDataFromAuthHeader(r.GetHeader("Authorization"), a.jwtKey)
	if err != nil {
		return nil, web.NewError(http.StatusUnauthorized, fmt.Errorf("unauthorized: %w", err))
	}

	// If the user ID is not found in the header, check the session
	if userID == 0 {
		strUserID, err := r.GetContext().GetSessionValue("user_id")
		if err != nil {
			return nil, web.NewError(http.StatusUnauthorized, errors.New("unauthorized: Please log in to access this link"))
		}

		userID, ok := strUserID.(uint)
		if !ok || userID <= 0 {
			return nil, web.NewError(http.StatusUnauthorized, errors.New("unauthorized: Please log in to access this link"))
		}
	}

	// Check if the user exists in the database
	user, err := a.GetUserByID(userID)
	if err != nil {
		return nil, web.NewError(http.StatusUnauthorized, fmt.Errorf("unauthorized: Please log in to access this link; %w", err))
	}

	return user, nil
}

func (a *Service) GetAuthMiddleware() web.Middleware {
	return func(r web.MiddlewareRequest) error {
		if a.isRouteIgnored(r.GetPath()) {
			return nil
		}

		user, err := a.getUserFromRequest(r)
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

func (a *Service) EnsureRole(role UserRole) web.Middleware {
	return func(r web.MiddlewareRequest) error {
		user, err := a.getUserFromRequest(r)
		if err != nil {
			return err
		}

		if user.Role >= role {
			return web.NewError(http.StatusForbidden, fmt.Errorf("forbidden: You do not have permission to access this resource"))
		}

		r.GetContext().PutSessionValue("user", user)

		return nil
	}
}
