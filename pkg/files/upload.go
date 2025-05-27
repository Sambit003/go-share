package files

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"

	"go-share/utils"

	"gorm.io/gorm"
)

// UploadFile manages the process of saving a new file to the system.
// It performs the following steps:
//  1. Constructs a unique path for the file within the storagePathBase, under a user-specific directory.
//  2. Ensures the target directory exists, creating it if necessary.
//  3. Writes the provided fileContent to the destination file on the filesystem.
//  4. If an encryptionKey is provided (and is of valid AES key length: 16, 24, or 32 bytes),
//     it encrypts the newly saved file using AES-GCM via the EncryptFile function.
//  5. Creates a File metadata record (including Name, ContentType, Path, Description, UserID, and IsEncrypted status).
//  6. Validates the metadata.
//  7. Saves the metadata record to the database using the File model's CreateFile method.
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
	// Sanitize fileName to prevent path traversal
	sanitizedFileName := filepath.Base(fileName)

	// Construct the full file path
	filePath := filepath.Join(storagePathBase, "user_"+strconv.Itoa(int(userID)), sanitizedFileName)

	// Ensure the directory exists with more restrictive permissions
	if err := os.MkdirAll(filepath.Dir(filePath), 0750); err != nil { // Changed from os.ModePerm to 0750
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// Create the destination file
	dst, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer func() {
		if cerr := dst.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("failed to close file: %w", cerr)
		}
	}()

	// Copy the fileContent to dst
	if _, err := io.Copy(dst, fileContent); err != nil {
		dst.Close()         // Close file before returning on error
		os.Remove(filePath) // Attempt to remove partially written file
		return nil, fmt.Errorf("failed to write to file: %w", err)
	}

	if err := dst.Close(); err != nil { // Explicitly close before encryption or further operations
		os.Remove(filePath) // Attempt to remove file if close fails
		return nil, fmt.Errorf("failed to close file after writing: %w", err)
	}

	isEncrypted := false
	if len(encryptionKey) > 0 {
		// Basic key length check (can be more sophisticated)
		if len(encryptionKey) != 16 && len(encryptionKey) != 24 && len(encryptionKey) != 32 {
			os.Remove(filePath) // Remove the plaintext file if encryption key is invalid
			return nil, fmt.Errorf("invalid encryption key length: must be 16, 24, or 32 bytes: %w", ErrInvalidKeyLength)
		}

		// Stream encryption: Modify EncryptFile to take io.Reader and io.Writer
		// For now, assuming EncryptFile still reads from filePath and overwrites it.
		// If EncryptFile is modified for streaming, the logic here will change significantly.
		if err := EncryptFile(filePath, encryptionKey); err != nil {
			os.Remove(filePath) // Remove the plaintext file if encryption failed
			return nil, fmt.Errorf("failed to encrypt file: %w", err)
		}
		isEncrypted = true
	}

	// Create a File model instance
	fileMetadata := &File{
		Name:        sanitizedFileName, // Use sanitized name
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
