package auth

import (
	"time"

	"github.com/unluckythoughts/go-microservice/utils"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"gorm.io/gorm"
)

type auth struct {
	db           *gorm.DB
	ignoreRoutes []string
	jwtKey       string
	tokenValid   time.Duration
	// Roles are defined as a map where the key is UserRole and the value is the role name
	// Higher value UserRoles have more privileges and can access all resources of lower value UserRoles
	userRoles         map[UserRole]string
	googleOAuthConfig oauth2.Config
}

type Options struct {
	DB                *gorm.DB
	JwtKey            string `env:"AUTH_JWT_KEY"`
	TokenValidInHours uint   `env:"AUTH_TOKEN_VALID" envDefault:"4"`
	IgnoreRoutes      []string
	UserRoles         map[UserRole]string `env:"AUTH_USER_ROLES" envDefault:"0:user,99:admin"`
	GoogleOauth       struct {
		ClientID     string `env:"CLIENT_ID"`
		ClientSecret string `env:"CLIENT_SECRET"`
		RedirectURI  string `env:"REDIRECT_URI"`
	} `envPrefix:"GOOGLE_OAUTH_"`
}

func NewAuthService(opts Options) *auth {
	utils.ParseEnvironmentVars(&opts)
	return &auth{
		db:           opts.DB,
		ignoreRoutes: opts.IgnoreRoutes,
		jwtKey:       opts.JwtKey,
		tokenValid:   time.Duration(opts.TokenValidInHours) * time.Hour,
		userRoles:    opts.UserRoles,
		googleOAuthConfig: oauth2.Config{
			ClientID:     opts.GoogleOauth.ClientID,
			ClientSecret: opts.GoogleOauth.ClientSecret,
			RedirectURL:  opts.GoogleOauth.RedirectURI,
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/userinfo.profile",
			},
			Endpoint: google.Endpoint,
		},
	}
}

func (a *auth) RoleName(role UserRole) string {
	return a.userRoles[role]
}

func (a *auth) GetUserRoles() map[UserRole]string {
	return a.userRoles
}
