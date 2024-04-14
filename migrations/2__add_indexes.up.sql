BEGIN;

CREATE INDEX IF NOT EXISTS idx_banners_chosen_version_id ON banners(chosen_version_id);

CREATE INDEX IF NOT EXISTS idx_banner_versions_banner_id_version_id ON banner_versions(banner_id, version_id);

CREATE INDEX IF NOT EXISTS idx_banner_versions_feature ON banner_versions(feature);

CREATE INDEX IF NOT EXISTS idx_banner_versions_updated_at ON banner_versions(updated_at DESC);

CREATE INDEX IF NOT EXISTS idx_banner_version_tags_tag ON banner_version_tags(tag);

COMMIT;