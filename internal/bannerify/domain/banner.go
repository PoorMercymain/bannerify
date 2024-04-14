package domain

import (
	"context"
)

type BannerServicePingProvider interface {
	Ping(ctx context.Context) error
}

type BannerServiceGetter interface {
	GetBanner(ctx context.Context, tagID int, featureID int, isAdmin bool, dbRequired bool) (string, error)
	ListBanners(ctx context.Context, tagID *int, featureID *int, limit int, offset int) ([]BannerListElement, error)
}

type BannerServiceVersioner interface {
	ListVersions(ctx context.Context, bannerID int, limit int, offset int) ([]VersionListElement, error)
	ChooseVersion(ctx context.Context, bannerID int, versionID int) error
}

type BannerServiceCreator interface {
	CreateBanner(ctx context.Context, banner Banner) (int, error)
}

type BannerServiceUpdater interface {
	UpdateBanner(ctx context.Context, bannerID int, banner Banner) error
}

type BannerServiceDeleter interface {
	DeleteBannerByID(ctx context.Context, bannerID int) error
	DeleteBannerByTagOrFeature(ctx context.Context, deleteCtx context.Context, tagID *int, featureID *int) error
}

type BannerRepositoryPingProvider interface {
	Ping(ctx context.Context) error
}

type BannerRepositoryGetter interface {
	GetBanner(ctx context.Context, tagID int, featureID int, isAdmin bool, dbRequired bool) (string, error)
	ListBanners(ctx context.Context, tagID *int, featureID *int, limit int, offset int) ([]BannerListElement, error)
}

type BannerRepositoryVersioner interface {
	ListVersions(ctx context.Context, bannerID int, limit int, offset int) ([]VersionListElement, error)
	ChooseVersion(ctx context.Context, bannerID int, versionID int) error
}

type BannerRepositoryCreator interface {
	CreateBanner(ctx context.Context, banner Banner) (int, error)
}

type BannerRepositoryUpdater interface {
	UpdateBanner(ctx context.Context, bannerID int, banner Banner) error
}

type BannerRepositoryDeleter interface {
	DeleteBannerByID(ctx context.Context, bannerID int) error
	DeleteBannerByTagOrFeature(ctx context.Context, deleteCtx context.Context, tagID *int, featureID *int) error
}
