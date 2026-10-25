-- Create app_releases table for internal auto-updates
CREATE TABLE IF NOT EXISTS "public"."app_releases" (
  "id" serial NOT NULL,
  "app_name" character varying(50) NOT NULL,
  "version" character varying(50) NOT NULL,
  "channel" character varying(50) NOT NULL DEFAULT 'stable',
  "release_notes" text NULL,
  "file_name" character varying(255) NOT NULL,
  "file_path" character varying(500) NOT NULL,
  "file_size_bytes" bigint NOT NULL,
  "checksum_sha256" character varying(64) NOT NULL,
  "is_mandatory" boolean NOT NULL DEFAULT true,
  "min_supported_version" character varying(50) NULL,
  "is_active" boolean NOT NULL DEFAULT true,
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id")
);

CREATE INDEX IF NOT EXISTS "idx_app_releases_lookup" ON "public"."app_releases" ("app_name", "channel", "is_active", "created_at" DESC);
