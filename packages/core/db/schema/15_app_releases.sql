-- =====================================================
-- APPLICATION RELEASES & UPDATES
-- =====================================================

CREATE TABLE IF NOT EXISTS app_releases (
    id SERIAL PRIMARY KEY,
    app_name VARCHAR(50) NOT NULL,
    version VARCHAR(50) NOT NULL,
    channel VARCHAR(50) NOT NULL DEFAULT 'stable',
    release_notes TEXT,
    file_name VARCHAR(255) NOT NULL,
    file_path VARCHAR(500) NOT NULL,
    file_size_bytes BIGINT NOT NULL,
    checksum_sha256 VARCHAR(64) NOT NULL,
    is_mandatory BOOLEAN NOT NULL DEFAULT true,
    min_supported_version VARCHAR(50),
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_app_releases_lookup 
ON app_releases(app_name, channel, is_active, created_at DESC);
