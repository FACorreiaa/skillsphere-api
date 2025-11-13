package handler

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"google.golang.org/protobuf/types/known/timestamppb"

	authv1 "github.com/FACorreiaa/skillsphere-proto/gen/go/auth/v1"

	"github.com/FACorreiaa/skillsphere-api/internal/domain/auth/repository"
	handler2 "github.com/FACorreiaa/skillsphere-api/internal/domain/auth/service"
)

// OAuthLogin handles OAuth login with Google/Apple
func (s *AuthService) OAuthLogin(ctx context.Context, req *connect.Request[authv1.OAuthLoginRequest]) (*connect.Response[authv1.OAuthLoginResponse], error) {
	// Validate input
	if req.Msg.Provider == "" || req.Msg.Code == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("provider and code are required"))
	}

	// Map provider
	provider := mapOAuthProvider(req.Msg.Provider)
	if provider == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("unsupported OAuth provider"))
	}

	// Complete OAuth flow (this is typically done via HTTP redirect flow)
	// For Connect RPC, client should get the code and send it here
	// In a real implementation, you'd exchange the code for user info

	// For now, simulate getting user from Gothic
	// In production, this would be handled via HTTP endpoints
	// and the client would send the resulting code/token here

	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("OAuth flow should be handled via HTTP endpoints. Use /service/{provider} and /service/{provider}/callback"))
}

// HTTP service for OAuth (to be used alongside Connect RPC)
func (s *AuthService) HandleOAuthStart(w http.ResponseWriter, r *http.Request) {
	// Start OAuth flow
	gothic.BeginAuthHandler(w, r)
}

func (s *AuthService) HandleOAuthCallback(w http.ResponseWriter, r *http.Request) {
	// Complete OAuth authentication
	gothUser, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ctx := r.Context()

	// Check if user exists with OAuth identity
	user, err := s.repo.GetUserByOAuthIdentity(ctx, gothUser.Provider, gothUser.UserID)

	if err != nil && errors.Is(err, repository.ErrUserNotFound) {
		// Create new user
		username := gothUser.NickName
		if username == "" {
			username = gothUser.Email
		}

		hashedPassword, _ := handler2.HashPassword(uuid.New().String()) // Random password for OAuth users

		user, err = s.repo.CreateUser(ctx, gothUser.Email, username, hashedPassword, gothUser.Name)
		if err != nil {
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
			return
		}

		// Mark email as verified for OAuth users
		s.repo.VerifyEmail(ctx, user.ID)

		// Create OAuth identity
		accessToken := &gothUser.AccessToken
		refreshToken := &gothUser.RefreshToken
		if err := s.repo.CreateOrUpdateOAuthIdentity(ctx, gothUser.Provider, gothUser.UserID, user.ID, accessToken, refreshToken); err != nil {
			http.Error(w, "Failed to create OAuth identity", http.StatusInternalServerError)
			return
		}

		// Send welcome email
		go s.emailService.SendWelcomeEmail(user.Email, user.DisplayName)
	} else if err != nil {
		http.Error(w, "Failed to get user", http.StatusInternalServerError)
		return
	} else {
		// Update OAuth tokens
		accessToken := &gothUser.AccessToken
		refreshToken := &gothUser.RefreshToken
		s.repo.CreateOrUpdateOAuthIdentity(ctx, gothUser.Provider, gothUser.UserID, user.ID, accessToken, refreshToken)
	}

	// Update last login
	go s.repo.UpdateLastLogin(context.Background(), user.ID)

	// Generate tokens
	tokens, err := s.tokenManager.GenerateTokenPair(
		user.ID.String(),
		user.Email,
		user.Username,
		user.Role,
	)
	if err != nil {
		http.Error(w, "Failed to generate tokens", http.StatusInternalServerError)
		return
	}

	// Store refresh token session
	hashedRefreshToken := hashToken(tokens.RefreshToken)
	userAgent := r.UserAgent()
	clientIP := r.RemoteAddr
	s.repo.CreateUserSession(ctx, user.ID, hashedRefreshToken, userAgent, clientIP, tokens.ExpiresAt.Add(7*24*time.Hour))

	// Redirect to frontend with tokens
	// In production, redirect to your frontend with tokens in URL params or use a different flow
	redirectURL := fmt.Sprintf("%s/service/callback?access_token=%s&refresh_token=%s",
		getEnv("FRONTEND_URL", "http://localhost:3000"),
		tokens.AccessToken,
		tokens.RefreshToken,
	)
	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}

// RequestPasswordReset handles password reset requests
func (s *AuthService) RequestPasswordReset(ctx context.Context, req *connect.Request[authv1.RequestPasswordResetRequest]) (*connect.Response[authv1.RequestPasswordResetResponse], error) {
	if req.Msg.Email == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("email is required"))
	}

	// Get user
	user, err := s.repo.GetUserByEmail(ctx, req.Msg.Email)
	if err != nil {
		// Don't reveal if user exists or not
		return connect.NewResponse(&authv1.RequestPasswordResetResponse{
			Success: true,
			Message: "If an account exists with this email, a password reset link has been sent",
		}), nil
	}

	// Generate reset token
	resetToken, err := handler2.GeneratePasswordResetToken()
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to generate reset token"))
	}

	// Hash and store token (1 hour expiration)
	hashedToken := hashToken(resetToken)
	expiresAt := time.Now().Add(1 * time.Hour)
	if err := s.repo.CreateUserToken(ctx, user.ID, hashedToken, "password_reset", expiresAt); err != nil {
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to create reset token"))
	}

	// Send reset email
	go s.emailService.SendPasswordResetEmail(user.Email, user.DisplayName, resetToken)

	return connect.NewResponse(&authv1.RequestPasswordResetResponse{
		Success: true,
		Message: "If an account exists with this email, a password reset link has been sent",
	}), nil
}

// ResetPassword handles password reset with token
func (s *AuthService) ResetPassword(ctx context.Context, req *connect.Request[authv1.ResetPasswordRequest]) (*connect.Response[authv1.ResetPasswordResponse], error) {
	if req.Msg.Token == "" || req.Msg.NewPassword == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("token and new password are required"))
	}

	// Validate new password
	if err := handler2.ValidatePassword(req.Msg.NewPassword); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	// Get and validate token
	hashedToken := hashToken(req.Msg.Token)
	token, err := s.repo.GetUserTokenByHash(ctx, hashedToken, "password_reset")
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid or expired reset token"))
	}

	// Hash new password
	hashedPassword, err := handler2.HashPassword(req.Msg.NewPassword)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to hash password"))
	}

	// Update password
	if err := s.repo.UpdatePassword(ctx, token.UserID, hashedPassword); err != nil {
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to update password"))
	}

	// Delete used token
	s.repo.DeleteUserToken(ctx, hashedToken)

	// Invalidate all sessions for security
	s.repo.DeleteAllUserSessions(ctx, token.UserID)

	return connect.NewResponse(&authv1.ResetPasswordResponse{
		Success: true,
		Message: "Password reset successfully",
	}), nil
}

// ChangePassword handles password change for authenticated users
func (s *AuthService) ChangePassword(ctx context.Context, req *connect.Request[authv1.ChangePasswordRequest]) (*connect.Response[authv1.ChangePasswordResponse], error) {
	// This requires authentication - extract user from context
	// Assuming service interceptor adds user ID to context
	userIDStr := ctx.Value("user_id")
	if userIDStr == nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("authentication required"))
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, errors.New("invalid user ID"))
	}

	if req.Msg.CurrentPassword == "" || req.Msg.NewPassword == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("current password and new password are required"))
	}

	// Get user
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to get user"))
	}

	// Verify current password
	if !handler2.ComparePassword(user.HashedPassword, req.Msg.CurrentPassword) {
		return nil, connect.NewError(connect.CodePermissionDenied, errors.New("current password is incorrect"))
	}

	// Validate new password
	if err := handler2.ValidatePassword(req.Msg.NewPassword); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	// Hash new password
	hashedPassword, err := handler2.HashPassword(req.Msg.NewPassword)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to hash password"))
	}

	// Update password
	if err := s.repo.UpdatePassword(ctx, userID, hashedPassword); err != nil {
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to update password"))
	}

	// Invalidate all sessions except current one
	// In production, you'd want to keep the current session active
	s.repo.DeleteAllUserSessions(ctx, userID)

	return connect.NewResponse(&authv1.ChangePasswordResponse{
		Success: true,
		Message: "Password changed successfully",
	}), nil
}

// VerifyEmail handles email verification
func (s *AuthService) VerifyEmail(ctx context.Context, req *connect.Request[authv1.VerifyEmailRequest]) (*connect.Response[authv1.VerifyEmailResponse], error) {
	if req.Msg.Token == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("verification token is required"))
	}

	// Get and validate token
	hashedToken := hashToken(req.Msg.Token)
	token, err := s.repo.GetUserTokenByHash(ctx, hashedToken, "email_verification")
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid or expired verification token"))
	}

	// Mark email as verified
	if err := s.repo.VerifyEmail(ctx, token.UserID); err != nil {
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to verify email"))
	}

	// Delete used token
	s.repo.DeleteUserToken(ctx, hashedToken)

	// Get user for welcome email
	user, _ := s.repo.GetUserByID(ctx, token.UserID)
	if user != nil {
		go s.emailService.SendWelcomeEmail(user.Email, user.DisplayName)
	}

	return connect.NewResponse(&authv1.VerifyEmailResponse{
		Success: true,
		Message: "Email verified successfully",
	}), nil
}

// ResendVerificationEmail resends the verification email
func (s *AuthService) ResendVerificationEmail(ctx context.Context, req *connect.Request[authv1.ResendVerificationEmailRequest]) (*connect.Response[authv1.ResendVerificationEmailResponse], error) {
	if req.Msg.Email == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("email is required"))
	}

	// Get user
	user, err := s.repo.GetUserByEmail(ctx, req.Msg.Email)
	if err != nil {
		// Don't reveal if user exists
		return connect.NewResponse(&authv1.ResendVerificationEmailResponse{
			Success: true,
			Message: "If an unverified account exists with this email, a verification link has been sent",
		}), nil
	}

	// Check if already verified
	if user.EmailVerifiedAt != nil {
		return connect.NewResponse(&authv1.ResendVerificationEmailResponse{
			Success: true,
			Message: "Email is already verified",
		}), nil
	}

	// Generate new verification token
	verificationToken, err := handler2.GenerateVerificationToken()
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to generate verification token"))
	}

	// Hash and store token
	hashedToken := hashToken(verificationToken)
	expiresAt := time.Now().Add(24 * time.Hour)
	if err := s.repo.CreateUserToken(ctx, user.ID, hashedToken, "email_verification", expiresAt); err != nil {
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to create verification token"))
	}

	// Send verification email
	go s.emailService.SendVerificationEmail(user.Email, user.DisplayName, verificationToken)

	return connect.NewResponse(&authv1.ResendVerificationEmailResponse{
		Success: true,
		Message: "Verification email sent",
	}), nil
}

// Helper functions

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func toProtoTimestamp(t time.Time) *timestamppb.Timestamp {
	return timestamppb.New(t)
}

func mapOAuthProvider(provider string) string {
	switch provider {
	case "google":
		return goth.ProviderGoogle
	case "apple":
		return "apple"
	default:
		return ""
	}
}

func getEnv(key, fallback string) string {
	if value := context.Background().Value(key); value != nil {
		return value.(string)
	}
	return fallback
}
