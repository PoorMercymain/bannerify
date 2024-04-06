package service

import (
	"context"
	"fmt"

	"github.com/PoorMercymain/bannerify/internal/bannerify/domain"
)

var (
	_ domain.BannerService = (*banner)(nil)
)

type banner struct {
	repo domain.BannerRepository
}

func NewBanner(repo domain.BannerRepository) *banner {
	return &banner{repo: repo}
}

func (s *banner) Ping(ctx context.Context) error {
	err := s.repo.Ping(ctx)
	if err != nil {
		return fmt.Errorf("service.Ping: %w", err)
	}

	return nil
}
