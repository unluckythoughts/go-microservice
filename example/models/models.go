package models

import (
	"regexp"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UserRole represents the role of the user
type UserRole string

// Mobile represents a mobile number
type Mobile string

const (
	UserRoleUser  UserRole = "user"
	UserRoleAdmin UserRole = "admin"
	UserRoleGuest UserRole = "guest"
)

// IsValid checks if the mobile number is valid
func (m Mobile) IsValid() bool {
	// Remove any spaces, dashes, or parentheses
	re := regexp.MustCompile(`[\s\-\(\)]`)
	mobile := re.ReplaceAllString(string(m), "")

	mobileRegex := regexp.MustCompile(`^\+?[0-9]{10,15}$`)
	return mobileRegex.MatchString(mobile)
}

// Hash generates a bcrypt hash for the given password
func Hash(password string) string {
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(hash)
}

// CheckPasswordHash compares a password with a hash
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// User represents a user in the system
type User struct {
	*gorm.Model
	Name         string    `gorm:"column:name;not null" json:"name" valid:"required~name is required"`
	Email        string    `gorm:"column:email;unique;not null" json:"email" valid:"required~email is required,email~email should be valid"`
	Password     string    `gorm:"column:password" json:"password,omitempty" valid:"required~password is required"`
	MobileNumber string    `gorm:"column:mobile_number" json:"mobile_number" valid:"optional,mobile~mobile number should be at least 10 numbers"`
	Role         UserRole  `gorm:"column:role;default:user;not null" json:"role"`
	IsVerified   bool      `gorm:"column:is_verified;default:false" json:"is_verified"`
	LastLogin    time.Time `gorm:"column:last_login" json:"last_login"`
	GoogleID     string    `gorm:"column:google_id" json:"-"`
	AppleID      string    `gorm:"column:apple_id" json:"-"`
}

// UserSettings represents user preferences and settings
type UserSettings struct {
	*gorm.Model
	UserID               uint   `gorm:"column:user_id;unique" json:"user_id"`
	User                 User   `gorm:"foreignKey:UserID" json:"-"`
	Theme                string `gorm:"column:theme;default:light" json:"theme"`
	Language             string `gorm:"column:language;default:en" json:"language"`
	NotificationsEnabled bool   `gorm:"column:notifications_enabled;default:true" json:"notifications_enabled"`
	TimeZone             string `gorm:"column:time_zone;default:UTC" json:"time_zone"`
}

// LoginRequest represents login request payload
type LoginRequest struct {
	Email    string `json:"email" valid:"email~email should be valid"`
	Mobile   string `json:"mobile_number" valid:"optional,mobile~mobile number should be valid"`
	Password string `json:"password" valid:"required~password is required"`
}

// SignupRequest represents signup request payload
type SignupRequest struct {
	Name         string `json:"name" valid:"required~name is required"`
	Email        string `json:"email" valid:"required~email is required,email~email should be valid"`
	Password     string `json:"password" valid:"required~password is required,length(8|50)~password should be between 8 and 50 characters"`
	MobileNumber string `json:"mobile_number" valid:"optional,mobile~mobile number should be valid"`
}
