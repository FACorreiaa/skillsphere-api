# Authentication System Setup Guide

## Overview

Complete JWT-based authentication system with OAuth (Google & Apple) for SkillSphere API using Connect RPC.

## Features

- ✅ JWT access & refresh tokens
- ✅ Email/password registration and login
- ✅ OAuth login (Google & Apple via Goth)
- ✅ Email verification
- ✅ Password reset flow
- ✅ Password change for authenticated users
- ✅ Session management
- ✅ Token validation
- ✅ Role-based authorization
- ✅ Secure password hashing (bcrypt)
- ✅ SMTP email sending

## Installation

### 1. Install Dependencies

```bash
go get github.com/golang-jwt/jwt/v5
go get github.com/markbates/goth
go get github.com/gorilla/sessions
go get golang.org/x/crypto/bcrypt
go get github.com/google/uuid
go get github.com/gorilla/mux
```

### 2. Setup Environment Variables

Copy `.env.example` to `.env` and configure:

```bash
cp .env.example .env
```

Required variables:
- `DATABASE_URL` - PostgreSQL connection string
- `JWT_SECRET` - Secret key for JWT signing
- `GOOGLE_CLIENT_ID` & `GOOGLE_CLIENT_SECRET` - Google OAuth credentials
- `APPLE_CLIENT_ID` & `APPLE_SECRET` - Apple OAuth credentials

### 3. Run Database Migrations

Ensure all migrations are applied:

```bash
goose -dir pkg/db/migrations postgres "$DATABASE_URL" up
```

## Usage

### Server Setup

See `examples/auth_server_example.go` for a complete example.

```go
// Initialize handler
authRepo := repository.NewAuthRepository(db)
tokenManager := auth.NewTokenManager(jwtSecret, jwtSecret, 15*time.Minute, 7*24*time.Hour)
emailService := services.NewEmailService()
ontologyEmitter := ontology.NewJSONEmitter(ontology.NewLogSender(logger))
authService := services.NewAuthService(authRepo, tokenManager, emailService, logger, ontologyEmitter, 7*24*time.Hour)

// Setup interceptor
authInterceptor := interceptors.NewAuthInterceptor(jwtSecret)

// Register Connect RPC service
authPath, authHandler := authv1connect.NewAuthServiceHandler(authService)
```

### Client Usage Examples

#### 1. Register

```bash
curl -X POST http://localhost:8080/skillsphere.auth.v1.AuthService/Register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "username": "johndoe",
    "password": "SecurePass123!",
    "display_name": "John Doe"
  }'
```

Response:
```json
{
  "user": {
    "user_id": "uuid",
    "email": "user@example.com",
    "username": "johndoe",
    "display_name": "John Doe",
    "role": "member"
  },
  "access_token": "eyJhbGc...",
  "refresh_token": "eyJhbGc...",
  "expires_at": "2025-11-14T..."
}
```

#### 2. Login

```bash
curl -X POST http://localhost:8080/skillsphere.auth.v1.AuthService/Login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePass123!"
  }'
```

#### 3. Refresh Token

```bash
curl -X POST http://localhost:8080/skillsphere.auth.v1.AuthService/RefreshToken \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "eyJhbGc..."
  }'
```

#### 4. OAuth Login (Google)

```bash
# Step 1: Redirect user to OAuth start
open http://localhost:8080/auth/google

# Step 2: User authenticates with Google
# Step 3: Callback returns to /service/google/callback
# Step 4: Frontend receives tokens via redirect
```

#### 5. Verify Email

```bash
curl -X POST http://localhost:8080/skillsphere.auth.v1.AuthService/VerifyEmail \
  -H "Content-Type: application/json" \
  -d '{
    "token": "verification-token-from-email"
  }'
```

#### 6. Request Password Reset

```bash
curl -X POST http://localhost:8080/skillsphere.auth.v1.AuthService/RequestPasswordReset \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com"
  }'
```

#### 7. Reset Password

```bash
curl -X POST http://localhost:8080/skillsphere.auth.v1.AuthService/ResetPassword \
  -H "Content-Type: application/json" \
  -d '{
    "token": "reset-token-from-email",
    "new_password": "NewSecurePass123!"
  }'
```

#### 8. Change Password (Authenticated)

```bash
curl -X POST http://localhost:8080/skillsphere.auth.v1.AuthService/ChangePassword \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -d '{
    "current_password": "CurrentPass123!",
    "new_password": "NewPass123!"
  }'
```

## OAuth Setup

### Google OAuth

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select existing
3. Enable Google+ API
4. Create OAuth 2.0 credentials
5. Add authorized redirect URI: `http://localhost:8080/auth/google/callback`
6. Copy Client ID and Client Secret to `.env`

### Apple OAuth

1. Go to [Apple Developer Portal](https://developer.apple.com/)
2. Create an App ID
3. Enable "Sign in with Apple"
4. Create a Service ID
5. Configure redirect URI: `http://localhost:8080/auth/apple/callback`
6. Generate a key for your Service ID
7. Copy credentials to `.env`

## Architecture

### Components

```
pkg/
├── auth/
│   ├── token_manager.go    # JWT generation/validation
│   ├── password.go          # Password hashing/validation
│   └── oauth.go             # OAuth provider setup
├── interceptors/
│   └── auth.go              # JWT authentication interceptor
├── repository/
│   └── auth_repository.go   # Database operations
└── services/
    ├── auth_service.go      # Main auth service
    ├── auth_service_oauth.go # OAuth handlers
    └── email_service.go     # Email sending
```

### Flow Diagrams

#### Registration Flow
```
Client → Register RPC → Validate Password → Hash Password
→ Create User → Generate Verification Token → Send Email
→ Generate JWT Tokens → Return Response
```

#### Login Flow
```
Client → Login RPC → Get User by Email → Verify Password
→ Check Active Status → Generate JWT Tokens → Create Session
→ Return Response
```

#### OAuth Flow
```
Client → /auth/google → Google Auth → /auth/google/callback
→ Get/Create User → Link OAuth Identity → Generate JWT Tokens
→ Redirect to Frontend with Tokens
```

## Security Considerations

### Production Recommendations

1. **JWT Secrets**
   - Use different secrets for access and refresh tokens
   - Rotate secrets periodically
   - Use strong, random secrets (min 32 bytes)

2. **HTTPS**
   - Always use HTTPS in production
   - Enable secure cookies for OAuth sessions

3. **Password Policy**
   - Minimum 8 characters
   - Requires uppercase, lowercase, digit, special character
   - Consider adding password history checks

4. **Rate Limiting**
   - Implement rate limiting on auth endpoints
   - Use exponential backoff for failed login attempts

5. **Session Management**
   - Implement session timeout
   - Clean up expired sessions regularly
   - Allow users to view/revoke active sessions

6. **OAuth**
   - Validate OAuth state parameter
   - Use PKCE for mobile apps
   - Store minimal OAuth tokens

## Testing

### Unit Tests

```bash
go test ./pkg/service/...
go test ./pkg/handler/...
go test ./pkg/repository/...
```

### Integration Tests

```bash
go test ./test/integration/auth_test.go
```

### Manual Testing with cURL

See client usage examples above.

## Troubleshooting

### Common Issues

1. **"missing authorization header"**
   - Ensure you're sending `Authorization: Bearer <token>` header
   - Check token hasn't expired

2. **"invalid token"**
   - Verify JWT_SECRET matches what was used to generate token
   - Check token format (should be three base64 parts)

3. **OAuth callback fails**
   - Verify redirect URI matches exactly what's configured in provider
   - Check SESSION_SECRET is set

4. **Email not sending**
   - Check SMTP credentials are correct
   - For Gmail, use App-Specific Password
   - Leave SMTP_HOST empty in development to skip emails

## API Reference

See generated proto documentation at `gen/go/auth/v1/`

All RPC methods:
- `Register(RegisterRequest) → RegisterResponse`
- `Login(LoginRequest) → LoginResponse`
- `Logout(LogoutRequest) → LogoutResponse`
- `RefreshToken(RefreshTokenRequest) → RefreshTokenResponse`
- `ValidateToken(ValidateTokenRequest) → ValidateTokenResponse`
- `OAuthLogin(OAuthLoginRequest) → OAuthLoginResponse`
- `RequestPasswordReset(RequestPasswordResetRequest) → RequestPasswordResetResponse`
- `ResetPassword(ResetPasswordRequest) → ResetPasswordResponse`
- `ChangePassword(ChangePasswordRequest) → ChangePasswordResponse`
- `VerifyEmail(VerifyEmailRequest) → VerifyEmailResponse`
- `ResendVerificationEmail(ResendVerificationEmailRequest) → ResendVerificationEmailResponse`

## License

MIT License
