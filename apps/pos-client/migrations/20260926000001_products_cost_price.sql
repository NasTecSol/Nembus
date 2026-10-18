-- +goose Up
-- =====================================================================
-- Migration: Add cost_price to products
-- =====================================================================
ALTER TABLE products ADD COLUMN IF NOT EXISTS cost_price DECIMAL(15,4) DEFAULT 0;
