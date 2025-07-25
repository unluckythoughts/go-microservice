package auth

import (
	"errors"
	"time"

	"github.com/unluckythoughts/go-microservice/utils"
	"gorm.io/gorm"
)

// CreateUser creates a new user with hashed password
func (a *Service) CreateUser(user *User) error {
	// Generate a unique random 16 character verify token
	token, err := utils.GenerateRandomString(16)
	if err != nil {
		return err
	}
	user.VerifyToken = token
	user.TokenExpiresAt = time.Now().Add(24 * time.Hour) // Token valid for 24 hours

	// Hash the password before saving
	if user.Password != "" {
		hashedPassword, err := utils.GetHash(user.Password)
		if err != nil {
			return err
		}
		user.Password = hashedPassword
	}
	return a.db.Create(user).Error
}

// GetUserByID retrieves a user by ID with their addresses
func (a *Service) GetUserByID(id uint) (*User, error) {
	var user User
	err := a.db.Preload("Addresses").First(&user, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// GetUserByEmail retrieves a user by email
func (a *Service) GetUserByEmail(email string) (*User, error) {
	var user User
	err := a.db.Preload("Addresses").Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// GetUserByMobile retrieves a user by mobile number
func (a *Service) GetUserByMobile(mobile string) (*User, error) {
	var user User
	err := a.db.Preload("Addresses").Where("mobile = ?", mobile).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// GetAllUsers retrieves all users with pagination
func (a *Service) GetAllUsers(offset, limit int) ([]User, error) {
	var users []User
	err := a.db.Preload("Addresses").Offset(offset).Limit(limit).Find(&users).Error
	return users, err
}

// UpdateUser updates an existing user with password hashing if needed
func (a *Service) UpdateUser(user *User) error {
	// Hash the password if it's being updated and not empty
	if user.Password != "" {
		hashedPassword, err := utils.GetHash(user.Password)
		if err != nil {
			return err
		}
		user.Password = hashedPassword
	}
	return a.db.Save(user).Error
}

// UpdateUserPartial updates specific fields of a user
func (a *Service) UpdateUserPartial(id uint, updates map[string]interface{}) error {
	result := a.db.Model(&User{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("user not found")
	}
	return nil
}

// DeleteUser soft deletes a user
func (a *Service) DeleteUser(id uint) error {
	result := a.db.Delete(&User{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("user not found")
	}
	return nil
}

// HardDeleteUser permanently deletes a user and all related data
func (a *Service) HardDeleteUser(id uint) error {
	// First delete related addresses
	if err := a.db.Unscoped().Where("user_id = ?", id).Error; err != nil {
		return err
	}

	// Then delete the user
	result := a.db.Unscoped().Delete(&User{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("user not found")
	}
	return nil
}

// CountUsers returns the total number of users
func (a *Service) CountUsers() (int64, error) {
	var count int64
	err := a.db.Model(&User{}).Count(&count).Error
	return count, err
}

// UserExists checks if a user exists by ID
func (a *Service) UserExists(id uint) (bool, error) {
	var count int64
	err := a.db.Model(&User{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

// EmailExists checks if an email is already taken
func (a *Service) EmailExists(email string) (bool, error) {
	var count int64
	err := a.db.Model(&User{}).Where("email = ?", email).Count(&count).Error
	return count > 0, err
}

func (a *Service) GetUserByGoogleID(googleID string) (*User, error) {
	var user User
	err := a.db.Where("google_id = ?", googleID).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// UpdateUserVerifyToken updates the verification token for a user
func (a *Service) UpdateUserVerifyToken(id uint, token string) error {
	updates := map[string]interface{}{
		"verify_token":     token,
		"token_expires_at": time.Now().Add(24 * time.Hour), // Token valid for 24 hours
	}
	result := a.db.Model(&User{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("user not found")
	}
	return nil
}

// ClearUserVerifyToken clears the verification token for a user
func (a *Service) ClearUserVerifyToken(id uint) error {
	result := a.db.Model(&User{}).Where("id = ?", id).Update("verify_token", "")
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("user not found")
	}
	return nil
}

// GetUserByVerifyToken retrieves a user by verification token
func (a *Service) VerifyUserToken(token string) (*User, error) {
	var user User
	err := a.db.Where("verify_token = ? AND verify_token != ''", token).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid or expired token")
		}
		return nil, err
	}

	if user.TokenExpiresAt.Before(time.Now()) {
		return nil, errors.New("token has expired")
	}

	return &user, nil
}

// Password-related functions
// UpdateUserPassword updates the password for a user
func (a *Service) UpdateUserPassword(userID uint, newPassword string) error {
	// Hash the new password
	hashedPassword, err := utils.GetHash(newPassword)
	if err != nil {
		return err
	}

	result := a.db.Model(&User{}).Where("id = ?", userID).Update("password", hashedPassword)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("user not found")
	}
	return nil
}

// VerifyUserPassword verifies a user's password
func (a *Service) VerifyUserPassword(userID uint, password string) (bool, error) {
	var user User
	err := a.db.Select("password").First(&user, userID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, errors.New("user not found")
		}
		return false, err
	}

	return utils.CompareValue(password, user.Password)
}

// VerifyUserPasswordByEmail verifies a user's password by email
func (a *Service) VerifyUserPasswordByEmail(email, password string) (*User, bool, error) {
	var user User
	err := a.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, false, errors.New("user not found")
		}
		return nil, false, err
	}

	isValid, err := utils.CompareValue(password, user.Password)
	if err != nil {
		return nil, false, err
	}

	return &user, isValid, nil
}

// VerifyUserPasswordByMobile verifies a user's password by mobile number
func (a *Service) VerifyUserPasswordByMobile(mobile, password string) (*User, bool, error) {
	var user User
	err := a.db.Where("mobile = ?", mobile).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, false, errors.New("user not found")
		}
		return nil, false, err
	}

	isValid, err := utils.CompareValue(password, user.Password)
	if err != nil {
		return nil, false, err
	}

	return &user, isValid, nil
}

// ChangeUserPassword changes a user's password after verifying the old password
func (a *Service) ChangeUserPassword(userID uint, oldPassword, newPassword string) error {
	// First verify the old password
	isValid, err := a.VerifyUserPassword(userID, oldPassword)
	if err != nil {
		return err
	}
	if !isValid {
		return errors.New("invalid current password")
	}

	// Update with new password
	return a.UpdateUserPassword(userID, newPassword)
}
