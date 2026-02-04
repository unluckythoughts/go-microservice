package auth

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Name           string `gorm:"column:name;not null" json:"name"`
	Email          string `gorm:"column:email;not null;uniqueIndex" json:"email"`
	EmailVerified  bool   `gorm:"column:email_verified;not null;default:false" json:"email_verified"`
	Mobile         Mobile `gorm:"column:mobile;uniqueIndex" json:"mobile,omitempty"`
	MobileVerified bool   `gorm:"column:mobile_verified;not null;default:false" json:"mobile_verified"`
	Password       string `gorm:"column:password;not null" json:"-"`
	Role           Role   `gorm:"column:role;type:int;not null;default:1" json:"role"`
	// Google OAuth fields
	GoogleID     string `gorm:"column:google_id" json:"-"`
	GoogleAvatar string `gorm:"column:google_avatar" json:"google_avatar,omitempty"`
}

var ErrExpiredToken = fmt.Errorf("verification token has expired")

type Verify struct {
	gorm.Model
	Target    string    `gorm:"column:target;not null;uniqueIndex" json:"-"`
	Token     string    `gorm:"column:token;not null;uniqueIndex" json:"-"`
	Verified  bool      `gorm:"column:verified;not null;default:false" json:"-"`
	ExpiresAt time.Time `gorm:"column:expires_at;not null" json:"-"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type Credentials struct {
	Email    string `json:"email" valid:"email~email is not valid"`
	Mobile   string `json:"mobile" valid:"mobile~mobile is not valid"`
	Password string `json:"password" valid:"required~password is required"`
}

type RegisterRequest struct {
	Email    string `json:"email" valid:"email~email is not valid"`
	Mobile   string `json:"mobile" valid:"mobile~mobile is not valid"`
	Password string `json:"password" valid:"required~password is required,length(10|64)~password must be between 10 and 64 characters,matches(^(?=.*[a-z])(?=.*[A-Z])(?=.*\\d)(?=.*[@$!%*?&#])[A-Za-z\\d@$!%*?&#]+$)~password must contain at least one uppercase letter, one lowercase letter, one digit, and one special character"`
	Name     string `json:"name" valid:"required~name is required"`
}

type UpdateUserRequest struct {
	Name   string `json:"name" valid:"required~name is required"`
	Email  string `json:"email" valid:"email~email is not valid"`
	Mobile string `json:"mobile" valid:"mobile~mobile is not valid"`
}

type UpdatePasswordRequest struct {
	VerifyToken string `json:"verify_token" valid:"required~verification token is required"`
	NewPassword string `json:"new_password" valid:"required~new password is required,length(10|64)~password must be between 10 and 64 characters,matches(^(?=.*[a-z])(?=.*[A-Z])(?=.*\\d)(?=.*[@$!%*?&#])[A-Za-z\\d@$!%*?&#]+$)~password must contain at least one uppercase letter, one lowercase letter, one digit, and one special character"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" valid:"required~old password is required"`
	NewPassword string `json:"new_password" valid:"required~new password is required,length(10|64)~password must be between 10 and 64 characters,matches(^(?=.*[a-z])(?=.*[A-Z])(?=.*\\d)(?=.*[@$!%*?&#])[A-Za-z\\d@$!%*?&#]+$)~password must contain at least one uppercase letter, one lowercase letter, one digit, and one special character"`
}

type SendVerificationRequest struct {
	Email  string `json:"email" valid:"email~email is not valid"`
	Mobile string `json:"mobile" valid:"mobile~mobile is not valid"`
}

// GoogleOAuthRequest represents the request for Google OAuth login
type GoogleOAuthRequest struct {
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
