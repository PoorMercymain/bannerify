package repository

import (
	"context"
	"fmt"

	"github.com/PoorMercymain/bannerify/internal/bannerify/domain"
)

var (
	_ domain.BannerRepository = (*banner)(nil)
)

type banner struct {
	db *postgres
}

func NewBanner(pg *postgres) *banner {
	return &banner{db: pg}
}

func (r *banner) Ping(ctx context.Context) error {
	err := r.db.Ping(ctx)
	if err != nil {
		return fmt.Errorf("repository.Ping: %w", err)
	}

	return nil
}
