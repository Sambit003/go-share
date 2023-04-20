package models

import "gorm.io/gorm"

type File struct {
	gorm.Model
	ID 	 	 	uint   `json:"id" gorm:"primaryKey"`
	Name     	string `json:"name" validate:"required"` 
	ContType 	string `json:"cont_type"`
	Content  	[]byte `json:"content"`
	Path     	string `json:"path"`
	Description string `json:"description" validate:"required"`
	UserID   	uint   `json:"user_id" gorm:"index; not null"`
}