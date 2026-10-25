-- +goose Up
-- 1. Supported Event Types
CREATE TABLE IF NOT EXISTS notification_event_types (
    id VARCHAR(50) PRIMARY KEY,
    category VARCHAR(50) NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Seed default core event types
INSERT INTO notification_event_types (id, category, description) VALUES
    ('po.submitted', 'procurement', 'Purchase Order submitted for review'),
    ('po.approved', 'procurement', 'Purchase Order approved'),
    ('po.rejected', 'procurement', 'Purchase Order rejected'),
    ('grn.posted', 'inventory', 'Goods Receipt Note posted'),
    ('inventory.low_stock', 'inventory', 'Item quantity at or below reorder level'),
    ('pos.receipt', 'sales', 'Customer POS E-Receipt')
ON CONFLICT (id) DO NOTHING;

-- 2. Channel Configurations per Tenant (AWS SES, Meta WhatsApp, WebSocket)
CREATE TABLE IF NOT EXISTS notification_channels (
    id SERIAL PRIMARY KEY,
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    channel_type VARCHAR(30) NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    config JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_tenant_channel UNIQUE (organization_id, channel_type)
);

-- 3. Message Templates (Multi-lingual & Multi-channel)
CREATE TABLE IF NOT EXISTS notification_templates (
    id SERIAL PRIMARY KEY,
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    event_type_id VARCHAR(50) NOT NULL REFERENCES notification_event_types(id),
    channel_type VARCHAR(30) NOT NULL,
    locale VARCHAR(10) NOT NULL DEFAULT 'en',
    subject_template VARCHAR(255),
    body_template TEXT NOT NULL,
    meta_template_name VARCHAR(100),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_tenant_event_template UNIQUE (organization_id, event_type_id, channel_type, locale)
);

-- 4. Notification Routing Rules
CREATE TABLE IF NOT EXISTS notification_rules (
    id SERIAL PRIMARY KEY,
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    event_type_id VARCHAR(50) NOT NULL REFERENCES notification_event_types(id),
    audience_type VARCHAR(30) NOT NULL,
    target_role_id INTEGER REFERENCES roles(id) ON DELETE SET NULL,
    channels JSONB NOT NULL DEFAULT '["websocket"]',
    conditions JSONB DEFAULT '{}',
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 5. The Transactional Outbox (Event Ledger)
CREATE TABLE IF NOT EXISTS notification_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    event_type_id VARCHAR(50) NOT NULL REFERENCES notification_event_types(id),
    entity_type VARCHAR(50) NOT NULL,
    entity_id VARCHAR(100) NOT NULL,
    idempotency_key VARCHAR(150) NOT NULL UNIQUE,
    payload JSONB NOT NULL,
    status VARCHAR(30) NOT NULL DEFAULT 'pending',
    retry_count INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    dispatched_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_notification_events_dispatch ON notification_events (organization_id, status, created_at);

-- 6. In-App User Notifications (Inbox / Bell UI)
CREATE TABLE IF NOT EXISTS user_notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    event_id UUID REFERENCES notification_events(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    action_url VARCHAR(500),
    is_read BOOLEAN DEFAULT FALSE,
    read_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_user_notifications_unread ON user_notifications (user_id, is_read, created_at DESC);

-- 7. Outbound Delivery Audit Log
CREATE TABLE IF NOT EXISTS notification_deliveries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES notification_events(id) ON DELETE CASCADE,
    channel_type VARCHAR(30) NOT NULL,
    recipient VARCHAR(255) NOT NULL,
    status VARCHAR(30) NOT NULL DEFAULT 'pending',
    provider_message_id VARCHAR(150),
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    sent_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_notification_deliveries_event ON notification_deliveries (event_id, channel_type, status);
