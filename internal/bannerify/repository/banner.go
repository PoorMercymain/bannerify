package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	appErrors "github.com/PoorMercymain/bannerify/errors"
	"github.com/PoorMercymain/bannerify/internal/bannerify/domain"
	"github.com/jackc/pgx/v5"
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

func (r *banner) GetBanner(ctx context.Context, tagID int, featureID int) (string, error) {
	conn, err := r.db.Acquire(ctx)
	if err != nil {
		return "", fmt.Errorf("repository.GetBanner: %w", err)
	}
	defer conn.Release()

	var bannerID int
	err = conn.QueryRow(ctx, "SELECT banner_id FROM tags_features_banner WHERE tag = $1 AND feature = $2", tagID, featureID).Scan(&bannerID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", fmt.Errorf("repository.GetBanner: %w", appErrors.ErrBannerNotFound)
		}

		return "", fmt.Errorf("repository.GetBanner: %w", err)
	}

	var data string
	err = conn.QueryRow(ctx, "SELECT data FROM banner_versions WHERE banner_id = $1 AND is_active = $2 AND choosen_version = $3", bannerID, true, true).Scan(&data)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", fmt.Errorf("repository.GetBanner: %w", appErrors.ErrBannerNotFound)
		}

		return "", fmt.Errorf("repository.GetBanner: %w", err)
	}

	return data, nil
}

func (r *banner) ListBanners(ctx context.Context, tagID int, featureID int, limit int, offset int) ([]domain.BannerListElement, error) {
	conn, err := r.db.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("repository.ListBanners: %w", err)
	}
	defer conn.Release()

	query := "SELECT bv.banner_id, array_agg(tfb.tag) AS tag_ids, tfb.feature AS feature_id, bv.data AS content, bv.is_active, bv.created_at, bv.updated_at FROM banner_versions bv JOIN tags_features_banner tfb ON bv.banner_id = tfb.banner_id WHERE bv.is_active = TRUE AND bv.chosen_version = TRUE"

	if tagID != -1 {
		query += fmt.Sprintf(" AND tfb.tag = %d", tagID)
	}

	if featureID != -1 {
		query += fmt.Sprintf(" AND tfb.feature = %d", featureID)
	}

	query += " GROUP BY bv.banner_id, tfb.feature, bv.data, bv.is_active, bv.created_at, bv.updated_at ORDER BY bv.updated_at DESC LIMIT $1 OFFSET $2 "


	rows, err := conn.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("repository.ListBanners: %w", err)
	}
	defer rows.Close()

	var (
		banners []domain.BannerListElement
		curElem domain.BannerListElement
		content string
		createdAt, updatedAt time.Time
	)

	for rows.Next() {
		if err = rows.Scan(&curElem.BannerID, &curElem.TagIDs, &curElem.FeatureID, &content, &curElem.IsActive, &createdAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("repository.ListBanners: %w", err)
		}

		curElem.Content = json.RawMessage(content)
		curElem.CreatedAt = createdAt.Format(time.RFC3339)
		curElem.UpdatedAt = updatedAt.Format(time.RFC3339)

		banners = append(banners, curElem)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository.ListBanners: %w", err)
	}

	return banners, nil
}
