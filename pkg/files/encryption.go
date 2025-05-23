package files

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
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
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
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
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

// EncryptFile reads the content of a file, encrypts it using EncryptData,
// and then writes the encrypted content back to the original file, overwriting it.
//
// Parameters:
//   - filePath: The path to the file to be encrypted.
//   - key: The AES encryption key (16, 24, or 32 bytes).
//
// Returns:
//   - An error if reading the file, encrypting its content, or writing the
//     encrypted content back to the file fails.
func EncryptFile(filePath string, key []byte) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	encryptedContent, err := EncryptData(content, key)
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, encryptedContent, 0644)
}

// DecryptFile reads the content of an encrypted file, decrypts it using DecryptData,
// and returns a *bytes.Reader for the decrypted content.
//
// Parameters:
//   - filePath: The path to the encrypted file.
//   - key: The AES decryption key (16, 24, or 32 bytes).
//
// Returns:
//   - A *bytes.Reader containing the decrypted file content.
//   - An error if reading the file or decrypting its content fails.
func DecryptFile(filePath string, key []byte) (*bytes.Reader, error) {
	encryptedContent, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	decryptedContent, err := DecryptData(encryptedContent, key)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(decryptedContent), nil
}
