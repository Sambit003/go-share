# Go-Share File Operations Library

## Overview

This library, part of the Go-Share application, provides functionalities for managing file uploads, downloads, and AES-GCM encryption/decryption. It is designed to store files on a local filesystem, with metadata managed in a PostgreSQL database via GORM.

## Features

*   **File Upload:** Handles multipart/form-data uploads to the local filesystem.
    *   Stores files in user-specific subdirectories.
    *   Records file metadata (name, path, content type, description, owner) in the database.
*   **File Download:** Allows authorized users to download their files.
*   **AES-GCM Encryption/Decryption:** Supports optional encryption of files at rest.
    *   Uses AES-GCM with 128, 192, or 256-bit keys.
    *   Encryption/decryption keys are passed via HTTP headers for simplicity in this example (see Security Notes).
*   **Basic User-based Authorization:** Ensures only file owners can download or modify their file metadata (though metadata modification is not directly exposed by `UploadFile`/`DownloadFile` but by `File` model methods).

## Configuration

The primary configuration for file storage is the base path where files will be saved on the server.

*   **`STORAGE_BASE_PATH` Environment Variable / `storage.base_path` in `config.yaml`:**
    *   This setting defines the root directory for all uploaded files.
    *   The application defaults to `./uploads` if not specified.
    *   It can be set via an environment variable (e.g., `STORAGE_BASE_PATH=/mnt/fileserver/uploads`) or within a `config.yaml` file (`storage.base_path: /mnt/fileserver/uploads`). Viper is used for configuration management, so environment variables typically take precedence if bound correctly.
*   **Directory Creation:**
    *   The application automatically creates the specified `StorageBasePath` directory (and any necessary parent directories) on startup if it doesn't already exist.
    *   User-specific subdirectories (`user_<userID>`) are created within the `StorageBasePath` upon the user's first file upload.

## Library Usage (Code Examples)

The core library functions are within the `pkg/files` package.

### Initialization

The library functions require a `*gorm.DB` instance for database operations. This is typically initialized in your application's main setup.

```go
// main.go or config/config.go (conceptual)
import (
    "gorm.io/gorm"
    "gorm.io/driver/postgres"
    "github.com/spf13/viper"
    // ... other imports
)

var DB *gorm.DB

func ConnectDatabase() {
    // Viper loads DB config from config.yaml or environment variables
    dbHost := viper.GetString("database.host")
    // ... other db config params
    dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable", ...)
    var err error
    DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatalf("Error connecting to database: %s", err)
    }
}

func main() {
    // Load Viper config (e.g., from config.yaml and .env)
    viper.SetDefault("storage.base_path", "./uploads")
    // ... other viper setup ...
    viper.ReadInConfig() // Or load .env first if using godotenv

    ConnectDatabase() // Initialize config.DB (or your global DB instance)
    // ... rest of your application setup
}
```

### Uploading a File

```go
package main // Or your relevant package

import (
    "fmt"
    "io"
    "log"
    "net/http"
    "os" // For example file reader

    "go-share/pkg/files" // Import the library
    "go-share/models"    // Assuming User model is here
    "gorm.io/gorm"       // For db instance
    "github.com/spf13/viper"
)

// Example usage in a handler or service
func handleFileUpload(db *gorm.DB, user models.User, fileReader io.Reader, fileName, contentType, description string, useEncryption bool) (*files.File, error) {
    storageBasePath := viper.GetString("storage.base_path")
    if storageBasePath == "" {
        storageBasePath = "./uploads" // Ensure a fallback if not configured
    }

    var encryptionKey []byte
    if useEncryption {
        // IMPORTANT: Key management is critical. 
        // This is a placeholder. DO NOT use hardcoded keys in production.
        // Keys should be securely generated, stored (e.g., Vault, KMS), and retrieved.
        encryptionKey = []byte("your-super-secret-32-byte-key!!") 
        if len(encryptionKey) != 32 { // Example for AES-256
             return nil, fmt.Errorf("invalid encryption key length")
        }
    }

    newFile, err := files.UploadFile(db, fileReader, fileName, contentType, description, user.ID, storageBasePath, encryptionKey)
    if err != nil {
        log.Printf("Error uploading file '%s': %v\n", fileName, err)
        return nil, err
    }

    fmt.Printf("File '%s' uploaded successfully. Path: %s, Encrypted: %t\n", newFile.Name, newFile.Path, newFile.IsEncrypted)
    return newFile, nil
}

// Example of how you might call it (simplified):
// func MyUploadHandler(w http.ResponseWriter, r *http.Request, db *gorm.DB, currentUser models.User) {
//     r.ParseMultipartForm(10 << 20) // 10 MB
//     formFile, handler, err := r.FormFile("uploadfile")
//     // ... error handling ...
//     defer formFile.Close()
//     description := r.FormValue("description")
//     encrypt := r.FormValue("encrypt") == "true"
//
//     // The X-Encryption-Key header is shown for API consistency, but the key for UploadFile
//     // would ideally come from a secure backend source, not directly from a client header for new uploads.
//     // For this library example, we pass it directly to UploadFile.
//
//     uploadedFile, err := handleFileUpload(db, currentUser, formFile, handler.Filename, handler.Header.Get("Content-Type"), description, encrypt)
//     // ... respond to client ...
// }
```

### Downloading a File

```go
package main // Or your relevant package

import (
    "fmt"
    "io"
    "log"
    "net/http" // For example: writing to ResponseWriter

    "go-share/pkg/files" // Import the library
    "go-share/models"    // Assuming User model is here
    "gorm.io/gorm"       // For db instance
)

// Example usage in a handler or service
func handleFileDownload(db *gorm.DB, fileID uint, user models.User, w http.ResponseWriter, r *http.Request) error {
    var decryptionKey []byte
    // If the file is known to be encrypted or if client indicates it (e.g. via a query param, not shown)
    // For this example, we check a header, similar to the API.
    decryptionKeyHeader := r.Header.Get("X-Decryption-Key")
    if decryptionKeyHeader != "" {
        // IMPORTANT: Secure key management is critical.
        decryptionKey = []byte(decryptionKeyHeader)
         if len(decryptionKey) != 16 && len(decryptionKey) != 24 && len(decryptionKey) != 32 {
             // Handle invalid key length error for client
             return fmt.Errorf("invalid decryption key length")
        }
    }

    fileStream, fileMetadata, err := files.DownloadFile(db, fileID, user.ID, decryptionKey)
    if err != nil {
        log.Printf("Error downloading file ID %d: %v\n", fileID, err)
        // Handle specific errors like "file not found", "unauthorized", "decryption failed"
        return err
    }
    defer fileStream.Close()

    // Set headers for download
    w.Header().Set("Content-Disposition", "attachment; filename=\""+fileMetadata.Name+"\"")
    ct := fileMetadata.ContentType
    if ct == "" {
        ct = "application/octet-stream" // Default
    }
    w.Header().Set("Content-Type", ct)

    // Content-Length can be tricky for encrypted files if original size isn't stored/retrievable.
    // The library currently only sets it for non-encrypted os.File instances.
    // if fileMetadata.Size > 0 && !fileMetadata.IsEncrypted { // Assuming Size field exists
    //     w.Header().Set("Content-Length", strconv.FormatInt(fileMetadata.Size, 10))
    // }

    _, err = io.Copy(w, fileStream)
    if err != nil {
        log.Printf("Error streaming file ID %d to client: %v\n", fileID, err)
        return err
    }
    return nil
}
```

## API Endpoints (Brief Overview)

The library functions are typically called by HTTP handlers. The Go-Share application exposes the following relevant endpoints:

*   **`POST /files`**: Uploads a file.
    *   Request Body: `multipart/form-data` with a `file` field for the file content and an optional `description` field.
    *   Optional Header: `X-Encryption-Key` (e.g., a 32-byte string) if the client wishes to encrypt the file.
*   **`GET /files/{id}`**: Downloads a file.
    *   Path Parameter: `{id}` is the ID of the file to download.
    *   Optional Header: `X-Decryption-Key` (e.g., a 32-byte string) if the file was previously encrypted and needs to be decrypted. This header is required if the file's `IsEncrypted` flag is true.

## Security Notes

*   **Key Management:** The example usage of `X-Encryption-Key` and `X-Decryption-Key` headers is a **major simplification** for demonstration purposes and is **NOT SUITABLE FOR PRODUCTION**. In a real-world application:
    *   Encryption keys should be securely generated and managed (e.g., using a Key Management Service like AWS KMS, Google Cloud KMS, HashiCorp Vault).
    *   Keys should not be directly passed by clients in headers for new uploads or for general decryption.
    *   Key identifiers or wrapped keys might be used, but the raw key material should be handled carefully on the backend.
*   **Authorization:** The library provides basic checks (e.g., `file.UserID == requestingUserID`). Ensure your application's authentication and authorization middleware are robust.
*   **Input Validation:** File names, descriptions, and other user-supplied data should be validated and sanitized to prevent injection attacks or path traversal issues (though `filepath.Join` helps with the latter). The `File` model includes validation tags.
*   **Storage Path Permissions:** The `StorageBasePath` and its subdirectories should have appropriate filesystem permissions to prevent unauthorized access. The application creates directories with `os.ModePerm` (0777 by default in Go's `os` package for `MkdirAll` if not overridden by umask, though the code explicitly passes `os.ModePerm`), which might be too permissive. Consider setting more restrictive permissions (e.g., 0750 or 0700) as appropriate.

## Running the Application (Example)

To run the full Go-Share application (which uses this library):

1.  **Configure:**
    *   Ensure you have a PostgreSQL database running and configured.
    *   Create a `.env` file or `config.yaml` in the project root.
    *   Example `.env`:
        ```env
        DB_HOST=localhost
        DB_PORT=5432
        DB_USER=youruser
        DB_PASSWORD=yourpassword
        DB_NAME=goshare_db
        JWT_SECRET=yourverysecurejwtsecret
        STORAGE_BASE_PATH=./my_secure_uploads
        ```
    *   Example `config.yaml`:
        ```yaml
        database:
          host: localhost
          port: "5432"
          user: youruser
          password: yourpassword
          name: goshare_db
        jwt_secret: yourverysecurejwtsecret
        storage:
          base_path: ./my_secure_uploads
        ```

2.  **Run:**
    ```bash
    go run main.go
    ```
    The server will typically start on port 8080.Okay, I've updated `README.md` with the library API details, usage examples, and configuration information.

I've completed the following steps:

1.  Added GoDoc comments to `pkg/files/file.go`.
2.  Added GoDoc comments to `pkg/files/upload.go`.
3.  Added GoDoc comments to `pkg/files/download.go`.
4.  Added GoDoc comments to `pkg/files/encryption.go`.
5.  Updated `README.md` with all the specified sections.

All documentation tasks for this subtask appear to be complete. I've tried to make the GoDoc comments informative and the README.md comprehensive for a developer using or understanding this library.
