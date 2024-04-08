package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	appErrors "github.com/PoorMercymain/bannerify/errors"
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

func (r *banner) GetBanner(ctx context.Context, tagID int, featureID int) (string, error) {
	const logErrPrefix = "repository.GetBanner: %w"

	conn, err := r.db.Acquire(ctx)
	if err != nil {
		return "", fmt.Errorf(logErrPrefix, err)
	}
	defer conn.Release()

	var data string
	err = conn.QueryRow(ctx, "SELECT bv.data FROM banner_versions bv JOIN banner_version_tags bvt ON bv.version_id = bvt.version_id WHERE bv.is_active = TRUE AND bv.feature = $1 AND bvt.tag = $2 AND bv.banner_id IN (SELECT banner_id FROM banners WHERE chosen_version_id = bv.version_id)", featureID, tagID).Scan(&data)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", fmt.Errorf(logErrPrefix, appErrors.ErrBannerNotFound)
		}

		return "", fmt.Errorf(logErrPrefix, err)
	}

	return data, nil
}

func (r *banner) ListBanners(ctx context.Context, tagID int, featureID int, limit int, offset int) ([]domain.BannerListElement, error) {
	const logErrPrefix = "repository.ListBanners: %w"

	conn, err := r.db.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf(logErrPrefix, err)
	}
	defer conn.Release()

	query := "SELECT bv.banner_id, bv.feature AS feature_id, bv.data, bv.is_active, bv.created_at, bv.updated_at, array_agg(DISTINCT bvt.tag) AS tag_ids FROM banner_versions bv JOIN banners b ON bv.banner_id = b.banner_id AND b.chosen_version_id = bv.version_id LEFT JOIN banner_version_tags bvt ON bv.version_id = bvt.version_id WHERE ($1 = -1 OR bv.feature = $1) GROUP BY bv.banner_id, bv.feature, bv.data, bv.is_active, bv.created_at, bv.updated_at HAVING ($2 = -1 OR bool_or(bvt.tag = $2)) ORDER BY bv.updated_at DESC LIMIT $3 OFFSET $4"

	rows, err := conn.Query(ctx, query, featureID, tagID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf(logErrPrefix, err)
	}
	defer rows.Close()

	var (
		banners              []domain.BannerListElement
		curElem              domain.BannerListElement
		content              string
		createdAt, updatedAt time.Time
	)

	for rows.Next() {
		if err = rows.Scan(&curElem.BannerID, &curElem.FeatureID, &content, &curElem.IsActive, &createdAt, &updatedAt, &curElem.TagIDs); err != nil {
			return nil, fmt.Errorf(logErrPrefix, err)
		}

		curElem.Content = json.RawMessage(content)
		curElem.CreatedAt = createdAt.Format(time.RFC3339)
		curElem.UpdatedAt = updatedAt.Format(time.RFC3339)

		banners = append(banners, curElem)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(logErrPrefix, err)
	}

	return banners, nil
}

func (r *banner) ListVersions(ctx context.Context, bannerID int, limit int, offset int) ([]domain.VersionListElement, error) {
	const logErrPrefix = "repository.ListVersions: %w"

	conn, err := r.db.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf(logErrPrefix, err)
	}
	defer conn.Release()

	query := "SELECT bv.version_id, array_agg(DISTINCT bvt.tag) AS tag_ids, bv.feature, bv.data, bv.is_active, bv.created_at, bv.updated_at, (b.chosen_version_id = bv.version_id) AS is_chosen FROM banner_versions bv LEFT JOIN banner_version_tags bvt ON bv.version_id = bvt.version_id JOIN banners b ON bv.banner_id = b.banner_id WHERE bv.banner_id = $1 GROUP BY bv.version_id, b.chosen_version_id ORDER BY bv.updated_at DESC LIMIT $2 OFFSET $3"

	rows, err := conn.Query(ctx, query, bannerID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf(logErrPrefix, err)
	}
	defer rows.Close()

	var (
		versions             []domain.VersionListElement
		curElem              domain.VersionListElement
		content              string
		createdAt, updatedAt time.Time
	)

	for rows.Next() {
		if err = rows.Scan(&curElem.VersionID, &curElem.TagIDs, &curElem.FeatureID, &content, &curElem.IsActive, &createdAt, &updatedAt, &curElem.IsChosen); err != nil {
			return nil, fmt.Errorf(logErrPrefix, err)
		}

		curElem.Content = json.RawMessage(content)
		curElem.CreatedAt = createdAt.Format(time.RFC3339)
		curElem.UpdatedAt = updatedAt.Format(time.RFC3339)

		versions = append(versions, curElem)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(logErrPrefix, err)
	}

	return versions, nil
}

func (r *banner) ChooseVersion(ctx context.Context, bannerID int, versionID int) error {
	const logErrPrefix = "repository.ChooseVersion: %w"

	err := r.db.WithTransaction(ctx, func(tx pgx.Tx) error {
		tag, err := tx.Exec(ctx, "UPDATE banners SET chosen_version_id = $1 WHERE banner_id = $2", versionID, bannerID)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr); pgErr.Code == pgerrcode.ForeignKeyViolation {
				return appErrors.ErrVersionNotFound
			}

			return err
		}

		if tag.RowsAffected() == 0 {
			return appErrors.ErrBannerNotFound
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf(logErrPrefix, err)
	}

	return nil
}

func (r *banner) CreateBanner(ctx context.Context, banner domain.Banner) (int, error) {
	const logErrPrefix = "repository.CreateBanner: %w"

	var bannerID int
	err := r.db.WithTransaction(ctx, func(tx pgx.Tx) error {
		err := tx.QueryRow(ctx, "INSERT INTO banners DEFAULT VALUES RETURNING banner_id").Scan(&bannerID)
		if err != nil {
			return err
		}

		var versionID int
		err = tx.QueryRow(ctx, "INSERT INTO banner_versions (banner_id, feature, data, is_active) VALUES ($1, $2, $3, $4) RETURNING version_id", bannerID, banner.FeatureID, string(banner.Content), banner.IsActive).Scan(&versionID)
		if err != nil {
			return err
		}

		tag, err := tx.Exec(ctx, "UPDATE banners SET chosen_version_id = $1 WHERE banner_id = $2", versionID, bannerID)
		if err != nil {
			return err
		}

		if tag.RowsAffected() == 0 {
			return appErrors.ErrNoRowsAffected
		}

		for _, tagID := range banner.TagIDs {
			tag, err := tx.Exec(ctx, "INSERT INTO banner_version_tags (version_id, tag) VALUES ($1, $2)", versionID, tagID)
			if err != nil {
				return err
			}

			if tag.RowsAffected() == 0 {
				return appErrors.ErrNoRowsAffected
			}
		}

		return nil
	})

	if err != nil {
		return 0, fmt.Errorf(logErrPrefix, err)
	}

	return bannerID, nil
}
