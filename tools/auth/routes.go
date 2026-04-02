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

	// Auth routes
	r.POST(prefix+"/auth/login", as.LoginHandler)
	r.POST(prefix+"/auth/register", as.GetRegisterHandlerForUserRole(userRole))
	r.GET(prefix+"/auth/verify/:target/:token", as.VerifyTokenHandler)
	r.PUT(prefix+"/auth/update-password", as.UpdatePasswordHandler)

	// Protected auth routes
	r.GET(prefix+"/auth/logout", as.EnsureRole(userRole), as.LogoutHandler)

	// Password reset and update routes
	r.PATCH(prefix+"/auth/reset-password/:target", as.EnsureRole(userRole), as.ResetPasswordHandler)
	r.PUT(prefix+"/auth/change-password", as.EnsureRole(userRole), as.ChangePasswordHandler)

	// User routes
	r.GET(prefix+"/auth/user", as.EnsureRole(userRole), as.GetUserHandler)
	r.PUT(prefix+"/auth/user", as.EnsureRole(userRole), as.UpdateUserHandler)

	return nil
}
