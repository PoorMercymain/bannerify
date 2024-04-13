BEGIN;

CREATE TABLE IF NOT EXISTS users (
    login TEXT PRIMARY KEY,
    hash TEXT NOT NULL,
    is_admin BOOLEAN DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS banners (
    banner_id SERIAL PRIMARY KEY,
    chosen_version_id INT NULL
);

CREATE TABLE IF NOT EXISTS banner_versions (
    version_id SERIAL PRIMARY KEY,
    banner_id INT NOT NULL,
    feature INT NOT NULL,
    data TEXT NOT NULL,
    is_active BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'fk_chosen_version'
    ) THEN
        ALTER TABLE banners
            ADD CONSTRAINT fk_chosen_version
            FOREIGN KEY (chosen_version_id)
            REFERENCES banner_versions(version_id)
            ON DELETE SET NULL;
    END IF;
END
$$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'fk_banner_id'
    ) THEN
        ALTER TABLE banner_versions
            ADD CONSTRAINT fk_banner_id
            FOREIGN KEY (banner_id)
            REFERENCES banners(banner_id)
            ON DELETE CASCADE;
    END IF;
END
$$;

CREATE TABLE banner_version_tags (
    version_id INT NOT NULL,
    tag INT NOT NULL,
    FOREIGN KEY (version_id) REFERENCES banner_versions(version_id) ON DELETE CASCADE,
    PRIMARY KEY (version_id, tag)
);

CREATE TABLE IF NOT EXISTS chosen_versions (
    banner_id INT NOT NULL,
    version_id INT NOT NULL,
    feature INT NOT NULL,
    tag INT NOT NULL,
    PRIMARY KEY (feature, tag),
    FOREIGN KEY (version_id) REFERENCES banner_versions(version_id) ON DELETE CASCADE
);

COMMIT;