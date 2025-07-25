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
	userRoles                map[UserRole]string
	defaultUserRole          UserRole
	defaultMobileCountryCode string
	googleOAuthConfig        oauth2.Config
}

type Options struct {
	// DB is the database connection, if nil, it will use the default database connection
	DB *gorm.DB
	// JwtKey is the secret key used to sign JWT tokens
	// If empty, it will panic
	JwtKey string `env:"AUTH_JWT_KEY"`
	// TokenValidInHours is the duration for which the JWT token is valid
	// Default is 4 hours
	TokenValidInHours uint `env:"AUTH_TOKEN_VALID" envDefault:"4"`
	// IgnoreRoutes are the routes that do not require authentication
	// Default is /api/v1/auth/login
	// This can be a comma-separated list of routes
	// e.g. /api/v1/auth/login,/api/v1/auth/register
	IgnoreRoutes []string `env:"AUTH_IGNORE_ROUTES" envDefault:"/api/v1/auth/login"`
	// Roles are defined as a map where the key is UserRole(uint) and the value is the role name.
	// Higher UserRole(uint) has more privileges and can access all resources of lower UserRole(uint).
	// Default roles are 0:user, 99:admin
	UserRoles map[UserRole]string `env:"AUTH_USER_ROLES" envDefault:"0:user,99:admin"`

	// Default UserRole for new users
	DefaultUserRole UserRole `env:"AUTH_DEFAULT_USER_ROLE" envDefault:"0"`

	// Default Mobile country code for new users
	DefaultMobileCountryCode string `env:"AUTH_DEFAULT_MOBILE_COUNTRY_CODE" envDefault:"+1"`

	// GoogleOauth contains the configuration for Google OAuth
	GoogleOauth struct {
		ClientID     string `env:"CLIENT_ID"`
		ClientSecret string `env:"CLIENT_SECRET"`
		RedirectURI  string `env:"REDIRECT_URI"`
	} `envPrefix:"AUTH_GOOGLE_"`
}

func getOptions(override Options) Options {
	opts := Options{}
	utils.ParseEnvironmentVars(&opts)

	if override.DB != nil {
		opts.DB = override.DB
	}
	if override.JwtKey != "" {
		opts.JwtKey = override.JwtKey
	}
	if override.TokenValidInHours > 0 {
		opts.TokenValidInHours = override.TokenValidInHours
	}
	if len(override.IgnoreRoutes) > 0 {
		opts.IgnoreRoutes = override.IgnoreRoutes
	}
	if len(override.UserRoles) > 0 {
		opts.UserRoles = override.UserRoles
	}
	if override.DefaultUserRole != 0 {
		opts.DefaultUserRole = override.DefaultUserRole
	}
	if override.DefaultMobileCountryCode != "" {
		opts.DefaultMobileCountryCode = override.DefaultMobileCountryCode
	}
	if override.GoogleOauth.ClientID != "" && override.GoogleOauth.ClientSecret != "" {
		opts.GoogleOauth.ClientID = override.GoogleOauth.ClientID
		opts.GoogleOauth.ClientSecret = override.GoogleOauth.ClientSecret
	}

	if override.GoogleOauth.RedirectURI != "" {
		opts.GoogleOauth.RedirectURI = override.GoogleOauth.RedirectURI
	} else if opts.GoogleOauth.RedirectURI == "" {
		panic("Google OAuth Redirect URI is required")
	}

	return opts
}

func NewAuthService(override Options) *auth {
	opts := getOptions(override)

	a := &auth{
		db:           opts.DB,
		ignoreRoutes: opts.IgnoreRoutes,
		jwtKey:       opts.JwtKey,
		tokenValid:   time.Duration(opts.TokenValidInHours) * time.Hour,
	}

	if len(opts.UserRoles) == 0 {
		opts.UserRoles = map[UserRole]string{
			0:  "user",
			99: "admin",
		}
	}

	if opts.GoogleOauth.ClientID != "" && opts.GoogleOauth.ClientSecret != "" {
		a.googleOAuthConfig = oauth2.Config{
			ClientID:     opts.GoogleOauth.ClientID,
			ClientSecret: opts.GoogleOauth.ClientSecret,
			RedirectURL:  opts.GoogleOauth.RedirectURI,
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/userinfo.profile",
			},
			Endpoint: google.Endpoint,
		}
	}

	return a
}

func (a *auth) RoleName(role UserRole) string {
	return a.userRoles[role]
}

func (a *auth) GetUserRoles() map[UserRole]string {
	return a.userRoles
}
