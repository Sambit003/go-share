// Package files provides functionalities for file management,
// including uploading, downloading, and encryption of files.
// It defines the core File model and its interactions with the database.
package files

import (
	"gorm.io/gorm"
	"go-share/utils"
	"errors"
)

// File represents the metadata of a shared file within the system.
// It includes details about the file's name, path, type, and ownership.
type File struct {
	gorm.Model           // GORM's base model (ID, CreatedAt, UpdatedAt, DeletedAt)
	Name        string `json:"name" validate:"required"` // Name of the file.
	ContentType string `json:"content_type"`             // MIME type of the file.
	Path        string `json:"path" validate:"required"` // Absolute path to the file on the server's filesystem.
	Description string `json:"description"`              // Optional description of the file.
	UserID      uint   `json:"user_id" gorm:"index; not null"` // ID of the user who owns this file.
	IsEncrypted bool   `json:"is_encrypted"`             // Flag indicating whether the file content is encrypted.
}

// CreateFile persists a new file record to the database.
// It first validates the File struct based on its validation tags.
// This method is typically called after a file has been successfully uploaded and its metadata populated.
func (f *File) CreateFile(db *gorm.DB) error {
	if err := utils.ValidateStruct(f); err != nil {
		return err
	}

	if err := db.Create(&f).Error; err != nil {
		return errors.New("error creating file")
	}

	return nil
}

// UpdateFile modifies an existing file record in the database.
// It first checks if the provided userID matches the UserID of the file,
// ensuring that only the owner can update the file information.
// Fields in updatedFile that are non-empty will be used to update the current file.
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
    // Path update might need careful consideration regarding the actual file on disk.
    if updatedFile.Path != "" {
        f.Path = updatedFile.Path
    }
    if updatedFile.Description != "" {
        f.Description = updatedFile.Description
    }
    // IsEncrypted status is typically set during upload and not directly updatable here.

	if err := db.Save(&f).Error; err != nil {
		return errors.New("error updating file")
	}

	return nil
}

// DeleteFile removes a file record from the database.
// It first checks if the provided userID matches the UserID of the file,
// ensuring that only the owner can delete the file.
// Note: This method only deletes the database record. The actual file on the
// filesystem is not removed by this method and should be handled separately if needed.
func (f *File) DeleteFile(db *gorm.DB, userID uint) error {
	if f.UserID != userID {
		return errors.New("unauthorized to delete file")
	}

	if err := db.Delete(&f).Error; err != nil {
		return errors.New("error deleting file") 
	}

	return nil
}