package repositories

import (
	"errors"
	"go-share/models"
	"go-share/utils"
	"gorm.io/gorm"
)

type UserRepository struct{
	DB *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{DB: db}
}

func (ur UserRepository) CreateUser(user *models.User) error {
	hashedPassword, err := utils.HashPassword(user.Password)
	if err != nil {
		return errors.New("error in password hashing")
	}
	user.Password = string(hashedPassword)

	if err := ur.DB.Create(&user).Error; err != nil {
		return errors.New("error in creating user")
	}

	return nil
}

func (ur UserRepository) ValidateUser(user *models.User) (*models.User, error) {
	var foundUser models.User

	if err := ur.DB.Where("email = ?", user.Email).First(&foundUser).Error; err != nil {
		return nil, errors.New("user not found")
	}

	if err := utils.ComparePassword(user.Password, foundUser.Password); err {
		return nil, errors.New("invalid password")
	}

	return &foundUser, nil
}

func (ur UserRepository) GetUserById(id string) (*models.User, error) {
	var user models.User

	if err := ur.DB.Where("id = ?", id).First(&user).Error; err != nil {
		return nil, errors.New("user not found")
	}

	return &user, nil
}

func (ur UserRepository) GetUserByEmail(email string) (*models.User, error) {
	var user models.User

	if err := ur.DB.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, errors.New("user not found")
	}

	return &user, nil
}