# SaaS Go Kit - Basic Example

This example demonstrates how to use the SaaS Go Kit authentication module in a real application.

## Features

- User registration and login
- JWT-based authentication
- Email verification (console output in dev mode)
- Password reset flow
- Profile management
- Rate limiting
- GORM integration with SQLite

## Setup

1. Copy the environment file:
```bash
cp .env.example .env
```

2. Install dependencies:
```bash
go mod download
```

3. Run the application:
```bash
go run .
```

The server will start on `http://localhost:8080`

## API Endpoints

### Public Endpoints

- `POST /api/v1/auth/register` - Register a new account
- `POST /api/v1/auth/login` - Login to an account
- `GET /api/v1/auth/verify-email?token=xxx` - Verify email address
- `POST /api/v1/auth/forgot-password` - Request password reset
- `POST /api/v1/auth/reset-password` - Reset password with token

### Protected Endpoints (require JWT token)

- `POST /api/v1/auth/refresh` - Refresh JWT token
- `POST /api/v1/auth/logout` - Logout
- `GET /api/v1/auth/me` - Get current user profile
- `PUT /api/v1/auth/profile` - Update profile
- `POST /api/v1/auth/change-password` - Change password
- `POST /api/v1/auth/send-verification` - Resend verification email

## Example Usage

### Register a new account
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "securepassword123",
    "company_name": "Example Corp"
  }'
```

### Login
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "securepassword123"
  }'
```

### Get profile (with JWT token)
```bash
curl -X GET http://localhost:8080/api/v1/auth/me \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## Development

In development mode, emails are printed to the console instead of being sent. Check the console output for verification links and password reset tokens.

## Architecture

This example demonstrates:

1. **Modular Design**: Auth module is registered as a self-contained unit
2. **Interface-based Storage**: GORM implementations of auth storage interfaces
3. **Provider Pattern**: Email and config providers are injected
4. **Event System**: Logging of auth events
5. **Rate Limiting**: Protection against brute force attacks
6. **Error Handling**: Consistent error responses using saas-go-kit/errors

## Customization

You can customize the auth behavior by:

1. Implementing custom storage providers
2. Adding custom password validators
3. Implementing real email providers
4. Adding custom event listeners
5. Implementing audit logging
6. Adding custom middleware