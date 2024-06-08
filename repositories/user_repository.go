package repositories

import (
	"errors"
	"go-share/models"
	"go-share/utils"

	"gorm.io/gorm"
)

// UserRepository interacts with the users table in the database.
type UserRepository struct {
	DB *gorm.DB
}

// NewUserRepository creates a new UserRepository instance.
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{DB: db}
}

// CreateUser creates a new user in the database. 
func (ur *UserRepository) CreateUser(user *models.User) error {
	hashedPassword, err := utils.HashPassword(user.Password)
	if err != nil {
		return err // Return the specific hashing error
	}
	user.Password = string(hashedPassword)

	if err := ur.DB.Create(&user).Error; err != nil {
		return errors.New("error creating user in database")
	}

	return nil
}

// GetUserByEmail retrieves a user by their email address.
func (ur *UserRepository) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	if err := ur.DB.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, errors.New("user not found") 
	}

	return &user, nil
}