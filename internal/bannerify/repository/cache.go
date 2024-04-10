package repository

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"

	appErrors "github.com/PoorMercymain/bannerify/errors"
)

type cache struct {
	*redis.Client
}

func NewCache(cachePort int) (*cache, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "redis:" + strconv.Itoa(cachePort),
		Password: "",
		DB:       0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("repository.NewCache: %w", err)
	}

	return &cache{rdb}, nil
}

func (c *cache) Get(ctx context.Context, key string) (string, error) {
	res, err := c.Client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", appErrors.ErrNotFoundInCache
		}

		return "", err
	}

	return res, nil
}

func (c *cache) Set(ctx context.Context, key string, value string) error {
	err := c.Client.Set(ctx, key, value, time.Minute*5).Err()
	if err != nil {
		return err
	}

	return nil
}
