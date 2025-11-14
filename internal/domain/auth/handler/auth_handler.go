package handler

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	authv1 "github.com/FACorreiaa/skillsphere-proto/gen/go/auth/v1"
	pb "github.com/FACorreiaa/skillsphere-proto/gen/go/auth/v1/authv1connect"

	"github.com/FACorreiaa/skillsphere-api/internal/domain/auth/common"
	"github.com/FACorreiaa/skillsphere-api/internal/domain/auth/presenter"
	"github.com/FACorreiaa/skillsphere-api/internal/domain/auth/service"
	"github.com/FACorreiaa/skillsphere-api/pkg/interceptors"
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

	return connect.NewResponse(presenter.RegisterResponse(result)), nil
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

	return connect.NewResponse(presenter.LoginResponse(result)), nil
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

	return connect.NewResponse(presenter.RefreshTokenResponse(tokens)), nil
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
		resp.ExpiresAt = presenter.Timestamp(claims.ExpiresAt.Time)
	}

	return connect.NewResponse(resp), nil
}

func metadataFromRequest[T any](req *connect.Request[T]) service.SessionMetadata {
	return service.SessionMetadata{
		UserAgent: req.Header().Get("User-Agent"),
		ClientIP:  req.Peer().Addr,
	}
}

func (h *AuthHandler) toConnectError(err error) error {
	switch {
	case errors.Is(err, common.ErrUserAlreadyExists):
		return connect.NewError(connect.CodeAlreadyExists, err)
	case errors.Is(err, common.ErrInvalidCredentials):
		return connect.NewError(connect.CodeUnauthenticated, err)
	case errors.Is(err, common.ErrInvalidToken), errors.Is(err, common.ErrSessionNotFound):
		return connect.NewError(connect.CodeUnauthenticated, err)
	case errors.Is(err, common.ErrUserNotFound):
		return connect.NewError(connect.CodeNotFound, err)
	case errors.Is(err, service.ErrAccountInactive):
		return connect.NewError(connect.CodePermissionDenied, err)
	case errors.Is(err, service.ErrPasswordTooShort),
		errors.Is(err, service.ErrPasswordNoDigit),
		errors.Is(err, service.ErrPasswordNoLowercase),
		errors.Is(err, service.ErrPasswordNoUppercase),
		errors.Is(err, service.ErrPasswordNoSpecial):
		return connect.NewError(connect.CodeInvalidArgument, err)
	default:
		return connect.NewError(connect.CodeInternal, err)
	}
}

// OAuthLogin is intentionally not implemented since OAuth flows happen via HTTP redirects.
func (h *AuthHandler) OAuthLogin(_ context.Context, _ *connect.Request[authv1.OAuthLoginRequest]) (*connect.Response[authv1.OAuthLoginResponse], error) {
	return nil, connect.NewError(
		connect.CodeUnimplemented,
		errors.New("oauth login must be completed using the HTTP redirect endpoints"),
	)
}

// RequestPasswordReset triggers a password reset email when possible.
func (h *AuthHandler) RequestPasswordReset(ctx context.Context, req *connect.Request[authv1.RequestPasswordResetRequest]) (*connect.Response[authv1.RequestPasswordResetResponse], error) {
	if req.Msg.Email == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("email is required"))
	}

	if err := h.service.RequestPasswordReset(ctx, req.Msg.Email); err != nil {
		return nil, h.toConnectError(err)
	}

	return connect.NewResponse(&authv1.RequestPasswordResetResponse{
		Success: true,
		Message: "If an account exists for this email, a reset link has been sent",
	}), nil
}

// ResetPassword updates a password using a reset token.
func (h *AuthHandler) ResetPassword(ctx context.Context, req *connect.Request[authv1.ResetPasswordRequest]) (*connect.Response[authv1.ResetPasswordResponse], error) {
	if req.Msg.ResetToken == "" || req.Msg.NewPassword == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("reset token and new password are required"))
	}

	if err := h.service.ResetPassword(ctx, req.Msg.ResetToken, req.Msg.NewPassword); err != nil {
		return nil, h.toConnectError(err)
	}

	return connect.NewResponse(&authv1.ResetPasswordResponse{
		Success: true,
		Message: "Password reset successfully",
	}), nil
}

// ChangePassword updates the password for an authenticated user.
func (h *AuthHandler) ChangePassword(ctx context.Context, req *connect.Request[authv1.ChangePasswordRequest]) (*connect.Response[authv1.ChangePasswordResponse], error) {
	if req.Msg.CurrentPassword == "" || req.Msg.NewPassword == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("current password and new password are required"))
	}

	userID := req.Msg.UserId
	if claims, err := interceptors.GetClaimsFromContext(ctx); err == nil && claims != nil && claims.UserID != "" {
		userID = claims.UserID
	}
	if userID == "" {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("authentication required"))
	}

	if err := h.service.ChangePassword(ctx, userID, req.Msg.CurrentPassword, req.Msg.NewPassword); err != nil {
		return nil, h.toConnectError(err)
	}

	return connect.NewResponse(&authv1.ChangePasswordResponse{
		Success: true,
	}), nil
}

// VerifyEmail confirms the user's email using a verification token.
func (h *AuthHandler) VerifyEmail(ctx context.Context, req *connect.Request[authv1.VerifyEmailRequest]) (*connect.Response[authv1.VerifyEmailResponse], error) {
	if req.Msg.VerificationToken == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("verification token is required"))
	}

	userID, err := h.service.VerifyEmail(ctx, req.Msg.VerificationToken)
	if err != nil {
		return nil, h.toConnectError(err)
	}

	return connect.NewResponse(&authv1.VerifyEmailResponse{
		Success: true,
		UserId:  userID.String(),
	}), nil
}

// ResendVerificationEmail sends another email verification message.
func (h *AuthHandler) ResendVerificationEmail(ctx context.Context, req *connect.Request[authv1.ResendVerificationEmailRequest]) (*connect.Response[authv1.ResendVerificationEmailResponse], error) {
	if req.Msg.Email == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("email is required"))
	}

	result, err := h.service.ResendVerificationEmail(ctx, req.Msg.Email)
	if err != nil {
		return nil, h.toConnectError(err)
	}

	message := "Verification email sent"
	if result != nil && result.AlreadyVerified {
		message = "Email address is already verified"
	}

	return connect.NewResponse(&authv1.ResendVerificationEmailResponse{
		Success: true,
		Message: message,
	}), nil
}
