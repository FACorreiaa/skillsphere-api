package handler

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	authv1 "github.com/FACorreiaa/skillsphere-proto/gen/go/auth/v1"

	"github.com/FACorreiaa/skillsphere-api/pkg/interceptors"
)

// OAuthLogin is intentionally not implemented since OAuth flows happen via HTTP redirects.
func (h *AuthHandler) OAuthLogin(ctx context.Context, req *connect.Request[authv1.OAuthLoginRequest]) (*connect.Response[authv1.OAuthLoginResponse], error) {
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
