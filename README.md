# GoShare - File Sharing API

GoShare is a Go-based API framework for building file sharing applications. It provides a starting point with essential features like user authentication, file management, and a basic API structure, allowing you to focus on your specific application requirements.

## Features

- **User Authentication:** Secure user registration and login with password hashing and JWT-based authentication.
- **File Management:** Create, read, update, and delete file metadata, with authorization checks to ensure data security.
- **API Structure:** Provides a basic RESTful API structure, making it easy to extend with additional endpoints.
- **Database Integration:** Uses GORM for seamless interaction with a PostgreSQL database.

## Getting Started

### Prerequisites

- Go 1.16 or later
- PostgreSQL

### Installation

1. **Clone the repository:**
   ```bash
   git clone https://github.com/your-username/go-share.git
   cd go-share
   ```

2. **Install dependencies:**
   ```bash
   go get github.com/golang-jwt/jwt/v5
   go get github.com/go-playground/validator/v10
   go get github.com/gorilla/mux
   go get gorm.io/driver/postgres
   go get gorm.io/gorm
   ```

3. **Configure `config.yaml`:**
   Create a `config.yaml` file in the project root directory and set the following:
   ```yaml
   database:
     host: your_db_host
     port: your_db_port
     user: your_db_user
     password: your_db_password
     name: your_db_name
   ```

4. **Run the server:**
   ```bash
   go run main.go
   ```

## Project Checklist

### Done:
- [x] User authentication (registration and login).
- [x] File metadata management (CRUD operations).
- [x] Basic RESTful API structure.
- [x] Database integration with GORM (PostgreSQL).
- [x] Code refactoring for best practices and readability.

### Upcoming:
- [ ] File storage implementation (local or cloud storage).
- [ ] Detailed API endpoint design and documentation.
- [ ] Rate limiting to prevent abuse.
- [ ] Metrics and monitoring setup.
- [ ] Production-ready deployment (Docker, etc.).
- [ ] Security hardening.

## Contributing

Contributions are welcome! To contribute to GoShare:
1. Fork the repository.
2. Create a new branch for your feature or bug fix.
3. Commit your changes with clear and concise commit messages.
4. Push your branch to your fork.
5. Open a pull request.

Please follow Go coding conventions and ensure that your code is well-tested.
