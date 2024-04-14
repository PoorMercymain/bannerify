package domain

import "context"

type AuthorizationService interface {
	Register(ctx context.Context, login string, password string, isAdmin bool) error
	CheckAuth(ctx context.Context, login string, password string) error
	IsAdmin(ctx context.Context, login string) (bool, error)
}

//go:generate mockgen -destination=mocks/authorization_repo_mock.gen.go -package=mocks . AuthorizationRepository
type AuthorizationRepository interface {
	Register(ctx context.Context, login string, passwordHash string, isAdmin bool) error
	GetPasswordHash(ctx context.Context, login string) (string, error)
	IsAdmin(ctx context.Context, login string) (bool, error)
}
