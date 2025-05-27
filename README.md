# Go-Share File Operations Library

## Overview

This Go-Share library helps you manage file uploads, downloads, and AES-GCM encryption/decryption. It stores files locally and uses PostgreSQL (via GORM) for metadata.

## Features

* **File Upload:** Upload files using `multipart/form-data`.
  * Files are stored in user-specific subdirectories.
  * Saves file details (name, path, content type, description, owner) in the database.
* **File Download:** Lets authorized users download their files.
* **AES-GCM Encryption/Decryption:** Optionally encrypt files at rest.
  * Uses AES-GCM with 128, 192, or 256-bit keys.
  * Pass encryption/decryption keys via HTTP headers (see Security Notes for why this is simplified for the example).
* **Basic User-based Authorization:** Only owners can download or change their file metadata. (The `UploadFile`/`DownloadFile` functions don't directly expose metadata modification; that's handled by `File` model methods).

## Configuration

The primary configuration for file storage is the base path where files will be saved on the server.

* **`STORAGE_BASE_PATH` Environment Variable / `storage.base_path` in `config.yaml`:**
  * This setting defines the root directory for all uploaded files.
  * The application defaults to `./uploads` if not specified.
  * It can be set via an environment variable (e.g., `STORAGE_BASE_PATH=/mnt/fileserver/uploads`) or within a `config.yaml` file (`storage.base_path: /mnt/fileserver/uploads`). Viper is used for configuration management, so environment variables typically take precedence if bound correctly.
* **Directory Creation:**
  * The application automatically creates the specified `StorageBasePath` directory (and any necessary parent directories) on startup if it doesn't already exist using `0750` permissions.
  * User-specific subdirectories (`user_<userID>`) are created within the `StorageBasePath` upon the user's first file upload with `0750` permissions.

## API Endpoints (Brief Overview)

The library functions are typically called by HTTP handlers. The Go-Share application exposes the following relevant endpoints:

* **`POST /files`**: Uploads a file.
  * Request Body: `multipart/form-data` with a `file` field for the file content and an optional `description` field.
  * Optional Header: `X-Encryption-Key` (e.g., a 32-byte string) if the client wishes to encrypt the file.
* **`GET /files/{id}`**: Downloads a file.
  * Path Parameter: `{id}` is the ID of the file to download.
  * Optional Header: `X-Decryption-Key` (e.g., a 32-byte string) if the file was previously encrypted and needs to be decrypted. This header is required if the file's `IsEncrypted` flag is true.

## Security Notes

* **Key Management:** The example usage of `X-Encryption-Key` and `X-Decryption-Key` headers is a **major simplification** for demonstration purposes and is **NOT SUITABLE FOR PRODUCTION**. In a real-world application:
  * Encryption keys should be securely generated and managed (e.g., using a Key Management Service like AWS KMS, Google Cloud KMS, HashiCorp Vault).
  * Keys should not be directly passed by clients in headers for new uploads or for general decryption.
  * Key identifiers or wrapped keys might be used, but the raw key material should be handled carefully on the backend.
* **Authorization:** The library provides basic checks (e.g., `file.UserID == requestingUserID`). Ensure your application's authentication and authorization middleware are robust.
* **Input Validation:** File names, descriptions, and other user-supplied data should be validated and sanitized to prevent injection attacks or path traversal issues. The `File` model includes validation tags, and `UploadFile` now uses `filepath.Base()` for filenames.
* **Storage Path Permissions:** The `StorageBasePath` and user-specific subdirectories are created with `0700` permissions by default. Review these permissions based on your specific security requirements.

## Running the Application (Example)

To run the full Go-Share application (which uses this library):

1. **Configure:**
    * Ensure you have a PostgreSQL database running and configured.
    * Create a `.env` file or `config.yaml` in the project root.
    * Example `.env`:

        ```env
        DB_HOST=localhost
        DB_PORT=5432
        DB_USER=youruser
        DB_PASSWORD=yourpassword
        DB_NAME=goshare_db
        JWT_SECRET=yourverysecurejwtsecret
        STORAGE_BASE_PATH=./my_secure_uploads
        ```

    * Example `config.yaml`:

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

2. **Run:**

    ```bash
    go run main.go
    ```

    The server will typically start on port 8080.
