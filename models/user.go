package models

import (
	"gorm.io/gorm"
)

var db *gorm.DB
type User struct {
	gorm.Model
	Email    string `gorm:"uniqueIndex" json:"email"`
	Password string `json:"password"`
}

func CreateUser(user *User) error {
	return db.Create(user).Error
}

func ValidateUser(user *User) (*User, error) {
	var foundUser User
	err := db.Where("email = ?", user.Email).First(&foundUser).Error
	if err != nil {
		return nil, err
	}
	return &foundUser, nil
}

func GetUser(id uint) (*User, error) {
	var user User
	err := db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

