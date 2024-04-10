package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/sync/singleflight"

	appErrors "github.com/PoorMercymain/bannerify/errors"
	"github.com/PoorMercymain/bannerify/internal/bannerify/domain"
	"github.com/PoorMercymain/bannerify/pkg/logger"
)

var (
	_ domain.BannerRepository = (*banner)(nil)
)

type banner struct {
	db    *postgres
	cache *cache
	sf *singleflight.Group
}

func NewBanner(pg *postgres, cache *cache) *banner {
	return &banner{db: pg, cache: cache, sf: &singleflight.Group{}}
}

func (r *banner) Ping(ctx context.Context) error {
	err := r.db.Ping(ctx)
	if err != nil {
		return fmt.Errorf("repository.Ping: %w", err)
	}

	return nil
}

func (r *banner) GetBanner(ctx context.Context, tagID int, featureID int, isAdmin bool) (string, error) {
	const logErrPrefix = "repository.GetBanner: %w"
	cacheKey := fmt.Sprintf("%d_%d_%t", tagID, featureID, isAdmin)

	res, cacheErr := r.cache.Get(ctx, cacheKey)
	if cacheErr != nil {
		data, err, _ := r.sf.Do(cacheKey, func() (interface{}, error) {
			data, err := r.getBanner(ctx, tagID, featureID, isAdmin, logErrPrefix)
			if err != nil {
				return "", err
			}

			if errors.Is(cacheErr, appErrors.ErrNotFoundInCache) {
				cacheErr := r.cache.Set(ctx, cacheKey, data)
				if cacheErr != nil {
					cacheErr = fmt.Errorf(logErrPrefix, cacheErr)
					logger.Logger().Error(cacheErr.Error())
				}
			} else {
				cacheErr = fmt.Errorf(logErrPrefix, cacheErr)
				logger.Logger().Error(cacheErr.Error())
			}

			return data, nil
		})

		if err != nil {
			return "", err
		}

		return data.(string), nil
	}

	return res, nil
}

func (r *banner) getBanner(ctx context.Context, tagID int, featureID int, isAdmin bool, logErrPrefix string) (string, error) {
	conn, err := r.db.Acquire(ctx)
	if err != nil {
		return "", fmt.Errorf(logErrPrefix, err)
	}
	defer conn.Release()

	var data string
	err = conn.QueryRow(ctx, "SELECT bv.data FROM banner_versions bv JOIN banner_version_tags bvt ON bv.version_id = bvt.version_id WHERE ((bv.is_active = TRUE) OR ($1 = TRUE)) AND bv.feature = $2 AND bvt.tag = $3 AND bv.banner_id IN (SELECT banner_id FROM banners WHERE chosen_version_id = bv.version_id)", isAdmin, featureID, tagID).Scan(&data)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", fmt.Errorf(logErrPrefix, appErrors.ErrBannerNotFound)
		}

		return "", fmt.Errorf(logErrPrefix, err)
	}

	return data, nil
}

func (r *banner) ListBanners(ctx context.Context, tagID *int, featureID *int, limit int, offset int) ([]domain.BannerListElement, error) {
	const logErrPrefix = "repository.ListBanners: %w"

	conn, err := r.db.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf(logErrPrefix, err)
	}
	defer conn.Release()

	query := "SELECT bv.banner_id, bv.feature AS feature_id, bv.data, bv.is_active, bv.created_at, bv.updated_at, array_agg(DISTINCT bvt.tag) AS tag_ids FROM banner_versions bv JOIN banners b ON bv.banner_id = b.banner_id AND b.chosen_version_id = bv.version_id LEFT JOIN banner_version_tags bvt ON bv.version_id = bvt.version_id WHERE ($1::INT IS NULL OR bv.feature = $1::INT) GROUP BY bv.banner_id, bv.feature, bv.data, bv.is_active, bv.created_at, bv.updated_at HAVING ($2::INT IS NULL OR bool_or(bvt.tag = $2::INT)) ORDER BY bv.updated_at DESC LIMIT $3 OFFSET $4"

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

		var featureID int
		err = tx.QueryRow(ctx, "SELECT feature FROM banner_versions WHERE version_id = $1", versionID).Scan(&featureID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return appErrors.ErrVersionNotFound
			}

			return err
		}

		_, err = tx.Exec(ctx, "DELETE FROM chosen_versions WHERE banner_id = $1", bannerID)
		if err != nil {
			return err
		}

		_, err = tx.Exec(ctx, "INSERT INTO chosen_versions (banner_id, version_id, feature, tag) SELECT $1 AS banner_id, $2 AS version_id, $3 AS feature, bvt.tag FROM banner_version_tags bvt WHERE bvt.version_id = $2", bannerID, versionID, featureID)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr); pgErr.Code == pgerrcode.UniqueViolation {
				return appErrors.ErrBannerTagUniqueViolation
			}

			return err
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

			logger.Logger().Infoln(tagID)
			_, err = tx.Exec(ctx, "INSERT INTO chosen_versions (banner_id, version_id, feature, tag) VALUES ($1, $2, $3, $4)", bannerID, versionID, banner.FeatureID, tagID)
			if err != nil {
				var pgErr *pgconn.PgError
				if errors.As(err, &pgErr); pgErr.Code == pgerrcode.UniqueViolation {
					logger.Logger().Infoln(pgErr.Message)
					return appErrors.ErrBannerTagUniqueViolation
				}

				return err
			}
		}

		return nil
	})

	if err != nil {
		return 0, fmt.Errorf(logErrPrefix, err)
	}

	return bannerID, nil
}

func (r *banner) UpdateBanner(ctx context.Context, bannerID int, banner domain.Banner) error {
	const logErrPrefix = "repository.UpdateBanner: %w"

	var versionID int
	err := r.db.WithTransaction(ctx, func(tx pgx.Tx) error {
		err := tx.QueryRow(ctx, "SELECT chosen_version_id FROM banners WHERE banner_id = $1", bannerID).Scan(&versionID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return appErrors.ErrBannerNotFound
			}

			return err
		}

		var newVersionID int
		var contentStr *string
		if banner.Content != nil {
			str := string(banner.Content)
			contentStr = &str
		}

		err = tx.QueryRow(ctx, "INSERT INTO banner_versions (banner_id, feature, data, is_active, created_at, updated_at) SELECT COALESCE($1, bv.banner_id), COALESCE($2, bv.feature), COALESCE($3, bv.data), COALESCE($4, bv.is_active), bv.created_at, CURRENT_TIMESTAMP FROM banner_versions bv WHERE bv.version_id = $5 RETURNING version_id", bannerID, banner.FeatureID, contentStr, banner.IsActive, versionID).Scan(&newVersionID)
		if err != nil {
			return err
		}

		_, err = tx.Exec(ctx, "DELETE FROM chosen_versions WHERE banner_id = $1", bannerID)
		if err != nil {
			return err
		}

		for _, tagID := range banner.TagIDs {
			tag, err := tx.Exec(ctx, "INSERT INTO banner_version_tags (version_id, tag) VALUES ($1, $2)", newVersionID, tagID)
			if err != nil {
				return err
			}

			if tag.RowsAffected() == 0 {
				return appErrors.ErrNoRowsAffected
			}

			_, err = tx.Exec(ctx, "INSERT INTO chosen_versions (banner_id, version_id, feature, tag) SELECT $1, $2, COALESCE($3, bv.feature), $4 FROM banner_versions bv WHERE bv.version_id = $5", bannerID, newVersionID, banner.FeatureID, tagID, versionID)
			if err != nil {
				var pgErr *pgconn.PgError
				if errors.As(err, &pgErr); pgErr.Code == pgerrcode.UniqueViolation {
					return appErrors.ErrBannerTagUniqueViolation
				}

				return err
			}
		}

		tag, err := tx.Exec(ctx, "UPDATE banners SET chosen_version_id = $1 WHERE banner_id = $2", newVersionID, bannerID)
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

func (r *banner) DeleteBannerByID(ctx context.Context, bannerID int) error {
	const logErrPrefix = "repository.DeleteBanner: %w"

	err := r.db.WithTransaction(ctx, func(tx pgx.Tx) error {
		tag, err := tx.Exec(ctx, "DELETE FROM banners WHERE banner_id = $1", bannerID)
		if err != nil {
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

func (r *banner) DeleteBannerByTagOrFeature(ctx context.Context, deleteCtx context.Context, tagID *int, featureID *int, wg *sync.WaitGroup) error {
	const logErrPrefix = "repository.DeleteBannerByTagOrFeature: %w"

	err := r.db.WithTransaction(ctx, func(tx pgx.Tx) error {
		var bannerExists bool
		err := tx.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM chosen_versions cv JOIN banner_versions bv ON cv.version_id = bv.version_id WHERE (cv.feature = $1 OR $1 IS NULL) AND (cv.tag = $2 OR $2 IS NULL) AND bv.banner_id IN (SELECT banner_id FROM banners WHERE chosen_version_id = bv.version_id))", featureID, tagID).Scan(&bannerExists)
		if err != nil {
			return err
		}

		if !bannerExists {
			return appErrors.ErrBannerNotFound
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf(logErrPrefix, err)
	}

	wg.Add(1)

	go func() {
		defer wg.Done()

		err := r.db.WithTransaction(deleteCtx, func(tx pgx.Tx) error {
			tag, err := tx.Exec(deleteCtx, "DELETE FROM banners WHERE banner_id IN (SELECT b.banner_id FROM banners b JOIN banner_versions bv ON b.chosen_version_id = bv.version_id LEFT JOIN banner_version_tags bvt ON bv.version_id = bvt.version_id WHERE ($1::INT IS NULL OR bv.feature = $1::INT) AND ($2::INT IS NULL OR bvt.tag = $2::INT))", featureID, tagID)
			if err != nil {
				return fmt.Errorf(logErrPrefix, err)
			}

			if tag.RowsAffected() == 0 {
				return fmt.Errorf(logErrPrefix, appErrors.ErrNoRowsAffected)
			}

			return nil
		})

		if err != nil {
			logger.Logger().Errorln(err.Error())
		}
	}()

	return nil
}
