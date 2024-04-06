package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	appErrors "github.com/PoorMercymain/bannerify/errors"
	"github.com/PoorMercymain/bannerify/internal/bannerify/domain"
)

var (
	_ domain.AuthorizationRepository = (*autorization)(nil)
)

type autorization struct {
	db *postgres
}

func NewAuthorization(pg *postgres) *autorization {
	return &autorization{db: pg}
}

func (r *autorization) Register(ctx context.Context, login string, passwordHash string, isAdmin bool) error {
	err := r.db.WithTransaction(ctx, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, "INSERT INTO users(login, hash, is_admin) VALUES($1, $2, $3)", login, passwordHash, isAdmin)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
				return appErrors.ErrAlreadyRegistered
			}

			return err
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("repository.Register: %w", err)
	}

	return nil
}

func (r *autorization) GetPasswordHash(ctx context.Context, login string) (string, error) {
	var hash string

	err := r.db.QueryRow(ctx, "SELECT hash FROM users WHERE login = $1", login).Scan(&hash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", fmt.Errorf("repository.GetPasswordHash: %w", appErrors.ErrUserNotFound)
		}

		return "", fmt.Errorf("repository.GetPasswordHash: %w", err)
	}

	return hash, nil
}

func (r *autorization) IsAdmin(ctx context.Context, login string) (bool, error) {
	var isAdmin bool
	err := r.db.QueryRow(ctx, "SELECT is_admin FROM users WHERE login = $1", login).Scan(&isAdmin)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, fmt.Errorf("repository.IsAdmin: %w", appErrors.ErrUserNotFound)
		}

		return false, fmt.Errorf("repository.IsAdmin: %w", err)
	}

	return isAdmin, nil
}
