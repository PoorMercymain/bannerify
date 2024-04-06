BEGIN;

CREATE TABLE IF NOT EXISTS users (
    login TEXT PRIMARY KEY,
    hash TEXT,
    is_admin BOOLEAN
);

CREATE TABLE IF NOT EXISTS banner_versions (
    banner_id SERIAL PRIMARY KEY,
    data JSONB NOT NULL,
    banner_version INT NOT NULL,
    is_active BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS tags_features_banner (
    tag INT NOT NULL,
    feature INT NOT NULL,
    banner_id INT NOT NULL,
    PRIMARY KEY (tag, feature),
    FOREIGN KEY (banner_id) REFERENCES banner_versions(banner_id) ON DELETE CASCADE
);
COMMIT;