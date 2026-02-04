package auth

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/unluckythoughts/go-microservice/v2/tools/web"
	"github.com/unluckythoughts/go-microservice/v2/utils"
	"gorm.io/gorm"
)

// CreateUser creates a new user with hashed password
func (a *Auth) CreateUser(user *User) error {
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
func (a *Auth) GetUserByID(id uint) (*User, error) {
	var user User
	err := a.db.Preload("Addresses").First(&user, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	utils.ClearValues(&user, "Password", "VerifyToken", "TokenExpiresAt", "GoogleID")
	return &user, nil
}

// GetUserByEmail retrieves a user by email
func (a *Auth) GetUserByEmail(email string) (*User, error) {
	var user User
	err := a.db.Preload("Addresses").Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	utils.ClearValues(&user, "Password", "VerifyToken", "TokenExpiresAt", "GoogleID")
	return &user, nil
}

// GetUserByMobile retrieves a user by mobile number
func (a *Auth) GetUserByMobile(mobile string) (*User, error) {
	var user User
	err := a.db.Preload("Addresses").Where("mobile = ?", mobile).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	utils.ClearValues(&user, "Password", "VerifyToken", "TokenExpiresAt", "GoogleID")
	return &user, nil
}

// GetAllUsers retrieves all users with pagination
func (a *Auth) GetAllUsers(offset, limit int) ([]User, error) {
	var users []User
	err := a.db.Preload("Addresses").Offset(offset).Limit(limit).Find(&users).Error
	utils.ClearValues(&users, "Password", "VerifyToken", "TokenExpiresAt", "GoogleID")
	return users, err
}

// UpdateUser updates an existing user with password hashing if needed
func (a *Auth) UpdateUser(user *User) error {
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

// UpdateEmailVerified updates the email_verified field for a user
func (a *Auth) UpdateEmailVerified(id uint, verified bool) error {
	result := a.db.Model(&User{}).Where("id = ?", id).Update("email_verified", verified)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("user not found")
	}
	return nil
}

// UpdateMobileVerified updates the mobile_verified field for a user
func (a *Auth) UpdateMobileVerified(id uint, verified bool) error {
	result := a.db.Model(&User{}).Where("id = ?", id).Update("mobile_verified", verified)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("user not found")
	}
	return nil
}

// UpdateUserPartial updates specific fields of a user
func (a *Auth) UpdateUserPartial(id uint, updates any) error {
	filteredUpdates := make(map[string]any)
	utils.FilterDBUpdates(updates, &filteredUpdates, "Password", "VerifyToken", "TokenExpiresAt", "GoogleID")
	result := a.db.Model(&User{}).Where("id = ?", id).Updates(filteredUpdates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("user not found")
	}
	return nil
}

// DeleteUser soft deletes a user
func (a *Auth) DeleteUser(id uint) error {
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
func (a *Auth) HardDeleteUser(id uint) error {
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
func (a *Auth) CountUsers() (int64, error) {
	var count int64
	err := a.db.Model(&User{}).Count(&count).Error
	return count, err
}

// UserExists checks if a user exists by ID
func (a *Auth) UserExists(id uint) (bool, error) {
	var count int64
	err := a.db.Model(&User{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

// EmailExists checks if an email is already taken
func (a *Auth) EmailExists(email string) (bool, error) {
	var count int64
	err := a.db.Model(&User{}).Where("email = ?", email).Count(&count).Error
	return count > 0, err
}

func (a *Auth) GetUserByGoogleID(googleID string) (*User, error) {
	var user User
	err := a.db.Where("google_id = ?", googleID).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	utils.ClearValues(&user, "Password", "VerifyToken", "TokenExpiresAt", "GoogleID")
	return &user, nil
}

// UpdateUserVerifyToken updates the verification token for a user
func (a *Auth) UpdateUserVerifyToken(id uint, token string) error {
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
func (a *Auth) ClearUserVerifyToken(id uint) error {
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
func (a *Auth) VerifyUserToken(token string) (*User, error) {
	var user User
	err := a.db.Where("verify_token = ? AND verify_token != ''", token).First(&user).Error
	if err != nil {
		return nil, web.NewError(http.StatusBadRequest, fmt.Errorf("invalid verify token: %w", err))
	}

	if user.TokenExpiresAt.Before(time.Now()) {
		return nil, web.NewError(http.StatusBadRequest, fmt.Errorf("token has expired"))
	}

	utils.ClearValues(&user, "Password", "VerifyToken", "TokenExpiresAt", "GoogleID")
	return &user, nil
}

// Password-related functions
// UpdateUserPassword updates the password for a user
func (a *Auth) UpdateUserPassword(userID uint, newPassword string) error {
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
func (a *Auth) VerifyUserPassword(userID uint, password string) (bool, error) {
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
func (a *Auth) VerifyUserPasswordByEmail(email, password string) (*User, bool, error) {
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

	utils.ClearValues(&user, "Password", "VerifyToken", "TokenExpiresAt", "GoogleID")
	return &user, isValid, nil
}

// VerifyUserPasswordByMobile verifies a user's password by mobile number
func (a *Auth) VerifyUserPasswordByMobile(mobile, password string) (*User, bool, error) {
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

	utils.ClearValues(&user, "Password", "VerifyToken", "TokenExpiresAt", "GoogleID")
	return &user, isValid, nil
}

// ChangeUserPassword changes a user's password after verifying the old password
func (a *Auth) ChangeUserPassword(userID uint, oldPassword, newPassword string) error {
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
