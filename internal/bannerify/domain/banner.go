package domain

import "context"

type BannerService interface {
	Ping(context.Context) error
}

type BannerRepository interface {
	Ping(context.Context) error
}