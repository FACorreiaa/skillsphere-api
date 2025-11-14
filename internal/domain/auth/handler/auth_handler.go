package handler

import (
	"context"
	"errors"
	"time"

	"connectrpc.com/connect"
	authv1 "github.com/FACorreiaa/skillsphere-proto/gen/go/auth/v1"
	pb "github.com/FACorreiaa/skillsphere-proto/gen/go/auth/v1/authv1connect"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/FACorreiaa/skillsphere-api/internal/domain/auth/common"
	"github.com/FACorreiaa/skillsphere-api/internal/domain/auth/repository"
	"github.com/FACorreiaa/skillsphere-api/internal/domain/auth/service"
)

// AuthHandler implements the AuthService Connect handlers.
type AuthHandler struct {
	pb.UnimplementedAuthServiceHandler
	service *service.AuthService
}

// NewAuthHandler constructs a new handler.
func NewAuthHandler(svc *service.AuthService) *AuthHandler {
	return &AuthHandler{
		service: svc,
	}
}

// Register handles user registration RPCs.
func (h *AuthHandler) Register(
	ctx context.Context,
	req *connect.Request[authv1.RegisterRequest],
) (*connect.Response[authv1.RegisterResponse], error) {
	if req.Msg.Email == "" || req.Msg.Password == "" || req.Msg.Username == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("email, username, and password are required"))
	}

	result, err := h.service.RegisterUser(ctx, service.RegisterParams{
		Email:       req.Msg.Email,
		Username:    req.Msg.Username,
		Password:    req.Msg.Password,
		DisplayName: req.Msg.DisplayName,
		Metadata:    metadataFromRequest(req),
	})
	if err != nil {
		return nil, h.toConnectError(err)
	}

	resp := &authv1.RegisterResponse{
		UserId:                    result.User.ID.String(),
		AccessToken:               result.Tokens.AccessToken,
		RefreshToken:              result.Tokens.RefreshToken,
		ExpiresAt:                 toProtoTimestamp(result.Tokens.ExpiresAt),
		EmailVerificationRequired: result.EmailVerificationRequired,
	}

	return connect.NewResponse(resp), nil
}

// Login authenticates a user.
func (h *AuthHandler) Login(ctx context.Context, req *connect.Request[authv1.LoginRequest]) (*connect.Response[authv1.LoginResponse], error) {
	if req.Msg.Email == "" || req.Msg.Password == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("email and password are required"))
	}

	result, err := h.service.Login(ctx, service.LoginParams{
		Email:    req.Msg.Email,
		Password: req.Msg.Password,
		Metadata: metadataFromRequest(req),
	})
	if err != nil {
		return nil, h.toConnectError(err)
	}

	resp := &authv1.LoginResponse{
		UserId:       result.User.ID.String(),
		AccessToken:  result.Tokens.AccessToken,
		RefreshToken: result.Tokens.RefreshToken,
		ExpiresAt:    toProtoTimestamp(result.Tokens.ExpiresAt),
		User:         toUserProfile(result.User),
	}

	return connect.NewResponse(resp), nil
}

// Logout deletes the refresh token session.
func (h *AuthHandler) Logout(ctx context.Context, req *connect.Request[authv1.LogoutRequest]) (*connect.Response[authv1.LogoutResponse], error) {
	if req.Msg.RefreshToken == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("refresh token is required"))
	}

	if err := h.service.Logout(ctx, req.Msg.RefreshToken); err != nil {
		return nil, h.toConnectError(err)
	}

	return connect.NewResponse(&authv1.LogoutResponse{
		Success: true,
	}), nil
}

// RefreshToken issues new access/refresh tokens.
func (h *AuthHandler) RefreshToken(ctx context.Context, req *connect.Request[authv1.RefreshTokenRequest]) (*connect.Response[authv1.RefreshTokenResponse], error) {
	if req.Msg.RefreshToken == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("refresh token is required"))
	}

	tokens, err := h.service.RefreshTokens(ctx, service.RefreshTokenParams{
		RefreshToken: req.Msg.RefreshToken,
		Metadata:     metadataFromRequest(req),
	})
	if err != nil {
		return nil, h.toConnectError(err)
	}

	return connect.NewResponse(&authv1.RefreshTokenResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresAt:    toProtoTimestamp(tokens.ExpiresAt),
	}), nil
}

// ValidateToken validates the provided access token.
func (h *AuthHandler) ValidateToken(ctx context.Context, req *connect.Request[authv1.ValidateTokenRequest]) (*connect.Response[authv1.ValidateTokenResponse], error) {
	if req.Msg.AccessToken == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("access token is required"))
	}

	claims, err := h.service.ValidateAccessToken(ctx, req.Msg.AccessToken)
	if err != nil {
		return connect.NewResponse(&authv1.ValidateTokenResponse{
			IsValid: false,
		}), nil
	}

	resp := &authv1.ValidateTokenResponse{
		IsValid: true,
		UserId:  claims.UserID,
	}
	if claims.ExpiresAt != nil {
		resp.ExpiresAt = timestamppb.New(claims.ExpiresAt.Time)
	}

	return connect.NewResponse(resp), nil
}

func toUserProfile(user *repository.User) *authv1.UserProfile {
	if user == nil {
		return nil
	}

	profile := &authv1.UserProfile{
		UserId:      user.ID.String(),
		Email:       user.Email,
		Username:    user.Username,
		DisplayName: user.DisplayName,
	}

	if user.AvatarURL != nil {
		profile.AvatarUrl = *user.AvatarURL
	}
	if user.EmailVerifiedAt != nil {
		profile.IsVerified = true
	}

	return profile
}

func metadataFromRequest[T any](req *connect.Request[T]) service.SessionMetadata {
	return service.SessionMetadata{
		UserAgent: req.Header().Get("User-Agent"),
		ClientIP:  req.Peer().Addr,
	}
}

func toProtoTimestamp(t time.Time) *timestamppb.Timestamp {
	if t.IsZero() {
		return nil
	}
	return timestamppb.New(t)
}

func (h *AuthHandler) toConnectError(err error) error {
	code := connect.CodeInternal
	switch {
	case errors.Is(err, common.ErrUserAlreadyExists):
		code = connect.CodeAlreadyExists
	case errors.Is(err, common.ErrInvalidCredentials):
		code = connect.CodeUnauthenticated
	case errors.Is(err, common.ErrInvalidToken), errors.Is(err, common.ErrSessionNotFound):
		code = connect.CodeUnauthenticated
	case errors.Is(err, common.ErrUserNotFound):
		code = connect.CodeNotFound
	case errors.Is(err, service.ErrAccountInactive):
		code = connect.CodePermissionDenied
	case errors.Is(err, service.ErrPasswordTooShort),
		errors.Is(err, service.ErrPasswordNoDigit),
		errors.Is(err, service.ErrPasswordNoLowercase),
		errors.Is(err, service.ErrPasswordNoUppercase),
		errors.Is(err, service.ErrPasswordNoSpecial):
		code = connect.CodeInvalidArgument
	default:
		code = connect.CodeInternal
	}
	return connect.NewError(code, err)
}
