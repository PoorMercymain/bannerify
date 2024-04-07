BEGIN;

CREATE TABLE IF NOT EXISTS auth (
    login TEXT PRIMARY KEY,
    hash TEXT NOT NULL,
    is_admin BOOLEAN DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS banners (
    banner_id SERIAL PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS banner_versions (
    version_id SERIAL,
    banner_id INT NOT NULL,
    data TEXT NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    chosen_version BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (banner_id) REFERENCES banners(banner_id) ON DELETE CASCADE,
    PRIMARY KEY (version_id, banner_id)
);

CREATE TABLE IF NOT EXISTS tags_features_banner (
    tag INT NOT NULL,
    feature INT NOT NULL,
    banner_id INT NOT NULL,
    PRIMARY KEY (tag, feature),
    FOREIGN KEY (banner_id) REFERENCES banners(banner_id) ON DELETE CASCADE
);
COMMIT;