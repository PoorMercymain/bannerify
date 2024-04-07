package domain

import (
	"context"
)

type BannerService interface {
	Ping(context.Context) error
	GetBanner(ctx context.Context, tagID int, featureID int) (string, error)
	ListBanners(ctx context.Context, tagID int, featureID int, limit int, offset int) ([]BannerListElement, error)
}

type BannerRepository interface {
	Ping(context.Context) error
	GetBanner(ctx context.Context, tagID int, featureID int) (string, error)
	ListBanners(ctx context.Context, tagID int, featureID int, limit int, offset int) ([]BannerListElement, error)
}
