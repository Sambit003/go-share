package files

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"os"
)

// EncryptData encrypts a byte slice using AES-GCM (Galois/Counter Mode).
// AES-GCM is an authenticated encryption mode that provides both confidentiality and integrity.
// The encryption key must be 16, 24, or 32 bytes long, corresponding to
// AES-128, AES-192, or AES-256, respectively.
// A random nonce is generated for each encryption operation and prepended to the ciphertext.
//
// Parameters:
//   - data: The plaintext data to encrypt.
//   - key: The AES encryption key.
//
// Returns:
//   - A byte slice containing the nonce prepended to the ciphertext.
//   - An error if key creation or GCM initialization fails.
//
// Note: Secure key management (generation, storage, distribution) is critical and
// is outside the scope of this function.
func EncryptData(data []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher block: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// gcm.Seal prepends the nonce to the ciphertext
	return gcm.Seal(nonce, nonce, data, nil), nil
}

// DecryptData decrypts a byte slice that was previously encrypted using AES-GCM
// by the EncryptData function.
// The encryption key must be the same as the one used for encryption and must be
// 16, 24, or 32 bytes long.
// The function expects the nonce to be prepended to the ciphertext.
//
// Parameters:
//   - data: The ciphertext, with the nonce prepended.
//   - key: The AES decryption key.
//
// Returns:
//   - A byte slice containing the original plaintext.
//   - An error if key creation, GCM initialization, or decryption (e.g., authentication failure) fails,
//     or if the ciphertext is too short to contain a nonce.
//
// Note: Secure key management is critical and is outside the scope of this function.
func DecryptData(data []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher block: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	decryptedData, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data: %w", err) // Wrap GCM open error
	}
	return decryptedData, nil
}

// EncryptFile encrypts a file using AES-GCM streaming.
// It reads from filePath, encrypts content, and writes to a temporary file,
// then replaces the original file.
func EncryptFile(filePath string, key []byte) error {
	inputFile, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open input file for encryption: %w", err)
	}
	defer inputFile.Close()

	tempFilePath := filePath + ".tmp"
	tempFile, err := os.Create(tempFilePath)
	if err != nil {
		return fmt.Errorf("failed to create temporary file for encryption: %w", err)
	}
	// Ensure tempFile is closed and removed in case of errors or successful rename
	defer func() {
		tempFile.Close()
		// Attempt to remove temp file. If os.Rename succeeded, this will (and should) fail.
		// If os.Rename failed or was not reached, this cleans up.
		os.Remove(tempFilePath)
	}()

	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("failed to create cipher block for encryption: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("failed to create GCM for encryption: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return fmt.Errorf("failed to generate nonce for encryption: %w", err)
	}

	if _, err := tempFile.Write(nonce); err != nil {
		return fmt.Errorf("failed to write nonce to temporary file: %w", err)
	}

	// GCM's Seal function can be used for streaming if we manage the ciphertext output.
	// However, a more explicit stream cipher mode like CTR or CFB is often used with GCM for just integrity.
	// For AES-GCM authenticated encryption stream:
	// We write nonce, then ciphertext. GCM handles both encryption and authentication tag.
	// The "streaming" part for very large files remains a TODO as it's complex with GCM.

	chunkSize := 64 * 1024 // 64 KB chunks
	buffer := make([]byte, chunkSize)

	for {
		n, err := inputFile.Read(buffer)
		if err != nil && err != io.EOF {
			return fmt.Errorf("failed to read chunk from input file: %w", err)
		}
		if n == 0 {
			break
		}

		encryptedChunk := gcm.Seal(nil, nonce, buffer[:n], nil)
		if _, err := tempFile.Write(encryptedChunk); err != nil {
			return fmt.Errorf("failed to write encrypted chunk to temporary file: %w", err)
		}
	}

	if err := tempFile.Close(); err != nil {
		// If close fails, os.Remove(tempFilePath) in defer will still run.
		return fmt.Errorf("failed to close temporary file after writing encrypted data: %w", err)
	}

	// Replace the original file with the temporary file
	if err := os.Rename(tempFilePath, filePath); err != nil {
		// If rename fails, os.Remove(tempFilePath) in defer will clean up the .tmp file.
		return fmt.Errorf("failed to replace original file with encrypted file: %w", err)
	}
	// If rename succeeds, the defer os.Remove(tempFilePath) will try to remove the *new* filePath + ".tmp"
	// which won't exist, which is fine. The original tempFilePath (which was renamed) is gone.

	return nil
}

// DecryptFile decrypts a file using AES-GCM.
// It reads the encrypted file, decrypts its content, and returns a *bytes.Reader.
// TODO: Implement true streaming decryption for large files.
func DecryptFile(filePath string, key []byte) (*bytes.Reader, error) {
	encryptedContent, err := os.ReadFile(filePath) // Reads the whole file
	if err != nil {
		return nil, fmt.Errorf("failed to read encrypted file for decryption: %w", err)
	}

	decryptedContent, err := DecryptData(encryptedContent, key)
	if err != nil {
		// This will catch GCM authentication errors like "cipher: message authentication failed"
		return nil, fmt.Errorf("failed to decrypt file data: %w", err)
	}

	return bytes.NewReader(decryptedContent), nil
}
