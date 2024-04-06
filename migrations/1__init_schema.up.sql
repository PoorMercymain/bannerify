BEGIN;

CREATE TABLE IF NOT EXISTS banners  (
    id SERIAL PRIMARY KEY,
    banner JSONB
);

CREATE TABLE IF NOT EXISTS feature_tag_banner (
    feature INTEGER,
    tag INTEGER,
    banner_id INTEGER,
    FOREIGN KEY(banner_id) REFERENCES banners(id) ON DELETE CASCADE,
    PRIMARY KEY(feature, tag)
);

COMMIT;