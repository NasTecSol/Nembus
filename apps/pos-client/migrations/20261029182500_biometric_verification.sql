-- +goose Up
-- Create "user_biometric_templates" table
CREATE TABLE "public"."user_biometric_templates" (
  "id" serial NOT NULL,
  "organization_id" integer NOT NULL,
  "user_id" integer NOT NULL,
  "finger_index" smallint NOT NULL,
  "fmd_format" character varying(50) NOT NULL DEFAULT 'ISO_19794_2_2005',
  "fmd_base64" text NOT NULL,
  "quality_score" smallint NULL,
  "reader_serial" character varying(128) NULL,
  "enrolled_by_user_id" integer NULL,
  "is_active" boolean NOT NULL DEFAULT true,
  "metadata" jsonb NULL DEFAULT '{}',
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "user_biometric_templates_user_id_finger_index_key" UNIQUE ("user_id", "finger_index"),
  CONSTRAINT "user_biometric_templates_enrolled_by_user_id_fkey" FOREIGN KEY ("enrolled_by_user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "user_biometric_templates_organization_id_fkey" FOREIGN KEY ("organization_id") REFERENCES "public"."organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "user_biometric_templates_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "user_biometric_templates_finger_index_check" CHECK ((finger_index >= 1) AND (finger_index <= 10)),
  CONSTRAINT "user_biometric_templates_quality_score_check" CHECK ((quality_score >= 1) AND (quality_score <= 5))
);
-- Create index "idx_user_biometric_templates_org_active" to table: "user_biometric_templates"
CREATE INDEX "idx_user_biometric_templates_org_active" ON "public"."user_biometric_templates" ("organization_id", "is_active");
-- Create index "idx_user_biometric_templates_updated_at" to table: "user_biometric_templates"
CREATE INDEX "idx_user_biometric_templates_updated_at" ON "public"."user_biometric_templates" ("updated_at");
-- Create index "idx_user_biometric_templates_user" to table: "user_biometric_templates"
CREATE INDEX "idx_user_biometric_templates_user" ON "public"."user_biometric_templates" ("user_id") WHERE (is_active = true);
-- Modify "cashier_sessions" table
ALTER TABLE "public"."cashier_sessions"
  ADD COLUMN "opened_by_supervisor_id" integer NULL REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  ADD COLUMN "closed_by_supervisor_id" integer NULL REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
