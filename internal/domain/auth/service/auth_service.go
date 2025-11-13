package service

import (
	"context"

	"github.com/FACorreiaa/skillsphere-api/internal/domain/auth/common"
	"github.com/FACorreiaa/skillsphere-api/internal/domain/auth/repository"
)

type AuthService struct {
	repo repository.AuthRepository
}

func NewAuthService(repo repository.AuthRepository) *AuthService {
	return &AuthService{repo: repo}
}

func (s *AuthService) RegisterUser(ctx context.Context, email, username, password, displayName string) (*repository.User, error) {
	hashedPassword := hashPassword(password) // implement

	return s.repo.CreateUser(ctx, email, username, hashedPassword, displayName)
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*repository.User, error) {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	if !checkPassword(password, user.HashedPassword) {
		return nil, common.ErrInvalidCredentials
	}

	err = s.repo.UpdateLastLogin(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return user, nil
}
