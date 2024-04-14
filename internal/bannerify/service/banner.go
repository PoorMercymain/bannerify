package service

import (
	"context"
	"fmt"

	"github.com/PoorMercymain/bannerify/internal/bannerify/domain"
)

var (
	_ domain.BannerServicePingProvider = (*pingProvider)(nil)
)

type pingProvider struct {
	repo domain.BannerRepositoryPingProvider
}

func NewPingProvider(repo domain.BannerRepositoryPingProvider) *pingProvider {
	return &pingProvider{repo: repo}
}

func (s *pingProvider) Ping(ctx context.Context) error {
	err := s.repo.Ping(ctx)
	if err != nil {
		return fmt.Errorf("service.Ping: %w", err)
	}

	return nil
}

var (
	_ domain.BannerServiceGetter = (*bannerGetter)(nil)
)

type bannerGetter struct {
	repo domain.BannerRepositoryGetter
}

func NewGetter(repo domain.BannerRepositoryGetter) *bannerGetter {
	return &bannerGetter{repo: repo}
}

func (s *bannerGetter) GetBanner(ctx context.Context, tagID int, featureID int, isAdmin bool, dbRequired bool) (string, error) {
	banner, err := s.repo.GetBanner(ctx, tagID, featureID, isAdmin, dbRequired)
	if err != nil {
		return "", fmt.Errorf("service.GetBanner: %w", err)
	}

	return banner, nil
}

func (s *bannerGetter) ListBanners(ctx context.Context, tagID *int, featureID *int, limit int, offset int) ([]domain.BannerListElement, error) {
	banners, err := s.repo.ListBanners(ctx, tagID, featureID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("service.ListBanners: %w", err)
	}

	return banners, nil
}

var (
	_ domain.BannerServiceVersioner = (*bannerVersioner)(nil)
)

type bannerVersioner struct {
	repo domain.BannerRepositoryVersioner
}

func NewVersioner(repo domain.BannerRepositoryVersioner) *bannerVersioner {
	return &bannerVersioner{repo: repo}
}

func (s *bannerVersioner) ListVersions(ctx context.Context, bannerID int, limit int, offset int) ([]domain.VersionListElement, error) {
	versions, err := s.repo.ListVersions(ctx, bannerID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("service.ListVersions: %w", err)
	}

	return versions, nil
}

func (s *bannerVersioner) ChooseVersion(ctx context.Context, bannerID int, versionID int) error {
	err := s.repo.ChooseVersion(ctx, bannerID, versionID)
	if err != nil {
		return fmt.Errorf("service.ChooseVersion: %w", err)
	}

	return nil
}

var (
	_ domain.BannerServiceCreator = (*bannerCreator)(nil)
)

type bannerCreator struct {
	repo domain.BannerRepositoryCreator
}

func NewCreator(repo domain.BannerRepositoryCreator) *bannerCreator {
	return &bannerCreator{repo: repo}
}

func (s *bannerCreator) CreateBanner(ctx context.Context, banner domain.Banner) (int, error) {
	bannerID, err := s.repo.CreateBanner(ctx, banner)
	if err != nil {
		return 0, fmt.Errorf("service.CreateBanner: %w", err)
	}

	return bannerID, nil
}

var (
	_ domain.BannerServiceUpdater = (*bannerUpdater)(nil)
)

type bannerUpdater struct {
	repo domain.BannerRepositoryUpdater
}

func NewUpdater(repo domain.BannerRepositoryUpdater) *bannerUpdater {
	return &bannerUpdater{repo: repo}
}

func (s *bannerUpdater) UpdateBanner(ctx context.Context, bannerID int, banner domain.Banner) error {
	err := s.repo.UpdateBanner(ctx, bannerID, banner)
	if err != nil {
		return fmt.Errorf("service.UpdateBanner: %w", err)
	}

	return nil
}

var (
	_ domain.BannerServiceDeleter = (*bannerDeleter)(nil)
)

type bannerDeleter struct {
	repo domain.BannerRepositoryDeleter
}

func NewDeleter(repo domain.BannerRepositoryDeleter) *bannerDeleter {
	return &bannerDeleter{repo: repo}
}

func (s *bannerDeleter) DeleteBannerByID(ctx context.Context, bannerID int) error {
	err := s.repo.DeleteBannerByID(ctx, bannerID)
	if err != nil {
		return fmt.Errorf("service.DeleteBannerByID: %w", err)
	}

	return nil
}

func (s *bannerDeleter) DeleteBannerByTagOrFeature(ctx context.Context, deleteCtx context.Context, tagID *int, featureID *int) error {
	err := s.repo.DeleteBannerByTagOrFeature(ctx, deleteCtx, tagID, featureID)
	if err != nil {
		return fmt.Errorf("service.DeleteBannerByTagOrFeature: %w", err)
	}

	return nil
}
