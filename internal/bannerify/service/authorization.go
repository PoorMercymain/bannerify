package service

import (
	"context"
	"fmt"

	"golang.org/x/crypto/bcrypt"

	appErrors "github.com/PoorMercymain/bannerify/errors"
	"github.com/PoorMercymain/bannerify/internal/bannerify/domain"
)

var (
	_ domain.AuthorizationService = (*autorization)(nil)
)

type autorization struct {
	repo domain.AuthorizationRepository
}

func NewAuthorization(repo domain.AuthorizationRepository) *autorization {
	return &autorization{repo: repo}
}

func (s *autorization) Register(ctx context.Context, login string, password string, isAdmin bool) error {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		return fmt.Errorf("service.Register: %w", err)
	}

	err = s.repo.Register(ctx, login, string(passwordHash), isAdmin)
	if err != nil {
		return fmt.Errorf("service.Register: %w", err)
	}

	return nil
}

func (s *autorization) CheckAuth(ctx context.Context, login string, password string) error {
	hash, err := s.repo.GetPasswordHash(ctx, login)
	if err != nil {
		return fmt.Errorf("service.CheckAuth: %w", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return fmt.Errorf("service.CheckAuth: %w", appErrors.ErrWrongPassword)
	}

	return nil
}

func (s *autorization) IsAdmin(ctx context.Context, login string) (bool, error) {
	isAdmin, err := s.repo.IsAdmin(ctx, login)
	if err != nil {
		return false, fmt.Errorf("service.IsAdmin: %w", err)
	}

	return isAdmin, nil
}
