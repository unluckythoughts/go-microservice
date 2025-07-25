package auth

import (
	"fmt"
	"regexp"
	"time"

	"gorm.io/gorm"
)

// Mobile is the mobile number of the user
type Mobile string

func (m Mobile) IsValid() bool {
	// Remove any spaces, dashes, or parentheses
	re := regexp.MustCompile(`[\s\-\(\)]`)
	mobile := re.ReplaceAllString(string(m), "")

	mobileRegex := regexp.MustCompile(`^\+?[0-9]{10,15}$`)
	return mobileRegex.MatchString(mobile)
}

func (m Mobile) getCountryCode() (string, bool) {
	if m.IsValid() {
		return string(m[:len(m)-10]), true
	}
	return "", false
}

func (m Mobile) getNumber() (string, bool) {
	if m.IsValid() {
		return string(m[len(m)-10:]), true
	}
	return "", false
}

type UserRole uint

func (r *UserRole) Value() string {
	return fmt.Sprintf("%d", r)
}

var (
	userRoleDefault UserRole = 0
)

type User struct {
	gorm.Model
	Name           string    `gorm:"column:name;not null" json:"name"`
	Email          string    `gorm:"column:email;not null;uniqueIndex" json:"email"`
	EmailVerified  bool      `gorm:"column:email_verified;not null;default:false" json:"email_verified"`
	Mobile         string    `gorm:"column:mobile;uniqueIndex" json:"mobile,omitempty"`
	MobileVerified bool      `gorm:"column:mobile_verified;not null;default:false" json:"mobile_verified"`
	Password       string    `gorm:"column:password;not null" json:"-"`
	Role           UserRole  `gorm:"column:role;type:int;not null;default:1" json:"role"`
	VerifyToken    string    `gorm:"column:verify_token;not null" json:"-"`
	TokenExpiresAt time.Time `gorm:"column:token_expires_at" json:"-"`
	// Google OAuth fields
	GoogleID     string `gorm:"column:google_id" json:"-"`
	GoogleAvatar string `gorm:"column:google_avatar" json:"google_avatar,omitempty"`
}

type loginRequest struct {
	Email    string `json:"email" valid:"email~email is not valid"`
	Mobile   string `json:"mobile" valid:"mobile~mobile is not valid"`
	Password string `json:"password" valid:"required~password is required"`
}

// GoogleOAuthRequest represents the request for Google OAuth login
type googleOAuthRequest struct {
	Code        string `json:"code" valid:"required~authorization code is required"`
	RedirectURI string `json:"redirect_uri" valid:"required~redirect URI is required"`
}

// AppleOAuthRequest represents the request for Apple OAuth login
// type appleOAuthRequest struct {
// 	Code        string `json:"code" valid:"required~authorization code is required"`
// 	RedirectURI string `json:"redirect_uri" valid:"required~redirect URI is required"`
// 	IDToken     string `json:"id_token" valid:"required~ID token is required"`
// }

// GoogleUserInfo represents the user info from Google OAuth
type googleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Locale        string `json:"locale"`
}
