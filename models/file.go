package models

import (
	"gorm.io/gorm"
	"go-share/utils"
	"errors"
)

// File represents a shared file.
type File struct {
	gorm.Model
	Name        string `json:"name" validate:"required"`
	ContentType string `json:"content_type"`
	Path        string `json:"path" validate:"required"`
	Description string `json:"description"`
	UserID      uint   `json:"user_id" gorm:"index; not null"`
}

// CreateFile creates a new file record in the database, ensuring it's associated with the user. 
func (f *File) CreateFile(db *gorm.DB) error {
	if err := utils.ValidateStruct(f); err != nil {
		return err
	}

	if err := db.Create(&f).Error; err != nil {
		return errors.New("error creating file")
	}

	return nil
}

// UpdateFile updates a file record. It checks for authorization before updating.
func (f *File) UpdateFile(db *gorm.DB, userID uint, updatedFile *File) error {
	if f.UserID != userID {
		return errors.New("unauthorized to update file")
	}

    if updatedFile.Name != "" {
        f.Name = updatedFile.Name
    }
    if updatedFile.ContentType != "" {
        f.ContentType = updatedFile.ContentType
    }
    if updatedFile.Path != "" {
        f.Path = updatedFile.Path
    }
    if updatedFile.Description != "" {
        f.Description = updatedFile.Description
    }

	if err := db.Save(&f).Error; err != nil {
		return errors.New("error updating file")
	}

	return nil
}

// DeleteFile deletes a file, checking for authorization before deletion.
func (f *File) DeleteFile(db *gorm.DB, userID uint) error {
	if f.UserID != userID {
		return errors.New("unauthorized to delete file")
	}

	if err := db.Delete(&f).Error; err != nil {
		return errors.New("error deleting file") 
	}

	return nil
}