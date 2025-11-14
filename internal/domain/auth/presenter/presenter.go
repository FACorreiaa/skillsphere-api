package presenter

import (
	"time"

	"github.com/FACorreiaa/skillsphere-api/internal/domain/auth/repository"
	"github.com/FACorreiaa/skillsphere-api/internal/domain/auth/service"
	authv1 "github.com/FACorreiaa/skillsphere-proto/gen/go/auth/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// RegisterResponse converts a register result into its RPC response.
func RegisterResponse(result *service.RegisterResult) *authv1.RegisterResponse {
	if result == nil {
		return &authv1.RegisterResponse{}
	}

	return &authv1.RegisterResponse{
		UserId:                    result.User.ID.String(),
		AccessToken:               result.Tokens.AccessToken,
		RefreshToken:              result.Tokens.RefreshToken,
		ExpiresAt:                 Timestamp(result.Tokens.ExpiresAt),
		EmailVerificationRequired: result.EmailVerificationRequired,
	}
}

// LoginResponse converts a login result into its RPC response.
func LoginResponse(result *service.LoginResult) *authv1.LoginResponse {
	if result == nil {
		return &authv1.LoginResponse{}
	}

	return &authv1.LoginResponse{
		UserId:       result.User.ID.String(),
		AccessToken:  result.Tokens.AccessToken,
		RefreshToken: result.Tokens.RefreshToken,
		ExpiresAt:    Timestamp(result.Tokens.ExpiresAt),
		User:         userProfile(result.User),
	}
}

// RefreshTokenResponse renders a token pair as RPC response.
func RefreshTokenResponse(tokens *service.TokenPair) *authv1.RefreshTokenResponse {
	if tokens == nil {
		return &authv1.RefreshTokenResponse{}
	}

	return &authv1.RefreshTokenResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresAt:    Timestamp(tokens.ExpiresAt),
	}
}

func userProfile(user *repository.User) *authv1.UserProfile {
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

// Timestamp converts time.Time into protobuf timestamp.
func Timestamp(t time.Time) *timestamppb.Timestamp {
	if t.IsZero() {
		return nil
	}
	return timestamppb.New(t)
}
