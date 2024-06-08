package models

import (
	"errors"
	"go-share/utils"
	"gorm.io/gorm"
)

// User represents a user in the system.
type User struct {
	gorm.Model
	Email    string `gorm:"uniqueIndex" json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// CreateUser creates a new user with a hashed password.
func (u *User) CreateUser(db *gorm.DB) error {
	hashedPassword, err := utils.HashPassword(u.Password)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)

	if err := db.Create(&u).Error; err != nil {
		return errors.New("error creating user")
	}
	return nil
}

// ValidateUserCredentials checks if the provided email and password match an existing user.
func (u *User) ValidateUserCredentials(db *gorm.DB) (*User, error) {
	var foundUser User
	if err := db.Where("email = ?", u.Email).First(&foundUser).Error; err != nil {
		return nil, errors.New("user not found")
	}

	if err := utils.ComparePassword(u.Password, foundUser.Password); err != nil {
		return nil, errors.New("invalid password")
	}

	return &foundUser, nil
}