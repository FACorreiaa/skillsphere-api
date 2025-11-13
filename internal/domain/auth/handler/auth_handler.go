package handler

import (
	"context"
	"errors"
	"fmt"
	"time"

	"connectrpc.com/connect"
	authv1 "github.com/FACorreiaa/skillsphere-proto/gen/go/auth/v1"
	pb "github.com/FACorreiaa/skillsphere-proto/gen/go/auth/v1/authv1connect"
	userv1 "github.com/FACorreiaa/skillsphere-proto/gen/go/user/v1"
	"github.com/google/uuid"

	"github.com/FACorreiaa/skillsphere-api/internal/domain/auth/repository"
	"github.com/FACorreiaa/skillsphere-api/internal/domain/auth/service"
)

// AuthHandler implements the AuthHandler RPC methods
type AuthHandler struct {
	pb.UnimplementedAuthServiceHandler
	service *service.AuthService
}

// NewAuthHandler creates a new service service
func NewAuthHandler(svc *service.AuthService) *AuthHandler {
	return &AuthHandler{
		service: svc
	}
}

// Register handles user registration
func (h *AuthHandler) Register(
	ctx context.Context,
	req *connect.Request[authv1.RegisterRequest],
) (*connect.Response[authv1.RegisterResponse], error) {

	user, err := h.service.RegisterUser(ctx,
		req.Msg.Email,
		req.Msg.Username,
		req.Msg.Password,
		req.Msg.DisplayName,
	)

	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&authv1.RegisterResponse{
		UserId: user.ID.String(),
		AccessToken:  user.AccessToken,
		RefreshToken: user.RefreshToken,
	}), nil
}

// Login handles user login
func (s *AuthHandler) Login(ctx context.Context, req *connect.Request[authv1.LoginRequest]) (*connect.Response[authv1.LoginResponse], error) {
	// Validate input
	if req.Msg.Email == "" || req.Msg.Password == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("email and password are required"))
	}

	// Get user
	user, err := s.repo.GetUserByEmail(ctx, req.Msg.Email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("invalid email or password"))
		}
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to get user"))
	}

	// Check if user is active
	if !user.IsActive {
		return nil, connect.NewError(connect.CodePermissionDenied, errors.New("account is deactivated"))
	}

	// Verify password
	if !handler2.ComparePassword(user.HashedPassword, req.Msg.Password) {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("invalid email or password"))
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
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to generate tokens"))
	}

	// Store refresh token session
	hashedRefreshToken := hashToken(tokens.RefreshToken)
	userAgent := req.Header().Get("User-Agent")
	clientIP := req.Peer().Addr // Get client IP
	_, err = s.repo.CreateUserSession(ctx, user.ID, hashedRefreshToken, userAgent, clientIP, tokens.ExpiresAt.Add(7*24*time.Hour))
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to create session"))
	}

	return connect.NewResponse(&authv1.LoginResponse{
		User: &userv1.User{
			UserId:      user.ID.String(),
			Email:       user.Email,
			Username:    user.Username,
			DisplayName: user.DisplayName,
			Role:        user.Role,
			IsActive:    user.IsActive,
			CreatedAt:   toProtoTimestamp(user.CreatedAt),
		},
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresAt:    toProtoTimestamp(tokens.ExpiresAt),
	}), nil
}

// Logout handles user logout
func (s *AuthHandler) Logout(ctx context.Context, req *connect.Request[authv1.LogoutRequest]) (*connect.Response[authv1.LogoutResponse], error) {
	if req.Msg.RefreshToken == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("refresh token is required"))
	}

	// Delete the session
	hashedToken := hashToken(req.Msg.RefreshToken)
	if err := s.repo.DeleteUserSession(ctx, hashedToken); err != nil {
		// Don't fail if session doesn't exist
		if !errors.Is(err, repository.ErrSessionNotFound) {
			return nil, connect.NewError(connect.CodeInternal, errors.New("failed to delete session"))
		}
	}

	return connect.NewResponse(&authv1.LogoutResponse{
		Success: true,
		Message: "Logged out successfully",
	}), nil
}

// RefreshToken handles token refresh
func (s *AuthHandler) RefreshToken(ctx context.Context, req *connect.Request[authv1.RefreshTokenRequest]) (*connect.Response[authv1.RefreshTokenResponse], error) {
	if req.Msg.RefreshToken == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("refresh token is required"))
	}

	// Validate refresh token
	claims, err := s.tokenManager.ValidateRefreshToken(req.Msg.RefreshToken)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("invalid refresh token"))
	}

	// Verify session exists
	hashedToken := hashToken(req.Msg.RefreshToken)
	session, err := s.repo.GetUserSessionByToken(ctx, hashedToken)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("session not found or expired"))
	}

	// Get user to ensure still active
	userID, _ := uuid.Parse(claims.UserID)
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("user not found"))
	}

	if !user.IsActive {
		return nil, connect.NewError(connect.CodePermissionDenied, errors.New("account is deactivated"))
	}

	// Generate new token pair
	tokens, err := s.tokenManager.GenerateTokenPair(
		user.ID.String(),
		user.Email,
		user.Username,
		user.Role,
	)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to generate tokens"))
	}

	// Delete old session and create new one
	s.repo.DeleteUserSession(ctx, hashedToken)
	newHashedToken := hashToken(tokens.RefreshToken)
	userAgent := req.Header().Get("User-Agent")
	clientIP := req.Peer().Addr
	_, err = s.repo.CreateUserSession(ctx, user.ID, newHashedToken, userAgent, clientIP, tokens.ExpiresAt.Add(7*24*time.Hour))
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to create session"))
	}

	return connect.NewResponse(&authv1.RefreshTokenResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresAt:    toProtoTimestamp(tokens.ExpiresAt),
	}), nil
}

// ValidateToken validates an access token
func (s *AuthHandler) ValidateToken(ctx context.Context, req *connect.Request[authv1.ValidateTokenRequest]) (*connect.Response[authv1.ValidateTokenResponse], error) {
	if req.Msg.AccessToken == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("access token is required"))
	}

	claims, err := s.tokenManager.ValidateAccessToken(req.Msg.AccessToken)
	if err != nil {
		return connect.NewResponse(&authv1.ValidateTokenResponse{
			Valid: false,
		}), nil
	}

	return connect.NewResponse(&authv1.ValidateTokenResponse{
		Valid:  true,
		UserId: claims.UserID,
		Email:  claims.Email,
		Role:   claims.Role,
	}), nil
}

// Continue in next message...
