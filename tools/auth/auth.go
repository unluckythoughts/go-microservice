package auth

import (
	"fmt"
	"time"

	"github.com/unluckythoughts/go-microservice/v2/utils"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"gorm.io/gorm"
)

type Auth struct {
	db           *gorm.DB
	l            *zap.Logger
	ignoreRoutes []string
	jwtKey       string
	tokenValid   time.Duration
	// Roles are defined as a map where the key is UserRole and the value is the role name
	// Higher value UserRoles have more privileges and can access all resources of lower value UserRoles
	userRoles                map[UserRole]string
	defaultMobileCountryCode string
	GoogleOauthConfig        oauth2.Config
}

type Options struct {
	// DB is the database connection, if nil, it will use the default database connection
	DB *gorm.DB
	// Logger is the logger to use, if nil, it will use the default logger
	Logger *zap.Logger
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
	IgnoreRoutes []string
	// Roles are defined as a map where the key is UserRole(uint) and the value is the role name.
	// Higher UserRole(uint) has more privileges and can access all resources of lower UserRole(uint).
	// Default roles are 0:user, 99:admin
	UserRoles map[UserRole]string
	// Default Mobile country code for new users
	DefaultMobileCountryCode string `env:"AUTH_DEFAULT_MOBILE_COUNTRY_CODE" envDefault:"+1"`

	// GoogleOauth contains the configuration for Google OAuth
	GoogleOauth struct {
		ClientID     string `env:"CLIENT_ID"`
		ClientSecret string `env:"CLIENT_SECRET"`
	} `envPrefix:"AUTH_GOOGLE_"`
}

func getOptions(override Options) Options {
	opts := Options{}
	utils.ParseEnvironmentVars(&opts)

	if override.DB != nil {
		opts.DB = override.DB
	}
	if override.Logger != nil {
		opts.Logger = override.Logger
	} else {
		panic("Logger is required for auth service")
	}
	if override.JwtKey != "" {
		opts.JwtKey = override.JwtKey
	}
	if override.TokenValidInHours > 0 {
		opts.TokenValidInHours = override.TokenValidInHours
	}
	if override.DefaultMobileCountryCode != "" {
		opts.DefaultMobileCountryCode = override.DefaultMobileCountryCode
	}
	if override.GoogleOauth.ClientID != "" && override.GoogleOauth.ClientSecret != "" {
		opts.GoogleOauth.ClientID = override.GoogleOauth.ClientID
		opts.GoogleOauth.ClientSecret = override.GoogleOauth.ClientSecret
	}

	opts.IgnoreRoutes = override.IgnoreRoutes
	opts.UserRoles = override.UserRoles

	return opts
}

func NewAuthService(override Options) *Auth {
	opts := getOptions(override)

	a := &Auth{
		db:           opts.DB,
		l:            opts.Logger,
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
		a.GoogleOauthConfig = oauth2.Config{
			ClientID:     opts.GoogleOauth.ClientID,
			ClientSecret: opts.GoogleOauth.ClientSecret,
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/userinfo.profile",
			},
			Endpoint: google.Endpoint,
		}
	}

	a.defaultMobileCountryCode = opts.DefaultMobileCountryCode

	return a
}

// RoleName returns the name of the role for the given UserRole
func (a *Auth) RoleName(role UserRole) string {
	return a.userRoles[role]
}

// addIgnoreRoute adds the given routes to the ignore list
// These routes do not require authentication
// This is useful for routes like login, register, etc.
func (a *Auth) AddIgnoreRoute(routes ...string) {
	a.ignoreRoutes = append(a.ignoreRoutes, routes...)
}

// FormatMobileNumber formats the mobile number to include the default country code
func (a *Auth) FormatMobileNumber(mobile string) string {
	if mobile == "" {
		return ""
	}

	if len(mobile) <= 10 {
		mobile = fmt.Sprintf("%s %s", a.defaultMobileCountryCode, mobile)
		return mobile
	}

	mobile = fmt.Sprintf("%s %s", mobile[:len(mobile)-10], mobile[len(mobile)-10:])
	return mobile
}

// GetUserRoles returns the map of user roles defined in the Auth.
// It does not take any input parameters and returns a map where the key is UserRole and the value is the role name.
func (a *Auth) GetUserRoles() map[UserRole]string {
	return a.userRoles
}
