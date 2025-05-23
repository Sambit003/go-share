package repositories

import (
	"errors"
	"go-share/pkg/files"

	"gorm.io/gorm"
)

// FileRepository handles interactions with the files table in the database.
type FileRepository struct {
	DB *gorm.DB
}

// NewFileRepository creates a new FileRepository.
func NewFileRepository(db *gorm.DB) *FileRepository {
	return &FileRepository{DB: db}
}

// CreateFile creates a new file associated with a user.
func (fr *FileRepository) CreateFile(file *files.File) error {
	if err := fr.DB.Create(&file).Error; err != nil {
		return errors.New("error creating file in database") 
	}

	return nil
}

// GetFiles retrieves all files (for now - pagination/filtering should be added).
func (fr *FileRepository) GetFiles() ([]files.File, error) {
	var files []files.File
	if err := fr.DB.Find(&files).Error; err != nil {
		return nil, errors.New("error retrieving files from database") 
	}
	return files, nil
}

// GetFile retrieves a file by its ID.
func (fr *FileRepository) GetFile(fileID uint) (*files.File, error) {
	var file files.File
	if err := fr.DB.First(&file, fileID).Error; err != nil {
		return nil, errors.New("file not found") 
	}

	return &file, nil
}

// UpdateFile updates a file's information.
func (fr *FileRepository) UpdateFile(file *files.File) error {
	if err := fr.DB.Save(&file).Error; err != nil {
		return errors.New("error updating file in database") 
	}

	return nil
}

// DeleteFile deletes a file by its ID.
func (fr *FileRepository) DeleteFile(fileID uint) error {
	if err := fr.DB.Delete(&files.File{}, fileID).Error; err != nil {
		return errors.New("error deleting file from database") 
	}

	return nil
}