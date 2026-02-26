package auth

import (
	"errors"
	"time"

	"github.com/unluckythoughts/go-microservice/v2/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (s *Service) CreateVerifyToken(target string) (string, error) {
	var token string
	var err error

	for {
		token, err = utils.GenerateRandomString(8)
		if err != nil {
			return "", err
		}

		var v Verify
		if err := s.db.Where("token = ?", token).First(&v).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// No existing token for this target, we can proceed to create a new one
				break
			}

			return "", err
		}
	}

	verify := &Verify{
		Target:    target,
		Token:     token,
		Verified:  false,
		ExpiresAt: time.Now().Add(time.Hour),
	}

	err = s.db.
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "target"}},
			DoUpdates: clause.AssignmentColumns([]string{"token", "verified", "expires_at"}),
		}).
		Create(verify).Error
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *Service) GetVerification(token string) (*Verify, error) {
	var verify Verify
	err := s.db.Where("token = ?", token).First(&verify).Error
	if err != nil {
		return nil, err
	}

	return &verify, nil
}

func (s *Service) IsVerified(target string) bool {
	var verify Verify
	err := s.db.Where("target = ?", target).First(&verify).Error
	if err != nil {
		return false
	}

	return verify.Verified
}

func (s *Service) VerifyToken(target string, token string) (bool, error) {
	var verify Verify
	err := s.db.
		Where("target = ? AND token = ?", target, token).
		First(&verify).Error
	if err != nil {
		return false, err
	}

	if verify.ExpiresAt.Before(time.Now()) {
		return false, ErrExpiredToken
	}

	verify.Verified = true
	err = s.db.Save(&verify).Error
	if err != nil {
		return false, err
	}

	return true, nil
}

// CreateUser creates a new user with hashed password
func (s *Service) CreateUser(user *User) error {
	// Hash the password before saving
	if user.Password.String() != "" {
		hashedPassword, err := utils.GetHash(user.Password.String())
		if err != nil {
			return err
		}
		user.Password = Password(hashedPassword)
	}
	return s.db.Create(user).Error
}

// GetUserByID retrieves a user by ID with their addresses
func (s *Service) GetUserByID(id uint) (*User, error) {
	var user User
	err := s.db.Preload("Addresses").First(&user, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	utils.ClearValues(&user, "Password", "GoogleID")
	return &user, nil
}

// GetUserByEmail retrieves a user by email
func (s *Service) GetUserByEmail(email string) (*User, error) {
	var user User
	err := s.db.Preload("Addresses").Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	utils.ClearValues(&user, "Password", "GoogleID")
	return &user, nil
}

// GetUserByMobile retrieves a user by mobile number
func (s *Service) GetUserByMobile(mobile string) (*User, error) {
	var user User
	err := s.db.Preload("Addresses").Where("mobile = ?", mobile).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	utils.ClearValues(&user, "Password", "GoogleID")
	return &user, nil
}

// GetAllUsers retrieves all users with pagination
func (s *Service) GetAllUsers(offset, limit int) ([]User, error) {
	var users []User
	err := s.db.Preload("Addresses").Offset(offset).Limit(limit).Find(&users).Error
	utils.ClearValues(&users, "Password", "GoogleID")
	return users, err
}

// UpdateUser updates an existing user with password hashing if needed
func (s *Service) UpdateUser(user *User) error {
	// Hash the password if it's being updated and not empty
	if user.Password.String() != "" {
		hashedPassword, err := utils.GetHash(user.Password.String())
		if err != nil {
			return err
		}
		user.Password = Password(hashedPassword)
	}
	return s.db.Save(user).Error
}

// UpdateEmailVerified updates the email_verified field for a user
func (s *Service) UpdateEmailVerified(id uint, verified bool) error {
	result := s.db.Model(&User{}).Where("id = ?", id).Update("email_verified", verified)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("user not found")
	}
	return nil
}

// UpdateMobileVerified updates the mobile_verified field for a user
func (s *Service) UpdateMobileVerified(id uint, verified bool) error {
	result := s.db.Model(&User{}).Where("id = ?", id).Update("mobile_verified", verified)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("user not found")
	}
	return nil
}

// UpdateUserPartial updates specific fields of a user
func (s *Service) UpdateUserPartial(id uint, updates any) error {
	filteredUpdates := make(map[string]any)
	utils.FilterDBUpdates(updates, &filteredUpdates, "Password", "GoogleID")
	result := s.db.Model(&User{}).Where("id = ?", id).Updates(filteredUpdates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("user not found")
	}
	return nil
}

// DeleteUser soft deletes a user
func (s *Service) DeleteUser(id uint) error {
	result := s.db.Delete(&User{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("user not found")
	}
	return nil
}

// HardDeleteUser permanently deletes a user and all related data
func (s *Service) HardDeleteUser(id uint) error {
	// First delete related addresses
	if err := s.db.Unscoped().Where("user_id = ?", id).Error; err != nil {
		return err
	}

	// Then delete the user
	result := s.db.Unscoped().Delete(&User{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("user not found")
	}
	return nil
}

// CountUsers returns the total number of users
func (s *Service) CountUsers() (int64, error) {
	var count int64
	err := s.db.Model(&User{}).Count(&count).Error
	return count, err
}

// UserExists checks if a user exists by ID
func (s *Service) UserExists(id uint) (bool, error) {
	var count int64
	err := s.db.Model(&User{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

// EmailExists checks if an email is already taken
func (s *Service) EmailExists(email string) (bool, error) {
	var count int64
	err := s.db.Model(&User{}).Where("email = ?", email).Count(&count).Error
	return count > 0, err
}

func (s *Service) GetUserByGoogleID(googleID string) (*User, error) {
	var user User
	err := s.db.Where("google_id = ?", googleID).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	utils.ClearValues(&user, "Password", "GoogleID")
	return &user, nil
}

// Password-related functions
// UpdateUserPassword updates the password for a user
func (s *Service) UpdateUserPassword(userID uint, newPassword Password) error {
	// Hash the new password
	hashedPassword, err := utils.GetHash(newPassword.String())
	if err != nil {
		return err
	}

	result := s.db.Model(&User{}).Where("id = ?", userID).Update("password", hashedPassword)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("user not found")
	}
	return nil
}

// VerifyUserPassword verifies a user's password
func (s *Service) VerifyUserPassword(userID uint, password Password) (bool, error) {
	var user User
	err := s.db.Select("password").First(&user, userID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, errors.New("user not found")
		}
		return false, err
	}

	return utils.CompareValue(password.String(), user.Password.String())
}

// VerifyUserPasswordByEmail verifies a user's password by email
func (s *Service) VerifyUserPasswordByEmail(email string, password Password) (*User, bool, error) {
	var user User
	err := s.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, false, errors.New("user not found")
		}
		return nil, false, err
	}

	isValid, err := utils.CompareValue(password.String(), user.Password.String())
	if err != nil {
		return nil, false, err
	}

	utils.ClearValues(&user, "Password", "VerifyToken", "TokenExpiresAt", "GoogleID")
	return &user, isValid, nil
}

// VerifyUserPasswordByMobile verifies a user's password by mobile number
func (s *Service) VerifyUserPasswordByMobile(mobile string, password Password) (*User, bool, error) {
	var user User
	err := s.db.Where("mobile = ?", mobile).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, false, errors.New("user not found")
		}
		return nil, false, err
	}

	isValid, err := utils.CompareValue(password.String(), user.Password.String())
	if err != nil {
		return nil, false, err
	}

	utils.ClearValues(&user, "Password", "VerifyToken", "TokenExpiresAt", "GoogleID")
	return &user, isValid, nil
}

// ChangeUserPassword changes a user's password after verifying the old password
func (s *Service) ChangeUserPassword(userID uint, oldPassword, newPassword Password) error {
	// First verify the old password
	isValid, err := s.VerifyUserPassword(userID, oldPassword)
	if err != nil {
		return err
	}
	if !isValid {
		return errors.New("invalid current password")
	}

	// Update with new password
	return s.UpdateUserPassword(userID, newPassword)
}
