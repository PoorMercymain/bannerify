package domain

import (
	"context"
)

type BannerService interface {
	Ping(context.Context) error
	GetBanner(ctx context.Context, tagID int, featureID int, isAdmin bool, dbRequired bool) (string, error)
	ListBanners(ctx context.Context, tagID *int, featureID *int, limit int, offset int) ([]BannerListElement, error)
	ListVersions(ctx context.Context, bannerID int, limit int, offset int) ([]VersionListElement, error)
	ChooseVersion(ctx context.Context, bannerID int, versionID int) error
	CreateBanner(ctx context.Context, banner Banner) (int, error)
	UpdateBanner(ctx context.Context, bannerID int, banner Banner) error
	DeleteBannerByID(ctx context.Context, bannerID int) error
	DeleteBannerByTagOrFeature(ctx context.Context, deleteCtx context.Context, tagID *int, featureID *int) error
}

type BannerRepository interface {
	Ping(context.Context) error
	GetBanner(ctx context.Context, tagID int, featureID int, isAdmin bool, dbRequired bool) (string, error)
	ListBanners(ctx context.Context, tagID *int, featureID *int, limit int, offset int) ([]BannerListElement, error)
	ListVersions(ctx context.Context, bannerID int, limit int, offset int) ([]VersionListElement, error)
	ChooseVersion(ctx context.Context, bannerID int, versionID int) error
	CreateBanner(ctx context.Context, banner Banner) (int, error)
	UpdateBanner(ctx context.Context, bannerID int, banner Banner) error
	DeleteBannerByID(ctx context.Context, bannerID int) error
	DeleteBannerByTagOrFeature(ctx context.Context, deleteCtx context.Context, tagID *int, featureID *int) error
}
