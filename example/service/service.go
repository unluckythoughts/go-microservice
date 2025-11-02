package service

import (
	"errors"
	"os"
	"time"

	"example/db"
	"example/models"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

// Service errors
var (
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type Service struct {
	db        *db.DB
	jwtSecret string
}

func NewService(database *gorm.DB) *Service {
	d := db.NewDB(database)

	// Get JWT secret from environment
	jwtSecret := os.Getenv("SESSION_SECRET")
	if jwtSecret == "" {
		panic("JWT session secret is not provided!")
	}

	return &Service{
		db:        d,
		jwtSecret: jwtSecret,
	}
}

// GetUserByID retrieves a user by ID
func (s *Service) GetUserByID(userID uint) (*models.User, error) {
	user, err := s.db.GetUserByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	// Clear password before returning
	user.Password = ""
	return user, nil
}

// GetUserByEmail retrieves a user by email
func (s *Service) GetUserByEmail(email string) (*models.User, error) {
	user, err := s.db.GetUserByEmail(email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

// GetUserByMobileNumber retrieves a user by mobile number
func (s *Service) GetUserByMobileNumber(mobileNumber string) (*models.User, error) {
	user, err := s.db.GetUserByMobileNumber(mobileNumber)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

// CreateAuthResponse creates authentication response with token
func (s *Service) CreateAuthResponse(r interface{}, user *models.User) (any, error) {
	// Generate JWT token
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"role":    user.Role,
		"exp":     time.Now().Add(time.Hour * 24 * 7).Unix(), // 7 days
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return nil, err
	}

	// Clear password before returning
	user.Password = ""

	return map[string]interface{}{
		"token": tokenString,
		"user":  user,
	}, nil
}

// CreateUser creates a new user
func (s *Service) CreateUser(user *models.User) error {
	return s.db.CreateUser(user)
}

// CreateUserSettings creates user settings
func (s *Service) CreateUserSettings(settings *models.UserSettings) error {
	return s.db.CreateUserSettings(settings)
}

// Login authenticates a user and returns a JWT token
func (s *Service) Login(email, mobile, password string) (string, *models.User, error) {
	if email == "" && mobile == "" {
		return "", nil, ErrInvalidCredentials
	}

	var user *models.User
	var err error

	// Try to find user by email or mobile
	if email != "" {
		user, err = s.db.GetUserByEmail(email)
	} else {
		user, err = s.db.GetUserByMobileNumber(mobile)
	}

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil, ErrUserNotFound
		}
		return "", nil, err
	}

	// Check password
	if !models.CheckPasswordHash(password, user.Password) {
		return "", nil, ErrInvalidCredentials
	}

	// Update last login
	user.LastLogin = time.Now()
	if err := s.db.UpdateUser(user); err != nil {
		// Log error but don't fail the login
		// In production, you might want to handle this differently
	}

	// Generate JWT token
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"role":    user.Role,
		"exp":     time.Now().Add(time.Hour * 24 * 7).Unix(), // 7 days
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", nil, err
	}

	// Clear password before returning
	user.Password = ""

	return tokenString, user, nil
}

// Signup creates a new user account
func (s *Service) Signup(name, email, password, mobileNumber string) (*models.User, error) {
	// Check if user already exists
	existingUser, _ := s.db.GetUserByEmail(email)
	if existingUser != nil {
		return nil, ErrUserAlreadyExists
	}

	// Create new user
	user := &models.User{
		Name:         name,
		Email:        email,
		Password:     password, // Will be hashed in DB layer
		MobileNumber: mobileNumber,
		Role:         models.UserRoleUser,
		IsVerified:   false,
	}

	if err := s.db.CreateUser(user); err != nil {
		return nil, err
	}

	// Create default user settings
	settings := &models.UserSettings{
		UserID:               user.ID,
		Theme:                "light",
		Language:             "en",
		NotificationsEnabled: true,
		TimeZone:             "UTC",
	}
	s.db.CreateUserSettings(settings)

	// Clear password before returning
	user.Password = ""

	return user, nil
}

// UpdateUser updates user information
func (s *Service) UpdateUser(user *models.User) error {
	return s.db.UpdateUser(user)
}

// DeleteUser deletes a user account
func (s *Service) DeleteUser(userID uint) error {
	return s.db.DeleteUser(userID)
}

// GetUserSettings retrieves user settings, creating defaults if not found
func (s *Service) GetUserSettings(userID uint) (*models.UserSettings, error) {
	settings, err := s.db.GetUserSettingsByUserID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create default settings if not found
			settings = &models.UserSettings{
				UserID:               userID,
				Theme:                "light",
				Language:             "en",
				NotificationsEnabled: true,
				TimeZone:             "UTC",
			}
			if err := s.db.CreateUserSettings(settings); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	return settings, nil
}

// UpdateUserSettings updates user settings
func (s *Service) UpdateUserSettings(userID uint, theme, language string, notificationsEnabled bool, timeZone string) (*models.UserSettings, error) {
	settings, err := s.db.GetUserSettingsByUserID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create new settings
			settings = &models.UserSettings{
				UserID: userID,
			}
		} else {
			return nil, err
		}
	}

	// Update settings fields
	settings.Theme = theme
	settings.Language = language
	settings.NotificationsEnabled = notificationsEnabled
	settings.TimeZone = timeZone

	if err := s.db.UpdateUserSettings(settings); err != nil {
		return nil, err
	}

	return settings, nil
}
