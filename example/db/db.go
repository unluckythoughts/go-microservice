package db

import (
	"example/models"

	"gorm.io/gorm"
)

type DB struct {
	gorm *gorm.DB
}

func NewDB(database *gorm.DB) *DB {
	return &DB{
		gorm: database,
	}
}

// User operations
func (db *DB) CreateUser(user *models.User) error {
	user.Password = models.Hash(user.Password)
	return db.gorm.Create(user).Error
}

func (db *DB) GetUserByID(id uint) (*models.User, error) {
	var user models.User
	err := db.gorm.First(&user, id).Error
	return &user, err
}

func (db *DB) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	err := db.gorm.Where("email = ?", email).First(&user).Error
	return &user, err
}

func (db *DB) GetUserByMobileNumber(mobileNumber string) (*models.User, error) {
	var user models.User
	err := db.gorm.Where("mobile_number = ?", mobileNumber).First(&user).Error
	return &user, err
}

func (db *DB) UpdateUser(user *models.User) error {
	return db.gorm.Save(user).Error
}

func (db *DB) DeleteUser(id uint) error {
	return db.gorm.Delete(&models.User{}, id).Error
}

// UserSettings operations
func (db *DB) CreateUserSettings(settings *models.UserSettings) error {
	return db.gorm.Create(settings).Error
}

func (db *DB) GetUserSettingsByUserID(userID uint) (*models.UserSettings, error) {
	var settings models.UserSettings
	err := db.gorm.Where("user_id = ?", userID).First(&settings).Error
	return &settings, err
}

func (db *DB) UpdateUserSettings(settings *models.UserSettings) error {
	return db.gorm.Save(settings).Error
}
