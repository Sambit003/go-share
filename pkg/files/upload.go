package files

import (
	"io"
	"os"
	"path/filepath"
	"strconv"
	"errors" // Added errors package

	// "go-share/repositories" // Removed to break import cycle
	"go-share/utils"
	"gorm.io/gorm"
)

// UploadFile manages the process of saving a new file to the system.
// It performs the following steps:
// 1. Constructs a unique path for the file within the storagePathBase, under a user-specific directory.
// 2. Ensures the target directory exists, creating it if necessary.
// 3. Writes the provided fileContent to the destination file on the filesystem.
// 4. If an encryptionKey is provided (and is of valid AES key length: 16, 24, or 32 bytes),
//    it encrypts the newly saved file using AES-GCM via the EncryptFile function.
// 5. Creates a File metadata record (including Name, ContentType, Path, Description, UserID, and IsEncrypted status).
// 6. Validates the metadata.
// 7. Saves the metadata record to the database using the File model's CreateFile method.
//
// Parameters:
//   - db: A *gorm.DB instance for database interactions.
//   - fileContent: An io.Reader from which the file's content will be read.
//   - fileName: The desired name for the file.
//   - contentType: The MIME type of the file (e.g., "image/jpeg", "text/plain").
//   - description: An optional description for the file.
//   - userID: The ID of the user uploading the file. This is used for associating the file and for namespacing the storage path.
//   - storagePathBase: The base directory on the server where files will be stored (e.g., "./uploads").
//   - encryptionKey: An optional byte slice representing the AES encryption key. If provided and valid, the file will be encrypted.
//     Key management (generation, storage, retrieval) is outside the scope of this function.
//
// Returns:
//   - A pointer to the newly created File metadata object if successful.
//   - An error if any step in the process fails (e.g., directory creation, file writing, encryption, database save).
//     Specific errors can indicate invalid encryption key length or encryption failure.
func UploadFile(db *gorm.DB, fileContent io.Reader, fileName string, contentType string, description string, userID uint, storagePathBase string, encryptionKey []byte) (*File, error) {
	// Construct the full file path
	filePath := filepath.Join(storagePathBase, "user_"+strconv.Itoa(int(userID)), fileName)

	// Ensure the directory exists
	if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
		return nil, err
	}

	// Create the destination file
	dst, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}
	defer dst.Close()

	// Copy the fileContent to dst
	if _, err := io.Copy(dst, fileContent); err != nil {
		return nil, err
	}
	dst.Close() // Close the file before encryption (if any)

	isEncrypted := false
	if len(encryptionKey) > 0 {
		// Basic key length check (can be more sophisticated)
		if len(encryptionKey) != 16 && len(encryptionKey) != 24 && len(encryptionKey) != 32 {
			// Optionally, remove the file if encryption key is invalid and we don't want to store unencrypted
			// os.Remove(filePath) 
			return nil, errors.New("invalid encryption key length: must be 16, 24, or 32 bytes")
		}
		if err := EncryptFile(filePath, encryptionKey); err != nil {
			// Optionally, remove the file if encryption failed
			// os.Remove(filePath)
			return nil, errors.New("failed to encrypt file: " + err.Error())
		}
		isEncrypted = true
	}

	// Create a File model instance
	fileMetadata := &File{
		Name:        fileName,
		ContentType: contentType,
		Path:        filePath, // Store the actual path
		Description: description,
		UserID:      userID,
		IsEncrypted: isEncrypted, // Set the IsEncrypted flag
	}

	// Validate the fileMetadata struct
	if err := utils.ValidateStruct(fileMetadata); err != nil {
		return nil, err
	}

	// Save the fileMetadata to the database using the model's method
	if err := fileMetadata.CreateFile(db); err != nil {
		return nil, err
	}

	return fileMetadata, nil
}
