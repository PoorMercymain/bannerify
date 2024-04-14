BEGIN;

DROP INDEX IF EXISTS idx_banners_chosen_version_id;

DROP INDEX IF EXISTS idx_banner_versions_banner_id_version_id;

DROP INDEX IF EXISTS idx_banner_versions_feature;

DROP INDEX IF EXISTS idx_banner_versions_updated_at;

DROP INDEX IF EXISTS idx_banner_version_tags_tag;

COMMIT;