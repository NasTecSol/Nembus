-- +goose Up
-- Migration: Add granular delta-fetch watermark columns to local_device_config for catalog, pricing, and promotions

ALTER TABLE local_device_config
    ADD COLUMN IF NOT EXISTS last_catalog_sync_at TIMESTAMPTZ DEFAULT '1970-01-01 00:00:00+00',
    ADD COLUMN IF NOT EXISTS last_price_sync_at   TIMESTAMPTZ DEFAULT '1970-01-01 00:00:00+00',
    ADD COLUMN IF NOT EXISTS last_promo_sync_at   TIMESTAMPTZ DEFAULT '1970-01-01 00:00:00+00';

-- Backfill from last_zatca_sync_at if already present
UPDATE local_device_config
SET last_catalog_sync_at = COALESCE(last_zatca_sync_at, '1970-01-01 00:00:00+00'),
    last_price_sync_at   = COALESCE(last_zatca_sync_at, '1970-01-01 00:00:00+00'),
    last_promo_sync_at   = COALESCE(last_zatca_sync_at, '1970-01-01 00:00:00+00')
WHERE last_catalog_sync_at = '1970-01-01 00:00:00+00';

-- +goose Down
ALTER TABLE local_device_config
    DROP COLUMN IF EXISTS last_catalog_sync_at,
    DROP COLUMN IF EXISTS last_price_sync_at,
    DROP COLUMN IF EXISTS last_promo_sync_at;
