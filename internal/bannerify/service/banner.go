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

func (s *banner) GetBanner(ctx context.Context, tagID int, featureID int, isAdmin bool, dbRequired bool) (string, error) {
	banner, err := s.repo.GetBanner(ctx, tagID, featureID, isAdmin, dbRequired)
	if err != nil {
		return "", fmt.Errorf("service.GetBanner: %w", err)
	}

	return banner, nil
}

func (s *banner) ListBanners(ctx context.Context, tagID *int, featureID *int, limit int, offset int) ([]domain.BannerListElement, error) {
	banners, err := s.repo.ListBanners(ctx, tagID, featureID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("service.ListBanners: %w", err)
	}

	return banners, nil
}

func (s *banner) ListVersions(ctx context.Context, bannerID int, limit int, offset int) ([]domain.VersionListElement, error) {
	versions, err := s.repo.ListVersions(ctx, bannerID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("service.ListVersions: %w", err)
	}

	return versions, nil
}

func (s *banner) ChooseVersion(ctx context.Context, bannerID int, versionID int) error {
	err := s.repo.ChooseVersion(ctx, bannerID, versionID)
	if err != nil {
		return fmt.Errorf("service.ChooseVersion: %w", err)
	}

	return nil
}

func (s *banner) CreateBanner(ctx context.Context, banner domain.Banner) (int, error) {
	bannerID, err := s.repo.CreateBanner(ctx, banner)
	if err != nil {
		return 0, fmt.Errorf("service.CreateBanner: %w", err)
	}

	return bannerID, nil
}

func (s *banner) UpdateBanner(ctx context.Context, bannerID int, banner domain.Banner) error {
	err := s.repo.UpdateBanner(ctx, bannerID, banner)
	if err != nil {
		return fmt.Errorf("service.UpdateBanner: %w", err)
	}

	return nil
}

func (s *banner) DeleteBannerByID(ctx context.Context, bannerID int) error {
	err := s.repo.DeleteBannerByID(ctx, bannerID)
	if err != nil {
		return fmt.Errorf("service.DeleteBannerByID: %w", err)
	}

	return nil
}

func (s *banner) DeleteBannerByTagOrFeature(ctx context.Context, deleteCtx context.Context, tagID *int, featureID *int) error {
	err := s.repo.DeleteBannerByTagOrFeature(ctx, deleteCtx, tagID, featureID)
	if err != nil {
		return fmt.Errorf("service.DeleteBannerByTagOrFeature: %w", err)
	}

	return nil
}
