package auth

import (
	"fmt"
	"strings"

	"github.com/unluckythoughts/go-microservice/v2/tools/web"
)

func RegisterAuthRoutes(r web.Router, prefix string, as *Service, userRole Role) error {
	if userRole == 0 {
		userRole = Role(1)
	}

	if prefix == "" {
		prefix = "/"
	} else if !strings.HasPrefix(prefix, "/") {
		return fmt.Errorf("route prefix has to start with '/'")
	}

	prefix = strings.TrimRight(prefix, "/")

	// Login route is public and does not require authentication
	r.POST(prefix+"/auth/login", as.LoginHandler)

	// Register route is public and does not require authentication
	// it registers a new user with the specified role (userRole)
	r.POST(prefix+"/auth/register", as.GetRegisterHandlerForUserRole(userRole))

	// Verify token route is public and does not require authentication
	// it verifies the token for the specified target (email or phone)
	r.GET(prefix+"/auth/verify/:target/:token", as.VerifyTokenHandler)

	// Update password route is protected and requires authentication
	// it updates the password for the authenticated user
	r.PUT(prefix+"/auth/update-password", as.UpdatePasswordHandler)

	// Logout route is protected and requires authentication
	// it logs out the authenticated user and invalidates the session
	r.GET(prefix+"/auth/logout", as.EnsureRole(userRole), as.LogoutHandler)

	// Reset password route is protected and requires authentication
	// it creates a verification token for resetting the password for the specified target (email or phone)
	r.PATCH(prefix+"/auth/reset-password/:target", as.EnsureRole(userRole), as.ResetPasswordHandler)

	// Change password route is protected and requires authentication
	// it changes the password for the authenticated user
	r.PUT(prefix+"/auth/change-password", as.EnsureRole(userRole), as.ChangePasswordHandler)

	// User route is protected and requires authentication
	// it retrieves the authenticated user's information
	r.GET(prefix+"/auth/user", as.EnsureRole(userRole), as.GetUserHandler)

	// Update user route is protected and requires authentication
	// it updates the authenticated user's information
	r.PUT(prefix+"/auth/user", as.EnsureRole(userRole), as.UpdateUserHandler)

	return nil
}
