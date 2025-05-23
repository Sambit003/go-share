package files

import (
	"errors"
	"io" // Added io
	"os"

	"gorm.io/gorm"
)

// DownloadFile manages the process of retrieving a file for download.
// It performs the following steps:
// 1. Retrieves the file's metadata from the database using the provided fileID.
// 2. Performs an authorization check to ensure the requesting userID matches the UserID associated with the file.
// 3. If the file is marked as IsEncrypted:
//    a. Checks if a decryptionKey is provided. If not, returns an error.
//    b. Validates the decryptionKey length (must be 16, 24, or 32 bytes for AES).
//    c. Calls DecryptFile to get an io.Reader for the decrypted file content.
//    d. Returns an io.NopCloser wrapping the decrypted content reader.
// 4. If the file is not encrypted, it opens the file directly from the filesystem using its stored Path.
//
// Parameters:
//   - db: A *gorm.DB instance for database interactions.
//   - fileID: The ID of the file to be downloaded.
//   - userID: The ID of the user attempting to download the file, used for authorization.
//   - decryptionKey: An optional byte slice representing the AES decryption key.
//     Required if the file is encrypted. Key management is outside the scope of this function.
//
// Returns:
//   - An io.ReadCloser from which the file content (decrypted, if applicable) can be read.
//     The caller is responsible for closing this ReadCloser.
//   - A pointer to the File metadata object.
//   - An error if any step fails (e.g., file not found, authorization failure, decryption failure,
//     missing decryption key for an encrypted file, or invalid key length).
func DownloadFile(db *gorm.DB, fileID uint, userID uint, decryptionKey []byte) (io.ReadCloser, *File, error) {
	var fileMetadata File
	if err := db.First(&fileMetadata, fileID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, errors.New("file not found")
		}
		return nil, nil, err // Other database error
	}

	// Authorization check
	if fileMetadata.UserID != userID {
		return nil, nil, errors.New("unauthorized: you do not have permission to download this file")
	}

	if fileMetadata.IsEncrypted {
		if len(decryptionKey) == 0 {
			return nil, &fileMetadata, errors.New("file is encrypted, decryption key required")
		}
		// Basic key length check (can be more sophisticated)
		if len(decryptionKey) != 16 && len(decryptionKey) != 24 && len(decryptionKey) != 32 {
			return nil, &fileMetadata, errors.New("invalid decryption key length: must be 16, 24, or 32 bytes")
		}

		decryptedReader, err := DecryptFile(fileMetadata.Path, decryptionKey)
		if err != nil {
			return nil, &fileMetadata, errors.New("failed to decrypt file: " + err.Error())
		}
		return io.NopCloser(decryptedReader), &fileMetadata, nil
	}

	// File is not encrypted, open it normally
	file, err := os.Open(fileMetadata.Path)
	if err != nil {
		return nil, nil, errors.New("error opening file: " + err.Error())
	}

	return file, &fileMetadata, nil
}
