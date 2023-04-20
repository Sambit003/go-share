package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Email    string `gorm:"uniqueIndex" json:"email"`
	Password string `json:"password"`
}

func GetUser(email string) (User, error) {
	var user User
	err := (*gorm.DB).Where("email = ?", email).First(&user).Error
	return user, err
}

func ValidateUser(email, password string) (User, error) {
	var user User
	err := (*gorm.DB).Where("email = ? AND password = ?", email, password).First(&user).Error
	return user, err
}

func CreateUser(email, password string) (User, error) {
	var user User
	err := (*gorm.DB).Create(&User{Email: email, Password: password}).Error
	return user, err
}