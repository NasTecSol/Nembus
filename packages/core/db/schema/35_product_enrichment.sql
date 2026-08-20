-- =====================================================
-- PRODUCT ENRICHMENT REVIEW QUEUE
-- =====================================================

CREATE TABLE product_enrichment_suggestions (
    id                       SERIAL PRIMARY KEY,
    organization_id         INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    product_id              INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    source_item_code        VARCHAR(100) NOT NULL,
    source_item_name        VARCHAR(255) NOT NULL,
    source_data_fingerprint VARCHAR(128) NOT NULL,
    contract_version        VARCHAR(50) NOT NULL,
    structured_current      JSONB NOT NULL DEFAULT '{}',
    proposed_brand          JSONB,
    proposed_category       JSONB,
    proposed_description    JSONB,
    unsupported_semantics   JSONB,
    source                   VARCHAR(20) NOT NULL DEFAULT 'ai'
        CHECK (source = 'ai'),
    provider                 VARCHAR(100),
    model                    VARCHAR(255),
    model_version            VARCHAR(100),
    status                   VARCHAR(20) NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'processing', 'in_review', 'approved', 'rejected', 'retryable', 'failed', 'applied')),
    reviewer_id              INTEGER REFERENCES users(id) ON DELETE SET NULL,
    reviewed_at              TIMESTAMP,
    applied_at               TIMESTAMP,
    attempt_count            INTEGER NOT NULL DEFAULT 0,
    next_attempt_at          TIMESTAMP,
    last_error_code          VARCHAR(100),
    created_at               TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at               TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_product_enrichment_suggestions_identity
        UNIQUE (organization_id, product_id, source_data_fingerprint, contract_version)
);

CREATE INDEX idx_product_enrichment_suggestions_organization_status
    ON product_enrichment_suggestions (organization_id, status, created_at DESC);
