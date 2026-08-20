-- Migration: durable Stage 2C enrichment execution metadata.
-- Forward-only addition; Stage 1 migration remains unchanged.

ALTER TABLE product_enrichment_suggestions
    ADD COLUMN IF NOT EXISTS attempt_count INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS next_attempt_at TIMESTAMP,
    ADD COLUMN IF NOT EXISTS last_error_code VARCHAR(100);

CREATE INDEX IF NOT EXISTS idx_product_enrichment_suggestions_due
    ON product_enrichment_suggestions (status, next_attempt_at, id)
    WHERE status IN ('pending', 'retryable');
