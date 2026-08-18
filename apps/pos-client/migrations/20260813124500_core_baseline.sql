-- +goose Up
-- Create extension "uuid-ossp"
CREATE EXTENSION "uuid-ossp" WITH SCHEMA "public" VERSION "1.1";
-- Create enum type "order_type"
CREATE TYPE "public"."order_type" AS ENUM ('standard', 'quote', 'subscription', 'return', 'exchange');
-- Create enum type "order_status_v2"
CREATE TYPE "public"."order_status_v2" AS ENUM ('draft', 'pending', 'confirmed', 'processing', 'partially_fulfilled', 'fulfilled', 'partially_shipped', 'shipped', 'delivered', 'cancelled', 'refunded', 'on_hold');
-- Create enum type "payment_status"
CREATE TYPE "public"."payment_status" AS ENUM ('unpaid', 'partially_paid', 'paid', 'refunded', 'partially_refunded', 'overdue');
-- Create enum type "fulfillment_status"
CREATE TYPE "public"."fulfillment_status" AS ENUM ('unfulfilled', 'partially_fulfilled', 'fulfilled', 'restocked');
-- Create enum type "cart_status"
CREATE TYPE "public"."cart_status" AS ENUM ('draft', 'active', 'abandoned', 'converted', 'expired');
-- Create enum type "cart_type"
CREATE TYPE "public"."cart_type" AS ENUM ('standard', 'quote', 'saved', 'wishlist', 'retail', 'wholesale');
-- Create enum type "invoice_type"
CREATE TYPE "public"."invoice_type" AS ENUM ('standard', 'proforma', 'credit_note', 'debit_note', 'recurring');
-- Create enum type "invoice_status"
CREATE TYPE "public"."invoice_status" AS ENUM ('draft', 'sent', 'viewed', 'partially_paid', 'paid', 'overdue', 'cancelled', 'refunded');
-- Create enum type "quote_status"
CREATE TYPE "public"."quote_status" AS ENUM ('draft', 'sent', 'viewed', 'accepted', 'declined', 'expired', 'converted');
-- Create enum type "zatca_doc_status"
CREATE TYPE "public"."zatca_doc_status" AS ENUM ('pending', 'cleared', 'reported', 'warning', 'rejected', 'failed');
-- Create "update_updated_at_column" function
-- +goose StatementBegin
CREATE FUNCTION "public"."update_updated_at_column" () RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$;
-- +goose StatementEnd
-- Create "organizations" table
CREATE TABLE "public"."organizations" (
  "id" serial NOT NULL,
  "name" character varying(255) NOT NULL,
  "code" character varying(50) NOT NULL,
  "legal_name" character varying(255) NULL,
  "tax_id" character varying(50) NULL,
  "currency_code" character varying(3) NULL DEFAULT 'SAR',
  "fiscal_year_variant" character varying(10) NULL,
  "is_active" boolean NULL DEFAULT true,
  "metadata" jsonb NULL DEFAULT '{}',
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "organizations_code_key" UNIQUE ("code")
);
-- Create index "idx_organizations_code" to table: "organizations"
CREATE INDEX "idx_organizations_code" ON "public"."organizations" ("code");
-- Create index "idx_organizations_is_active" to table: "organizations"
CREATE INDEX "idx_organizations_is_active" ON "public"."organizations" ("is_active");
-- Create "users" table
CREATE TABLE "public"."users" (
  "id" serial NOT NULL,
  "organization_id" integer NOT NULL,
  "username" character varying(100) NOT NULL,
  "email" character varying(255) NOT NULL,
  "password_hash" character varying(255) NOT NULL,
  "first_name" character varying(100) NULL,
  "last_name" character varying(100) NULL,
  "employee_code" character varying(50) NULL,
  "is_active" boolean NULL DEFAULT true,
  "metadata" jsonb NULL DEFAULT '{}',
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "users_email_key" UNIQUE ("email"),
  CONSTRAINT "users_username_key" UNIQUE ("username"),
  CONSTRAINT "users_organization_id_fkey" FOREIGN KEY ("organization_id") REFERENCES "public"."organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_users_email" to table: "users"
CREATE INDEX "idx_users_email" ON "public"."users" ("email");
-- Create index "idx_users_employee_code" to table: "users"
CREATE INDEX "idx_users_employee_code" ON "public"."users" ("employee_code");
-- Create index "idx_users_is_active" to table: "users"
CREATE INDEX "idx_users_is_active" ON "public"."users" ("is_active");
-- Create index "idx_users_organization_id" to table: "users"
CREATE INDEX "idx_users_organization_id" ON "public"."users" ("organization_id");
-- Create index "idx_users_username" to table: "users"
CREATE INDEX "idx_users_username" ON "public"."users" ("username");
-- Create "currencies" table
CREATE TABLE "public"."currencies" (
  "code" character varying(3) NOT NULL,
  "name" character varying(50) NOT NULL,
  "symbol" character varying(10) NOT NULL,
  "decimal_places" integer NULL DEFAULT 2,
  "is_active" boolean NULL DEFAULT true,
  PRIMARY KEY ("code")
);
-- Create "payment_terms" table
CREATE TABLE "public"."payment_terms" (
  "id" serial NOT NULL,
  "organization_id" integer NOT NULL,
  "code" character varying(50) NOT NULL,
  "name" character varying(100) NOT NULL,
  "due_days" integer NOT NULL DEFAULT 0,
  "discount_days" integer NULL DEFAULT 0,
  "discount_percentage" numeric(5,2) NULL DEFAULT 0.00,
  "late_fee_percentage" numeric(5,2) NULL DEFAULT 0.00,
  "is_active" boolean NULL DEFAULT true,
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "payment_terms_code_key" UNIQUE ("code"),
  CONSTRAINT "payment_terms_organization_id_fkey" FOREIGN KEY ("organization_id") REFERENCES "public"."organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create "business_partners" table
CREATE TABLE "public"."business_partners" (
  "id" serial NOT NULL,
  "organization_id" integer NOT NULL,
  "code" character varying(50) NOT NULL,
  "name" character varying(255) NOT NULL,
  "partner_role" character varying(20) NOT NULL,
  "tax_id" character varying(50) NULL,
  "currency_code" character varying(3) NULL DEFAULT 'SAR',
  "credit_limit" numeric(15,2) NULL DEFAULT 0.00,
  "outstanding_balance" numeric(15,2) NULL DEFAULT 0.00,
  "payment_terms_id" integer NULL,
  "sales_rep_user_id" integer NULL,
  "is_active" boolean NULL DEFAULT true,
  "metadata" jsonb NULL DEFAULT '{}',
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "business_partners_code_key" UNIQUE ("code"),
  CONSTRAINT "business_partners_currency_code_fkey" FOREIGN KEY ("currency_code") REFERENCES "public"."currencies" ("code") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "business_partners_organization_id_fkey" FOREIGN KEY ("organization_id") REFERENCES "public"."organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "business_partners_payment_terms_id_fkey" FOREIGN KEY ("payment_terms_id") REFERENCES "public"."payment_terms" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "business_partners_sales_rep_user_id_fkey" FOREIGN KEY ("sales_rep_user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "business_partners_partner_role_check" CHECK ((partner_role)::text = ANY ((ARRAY['supplier'::character varying, 'vendor'::character varying, 'special_customer'::character varying, 'corporate_group'::character varying])::text[]))
);
-- Create index "idx_business_partners_code" to table: "business_partners"
CREATE INDEX "idx_business_partners_code" ON "public"."business_partners" ("code");
-- Create index "idx_business_partners_is_active" to table: "business_partners"
CREATE INDEX "idx_business_partners_is_active" ON "public"."business_partners" ("is_active");
-- Create index "idx_business_partners_organization_id" to table: "business_partners"
CREATE INDEX "idx_business_partners_organization_id" ON "public"."business_partners" ("organization_id");
-- Create index "idx_business_partners_partner_role" to table: "business_partners"
CREATE INDEX "idx_business_partners_partner_role" ON "public"."business_partners" ("partner_role");
-- Create "units_of_measure" table
CREATE TABLE "public"."units_of_measure" (
  "id" serial NOT NULL,
  "code" character varying(20) NOT NULL,
  "name" character varying(50) NOT NULL,
  "uom_type" character varying(20) NULL,
  "decimal_places" integer NULL DEFAULT 2,
  "is_active" boolean NULL DEFAULT true,
  "metadata" jsonb NULL DEFAULT '{}',
  PRIMARY KEY ("id"),
  CONSTRAINT "units_of_measure_code_key" UNIQUE ("code")
);
-- Create index "idx_units_of_measure_code" to table: "units_of_measure"
CREATE INDEX "idx_units_of_measure_code" ON "public"."units_of_measure" ("code");
-- Create index "idx_units_of_measure_uom_type" to table: "units_of_measure"
CREATE INDEX "idx_units_of_measure_uom_type" ON "public"."units_of_measure" ("uom_type");
-- Create "brands" table
CREATE TABLE "public"."brands" (
  "id" serial NOT NULL,
  "name" character varying(255) NOT NULL,
  "code" character varying(50) NOT NULL,
  "is_active" boolean NULL DEFAULT true,
  "metadata" jsonb NULL DEFAULT '{}',
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "brands_code_key" UNIQUE ("code")
);
-- Create index "idx_brands_code" to table: "brands"
CREATE INDEX "idx_brands_code" ON "public"."brands" ("code");
-- Create index "idx_brands_is_active" to table: "brands"
CREATE INDEX "idx_brands_is_active" ON "public"."brands" ("is_active");
-- Create "product_categories" table
CREATE TABLE "public"."product_categories" (
  "id" serial NOT NULL,
  "parent_category_id" integer NULL,
  "name" character varying(255) NOT NULL,
  "code" character varying(50) NOT NULL,
  "description" text NULL,
  "category_level" integer NULL DEFAULT 1,
  "is_active" boolean NULL DEFAULT true,
  "metadata" jsonb NULL DEFAULT '{}',
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "product_categories_code_key" UNIQUE ("code"),
  CONSTRAINT "product_categories_parent_category_id_fkey" FOREIGN KEY ("parent_category_id") REFERENCES "public"."product_categories" ("id") ON UPDATE NO ACTION ON DELETE SET NULL
);
-- Create index "idx_product_categories_code" to table: "product_categories"
CREATE INDEX "idx_product_categories_code" ON "public"."product_categories" ("code");
-- Create index "idx_product_categories_is_active" to table: "product_categories"
CREATE INDEX "idx_product_categories_is_active" ON "public"."product_categories" ("is_active");
-- Create index "idx_product_categories_parent_category_id" to table: "product_categories"
CREATE INDEX "idx_product_categories_parent_category_id" ON "public"."product_categories" ("parent_category_id");
-- Create "tax_categories" table
CREATE TABLE "public"."tax_categories" (
  "id" serial NOT NULL,
  "name" character varying(100) NOT NULL,
  "code" character varying(50) NOT NULL,
  "tax_rate" numeric(5,2) NOT NULL,
  "is_inclusive" boolean NULL DEFAULT false,
  "is_active" boolean NULL DEFAULT true,
  "metadata" jsonb NULL DEFAULT '{}',
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "tax_categories_code_key" UNIQUE ("code")
);
-- Create index "idx_tax_categories_code" to table: "tax_categories"
CREATE INDEX "idx_tax_categories_code" ON "public"."tax_categories" ("code");
-- Create index "idx_tax_categories_is_active" to table: "tax_categories"
CREATE INDEX "idx_tax_categories_is_active" ON "public"."tax_categories" ("is_active");
-- Create "products" table
CREATE TABLE "public"."products" (
  "id" serial NOT NULL,
  "organization_id" integer NOT NULL,
  "sku" character varying(100) NOT NULL,
  "name" character varying(255) NOT NULL,
  "description" text NULL,
  "category_id" integer NULL,
  "brand_id" integer NULL,
  "base_uom_id" integer NULL,
  "product_type" character varying(50) NULL,
  "tax_category_id" integer NULL,
  "is_serialized" boolean NULL DEFAULT false,
  "is_batch_managed" boolean NULL DEFAULT false,
  "is_active" boolean NULL DEFAULT true,
  "is_sellable" boolean NULL DEFAULT true,
  "is_purchasable" boolean NULL DEFAULT true,
  "allow_decimal_quantity" boolean NULL DEFAULT false,
  "track_inventory" boolean NULL DEFAULT true,
  "metadata" jsonb NULL DEFAULT '{}',
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "products_organization_id_sku_key" UNIQUE ("organization_id", "sku"),
  CONSTRAINT "products_base_uom_id_fkey" FOREIGN KEY ("base_uom_id") REFERENCES "public"."units_of_measure" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "products_brand_id_fkey" FOREIGN KEY ("brand_id") REFERENCES "public"."brands" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "products_category_id_fkey" FOREIGN KEY ("category_id") REFERENCES "public"."product_categories" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "products_organization_id_fkey" FOREIGN KEY ("organization_id") REFERENCES "public"."organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "products_tax_category_id_fkey" FOREIGN KEY ("tax_category_id") REFERENCES "public"."tax_categories" ("id") ON UPDATE NO ACTION ON DELETE SET NULL
);
-- Create index "idx_products_active_sellable" to table: "products"
CREATE INDEX "idx_products_active_sellable" ON "public"."products" ("is_active", "is_sellable") WHERE ((is_active = true) AND (is_sellable = true));
-- Create index "idx_products_brand_id" to table: "products"
CREATE INDEX "idx_products_brand_id" ON "public"."products" ("brand_id");
-- Create index "idx_products_category_id" to table: "products"
CREATE INDEX "idx_products_category_id" ON "public"."products" ("category_id");
-- Create index "idx_products_is_active" to table: "products"
CREATE INDEX "idx_products_is_active" ON "public"."products" ("is_active");
-- Create index "idx_products_is_purchasable" to table: "products"
CREATE INDEX "idx_products_is_purchasable" ON "public"."products" ("is_purchasable");
-- Create index "idx_products_is_sellable" to table: "products"
CREATE INDEX "idx_products_is_sellable" ON "public"."products" ("is_sellable");
-- Create index "idx_products_organization_id" to table: "products"
CREATE INDEX "idx_products_organization_id" ON "public"."products" ("organization_id");
-- Create index "idx_products_product_type" to table: "products"
CREATE INDEX "idx_products_product_type" ON "public"."products" ("product_type");
-- Create index "idx_products_sku" to table: "products"
CREATE INDEX "idx_products_sku" ON "public"."products" ("sku");
-- Create index "idx_products_sku_varchar_pattern" to table: "products"
CREATE INDEX "idx_products_sku_varchar_pattern" ON "public"."products" ("sku" varchar_pattern_ops);
-- Create "product_variants" table
CREATE TABLE "public"."product_variants" (
  "id" serial NOT NULL,
  "product_id" integer NOT NULL,
  "variant_sku" character varying(100) NOT NULL,
  "variant_name" character varying(255) NULL,
  "variant_attributes" jsonb NOT NULL,
  "is_active" boolean NULL DEFAULT true,
  "metadata" jsonb NULL DEFAULT '{}',
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "product_variants_variant_sku_key" UNIQUE ("variant_sku"),
  CONSTRAINT "product_variants_product_id_fkey" FOREIGN KEY ("product_id") REFERENCES "public"."products" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_product_variants_is_active" to table: "product_variants"
CREATE INDEX "idx_product_variants_is_active" ON "public"."product_variants" ("is_active");
-- Create index "idx_product_variants_product_id" to table: "product_variants"
CREATE INDEX "idx_product_variants_product_id" ON "public"."product_variants" ("product_id");
-- Create index "idx_product_variants_variant_sku" to table: "product_variants"
CREATE INDEX "idx_product_variants_variant_sku" ON "public"."product_variants" ("variant_sku");
-- Create "bp_price_contracts" table
CREATE TABLE "public"."bp_price_contracts" (
  "id" serial NOT NULL,
  "organization_id" integer NOT NULL,
  "business_partner_id" integer NOT NULL,
  "product_id" integer NOT NULL,
  "product_variant_id" integer NULL,
  "contract_price" numeric(15,4) NOT NULL,
  "discount_percentage" numeric(5,2) NULL DEFAULT 0.00,
  "min_quantity" numeric(15,3) NULL DEFAULT 1,
  "valid_from" date NULL,
  "valid_to" date NULL,
  "is_active" boolean NULL DEFAULT true,
  "notes" text NULL,
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "bp_price_contracts_business_partner_id_product_id_product_v_key" UNIQUE ("business_partner_id", "product_id", "product_variant_id"),
  CONSTRAINT "bp_price_contracts_business_partner_id_fkey" FOREIGN KEY ("business_partner_id") REFERENCES "public"."business_partners" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "bp_price_contracts_organization_id_fkey" FOREIGN KEY ("organization_id") REFERENCES "public"."organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "bp_price_contracts_product_id_fkey" FOREIGN KEY ("product_id") REFERENCES "public"."products" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "bp_price_contracts_product_variant_id_fkey" FOREIGN KEY ("product_variant_id") REFERENCES "public"."product_variants" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_bp_price_contracts_bp_product" to table: "bp_price_contracts"
CREATE INDEX "idx_bp_price_contracts_bp_product" ON "public"."bp_price_contracts" ("business_partner_id", "product_id");
-- Create index "idx_bp_price_contracts_is_active" to table: "bp_price_contracts"
CREATE INDEX "idx_bp_price_contracts_is_active" ON "public"."bp_price_contracts" ("is_active");
-- Create trigger "trg_bp_price_contracts_updated_at"
CREATE TRIGGER "trg_bp_price_contracts_updated_at" BEFORE UPDATE ON "public"."bp_price_contracts" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create trigger "trg_brands_updated_at"
CREATE TRIGGER "trg_brands_updated_at" BEFORE UPDATE ON "public"."brands" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create trigger "update_brands_updated_at"
CREATE TRIGGER "update_brands_updated_at" BEFORE UPDATE ON "public"."brands" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create trigger "trg_business_partners_updated_at"
CREATE TRIGGER "trg_business_partners_updated_at" BEFORE UPDATE ON "public"."business_partners" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create "stores" table
CREATE TABLE "public"."stores" (
  "id" serial NOT NULL,
  "organization_id" integer NOT NULL,
  "parent_store_id" integer NULL,
  "name" character varying(255) NOT NULL,
  "code" character varying(50) NOT NULL,
  "store_type" character varying(50) NULL,
  "is_warehouse" boolean NULL DEFAULT false,
  "is_pos_enabled" boolean NULL DEFAULT false,
  "timezone" character varying(50) NULL DEFAULT 'UTC',
  "is_active" boolean NULL DEFAULT true,
  "metadata" jsonb NULL DEFAULT '{}',
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "stores_organization_id_code_key" UNIQUE ("organization_id", "code"),
  CONSTRAINT "stores_organization_id_fkey" FOREIGN KEY ("organization_id") REFERENCES "public"."organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "stores_parent_store_id_fkey" FOREIGN KEY ("parent_store_id") REFERENCES "public"."stores" ("id") ON UPDATE NO ACTION ON DELETE SET NULL
);
-- Create index "idx_stores_code" to table: "stores"
CREATE INDEX "idx_stores_code" ON "public"."stores" ("code");
-- Create index "idx_stores_is_active" to table: "stores"
CREATE INDEX "idx_stores_is_active" ON "public"."stores" ("is_active");
-- Create index "idx_stores_organization_id" to table: "stores"
CREATE INDEX "idx_stores_organization_id" ON "public"."stores" ("organization_id");
-- Create index "idx_stores_parent_store_id" to table: "stores"
CREATE INDEX "idx_stores_parent_store_id" ON "public"."stores" ("parent_store_id");
-- Create index "idx_stores_store_type" to table: "stores"
CREATE INDEX "idx_stores_store_type" ON "public"."stores" ("store_type");
-- Create "cashiers" table
CREATE TABLE "public"."cashiers" (
  "id" serial NOT NULL,
  "user_id" integer NOT NULL,
  "store_id" integer NOT NULL,
  "cashier_code" character varying(50) NOT NULL,
  "drawer_limit" numeric(15,2) NULL,
  "discount_limit" numeric(5,2) NULL,
  "is_active" boolean NULL DEFAULT true,
  "metadata" jsonb NULL DEFAULT '{}',
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "cashiers_store_id_cashier_code_key" UNIQUE ("store_id", "cashier_code"),
  CONSTRAINT "cashiers_store_id_fkey" FOREIGN KEY ("store_id") REFERENCES "public"."stores" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "cashiers_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "cashiers_discount_limit_check" CHECK ((discount_limit >= (0)::numeric) AND (discount_limit <= (100)::numeric))
);
-- Create index "idx_cashiers_is_active" to table: "cashiers"
CREATE INDEX "idx_cashiers_is_active" ON "public"."cashiers" ("is_active");
-- Create index "idx_cashiers_store_id" to table: "cashiers"
CREATE INDEX "idx_cashiers_store_id" ON "public"."cashiers" ("store_id");
-- Create index "idx_cashiers_user_id" to table: "cashiers"
CREATE INDEX "idx_cashiers_user_id" ON "public"."cashiers" ("user_id");
-- Create "price_lists" table
CREATE TABLE "public"."price_lists" (
  "id" serial NOT NULL,
  "name" character varying(100) NOT NULL,
  "code" character varying(50) NOT NULL,
  "price_list_type" character varying(50) NULL,
  "currency_code" character varying(3) NULL DEFAULT 'USD',
  "valid_from" date NULL,
  "valid_to" date NULL,
  "is_default" boolean NULL DEFAULT false,
  "is_active" boolean NULL DEFAULT true,
  "metadata" jsonb NULL DEFAULT '{}',
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "price_lists_code_key" UNIQUE ("code")
);
-- Create index "idx_price_lists_code" to table: "price_lists"
CREATE INDEX "idx_price_lists_code" ON "public"."price_lists" ("code");
-- Create index "idx_price_lists_is_active" to table: "price_lists"
CREATE INDEX "idx_price_lists_is_active" ON "public"."price_lists" ("is_active");
-- Create index "idx_price_lists_valid_from" to table: "price_lists"
CREATE INDEX "idx_price_lists_valid_from" ON "public"."price_lists" ("valid_from");
-- Create index "idx_price_lists_valid_to" to table: "price_lists"
CREATE INDEX "idx_price_lists_valid_to" ON "public"."price_lists" ("valid_to");
-- Create "customers" table
CREATE TABLE "public"."customers" (
  "id" serial NOT NULL,
  "organization_id" integer NOT NULL,
  "customer_code" character varying(50) NOT NULL,
  "name" character varying(255) NOT NULL,
  "email" character varying(255) NULL,
  "phone" character varying(50) NULL,
  "address" text NULL,
  "customer_type" character varying(50) NULL,
  "price_list_id" integer NULL,
  "credit_limit" numeric(15,2) NULL DEFAULT 0,
  "outstanding_balance" numeric(15,2) NULL DEFAULT 0,
  "loyalty_points" numeric(15,2) NULL DEFAULT 0,
  "is_active" boolean NULL DEFAULT true,
  "metadata" jsonb NULL DEFAULT '{}',
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  "business_partner_id" integer NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "customers_organization_id_customer_code_key" UNIQUE ("organization_id", "customer_code"),
  CONSTRAINT "customers_business_partner_id_fkey" FOREIGN KEY ("business_partner_id") REFERENCES "public"."business_partners" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "customers_organization_id_fkey" FOREIGN KEY ("organization_id") REFERENCES "public"."organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "customers_price_list_id_fkey" FOREIGN KEY ("price_list_id") REFERENCES "public"."price_lists" ("id") ON UPDATE NO ACTION ON DELETE SET NULL
);
-- Create index "idx_customers_business_partner_id" to table: "customers"
CREATE INDEX "idx_customers_business_partner_id" ON "public"."customers" ("business_partner_id");
-- Create index "idx_customers_customer_code" to table: "customers"
CREATE INDEX "idx_customers_customer_code" ON "public"."customers" ("customer_code");
-- Create index "idx_customers_customer_type" to table: "customers"
CREATE INDEX "idx_customers_customer_type" ON "public"."customers" ("customer_type");
-- Create index "idx_customers_is_active" to table: "customers"
CREATE INDEX "idx_customers_is_active" ON "public"."customers" ("is_active");
-- Create index "idx_customers_organization_id" to table: "customers"
CREATE INDEX "idx_customers_organization_id" ON "public"."customers" ("organization_id");
-- Create "pos_terminals" table
CREATE TABLE "public"."pos_terminals" (
  "id" serial NOT NULL,
  "store_id" integer NOT NULL,
  "terminal_code" character varying(50) NOT NULL,
  "terminal_name" character varying(100) NULL,
  "device_id" character varying(100) NULL,
  "is_active" boolean NULL DEFAULT true,
  "metadata" jsonb NULL DEFAULT '{}',
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "pos_terminals_store_id_terminal_code_key" UNIQUE ("store_id", "terminal_code"),
  CONSTRAINT "pos_terminals_store_id_fkey" FOREIGN KEY ("store_id") REFERENCES "public"."stores" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_pos_terminals_is_active" to table: "pos_terminals"
CREATE INDEX "idx_pos_terminals_is_active" ON "public"."pos_terminals" ("is_active");
-- Create index "idx_pos_terminals_store_id" to table: "pos_terminals"
CREATE INDEX "idx_pos_terminals_store_id" ON "public"."pos_terminals" ("store_id");
-- Create "carts" table
CREATE TABLE "public"."carts" (
  "id" uuid NOT NULL DEFAULT public.uuid_generate_v4(),
  "cart_number" character varying(50) NOT NULL,
  "organization_id" integer NOT NULL,
  "store_id" integer NULL,
  "customer_id" integer NULL,
  "guest_identifier" character varying(255) NULL,
  "guest_email" character varying(255) NULL,
  "guest_phone" character varying(50) NULL,
  "cart_status" "public"."cart_status" NOT NULL DEFAULT 'draft',
  "cart_type" "public"."cart_type" NOT NULL DEFAULT 'standard',
  "channel" character varying(50) NULL DEFAULT 'online',
  "payment_method" character varying(100) NULL,
  "payment_gateway" character varying(100) NULL,
  "device_info" jsonb NULL DEFAULT '{}',
  "created_by_user_id" integer NULL,
  "cashier_id" integer NULL,
  "pos_terminal_id" integer NULL,
  "subtotal" numeric(15,2) NULL DEFAULT 0.00,
  "discount_amount" numeric(15,2) NULL DEFAULT 0.00,
  "tax_amount" numeric(15,2) NULL DEFAULT 0.00,
  "shipping_amount" numeric(15,2) NULL DEFAULT 0.00,
  "total_amount" numeric(15,2) NULL DEFAULT 0.00,
  "coupon_code" character varying(100) NULL,
  "discount_code" character varying(100) NULL,
  "promotional_credits" numeric(15,2) NULL DEFAULT 0.00,
  "shipping_address" jsonb NULL,
  "billing_address" jsonb NULL,
  "shipping_method" character varying(100) NULL,
  "converted_to_order_id" uuid NULL,
  "converted_at" timestamp NULL,
  "last_activity_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  "expires_at" timestamp NULL,
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  "metadata" jsonb NULL DEFAULT '{}',
  "notes" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "carts_cart_number_key" UNIQUE ("cart_number"),
  CONSTRAINT "carts_cashier_id_fkey" FOREIGN KEY ("cashier_id") REFERENCES "public"."cashiers" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "carts_created_by_user_id_fkey" FOREIGN KEY ("created_by_user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "carts_customer_id_fkey" FOREIGN KEY ("customer_id") REFERENCES "public"."customers" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "carts_organization_id_fkey" FOREIGN KEY ("organization_id") REFERENCES "public"."organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "carts_pos_terminal_id_fkey" FOREIGN KEY ("pos_terminal_id") REFERENCES "public"."pos_terminals" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "carts_store_id_fkey" FOREIGN KEY ("store_id") REFERENCES "public"."stores" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "chk_cart_customer" CHECK ((customer_id IS NOT NULL) OR (guest_identifier IS NOT NULL))
);
-- Create index "idx_carts_cart_number" to table: "carts"
CREATE INDEX "idx_carts_cart_number" ON "public"."carts" ("cart_number");
-- Create index "idx_carts_cart_status" to table: "carts"
CREATE INDEX "idx_carts_cart_status" ON "public"."carts" ("cart_status");
-- Create index "idx_carts_cart_type" to table: "carts"
CREATE INDEX "idx_carts_cart_type" ON "public"."carts" ("cart_type");
-- Create index "idx_carts_channel" to table: "carts"
CREATE INDEX "idx_carts_channel" ON "public"."carts" ("channel");
-- Create index "idx_carts_created_at" to table: "carts"
CREATE INDEX "idx_carts_created_at" ON "public"."carts" ("created_at");
-- Create index "idx_carts_customer_id" to table: "carts"
CREATE INDEX "idx_carts_customer_id" ON "public"."carts" ("customer_id");
-- Create index "idx_carts_expires_at" to table: "carts"
CREATE INDEX "idx_carts_expires_at" ON "public"."carts" ("expires_at");
-- Create index "idx_carts_guest_identifier" to table: "carts"
CREATE INDEX "idx_carts_guest_identifier" ON "public"."carts" ("guest_identifier");
-- Create index "idx_carts_last_activity_at" to table: "carts"
CREATE INDEX "idx_carts_last_activity_at" ON "public"."carts" ("last_activity_at");
-- Create index "idx_carts_organization_id" to table: "carts"
CREATE INDEX "idx_carts_organization_id" ON "public"."carts" ("organization_id");
-- Create index "idx_carts_store_id" to table: "carts"
CREATE INDEX "idx_carts_store_id" ON "public"."carts" ("store_id");
-- Set comment to table: "carts"
COMMENT ON TABLE "public"."carts" IS 'Shopping carts for online and POS channels, supporting both registered customers and guests';
-- Create "update_cart_activity" function
-- +goose StatementBegin
CREATE FUNCTION "public"."update_cart_activity" () RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    UPDATE carts 
    SET last_activity_at = CURRENT_TIMESTAMP
    WHERE id = COALESCE(NEW.cart_id, OLD.cart_id);
    RETURN COALESCE(NEW, OLD);
END;
$$;
-- +goose StatementEnd
-- Create "cart_items" table
CREATE TABLE "public"."cart_items" (
  "id" uuid NOT NULL DEFAULT public.uuid_generate_v4(),
  "cart_id" uuid NOT NULL,
  "organization_id" integer NOT NULL,
  "product_id" integer NOT NULL,
  "product_variant_id" integer NULL,
  "quantity" numeric(15,3) NOT NULL,
  "uom_id" integer NULL,
  "unit_price" numeric(15,2) NOT NULL,
  "discount_amount" numeric(15,2) NULL DEFAULT 0.00,
  "tax_amount" numeric(15,2) NULL DEFAULT 0.00,
  "line_total" numeric(15,2) NOT NULL,
  "price_list_id" integer NULL,
  "tax_category_id" integer NULL,
  "batch_number" character varying(100) NULL,
  "serial_number" character varying(100) NULL,
  "customization_details" jsonb NULL DEFAULT '{}',
  "notes" text NULL,
  "metadata" jsonb NULL DEFAULT '{}',
  "added_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "cart_items_cart_id_product_id_product_variant_id_batch_numb_key" UNIQUE ("cart_id", "product_id", "product_variant_id", "batch_number", "serial_number"),
  CONSTRAINT "cart_items_cart_id_fkey" FOREIGN KEY ("cart_id") REFERENCES "public"."carts" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "cart_items_organization_id_fkey" FOREIGN KEY ("organization_id") REFERENCES "public"."organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "cart_items_price_list_id_fkey" FOREIGN KEY ("price_list_id") REFERENCES "public"."price_lists" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "cart_items_product_id_fkey" FOREIGN KEY ("product_id") REFERENCES "public"."products" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT,
  CONSTRAINT "cart_items_product_variant_id_fkey" FOREIGN KEY ("product_variant_id") REFERENCES "public"."product_variants" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "cart_items_tax_category_id_fkey" FOREIGN KEY ("tax_category_id") REFERENCES "public"."tax_categories" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "cart_items_uom_id_fkey" FOREIGN KEY ("uom_id") REFERENCES "public"."units_of_measure" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "cart_items_quantity_check" CHECK (quantity > (0)::numeric)
);
-- Create index "idx_cart_items_added_at" to table: "cart_items"
CREATE INDEX "idx_cart_items_added_at" ON "public"."cart_items" ("added_at");
-- Create index "idx_cart_items_cart_id" to table: "cart_items"
CREATE INDEX "idx_cart_items_cart_id" ON "public"."cart_items" ("cart_id");
-- Create index "idx_cart_items_product_id" to table: "cart_items"
CREATE INDEX "idx_cart_items_product_id" ON "public"."cart_items" ("product_id");
-- Create index "idx_cart_items_product_variant_id" to table: "cart_items"
CREATE INDEX "idx_cart_items_product_variant_id" ON "public"."cart_items" ("product_variant_id");
-- Set comment to table: "cart_items"
COMMENT ON TABLE "public"."cart_items" IS 'Line items in shopping carts with pricing and customization details';
-- Create trigger "cart_items_activity_trigger"
CREATE TRIGGER "cart_items_activity_trigger" AFTER DELETE OR INSERT OR UPDATE ON "public"."cart_items" FOR EACH ROW EXECUTE FUNCTION "public"."update_cart_activity"();
-- Create trigger "trg_cart_items_updated_at"
CREATE TRIGGER "trg_cart_items_updated_at" BEFORE UPDATE ON "public"."cart_items" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create "cart_activity_log" table
CREATE TABLE "public"."cart_activity_log" (
  "id" bigserial NOT NULL,
  "cart_id" uuid NOT NULL,
  "organization_id" integer NOT NULL,
  "activity_type" character varying(50) NOT NULL,
  "description" text NULL,
  "performed_by_user_id" integer NULL,
  "ip_address" inet NULL,
  "user_agent" text NULL,
  "old_value" jsonb NULL,
  "new_value" jsonb NULL,
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "cart_activity_log_cart_id_fkey" FOREIGN KEY ("cart_id") REFERENCES "public"."carts" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "cart_activity_log_organization_id_fkey" FOREIGN KEY ("organization_id") REFERENCES "public"."organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "cart_activity_log_performed_by_user_id_fkey" FOREIGN KEY ("performed_by_user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL
);
-- Create index "idx_cart_activity_log_activity_type" to table: "cart_activity_log"
CREATE INDEX "idx_cart_activity_log_activity_type" ON "public"."cart_activity_log" ("activity_type");
-- Create index "idx_cart_activity_log_cart_id" to table: "cart_activity_log"
CREATE INDEX "idx_cart_activity_log_cart_id" ON "public"."cart_activity_log" ("cart_id");
-- Create index "idx_cart_activity_log_created_at" to table: "cart_activity_log"
CREATE INDEX "idx_cart_activity_log_created_at" ON "public"."cart_activity_log" ("created_at");
-- Set comment to table: "cart_activity_log"
COMMENT ON TABLE "public"."cart_activity_log" IS 'Audit trail of all cart activities and changes';
-- Create "log_cart_status_change" function
-- +goose StatementBegin
CREATE FUNCTION "public"."log_cart_status_change" () RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    IF OLD.cart_status IS DISTINCT FROM NEW.cart_status THEN
        INSERT INTO cart_activity_log (
            cart_id, 
            organization_id, 
            activity_type, 
            description, 
            old_value, 
            new_value
        )
        VALUES (
            NEW.id,
            NEW.organization_id,
            'status_changed',
            'Cart status changed from ' || OLD.cart_status || ' to ' || NEW.cart_status,
            jsonb_build_object('status', OLD.cart_status),
            jsonb_build_object('status', NEW.cart_status)
        );
    END IF;
    RETURN NEW;
END;
$$;
-- +goose StatementEnd
-- Create trigger "cart_status_change_trigger"
CREATE TRIGGER "cart_status_change_trigger" AFTER UPDATE ON "public"."carts" FOR EACH ROW WHEN (old.cart_status IS DISTINCT FROM new.cart_status) EXECUTE FUNCTION "public"."log_cart_status_change"();
-- Create trigger "trg_carts_updated_at"
CREATE TRIGGER "trg_carts_updated_at" BEFORE UPDATE ON "public"."carts" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create "cashier_sessions" table
CREATE TABLE "public"."cashier_sessions" (
  "id" serial NOT NULL,
  "cashier_id" integer NOT NULL,
  "pos_terminal_id" integer NOT NULL,
  "session_number" character varying(50) NOT NULL,
  "opening_time" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "closing_time" timestamp NULL,
  "opening_balance" numeric(15,2) NULL DEFAULT 0,
  "closing_balance" numeric(15,2) NULL,
  "expected_balance" numeric(15,2) NULL,
  "variance" numeric(15,2) NULL,
  "status" character varying(20) NULL DEFAULT 'open',
  "metadata" jsonb NULL DEFAULT '{}',
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "cashier_sessions_cashier_id_fkey" FOREIGN KEY ("cashier_id") REFERENCES "public"."cashiers" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "cashier_sessions_pos_terminal_id_fkey" FOREIGN KEY ("pos_terminal_id") REFERENCES "public"."pos_terminals" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_cashier_sessions_cashier_id" to table: "cashier_sessions"
CREATE INDEX "idx_cashier_sessions_cashier_id" ON "public"."cashier_sessions" ("cashier_id");
-- Create index "idx_cashier_sessions_opening_time" to table: "cashier_sessions"
CREATE INDEX "idx_cashier_sessions_opening_time" ON "public"."cashier_sessions" ("opening_time");
-- Create index "idx_cashier_sessions_pos_terminal_id" to table: "cashier_sessions"
CREATE INDEX "idx_cashier_sessions_pos_terminal_id" ON "public"."cashier_sessions" ("pos_terminal_id");
-- Create index "idx_cashier_sessions_status" to table: "cashier_sessions"
CREATE INDEX "idx_cashier_sessions_status" ON "public"."cashier_sessions" ("status");
-- Create trigger "trg_cashier_sessions_updated_at"
CREATE TRIGGER "trg_cashier_sessions_updated_at" BEFORE UPDATE ON "public"."cashier_sessions" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create trigger "trg_cashiers_updated_at"
CREATE TRIGGER "trg_cashiers_updated_at" BEFORE UPDATE ON "public"."cashiers" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create "combo_bundles" table
CREATE TABLE "public"."combo_bundles" (
  "id" serial NOT NULL,
  "store_id" integer NOT NULL,
  "code" character varying(50) NOT NULL,
  "name" character varying(255) NOT NULL,
  "description" text NULL,
  "bundle_price" numeric(15,2) NOT NULL,
  "bundle_type" character varying(30) NULL DEFAULT 'fixed',
  "is_active" boolean NULL DEFAULT true,
  "valid_from" date NULL,
  "valid_to" date NULL,
  "display_order" integer NULL DEFAULT 0,
  "metadata" jsonb NULL DEFAULT '{}',
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "combo_bundles_store_id_code_key" UNIQUE ("store_id", "code"),
  CONSTRAINT "combo_bundles_store_id_fkey" FOREIGN KEY ("store_id") REFERENCES "public"."stores" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "combo_bundles_bundle_type_check" CHECK ((bundle_type)::text = ANY ((ARRAY['fixed'::character varying, 'build_your_own'::character varying, 'meal_deal'::character varying, 'bogo'::character varying])::text[]))
);
-- Create index "idx_combo_bundles_is_active" to table: "combo_bundles"
CREATE INDEX "idx_combo_bundles_is_active" ON "public"."combo_bundles" ("is_active");
-- Create index "idx_combo_bundles_store_id" to table: "combo_bundles"
CREATE INDEX "idx_combo_bundles_store_id" ON "public"."combo_bundles" ("store_id");
-- Create trigger "trg_combo_bundles_updated_at"
CREATE TRIGGER "trg_combo_bundles_updated_at" BEFORE UPDATE ON "public"."combo_bundles" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create trigger "trg_customers_updated_at"
CREATE TRIGGER "trg_customers_updated_at" BEFORE UPDATE ON "public"."customers" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create trigger "update_customers_updated_at"
CREATE TRIGGER "update_customers_updated_at" BEFORE UPDATE ON "public"."customers" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create "discount_analytics" table
CREATE TABLE "public"."discount_analytics" (
  "id" serial NOT NULL,
  "organization_id" integer NOT NULL,
  "store_id" integer NULL,
  "cashier_id" integer NULL,
  "product_id" integer NULL,
  "discount_type" character varying(50) NULL,
  "date" date NOT NULL,
  "month" integer NULL,
  "quarter" integer NULL,
  "year" integer NULL,
  "total_discounts_given" numeric(15,2) NULL DEFAULT 0,
  "transactions_with_discount" integer NULL DEFAULT 0,
  "total_transactions" integer NULL DEFAULT 0,
  "discount_percentage" numeric(5,2) NULL,
  "revenue_impact" numeric(15,2) NULL DEFAULT 0,
  "metadata" jsonb NULL DEFAULT '{}',
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "discount_analytics_organization_id_fkey" FOREIGN KEY ("organization_id") REFERENCES "public"."organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "fk_discount_analytics_cashier" FOREIGN KEY ("cashier_id") REFERENCES "public"."cashiers" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "fk_discount_analytics_product" FOREIGN KEY ("product_id") REFERENCES "public"."products" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "fk_discount_analytics_store" FOREIGN KEY ("store_id") REFERENCES "public"."stores" ("id") ON UPDATE NO ACTION ON DELETE SET NULL
);
-- Create index "idx_discount_analytics_cashier_id" to table: "discount_analytics"
CREATE INDEX "idx_discount_analytics_cashier_id" ON "public"."discount_analytics" ("cashier_id");
-- Create index "idx_discount_analytics_date" to table: "discount_analytics"
CREATE INDEX "idx_discount_analytics_date" ON "public"."discount_analytics" ("date");
-- Create index "idx_discount_analytics_organization_id" to table: "discount_analytics"
CREATE INDEX "idx_discount_analytics_organization_id" ON "public"."discount_analytics" ("organization_id");
-- Create index "idx_discount_analytics_store_id" to table: "discount_analytics"
CREATE INDEX "idx_discount_analytics_store_id" ON "public"."discount_analytics" ("store_id");
-- Create trigger "trg_discount_analytics_updated_at"
CREATE TRIGGER "trg_discount_analytics_updated_at" BEFORE UPDATE ON "public"."discount_analytics" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create trigger "update_discount_analytics_updated_at"
CREATE TRIGGER "update_discount_analytics_updated_at" BEFORE UPDATE ON "public"."discount_analytics" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create "draft_cart_templates" table
CREATE TABLE "public"."draft_cart_templates" (
  "id" uuid NOT NULL DEFAULT public.uuid_generate_v4(),
  "organization_id" integer NOT NULL,
  "customer_id" integer NOT NULL,
  "template_name" character varying(255) NOT NULL,
  "description" text NULL,
  "template_type" character varying(50) NULL DEFAULT 'saved_cart',
  "is_favorite" boolean NULL DEFAULT false,
  "auto_reorder_enabled" boolean NULL DEFAULT false,
  "reorder_frequency_days" integer NULL,
  "next_reorder_date" date NULL,
  "total_items" integer NULL DEFAULT 0,
  "estimated_total" numeric(15,2) NULL DEFAULT 0.00,
  "metadata" jsonb NULL DEFAULT '{}',
  "notes" text NULL,
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "draft_cart_templates_organization_id_customer_id_template_n_key" UNIQUE ("organization_id", "customer_id", "template_name"),
  CONSTRAINT "draft_cart_templates_customer_id_fkey" FOREIGN KEY ("customer_id") REFERENCES "public"."customers" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "draft_cart_templates_organization_id_fkey" FOREIGN KEY ("organization_id") REFERENCES "public"."organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_draft_cart_templates_auto_reorder" to table: "draft_cart_templates"
CREATE INDEX "idx_draft_cart_templates_auto_reorder" ON "public"."draft_cart_templates" ("auto_reorder_enabled");
-- Create index "idx_draft_cart_templates_customer_id" to table: "draft_cart_templates"
CREATE INDEX "idx_draft_cart_templates_customer_id" ON "public"."draft_cart_templates" ("customer_id");
-- Create index "idx_draft_cart_templates_is_favorite" to table: "draft_cart_templates"
CREATE INDEX "idx_draft_cart_templates_is_favorite" ON "public"."draft_cart_templates" ("is_favorite");
-- Create index "idx_draft_cart_templates_next_reorder_date" to table: "draft_cart_templates"
CREATE INDEX "idx_draft_cart_templates_next_reorder_date" ON "public"."draft_cart_templates" ("next_reorder_date");
-- Create index "idx_draft_cart_templates_organization_id" to table: "draft_cart_templates"
CREATE INDEX "idx_draft_cart_templates_organization_id" ON "public"."draft_cart_templates" ("organization_id");
-- Create index "idx_draft_cart_templates_template_type" to table: "draft_cart_templates"
CREATE INDEX "idx_draft_cart_templates_template_type" ON "public"."draft_cart_templates" ("template_type");
-- Set comment to table: "draft_cart_templates"
COMMENT ON TABLE "public"."draft_cart_templates" IS 'Saved carts and wishlists for quick reordering';
-- Create "draft_cart_template_items" table
CREATE TABLE "public"."draft_cart_template_items" (
  "id" uuid NOT NULL DEFAULT public.uuid_generate_v4(),
  "template_id" uuid NOT NULL,
  "organization_id" integer NOT NULL,
  "product_id" integer NOT NULL,
  "product_variant_id" integer NULL,
  "quantity" numeric(15,3) NOT NULL,
  "uom_id" integer NULL,
  "last_known_price" numeric(15,2) NULL,
  "priority" integer NULL DEFAULT 0,
  "notes" text NULL,
  "metadata" jsonb NULL DEFAULT '{}',
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "draft_cart_template_items_organization_id_fkey" FOREIGN KEY ("organization_id") REFERENCES "public"."organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "draft_cart_template_items_product_id_fkey" FOREIGN KEY ("product_id") REFERENCES "public"."products" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "draft_cart_template_items_product_variant_id_fkey" FOREIGN KEY ("product_variant_id") REFERENCES "public"."product_variants" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "draft_cart_template_items_template_id_fkey" FOREIGN KEY ("template_id") REFERENCES "public"."draft_cart_templates" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "draft_cart_template_items_uom_id_fkey" FOREIGN KEY ("uom_id") REFERENCES "public"."units_of_measure" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "draft_cart_template_items_quantity_check" CHECK (quantity > (0)::numeric)
);
-- Create index "idx_draft_cart_template_items_product_id" to table: "draft_cart_template_items"
CREATE INDEX "idx_draft_cart_template_items_product_id" ON "public"."draft_cart_template_items" ("product_id");
-- Create index "idx_draft_cart_template_items_template_id" to table: "draft_cart_template_items"
CREATE INDEX "idx_draft_cart_template_items_template_id" ON "public"."draft_cart_template_items" ("template_id");
-- Create trigger "trg_draft_cart_template_items_updated_at"
CREATE TRIGGER "trg_draft_cart_template_items_updated_at" BEFORE UPDATE ON "public"."draft_cart_template_items" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create trigger "trg_draft_cart_templates_updated_at"
CREATE TRIGGER "trg_draft_cart_templates_updated_at" BEFORE UPDATE ON "public"."draft_cart_templates" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create "suppliers" table
CREATE TABLE "public"."suppliers" (
  "id" serial NOT NULL,
  "organization_id" integer NOT NULL,
  "code" character varying(50) NOT NULL,
  "name" character varying(255) NOT NULL,
  "supplier_type" character varying(50) NULL,
  "credit_limit" numeric(15,2) NULL DEFAULT 0,
  "contact_person" character varying(100) NULL,
  "email" character varying(255) NULL,
  "phone" character varying(50) NULL,
  "address" text NULL,
  "currency_code" character varying(3) NULL DEFAULT 'USD',
  "payment_terms" character varying(100) NULL,
  "tax_id" character varying(50) NULL,
  "is_active" boolean NULL DEFAULT true,
  "metadata" jsonb NULL DEFAULT '{}',
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "suppliers_organization_id_code_key" UNIQUE ("organization_id", "code"),
  CONSTRAINT "suppliers_organization_id_fkey" FOREIGN KEY ("organization_id") REFERENCES "public"."organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_suppliers_code" to table: "suppliers"
CREATE INDEX "idx_suppliers_code" ON "public"."suppliers" ("code");
-- Create index "idx_suppliers_is_active" to table: "suppliers"
CREATE INDEX "idx_suppliers_is_active" ON "public"."suppliers" ("is_active");
-- Create index "idx_suppliers_organization_id" to table: "suppliers"
CREATE INDEX "idx_suppliers_organization_id" ON "public"."suppliers" ("organization_id");
-- Create "purchase_orders" table
CREATE TABLE "public"."purchase_orders" (
  "id" serial NOT NULL,
  "organization_id" integer NOT NULL,
  "po_number" character varying(50) NOT NULL,
  "supplier_id" integer NOT NULL,
  "store_id" integer NOT NULL,
  "po_date" date NOT NULL,
  "expected_delivery_date" date NULL,
  "status" character varying(50) NULL DEFAULT 'draft',
  "subtotal" numeric(15,2) NULL DEFAULT 0,
  "discount_amount" numeric(15,2) NULL DEFAULT 0,
  "tax_amount" numeric(15,2) NULL DEFAULT 0,
  "total_amount" numeric(15,2) NULL DEFAULT 0,
  "price_list_id" integer NULL,
  "created_by" integer NULL,
  "approved_by" integer NULL,
  "metadata" jsonb NULL DEFAULT '{}',
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  "business_partner_id" integer NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "purchase_orders_po_number_key" UNIQUE ("po_number"),
  CONSTRAINT "purchase_orders_approved_by_fkey" FOREIGN KEY ("approved_by") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "purchase_orders_business_partner_id_fkey" FOREIGN KEY ("business_partner_id") REFERENCES "public"."business_partners" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "purchase_orders_created_by_fkey" FOREIGN KEY ("created_by") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "purchase_orders_organization_id_fkey" FOREIGN KEY ("organization_id") REFERENCES "public"."organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "purchase_orders_price_list_id_fkey" FOREIGN KEY ("price_list_id") REFERENCES "public"."price_lists" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "purchase_orders_store_id_fkey" FOREIGN KEY ("store_id") REFERENCES "public"."stores" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "purchase_orders_supplier_id_fkey" FOREIGN KEY ("supplier_id") REFERENCES "public"."suppliers" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_po_business_partner_id" to table: "purchase_orders"
CREATE INDEX "idx_po_business_partner_id" ON "public"."purchase_orders" ("business_partner_id");
-- Create index "idx_purchase_orders_organization_id" to table: "purchase_orders"
CREATE INDEX "idx_purchase_orders_organization_id" ON "public"."purchase_orders" ("organization_id");
-- Create index "idx_purchase_orders_po_date" to table: "purchase_orders"
CREATE INDEX "idx_purchase_orders_po_date" ON "public"."purchase_orders" ("po_date");
-- Create index "idx_purchase_orders_po_number" to table: "purchase_orders"
CREATE INDEX "idx_purchase_orders_po_number" ON "public"."purchase_orders" ("po_number");
-- Create index "idx_purchase_orders_status" to table: "purchase_orders"
CREATE INDEX "idx_purchase_orders_status" ON "public"."purchase_orders" ("status");
-- Create index "idx_purchase_orders_store_id" to table: "purchase_orders"
CREATE INDEX "idx_purchase_orders_store_id" ON "public"."purchase_orders" ("store_id");
-- Create index "idx_purchase_orders_supplier_id" to table: "purchase_orders"
CREATE INDEX "idx_purchase_orders_supplier_id" ON "public"."purchase_orders" ("supplier_id");
-- Create "goods_receipt_notes" table
CREATE TABLE "public"."goods_receipt_notes" (
  "id" serial NOT NULL,
  "organization_id" integer NOT NULL,
  "grn_number" character varying(50) NOT NULL,
  "purchase_order_id" integer NULL,
  "supplier_id" integer NOT NULL,
  "store_id" integer NOT NULL,
  "received_by" integer NULL,
  "receipt_date" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  "delivery_note_number" character varying(100) NULL,
  "status" character varying(50) NULL DEFAULT 'posted',
  "notes" text NULL,
  "metadata" jsonb NULL DEFAULT '{}',
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  "business_partner_id" integer NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "goods_receipt_notes_grn_number_key" UNIQUE ("grn_number"),
  CONSTRAINT "goods_receipt_notes_business_partner_id_fkey" FOREIGN KEY ("business_partner_id") REFERENCES "public"."business_partners" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "goods_receipt_notes_organization_id_fkey" FOREIGN KEY ("organization_id") REFERENCES "public"."organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "goods_receipt_notes_purchase_order_id_fkey" FOREIGN KEY ("purchase_order_id") REFERENCES "public"."purchase_orders" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "goods_receipt_notes_received_by_fkey" FOREIGN KEY ("received_by") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "goods_receipt_notes_store_id_fkey" FOREIGN KEY ("store_id") REFERENCES "public"."stores" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "goods_receipt_notes_supplier_id_fkey" FOREIGN KEY ("supplier_id") REFERENCES "public"."suppliers" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_grn_business_partner_id" to table: "goods_receipt_notes"
CREATE INDEX "idx_grn_business_partner_id" ON "public"."goods_receipt_notes" ("business_partner_id");
-- Create trigger "update_goods_receipt_notes_updated_at"
CREATE TRIGGER "update_goods_receipt_notes_updated_at" BEFORE UPDATE ON "public"."goods_receipt_notes" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create "inventory_analytics" table
CREATE TABLE "public"."inventory_analytics" (
  "id" serial NOT NULL,
  "organization_id" integer NOT NULL,
  "store_id" integer NULL,
  "product_id" integer NULL,
  "category_id" integer NULL,
  "date" date NOT NULL,
  "month" integer NULL,
  "quarter" integer NULL,
  "year" integer NULL,
  "opening_stock" numeric(15,3) NULL DEFAULT 0,
  "stock_in" numeric(15,3) NULL DEFAULT 0,
  "stock_out" numeric(15,3) NULL DEFAULT 0,
  "receipts" numeric(15,3) NULL DEFAULT 0,
  "issues" numeric(15,3) NULL DEFAULT 0,
  "adjustments" numeric(15,3) NULL DEFAULT 0,
  "closing_stock" numeric(15,3) NULL DEFAULT 0,
  "average_stock" numeric(15,3) NULL DEFAULT 0,
  "stock_value" numeric(15,2) NULL DEFAULT 0,
  "turnover_rate" numeric(5,2) NULL,
  "stock_turnover_ratio" numeric(5,2) NULL,
  "days_of_inventory" numeric(15,3) NULL DEFAULT 0,
  "days_in_stock" numeric(5,2) NULL,
  "low_stock_alerts" integer NULL DEFAULT 0,
  "out_of_stock_days" integer NULL DEFAULT 0,
  "metadata" jsonb NULL DEFAULT '{}',
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_inventory_analytics_category" FOREIGN KEY ("category_id") REFERENCES "public"."product_categories" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "fk_inventory_analytics_product" FOREIGN KEY ("product_id") REFERENCES "public"."products" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "fk_inventory_analytics_store" FOREIGN KEY ("store_id") REFERENCES "public"."stores" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "inventory_analytics_organization_id_fkey" FOREIGN KEY ("organization_id") REFERENCES "public"."organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_inventory_analytics_date" to table: "inventory_analytics"
CREATE INDEX "idx_inventory_analytics_date" ON "public"."inventory_analytics" ("date");
-- Create index "idx_inventory_analytics_organization_id" to table: "inventory_analytics"
CREATE INDEX "idx_inventory_analytics_organization_id" ON "public"."inventory_analytics" ("organization_id");
-- Create index "idx_inventory_analytics_product_id" to table: "inventory_analytics"
CREATE INDEX "idx_inventory_analytics_product_id" ON "public"."inventory_analytics" ("product_id");
-- Create index "idx_inventory_analytics_store_id" to table: "inventory_analytics"
CREATE INDEX "idx_inventory_analytics_store_id" ON "public"."inventory_analytics" ("store_id");
-- Create trigger "trg_inventory_analytics_updated_at"
CREATE TRIGGER "trg_inventory_analytics_updated_at" BEFORE UPDATE ON "public"."inventory_analytics" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create trigger "update_inventory_analytics_updated_at"
CREATE TRIGGER "update_inventory_analytics_updated_at" BEFORE UPDATE ON "public"."inventory_analytics" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create "storage_locations" table
CREATE TABLE "public"."storage_locations" (
  "id" serial NOT NULL,
  "store_id" integer NOT NULL,
  "code" character varying(50) NOT NULL,
  "name" character varying(255) NOT NULL,
  "location_type" character varying(50) NULL,
  "parent_location_id" integer NULL,
  "is_active" boolean NULL DEFAULT true,
  "metadata" jsonb NULL DEFAULT '{}',
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "storage_locations_store_id_code_key" UNIQUE ("store_id", "code"),
  CONSTRAINT "storage_locations_parent_location_id_fkey" FOREIGN KEY ("parent_location_id") REFERENCES "public"."storage_locations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "storage_locations_store_id_fkey" FOREIGN KEY ("store_id") REFERENCES "public"."stores" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_storage_locations_code" to table: "storage_locations"
CREATE INDEX "idx_storage_locations_code" ON "public"."storage_locations" ("code");
-- Create index "idx_storage_locations_parent_location_id" to table: "storage_locations"
CREATE INDEX "idx_storage_locations_parent_location_id" ON "public"."storage_locations" ("parent_location_id");
-- Create index "idx_storage_locations_store_id" to table: "storage_locations"
CREATE INDEX "idx_storage_locations_store_id" ON "public"."storage_locations" ("store_id");
-- Create "inventory_stock" table
CREATE TABLE "public"."inventory_stock" (
  "id" serial NOT NULL,
  "product_id" integer NOT NULL,
  "product_variant_id" integer NULL,
  "store_id" integer NOT NULL,
  "storage_location_id" integer NULL,
  "quantity_on_hand" numeric(15,3) NULL DEFAULT 0,
  "quantity_allocated" numeric(15,3) NULL DEFAULT 0,
  "quantity_available" numeric(15,3) NULL DEFAULT 0,
  "quantity_on_order" numeric(15,3) NULL DEFAULT 0,
  "quantity_in_transit" numeric(15,3) NULL DEFAULT 0,
  "reorder_level" numeric(15,3) NULL,
  "reorder_quantity" numeric(15,3) NULL,
  "max_stock_level" numeric(15,3) NULL,
  "last_counted_at" timestamp NULL,
  "metadata" jsonb NULL DEFAULT '{}',
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "inventory_stock_product_id_fkey" FOREIGN KEY ("product_id") REFERENCES "public"."products" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "inventory_stock_product_variant_id_fkey" FOREIGN KEY ("product_variant_id") REFERENCES "public"."product_variants" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "inventory_stock_storage_location_id_fkey" FOREIGN KEY ("storage_location_id") REFERENCES "public"."storage_locations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "inventory_stock_store_id_fkey" FOREIGN KEY ("store_id") REFERENCES "public"."stores" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_inventory_stock_product_id" to table: "inventory_stock"
CREATE INDEX "idx_inventory_stock_product_id" ON "public"."inventory_stock" ("product_id");
-- Create index "idx_inventory_stock_product_variant_id" to table: "inventory_stock"
CREATE INDEX "idx_inventory_stock_product_variant_id" ON "public"."inventory_stock" ("product_variant_id");
-- Create index "idx_inventory_stock_storage_location_id" to table: "inventory_stock"
CREATE INDEX "idx_inventory_stock_storage_location_id" ON "public"."inventory_stock" ("storage_location_id");
-- Create index "idx_inventory_stock_store_id" to table: "inventory_stock"
CREATE INDEX "idx_inventory_stock_store_id" ON "public"."inventory_stock" ("store_id");
-- Create index "idx_inventory_stock_store_product_qty" to table: "inventory_stock"
CREATE INDEX "idx_inventory_stock_store_product_qty" ON "public"."inventory_stock" ("store_id", "product_id", "quantity_available");
-- Create trigger "trg_inventory_stock_updated_at"
CREATE TRIGGER "trg_inventory_stock_updated_at" BEFORE UPDATE ON "public"."inventory_stock" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create trigger "update_inventory_stock_updated_at"
CREATE TRIGGER "update_inventory_stock_updated_at" BEFORE UPDATE ON "public"."inventory_stock" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create "sales_orders_v2" table
CREATE TABLE "public"."sales_orders_v2" (
  "id" uuid NOT NULL DEFAULT public.uuid_generate_v4(),
  "order_number" character varying(50) NOT NULL,
  "organization_id" integer NOT NULL,
  "store_id" integer NULL,
  "customer_id" integer NULL,
  "customer_name" character varying(255) NULL,
  "customer_email" character varying(255) NULL,
  "customer_phone" character varying(50) NULL,
  "order_type" "public"."order_type" NOT NULL DEFAULT 'standard',
  "order_status" "public"."order_status_v2" NOT NULL DEFAULT 'draft',
  "payment_status" "public"."payment_status" NOT NULL DEFAULT 'unpaid',
  "fulfillment_status" "public"."fulfillment_status" NOT NULL DEFAULT 'unfulfilled',
  "sales_channel" character varying(50) NULL DEFAULT 'online',
  "order_source" character varying(100) NULL,
  "referral_source" character varying(255) NULL,
  "source_cart_id" uuid NULL,
  "created_by_user_id" integer NULL,
  "assigned_to_user_id" integer NULL,
  "order_date" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  "confirmed_date" timestamp NULL,
  "expected_delivery_date" date NULL,
  "actual_delivery_date" date NULL,
  "cancelled_date" timestamp NULL,
  "subtotal" numeric(15,2) NULL DEFAULT 0.00,
  "discount_amount" numeric(15,2) NULL DEFAULT 0.00,
  "tax_amount" numeric(15,2) NULL DEFAULT 0.00,
  "shipping_amount" numeric(15,2) NULL DEFAULT 0.00,
  "adjustment_amount" numeric(15,2) NULL DEFAULT 0.00,
  "total_amount" numeric(15,2) NULL DEFAULT 0.00,
  "paid_amount" numeric(15,2) NULL DEFAULT 0.00,
  "refunded_amount" numeric(15,2) NULL DEFAULT 0.00,
  "balance_due" numeric(15,2) NULL DEFAULT 0.00,
  "coupon_code" character varying(100) NULL,
  "discount_codes" text[] NULL,
  "promotional_credits" numeric(15,2) NULL DEFAULT 0.00,
  "shipping_address" jsonb NOT NULL,
  "billing_address" jsonb NOT NULL,
  "shipping_method" character varying(100) NULL,
  "shipping_carrier" character varying(100) NULL,
  "tracking_number" character varying(255) NULL,
  "tracking_url" text NULL,
  "payment_method" character varying(100) NULL,
  "payment_gateway" character varying(100) NULL,
  "payment_terms" character varying(100) NULL,
  "payment_due_date" date NULL,
  "pos_terminal_id" integer NULL,
  "cashier_id" integer NULL,
  "is_gift" boolean NULL DEFAULT false,
  "gift_message" text NULL,
  "special_instructions" text NULL,
  "internal_notes" text NULL,
  "tags" text[] NULL,
  "priority" character varying(20) NULL DEFAULT 'normal',
  "metadata" jsonb NULL DEFAULT '{}',
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "sales_orders_v2_order_number_key" UNIQUE ("order_number"),
  CONSTRAINT "sales_orders_v2_assigned_to_user_id_fkey" FOREIGN KEY ("assigned_to_user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "sales_orders_v2_cashier_id_fkey" FOREIGN KEY ("cashier_id") REFERENCES "public"."cashiers" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "sales_orders_v2_created_by_user_id_fkey" FOREIGN KEY ("created_by_user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "sales_orders_v2_customer_id_fkey" FOREIGN KEY ("customer_id") REFERENCES "public"."customers" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "sales_orders_v2_organization_id_fkey" FOREIGN KEY ("organization_id") REFERENCES "public"."organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "sales_orders_v2_pos_terminal_id_fkey" FOREIGN KEY ("pos_terminal_id") REFERENCES "public"."pos_terminals" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "sales_orders_v2_source_cart_id_fkey" FOREIGN KEY ("source_cart_id") REFERENCES "public"."carts" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "sales_orders_v2_store_id_fkey" FOREIGN KEY ("store_id") REFERENCES "public"."stores" ("id") ON UPDATE NO ACTION ON DELETE SET NULL
);
-- Create index "idx_sales_orders_v2_created_at" to table: "sales_orders_v2"
CREATE INDEX "idx_sales_orders_v2_created_at" ON "public"."sales_orders_v2" ("created_at");
-- Create index "idx_sales_orders_v2_customer_id" to table: "sales_orders_v2"
CREATE INDEX "idx_sales_orders_v2_customer_id" ON "public"."sales_orders_v2" ("customer_id");
-- Create index "idx_sales_orders_v2_fulfillment_status" to table: "sales_orders_v2"
CREATE INDEX "idx_sales_orders_v2_fulfillment_status" ON "public"."sales_orders_v2" ("fulfillment_status");
-- Create index "idx_sales_orders_v2_order_date" to table: "sales_orders_v2"
CREATE INDEX "idx_sales_orders_v2_order_date" ON "public"."sales_orders_v2" ("order_date");
-- Create index "idx_sales_orders_v2_order_number" to table: "sales_orders_v2"
CREATE INDEX "idx_sales_orders_v2_order_number" ON "public"."sales_orders_v2" ("order_number");
-- Create index "idx_sales_orders_v2_order_status" to table: "sales_orders_v2"
CREATE INDEX "idx_sales_orders_v2_order_status" ON "public"."sales_orders_v2" ("order_status");
-- Create index "idx_sales_orders_v2_order_type" to table: "sales_orders_v2"
CREATE INDEX "idx_sales_orders_v2_order_type" ON "public"."sales_orders_v2" ("order_type");
-- Create index "idx_sales_orders_v2_organization_id" to table: "sales_orders_v2"
CREATE INDEX "idx_sales_orders_v2_organization_id" ON "public"."sales_orders_v2" ("organization_id");
-- Create index "idx_sales_orders_v2_payment_status" to table: "sales_orders_v2"
CREATE INDEX "idx_sales_orders_v2_payment_status" ON "public"."sales_orders_v2" ("payment_status");
-- Create index "idx_sales_orders_v2_sales_channel" to table: "sales_orders_v2"
CREATE INDEX "idx_sales_orders_v2_sales_channel" ON "public"."sales_orders_v2" ("sales_channel");
-- Create index "idx_sales_orders_v2_source_cart_id" to table: "sales_orders_v2"
CREATE INDEX "idx_sales_orders_v2_source_cart_id" ON "public"."sales_orders_v2" ("source_cart_id");
-- Create index "idx_sales_orders_v2_store_id" to table: "sales_orders_v2"
CREATE INDEX "idx_sales_orders_v2_store_id" ON "public"."sales_orders_v2" ("store_id");
-- Set comment to table: "sales_orders_v2"
COMMENT ON TABLE "public"."sales_orders_v2" IS 'Enhanced order management with comprehensive tracking across all sales channels';
-- Create "invoices" table
CREATE TABLE "public"."invoices" (
  "id" uuid NOT NULL DEFAULT public.uuid_generate_v4(),
  "invoice_number" character varying(50) NOT NULL,
  "organization_id" integer NOT NULL,
  "store_id" integer NULL,
  "customer_id" integer NOT NULL,
  "customer_name" character varying(255) NOT NULL,
  "customer_email" character varying(255) NULL,
  "customer_phone" character varying(50) NULL,
  "customer_tax_id" character varying(50) NULL,
  "invoice_type" "public"."invoice_type" NOT NULL DEFAULT 'standard',
  "invoice_status" "public"."invoice_status" NOT NULL DEFAULT 'draft',
  "sales_order_id" uuid NULL,
  "related_invoice_id" uuid NULL,
  "invoice_date" date NOT NULL DEFAULT CURRENT_DATE,
  "due_date" date NOT NULL,
  "sent_date" date NULL,
  "paid_date" date NULL,
  "subtotal" numeric(15,2) NULL DEFAULT 0.00,
  "discount_amount" numeric(15,2) NULL DEFAULT 0.00,
  "tax_amount" numeric(15,2) NULL DEFAULT 0.00,
  "shipping_amount" numeric(15,2) NULL DEFAULT 0.00,
  "adjustment_amount" numeric(15,2) NULL DEFAULT 0.00,
  "total_amount" numeric(15,2) NOT NULL,
  "paid_amount" numeric(15,2) NULL DEFAULT 0.00,
  "credit_applied" numeric(15,2) NULL DEFAULT 0.00,
  "balance_due" numeric(15,2) NOT NULL,
  "payment_terms" character varying(100) NULL,
  "currency_code" character varying(3) NULL DEFAULT 'USD',
  "exchange_rate" numeric(15,6) NULL DEFAULT 1.000000,
  "billing_address" jsonb NOT NULL,
  "shipping_address" jsonb NULL,
  "is_recurring" boolean NULL DEFAULT false,
  "recurrence_pattern" character varying(50) NULL,
  "next_invoice_date" date NULL,
  "pdf_url" text NULL,
  "document_hash" character varying(255) NULL,
  "reminder_sent_count" integer NULL DEFAULT 0,
  "last_reminder_sent_at" timestamp NULL,
  "notes" text NULL,
  "internal_notes" text NULL,
  "reference_number" character varying(100) NULL,
  "created_by_user_id" integer NULL,
  "metadata" jsonb NULL DEFAULT '{}',
  "tags" text[] NULL,
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "invoices_invoice_number_key" UNIQUE ("invoice_number"),
  CONSTRAINT "invoices_created_by_user_id_fkey" FOREIGN KEY ("created_by_user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "invoices_customer_id_fkey" FOREIGN KEY ("customer_id") REFERENCES "public"."customers" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT,
  CONSTRAINT "invoices_organization_id_fkey" FOREIGN KEY ("organization_id") REFERENCES "public"."organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "invoices_related_invoice_id_fkey" FOREIGN KEY ("related_invoice_id") REFERENCES "public"."invoices" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "invoices_sales_order_id_fkey" FOREIGN KEY ("sales_order_id") REFERENCES "public"."sales_orders_v2" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "invoices_store_id_fkey" FOREIGN KEY ("store_id") REFERENCES "public"."stores" ("id") ON UPDATE NO ACTION ON DELETE SET NULL
);
-- Create index "idx_invoices_customer_id" to table: "invoices"
CREATE INDEX "idx_invoices_customer_id" ON "public"."invoices" ("customer_id");
-- Create index "idx_invoices_due_date" to table: "invoices"
CREATE INDEX "idx_invoices_due_date" ON "public"."invoices" ("due_date");
-- Create index "idx_invoices_invoice_date" to table: "invoices"
CREATE INDEX "idx_invoices_invoice_date" ON "public"."invoices" ("invoice_date");
-- Create index "idx_invoices_invoice_number" to table: "invoices"
CREATE INDEX "idx_invoices_invoice_number" ON "public"."invoices" ("invoice_number");
-- Create index "idx_invoices_invoice_status" to table: "invoices"
CREATE INDEX "idx_invoices_invoice_status" ON "public"."invoices" ("invoice_status");
-- Create index "idx_invoices_invoice_type" to table: "invoices"
CREATE INDEX "idx_invoices_invoice_type" ON "public"."invoices" ("invoice_type");
-- Create index "idx_invoices_is_recurring" to table: "invoices"
CREATE INDEX "idx_invoices_is_recurring" ON "public"."invoices" ("is_recurring");
-- Create index "idx_invoices_next_invoice_date" to table: "invoices"
CREATE INDEX "idx_invoices_next_invoice_date" ON "public"."invoices" ("next_invoice_date");
-- Create index "idx_invoices_organization_id" to table: "invoices"
CREATE INDEX "idx_invoices_organization_id" ON "public"."invoices" ("organization_id");
-- Create index "idx_invoices_sales_order_id" to table: "invoices"
CREATE INDEX "idx_invoices_sales_order_id" ON "public"."invoices" ("sales_order_id");
-- Create index "idx_invoices_store_id" to table: "invoices"
CREATE INDEX "idx_invoices_store_id" ON "public"."invoices" ("store_id");
-- Set comment to table: "invoices"
COMMENT ON TABLE "public"."invoices" IS 'Customer invoices with payment tracking and recurring billing support';
-- Create "sales_order_lines_v2" table
CREATE TABLE "public"."sales_order_lines_v2" (
  "id" uuid NOT NULL DEFAULT public.uuid_generate_v4(),
  "sales_order_id" uuid NOT NULL,
  "organization_id" integer NOT NULL,
  "line_number" integer NOT NULL,
  "product_id" integer NOT NULL,
  "product_variant_id" integer NULL,
  "product_name" character varying(255) NOT NULL,
  "product_sku" character varying(100) NULL,
  "quantity_ordered" numeric(15,3) NOT NULL,
  "quantity_fulfilled" numeric(15,3) NULL DEFAULT 0.00,
  "quantity_cancelled" numeric(15,3) NULL DEFAULT 0.00,
  "quantity_returned" numeric(15,3) NULL DEFAULT 0.00,
  "uom_id" integer NULL,
  "unit_price" numeric(15,2) NOT NULL,
  "discount_amount" numeric(15,2) NULL DEFAULT 0.00,
  "discount_percentage" numeric(5,2) NULL DEFAULT 0.00,
  "tax_amount" numeric(15,2) NULL DEFAULT 0.00,
  "line_total" numeric(15,2) NOT NULL,
  "tax_category_id" integer NULL,
  "tax_rate" numeric(5,2) NULL,
  "batch_number" character varying(100) NULL,
  "serial_numbers" text[] NULL,
  "expiry_date" date NULL,
  "line_status" character varying(50) NULL DEFAULT 'pending',
  "customization_details" jsonb NULL DEFAULT '{}',
  "unit_cost" numeric(15,2) NULL,
  "notes" text NULL,
  "metadata" jsonb NULL DEFAULT '{}',
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "sales_order_lines_v2_sales_order_id_line_number_key" UNIQUE ("sales_order_id", "line_number"),
  CONSTRAINT "sales_order_lines_v2_organization_id_fkey" FOREIGN KEY ("organization_id") REFERENCES "public"."organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "sales_order_lines_v2_product_id_fkey" FOREIGN KEY ("product_id") REFERENCES "public"."products" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT,
  CONSTRAINT "sales_order_lines_v2_product_variant_id_fkey" FOREIGN KEY ("product_variant_id") REFERENCES "public"."product_variants" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "sales_order_lines_v2_sales_order_id_fkey" FOREIGN KEY ("sales_order_id") REFERENCES "public"."sales_orders_v2" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "sales_order_lines_v2_tax_category_id_fkey" FOREIGN KEY ("tax_category_id") REFERENCES "public"."tax_categories" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "sales_order_lines_v2_uom_id_fkey" FOREIGN KEY ("uom_id") REFERENCES "public"."units_of_measure" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "sales_order_lines_v2_quantity_ordered_check" CHECK (quantity_ordered > (0)::numeric)
);
-- Create index "idx_sales_order_lines_v2_line_status" to table: "sales_order_lines_v2"
CREATE INDEX "idx_sales_order_lines_v2_line_status" ON "public"."sales_order_lines_v2" ("line_status");
-- Create index "idx_sales_order_lines_v2_product_id" to table: "sales_order_lines_v2"
CREATE INDEX "idx_sales_order_lines_v2_product_id" ON "public"."sales_order_lines_v2" ("product_id");
-- Create index "idx_sales_order_lines_v2_product_variant_id" to table: "sales_order_lines_v2"
CREATE INDEX "idx_sales_order_lines_v2_product_variant_id" ON "public"."sales_order_lines_v2" ("product_variant_id");
-- Create index "idx_sales_order_lines_v2_sales_order_id" to table: "sales_order_lines_v2"
CREATE INDEX "idx_sales_order_lines_v2_sales_order_id" ON "public"."sales_order_lines_v2" ("sales_order_id");
-- Set comment to table: "sales_order_lines_v2"
COMMENT ON TABLE "public"."sales_order_lines_v2" IS 'Order line items with fulfillment tracking';
-- Create "invoice_lines" table
CREATE TABLE "public"."invoice_lines" (
  "id" uuid NOT NULL DEFAULT public.uuid_generate_v4(),
  "invoice_id" uuid NOT NULL,
  "organization_id" integer NOT NULL,
  "line_number" integer NOT NULL,
  "description" text NOT NULL,
  "item_type" character varying(50) NULL DEFAULT 'product',
  "product_id" integer NULL,
  "product_variant_id" integer NULL,
  "product_sku" character varying(100) NULL,
  "order_line_id" uuid NULL,
  "quantity" numeric(15,3) NULL DEFAULT 1.000,
  "unit_price" numeric(15,2) NOT NULL,
  "discount_amount" numeric(15,2) NULL DEFAULT 0.00,
  "tax_amount" numeric(15,2) NULL DEFAULT 0.00,
  "line_total" numeric(15,2) NOT NULL,
  "tax_category_id" integer NULL,
  "tax_rate" numeric(5,2) NULL,
  "uom_id" integer NULL,
  "metadata" jsonb NULL DEFAULT '{}',
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "invoice_lines_invoice_id_line_number_key" UNIQUE ("invoice_id", "line_number"),
  CONSTRAINT "invoice_lines_invoice_id_fkey" FOREIGN KEY ("invoice_id") REFERENCES "public"."invoices" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "invoice_lines_order_line_id_fkey" FOREIGN KEY ("order_line_id") REFERENCES "public"."sales_order_lines_v2" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "invoice_lines_organization_id_fkey" FOREIGN KEY ("organization_id") REFERENCES "public"."organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "invoice_lines_product_id_fkey" FOREIGN KEY ("product_id") REFERENCES "public"."products" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "invoice_lines_product_variant_id_fkey" FOREIGN KEY ("product_variant_id") REFERENCES "public"."product_variants" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "invoice_lines_tax_category_id_fkey" FOREIGN KEY ("tax_category_id") REFERENCES "public"."tax_categories" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "invoice_lines_uom_id_fkey" FOREIGN KEY ("uom_id") REFERENCES "public"."units_of_measure" ("id") ON UPDATE NO ACTION ON DELETE SET NULL
);
-- Create index "idx_invoice_lines_invoice_id" to table: "invoice_lines"
CREATE INDEX "idx_invoice_lines_invoice_id" ON "public"."invoice_lines" ("invoice_id");
-- Create index "idx_invoice_lines_order_line_id" to table: "invoice_lines"
CREATE INDEX "idx_invoice_lines_order_line_id" ON "public"."invoice_lines" ("order_line_id");
-- Create index "idx_invoice_lines_product_id" to table: "invoice_lines"
CREATE INDEX "idx_invoice_lines_product_id" ON "public"."invoice_lines" ("product_id");
-- Create "calculate_invoice_totals" function
-- +goose StatementBegin
CREATE FUNCTION "public"."calculate_invoice_totals" () RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE
    v_subtotal DECIMAL(15,2);
    v_tax_amount DECIMAL(15,2);
BEGIN
    -- Calculate from invoice lines
    SELECT 
        COALESCE(SUM(line_total - tax_amount), 0),
        COALESCE(SUM(tax_amount), 0)
    INTO v_subtotal, v_tax_amount
    FROM invoice_lines
    WHERE invoice_id = COALESCE(NEW.invoice_id, OLD.invoice_id);
    
    -- Update the invoice
    UPDATE invoices
    SET 
        subtotal = v_subtotal,
        tax_amount = v_tax_amount,
        total_amount = v_subtotal + v_tax_amount + COALESCE(shipping_amount, 0) + COALESCE(adjustment_amount, 0) - COALESCE(discount_amount, 0),
        balance_due = (v_subtotal + v_tax_amount + COALESCE(shipping_amount, 0) + COALESCE(adjustment_amount, 0) - COALESCE(discount_amount, 0)) - COALESCE(paid_amount, 0) - COALESCE(credit_applied, 0)
    WHERE id = COALESCE(NEW.invoice_id, OLD.invoice_id);
    
    RETURN COALESCE(NEW, OLD);
END;
$$;
-- +goose StatementEnd
-- Create trigger "calculate_invoice_totals_trigger"
CREATE TRIGGER "calculate_invoice_totals_trigger" AFTER DELETE OR INSERT OR UPDATE ON "public"."invoice_lines" FOR EACH ROW EXECUTE FUNCTION "public"."calculate_invoice_totals"();
-- Create trigger "trg_invoice_lines_updated_at"
CREATE TRIGGER "trg_invoice_lines_updated_at" BEFORE UPDATE ON "public"."invoice_lines" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create "invoice_payments" table
CREATE TABLE "public"."invoice_payments" (
  "id" uuid NOT NULL DEFAULT public.uuid_generate_v4(),
  "invoice_id" uuid NOT NULL,
  "organization_id" integer NOT NULL,
  "payment_number" character varying(50) NOT NULL,
  "payment_date" date NOT NULL DEFAULT CURRENT_DATE,
  "payment_amount" numeric(15,2) NOT NULL,
  "payment_method" character varying(100) NOT NULL,
  "payment_gateway" character varying(100) NULL,
  "payment_reference" character varying(255) NULL,
  "currency_code" character varying(3) NULL DEFAULT 'USD',
  "exchange_rate" numeric(15,6) NULL DEFAULT 1.000000,
  "bank_account_id" integer NULL,
  "reconciled" boolean NULL DEFAULT false,
  "reconciled_date" date NULL,
  "notes" text NULL,
  "received_by_user_id" integer NULL,
  "metadata" jsonb NULL DEFAULT '{}',
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "invoice_payments_payment_number_key" UNIQUE ("payment_number"),
  CONSTRAINT "invoice_payments_invoice_id_fkey" FOREIGN KEY ("invoice_id") REFERENCES "public"."invoices" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "invoice_payments_organization_id_fkey" FOREIGN KEY ("organization_id") REFERENCES "public"."organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "invoice_payments_received_by_user_id_fkey" FOREIGN KEY ("received_by_user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "invoice_payments_payment_amount_check" CHECK (payment_amount > (0)::numeric)
);
-- Create index "idx_invoice_payments_invoice_id" to table: "invoice_payments"
CREATE INDEX "idx_invoice_payments_invoice_id" ON "public"."invoice_payments" ("invoice_id");
-- Create index "idx_invoice_payments_payment_date" to table: "invoice_payments"
CREATE INDEX "idx_invoice_payments_payment_date" ON "public"."invoice_payments" ("payment_date");
-- Create index "idx_invoice_payments_payment_number" to table: "invoice_payments"
CREATE INDEX "idx_invoice_payments_payment_number" ON "public"."invoice_payments" ("payment_number");
-- Create index "idx_invoice_payments_reconciled" to table: "invoice_payments"
CREATE INDEX "idx_invoice_payments_reconciled" ON "public"."invoice_payments" ("reconciled");
-- Set comment to table: "invoice_payments"
COMMENT ON TABLE "public"."invoice_payments" IS 'Payment records against invoices';
-- Create "update_invoice_payment" function
-- +goose StatementBegin
CREATE FUNCTION "public"."update_invoice_payment" () RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE
    v_total_paid DECIMAL(15,2);
BEGIN
    -- Calculate total paid
    SELECT COALESCE(SUM(payment_amount), 0)
    INTO v_total_paid
    FROM invoice_payments
    WHERE invoice_id = COALESCE(NEW.invoice_id, OLD.invoice_id);
    
    -- Update invoice
    UPDATE invoices
    SET 
        paid_amount = v_total_paid,
        balance_due = total_amount - v_total_paid - COALESCE(credit_applied, 0),
        invoice_status = CASE
            WHEN v_total_paid = 0 THEN 'sent'::invoice_status
            WHEN v_total_paid >= total_amount - COALESCE(credit_applied, 0) THEN 'paid'::invoice_status
            ELSE 'partially_paid'::invoice_status
        END,
        paid_date = CASE 
            WHEN v_total_paid >= total_amount - COALESCE(credit_applied, 0) THEN CURRENT_DATE
            ELSE NULL
        END
    WHERE id = COALESCE(NEW.invoice_id, OLD.invoice_id);
    
    RETURN COALESCE(NEW, OLD);
END;
$$;
-- +goose StatementEnd
-- Create trigger "update_invoice_payment_trigger"
CREATE TRIGGER "update_invoice_payment_trigger" AFTER DELETE OR INSERT OR UPDATE ON "public"."invoice_payments" FOR EACH ROW EXECUTE FUNCTION "public"."update_invoice_payment"();
-- Create trigger "trg_invoice_payments_updated_at"
CREATE TRIGGER "trg_invoice_payments_updated_at" BEFORE UPDATE ON "public"."invoice_payments" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create trigger "trg_invoices_updated_at"
CREATE TRIGGER "trg_invoices_updated_at" BEFORE UPDATE ON "public"."invoices" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create "loyalty_redemption_rules" table
CREATE TABLE "public"."loyalty_redemption_rules" (
  "id" serial NOT NULL,
  "organization_id" integer NOT NULL,
  "rule_name" character varying(255) NOT NULL,
  "points_earning_rate" numeric(10,4) NULL DEFAULT 1,
  "points_redemption_rate" numeric(10,4) NULL DEFAULT 1,
  "min_points_to_redeem" numeric(15,2) NULL DEFAULT 0,
  "max_points_per_txn" numeric(15,2) NULL,
  "max_redemption_percent" numeric(5,2) NULL,
  "eligible_product_types" text[] NULL DEFAULT '{}',
  "expiry_days" integer NULL,
  "is_active" boolean NULL DEFAULT true,
  "valid_from" date NULL,
  "valid_to" date NULL,
  "metadata" jsonb NULL DEFAULT '{}',
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "loyalty_redemption_rules_organization_id_fkey" FOREIGN KEY ("organization_id") REFERENCES "public"."organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "loyalty_redemption_rules_max_redemption_percent_check" CHECK ((max_redemption_percent >= (0)::numeric) AND (max_redemption_percent <= (100)::numeric))
);
-- Create trigger "trg_loyalty_redemption_rules_updated_at"
CREATE TRIGGER "trg_loyalty_redemption_rules_updated_at" BEFORE UPDATE ON "public"."loyalty_redemption_rules" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create "menu_categories" table
CREATE TABLE "public"."menu_categories" (
  "id" serial NOT NULL,
  "store_id" integer NOT NULL,
  "parent_category_id" integer NULL,
  "name" character varying(255) NOT NULL,
  "code" character varying(50) NOT NULL,
  "description" text NULL,
  "category_level" integer NULL DEFAULT 1,
  "display_order" integer NULL DEFAULT 0,
  "icon" character varying(100) NULL,
  "image_url" text NULL,
  "is_active" boolean NULL DEFAULT true,
  "metadata" jsonb NULL DEFAULT '{}',
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "menu_categories_store_id_code_key" UNIQUE ("store_id", "code"),
  CONSTRAINT "menu_categories_parent_category_id_fkey" FOREIGN KEY ("parent_category_id") REFERENCES "public"."menu_categories" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "menu_categories_store_id_fkey" FOREIGN KEY ("store_id") REFERENCES "public"."stores" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_menu_categories_display_order" to table: "menu_categories"
CREATE INDEX "idx_menu_categories_display_order" ON "public"."menu_categories" ("display_order");
-- Create index "idx_menu_categories_is_active" to table: "menu_categories"
CREATE INDEX "idx_menu_categories_is_active" ON "public"."menu_categories" ("is_active");
-- Create index "idx_menu_categories_parent_id" to table: "menu_categories"
CREATE INDEX "idx_menu_categories_parent_id" ON "public"."menu_categories" ("parent_category_id");
-- Create index "idx_menu_categories_store_id" to table: "menu_categories"
CREATE INDEX "idx_menu_categories_store_id" ON "public"."menu_categories" ("store_id");
-- Create trigger "trg_menu_categories_updated_at"
CREATE TRIGGER "trg_menu_categories_updated_at" BEFORE UPDATE ON "public"."menu_categories" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create "recipes" table
CREATE TABLE "public"."recipes" (
  "id" serial NOT NULL,
  "organization_id" integer NOT NULL,
  "recipe_code" character varying(50) NOT NULL,
  "recipe_name" character varying(255) NOT NULL,
  "description" text NULL,
  "finished_product_id" integer NULL,
  "yield_quantity" numeric(15,3) NULL DEFAULT 1,
  "yield_uom_id" integer NULL,
  "preparation_steps" text NULL,
  "preparation_time_min" integer NULL DEFAULT 0,
  "cooking_time_min" integer NULL DEFAULT 0,
  "is_active" boolean NULL DEFAULT true,
  "metadata" jsonb NULL DEFAULT '{}',
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "recipes_organization_id_recipe_code_key" UNIQUE ("organization_id", "recipe_code"),
  CONSTRAINT "recipes_finished_product_id_fkey" FOREIGN KEY ("finished_product_id") REFERENCES "public"."products" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "recipes_organization_id_fkey" FOREIGN KEY ("organization_id") REFERENCES "public"."organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "recipes_yield_uom_id_fkey" FOREIGN KEY ("yield_uom_id") REFERENCES "public"."units_of_measure" ("id") ON UPDATE NO ACTION ON DELETE SET NULL
);
-- Create index "idx_recipes_code" to table: "recipes"
CREATE INDEX "idx_recipes_code" ON "public"."recipes" ("recipe_code");
-- Create index "idx_recipes_finished_product_id" to table: "recipes"
CREATE INDEX "idx_recipes_finished_product_id" ON "public"."recipes" ("finished_product_id");
-- Create index "idx_recipes_is_active" to table: "recipes"
CREATE INDEX "idx_recipes_is_active" ON "public"."recipes" ("is_active");
-- Create index "idx_recipes_organization_id" to table: "recipes"
CREATE INDEX "idx_recipes_organization_id" ON "public"."recipes" ("organization_id");
-- Create "menu_items" table
CREATE TABLE "public"."menu_items" (
  "id" serial NOT NULL,
  "store_id" integer NOT NULL,
  "menu_category_id" integer NOT NULL,
  "product_id" integer NULL,
  "recipe_id" integer NULL,
  "name" character varying(255) NOT NULL,
  "short_name" character varying(50) NULL,
  "description" text NULL,
  "image_url" text NULL,
  "base_price" numeric(15,2) NOT NULL,
  "cost_price" numeric(15,2) NULL DEFAULT 0,
  "preparation_time_min" integer NULL DEFAULT 0,
  "tax_category_id" integer NULL,
  "is_available" boolean NULL DEFAULT true,
  "is_active" boolean NULL DEFAULT true,
  "display_order" integer NULL DEFAULT 0,
  "metadata" jsonb NULL DEFAULT '{}',
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_menu_items_recipe" FOREIGN KEY ("recipe_id") REFERENCES "public"."recipes" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "menu_items_menu_category_id_fkey" FOREIGN KEY ("menu_category_id") REFERENCES "public"."menu_categories" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "menu_items_product_id_fkey" FOREIGN KEY ("product_id") REFERENCES "public"."products" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "menu_items_store_id_fkey" FOREIGN KEY ("store_id") REFERENCES "public"."stores" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "menu_items_tax_category_id_fkey" FOREIGN KEY ("tax_category_id") REFERENCES "public"."tax_categories" ("id") ON UPDATE NO ACTION ON DELETE SET NULL
);
-- Create index "idx_menu_items_category_id" to table: "menu_items"
CREATE INDEX "idx_menu_items_category_id" ON "public"."menu_items" ("menu_category_id");
-- Create index "idx_menu_items_display_order" to table: "menu_items"
CREATE INDEX "idx_menu_items_display_order" ON "public"."menu_items" ("display_order");
-- Create index "idx_menu_items_is_active" to table: "menu_items"
CREATE INDEX "idx_menu_items_is_active" ON "public"."menu_items" ("is_active");
-- Create index "idx_menu_items_is_available" to table: "menu_items"
CREATE INDEX "idx_menu_items_is_available" ON "public"."menu_items" ("is_available");
-- Create index "idx_menu_items_product_id" to table: "menu_items"
CREATE INDEX "idx_menu_items_product_id" ON "public"."menu_items" ("product_id");
-- Create index "idx_menu_items_recipe_id" to table: "menu_items"
CREATE INDEX "idx_menu_items_recipe_id" ON "public"."menu_items" ("recipe_id");
-- Create index "idx_menu_items_store_id" to table: "menu_items"
CREATE INDEX "idx_menu_items_store_id" ON "public"."menu_items" ("store_id");
-- Create trigger "trg_menu_items_updated_at"
CREATE TRIGGER "trg_menu_items_updated_at" BEFORE UPDATE ON "public"."menu_items" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create "menu_modifier_groups" table
CREATE TABLE "public"."menu_modifier_groups" (
  "id" serial NOT NULL,
  "store_id" integer NOT NULL,
  "name" character varying(100) NOT NULL,
  "code" character varying(50) NOT NULL,
  "selection_type" character varying(20) NULL DEFAULT 'optional',
  "min_selections" integer NULL DEFAULT 0,
  "max_selections" integer NULL,
  "is_active" boolean NULL DEFAULT true,
  "display_order" integer NULL DEFAULT 0,
  "metadata" jsonb NULL DEFAULT '{}',
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "menu_modifier_groups_store_id_code_key" UNIQUE ("store_id", "code"),
  CONSTRAINT "menu_modifier_groups_store_id_fkey" FOREIGN KEY ("store_id") REFERENCES "public"."stores" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "chk_modifier_group_selections" CHECK ((min_selections >= 0) AND ((max_selections IS NULL) OR (max_selections >= min_selections))),
  CONSTRAINT "menu_modifier_groups_selection_type_check" CHECK ((selection_type)::text = ANY ((ARRAY['required'::character varying, 'optional'::character varying, 'multiple'::character varying])::text[]))
);
-- Create trigger "trg_menu_modifier_groups_updated_at"
CREATE TRIGGER "trg_menu_modifier_groups_updated_at" BEFORE UPDATE ON "public"."menu_modifier_groups" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create "modules" table
CREATE TABLE "public"."modules" (
  "id" serial NOT NULL,
  "name" character varying(100) NOT NULL,
  "code" character varying(50) NOT NULL,
  "description" text NULL,
  "icon" character varying(100) NULL,
  "is_active" boolean NULL DEFAULT true,
  "display_order" integer NULL DEFAULT 0,
  "metadata" jsonb NULL DEFAULT '{}',
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "modules_code_key" UNIQUE ("code")
);
-- Create index "idx_modules_code" to table: "modules"
CREATE INDEX "idx_modules_code" ON "public"."modules" ("code");
-- Create index "idx_modules_display_order" to table: "modules"
CREATE INDEX "idx_modules_display_order" ON "public"."modules" ("display_order");
-- Create index "idx_modules_is_active" to table: "modules"
CREATE INDEX "idx_modules_is_active" ON "public"."modules" ("is_active");
-- Create "menus" table
CREATE TABLE "public"."menus" (
  "id" serial NOT NULL,
  "module_id" integer NOT NULL,
  "parent_menu_id" integer NULL,
  "name" character varying(100) NOT NULL,
  "code" character varying(50) NOT NULL,
  "route_path" character varying(255) NULL,
  "icon" character varying(100) NULL,
  "display_order" integer NULL DEFAULT 0,
  "is_active" boolean NULL DEFAULT true,
  "metadata" jsonb NULL DEFAULT '{}',
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "menus_module_id_code_key" UNIQUE ("module_id", "code"),
  CONSTRAINT "menus_module_id_fkey" FOREIGN KEY ("module_id") REFERENCES "public"."modules" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "menus_parent_menu_id_fkey" FOREIGN KEY ("parent_menu_id") REFERENCES "public"."menus" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_menus_display_order" to table: "menus"
CREATE INDEX "idx_menus_display_order" ON "public"."menus" ("display_order");
-- Create index "idx_menus_is_active" to table: "menus"
CREATE INDEX "idx_menus_is_active" ON "public"."menus" ("is_active");
-- Create index "idx_menus_module_id" to table: "menus"
CREATE INDEX "idx_menus_module_id" ON "public"."menus" ("module_id");
-- Create index "idx_menus_parent_menu_id" to table: "menus"
CREATE INDEX "idx_menus_parent_menu_id" ON "public"."menus" ("parent_menu_id");
-- Create trigger "trg_menus_updated_at"
CREATE TRIGGER "trg_menus_updated_at" BEFORE UPDATE ON "public"."menus" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create trigger "update_menus_updated_at"
CREATE TRIGGER "update_menus_updated_at" BEFORE UPDATE ON "public"."menus" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create trigger "trg_modules_updated_at"
CREATE TRIGGER "trg_modules_updated_at" BEFORE UPDATE ON "public"."modules" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create trigger "update_modules_updated_at"
CREATE TRIGGER "update_modules_updated_at" BEFORE UPDATE ON "public"."modules" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create "order_fulfillments" table
CREATE TABLE "public"."order_fulfillments" (
  "id" uuid NOT NULL DEFAULT public.uuid_generate_v4(),
  "sales_order_id" uuid NOT NULL,
  "organization_id" integer NOT NULL,
  "fulfillment_number" character varying(50) NOT NULL,
  "fulfillment_status" character varying(50) NULL DEFAULT 'pending',
  "shipment_status" character varying(50) NULL DEFAULT 'pending',
  "fulfillment_store_id" integer NULL,
  "shipping_carrier" character varying(100) NULL,
  "shipping_method" character varying(100) NULL,
  "tracking_number" character varying(255) NULL,
  "tracking_url" text NULL,
  "picked_at" timestamp NULL,
  "packed_at" timestamp NULL,
  "shipped_at" timestamp NULL,
  "estimated_delivery_date" date NULL,
  "actual_delivery_date" date NULL,
  "picked_by_user_id" integer NULL,
  "packed_by_user_id" integer NULL,
  "notes" text NULL,
  "metadata" jsonb NULL DEFAULT '{}',
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "order_fulfillments_fulfillment_number_key" UNIQUE ("fulfillment_number"),
  CONSTRAINT "order_fulfillments_fulfillment_store_id_fkey" FOREIGN KEY ("fulfillment_store_id") REFERENCES "public"."stores" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "order_fulfillments_organization_id_fkey" FOREIGN KEY ("organization_id") REFERENCES "public"."organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "order_fulfillments_packed_by_user_id_fkey" FOREIGN KEY ("packed_by_user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "order_fulfillments_picked_by_user_id_fkey" FOREIGN KEY ("picked_by_user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "order_fulfillments_sales_order_id_fkey" FOREIGN KEY ("sales_order_id") REFERENCES "public"."sales_orders_v2" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_order_fulfillments_fulfillment_number" to table: "order_fulfillments"
CREATE INDEX "idx_order_fulfillments_fulfillment_number" ON "public"."order_fulfillments" ("fulfillment_number");
-- Create index "idx_order_fulfillments_fulfillment_status" to table: "order_fulfillments"
CREATE INDEX "idx_order_fulfillments_fulfillment_status" ON "public"."order_fulfillments" ("fulfillment_status");
-- Create index "idx_order_fulfillments_sales_order_id" to table: "order_fulfillments"
CREATE INDEX "idx_order_fulfillments_sales_order_id" ON "public"."order_fulfillments" ("sales_order_id");
-- Create index "idx_order_fulfillments_shipment_status" to table: "order_fulfillments"
CREATE INDEX "idx_order_fulfillments_shipment_status" ON "public"."order_fulfillments" ("shipment_status");
-- Set comment to table: "order_fulfillments"
COMMENT ON TABLE "public"."order_fulfillments" IS 'Shipment and fulfillment tracking for orders';
-- Create trigger "trg_order_fulfillments_updated_at"
CREATE TRIGGER "trg_order_fulfillments_updated_at" BEFORE UPDATE ON "public"."order_fulfillments" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create trigger "trg_organizations_updated_at"
CREATE TRIGGER "trg_organizations_updated_at" BEFORE UPDATE ON "public"."organizations" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create trigger "update_organizations_updated_at"
CREATE TRIGGER "update_organizations_updated_at" BEFORE UPDATE ON "public"."organizations" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create trigger "trg_pos_terminals_updated_at"
CREATE TRIGGER "trg_pos_terminals_updated_at" BEFORE UPDATE ON "public"."pos_terminals" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create trigger "update_pos_terminals_updated_at"
CREATE TRIGGER "update_pos_terminals_updated_at" BEFORE UPDATE ON "public"."pos_terminals" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create "pos_transactions" table
CREATE TABLE "public"."pos_transactions" (
  "id" serial NOT NULL,
  "store_id" integer NOT NULL,
  "cashier_id" integer NOT NULL,
  "cashier_session_id" integer NOT NULL,
  "customer_id" integer NULL,
  "pos_terminal_id" integer NULL,
  "transaction_number" character varying(50) NOT NULL,
  "transaction_date" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  "transaction_type" character varying(50) NULL,
  "subtotal" numeric(15,2) NULL DEFAULT 0,
  "discount_amount" numeric(15,2) NULL DEFAULT 0,
  "tax_amount" numeric(15,2) NULL DEFAULT 0,
  "total_amount" numeric(15,2) NULL DEFAULT 0,
  "total_cost" numeric(15,2) NULL DEFAULT 0,
  "amount_paid" numeric(15,2) NULL DEFAULT 0,
  "change_given" numeric(15,2) NULL DEFAULT 0,
  "status" character varying(50) NULL DEFAULT 'completed',
  "price_list_id" integer NULL,
  "sales_order_id" uuid NULL,
  "source_cart_id" uuid NULL,
  "voided_by" integer NULL,
  "voided_at" timestamp NULL,
  "metadata" jsonb NULL DEFAULT '{}',
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "pos_transactions_transaction_number_key" UNIQUE ("transaction_number"),
  CONSTRAINT "pos_transactions_cashier_id_fkey" FOREIGN KEY ("cashier_id") REFERENCES "public"."cashiers" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "pos_transactions_cashier_session_id_fkey" FOREIGN KEY ("cashier_session_id") REFERENCES "public"."cashier_sessions" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "pos_transactions_customer_id_fkey" FOREIGN KEY ("customer_id") REFERENCES "public"."customers" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "pos_transactions_pos_terminal_id_fkey" FOREIGN KEY ("pos_terminal_id") REFERENCES "public"."pos_terminals" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "pos_transactions_price_list_id_fkey" FOREIGN KEY ("price_list_id") REFERENCES "public"."price_lists" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "pos_transactions_sales_order_id_fkey" FOREIGN KEY ("sales_order_id") REFERENCES "public"."sales_orders_v2" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "pos_transactions_source_cart_id_fkey" FOREIGN KEY ("source_cart_id") REFERENCES "public"."carts" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "pos_transactions_store_id_fkey" FOREIGN KEY ("store_id") REFERENCES "public"."stores" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "pos_transactions_voided_by_fkey" FOREIGN KEY ("voided_by") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL
);
-- Create index "idx_pos_transactions_cashier_id" to table: "pos_transactions"
CREATE INDEX "idx_pos_transactions_cashier_id" ON "public"."pos_transactions" ("cashier_id");
-- Create index "idx_pos_transactions_cashier_session_id" to table: "pos_transactions"
CREATE INDEX "idx_pos_transactions_cashier_session_id" ON "public"."pos_transactions" ("cashier_session_id");
-- Create index "idx_pos_transactions_customer_id" to table: "pos_transactions"
CREATE INDEX "idx_pos_transactions_customer_id" ON "public"."pos_transactions" ("customer_id");
-- Create index "idx_pos_transactions_status" to table: "pos_transactions"
CREATE INDEX "idx_pos_transactions_status" ON "public"."pos_transactions" ("status");
-- Create index "idx_pos_transactions_store_id" to table: "pos_transactions"
CREATE INDEX "idx_pos_transactions_store_id" ON "public"."pos_transactions" ("store_id");
-- Create index "idx_pos_transactions_transaction_date" to table: "pos_transactions"
CREATE INDEX "idx_pos_transactions_transaction_date" ON "public"."pos_transactions" ("transaction_date");
-- Create index "idx_pos_transactions_transaction_number" to table: "pos_transactions"
CREATE INDEX "idx_pos_transactions_transaction_number" ON "public"."pos_transactions" ("transaction_number");
-- Create trigger "trg_pos_transactions_updated_at"
CREATE TRIGGER "trg_pos_transactions_updated_at" BEFORE UPDATE ON "public"."pos_transactions" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create trigger "trg_price_lists_updated_at"
CREATE TRIGGER "trg_price_lists_updated_at" BEFORE UPDATE ON "public"."price_lists" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create trigger "update_price_lists_updated_at"
CREATE TRIGGER "update_price_lists_updated_at" BEFORE UPDATE ON "public"."price_lists" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create "product_batches" table
CREATE TABLE "public"."product_batches" (
  "id" serial NOT NULL,
  "product_id" integer NOT NULL,
  "product_variant_id" integer NULL,
  "batch_number" character varying(100) NOT NULL,
  "manufacturing_date" date NULL,
  "expiry_date" date NULL,
  "store_id" integer NULL,
  "quantity_available" numeric(15,3) NULL DEFAULT 0,
  "status" character varying(50) NULL DEFAULT 'active',
  "metadata" jsonb NULL DEFAULT '{}',
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "product_batches_product_id_batch_number_store_id_key" UNIQUE ("product_id", "batch_number", "store_id"),
  CONSTRAINT "product_batches_product_id_fkey" FOREIGN KEY ("product_id") REFERENCES "public"."products" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "product_batches_product_variant_id_fkey" FOREIGN KEY ("product_variant_id") REFERENCES "public"."product_variants" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "product_batches_store_id_fkey" FOREIGN KEY ("store_id") REFERENCES "public"."stores" ("id") ON UPDATE NO ACTION ON DELETE SET NULL
);
-- Create index "idx_product_batches_batch_number" to table: "product_batches"
CREATE INDEX "idx_product_batches_batch_number" ON "public"."product_batches" ("batch_number");
-- Create index "idx_product_batches_expiry_date" to table: "product_batches"
CREATE INDEX "idx_product_batches_expiry_date" ON "public"."product_batches" ("expiry_date");
-- Create index "idx_product_batches_product_id" to table: "product_batches"
CREATE INDEX "idx_product_batches_product_id" ON "public"."product_batches" ("product_id");
-- Create index "idx_product_batches_status" to table: "product_batches"
CREATE INDEX "idx_product_batches_status" ON "public"."product_batches" ("status");
-- Create index "idx_product_batches_store_id" to table: "product_batches"
CREATE INDEX "idx_product_batches_store_id" ON "public"."product_batches" ("store_id");
-- Create trigger "trg_product_batches_updated_at"
CREATE TRIGGER "trg_product_batches_updated_at" BEFORE UPDATE ON "public"."product_batches" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create trigger "update_product_batches_updated_at"
CREATE TRIGGER "update_product_batches_updated_at" BEFORE UPDATE ON "public"."product_batches" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create trigger "trg_product_categories_updated_at"
CREATE TRIGGER "trg_product_categories_updated_at" BEFORE UPDATE ON "public"."product_categories" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create trigger "update_product_categories_updated_at"
CREATE TRIGGER "update_product_categories_updated_at" BEFORE UPDATE ON "public"."product_categories" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create "product_prices" table
CREATE TABLE "public"."product_prices" (
  "id" serial NOT NULL,
  "product_id" integer NOT NULL,
  "product_variant_id" integer NULL,
  "price_list_id" integer NOT NULL,
  "uom_id" integer NULL,
  "price" numeric(15,2) NOT NULL,
  "min_quantity" numeric(15,3) NULL DEFAULT 1,
  "max_quantity" numeric(15,3) NULL,
  "valid_from" date NULL,
  "valid_to" date NULL,
  "is_active" boolean NULL DEFAULT true,
  "metadata" jsonb NULL DEFAULT '{}',
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "product_prices_price_list_id_fkey" FOREIGN KEY ("price_list_id") REFERENCES "public"."price_lists" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "product_prices_product_id_fkey" FOREIGN KEY ("product_id") REFERENCES "public"."products" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "product_prices_product_variant_id_fkey" FOREIGN KEY ("product_variant_id") REFERENCES "public"."product_variants" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "product_prices_uom_id_fkey" FOREIGN KEY ("uom_id") REFERENCES "public"."units_of_measure" ("id") ON UPDATE NO ACTION ON DELETE SET NULL
);
-- Create index "idx_product_prices_is_active" to table: "product_prices"
CREATE INDEX "idx_product_prices_is_active" ON "public"."product_prices" ("is_active");
-- Create index "idx_product_prices_price_list_id" to table: "product_prices"
CREATE INDEX "idx_product_prices_price_list_id" ON "public"."product_prices" ("price_list_id");
-- Create index "idx_product_prices_product_id" to table: "product_prices"
CREATE INDEX "idx_product_prices_product_id" ON "public"."product_prices" ("product_id");
-- Create index "idx_product_prices_product_variant_id" to table: "product_prices"
CREATE INDEX "idx_product_prices_product_variant_id" ON "public"."product_prices" ("product_variant_id");
-- Create trigger "trg_product_prices_updated_at"
CREATE TRIGGER "trg_product_prices_updated_at" BEFORE UPDATE ON "public"."product_prices" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create trigger "update_product_prices_updated_at"
CREATE TRIGGER "update_product_prices_updated_at" BEFORE UPDATE ON "public"."product_prices" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create "product_serial_numbers" table
CREATE TABLE "public"."product_serial_numbers" (
  "id" serial NOT NULL,
  "product_id" integer NOT NULL,
  "product_variant_id" integer NULL,
  "serial_number" character varying(100) NOT NULL,
  "status" character varying(50) NULL DEFAULT 'in_stock',
  "current_store_id" integer NULL,
  "manufacturing_date" date NULL,
  "expiry_date" date NULL,
  "metadata" jsonb NULL DEFAULT '{}',
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "product_serial_numbers_serial_number_key" UNIQUE ("serial_number"),
  CONSTRAINT "product_serial_numbers_current_store_id_fkey" FOREIGN KEY ("current_store_id") REFERENCES "public"."stores" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "product_serial_numbers_product_id_fkey" FOREIGN KEY ("product_id") REFERENCES "public"."products" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "product_serial_numbers_product_variant_id_fkey" FOREIGN KEY ("product_variant_id") REFERENCES "public"."product_variants" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_product_serial_numbers_current_store_id" to table: "product_serial_numbers"
CREATE INDEX "idx_product_serial_numbers_current_store_id" ON "public"."product_serial_numbers" ("current_store_id");
-- Create index "idx_product_serial_numbers_product_id" to table: "product_serial_numbers"
CREATE INDEX "idx_product_serial_numbers_product_id" ON "public"."product_serial_numbers" ("product_id");
-- Create index "idx_product_serial_numbers_serial_number" to table: "product_serial_numbers"
CREATE INDEX "idx_product_serial_numbers_serial_number" ON "public"."product_serial_numbers" ("serial_number");
-- Create index "idx_product_serial_numbers_status" to table: "product_serial_numbers"
CREATE INDEX "idx_product_serial_numbers_status" ON "public"."product_serial_numbers" ("status");
-- Create trigger "trg_product_serial_numbers_updated_at"
CREATE TRIGGER "trg_product_serial_numbers_updated_at" BEFORE UPDATE ON "public"."product_serial_numbers" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create trigger "update_product_serial_numbers_updated_at"
CREATE TRIGGER "update_product_serial_numbers_updated_at" BEFORE UPDATE ON "public"."product_serial_numbers" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create trigger "trg_product_variants_updated_at"
CREATE TRIGGER "trg_product_variants_updated_at" BEFORE UPDATE ON "public"."product_variants" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create trigger "update_product_variants_updated_at"
CREATE TRIGGER "update_product_variants_updated_at" BEFORE UPDATE ON "public"."product_variants" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create trigger "trg_products_updated_at"
CREATE TRIGGER "trg_products_updated_at" BEFORE UPDATE ON "public"."products" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create trigger "update_products_updated_at"
CREATE TRIGGER "update_products_updated_at" BEFORE UPDATE ON "public"."products" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create "profit_loss_analytics" table
CREATE TABLE "public"."profit_loss_analytics" (
  "id" serial NOT NULL,
  "organization_id" integer NOT NULL,
  "store_id" integer NULL,
  "date" date NOT NULL,
  "period_type" character varying(20) NULL,
  "month" integer NULL,
  "quarter" integer NULL,
  "year" integer NULL,
  "gross_revenue" numeric(15,2) NULL DEFAULT 0,
  "sales_discounts" numeric(15,2) NULL DEFAULT 0,
  "sales_returns" numeric(15,2) NULL DEFAULT 0,
  "net_revenue" numeric(15,2) NULL DEFAULT 0,
  "opening_inventory_value" numeric(15,2) NULL DEFAULT 0,
  "purchases" numeric(15,2) NULL DEFAULT 0,
  "closing_inventory_value" numeric(15,2) NULL DEFAULT 0,
  "cogs" numeric(15,2) NULL DEFAULT 0,
  "gross_profit" numeric(15,2) NULL DEFAULT 0,
  "gross_profit_margin" numeric(5,2) NULL,
  "total_expenses" numeric(15,2) NULL DEFAULT 0,
  "net_profit" numeric(15,2) NULL DEFAULT 0,
  "net_profit_margin" numeric(5,2) NULL,
  "metadata" jsonb NULL DEFAULT '{}',
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_profit_loss_analytics_store" FOREIGN KEY ("store_id") REFERENCES "public"."stores" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "profit_loss_analytics_organization_id_fkey" FOREIGN KEY ("organization_id") REFERENCES "public"."organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_profit_loss_analytics_date" to table: "profit_loss_analytics"
CREATE INDEX "idx_profit_loss_analytics_date" ON "public"."profit_loss_analytics" ("date");
-- Create index "idx_profit_loss_analytics_organization_id" to table: "profit_loss_analytics"
CREATE INDEX "idx_profit_loss_analytics_organization_id" ON "public"."profit_loss_analytics" ("organization_id");
-- Create index "idx_profit_loss_analytics_period_type" to table: "profit_loss_analytics"
CREATE INDEX "idx_profit_loss_analytics_period_type" ON "public"."profit_loss_analytics" ("period_type");
-- Create index "idx_profit_loss_analytics_store_id" to table: "profit_loss_analytics"
CREATE INDEX "idx_profit_loss_analytics_store_id" ON "public"."profit_loss_analytics" ("store_id");
-- Create trigger "trg_profit_loss_analytics_updated_at"
CREATE TRIGGER "trg_profit_loss_analytics_updated_at" BEFORE UPDATE ON "public"."profit_loss_analytics" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create trigger "update_profit_loss_analytics_updated_at"
CREATE TRIGGER "update_profit_loss_analytics_updated_at" BEFORE UPDATE ON "public"."profit_loss_analytics" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create "fn_sync_promotion_to_product_prices" function
-- +goose StatementBegin
CREATE FUNCTION "public"."fn_sync_promotion_to_product_prices" () RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE
    v_promo_pl_id INTEGER;
    v_target_product_id INTEGER;
    v_retail_pp RECORD;
    v_calculated_price NUMERIC(15,2);
    v_discount_percent_str VARCHAR;
BEGIN
    -- Locate existing PROMO price list
    SELECT id INTO v_promo_pl_id
    FROM price_lists
    WHERE (code = 'PROMO' OR price_list_type = 'promotional')
      AND is_active = true
    ORDER BY id ASC
    LIMIT 1;

    -- If no PROMO price list exists, dynamically create one
    IF v_promo_pl_id IS NULL AND (TG_OP = 'INSERT' OR TG_OP = 'UPDATE') THEN
        INSERT INTO price_lists (code, name, price_list_type, currency_code, is_active)
        VALUES (
            'PROMO', 
            'Promotional Price List', 
            'promotional', 
            'SAR', 
            true
        )
        ON CONFLICT (code) DO UPDATE SET is_active = true
        RETURNING id INTO v_promo_pl_id;
    END IF;

    -- If DELETING or DEACTIVATING promotion, remove/deactivate promo package prices
    IF TG_OP = 'DELETE' OR (TG_OP = 'UPDATE' AND NEW.is_active = false) THEN
        DELETE FROM product_prices
        WHERE price_list_id = v_promo_pl_id
          AND metadata->>'promotion_id' = COALESCE(OLD.id, NEW.id)::text;
        
        IF TG_OP = 'DELETE' THEN
            RETURN OLD;
        END IF;
        RETURN NEW;
    END IF;

    -- If INSERTING or UPDATING active promotion, generate promotional package prices
    IF NEW.is_active = true AND v_promo_pl_id IS NOT NULL THEN
        -- Format discount percent string for metadata
        IF NEW.promotion_type = 'percentage_discount' AND NEW.discount_value IS NOT NULL THEN
            v_discount_percent_str := CONCAT(TRIM(TRAILING '.' FROM TRIM(TRAILING '0' FROM NEW.discount_value::text)), '%');
        ELSE
            v_discount_percent_str := NULL;
        END IF;

        -- Find all target products matching the promotion criteria
        FOR v_target_product_id IN (
            SELECT p.id
            FROM products p
            WHERE p.organization_id = NEW.organization_id
              AND p.is_active = true
              AND (
                  NEW.applies_to = 'all'
                  OR (NEW.applies_to = 'product' AND p.id = ANY(NEW.target_product_ids))
                  OR (NEW.applies_to = 'category' AND p.category_id = ANY(NEW.target_category_ids))
              )
        ) LOOP
            -- Loop through existing retail package prices for this product to compute promotional price per UOM
            FOR v_retail_pp IN (
                SELECT pp.*
                FROM product_prices pp
                JOIN price_lists pl ON pp.price_list_id = pl.id
                WHERE pp.product_id = v_target_product_id
                  AND pl.price_list_type = 'retail'
                  AND pp.is_active = true
            ) LOOP
                -- Calculate promo price
                IF NEW.promotion_type = 'percentage_discount' AND NEW.discount_value IS NOT NULL THEN
                    v_calculated_price := ROUND(v_retail_pp.price * (1.0 - (NEW.discount_value / 100.0)), 2);
                ELSIF NEW.promotion_type = 'fixed_discount' AND NEW.discount_value IS NOT NULL THEN
                    v_calculated_price := GREATEST(0.00, v_retail_pp.price - NEW.discount_value);
                ELSE
                    v_calculated_price := v_retail_pp.price;
                END IF;

                -- Remove previous promo price entry for this product + UOM + PROMO price list if it exists
                DELETE FROM product_prices
                WHERE product_id = v_target_product_id
                  AND price_list_id = v_promo_pl_id
                  AND (product_variant_id = v_retail_pp.product_variant_id OR (product_variant_id IS NULL AND v_retail_pp.product_variant_id IS NULL))
                  AND (uom_id = v_retail_pp.uom_id OR (uom_id IS NULL AND v_retail_pp.uom_id IS NULL));

                -- Insert new promotional package price into product_prices
                INSERT INTO product_prices (
                    product_id,
                    product_variant_id,
                    price_list_id,
                    uom_id,
                    price,
                    min_quantity,
                    max_quantity,
                    valid_from,
                    valid_to,
                    is_active,
                    metadata
                ) VALUES (
                    v_target_product_id,
                    v_retail_pp.product_variant_id,
                    v_promo_pl_id,
                    v_retail_pp.uom_id,
                    v_calculated_price,
                    v_retail_pp.min_quantity,
                    v_retail_pp.max_quantity,
                    NEW.valid_from,
                    NEW.valid_to,
                    true,
                    jsonb_build_object(
                        'promotion_id', NEW.id,
                        'promotion_name', NEW.name,
                        'discount_percent', v_discount_percent_str
                    )
                );
            END LOOP;
        END LOOP;
    END IF;

    RETURN NEW;
END;
$$;
-- +goose StatementEnd
-- Create "promotions" table
CREATE TABLE "public"."promotions" (
  "id" serial NOT NULL,
  "organization_id" integer NOT NULL,
  "code" character varying(50) NOT NULL,
  "name" character varying(255) NOT NULL,
  "description" text NULL,
  "promotion_type" character varying(50) NOT NULL,
  "action_metadata" jsonb NULL DEFAULT '{}',
  "valid_from" timestamp NULL,
  "valid_to" timestamp NULL,
  "schedule_json" jsonb NULL DEFAULT '{}',
  "applies_to" character varying(50) NULL DEFAULT 'all',
  "target_product_ids" integer[] NULL DEFAULT '{}',
  "target_category_ids" integer[] NULL DEFAULT '{}',
  "target_customer_types" text[] NULL DEFAULT '{}',
  "min_order_amount" numeric(15,2) NULL,
  "min_quantity" numeric(15,3) NULL,
  "coupon_code" character varying(50) NULL,
  "usage_limit" integer NULL,
  "usage_count" integer NULL DEFAULT 0,
  "usage_per_customer" integer NULL,
  "discount_value" numeric(15,4) NULL,
  "is_stackable" boolean NULL DEFAULT false,
  "is_active" boolean NULL DEFAULT true,
  "store_ids" integer[] NULL DEFAULT '{}',
  "created_by" integer NULL,
  "metadata" jsonb NULL DEFAULT '{}',
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "promotions_organization_id_code_key" UNIQUE ("organization_id", "code"),
  CONSTRAINT "promotions_created_by_fkey" FOREIGN KEY ("created_by") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "promotions_organization_id_fkey" FOREIGN KEY ("organization_id") REFERENCES "public"."organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "promotions_applies_to_check" CHECK ((applies_to)::text = ANY ((ARRAY['all'::character varying, 'category'::character varying, 'product'::character varying, 'customer_type'::character varying, 'price_list'::character varying])::text[])),
  CONSTRAINT "promotions_promotion_type_check" CHECK ((promotion_type)::text = ANY ((ARRAY['percentage_discount'::character varying, 'fixed_discount'::character varying, 'bogo'::character varying, 'buy_x_get_y'::character varying, 'free_item'::character varying, 'bundle_price'::character varying, 'points_multiplier'::character varying, 'happy_hour'::character varying])::text[]))
);
-- Create trigger "trg_sync_promotion_to_product_prices"
CREATE TRIGGER "trg_sync_promotion_to_product_prices" AFTER DELETE OR INSERT OR UPDATE ON "public"."promotions" FOR EACH ROW EXECUTE FUNCTION "public"."fn_sync_promotion_to_product_prices"();
-- Create trigger "trg_promotions_updated_at"
CREATE TRIGGER "trg_promotions_updated_at" BEFORE UPDATE ON "public"."promotions" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create "purchase_analytics" table
CREATE TABLE "public"."purchase_analytics" (
  "id" serial NOT NULL,
  "organization_id" integer NOT NULL,
  "store_id" integer NULL,
  "supplier_id" integer NULL,
  "product_id" integer NULL,
  "category_id" integer NULL,
  "date" date NOT NULL,
  "month" integer NULL,
  "quarter" integer NULL,
  "year" integer NULL,
  "units_purchased" numeric(15,3) NULL DEFAULT 0,
  "total_cost" numeric(15,2) NULL DEFAULT 0,
  "discounts" numeric(15,2) NULL DEFAULT 0,
  "taxes" numeric(15,2) NULL DEFAULT 0,
  "net_cost" numeric(15,2) NULL DEFAULT 0,
  "orders" integer NULL DEFAULT 0,
  "total_orders" integer NULL DEFAULT 0,
  "total_quantity" numeric(15,3) NULL DEFAULT 0,
  "total_amount" numeric(15,2) NULL DEFAULT 0,
  "discounts_received" numeric(15,2) NULL DEFAULT 0,
  "taxes_paid" numeric(15,2) NULL DEFAULT 0,
  "net_amount" numeric(15,2) NULL DEFAULT 0,
  "average_order_value" numeric(15,2) NULL,
  "metadata" jsonb NULL DEFAULT '{}',
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_purchase_analytics_category" FOREIGN KEY ("category_id") REFERENCES "public"."product_categories" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "fk_purchase_analytics_product" FOREIGN KEY ("product_id") REFERENCES "public"."products" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "fk_purchase_analytics_store" FOREIGN KEY ("store_id") REFERENCES "public"."stores" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "fk_purchase_analytics_supplier" FOREIGN KEY ("supplier_id") REFERENCES "public"."suppliers" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "purchase_analytics_organization_id_fkey" FOREIGN KEY ("organization_id") REFERENCES "public"."organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_purchase_analytics_date" to table: "purchase_analytics"
CREATE INDEX "idx_purchase_analytics_date" ON "public"."purchase_analytics" ("date");
-- Create index "idx_purchase_analytics_organization_id" to table: "purchase_analytics"
CREATE INDEX "idx_purchase_analytics_organization_id" ON "public"."purchase_analytics" ("organization_id");
-- Create index "idx_purchase_analytics_product_id" to table: "purchase_analytics"
CREATE INDEX "idx_purchase_analytics_product_id" ON "public"."purchase_analytics" ("product_id");
-- Create index "idx_purchase_analytics_store_id" to table: "purchase_analytics"
CREATE INDEX "idx_purchase_analytics_store_id" ON "public"."purchase_analytics" ("store_id");
-- Create index "idx_purchase_analytics_supplier_id" to table: "purchase_analytics"
CREATE INDEX "idx_purchase_analytics_supplier_id" ON "public"."purchase_analytics" ("supplier_id");
-- Create trigger "trg_purchase_analytics_updated_at"
CREATE TRIGGER "trg_purchase_analytics_updated_at" BEFORE UPDATE ON "public"."purchase_analytics" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create trigger "update_purchase_analytics_updated_at"
CREATE TRIGGER "update_purchase_analytics_updated_at" BEFORE UPDATE ON "public"."purchase_analytics" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create trigger "trg_purchase_orders_updated_at"
CREATE TRIGGER "trg_purchase_orders_updated_at" BEFORE UPDATE ON "public"."purchase_orders" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create trigger "update_purchase_orders_updated_at"
CREATE TRIGGER "update_purchase_orders_updated_at" BEFORE UPDATE ON "public"."purchase_orders" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create "quotes" table
CREATE TABLE "public"."quotes" (
  "id" uuid NOT NULL DEFAULT public.uuid_generate_v4(),
  "quote_number" character varying(50) NOT NULL,
  "organization_id" integer NOT NULL,
  "store_id" integer NULL,
  "customer_id" integer NULL,
  "customer_name" character varying(255) NOT NULL,
  "customer_email" character varying(255) NULL,
  "customer_phone" character varying(50) NULL,
  "quote_status" "public"."quote_status" NOT NULL DEFAULT 'draft',
  "quote_date" date NOT NULL DEFAULT CURRENT_DATE,
  "valid_until" date NOT NULL,
  "sent_date" date NULL,
  "accepted_date" date NULL,
  "converted_date" date NULL,
  "subtotal" numeric(15,2) NULL DEFAULT 0.00,
  "discount_amount" numeric(15,2) NULL DEFAULT 0.00,
  "tax_amount" numeric(15,2) NULL DEFAULT 0.00,
  "total_amount" numeric(15,2) NOT NULL,
  "converted_to_order_id" uuid NULL,
  "payment_terms" character varying(100) NULL,
  "delivery_terms" text NULL,
  "terms_and_conditions" text NULL,
  "notes" text NULL,
  "internal_notes" text NULL,
  "created_by_user_id" integer NULL,
  "metadata" jsonb NULL DEFAULT '{}',
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "quotes_quote_number_key" UNIQUE ("quote_number"),
  CONSTRAINT "quotes_converted_to_order_id_fkey" FOREIGN KEY ("converted_to_order_id") REFERENCES "public"."sales_orders_v2" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "quotes_created_by_user_id_fkey" FOREIGN KEY ("created_by_user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "quotes_customer_id_fkey" FOREIGN KEY ("customer_id") REFERENCES "public"."customers" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "quotes_organization_id_fkey" FOREIGN KEY ("organization_id") REFERENCES "public"."organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "quotes_store_id_fkey" FOREIGN KEY ("store_id") REFERENCES "public"."stores" ("id") ON UPDATE NO ACTION ON DELETE SET NULL
);
-- Create index "idx_quotes_converted_to_order_id" to table: "quotes"
CREATE INDEX "idx_quotes_converted_to_order_id" ON "public"."quotes" ("converted_to_order_id");
-- Create index "idx_quotes_customer_id" to table: "quotes"
CREATE INDEX "idx_quotes_customer_id" ON "public"."quotes" ("customer_id");
-- Create index "idx_quotes_organization_id" to table: "quotes"
CREATE INDEX "idx_quotes_organization_id" ON "public"."quotes" ("organization_id");
-- Create index "idx_quotes_quote_date" to table: "quotes"
CREATE INDEX "idx_quotes_quote_date" ON "public"."quotes" ("quote_date");
-- Create index "idx_quotes_quote_number" to table: "quotes"
CREATE INDEX "idx_quotes_quote_number" ON "public"."quotes" ("quote_number");
-- Create index "idx_quotes_quote_status" to table: "quotes"
CREATE INDEX "idx_quotes_quote_status" ON "public"."quotes" ("quote_status");
-- Create index "idx_quotes_valid_until" to table: "quotes"
CREATE INDEX "idx_quotes_valid_until" ON "public"."quotes" ("valid_until");
-- Set comment to table: "quotes"
COMMENT ON TABLE "public"."quotes" IS 'Sales quotations with approval workflow';
-- Create "quote_lines" table
CREATE TABLE "public"."quote_lines" (
  "id" uuid NOT NULL DEFAULT public.uuid_generate_v4(),
  "quote_id" uuid NOT NULL,
  "organization_id" integer NOT NULL,
  "line_number" integer NOT NULL,
  "product_id" integer NULL,
  "product_variant_id" integer NULL,
  "description" text NOT NULL,
  "quantity" numeric(15,3) NOT NULL,
  "unit_price" numeric(15,2) NOT NULL,
  "discount_amount" numeric(15,2) NULL DEFAULT 0.00,
  "tax_amount" numeric(15,2) NULL DEFAULT 0.00,
  "line_total" numeric(15,2) NOT NULL,
  "uom_id" integer NULL,
  "notes" text NULL,
  "metadata" jsonb NULL DEFAULT '{}',
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "quote_lines_quote_id_line_number_key" UNIQUE ("quote_id", "line_number"),
  CONSTRAINT "quote_lines_organization_id_fkey" FOREIGN KEY ("organization_id") REFERENCES "public"."organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "quote_lines_product_id_fkey" FOREIGN KEY ("product_id") REFERENCES "public"."products" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "quote_lines_product_variant_id_fkey" FOREIGN KEY ("product_variant_id") REFERENCES "public"."product_variants" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "quote_lines_quote_id_fkey" FOREIGN KEY ("quote_id") REFERENCES "public"."quotes" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "quote_lines_uom_id_fkey" FOREIGN KEY ("uom_id") REFERENCES "public"."units_of_measure" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "quote_lines_quantity_check" CHECK (quantity > (0)::numeric)
);
-- Create index "idx_quote_lines_product_id" to table: "quote_lines"
CREATE INDEX "idx_quote_lines_product_id" ON "public"."quote_lines" ("product_id");
-- Create index "idx_quote_lines_quote_id" to table: "quote_lines"
CREATE INDEX "idx_quote_lines_quote_id" ON "public"."quote_lines" ("quote_id");
-- Create trigger "trg_quote_lines_updated_at"
CREATE TRIGGER "trg_quote_lines_updated_at" BEFORE UPDATE ON "public"."quote_lines" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create trigger "trg_quotes_updated_at"
CREATE TRIGGER "trg_quotes_updated_at" BEFORE UPDATE ON "public"."quotes" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create trigger "trg_recipes_updated_at"
CREATE TRIGGER "trg_recipes_updated_at" BEFORE UPDATE ON "public"."recipes" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create "restaurant_tables" table
CREATE TABLE "public"."restaurant_tables" (
  "id" serial NOT NULL,
  "store_id" integer NOT NULL,
  "table_number" character varying(20) NOT NULL,
  "table_name" character varying(100) NULL,
  "section" character varying(50) NULL,
  "capacity" integer NULL DEFAULT 4,
  "is_active" boolean NULL DEFAULT true,
  "metadata" jsonb NULL DEFAULT '{}',
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "restaurant_tables_store_id_table_number_key" UNIQUE ("store_id", "table_number"),
  CONSTRAINT "restaurant_tables_store_id_fkey" FOREIGN KEY ("store_id") REFERENCES "public"."stores" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_restaurant_tables_is_active" to table: "restaurant_tables"
CREATE INDEX "idx_restaurant_tables_is_active" ON "public"."restaurant_tables" ("is_active");
-- Create index "idx_restaurant_tables_section" to table: "restaurant_tables"
CREATE INDEX "idx_restaurant_tables_section" ON "public"."restaurant_tables" ("section");
-- Create index "idx_restaurant_tables_store_id" to table: "restaurant_tables"
CREATE INDEX "idx_restaurant_tables_store_id" ON "public"."restaurant_tables" ("store_id");
-- Create "restaurant_orders" table
CREATE TABLE "public"."restaurant_orders" (
  "id" serial NOT NULL,
  "store_id" integer NOT NULL,
  "table_id" integer NULL,
  "cashier_id" integer NULL,
  "cashier_session_id" integer NULL,
  "customer_id" integer NULL,
  "order_number" character varying(50) NOT NULL,
  "order_source" character varying(30) NOT NULL DEFAULT 'counter',
  "status" character varying(30) NOT NULL DEFAULT 'pending',
  "subtotal" numeric(15,2) NULL DEFAULT 0,
  "discount_amount" numeric(15,2) NULL DEFAULT 0,
  "tax_amount" numeric(15,2) NULL DEFAULT 0,
  "total_amount" numeric(15,2) NULL DEFAULT 0,
  "amount_paid" numeric(15,2) NULL DEFAULT 0,
  "change_given" numeric(15,2) NULL DEFAULT 0,
  "notes" text NULL,
  "pos_transaction_id" integer NULL,
  "ordered_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  "confirmed_at" timestamp NULL,
  "served_at" timestamp NULL,
  "paid_at" timestamp NULL,
  "metadata" jsonb NULL DEFAULT '{}',
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "restaurant_orders_store_id_order_number_key" UNIQUE ("store_id", "order_number"),
  CONSTRAINT "restaurant_orders_cashier_id_fkey" FOREIGN KEY ("cashier_id") REFERENCES "public"."cashiers" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "restaurant_orders_cashier_session_id_fkey" FOREIGN KEY ("cashier_session_id") REFERENCES "public"."cashier_sessions" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "restaurant_orders_customer_id_fkey" FOREIGN KEY ("customer_id") REFERENCES "public"."customers" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "restaurant_orders_pos_transaction_id_fkey" FOREIGN KEY ("pos_transaction_id") REFERENCES "public"."pos_transactions" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "restaurant_orders_store_id_fkey" FOREIGN KEY ("store_id") REFERENCES "public"."stores" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "restaurant_orders_table_id_fkey" FOREIGN KEY ("table_id") REFERENCES "public"."restaurant_tables" ("id") ON UPDATE NO ACTION ON DELETE SET NULL
);
-- Create index "idx_restaurant_orders_cashier_id" to table: "restaurant_orders"
CREATE INDEX "idx_restaurant_orders_cashier_id" ON "public"."restaurant_orders" ("cashier_id");
-- Create index "idx_restaurant_orders_customer_id" to table: "restaurant_orders"
CREATE INDEX "idx_restaurant_orders_customer_id" ON "public"."restaurant_orders" ("customer_id");
-- Create index "idx_restaurant_orders_ordered_at" to table: "restaurant_orders"
CREATE INDEX "idx_restaurant_orders_ordered_at" ON "public"."restaurant_orders" ("ordered_at");
-- Create index "idx_restaurant_orders_pos_txn_id" to table: "restaurant_orders"
CREATE INDEX "idx_restaurant_orders_pos_txn_id" ON "public"."restaurant_orders" ("pos_transaction_id");
-- Create index "idx_restaurant_orders_session_id" to table: "restaurant_orders"
CREATE INDEX "idx_restaurant_orders_session_id" ON "public"."restaurant_orders" ("cashier_session_id");
-- Create index "idx_restaurant_orders_source" to table: "restaurant_orders"
CREATE INDEX "idx_restaurant_orders_source" ON "public"."restaurant_orders" ("order_source");
-- Create index "idx_restaurant_orders_status" to table: "restaurant_orders"
CREATE INDEX "idx_restaurant_orders_status" ON "public"."restaurant_orders" ("status");
-- Create index "idx_restaurant_orders_store_id" to table: "restaurant_orders"
CREATE INDEX "idx_restaurant_orders_store_id" ON "public"."restaurant_orders" ("store_id");
-- Create index "idx_restaurant_orders_store_status_time" to table: "restaurant_orders"
CREATE INDEX "idx_restaurant_orders_store_status_time" ON "public"."restaurant_orders" ("store_id", "status", "ordered_at");
-- Create index "idx_restaurant_orders_table_id" to table: "restaurant_orders"
CREATE INDEX "idx_restaurant_orders_table_id" ON "public"."restaurant_orders" ("table_id");
-- Create "restaurant_order_items" table
CREATE TABLE "public"."restaurant_order_items" (
  "id" serial NOT NULL,
  "order_id" integer NOT NULL,
  "menu_item_id" integer NOT NULL,
  "quantity" numeric(15,3) NOT NULL DEFAULT 1,
  "unit_price" numeric(15,4) NOT NULL,
  "modifiers_snapshot" jsonb NULL DEFAULT '[]',
  "modifiers_total" numeric(15,2) NULL DEFAULT 0,
  "discount_amount" numeric(15,2) NULL DEFAULT 0,
  "tax_amount" numeric(15,2) NULL DEFAULT 0,
  "subtotal" numeric(15,2) NOT NULL,
  "line_number" integer NULL,
  "notes" text NULL,
  "status" character varying(30) NULL DEFAULT 'pending',
  "metadata" jsonb NULL DEFAULT '{}',
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "restaurant_order_items_menu_item_id_fkey" FOREIGN KEY ("menu_item_id") REFERENCES "public"."menu_items" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "restaurant_order_items_order_id_fkey" FOREIGN KEY ("order_id") REFERENCES "public"."restaurant_orders" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_restaurant_order_items_menu_item" to table: "restaurant_order_items"
CREATE INDEX "idx_restaurant_order_items_menu_item" ON "public"."restaurant_order_items" ("menu_item_id");
-- Create index "idx_restaurant_order_items_order_id" to table: "restaurant_order_items"
CREATE INDEX "idx_restaurant_order_items_order_id" ON "public"."restaurant_order_items" ("order_id");
-- Create index "idx_restaurant_order_items_status" to table: "restaurant_order_items"
CREATE INDEX "idx_restaurant_order_items_status" ON "public"."restaurant_order_items" ("status");
-- Create trigger "trg_restaurant_order_items_updated_at"
CREATE TRIGGER "trg_restaurant_order_items_updated_at" BEFORE UPDATE ON "public"."restaurant_order_items" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create trigger "trg_restaurant_orders_updated_at"
CREATE TRIGGER "trg_restaurant_orders_updated_at" BEFORE UPDATE ON "public"."restaurant_orders" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create trigger "trg_restaurant_tables_updated_at"
CREATE TRIGGER "trg_restaurant_tables_updated_at" BEFORE UPDATE ON "public"."restaurant_tables" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create "roles" table
CREATE TABLE "public"."roles" (
  "id" serial NOT NULL,
  "name" character varying(100) NOT NULL,
  "code" character varying(50) NOT NULL,
  "description" text NULL,
  "is_system_role" boolean NULL DEFAULT false,
  "is_active" boolean NULL DEFAULT true,
  "metadata" jsonb NULL DEFAULT '{}',
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "roles_code_key" UNIQUE ("code"),
  CONSTRAINT "roles_name_key" UNIQUE ("name")
);
-- Create index "idx_roles_code" to table: "roles"
CREATE INDEX "idx_roles_code" ON "public"."roles" ("code");
-- Create index "idx_roles_is_active" to table: "roles"
CREATE INDEX "idx_roles_is_active" ON "public"."roles" ("is_active");
-- Create "submenus" table
CREATE TABLE "public"."submenus" (
  "id" serial NOT NULL,
  "menu_id" integer NOT NULL,
  "parent_submenu_id" integer NULL,
  "name" character varying(100) NOT NULL,
  "code" character varying(50) NOT NULL,
  "route_path" character varying(255) NULL,
  "icon" character varying(100) NULL,
  "display_order" integer NULL DEFAULT 0,
  "is_active" boolean NULL DEFAULT true,
  "metadata" jsonb NULL DEFAULT '{}',
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "submenus_menu_id_code_key" UNIQUE ("menu_id", "code"),
  CONSTRAINT "submenus_menu_id_fkey" FOREIGN KEY ("menu_id") REFERENCES "public"."menus" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "submenus_parent_submenu_id_fkey" FOREIGN KEY ("parent_submenu_id") REFERENCES "public"."submenus" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_submenus_display_order" to table: "submenus"
CREATE INDEX "idx_submenus_display_order" ON "public"."submenus" ("display_order");
-- Create index "idx_submenus_is_active" to table: "submenus"
CREATE INDEX "idx_submenus_is_active" ON "public"."submenus" ("is_active");
-- Create index "idx_submenus_menu_id" to table: "submenus"
CREATE INDEX "idx_submenus_menu_id" ON "public"."submenus" ("menu_id");
-- Create index "idx_submenus_parent_submenu_id" to table: "submenus"
CREATE INDEX "idx_submenus_parent_submenu_id" ON "public"."submenus" ("parent_submenu_id");
-- Create "role_ui_customizations" table
CREATE TABLE "public"."role_ui_customizations" (
  "id" serial NOT NULL,
  "role_id" integer NOT NULL,
  "submenu_id" integer NOT NULL,
  "customization_data" jsonb NULL,
  "metadata" jsonb NULL DEFAULT '{}',
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "role_ui_customizations_role_id_submenu_id_key" UNIQUE ("role_id", "submenu_id"),
  CONSTRAINT "role_ui_customizations_role_id_fkey" FOREIGN KEY ("role_id") REFERENCES "public"."roles" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "role_ui_customizations_submenu_id_fkey" FOREIGN KEY ("submenu_id") REFERENCES "public"."submenus" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create trigger "trg_role_ui_customizations_updated_at"
CREATE TRIGGER "trg_role_ui_customizations_updated_at" BEFORE UPDATE ON "public"."role_ui_customizations" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create trigger "update_role_ui_customizations_updated_at"
CREATE TRIGGER "update_role_ui_customizations_updated_at" BEFORE UPDATE ON "public"."role_ui_customizations" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create trigger "trg_roles_updated_at"
CREATE TRIGGER "trg_roles_updated_at" BEFORE UPDATE ON "public"."roles" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create trigger "update_roles_updated_at"
CREATE TRIGGER "update_roles_updated_at" BEFORE UPDATE ON "public"."roles" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create "sales_analytics" table
CREATE TABLE "public"."sales_analytics" (
  "id" serial NOT NULL,
  "organization_id" integer NOT NULL,
  "store_id" integer NULL,
  "product_id" integer NULL,
  "category_id" integer NULL,
  "customer_id" integer NULL,
  "date" date NOT NULL,
  "hour" integer NULL,
  "day_of_week" integer NULL,
  "month" integer NULL,
  "quarter" integer NULL,
  "year" integer NULL,
  "units_sold" numeric(15,3) NULL DEFAULT 0,
  "revenue" numeric(15,2) NULL DEFAULT 0,
  "discounts" numeric(15,2) NULL DEFAULT 0,
  "taxes" numeric(15,2) NULL DEFAULT 0,
  "net_revenue" numeric(15,2) NULL DEFAULT 0,
  "transactions" integer NULL DEFAULT 0,
  "payment_method" character varying(50) NULL,
  "payment_gateway" character varying(50) NULL,
  "average_order_value" numeric(15,2) NULL,
  "metadata" jsonb NULL DEFAULT '{}',
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_sales_analytics_category" FOREIGN KEY ("category_id") REFERENCES "public"."product_categories" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "fk_sales_analytics_customer" FOREIGN KEY ("customer_id") REFERENCES "public"."customers" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "fk_sales_analytics_product" FOREIGN KEY ("product_id") REFERENCES "public"."products" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "fk_sales_analytics_store" FOREIGN KEY ("store_id") REFERENCES "public"."stores" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "sales_analytics_organization_id_fkey" FOREIGN KEY ("organization_id") REFERENCES "public"."organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_sales_analytics_category_id" to table: "sales_analytics"
CREATE INDEX "idx_sales_analytics_category_id" ON "public"."sales_analytics" ("category_id");
-- Create index "idx_sales_analytics_customer_id" to table: "sales_analytics"
CREATE INDEX "idx_sales_analytics_customer_id" ON "public"."sales_analytics" ("customer_id");
-- Create index "idx_sales_analytics_date" to table: "sales_analytics"
CREATE INDEX "idx_sales_analytics_date" ON "public"."sales_analytics" ("date");
-- Create index "idx_sales_analytics_organization_id" to table: "sales_analytics"
CREATE INDEX "idx_sales_analytics_organization_id" ON "public"."sales_analytics" ("organization_id");
-- Create index "idx_sales_analytics_product_id" to table: "sales_analytics"
CREATE INDEX "idx_sales_analytics_product_id" ON "public"."sales_analytics" ("product_id");
-- Create index "idx_sales_analytics_store_id" to table: "sales_analytics"
CREATE INDEX "idx_sales_analytics_store_id" ON "public"."sales_analytics" ("store_id");
-- Create index "idx_sales_analytics_year_month" to table: "sales_analytics"
CREATE INDEX "idx_sales_analytics_year_month" ON "public"."sales_analytics" ("year", "month");
-- Create trigger "trg_sales_analytics_updated_at"
CREATE TRIGGER "trg_sales_analytics_updated_at" BEFORE UPDATE ON "public"."sales_analytics" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create trigger "update_sales_analytics_updated_at"
CREATE TRIGGER "update_sales_analytics_updated_at" BEFORE UPDATE ON "public"."sales_analytics" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create "calculate_order_totals" function
-- +goose StatementBegin
CREATE FUNCTION "public"."calculate_order_totals" () RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE
    v_subtotal DECIMAL(15,2);
    v_tax_amount DECIMAL(15,2);
BEGIN
    -- Calculate subtotal and tax from order lines
    SELECT 
        COALESCE(SUM(line_total - tax_amount), 0),
        COALESCE(SUM(tax_amount), 0)
    INTO v_subtotal, v_tax_amount
    FROM sales_order_lines_v2
    WHERE sales_order_id = COALESCE(NEW.sales_order_id, OLD.sales_order_id);
    
    -- Update the order
    UPDATE sales_orders_v2
    SET 
        subtotal = v_subtotal,
        tax_amount = v_tax_amount,
        total_amount = v_subtotal + v_tax_amount + COALESCE(shipping_amount, 0) + COALESCE(adjustment_amount, 0) - COALESCE(discount_amount, 0),
        balance_due = (v_subtotal + v_tax_amount + COALESCE(shipping_amount, 0) + COALESCE(adjustment_amount, 0) - COALESCE(discount_amount, 0)) - COALESCE(paid_amount, 0)
    WHERE id = COALESCE(NEW.sales_order_id, OLD.sales_order_id);
    
    RETURN COALESCE(NEW, OLD);
END;
$$;
-- +goose StatementEnd
-- Create trigger "calculate_order_totals_trigger"
CREATE TRIGGER "calculate_order_totals_trigger" AFTER DELETE OR INSERT OR UPDATE ON "public"."sales_order_lines_v2" FOR EACH ROW EXECUTE FUNCTION "public"."calculate_order_totals"();
-- Create "stock_movements" table
CREATE TABLE "public"."stock_movements" (
  "id" serial NOT NULL,
  "movement_type" character varying(50) NOT NULL,
  "reference_type" character varying(50) NULL,
  "reference_id" integer NULL,
  "product_id" integer NOT NULL,
  "product_variant_id" integer NULL,
  "from_store_id" integer NULL,
  "to_store_id" integer NULL,
  "from_location_id" integer NULL,
  "to_location_id" integer NULL,
  "quantity" numeric(15,3) NOT NULL,
  "uom_id" integer NULL,
  "batch_number" character varying(100) NULL,
  "serial_number" character varying(100) NULL,
  "movement_date" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  "posted_by" integer NULL,
  "status" character varying(50) NULL DEFAULT 'completed',
  "cost_per_unit" numeric(15,4) NULL,
  "total_value" numeric(15,2) NULL,
  "metadata" jsonb NULL DEFAULT '{}',
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "stock_movements_from_location_id_fkey" FOREIGN KEY ("from_location_id") REFERENCES "public"."storage_locations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "stock_movements_from_store_id_fkey" FOREIGN KEY ("from_store_id") REFERENCES "public"."stores" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "stock_movements_posted_by_fkey" FOREIGN KEY ("posted_by") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "stock_movements_product_id_fkey" FOREIGN KEY ("product_id") REFERENCES "public"."products" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "stock_movements_product_variant_id_fkey" FOREIGN KEY ("product_variant_id") REFERENCES "public"."product_variants" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "stock_movements_to_location_id_fkey" FOREIGN KEY ("to_location_id") REFERENCES "public"."storage_locations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "stock_movements_to_store_id_fkey" FOREIGN KEY ("to_store_id") REFERENCES "public"."stores" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "stock_movements_uom_id_fkey" FOREIGN KEY ("uom_id") REFERENCES "public"."units_of_measure" ("id") ON UPDATE NO ACTION ON DELETE SET NULL
);
-- Create index "idx_stock_movements_from_store_id" to table: "stock_movements"
CREATE INDEX "idx_stock_movements_from_store_id" ON "public"."stock_movements" ("from_store_id");
-- Create index "idx_stock_movements_movement_date" to table: "stock_movements"
CREATE INDEX "idx_stock_movements_movement_date" ON "public"."stock_movements" ("movement_date");
-- Create index "idx_stock_movements_movement_type" to table: "stock_movements"
CREATE INDEX "idx_stock_movements_movement_type" ON "public"."stock_movements" ("movement_type");
-- Create index "idx_stock_movements_product_id" to table: "stock_movements"
CREATE INDEX "idx_stock_movements_product_id" ON "public"."stock_movements" ("product_id");
-- Create index "idx_stock_movements_reference_type_id" to table: "stock_movements"
CREATE INDEX "idx_stock_movements_reference_type_id" ON "public"."stock_movements" ("reference_type", "reference_id");
-- Create index "idx_stock_movements_to_store_id" to table: "stock_movements"
CREATE INDEX "idx_stock_movements_to_store_id" ON "public"."stock_movements" ("to_store_id");
-- Create "fn_trigger_allocate_inventory_on_order_line" function
-- +goose StatementBegin
CREATE FUNCTION "public"."fn_trigger_allocate_inventory_on_order_line" () RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE
    v_order RECORD;
    v_quantity DECIMAL(15,3);
BEGIN
    -- Get the parent order
    SELECT order_status, store_id INTO v_order
    FROM sales_orders_v2
    WHERE id = NEW.sales_order_id;

    -- Only process if order status is 'pending' or 'confirmed'
    IF v_order.order_status IN ('pending', 'confirmed') AND v_order.store_id IS NOT NULL THEN
        v_quantity := NEW.quantity_ordered;
        
        IF v_quantity > 0 THEN
            -- Allocate stock (increase allocated, decrease available)
            UPDATE inventory_stock
            SET 
                quantity_allocated = quantity_allocated + v_quantity,
                quantity_available = GREATEST(0, quantity_on_hand - (quantity_allocated + v_quantity)),
                updated_at = CURRENT_TIMESTAMP
            WHERE product_id = NEW.product_id
              AND (product_variant_id = NEW.product_variant_id 
                   OR (product_variant_id IS NULL AND NEW.product_variant_id IS NULL))
              AND store_id = v_order.store_id;

            -- Record the allocation movement
            INSERT INTO stock_movements (
                movement_type,
                reference_type,
                reference_id,
                product_id,
                product_variant_id,
                from_store_id,
                quantity,
                uom_id,
                status,
                metadata
            )
            VALUES (
                'allocation',
                'sales_order',
                NULL,
                NEW.product_id,
                NEW.product_variant_id,
                v_order.store_id,
                v_quantity,
                NEW.uom_id,
                'completed',
                jsonb_build_object(
                    'sales_order_id', NEW.sales_order_id::TEXT,
                    'order_line_id', NEW.id::TEXT,
                    'order_status', v_order.order_status
                )
            );
        END IF;
    END IF;

    RETURN NEW;
END;
$$;
-- +goose StatementEnd
-- Create trigger "trg_allocate_inventory_on_order_line_insert"
CREATE TRIGGER "trg_allocate_inventory_on_order_line_insert" AFTER INSERT ON "public"."sales_order_lines_v2" FOR EACH ROW EXECUTE FUNCTION "public"."fn_trigger_allocate_inventory_on_order_line"();
-- Create trigger "trg_sales_order_lines_v2_updated_at"
CREATE TRIGGER "trg_sales_order_lines_v2_updated_at" BEFORE UPDATE ON "public"."sales_order_lines_v2" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create "sales_orders" table
CREATE TABLE "public"."sales_orders" (
  "id" serial NOT NULL,
  "organization_id" integer NOT NULL,
  "order_number" character varying(50) NOT NULL,
  "customer_id" integer NULL,
  "store_id" integer NOT NULL,
  "order_date" date NOT NULL,
  "delivery_date" date NULL,
  "status" character varying(50) NULL DEFAULT 'draft',
  "subtotal" numeric(15,2) NULL DEFAULT 0,
  "discount_amount" numeric(15,2) NULL DEFAULT 0,
  "tax_amount" numeric(15,2) NULL DEFAULT 0,
  "total_amount" numeric(15,2) NULL DEFAULT 0,
  "price_list_id" integer NULL,
  "created_by" integer NULL,
  "metadata" jsonb NULL DEFAULT '{}',
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "sales_orders_order_number_key" UNIQUE ("order_number"),
  CONSTRAINT "sales_orders_created_by_fkey" FOREIGN KEY ("created_by") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "sales_orders_customer_id_fkey" FOREIGN KEY ("customer_id") REFERENCES "public"."customers" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "sales_orders_organization_id_fkey" FOREIGN KEY ("organization_id") REFERENCES "public"."organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "sales_orders_price_list_id_fkey" FOREIGN KEY ("price_list_id") REFERENCES "public"."price_lists" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "sales_orders_store_id_fkey" FOREIGN KEY ("store_id") REFERENCES "public"."stores" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_sales_orders_customer_id" to table: "sales_orders"
CREATE INDEX "idx_sales_orders_customer_id" ON "public"."sales_orders" ("customer_id");
-- Create index "idx_sales_orders_order_date" to table: "sales_orders"
CREATE INDEX "idx_sales_orders_order_date" ON "public"."sales_orders" ("order_date");
-- Create index "idx_sales_orders_order_number" to table: "sales_orders"
CREATE INDEX "idx_sales_orders_order_number" ON "public"."sales_orders" ("order_number");
-- Create index "idx_sales_orders_organization_id" to table: "sales_orders"
CREATE INDEX "idx_sales_orders_organization_id" ON "public"."sales_orders" ("organization_id");
-- Create index "idx_sales_orders_status" to table: "sales_orders"
CREATE INDEX "idx_sales_orders_status" ON "public"."sales_orders" ("status");
-- Create index "idx_sales_orders_store_id" to table: "sales_orders"
CREATE INDEX "idx_sales_orders_store_id" ON "public"."sales_orders" ("store_id");
-- Create trigger "update_sales_orders_updated_at"
CREATE TRIGGER "update_sales_orders_updated_at" BEFORE UPDATE ON "public"."sales_orders" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create "stock_reservations" table
CREATE TABLE "public"."stock_reservations" (
  "id" serial NOT NULL,
  "reservation_number" character varying(50) NOT NULL,
  "product_id" integer NOT NULL,
  "product_variant_id" integer NULL,
  "store_id" integer NOT NULL,
  "reference_type" character varying(50) NOT NULL,
  "reference_id" character varying(100) NOT NULL,
  "quantity_reserved" numeric(15,3) NOT NULL,
  "reserved_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  "expires_at" timestamp NULL,
  "status" character varying(30) NULL DEFAULT 'active',
  "reserved_by" integer NULL,
  "notes" text NULL,
  "metadata" jsonb NULL DEFAULT '{}',
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "stock_reservations_reservation_number_key" UNIQUE ("reservation_number"),
  CONSTRAINT "stock_reservations_product_id_fkey" FOREIGN KEY ("product_id") REFERENCES "public"."products" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "stock_reservations_product_variant_id_fkey" FOREIGN KEY ("product_variant_id") REFERENCES "public"."product_variants" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "stock_reservations_reserved_by_fkey" FOREIGN KEY ("reserved_by") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "stock_reservations_store_id_fkey" FOREIGN KEY ("store_id") REFERENCES "public"."stores" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "stock_reservations_quantity_reserved_check" CHECK (quantity_reserved > (0)::numeric),
  CONSTRAINT "stock_reservations_reference_type_check" CHECK ((reference_type)::text = ANY ((ARRAY['sales_order'::character varying, 'pos_transaction'::character varying, 'cart'::character varying, 'transfer'::character varying, 'manual'::character varying])::text[])),
  CONSTRAINT "stock_reservations_status_check" CHECK ((status)::text = ANY ((ARRAY['active'::character varying, 'fulfilled'::character varying, 'cancelled'::character varying, 'expired'::character varying])::text[]))
);
-- Create index "idx_stock_reservations_expires_at" to table: "stock_reservations"
CREATE INDEX "idx_stock_reservations_expires_at" ON "public"."stock_reservations" ("expires_at");
-- Create index "idx_stock_reservations_product_id" to table: "stock_reservations"
CREATE INDEX "idx_stock_reservations_product_id" ON "public"."stock_reservations" ("product_id");
-- Create index "idx_stock_reservations_product_variant_id" to table: "stock_reservations"
CREATE INDEX "idx_stock_reservations_product_variant_id" ON "public"."stock_reservations" ("product_variant_id");
-- Create index "idx_stock_reservations_reference" to table: "stock_reservations"
CREATE INDEX "idx_stock_reservations_reference" ON "public"."stock_reservations" ("reference_type", "reference_id");
-- Create index "idx_stock_reservations_status" to table: "stock_reservations"
CREATE INDEX "idx_stock_reservations_status" ON "public"."stock_reservations" ("status");
-- Create index "idx_stock_reservations_store_id" to table: "stock_reservations"
CREATE INDEX "idx_stock_reservations_store_id" ON "public"."stock_reservations" ("store_id");
-- Create "fn_trigger_deduct_inventory_on_fulfillment" function
-- +goose StatementBegin
CREATE FUNCTION "public"."fn_trigger_deduct_inventory_on_fulfillment" () RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE
    v_order_line RECORD;
    v_fulfilled_qty DECIMAL(15,3);
    v_reservation RECORD;
BEGIN
    -- Only process when order status changes to 'fulfilled'
    IF NOT (OLD.order_status IS DISTINCT FROM NEW.order_status AND NEW.order_status = 'fulfilled') THEN
        RETURN NEW;
    END IF;
        
    -- Ensure store_id is present
    IF NEW.store_id IS NULL THEN
        RAISE WARNING 'Order % has no store_id, skipping inventory deduction', NEW.id;
        RETURN NEW;
    END IF;

    -- Loop through all order lines
    FOR v_order_line IN
        SELECT 
            id,
            product_id,
            product_variant_id,
            quantity_ordered,
            quantity_fulfilled,
            uom_id
        FROM sales_order_lines_v2
        WHERE sales_order_id = NEW.id
    LOOP
        -- Determine the fulfilled quantity
        IF v_order_line.quantity_fulfilled IS NOT NULL AND v_order_line.quantity_fulfilled > 0 THEN
            v_fulfilled_qty := v_order_line.quantity_fulfilled;
        ELSE
            v_fulfilled_qty := v_order_line.quantity_ordered;
        END IF;
        
        IF v_fulfilled_qty <= 0 THEN
            CONTINUE;
        END IF;

        -- FULFILLMENT: Deduct from on-hand and reduce allocated
        UPDATE inventory_stock
        SET 
            quantity_on_hand = quantity_on_hand - v_fulfilled_qty,
            quantity_allocated = GREATEST(0, quantity_allocated - v_fulfilled_qty),
            quantity_available = GREATEST(0, 
                (quantity_on_hand - v_fulfilled_qty) - 
                GREATEST(0, quantity_allocated - v_fulfilled_qty)
            ),
            updated_at = CURRENT_TIMESTAMP
        WHERE product_id = v_order_line.product_id
          AND (product_variant_id = v_order_line.product_variant_id 
               OR (product_variant_id IS NULL AND v_order_line.product_variant_id IS NULL))
          AND store_id = NEW.store_id;

        IF NOT FOUND THEN
            RAISE WARNING 'No inventory_stock record found for product_id=%, product_variant_id=%, store_id=%. Movement recorded but stock not updated.',
                v_order_line.product_id, v_order_line.product_variant_id, NEW.store_id;
        END IF;

        -- Record the stock movement for auditing
        INSERT INTO stock_movements (
            movement_type,
            reference_type,
            reference_id,
            product_id,
            product_variant_id,
            from_store_id,
            quantity,
            uom_id,
            status,
            metadata
        )
        VALUES (
            'sale',
            'sales_order',
            NULL,
            v_order_line.product_id,
            v_order_line.product_variant_id,
            NEW.store_id,
            v_fulfilled_qty,
            v_order_line.uom_id,
            'completed',
            jsonb_build_object(
                'sales_order_id', NEW.id::TEXT,
                'sales_order_number', NEW.order_number,
                'order_line_id', v_order_line.id::TEXT,
                'order_status', NEW.order_status
            )
        );

        -- Mark active reservations as 'fulfilled' when order is fulfilled
        FOR v_reservation IN
            SELECT id, quantity_reserved
            FROM stock_reservations
            WHERE reference_type = 'sales_order'
              AND reference_id = NEW.id::TEXT
              AND product_id = v_order_line.product_id
              AND (product_variant_id = v_order_line.product_variant_id 
                   OR (product_variant_id IS NULL AND v_order_line.product_variant_id IS NULL))
              AND store_id = NEW.store_id
              AND status = 'active'
        LOOP
            UPDATE stock_reservations
            SET 
                status = 'fulfilled',
                updated_at = CURRENT_TIMESTAMP
            WHERE id = v_reservation.id;
        END LOOP;

    END LOOP;

    RETURN NEW;
END;
$$;
-- +goose StatementEnd
-- Create trigger "trg_deduct_inventory_on_fulfillment"
CREATE TRIGGER "trg_deduct_inventory_on_fulfillment" AFTER UPDATE ON "public"."sales_orders_v2" FOR EACH ROW WHEN ((old.order_status IS DISTINCT FROM new.order_status) AND (new.order_status = 'fulfilled'::public.order_status_v2)) EXECUTE FUNCTION "public"."fn_trigger_deduct_inventory_on_fulfillment"();
-- Create trigger "trg_sales_orders_v2_updated_at"
CREATE TRIGGER "trg_sales_orders_v2_updated_at" BEFORE UPDATE ON "public"."sales_orders_v2" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create "sales_returns" table
CREATE TABLE "public"."sales_returns" (
  "id" serial NOT NULL,
  "return_number" character varying(50) NOT NULL,
  "store_id" integer NOT NULL,
  "cashier_id" integer NULL,
  "cashier_session_id" integer NULL,
  "customer_id" integer NULL,
  "original_transaction_id" integer NULL,
  "return_date" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  "return_reason" character varying(255) NULL,
  "status" character varying(30) NULL DEFAULT 'pending',
  "subtotal" numeric(15,2) NULL DEFAULT 0,
  "tax_amount" numeric(15,2) NULL DEFAULT 0,
  "total_refund_amount" numeric(15,2) NULL DEFAULT 0,
  "refund_method" character varying(50) NULL,
  "refund_reference" character varying(100) NULL,
  "approved_by" integer NULL,
  "notes" text NULL,
  "metadata" jsonb NULL DEFAULT '{}',
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "sales_returns_return_number_key" UNIQUE ("return_number"),
  CONSTRAINT "sales_returns_approved_by_fkey" FOREIGN KEY ("approved_by") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "sales_returns_cashier_id_fkey" FOREIGN KEY ("cashier_id") REFERENCES "public"."cashiers" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "sales_returns_cashier_session_id_fkey" FOREIGN KEY ("cashier_session_id") REFERENCES "public"."cashier_sessions" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "sales_returns_customer_id_fkey" FOREIGN KEY ("customer_id") REFERENCES "public"."customers" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "sales_returns_original_transaction_id_fkey" FOREIGN KEY ("original_transaction_id") REFERENCES "public"."pos_transactions" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "sales_returns_store_id_fkey" FOREIGN KEY ("store_id") REFERENCES "public"."stores" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "sales_returns_status_check" CHECK ((status)::text = ANY ((ARRAY['pending'::character varying, 'approved'::character varying, 'completed'::character varying, 'cancelled'::character varying])::text[]))
);
-- Create index "idx_sales_returns_original_transaction" to table: "sales_returns"
CREATE INDEX "idx_sales_returns_original_transaction" ON "public"."sales_returns" ("original_transaction_id");
-- Create index "idx_sales_returns_status" to table: "sales_returns"
CREATE INDEX "idx_sales_returns_status" ON "public"."sales_returns" ("status");
-- Create index "idx_sales_returns_store_id" to table: "sales_returns"
CREATE INDEX "idx_sales_returns_store_id" ON "public"."sales_returns" ("store_id");
-- Create trigger "trg_sales_returns_updated_at"
CREATE TRIGGER "trg_sales_returns_updated_at" BEFORE UPDATE ON "public"."sales_returns" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create trigger "trg_stock_reservations_updated_at"
CREATE TRIGGER "trg_stock_reservations_updated_at" BEFORE UPDATE ON "public"."stock_reservations" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create trigger "trg_stores_updated_at"
CREATE TRIGGER "trg_stores_updated_at" BEFORE UPDATE ON "public"."stores" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create trigger "update_stores_updated_at"
CREATE TRIGGER "update_stores_updated_at" BEFORE UPDATE ON "public"."stores" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create trigger "trg_submenus_updated_at"
CREATE TRIGGER "trg_submenus_updated_at" BEFORE UPDATE ON "public"."submenus" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create trigger "update_submenus_updated_at"
CREATE TRIGGER "update_submenus_updated_at" BEFORE UPDATE ON "public"."submenus" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create trigger "trg_suppliers_updated_at"
CREATE TRIGGER "trg_suppliers_updated_at" BEFORE UPDATE ON "public"."suppliers" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create trigger "update_suppliers_updated_at"
CREATE TRIGGER "update_suppliers_updated_at" BEFORE UPDATE ON "public"."suppliers" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create "tenants" table
CREATE TABLE "public"."tenants" (
  "id" uuid NOT NULL DEFAULT public.uuid_generate_v4(),
  "tenant_name" character varying(255) NOT NULL,
  "slug" character varying(100) NOT NULL,
  "db_conn_str" text NOT NULL,
  "is_active" boolean NULL DEFAULT true,
  "settings" jsonb NULL DEFAULT '{}',
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "tenants_slug_key" UNIQUE ("slug")
);
-- Create index "idx_tenants_is_active" to table: "tenants"
CREATE INDEX "idx_tenants_is_active" ON "public"."tenants" ("is_active");
-- Create index "idx_tenants_slug" to table: "tenants"
CREATE INDEX "idx_tenants_slug" ON "public"."tenants" ("slug");
-- Create trigger "trg_tenants_updated_at"
CREATE TRIGGER "trg_tenants_updated_at" BEFORE UPDATE ON "public"."tenants" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create trigger "update_tenants_updated_at"
CREATE TRIGGER "update_tenants_updated_at" BEFORE UPDATE ON "public"."tenants" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create "product_uom_conversions" table
CREATE TABLE "public"."product_uom_conversions" (
  "id" serial NOT NULL,
  "product_id" integer NOT NULL,
  "from_uom_id" integer NOT NULL,
  "to_uom_id" integer NOT NULL,
  "conversion_factor" numeric(15,6) NOT NULL,
  "is_default" boolean NULL DEFAULT false,
  "metadata" jsonb NULL DEFAULT '{}',
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "product_uom_conversions_product_id_from_uom_id_to_uom_id_key" UNIQUE ("product_id", "from_uom_id", "to_uom_id"),
  CONSTRAINT "product_uom_conversions_from_uom_id_fkey" FOREIGN KEY ("from_uom_id") REFERENCES "public"."units_of_measure" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "product_uom_conversions_product_id_fkey" FOREIGN KEY ("product_id") REFERENCES "public"."products" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "product_uom_conversions_to_uom_id_fkey" FOREIGN KEY ("to_uom_id") REFERENCES "public"."units_of_measure" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create "fn_convert_uom_quantity" function
-- +goose StatementBegin
CREATE FUNCTION "public"."fn_convert_uom_quantity" ("p_product_id" integer, "p_from_uom_code" character varying, "p_quantity" numeric) RETURNS numeric LANGUAGE plpgsql AS $$
DECLARE
    v_base_uom_id INTEGER;
    v_from_uom_id INTEGER;
    v_base_quantity NUMERIC;
BEGIN
    -- Get base UOM for product
    SELECT base_uom_id INTO v_base_uom_id
    FROM products
    WHERE id = p_product_id;
    
    -- Get from UOM ID
    SELECT id INTO v_from_uom_id
    FROM units_of_measure
    WHERE code = p_from_uom_code;
    
    -- If from_uom is already base_uom, return as is
    IF v_from_uom_id = v_base_uom_id THEN
        RETURN p_quantity;
    END IF;
    
    -- Calculate conversion
    WITH RECURSIVE uom_path AS (
        -- Base case: direct conversion
        SELECT 
            from_uom_id,
            to_uom_id,
            conversion_factor::NUMERIC,
            1 as level
        FROM product_uom_conversions
        WHERE product_id = p_product_id
            AND from_uom_id = v_from_uom_id
        
        UNION ALL
        
        -- Recursive case: chain conversions
        SELECT 
            puc.from_uom_id,
            puc.to_uom_id,
            (up.conversion_factor * puc.conversion_factor)::NUMERIC,
            up.level + 1
        FROM product_uom_conversions puc
        JOIN uom_path up ON puc.from_uom_id = up.to_uom_id
        WHERE puc.product_id = p_product_id
            AND up.level < 10  -- Prevent infinite loops
    )
    SELECT p_quantity * conversion_factor INTO v_base_quantity
    FROM uom_path
    WHERE to_uom_id = v_base_uom_id
    ORDER BY level
    LIMIT 1;
    
    RETURN COALESCE(v_base_quantity, p_quantity);
END;
$$;
-- +goose StatementEnd
-- Create "transfer_requests" table
CREATE TABLE "public"."transfer_requests" (
  "id" serial NOT NULL,
  "organization_id" integer NOT NULL,
  "transfer_number" character varying(50) NOT NULL,
  "from_store_id" integer NOT NULL,
  "to_store_id" integer NOT NULL,
  "status" character varying(50) NOT NULL DEFAULT 'draft',
  "requested_by" integer NULL,
  "approved_by" integer NULL,
  "shipped_by" integer NULL,
  "received_by" integer NULL,
  "request_date" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  "expected_delivery_date" date NULL,
  "shipped_at" timestamp NULL,
  "received_at" timestamp NULL,
  "notes" text NULL,
  "metadata" jsonb NULL DEFAULT '{}',
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "transfer_requests_transfer_number_key" UNIQUE ("transfer_number"),
  CONSTRAINT "transfer_requests_approved_by_fkey" FOREIGN KEY ("approved_by") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "transfer_requests_from_store_id_fkey" FOREIGN KEY ("from_store_id") REFERENCES "public"."stores" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "transfer_requests_organization_id_fkey" FOREIGN KEY ("organization_id") REFERENCES "public"."organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "transfer_requests_received_by_fkey" FOREIGN KEY ("received_by") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "transfer_requests_requested_by_fkey" FOREIGN KEY ("requested_by") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "transfer_requests_shipped_by_fkey" FOREIGN KEY ("shipped_by") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "transfer_requests_to_store_id_fkey" FOREIGN KEY ("to_store_id") REFERENCES "public"."stores" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create "transfer_request_items" table
CREATE TABLE "public"."transfer_request_items" (
  "id" serial NOT NULL,
  "transfer_request_id" integer NOT NULL,
  "product_id" integer NOT NULL,
  "product_variant_id" integer NULL,
  "from_location_id" integer NULL,
  "to_location_id" integer NULL,
  "requested_quantity" numeric(15,3) NOT NULL,
  "approved_quantity" numeric(15,3) NULL DEFAULT 0,
  "shipped_quantity" numeric(15,3) NULL DEFAULT 0,
  "received_quantity" numeric(15,3) NULL DEFAULT 0,
  "uom_id" integer NULL,
  "batch_number" character varying(100) NULL,
  "notes" text NULL,
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "transfer_request_items_from_location_id_fkey" FOREIGN KEY ("from_location_id") REFERENCES "public"."storage_locations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "transfer_request_items_product_id_fkey" FOREIGN KEY ("product_id") REFERENCES "public"."products" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "transfer_request_items_product_variant_id_fkey" FOREIGN KEY ("product_variant_id") REFERENCES "public"."product_variants" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "transfer_request_items_to_location_id_fkey" FOREIGN KEY ("to_location_id") REFERENCES "public"."storage_locations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "transfer_request_items_transfer_request_id_fkey" FOREIGN KEY ("transfer_request_id") REFERENCES "public"."transfer_requests" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "transfer_request_items_uom_id_fkey" FOREIGN KEY ("uom_id") REFERENCES "public"."units_of_measure" ("id") ON UPDATE NO ACTION ON DELETE SET NULL
);
-- Create "fn_log_transfer_request_history" function
-- +goose StatementBegin
CREATE FUNCTION "public"."fn_log_transfer_request_history" () RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE
    v_history_entry JSONB;
    v_items_array JSONB := '[]'::jsonb;
    v_item RECORD;
    v_user_id INTEGER;
    v_user_name VARCHAR;
    v_from_store_name VARCHAR;
    v_to_store_name VARCHAR;
    
    -- Quantities
    v_qty_moved NUMERIC;
    v_base_qty_moved NUMERIC;
    v_base_uom_code VARCHAR;
    
    -- Stock tracking variables
    v_from_before_on_hand NUMERIC := 0;
    v_from_after_on_hand NUMERIC := 0;
    v_to_before_on_hand NUMERIC := 0;
    v_to_after_on_hand NUMERIC := 0;
    v_to_before_transit NUMERIC := 0;
    v_to_after_transit NUMERIC := 0;
BEGIN
    -- If it's an UPDATE, and status did not change, skip history logging
    IF TG_OP = 'UPDATE' AND OLD.status = NEW.status THEN
        RETURN NEW;
    END IF;

    -- Ensure metadata block is initialized as a JSON object (avoids "cannot set path in scalar")
    IF NEW.metadata IS NULL OR jsonb_typeof(NEW.metadata) != 'object' THEN
        NEW.metadata := '{}'::jsonb;
    END IF;

    -- Determine who changed the status based on current state
    IF NEW.status = 'draft' THEN v_user_id := NEW.requested_by;
    ELSIF NEW.status = 'approved' THEN v_user_id := NEW.approved_by;
    ELSIF NEW.status = 'shipped' THEN v_user_id := NEW.shipped_by;
    ELSIF NEW.status = 'received' THEN v_user_id := NEW.received_by;
    ELSE v_user_id := COALESCE(NEW.approved_by, NEW.requested_by);
    END IF;

    -- Fetch user name and store names for auditing
    SELECT username INTO v_user_name FROM users WHERE id = v_user_id;
    SELECT name INTO v_from_store_name FROM stores WHERE id = NEW.from_store_id;
    SELECT name INTO v_to_store_name FROM stores WHERE id = NEW.to_store_id;

    -- Loop through all products/items linked to this transfer request
    FOR v_item IN (
        SELECT tri.*, p.sku, p.name AS prod_name, u.code AS req_uom_code, p.base_uom_id
        FROM transfer_request_items tri
        JOIN products p ON tri.product_id = p.id
        LEFT JOIN units_of_measure u ON tri.uom_id = u.id
        WHERE tri.transfer_request_id = NEW.id
    ) LOOP
        -- Retrieve base UOM code
        SELECT code INTO v_base_uom_code FROM units_of_measure WHERE id = v_item.base_uom_id;

        -- Get current live stock at sending Store (Source)
        SELECT COALESCE(quantity_on_hand, 0)
        INTO v_from_after_on_hand
        FROM inventory_stock
        WHERE product_id = v_item.product_id 
          AND store_id = NEW.from_store_id 
          AND (product_variant_id = v_item.product_variant_id OR (product_variant_id IS NULL AND v_item.product_variant_id IS NULL));

        -- Get current live stock at receiving Store (Destination)
        SELECT COALESCE(quantity_on_hand, 0), COALESCE(quantity_in_transit, 0)
        INTO v_to_after_on_hand, v_to_after_transit
        FROM inventory_stock
        WHERE product_id = v_item.product_id 
          AND store_id = NEW.to_store_id 
          AND (product_variant_id = v_item.product_variant_id OR (product_variant_id IS NULL AND v_item.product_variant_id IS NULL));

        -- Resolve the quantity moved in the current state transition
        IF NEW.status = 'shipped' THEN
            v_qty_moved := COALESCE(v_item.shipped_quantity, v_item.requested_quantity);
        ELSIF NEW.status = 'received' THEN
            v_qty_moved := COALESCE(v_item.received_quantity, v_item.shipped_quantity);
        ELSE
            v_qty_moved := v_item.requested_quantity;
        END IF;

        -- Convert to base unit quantity for inventory calculations
        v_base_qty_moved := fn_convert_uom_quantity(v_item.product_id, COALESCE(v_item.req_uom_code, ''), v_qty_moved);
        IF v_base_qty_moved IS NULL THEN
            v_base_qty_moved := v_qty_moved;
        END IF;

        -- Calculate predicted stock outcomes based on transition states using base quantities
        IF NEW.status = 'shipped' THEN
            -- Shipped: Stock has been deducted from source, and added to transit at destination
            v_from_before_on_hand := v_from_after_on_hand + v_base_qty_moved;
            
            v_to_before_transit   := v_to_after_transit - v_base_qty_moved;
            v_to_before_on_hand   := v_to_after_on_hand;
        ELSIF NEW.status = 'received' THEN
            -- Received: Stock was moved from transit to on_hand at destination
            v_from_before_on_hand := v_from_after_on_hand;
            v_from_after_on_hand  := v_from_after_on_hand;
            
            v_to_before_transit   := v_to_after_transit + v_base_qty_moved;
            v_to_before_on_hand   := v_to_after_on_hand - v_base_qty_moved;
        ELSE
            -- Default/Approval state: Physical stock levels unchanged
            v_from_before_on_hand := v_from_after_on_hand;
            v_to_before_transit   := v_to_after_transit;
            v_to_before_on_hand   := v_to_after_on_hand;
        END IF;

        -- Append item audit blocks to array
        v_items_array := v_items_array || jsonb_build_array(
            jsonb_build_object(
                'product_id', v_item.product_id,
                'product_variant_id', v_item.product_variant_id,
                'sku', v_item.sku,
                'product_name', v_item.prod_name,
                'requested_quantity', v_item.requested_quantity,
                'shipped_quantity', v_item.shipped_quantity,
                'received_quantity', v_item.received_quantity,
                'uom', COALESCE(v_item.req_uom_code, ''),
                'converted_base_quantity', v_base_qty_moved,
                'base_uom', COALESCE(v_base_uom_code, ''),
                'inventory_snapshot', jsonb_build_object(
                    'source_store', jsonb_build_object(
                        'store_name', v_from_store_name,
                        'before_on_hand', COALESCE(v_from_before_on_hand, 0),
                        'after_on_hand', COALESCE(v_from_after_on_hand, 0),
                        'deducted', CASE WHEN NEW.status = 'shipped' THEN v_base_qty_moved ELSE 0 END
                    ),
                    'destination_store', jsonb_build_object(
                        'store_name', v_to_store_name,
                        'before_on_hand', COALESCE(v_to_before_on_hand, 0),
                        'after_on_hand', COALESCE(v_to_after_on_hand, 0),
                        'before_in_transit', COALESCE(v_to_before_transit, 0),
                        'after_in_transit', COALESCE(v_to_after_transit, 0),
                        'added_received', CASE WHEN NEW.status = 'received' THEN v_base_qty_moved ELSE 0 END
                    )
                )
            )
        );
    END LOOP;

    -- Build the history state entry
    v_history_entry := jsonb_build_object(
        'status', NEW.status,
        'changed_at', CURRENT_TIMESTAMP,
        'user_details', jsonb_build_object('id', v_user_id, 'username', COALESCE(v_user_name, 'system')),
        'notes', COALESCE(NEW.notes, ''),
        'transfer_items_snapshot', v_items_array
    );

    -- Build nested history array key
    IF NOT (NEW.metadata ? 'history') THEN
        NEW.metadata := jsonb_set(NEW.metadata, '{history}', '[]'::jsonb);
    END IF;

    -- Append the state audit to the history queue
    NEW.metadata := jsonb_set(
        NEW.metadata, 
        '{history}', 
        (NEW.metadata->'history') || jsonb_build_array(v_history_entry)
    );

    RETURN NEW;
END;
$$;
-- +goose StatementEnd
-- Create trigger "trg_transfer_request_history"
CREATE TRIGGER "trg_transfer_request_history" BEFORE INSERT OR UPDATE ON "public"."transfer_requests" FOR EACH ROW EXECUTE FUNCTION "public"."fn_log_transfer_request_history"();
-- Create trigger "update_transfer_requests_updated_at"
CREATE TRIGGER "update_transfer_requests_updated_at" BEFORE UPDATE ON "public"."transfer_requests" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create "ui_settings" table
CREATE TABLE "public"."ui_settings" (
  "id" serial NOT NULL,
  "submenu_id" integer NULL,
  "setting_key" character varying(100) NOT NULL,
  "setting_value" jsonb NOT NULL,
  "description" text NULL,
  "metadata" jsonb NULL DEFAULT '{}',
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "ui_settings_submenu_id_setting_key_key" UNIQUE ("submenu_id", "setting_key"),
  CONSTRAINT "ui_settings_submenu_id_fkey" FOREIGN KEY ("submenu_id") REFERENCES "public"."submenus" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create trigger "trg_ui_settings_updated_at"
CREATE TRIGGER "trg_ui_settings_updated_at" BEFORE UPDATE ON "public"."ui_settings" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create trigger "update_ui_settings_updated_at"
CREATE TRIGGER "update_ui_settings_updated_at" BEFORE UPDATE ON "public"."ui_settings" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create trigger "trg_users_updated_at"
CREATE TRIGGER "trg_users_updated_at" BEFORE UPDATE ON "public"."users" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create trigger "update_users_updated_at"
CREATE TRIGGER "update_users_updated_at" BEFORE UPDATE ON "public"."users" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();
-- Create "fn_approve_transfer_request" function
-- +goose StatementBegin
CREATE FUNCTION "public"."fn_approve_transfer_request" ("p_transfer_request_id" integer, "p_approved_by" integer) RETURNS TABLE ("success" boolean, "message" text) LANGUAGE plpgsql AS $$
DECLARE
    v_req RECORD;
BEGIN
    SELECT * INTO v_req FROM transfer_requests WHERE id = p_transfer_request_id FOR UPDATE;
    IF v_req IS NULL THEN
        RETURN QUERY SELECT false, 'Transfer request not found.';
        RETURN;
    END IF;

    IF v_req.status NOT IN ('draft', 'pending_approval') THEN
        RETURN QUERY SELECT false, 'Transfer request can only be approved from draft or pending_approval state.';
        RETURN;
    END IF;

    UPDATE transfer_requests
    SET status = 'approved',
        approved_by = p_approved_by,
        updated_at = CURRENT_TIMESTAMP
    WHERE id = p_transfer_request_id;

    RETURN QUERY SELECT true, 'Transfer request approved successfully.';
END;
$$;
-- +goose StatementEnd
-- Create "fn_calculate_loyalty_earned" function
-- +goose StatementBegin
CREATE FUNCTION "public"."fn_calculate_loyalty_earned" ("p_transaction_id" integer) RETURNS TABLE ("points_earned" numeric, "rule_applied" character varying, "customer_id" integer) LANGUAGE plpgsql AS $$
DECLARE
    v_txn        RECORD;
    v_rule       RECORD;
    v_points     DECIMAL(15,2) := 0;
BEGIN
    SELECT pt.*, pt.total_amount, pt.customer_id AS cust_id
    INTO v_txn
    FROM pos_transactions pt
    WHERE pt.id = p_transaction_id;

    IF NOT FOUND OR v_txn.cust_id IS NULL THEN
        RETURN QUERY SELECT 0::DECIMAL(15,2), 'No customer on transaction'::VARCHAR(255), NULL::INTEGER;
        RETURN;
    END IF;

    SELECT * INTO v_rule
    FROM loyalty_redemption_rules
    WHERE is_active = true
      AND (valid_from IS NULL OR valid_from <= CURRENT_DATE)
      AND (valid_to   IS NULL OR valid_to   >= CURRENT_DATE)
    ORDER BY id DESC LIMIT 1;

    IF NOT FOUND THEN
        RETURN QUERY SELECT 0::DECIMAL(15,2), 'No active loyalty rule found'::VARCHAR(255), v_txn.cust_id;
        RETURN;
    END IF;

    v_points := FLOOR(v_txn.total_amount * v_rule.points_earning_rate);

    -- Update customer loyalty_points balance
    UPDATE customers
    SET loyalty_points = loyalty_points + v_points,
        updated_at     = CURRENT_TIMESTAMP
    WHERE id = v_txn.cust_id;

    RETURN QUERY SELECT v_points, v_rule.rule_name::VARCHAR(255), v_txn.cust_id;
END;
$$;
-- +goose StatementEnd
-- Create "recipe_ingredients" table
CREATE TABLE "public"."recipe_ingredients" (
  "id" serial NOT NULL,
  "recipe_id" integer NOT NULL,
  "product_id" integer NOT NULL,
  "product_variant_id" integer NULL,
  "quantity" numeric(15,3) NOT NULL,
  "uom_id" integer NULL,
  "is_optional" boolean NULL DEFAULT false,
  "is_byproduct" boolean NULL DEFAULT false,
  "line_number" integer NULL,
  "metadata" jsonb NULL DEFAULT '{}',
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "recipe_ingredients_recipe_id_product_id_product_variant_id_key" UNIQUE ("recipe_id", "product_id", "product_variant_id"),
  CONSTRAINT "recipe_ingredients_product_id_fkey" FOREIGN KEY ("product_id") REFERENCES "public"."products" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "recipe_ingredients_product_variant_id_fkey" FOREIGN KEY ("product_variant_id") REFERENCES "public"."product_variants" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "recipe_ingredients_recipe_id_fkey" FOREIGN KEY ("recipe_id") REFERENCES "public"."recipes" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "recipe_ingredients_uom_id_fkey" FOREIGN KEY ("uom_id") REFERENCES "public"."units_of_measure" ("id") ON UPDATE NO ACTION ON DELETE SET NULL
);
-- Create index "idx_recipe_ingredients_product_id" to table: "recipe_ingredients"
CREATE INDEX "idx_recipe_ingredients_product_id" ON "public"."recipe_ingredients" ("product_id");
-- Create index "idx_recipe_ingredients_recipe_id" to table: "recipe_ingredients"
CREATE INDEX "idx_recipe_ingredients_recipe_id" ON "public"."recipe_ingredients" ("recipe_id");
-- Create index "idx_recipe_ingredients_variant_id" to table: "recipe_ingredients"
CREATE INDEX "idx_recipe_ingredients_variant_id" ON "public"."recipe_ingredients" ("product_variant_id");
-- Create "vw_recipe_bom" view
CREATE VIEW "public"."vw_recipe_bom" (
  "recipe_id",
  "recipe_code",
  "recipe_name",
  "yield_quantity",
  "organization_id",
  "ingredient_line_id",
  "line_number",
  "ingredient_qty",
  "is_optional",
  "is_byproduct",
  "product_id",
  "sku",
  "product_name",
  "variant_id",
  "variant_name",
  "uom_id",
  "uom_code",
  "uom_name",
  "unit_cost_estimate",
  "line_cost_estimate"
) AS SELECT r.id AS recipe_id,
    r.recipe_code,
    r.recipe_name,
    r.yield_quantity,
    r.organization_id,
    ri.id AS ingredient_line_id,
    ri.line_number,
    ri.quantity AS ingredient_qty,
    ri.is_optional,
    ri.is_byproduct,
    p.id AS product_id,
    p.sku,
    p.name AS product_name,
    pv.id AS variant_id,
    pv.variant_name,
    uom.id AS uom_id,
    uom.code AS uom_code,
    uom.name AS uom_name,
    pp.price AS unit_cost_estimate,
    round(ri.quantity * COALESCE(pp.price, 0::numeric), 4) AS line_cost_estimate
   FROM public.recipes r
     JOIN public.recipe_ingredients ri ON r.id = ri.recipe_id
     JOIN public.products p ON ri.product_id = p.id
     LEFT JOIN public.product_variants pv ON ri.product_variant_id = pv.id
     LEFT JOIN public.units_of_measure uom ON ri.uom_id = uom.id
     LEFT JOIN public.product_prices pp ON p.id = pp.product_id AND pp.price_list_id = (( SELECT price_lists.id
           FROM public.price_lists
          WHERE price_lists.code::text = 'RETAIL_SAR'::text AND price_lists.is_active = true
         LIMIT 1)) AND pp.is_active = true
  WHERE r.is_active = true;
-- Create "fn_calculate_recipe_cost" function
-- +goose StatementBegin
CREATE FUNCTION "public"."fn_calculate_recipe_cost" ("p_recipe_id" integer) RETURNS numeric LANGUAGE plpgsql AS $$
DECLARE
    v_total_cost NUMERIC := 0;
BEGIN
    SELECT COALESCE(SUM(vb.line_cost_estimate), 0)
      INTO v_total_cost
    FROM vw_recipe_bom vb
    WHERE vb.recipe_id    = p_recipe_id
      AND vb.is_byproduct = false
      AND vb.is_optional  = false;

    RETURN v_total_cost;
END;
$$;
-- +goose StatementEnd
-- Create "menu_item_modifiers" table
CREATE TABLE "public"."menu_item_modifiers" (
  "id" serial NOT NULL,
  "menu_item_id" integer NOT NULL,
  "modifier_name" character varying(100) NOT NULL,
  "modifier_type" character varying(30) NOT NULL DEFAULT 'addon',
  "price_adjustment" numeric(15,2) NULL DEFAULT 0,
  "is_active" boolean NULL DEFAULT true,
  "display_order" integer NULL DEFAULT 0,
  "metadata" jsonb NULL DEFAULT '{}',
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "menu_item_modifiers_menu_item_id_fkey" FOREIGN KEY ("menu_item_id") REFERENCES "public"."menu_items" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_menu_item_modifiers_is_active" to table: "menu_item_modifiers"
CREATE INDEX "idx_menu_item_modifiers_is_active" ON "public"."menu_item_modifiers" ("is_active");
-- Create index "idx_menu_item_modifiers_item_id" to table: "menu_item_modifiers"
CREATE INDEX "idx_menu_item_modifiers_item_id" ON "public"."menu_item_modifiers" ("menu_item_id");
-- Create "fn_get_item_modifiers" function
-- +goose StatementBegin
CREATE FUNCTION "public"."fn_get_item_modifiers" ("p_menu_item_id" integer) RETURNS TABLE ("modifier_id" integer, "modifier_name" character varying, "modifier_type" character varying, "price_adjustment" numeric, "display_order" integer) LANGUAGE plpgsql AS $$
BEGIN
    RETURN QUERY
    SELECT
        m.id,
        m.modifier_name::VARCHAR,
        m.modifier_type::VARCHAR,
        m.price_adjustment,
        m.display_order
    FROM menu_item_modifiers m
    WHERE m.menu_item_id = p_menu_item_id
      AND m.is_active    = true
    ORDER BY m.display_order;
END;
$$;
-- +goose StatementEnd
-- Create "fn_get_kds_orders" function
-- +goose StatementBegin
CREATE FUNCTION "public"."fn_get_kds_orders" ("p_store_id" integer, "p_statuses" character varying[] DEFAULT ARRAY['pending'::text, 'confirmed'::text, 'preparing'::text]) RETURNS TABLE ("order_id" integer, "order_number" character varying, "table_number" character varying, "waiter_name" character varying, "order_status" character varying, "ordered_at" timestamp, "minutes_elapsed" numeric, "item_id" integer, "item_name" character varying, "item_short_name" character varying, "item_qty" numeric, "item_notes" text, "item_modifiers" jsonb, "item_status" character varying) LANGUAGE plpgsql AS $$
BEGIN
    RETURN QUERY
    SELECT
        ro.id,
        ro.order_number::VARCHAR,
        rt.table_number::VARCHAR,
        (u.first_name || ' ' || u.last_name)::VARCHAR,
        ro.status::VARCHAR,
        ro.ordered_at,
        ROUND(EXTRACT(EPOCH FROM (CURRENT_TIMESTAMP - ro.ordered_at)) / 60.0, 1),
        roi.id,
        mi.name::VARCHAR,
        mi.short_name::VARCHAR,
        roi.quantity,
        roi.notes,
        roi.modifiers_snapshot,
        roi.status::VARCHAR
    FROM restaurant_orders ro
    LEFT JOIN restaurant_tables rt      ON ro.table_id = rt.id
    LEFT JOIN cashiers c                ON ro.cashier_id = c.id
    LEFT JOIN users u                   ON c.user_id = u.id
    JOIN  restaurant_order_items roi    ON ro.id = roi.order_id
    JOIN  menu_items mi                 ON roi.menu_item_id = mi.id
    WHERE ro.store_id = p_store_id
      AND ro.status = ANY(p_statuses)
    ORDER BY ro.ordered_at, roi.line_number;
END;
$$;
-- +goose StatementEnd
-- Create "vw_restaurant_menu" view
CREATE VIEW "public"."vw_restaurant_menu" (
  "menu_item_id",
  "store_id",
  "item_name",
  "short_name",
  "description",
  "image_url",
  "base_price",
  "cost_price",
  "preparation_time_min",
  "is_available",
  "is_active",
  "display_order",
  "item_metadata",
  "category_id",
  "category_name",
  "category_code",
  "parent_category_id",
  "category_display_order",
  "category_image_url",
  "parent_category_name",
  "tax_category_id",
  "tax_rate",
  "tax_is_inclusive",
  "recipe_id",
  "recipe_name",
  "recipe_yield",
  "product_id",
  "product_sku",
  "active_modifier_count",
  "margin_percent"
) AS SELECT mi.id AS menu_item_id,
    mi.store_id,
    mi.name AS item_name,
    mi.short_name,
    mi.description,
    mi.image_url,
    mi.base_price,
    mi.cost_price,
    mi.preparation_time_min,
    mi.is_available,
    mi.is_active,
    mi.display_order,
    mi.metadata AS item_metadata,
    mc.id AS category_id,
    mc.name AS category_name,
    mc.code AS category_code,
    mc.parent_category_id,
    mc.display_order AS category_display_order,
    mc.image_url AS category_image_url,
    mc_parent.name AS parent_category_name,
    tc.id AS tax_category_id,
    tc.tax_rate,
    tc.is_inclusive AS tax_is_inclusive,
    mi.recipe_id,
    r.recipe_name,
    r.yield_quantity AS recipe_yield,
    mi.product_id,
    p.sku AS product_sku,
    (( SELECT count(*) AS count
           FROM public.menu_item_modifiers m
          WHERE m.menu_item_id = mi.id AND m.is_active = true))::integer AS active_modifier_count,
        CASE
            WHEN mi.base_price > 0::numeric AND mi.cost_price > 0::numeric THEN round((mi.base_price - mi.cost_price) / mi.base_price * 100::numeric, 2)
            ELSE NULL::numeric
        END AS margin_percent
   FROM public.menu_items mi
     JOIN public.menu_categories mc ON mi.menu_category_id = mc.id
     LEFT JOIN public.menu_categories mc_parent ON mc.parent_category_id = mc_parent.id
     LEFT JOIN public.tax_categories tc ON mi.tax_category_id = tc.id
     LEFT JOIN public.recipes r ON mi.recipe_id = r.id
     LEFT JOIN public.products p ON mi.product_id = p.id
  WHERE mi.is_active = true;
-- Create "fn_get_restaurant_menu" function
-- +goose StatementBegin
CREATE FUNCTION "public"."fn_get_restaurant_menu" ("p_store_id" integer, "p_category_id" integer DEFAULT NULL::integer, "p_include_unavail" boolean DEFAULT false) RETURNS TABLE ("menu_item_id" integer, "item_name" character varying, "short_name" character varying, "description" text, "image_url" text, "base_price" numeric, "preparation_time_min" integer, "is_available" boolean, "category_id" integer, "category_name" character varying, "parent_category_name" character varying, "tax_rate" numeric, "tax_is_inclusive" boolean, "recipe_id" integer, "product_id" integer, "active_modifier_count" integer, "margin_percent" numeric) LANGUAGE plpgsql AS $$
BEGIN
    RETURN QUERY
    SELECT
        vm.menu_item_id,
        vm.item_name::VARCHAR,
        vm.short_name::VARCHAR,
        vm.description,
        vm.image_url,
        vm.base_price,
        vm.preparation_time_min,
        vm.is_available,
        vm.category_id,
        vm.category_name::VARCHAR,
        vm.parent_category_name::VARCHAR,
        vm.tax_rate,
        vm.tax_is_inclusive,
        vm.recipe_id,
        vm.product_id,
        vm.active_modifier_count,
        vm.margin_percent
    FROM vw_restaurant_menu vm
    WHERE vm.store_id = p_store_id
      AND (p_category_id IS NULL OR vm.category_id = p_category_id)
      AND (p_include_unavail = true OR vm.is_available = true)
    ORDER BY vm.category_display_order, vm.display_order;
END;
$$;
-- +goose StatementEnd
-- Create "waste_logs" table
CREATE TABLE "public"."waste_logs" (
  "id" serial NOT NULL,
  "store_id" integer NOT NULL,
  "product_id" integer NULL,
  "menu_item_id" integer NULL,
  "recipe_id" integer NULL,
  "waste_source" character varying(30) NOT NULL DEFAULT 'kitchen',
  "quantity" numeric(15,3) NOT NULL,
  "uom_id" integer NULL,
  "unit_cost" numeric(15,4) NULL DEFAULT 0,
  "total_cost" numeric(15,2) NULL DEFAULT 0,
  "reason" text NULL,
  "logged_by" integer NULL,
  "order_id" integer NULL,
  "wasted_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  "metadata" jsonb NULL DEFAULT '{}',
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "waste_logs_logged_by_fkey" FOREIGN KEY ("logged_by") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "waste_logs_menu_item_id_fkey" FOREIGN KEY ("menu_item_id") REFERENCES "public"."menu_items" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "waste_logs_order_id_fkey" FOREIGN KEY ("order_id") REFERENCES "public"."restaurant_orders" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "waste_logs_product_id_fkey" FOREIGN KEY ("product_id") REFERENCES "public"."products" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "waste_logs_recipe_id_fkey" FOREIGN KEY ("recipe_id") REFERENCES "public"."recipes" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "waste_logs_store_id_fkey" FOREIGN KEY ("store_id") REFERENCES "public"."stores" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "waste_logs_uom_id_fkey" FOREIGN KEY ("uom_id") REFERENCES "public"."units_of_measure" ("id") ON UPDATE NO ACTION ON DELETE SET NULL
);
-- Create index "idx_waste_logs_menu_item_id" to table: "waste_logs"
CREATE INDEX "idx_waste_logs_menu_item_id" ON "public"."waste_logs" ("menu_item_id");
-- Create index "idx_waste_logs_order_id" to table: "waste_logs"
CREATE INDEX "idx_waste_logs_order_id" ON "public"."waste_logs" ("order_id");
-- Create index "idx_waste_logs_product_id" to table: "waste_logs"
CREATE INDEX "idx_waste_logs_product_id" ON "public"."waste_logs" ("product_id");
-- Create index "idx_waste_logs_recipe_id" to table: "waste_logs"
CREATE INDEX "idx_waste_logs_recipe_id" ON "public"."waste_logs" ("recipe_id");
-- Create index "idx_waste_logs_store_id" to table: "waste_logs"
CREATE INDEX "idx_waste_logs_store_id" ON "public"."waste_logs" ("store_id");
-- Create index "idx_waste_logs_store_source_date" to table: "waste_logs"
CREATE INDEX "idx_waste_logs_store_source_date" ON "public"."waste_logs" ("store_id", "waste_source", "wasted_at");
-- Create index "idx_waste_logs_waste_source" to table: "waste_logs"
CREATE INDEX "idx_waste_logs_waste_source" ON "public"."waste_logs" ("waste_source");
-- Create index "idx_waste_logs_wasted_at" to table: "waste_logs"
CREATE INDEX "idx_waste_logs_wasted_at" ON "public"."waste_logs" ("wasted_at");
-- Create "fn_get_waste_report" function
-- +goose StatementBegin
CREATE FUNCTION "public"."fn_get_waste_report" ("p_store_id" integer, "p_from_date" date, "p_to_date" date, "p_waste_source" character varying DEFAULT NULL::character varying) RETURNS TABLE ("waste_date" date, "waste_source" character varying, "product_id" integer, "product_name" character varying, "menu_item_id" integer, "menu_item_name" character varying, "quantity" numeric, "uom_code" character varying, "total_cost" numeric, "reason" text, "logged_by_name" character varying) LANGUAGE plpgsql AS $$
BEGIN
    RETURN QUERY
    SELECT
        DATE(wl.wasted_at),
        wl.waste_source::VARCHAR,
        wl.product_id,
        p.name::VARCHAR,
        wl.menu_item_id,
        mi.name::VARCHAR,
        wl.quantity,
        uom.code::VARCHAR,
        wl.total_cost,
        wl.reason,
        (u.first_name || ' ' || u.last_name)::VARCHAR
    FROM waste_logs wl
    LEFT JOIN products p            ON wl.product_id   = p.id
    LEFT JOIN menu_items mi         ON wl.menu_item_id = mi.id
    LEFT JOIN units_of_measure uom  ON wl.uom_id       = uom.id
    LEFT JOIN users u               ON wl.logged_by    = u.id
    WHERE wl.store_id           = p_store_id
      AND DATE(wl.wasted_at)    BETWEEN p_from_date AND p_to_date
      AND (p_waste_source IS NULL OR wl.waste_source = p_waste_source)
    ORDER BY wl.wasted_at DESC;
END;
$$;
-- +goose StatementEnd
-- Create "product_barcodes" table
CREATE TABLE "public"."product_barcodes" (
  "id" serial NOT NULL,
  "product_id" integer NOT NULL,
  "product_variant_id" integer NULL,
  "barcode" character varying(100) NOT NULL,
  "barcode_type" character varying(50) NULL,
  "is_primary" boolean NULL DEFAULT false,
  "metadata" jsonb NULL DEFAULT '{}',
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "product_barcodes_barcode_key" UNIQUE ("barcode"),
  CONSTRAINT "product_barcodes_product_id_fkey" FOREIGN KEY ("product_id") REFERENCES "public"."products" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "product_barcodes_product_variant_id_fkey" FOREIGN KEY ("product_variant_id") REFERENCES "public"."product_variants" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_product_barcodes_barcode" to table: "product_barcodes"
CREATE INDEX "idx_product_barcodes_barcode" ON "public"."product_barcodes" ("barcode");
-- Create index "idx_product_barcodes_barcode_lookup" to table: "product_barcodes"
CREATE INDEX "idx_product_barcodes_barcode_lookup" ON "public"."product_barcodes" ("barcode") WHERE (is_primary = true);
-- Create index "idx_product_barcodes_product_id" to table: "product_barcodes"
CREATE INDEX "idx_product_barcodes_product_id" ON "public"."product_barcodes" ("product_id");
-- Create index "idx_product_barcodes_product_variant_id" to table: "product_barcodes"
CREATE INDEX "idx_product_barcodes_product_variant_id" ON "public"."product_barcodes" ("product_variant_id");
-- Create "vw_pos_product_catalog" view
CREATE VIEW "public"."vw_pos_product_catalog" (
  "product_id",
  "product_variant_id",
  "base_sku",
  "sku",
  "base_product_name",
  "product_name",
  "variant_attributes",
  "description",
  "product_type",
  "category_id",
  "category_name",
  "category_code",
  "parent_category_id",
  "parent_category_name",
  "brand_id",
  "brand_name",
  "uom_id",
  "uom_code",
  "uom_name",
  "decimal_places",
  "barcode",
  "barcode_type",
  "tax_category_id",
  "tax_category_name",
  "tax_rate",
  "tax_is_inclusive",
  "retail_price",
  "retail_price_id",
  "promo_price",
  "promo_price_id",
  "promo_min_quantity",
  "promo_valid_from",
  "promo_valid_to",
  "promotion_name",
  "discount_percent",
  "effective_price",
  "has_active_promotion",
  "is_active",
  "is_sellable",
  "is_serialized",
  "is_batch_managed",
  "allow_decimal_quantity",
  "track_inventory",
  "product_metadata"
) AS SELECT p.id AS product_id,
    pv.id AS product_variant_id,
    p.sku AS base_sku,
    COALESCE(pv.variant_sku, p.sku) AS sku,
    p.name AS base_product_name,
    COALESCE(pv.variant_name, p.name) AS product_name,
    pv.variant_attributes,
    p.description,
    p.product_type,
    pc.id AS category_id,
    pc.name AS category_name,
    pc.code AS category_code,
    pc_parent.id AS parent_category_id,
    pc_parent.name AS parent_category_name,
    b.id AS brand_id,
    b.name AS brand_name,
    uom.id AS uom_id,
    uom.code AS uom_code,
    uom.name AS uom_name,
    uom.decimal_places,
    COALESCE(pb_variant.barcode, pb_base.barcode) AS barcode,
    COALESCE(pb_variant.barcode_type, pb_base.barcode_type) AS barcode_type,
    tc.id AS tax_category_id,
    tc.name AS tax_category_name,
    tc.tax_rate,
    tc.is_inclusive AS tax_is_inclusive,
    COALESCE(pp_retail_v.price, pp_retail.price) AS retail_price,
    COALESCE(pp_retail_v.id, pp_retail.id) AS retail_price_id,
    COALESCE(pp_promo_v.price, pp_promo.price) AS promo_price,
    COALESCE(pp_promo_v.id, pp_promo.id) AS promo_price_id,
    COALESCE(pp_promo_v.min_quantity, pp_promo.min_quantity) AS promo_min_quantity,
    COALESCE(pp_promo_v.valid_from, pp_promo.valid_from) AS promo_valid_from,
    COALESCE(pp_promo_v.valid_to, pp_promo.valid_to) AS promo_valid_to,
    COALESCE(pp_promo_v.metadata ->> 'promotion_name'::text, pp_promo.metadata ->> 'promotion_name'::text) AS promotion_name,
    COALESCE(pp_promo_v.metadata ->> 'discount_percent'::text, pp_promo.metadata ->> 'discount_percent'::text) AS discount_percent,
        CASE
            WHEN COALESCE(pp_promo_v.id, pp_promo.id) IS NOT NULL AND COALESCE(pp_promo_v.is_active, pp_promo.is_active) = true AND COALESCE(pp_promo_v.valid_from, pp_promo.valid_from) <= CURRENT_DATE AND (COALESCE(pp_promo_v.valid_to, pp_promo.valid_to) IS NULL OR COALESCE(pp_promo_v.valid_to, pp_promo.valid_to) >= CURRENT_DATE) THEN COALESCE(pp_promo_v.price, pp_promo.price)
            ELSE COALESCE(pp_retail_v.price, pp_retail.price)
        END AS effective_price,
        CASE
            WHEN COALESCE(pp_promo_v.id, pp_promo.id) IS NOT NULL AND COALESCE(pp_promo_v.is_active, pp_promo.is_active) = true AND COALESCE(pp_promo_v.valid_from, pp_promo.valid_from) <= CURRENT_DATE AND (COALESCE(pp_promo_v.valid_to, pp_promo.valid_to) IS NULL OR COALESCE(pp_promo_v.valid_to, pp_promo.valid_to) >= CURRENT_DATE) THEN true
            ELSE false
        END AS has_active_promotion,
    p.is_active,
    p.is_sellable,
    p.is_serialized,
    p.is_batch_managed,
    p.allow_decimal_quantity,
    p.track_inventory,
    p.metadata AS product_metadata
   FROM public.products p
     LEFT JOIN public.product_variants pv ON pv.product_id = p.id AND pv.is_active = true
     LEFT JOIN public.product_categories pc ON p.category_id = pc.id
     LEFT JOIN public.product_categories pc_parent ON pc.parent_category_id = pc_parent.id
     LEFT JOIN public.brands b ON p.brand_id = b.id
     LEFT JOIN public.units_of_measure uom ON p.base_uom_id = uom.id
     LEFT JOIN public.product_barcodes pb_base ON p.id = pb_base.product_id AND pb_base.product_variant_id IS NULL AND pb_base.is_primary = true
     LEFT JOIN public.product_barcodes pb_variant ON pv.id = pb_variant.product_variant_id AND pb_variant.is_primary = true
     LEFT JOIN public.tax_categories tc ON p.tax_category_id = tc.id
     LEFT JOIN public.product_prices pp_retail ON p.id = pp_retail.product_id AND pp_retail.product_variant_id IS NULL AND pp_retail.price_list_id = (( SELECT price_lists.id
           FROM public.price_lists
          WHERE price_lists.code::text = 'RETAIL_SAR'::text AND price_lists.is_active = true
         LIMIT 1)) AND pp_retail.is_active = true
     LEFT JOIN public.product_prices pp_retail_v ON p.id = pp_retail_v.product_id AND pp_retail_v.product_variant_id = pv.id AND pp_retail_v.price_list_id = (( SELECT price_lists.id
           FROM public.price_lists
          WHERE price_lists.code::text = 'RETAIL_SAR'::text AND price_lists.is_active = true
         LIMIT 1)) AND pp_retail_v.is_active = true
     LEFT JOIN public.product_prices pp_promo ON p.id = pp_promo.product_id AND pp_promo.product_variant_id IS NULL AND pp_promo.price_list_id = (( SELECT price_lists.id
           FROM public.price_lists
          WHERE price_lists.code::text = 'PROMO_SAR'::text AND price_lists.is_active = true
         LIMIT 1)) AND pp_promo.is_active = true
     LEFT JOIN public.product_prices pp_promo_v ON p.id = pp_promo_v.product_id AND pp_promo_v.product_variant_id = pv.id AND pp_promo_v.price_list_id = (( SELECT price_lists.id
           FROM public.price_lists
          WHERE price_lists.code::text = 'PROMO_SAR'::text AND price_lists.is_active = true
         LIMIT 1)) AND pp_promo_v.is_active = true
  WHERE p.is_active = true AND p.is_sellable = true
  ORDER BY pc.name, p.name, pv.variant_name;
-- Create "fn_pos_get_product_by_barcode" function
-- +goose StatementBegin
CREATE FUNCTION "public"."fn_pos_get_product_by_barcode" ("p_barcode" character varying, "p_store_id" integer) RETURNS TABLE ("product_id" integer, "sku" character varying, "product_name" character varying, "description" text, "category_name" character varying, "brand_name" character varying, "barcode" character varying, "uom_code" character varying, "decimal_places" integer, "retail_price" numeric, "promo_price" numeric, "effective_price" numeric, "has_promotion" boolean, "promotion_name" character varying, "promo_min_quantity" numeric, "tax_rate" numeric, "tax_is_inclusive" boolean, "quantity_available" numeric, "is_in_stock" boolean, "allow_decimal_quantity" boolean, "is_serialized" boolean, "is_batch_managed" boolean, "product_metadata" jsonb, "package_n_price" jsonb, "product_uom_conversions" jsonb) LANGUAGE plpgsql AS $$
BEGIN
    RETURN QUERY
    SELECT 
        cat.product_id,
        cat.sku::VARCHAR,
        cat.product_name::VARCHAR,
        cat.description,
        cat.category_name::VARCHAR,
        cat.brand_name::VARCHAR,
        cat.barcode::VARCHAR,
        cat.uom_code::VARCHAR,
        (cat.decimal_places)::INTEGER,
        cat.retail_price,
        COALESCE(cat.promo_price, promo_rule.calculated_promo_price) AS promo_price,
        COALESCE(cat.promo_price, promo_rule.calculated_promo_price, cat.retail_price) AS effective_price,
        (cat.has_active_promotion OR (promo_rule.promo_name IS NOT NULL)) AS has_promotion,
        COALESCE(cat.promotion_name, promo_rule.promo_name)::VARCHAR AS promotion_name,
        COALESCE(cat.promo_min_quantity, promo_rule.promo_min_qty) AS promo_min_quantity,
        cat.tax_rate,
        cat.tax_is_inclusive,
        COALESCE(inv.quantity_available, 0)::NUMERIC,
        (COALESCE(inv.quantity_available, 0) > 0),
        cat.allow_decimal_quantity,
        cat.is_serialized,
        cat.is_batch_managed,
        cat.product_metadata,
        (SELECT COALESCE(jsonb_agg(s.rec ORDER BY s.pl_code, s.uom_code), '[]'::jsonb)
         FROM (
             SELECT 
                 pl.code AS pl_code,
                 uom.code AS uom_code,
                 jsonb_build_object(
                     'product_name', p.name,
                     'price_list_id', pl.id,
                     'price_list_code', pl.code,
                     'price_list_name', pl.name,
                     'price_list', pl.name,
                     'price_list_type', pl.price_list_type,
                     'currency_code', pl.currency_code,
                     'uom_id', uom.id,
                     'uom_code', uom.code,
                     'uom', uom.code,
                     'uom_name', uom.name,
                     'decimal_places', uom.decimal_places,
                     'price', pp.price,
                     'min_quantity', pp.min_quantity,
                     'max_quantity', pp.max_quantity,
                     'valid_from', pp.valid_from,
                     'valid_to', pp.valid_to,
                     'metadata', COALESCE(pp.metadata, '{}'::jsonb),
                     'barcodes', (SELECT COALESCE(jsonb_agg(pb.barcode), '[]'::jsonb) FROM product_barcodes pb WHERE pb.product_id = pp.product_id)
                 ) AS rec
             FROM product_prices pp
             INNER JOIN products p ON pp.product_id = p.id
             INNER JOIN price_lists pl ON pp.price_list_id = pl.id AND pl.is_active = true
             LEFT JOIN units_of_measure uom ON pp.uom_id = uom.id
             WHERE pp.product_id = cat.product_id
               AND pp.is_active = true
               AND (pp.valid_from IS NULL OR pp.valid_from <= CURRENT_DATE)
               AND (pp.valid_to IS NULL OR pp.valid_to >= CURRENT_DATE)
         ) AS s),
        (SELECT COALESCE(jsonb_agg(conv.cv ORDER BY conv.fu_code, conv.tu_code), '[]'::jsonb)
         FROM (
             SELECT fu.code AS fu_code, tu.code AS tu_code,
                    jsonb_build_object(
                        'from_uom_id', fu.id, 'from_uom_code', fu.code, 'from_uom_name', fu.name,
                        'to_uom_id', tu.id, 'to_uom_code', tu.code, 'to_uom_name', tu.name,
                        'conversion_factor', puc.conversion_factor
                    ) AS cv
             FROM product_uom_conversions puc
             JOIN units_of_measure fu ON puc.from_uom_id = fu.id
             JOIN units_of_measure tu ON puc.to_uom_id = tu.id
             WHERE puc.product_id = cat.product_id
         ) AS conv)
    FROM vw_pos_product_catalog cat
    LEFT JOIN LATERAL (
        SELECT 
            pr.name AS promo_name,
            pr.min_quantity AS promo_min_qty,
            pr.discount_value,
            pr.promotion_type,
            CASE 
                WHEN pr.promotion_type = 'percentage_discount' AND cat.retail_price IS NOT NULL THEN
                    ROUND(cat.retail_price * (1.0 - (pr.discount_value / 100.0)), 2)
                WHEN pr.promotion_type = 'fixed_discount' AND cat.retail_price IS NOT NULL THEN
                    GREATEST(0.00, cat.retail_price - pr.discount_value)
                ELSE cat.retail_price
            END AS calculated_promo_price
        FROM promotions pr
        WHERE pr.is_active = true
          AND (pr.valid_from IS NULL OR pr.valid_from <= CURRENT_TIMESTAMP)
          AND (pr.valid_to IS NULL OR pr.valid_to >= CURRENT_TIMESTAMP)
          AND (cardinality(pr.store_ids) = 0 OR p_store_id = ANY(pr.store_ids))
          AND (
              pr.applies_to = 'all'
              OR (pr.applies_to = 'product' AND cat.product_id = ANY(pr.target_product_ids))
              OR (pr.applies_to = 'category' AND cat.category_id = ANY(pr.target_category_ids))
          )
        ORDER BY pr.created_at DESC
        LIMIT 1
    ) promo_rule ON true
    LEFT JOIN inventory_stock inv ON cat.product_id = inv.product_id AND inv.store_id = p_store_id
    WHERE cat.barcode = p_barcode
    LIMIT 1;
END;
$$;
-- +goose StatementEnd
-- Create "fn_pos_get_products_by_category" function
-- +goose StatementBegin
CREATE FUNCTION "public"."fn_pos_get_products_by_category" ("p_category_id" integer, "p_store_id" integer, "p_include_subcategories" boolean DEFAULT true) RETURNS TABLE ("product_id" integer, "sku" character varying, "product_name" character varying, "category_name" character varying, "brand_name" character varying, "barcode" character varying, "effective_price" numeric, "has_promotion" boolean, "promotion_name" character varying, "quantity_available" numeric, "is_in_stock" boolean, "package_n_price" jsonb, "product_uom_conversions" jsonb) LANGUAGE plpgsql AS $$
BEGIN
    RETURN QUERY
    SELECT 
        cat.product_id,
        cat.sku::VARCHAR,
        cat.product_name::VARCHAR,
        cat.category_name::VARCHAR,
        cat.brand_name::VARCHAR,
        cat.barcode::VARCHAR,
        cat.effective_price,
        cat.has_active_promotion,
        cat.promotion_name::VARCHAR,
        COALESCE(inv.quantity_available, 0)::NUMERIC,
        (COALESCE(inv.quantity_available, 0) > 0),
        (SELECT COALESCE(jsonb_agg(s.rec ORDER BY s.pl_code, s.uom_code), '[]'::jsonb)
         FROM (
             SELECT 
                 pl.code AS pl_code,
                 uom.code AS uom_code,
                 jsonb_build_object(
                     'product_name', p.name,
                     'price_list_id', pl.id,
                     'price_list_code', pl.code,
                     'price_list_name', pl.name,
                     'price_list', pl.name,
                     'price_list_type', pl.price_list_type,
                     'currency_code', pl.currency_code,
                     'uom_id', uom.id,
                     'uom_code', uom.code,
                     'uom', uom.code,
                     'uom_name', uom.name,
                     'decimal_places', uom.decimal_places,
                     'price', pp.price,
                     'min_quantity', pp.min_quantity,
                     'max_quantity', pp.max_quantity,
                     'valid_from', pp.valid_from,
                     'valid_to', pp.valid_to,
                     'metadata', COALESCE(pp.metadata, '{}'::jsonb),
                     'barcodes', (SELECT COALESCE(jsonb_agg(pb.barcode), '[]'::jsonb) FROM product_barcodes pb WHERE pb.product_id = pp.product_id)
                 ) AS rec
             FROM product_prices pp
             INNER JOIN products p ON pp.product_id = p.id
             INNER JOIN price_lists pl ON pp.price_list_id = pl.id AND pl.is_active = true
             LEFT JOIN units_of_measure uom ON pp.uom_id = uom.id
             WHERE pp.product_id = cat.product_id
               AND pp.is_active = true
               AND (pp.valid_from IS NULL OR pp.valid_from <= CURRENT_DATE)
               AND (pp.valid_to IS NULL OR pp.valid_to >= CURRENT_DATE)
         ) AS s),
        (SELECT COALESCE(jsonb_agg(conv.cv ORDER BY conv.fu_code, conv.tu_code), '[]'::jsonb)
         FROM (
             SELECT fu.code AS fu_code, tu.code AS tu_code,
                    jsonb_build_object(
                        'from_uom_id', fu.id, 'from_uom_code', fu.code, 'from_uom_name', fu.name,
                        'to_uom_id', tu.id, 'to_uom_code', tu.code, 'to_uom_name', tu.name,
                        'conversion_factor', puc.conversion_factor
                    ) AS cv
             FROM product_uom_conversions puc
             JOIN units_of_measure fu ON puc.from_uom_id = fu.id
             JOIN units_of_measure tu ON puc.to_uom_id = tu.id
             WHERE puc.product_id = cat.product_id
         ) AS conv)
    FROM vw_pos_product_catalog cat
    LEFT JOIN inventory_stock inv ON cat.product_id = inv.product_id AND inv.store_id = p_store_id
    WHERE 
        (cat.category_id = p_category_id 
         OR (p_include_subcategories = true AND cat.parent_category_id = p_category_id))
        AND COALESCE(inv.quantity_available, 0) > 0
    ORDER BY cat.product_name;
END;
$$;
-- +goose StatementEnd
-- Create "fn_pos_get_products_with_stock" function
-- +goose StatementBegin
CREATE FUNCTION "public"."fn_pos_get_products_with_stock" ("p_store_id" integer, "p_category_id" integer DEFAULT NULL::integer, "p_search_term" character varying DEFAULT NULL::character varying, "p_include_out_of_stock" boolean DEFAULT false) RETURNS TABLE ("product_id" integer, "product_variant_id" integer, "sku" character varying, "product_name" character varying, "variant_attributes" jsonb, "description" text, "category_id" integer, "category_name" character varying, "brand_name" character varying, "barcode" character varying, "uom_code" character varying, "decimal_places" integer, "retail_price" numeric, "promo_price" numeric, "effective_price" numeric, "has_promotion" boolean, "promotion_name" character varying, "discount_percent" character varying, "promo_min_quantity" numeric, "tax_rate" numeric, "tax_is_inclusive" boolean, "quantity_available" numeric, "quantity_on_hand" numeric, "quantity_allocated" numeric, "is_in_stock" boolean, "is_low_stock" boolean, "reorder_level" numeric, "allow_decimal_quantity" boolean, "is_serialized" boolean, "is_batch_managed" boolean, "product_metadata" jsonb, "product_variants" jsonb, "package_n_price" jsonb, "product_uom_conversions" jsonb) LANGUAGE plpgsql AS $$
BEGIN
    RETURN QUERY
    SELECT
        cat.product_id,
        cat.product_variant_id,
        cat.sku::VARCHAR,
        cat.product_name::VARCHAR,
        cat.variant_attributes,
        cat.description,
        cat.category_id,
        cat.category_name::VARCHAR,
        cat.brand_name::VARCHAR,
        cat.barcode::VARCHAR,
        cat.uom_code::VARCHAR,
        cat.decimal_places::INTEGER,
        cat.retail_price,
        COALESCE(cat.promo_price, promo_rule.calculated_promo_price) AS promo_price,
        COALESCE(cat.promo_price, promo_rule.calculated_promo_price, cat.retail_price) AS effective_price,
        (cat.has_active_promotion OR (promo_rule.promo_name IS NOT NULL)) AS has_promotion,
        COALESCE(cat.promotion_name, promo_rule.promo_name)::VARCHAR AS promotion_name,
        COALESCE(cat.discount_percent, promo_rule.calc_discount_percent)::VARCHAR AS discount_percent,
        COALESCE(cat.promo_min_quantity, promo_rule.promo_min_qty) AS promo_min_quantity,
        cat.tax_rate,
        cat.tax_is_inclusive,
        -- FIX 9.3: Join on both product_id AND product_variant_id
        COALESCE(inv.quantity_available, 0)::NUMERIC,
        COALESCE(inv.quantity_on_hand,   0)::NUMERIC,
        COALESCE(inv.quantity_allocated, 0)::NUMERIC,
        (COALESCE(inv.quantity_available, 0) > 0),
        (COALESCE(inv.quantity_available, 0) <= COALESCE(inv.reorder_level, 0) AND COALESCE(inv.quantity_available, 0) > 0),
        COALESCE(inv.reorder_level, 0)::NUMERIC,
        cat.allow_decimal_quantity,
        cat.is_serialized,
        cat.is_batch_managed,
        cat.product_metadata,
        (SELECT COALESCE(jsonb_agg(
            jsonb_build_object(
                'id', pv.id,
                'product_id', pv.product_id,
                'variant_sku', pv.variant_sku,
                'variant_name', pv.variant_name,
                'variant_attributes', pv.variant_attributes,
                'price', (
                    SELECT ppv.price
                    FROM product_prices ppv
                    INNER JOIN price_lists plv ON ppv.price_list_id = plv.id AND plv.is_active = true
                    WHERE ppv.product_id = pv.product_id
                      AND ppv.product_variant_id = pv.id
                      AND ppv.is_active = true
                      AND (ppv.valid_from IS NULL OR ppv.valid_from <= CURRENT_DATE)
                      AND (ppv.valid_to   IS NULL OR ppv.valid_to   >= CURRENT_DATE)
                    ORDER BY ppv.valid_from DESC NULLS LAST, ppv.id DESC
                    LIMIT 1
                ),
                'is_active', pv.is_active,
                'metadata', COALESCE(pv.metadata, '{}'::jsonb),
                'created_at', pv.created_at,
                'updated_at', pv.updated_at
            )
            ORDER BY pv.id
        ), '[]'::jsonb)
         FROM product_variants pv
         WHERE pv.product_id = cat.product_id),
        -- FIX 9.2: Include variant_id and variant_attributes in package_n_price JSON
        (SELECT COALESCE(jsonb_agg(s.rec ORDER BY s.pl_code, s.uom_code), '[]'::jsonb)
         FROM (
             SELECT pl.code AS pl_code, uom.code AS uom_code,
                    jsonb_build_object(
                        'product_name',        cat.product_name,
                        'price_list_id',       pl.id,
                        'price_list_code',     pl.code,
                        'price_list_name',     pl.name,
                        'price_list',          pl.name,
                        'price_list_type',     pl.price_list_type,
                        'currency_code',       pl.currency_code,
                        'uom_id',              uom.id,
                        'uom_code',            uom.code,
                        'uom',                 uom.code,
                        'uom_name',            uom.name,
                        'decimal_places',      uom.decimal_places,
                        'price',               pp.price,
                        'min_quantity',        pp.min_quantity,
                        'max_quantity',        pp.max_quantity,
                        'valid_from',          pp.valid_from,
                        'valid_to',            pp.valid_to,
                        'metadata',            COALESCE(pp.metadata, '{}'::jsonb),
                        'barcodes',            (SELECT COALESCE(jsonb_agg(pb.barcode), '[]'::jsonb)
                                                FROM product_barcodes pb
                                                WHERE pb.product_id = pp.product_id
                                                  AND (pb.product_variant_id = cat.product_variant_id
                                                       OR (cat.product_variant_id IS NULL AND pb.product_variant_id IS NULL)))
                    ) AS rec
             FROM product_prices pp
             INNER JOIN price_lists pl ON pp.price_list_id = pl.id AND pl.is_active = true
             LEFT JOIN units_of_measure uom ON pp.uom_id = uom.id
             WHERE pp.product_id = cat.product_id
               AND (pp.product_variant_id = cat.product_variant_id
                    OR (cat.product_variant_id IS NULL AND pp.product_variant_id IS NULL))
               AND pp.is_active = true
               AND (pp.valid_from IS NULL OR pp.valid_from <= CURRENT_DATE)
               AND (pp.valid_to   IS NULL OR pp.valid_to   >= CURRENT_DATE)
         ) AS s),
        (SELECT COALESCE(jsonb_agg(conv.cv ORDER BY conv.fu_code, conv.tu_code), '[]'::jsonb)
         FROM (
             SELECT fu.code AS fu_code, tu.code AS tu_code,
                    jsonb_build_object(
                        'from_uom_id', fu.id, 'from_uom_code', fu.code, 'from_uom_name', fu.name,
                        'to_uom_id',   tu.id, 'to_uom_code',   tu.code, 'to_uom_name',   tu.name,
                        'conversion_factor', puc.conversion_factor
                    ) AS cv
             FROM product_uom_conversions puc
             JOIN units_of_measure fu ON puc.from_uom_id = fu.id
             JOIN units_of_measure tu ON puc.to_uom_id   = tu.id
             WHERE puc.product_id = cat.product_id
         ) AS conv)
    FROM vw_pos_product_catalog cat
    LEFT JOIN LATERAL (
        SELECT 
            pr.name AS promo_name,
            pr.min_quantity AS promo_min_qty,
            pr.discount_value,
            pr.promotion_type,
            CASE 
                WHEN pr.promotion_type = 'percentage_discount' AND cat.retail_price IS NOT NULL THEN
                    ROUND(cat.retail_price * (1.0 - (pr.discount_value / 100.0)), 2)
                WHEN pr.promotion_type = 'fixed_discount' AND cat.retail_price IS NOT NULL THEN
                    GREATEST(0.00, cat.retail_price - pr.discount_value)
                ELSE cat.retail_price
            END AS calculated_promo_price,
            CASE
                WHEN pr.promotion_type = 'percentage_discount' AND pr.discount_value IS NOT NULL THEN
                    CONCAT(TRIM(TRAILING '.' FROM TRIM(TRAILING '0' FROM pr.discount_value::text)), '%')
                ELSE NULL
            END AS calc_discount_percent
        FROM promotions pr
        WHERE pr.is_active = true
          AND (pr.valid_from IS NULL OR pr.valid_from <= CURRENT_TIMESTAMP)
          AND (pr.valid_to IS NULL OR pr.valid_to >= CURRENT_TIMESTAMP)
          AND (cardinality(pr.store_ids) = 0 OR p_store_id = ANY(pr.store_ids))
          AND (
              pr.applies_to = 'all'
              OR (pr.applies_to = 'product' AND cat.product_id = ANY(pr.target_product_ids))
              OR (pr.applies_to = 'category' AND cat.category_id = ANY(pr.target_category_ids))
          )
        ORDER BY pr.created_at DESC
        LIMIT 1
    ) promo_rule ON true
    -- FIX 9.3: correct variant-aware inventory join
    LEFT JOIN inventory_stock inv
        ON inv.product_id = cat.product_id
        AND inv.store_id = p_store_id
        AND (inv.product_variant_id = cat.product_variant_id
             OR (cat.product_variant_id IS NULL AND inv.product_variant_id IS NULL))
    WHERE
        (p_category_id IS NULL OR cat.category_id = p_category_id)
        AND (p_search_term IS NULL
             OR cat.product_name ILIKE '%' || p_search_term || '%'
             OR cat.sku         ILIKE '%' || p_search_term || '%'
             OR cat.barcode     ILIKE '%' || p_search_term || '%')
        AND (p_include_out_of_stock = true OR COALESCE(inv.quantity_available, 0) > 0)
    ORDER BY cat.category_name, cat.product_name;
END;
$$;
-- +goose StatementEnd
-- Create "fn_pos_search_products" function
-- +goose StatementBegin
CREATE FUNCTION "public"."fn_pos_search_products" ("p_search_term" character varying, "p_store_id" integer, "p_limit" integer DEFAULT 50) RETURNS TABLE ("product_id" integer, "sku" character varying, "product_name" character varying, "category_name" character varying, "brand_name" character varying, "barcode" character varying, "effective_price" numeric, "has_promotion" boolean, "quantity_available" numeric, "is_in_stock" boolean, "relevance_score" integer, "package_n_price" jsonb, "product_uom_conversions" jsonb) LANGUAGE plpgsql AS $$
BEGIN
    RETURN QUERY
    SELECT 
        cat.product_id,
        cat.sku::VARCHAR,
        cat.product_name::VARCHAR,
        cat.category_name::VARCHAR,
        cat.brand_name::VARCHAR,
        cat.barcode::VARCHAR,
        cat.effective_price,
        cat.has_active_promotion,
        COALESCE(inv.quantity_available, 0)::NUMERIC,
        (COALESCE(inv.quantity_available, 0) > 0),
        (CASE 
            WHEN cat.sku ILIKE p_search_term THEN 100
            WHEN cat.product_name ILIKE p_search_term THEN 90
            WHEN cat.barcode = p_search_term THEN 95
            WHEN cat.sku ILIKE p_search_term || '%' THEN 80
            WHEN cat.product_name ILIKE p_search_term || '%' THEN 70
            WHEN cat.sku ILIKE '%' || p_search_term || '%' THEN 60
            WHEN cat.product_name ILIKE '%' || p_search_term || '%' THEN 50
            ELSE 40
        END)::INTEGER,
        (SELECT COALESCE(jsonb_agg(s.rec ORDER BY s.pl_code, s.uom_code), '[]'::jsonb)
         FROM (
             SELECT 
                 pl.code AS pl_code,
                 uom.code AS uom_code,
                 jsonb_build_object(
                     'product_name', p.name,
                     'price_list_id', pl.id,
                     'price_list_code', pl.code,
                     'price_list_name', pl.name,
                     'price_list', pl.name,
                     'price_list_type', pl.price_list_type,
                     'currency_code', pl.currency_code,
                     'uom_id', uom.id,
                     'uom_code', uom.code,
                     'uom', uom.code,
                     'uom_name', uom.name,
                     'decimal_places', uom.decimal_places,
                     'price', pp.price,
                     'min_quantity', pp.min_quantity,
                     'max_quantity', pp.max_quantity,
                     'valid_from', pp.valid_from,
                     'valid_to', pp.valid_to,
                     'metadata', COALESCE(pp.metadata, '{}'::jsonb),
                     'barcodes', (SELECT COALESCE(jsonb_agg(pb.barcode), '[]'::jsonb) FROM product_barcodes pb WHERE pb.product_id = pp.product_id)
                 ) AS rec
             FROM product_prices pp
             INNER JOIN products p ON pp.product_id = p.id
             INNER JOIN price_lists pl ON pp.price_list_id = pl.id AND pl.is_active = true
             LEFT JOIN units_of_measure uom ON pp.uom_id = uom.id
             WHERE pp.product_id = cat.product_id
               AND pp.is_active = true
               AND (pp.valid_from IS NULL OR pp.valid_from <= CURRENT_DATE)
               AND (pp.valid_to IS NULL OR pp.valid_to >= CURRENT_DATE)
         ) AS s),
        (SELECT COALESCE(jsonb_agg(conv.cv ORDER BY conv.fu_code, conv.tu_code), '[]'::jsonb)
         FROM (
             SELECT fu.code AS fu_code, tu.code AS tu_code,
                    jsonb_build_object(
                        'from_uom_id', fu.id, 'from_uom_code', fu.code, 'from_uom_name', fu.name,
                        'to_uom_id', tu.id, 'to_uom_code', tu.code, 'to_uom_name', tu.name,
                        'conversion_factor', puc.conversion_factor
                    ) AS cv
             FROM product_uom_conversions puc
             JOIN units_of_measure fu ON puc.from_uom_id = fu.id
             JOIN units_of_measure tu ON puc.to_uom_id = tu.id
             WHERE puc.product_id = cat.product_id
         ) AS conv)
    FROM vw_pos_product_catalog cat
    LEFT JOIN inventory_stock inv ON cat.product_id = inv.product_id AND inv.store_id = p_store_id
    WHERE 
        cat.product_name ILIKE '%' || p_search_term || '%'
        OR cat.sku ILIKE '%' || p_search_term || '%'
        OR cat.barcode ILIKE '%' || p_search_term || '%'
    ORDER BY 11 DESC, cat.product_name
    LIMIT p_limit;
END;
$$;
-- +goose StatementEnd
-- Create "purchase_order_lines" table
CREATE TABLE "public"."purchase_order_lines" (
  "id" serial NOT NULL,
  "purchase_order_id" integer NOT NULL,
  "product_id" integer NOT NULL,
  "product_variant_id" integer NULL,
  "quantity" numeric(15,3) NOT NULL,
  "uom_id" integer NULL,
  "unit_price" numeric(15,4) NOT NULL,
  "discount_amount" numeric(15,2) NULL DEFAULT 0,
  "tax_amount" numeric(15,2) NULL DEFAULT 0,
  "subtotal" numeric(15,2) NOT NULL,
  "line_total" numeric(15,2) NULL DEFAULT 0,
  "received_quantity" numeric(15,3) NULL DEFAULT 0,
  "line_number" integer NULL,
  "metadata" jsonb NULL DEFAULT '{}',
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "purchase_order_lines_product_id_fkey" FOREIGN KEY ("product_id") REFERENCES "public"."products" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "purchase_order_lines_product_variant_id_fkey" FOREIGN KEY ("product_variant_id") REFERENCES "public"."product_variants" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "purchase_order_lines_purchase_order_id_fkey" FOREIGN KEY ("purchase_order_id") REFERENCES "public"."purchase_orders" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "purchase_order_lines_uom_id_fkey" FOREIGN KEY ("uom_id") REFERENCES "public"."units_of_measure" ("id") ON UPDATE NO ACTION ON DELETE SET NULL
);
-- Create index "idx_purchase_order_lines_product_id" to table: "purchase_order_lines"
CREATE INDEX "idx_purchase_order_lines_product_id" ON "public"."purchase_order_lines" ("product_id");
-- Create index "idx_purchase_order_lines_purchase_order_id" to table: "purchase_order_lines"
CREATE INDEX "idx_purchase_order_lines_purchase_order_id" ON "public"."purchase_order_lines" ("purchase_order_id");
-- Create "goods_receipt_note_items" table
CREATE TABLE "public"."goods_receipt_note_items" (
  "id" serial NOT NULL,
  "grn_id" integer NOT NULL,
  "purchase_order_line_id" integer NULL,
  "product_id" integer NOT NULL,
  "product_variant_id" integer NULL,
  "storage_location_id" integer NULL,
  "quantity_received" numeric(15,3) NOT NULL,
  "quantity_rejected" numeric(15,3) NULL DEFAULT 0,
  "uom_id" integer NULL,
  "unit_cost" numeric(15,4) NULL,
  "batch_number" character varying(100) NULL,
  "expiry_date" date NULL,
  "rejection_reason" text NULL,
  "notes" text NULL,
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "goods_receipt_note_items_grn_id_fkey" FOREIGN KEY ("grn_id") REFERENCES "public"."goods_receipt_notes" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "goods_receipt_note_items_product_id_fkey" FOREIGN KEY ("product_id") REFERENCES "public"."products" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "goods_receipt_note_items_product_variant_id_fkey" FOREIGN KEY ("product_variant_id") REFERENCES "public"."product_variants" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "goods_receipt_note_items_purchase_order_line_id_fkey" FOREIGN KEY ("purchase_order_line_id") REFERENCES "public"."purchase_order_lines" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "goods_receipt_note_items_storage_location_id_fkey" FOREIGN KEY ("storage_location_id") REFERENCES "public"."storage_locations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "goods_receipt_note_items_uom_id_fkey" FOREIGN KEY ("uom_id") REFERENCES "public"."units_of_measure" ("id") ON UPDATE NO ACTION ON DELETE SET NULL
);
-- Create "fn_process_goods_receipt" function
-- +goose StatementBegin
CREATE FUNCTION "public"."fn_process_goods_receipt" ("p_grn_id" integer) RETURNS TABLE ("success" boolean, "message" text) LANGUAGE plpgsql AS $$
DECLARE
    v_grn RECORD;
    v_item RECORD;
    v_all_po_received BOOLEAN := true;
    v_any_po_received BOOLEAN := false;
    v_po_line RECORD;
BEGIN
    SELECT * INTO v_grn FROM goods_receipt_notes WHERE id = p_grn_id FOR UPDATE;
    IF v_grn IS NULL THEN
        RETURN QUERY SELECT false, 'Goods receipt note not found.';
        RETURN;
    END IF;

    IF v_grn.status = 'completed' THEN
        RETURN QUERY SELECT false, 'Goods receipt note is already completed.';
        RETURN;
    END IF;

    FOR v_item IN SELECT * FROM goods_receipt_note_items WHERE grn_id = p_grn_id FOR UPDATE LOOP
        IF v_item.quantity_received <= 0 THEN
            CONTINUE;
        END IF;

        -- Update PO Line if associated
        IF v_item.purchase_order_line_id IS NOT NULL THEN
            UPDATE purchase_order_lines
            SET received_quantity = received_quantity + v_item.quantity_received
            WHERE id = v_item.purchase_order_line_id;
        END IF;

        -- Increment inventory stock at store
        INSERT INTO inventory_stock (
            product_id, product_variant_id, store_id, storage_location_id,
            quantity_on_hand, quantity_available
        ) VALUES (
            v_item.product_id, v_item.product_variant_id, v_grn.store_id, v_item.storage_location_id,
            v_item.quantity_received, v_item.quantity_received
        )
        ON CONFLICT (product_id, COALESCE(product_variant_id, -1), store_id)
        DO UPDATE SET
            quantity_on_hand = inventory_stock.quantity_on_hand + EXCLUDED.quantity_on_hand,
            quantity_available = inventory_stock.quantity_available + EXCLUDED.quantity_available,
            updated_at = CURRENT_TIMESTAMP;

        -- Insert stock movement (purchase_receipt)
        INSERT INTO stock_movements (
            movement_type, reference_type, reference_id, product_id, product_variant_id,
            to_store_id, to_location_id, quantity, uom_id, batch_number,
            posted_by, status, cost_per_unit, total_value, metadata
        ) VALUES (
            'purchase_receipt', 'goods_receipt_note', p_grn_id, v_item.product_id, v_item.product_variant_id,
            v_grn.store_id, v_item.storage_location_id, v_item.quantity_received, v_item.uom_id, COALESCE(v_item.batch_number, ''),
            v_grn.received_by, 'completed', COALESCE(v_item.unit_cost, 0), (COALESCE(v_item.unit_cost, 0) * v_item.quantity_received),
            jsonb_build_object('grn_number', v_grn.grn_number, 'delivery_note_number', COALESCE(v_grn.delivery_note_number, ''))
        );
    END LOOP;

    -- Update GRN status
    UPDATE goods_receipt_notes
    SET status = 'completed',
        updated_at = CURRENT_TIMESTAMP
    WHERE id = p_grn_id;

    -- Update Purchase Order status if PO ID present
    IF v_grn.purchase_order_id IS NOT NULL THEN
        FOR v_po_line IN SELECT quantity, received_quantity FROM purchase_order_lines WHERE purchase_order_id = v_grn.purchase_order_id LOOP
            IF v_po_line.received_quantity < v_po_line.quantity THEN
                v_all_po_received := false;
            END IF;
            IF v_po_line.received_quantity > 0 THEN
                v_any_po_received := true;
            END IF;
        END LOOP;

        IF v_all_po_received THEN
            UPDATE purchase_orders SET status = 'received', updated_at = CURRENT_TIMESTAMP WHERE id = v_grn.purchase_order_id;
        ELSIF v_any_po_received THEN
            UPDATE purchase_orders SET status = 'partially_received', updated_at = CURRENT_TIMESTAMP WHERE id = v_grn.purchase_order_id;
        END IF;
    END IF;

    RETURN QUERY SELECT true, 'Goods receipt processed successfully.';
END;
$$;
-- +goose StatementEnd
-- Create "fn_process_stock_transfer" function
-- +goose StatementBegin
CREATE FUNCTION "public"."fn_process_stock_transfer" ("p_from_store_id" integer, "p_to_store_id" integer, "p_product_id" integer, "p_product_variant_id" integer, "p_quantity" numeric, "p_from_location_id" integer DEFAULT NULL::integer, "p_to_location_id" integer DEFAULT NULL::integer, "p_batch_number" character varying DEFAULT NULL::character varying, "p_performed_by" integer DEFAULT NULL::integer, "p_notes" text DEFAULT NULL::text) RETURNS TABLE ("success" boolean, "message" text, "movement_id" integer) LANGUAGE plpgsql AS $$
DECLARE
    v_available  DECIMAL(15,3);
    v_movement_id INTEGER;
    v_ref_num    VARCHAR(50);
BEGIN
    -- Validate same store check
    IF p_from_store_id = p_to_store_id THEN
        RETURN QUERY SELECT false, 'Source and destination stores must differ.', NULL::INTEGER;
        RETURN;
    END IF;

    -- Lock and read available stock at source
    SELECT quantity_available INTO v_available
    FROM inventory_stock
    WHERE product_id = p_product_id
      AND (product_variant_id = p_product_variant_id OR (product_variant_id IS NULL AND p_product_variant_id IS NULL))
      AND store_id = p_from_store_id
    FOR UPDATE;

    IF v_available IS NULL OR v_available < p_quantity THEN
        RETURN QUERY SELECT false,
            format('Insufficient stock. Available: %s, Requested: %s', COALESCE(v_available, 0), p_quantity),
            NULL::INTEGER;
        RETURN;
    END IF;

    v_ref_num := 'TRF-' || to_char(CURRENT_TIMESTAMP, 'YYYYMMDDHH24MISS') || '-' || p_from_store_id || '-' || p_to_store_id;

    -- Deduct from source
    UPDATE inventory_stock
    SET quantity_on_hand   = quantity_on_hand   - p_quantity,
        quantity_available = quantity_available - p_quantity,
        quantity_in_transit = quantity_in_transit + p_quantity,
        updated_at = CURRENT_TIMESTAMP
    WHERE product_id = p_product_id
      AND (product_variant_id = p_product_variant_id OR (product_variant_id IS NULL AND p_product_variant_id IS NULL))
      AND store_id = p_from_store_id;

    -- Add to destination
    INSERT INTO inventory_stock (product_id, product_variant_id, store_id, storage_location_id,
        quantity_on_hand, quantity_available, quantity_in_transit)
    VALUES (p_product_id, p_product_variant_id, p_to_store_id, p_to_location_id,
            p_quantity, p_quantity, 0)
    ON CONFLICT (product_id, COALESCE(product_variant_id, -1), store_id)
    DO UPDATE SET
        quantity_on_hand   = inventory_stock.quantity_on_hand   + EXCLUDED.quantity_on_hand,
        quantity_available = inventory_stock.quantity_available + EXCLUDED.quantity_available,
        updated_at = CURRENT_TIMESTAMP;

    -- Clear in-transit at source
    UPDATE inventory_stock
    SET quantity_in_transit = GREATEST(0, quantity_in_transit - p_quantity),
        updated_at = CURRENT_TIMESTAMP
    WHERE product_id = p_product_id
      AND (product_variant_id = p_product_variant_id OR (product_variant_id IS NULL AND p_product_variant_id IS NULL))
      AND store_id = p_from_store_id;

    -- Record movement
    INSERT INTO stock_movements (movement_type, reference_type, product_id, product_variant_id,
        from_store_id, to_store_id, from_location_id, to_location_id,
        quantity, batch_number, posted_by, status,
        metadata)
    VALUES ('transfer', 'manual', p_product_id, p_product_variant_id,
            p_from_store_id, p_to_store_id, p_from_location_id, p_to_location_id,
            p_quantity, p_batch_number, p_performed_by, 'completed',
            jsonb_build_object('reference_number', v_ref_num, 'notes', p_notes))
    RETURNING id INTO v_movement_id;

    RETURN QUERY SELECT true, 'Transfer completed successfully. Ref: ' || v_ref_num, v_movement_id;
END;
$$;
-- +goose StatementEnd
-- Create "fn_receive_transfer_request" function
-- +goose StatementBegin
CREATE FUNCTION "public"."fn_receive_transfer_request" ("p_transfer_request_id" integer, "p_received_by" integer) RETURNS TABLE ("success" boolean, "message" text) LANGUAGE plpgsql AS $$
DECLARE
    v_req RECORD;
    v_item RECORD;
    v_qty DECIMAL(15,3);
    v_uom_code VARCHAR;
    v_base_qty DECIMAL(15,3);
BEGIN
    SELECT * INTO v_req FROM transfer_requests WHERE id = p_transfer_request_id FOR UPDATE;
    IF v_req IS NULL THEN
        RETURN QUERY SELECT false, 'Transfer request not found.';
        RETURN;
    END IF;

    IF v_req.status NOT IN ('shipped', 'partially_received') THEN
        RETURN QUERY SELECT false, 'Transfer request must be shipped or partially_received to receive.';
        RETURN;
    END IF;

    FOR v_item IN SELECT * FROM transfer_request_items WHERE transfer_request_id = p_transfer_request_id FOR UPDATE LOOP
        v_qty := CASE WHEN v_item.shipped_quantity > 0 THEN v_item.shipped_quantity ELSE v_item.requested_quantity END;
        IF v_qty <= 0 THEN
            CONTINUE;
        END IF;

        -- Retrieve UOM code for conversion
        SELECT code INTO v_uom_code FROM units_of_measure WHERE id = v_item.uom_id;
        
        -- Convert requested qty to base unit qty
        v_base_qty := fn_convert_uom_quantity(v_item.product_id, v_uom_code, v_qty);
        IF v_base_qty IS NULL THEN
            v_base_qty := v_qty;
        END IF;

        -- Decrement in-transit and increment on_hand & available at destination store using v_base_qty
        UPDATE inventory_stock
        SET quantity_in_transit = GREATEST(0, quantity_in_transit - v_base_qty),
            quantity_on_hand = quantity_on_hand + v_base_qty,
            quantity_available = quantity_available + v_base_qty,
            updated_at = CURRENT_TIMESTAMP
        WHERE product_id = v_item.product_id
          AND (product_variant_id = v_item.product_variant_id OR (product_variant_id IS NULL AND v_item.product_variant_id IS NULL))
          AND store_id = v_req.to_store_id;

        -- Update item received_quantity
        UPDATE transfer_request_items
        SET received_quantity = v_qty
        WHERE id = v_item.id;

        -- Record stock movement (transfer_in / completed) using v_base_qty
        INSERT INTO stock_movements (
            movement_type, reference_type, reference_id, product_id, product_variant_id,
            from_store_id, to_store_id, from_location_id, to_location_id,
            quantity, uom_id, batch_number, posted_by, status, metadata
        ) VALUES (
            'transfer_in', 'transfer_request', p_transfer_request_id, v_item.product_id, v_item.product_variant_id,
            v_req.from_store_id, v_req.to_store_id, v_item.from_location_id, v_item.to_location_id,
            v_base_qty, v_item.uom_id, v_item.batch_number, p_received_by, 'completed',
            jsonb_build_object('transfer_number', v_req.transfer_number)
        );
    END LOOP;

    -- Update header
    UPDATE transfer_requests
    SET status = 'received',
        received_by = p_received_by,
        received_at = CURRENT_TIMESTAMP,
        updated_at = CURRENT_TIMESTAMP
    WHERE id = p_transfer_request_id;

    RETURN QUERY SELECT true, 'Transfer request received successfully.';
END;
$$;
-- +goose StatementEnd
-- Create "stock_counts" table
CREATE TABLE "public"."stock_counts" (
  "id" serial NOT NULL,
  "count_number" character varying(50) NOT NULL,
  "store_id" integer NOT NULL,
  "storage_location_id" integer NULL,
  "count_type" character varying(50) NULL,
  "status" character varying(50) NULL DEFAULT 'planned',
  "scheduled_date" date NULL,
  "started_at" timestamp NULL,
  "completed_at" timestamp NULL,
  "counted_by" integer NULL,
  "approved_by" integer NULL,
  "metadata" jsonb NULL DEFAULT '{}',
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "stock_counts_count_number_key" UNIQUE ("count_number"),
  CONSTRAINT "stock_counts_approved_by_fkey" FOREIGN KEY ("approved_by") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "stock_counts_counted_by_fkey" FOREIGN KEY ("counted_by") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "stock_counts_storage_location_id_fkey" FOREIGN KEY ("storage_location_id") REFERENCES "public"."storage_locations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "stock_counts_store_id_fkey" FOREIGN KEY ("store_id") REFERENCES "public"."stores" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_stock_counts_count_number" to table: "stock_counts"
CREATE INDEX "idx_stock_counts_count_number" ON "public"."stock_counts" ("count_number");
-- Create index "idx_stock_counts_status" to table: "stock_counts"
CREATE INDEX "idx_stock_counts_status" ON "public"."stock_counts" ("status");
-- Create index "idx_stock_counts_store_id" to table: "stock_counts"
CREATE INDEX "idx_stock_counts_store_id" ON "public"."stock_counts" ("store_id");
-- Create "stock_count_lines" table
CREATE TABLE "public"."stock_count_lines" (
  "id" serial NOT NULL,
  "stock_count_id" integer NOT NULL,
  "product_id" integer NOT NULL,
  "product_variant_id" integer NULL,
  "storage_location_id" integer NULL,
  "expected_quantity" numeric(15,3) NULL DEFAULT 0,
  "system_quantity" numeric(15,3) NULL DEFAULT 0,
  "counted_quantity" numeric(15,3) NULL DEFAULT 0,
  "variance" numeric(15,3) NULL DEFAULT 0,
  "variance_value" numeric(15,2) NULL DEFAULT 0,
  "counted_at" timestamp NULL,
  "uom_id" integer NULL,
  "batch_number" character varying(100) NULL,
  "serial_number" character varying(100) NULL,
  "metadata" jsonb NULL DEFAULT '{}',
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "stock_count_lines_product_id_fkey" FOREIGN KEY ("product_id") REFERENCES "public"."products" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "stock_count_lines_product_variant_id_fkey" FOREIGN KEY ("product_variant_id") REFERENCES "public"."product_variants" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "stock_count_lines_stock_count_id_fkey" FOREIGN KEY ("stock_count_id") REFERENCES "public"."stock_counts" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "stock_count_lines_storage_location_id_fkey" FOREIGN KEY ("storage_location_id") REFERENCES "public"."storage_locations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "stock_count_lines_uom_id_fkey" FOREIGN KEY ("uom_id") REFERENCES "public"."units_of_measure" ("id") ON UPDATE NO ACTION ON DELETE SET NULL
);
-- Create index "idx_stock_count_lines_product_id" to table: "stock_count_lines"
CREATE INDEX "idx_stock_count_lines_product_id" ON "public"."stock_count_lines" ("product_id");
-- Create index "idx_stock_count_lines_stock_count_id" to table: "stock_count_lines"
CREATE INDEX "idx_stock_count_lines_stock_count_id" ON "public"."stock_count_lines" ("stock_count_id");
-- Create "fn_reconcile_stock_count" function
-- +goose StatementBegin
CREATE FUNCTION "public"."fn_reconcile_stock_count" ("p_count_id" integer) RETURNS TABLE ("success" boolean, "message" text, "lines_updated" integer) LANGUAGE plpgsql AS $$
DECLARE
    v_count        RECORD;
    v_line         RECORD;
    v_lines_updated INTEGER := 0;
BEGIN
    SELECT * INTO v_count FROM stock_counts WHERE id = p_count_id;

    IF NOT FOUND THEN
        RETURN QUERY SELECT false, 'Stock count not found.', 0;
        RETURN;
    END IF;
    IF v_count.status <> 'approved' THEN
        RETURN QUERY SELECT false, 'Stock count must be in ''approved'' status to reconcile.', 0;
        RETURN;
    END IF;

    FOR v_line IN
        SELECT * FROM stock_count_lines WHERE stock_count_id = p_count_id
    LOOP
        -- Update inventory_stock to match counted quantity
        UPDATE inventory_stock
        SET quantity_on_hand   = v_line.counted_quantity,
            quantity_available = GREATEST(0, v_line.counted_quantity - COALESCE(quantity_allocated, 0)),
            last_counted_at    = CURRENT_TIMESTAMP,
            updated_at         = CURRENT_TIMESTAMP
        WHERE product_id = v_line.product_id
          AND (product_variant_id = v_line.product_variant_id
               OR (product_variant_id IS NULL AND v_line.product_variant_id IS NULL))
          AND store_id = v_count.store_id;

        -- Record the adjustment movement
        INSERT INTO stock_movements (movement_type, reference_type, reference_id, product_id,
            product_variant_id, to_store_id, quantity, batch_number, serial_number,
            posted_by, status, metadata)
        VALUES ('count_adjustment', 'stock_count', p_count_id, v_line.product_id,
                v_line.product_variant_id, v_count.store_id,
                v_line.counted_quantity - v_line.system_quantity,
                v_line.batch_number, v_line.serial_number,
                v_count.approved_by, 'completed',
                jsonb_build_object('count_id', p_count_id, 'variance', v_line.variance));

        v_lines_updated := v_lines_updated + 1;
    END LOOP;

    -- Mark count as reconciled
    UPDATE stock_counts SET status = 'approved', updated_at = CURRENT_TIMESTAMP
    WHERE id = p_count_id;

    RETURN QUERY SELECT true, format('Reconciliation complete. %s lines adjusted.', v_lines_updated), v_lines_updated;
END;
$$;
-- +goose StatementEnd
-- Create "pos_payments" table
CREATE TABLE "public"."pos_payments" (
  "id" serial NOT NULL,
  "transaction_id" integer NOT NULL,
  "payment_method" character varying(50) NOT NULL,
  "payment_gateway" character varying(50) NULL,
  "amount" numeric(15,2) NOT NULL,
  "payment_reference" character varying(100) NULL,
  "reference_number" character varying(100) NULL,
  "payment_date" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  "metadata" jsonb NULL DEFAULT '{}',
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "pos_payments_transaction_id_fkey" FOREIGN KEY ("transaction_id") REFERENCES "public"."pos_transactions" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_pos_payments_payment_method" to table: "pos_payments"
CREATE INDEX "idx_pos_payments_payment_method" ON "public"."pos_payments" ("payment_method");
-- Create index "idx_pos_payments_transaction_id" to table: "pos_payments"
CREATE INDEX "idx_pos_payments_transaction_id" ON "public"."pos_payments" ("transaction_id");
-- Create "pos_transaction_lines" table
CREATE TABLE "public"."pos_transaction_lines" (
  "id" serial NOT NULL,
  "transaction_id" integer NOT NULL,
  "product_id" integer NOT NULL,
  "product_variant_id" integer NULL,
  "quantity" numeric(15,3) NOT NULL,
  "uom_id" integer NULL,
  "unit_price" numeric(15,4) NOT NULL,
  "discount_amount" numeric(15,2) NULL DEFAULT 0,
  "tax_amount" numeric(15,2) NULL DEFAULT 0,
  "subtotal" numeric(15,2) NOT NULL,
  "line_total" numeric(15,2) NULL DEFAULT 0,
  "cost_price" numeric(15,2) NULL DEFAULT 0,
  "line_number" integer NULL,
  "serial_number" character varying(100) NULL,
  "batch_number" character varying(100) NULL,
  "metadata" jsonb NULL DEFAULT '{}',
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "pos_transaction_lines_product_id_fkey" FOREIGN KEY ("product_id") REFERENCES "public"."products" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "pos_transaction_lines_product_variant_id_fkey" FOREIGN KEY ("product_variant_id") REFERENCES "public"."product_variants" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "pos_transaction_lines_transaction_id_fkey" FOREIGN KEY ("transaction_id") REFERENCES "public"."pos_transactions" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "pos_transaction_lines_uom_id_fkey" FOREIGN KEY ("uom_id") REFERENCES "public"."units_of_measure" ("id") ON UPDATE NO ACTION ON DELETE SET NULL
);
-- Create index "idx_pos_transaction_lines_product_id" to table: "pos_transaction_lines"
CREATE INDEX "idx_pos_transaction_lines_product_id" ON "public"."pos_transaction_lines" ("product_id");
-- Create index "idx_pos_transaction_lines_transaction_id" to table: "pos_transaction_lines"
CREATE INDEX "idx_pos_transaction_lines_transaction_id" ON "public"."pos_transaction_lines" ("transaction_id");
-- Create "fn_refresh_daily_analytics" function
-- +goose StatementBegin
CREATE FUNCTION "public"."fn_refresh_daily_analytics" ("p_date" date DEFAULT CURRENT_DATE) RETURNS void LANGUAGE plpgsql AS $$
BEGIN
    -- Refresh sales_analytics from pos_transactions and sales_orders_v2
    INSERT INTO sales_analytics (
        organization_id, store_id, product_id, category_id,
        date, hour, day_of_week, month, quarter, year,
        units_sold, revenue, discounts, taxes, net_revenue, transactions,
        payment_method, payment_gateway
    )
    SELECT
        organization_id, store_id, product_id, category_id,
        date, hour, day_of_week, month, quarter, year,
        SUM(units_sold), SUM(revenue), SUM(discounts), SUM(taxes), SUM(net_revenue), SUM(transactions),
        payment_method, payment_gateway
    FROM (
        -- Data from POS transactions
        SELECT
            s.organization_id,
            pt.store_id,
            ptl.product_id,
            p.category_id,
            p_date AS date,
            EXTRACT(HOUR FROM pt.transaction_date)::INTEGER AS hour,
            EXTRACT(DOW  FROM pt.transaction_date)::INTEGER AS day_of_week,
            EXTRACT(MONTH FROM p_date)::INTEGER AS month,
            EXTRACT(QUARTER FROM p_date)::INTEGER AS quarter,
            EXTRACT(YEAR FROM p_date)::INTEGER AS year,
            ptl.quantity AS units_sold,
            ptl.line_total AS revenue,
            ptl.discount_amount AS discounts,
            ptl.tax_amount AS taxes,
            (ptl.line_total - ptl.tax_amount) AS net_revenue,
            1 AS transactions,
            pp.payment_method,
            pp.payment_gateway
        FROM pos_transactions pt
        JOIN pos_transaction_lines ptl ON ptl.transaction_id = pt.id
        LEFT JOIN pos_payments pp ON pp.transaction_id = pt.id
        JOIN products p ON p.id = ptl.product_id
        JOIN stores s ON s.id = pt.store_id
        WHERE DATE(pt.transaction_date) = p_date
          AND pt.status = 'completed'

        UNION ALL

        -- Data from sales_orders_v2 (excluding those that were synced to POS to avoid double counting)
        SELECT
            o.organization_id,
            o.store_id,
            ol.product_id,
            p.category_id,
            p_date AS date,
            EXTRACT(HOUR FROM o.order_date)::INTEGER AS hour,
            EXTRACT(DOW  FROM o.order_date)::INTEGER AS day_of_week,
            EXTRACT(MONTH FROM p_date)::INTEGER AS month,
            EXTRACT(QUARTER FROM p_date)::INTEGER AS quarter,
            EXTRACT(YEAR FROM p_date)::INTEGER AS year,
            ol.quantity_ordered::DECIMAL(15,3) AS units_sold,
            ol.line_total AS revenue,
            COALESCE(ol.discount_amount, 0) AS discounts,
            COALESCE(ol.tax_amount, 0) AS taxes,
            (ol.line_total - COALESCE(ol.tax_amount, 0)) AS net_revenue,
            1 AS transactions,
            o.payment_method,
            o.payment_gateway
        FROM sales_orders_v2 o
        JOIN sales_order_lines_v2 ol ON ol.sales_order_id = o.id
        JOIN products p ON p.id = ol.product_id
        WHERE DATE(o.order_date) = p_date
          AND o.order_status IN ('confirmed', 'processing', 'partially_fulfilled', 'fulfilled', 'shipped', 'delivered')
          AND NOT EXISTS (SELECT 1 FROM pos_transactions WHERE sales_order_id = o.id)
    ) aggregated_sales
    GROUP BY organization_id, store_id, product_id, category_id,
             date, hour, day_of_week, month, quarter, year,
             payment_method, payment_gateway
    ON CONFLICT (organization_id, store_id, product_id, date, hour, payment_method, payment_gateway) 
    DO UPDATE SET
        units_sold = EXCLUDED.units_sold,
        revenue = EXCLUDED.revenue,
        discounts = EXCLUDED.discounts,
        taxes = EXCLUDED.taxes,
        net_revenue = EXCLUDED.net_revenue,
        transactions = EXCLUDED.transactions,
        updated_at = CURRENT_TIMESTAMP;

    -- Refresh profit_loss_analytics
    INSERT INTO profit_loss_analytics (
        organization_id, store_id, date, period_type, month, quarter, year,
        gross_revenue, sales_discounts, net_revenue, cogs
    )
    SELECT
        s.organization_id,
        pt.store_id,
        p_date,
        'daily',
        EXTRACT(MONTH FROM p_date)::INTEGER,
        EXTRACT(QUARTER FROM p_date)::INTEGER,
        EXTRACT(YEAR FROM p_date)::INTEGER,
        SUM(pt.total_amount),
        SUM(pt.discount_amount),
        SUM(pt.total_amount - pt.discount_amount),
        SUM(pt.total_cost)
    FROM pos_transactions pt
    JOIN stores s ON s.id = pt.store_id
    WHERE DATE(pt.transaction_date) = p_date
      AND pt.status = 'completed'
    GROUP BY s.organization_id, pt.store_id
    ON CONFLICT DO NOTHING;
END;
$$;
-- +goose StatementEnd
-- Create "fn_ship_transfer_request" function
-- +goose StatementBegin
CREATE FUNCTION "public"."fn_ship_transfer_request" ("p_transfer_request_id" integer, "p_shipped_by" integer) RETURNS TABLE ("success" boolean, "message" text) LANGUAGE plpgsql AS $$
DECLARE
    v_req RECORD;
    v_item RECORD;
    v_available DECIMAL(15,3);
    v_qty DECIMAL(15,3);
    v_uom_code VARCHAR;
    v_base_qty DECIMAL(15,3);
BEGIN
    SELECT * INTO v_req FROM transfer_requests WHERE id = p_transfer_request_id FOR UPDATE;
    IF v_req IS NULL THEN
        RETURN QUERY SELECT false, 'Transfer request not found.';
        RETURN;
    END IF;

    IF v_req.status NOT IN ('approved', 'pending_approval', 'draft') THEN
        RETURN QUERY SELECT false, 'Transfer request must be approved to ship.';
        RETURN;
    END IF;

    FOR v_item IN SELECT * FROM transfer_request_items WHERE transfer_request_id = p_transfer_request_id FOR UPDATE LOOP
        v_qty := CASE WHEN v_item.approved_quantity > 0 THEN v_item.approved_quantity ELSE v_item.requested_quantity END;
        IF v_qty <= 0 THEN
            CONTINUE;
        END IF;

        -- Retrieve UOM code for conversion
        SELECT code INTO v_uom_code FROM units_of_measure WHERE id = v_item.uom_id;
        
        -- Convert requested qty to base unit qty
        v_base_qty := fn_convert_uom_quantity(v_item.product_id, v_uom_code, v_qty);
        IF v_base_qty IS NULL THEN
            v_base_qty := v_qty;
        END IF;

        -- Check available stock at source using v_base_qty
        SELECT quantity_available INTO v_available
        FROM inventory_stock
        WHERE product_id = v_item.product_id
          AND (product_variant_id = v_item.product_variant_id OR (product_variant_id IS NULL AND v_item.product_variant_id IS NULL))
          AND store_id = v_req.from_store_id
        FOR UPDATE;

        IF v_available IS NULL OR v_available < v_base_qty THEN
            RETURN QUERY SELECT false, format('Insufficient stock for product ID %s at source store.', v_item.product_id);
            RETURN;
        END IF;

        -- Deduct from source store using v_base_qty
        UPDATE inventory_stock
        SET quantity_on_hand = quantity_on_hand - v_base_qty,
            quantity_available = quantity_available - v_base_qty,
            updated_at = CURRENT_TIMESTAMP
        WHERE product_id = v_item.product_id
          AND (product_variant_id = v_item.product_variant_id OR (product_variant_id IS NULL AND v_item.product_variant_id IS NULL))
          AND store_id = v_req.from_store_id;

        -- Increment quantity_in_transit at destination store using v_base_qty
        INSERT INTO inventory_stock (product_id, product_variant_id, store_id, storage_location_id,
            quantity_on_hand, quantity_available, quantity_in_transit)
        VALUES (v_item.product_id, v_item.product_variant_id, v_req.to_store_id, v_item.to_location_id,
                0, 0, v_base_qty)
        ON CONFLICT (product_id, COALESCE(product_variant_id, -1), store_id)
        DO UPDATE SET
            quantity_in_transit = inventory_stock.quantity_in_transit + EXCLUDED.quantity_in_transit,
            updated_at = CURRENT_TIMESTAMP;

        -- Update item shipped_quantity
        UPDATE transfer_request_items
        SET shipped_quantity = v_qty,
            approved_quantity = v_qty
        WHERE id = v_item.id;

        -- Record stock movement (transfer_out / shipped) using v_base_qty
        INSERT INTO stock_movements (
            movement_type, reference_type, reference_id, product_id, product_variant_id,
            from_store_id, to_store_id, from_location_id, to_location_id,
            quantity, uom_id, batch_number, posted_by, status, metadata
        ) VALUES (
            'transfer_out', 'transfer_request', p_transfer_request_id, v_item.product_id, v_item.product_variant_id,
            v_req.from_store_id, v_req.to_store_id, v_item.from_location_id, v_item.to_location_id,
            v_base_qty, v_item.uom_id, v_item.batch_number, p_shipped_by, 'shipped',
            jsonb_build_object('transfer_number', v_req.transfer_number)
        );
    END LOOP;

    -- Update header
    UPDATE transfer_requests
    SET status = 'shipped',
        shipped_by = p_shipped_by,
        shipped_at = CURRENT_TIMESTAMP,
        updated_at = CURRENT_TIMESTAMP
    WHERE id = p_transfer_request_id;

    RETURN QUERY SELECT true, 'Transfer request shipped successfully.';
END;
$$;
-- +goose StatementEnd
-- Create "audit_logs" table
CREATE TABLE "public"."audit_logs" (
  "id" bigserial NOT NULL,
  "organization_id" integer NULL,
  "table_name" character varying(100) NOT NULL,
  "record_id" character varying(100) NOT NULL,
  "action" character varying(20) NOT NULL,
  "old_values" jsonb NULL,
  "new_values" jsonb NULL,
  "changed_fields" text[] NULL,
  "performed_by" integer NULL,
  "ip_address" inet NULL,
  "user_agent" text NULL,
  "session_id" character varying(255) NULL,
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "audit_logs_organization_id_fkey" FOREIGN KEY ("organization_id") REFERENCES "public"."organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "audit_logs_performed_by_fkey" FOREIGN KEY ("performed_by") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "audit_logs_action_check" CHECK ((action)::text = ANY ((ARRAY['INSERT'::character varying, 'UPDATE'::character varying, 'DELETE'::character varying, 'SELECT'::character varying])::text[]))
);
-- Create "chart_of_accounts" table
CREATE TABLE "public"."chart_of_accounts" (
  "id" serial NOT NULL,
  "organization_id" integer NOT NULL,
  "account_code" character varying(50) NOT NULL,
  "account_name" character varying(255) NOT NULL,
  "account_type" character varying(30) NOT NULL,
  "parent_account_id" integer NULL,
  "is_active" boolean NULL DEFAULT true,
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "chart_of_accounts_account_code_key" UNIQUE ("account_code"),
  CONSTRAINT "chart_of_accounts_organization_id_fkey" FOREIGN KEY ("organization_id") REFERENCES "public"."organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "chart_of_accounts_parent_account_id_fkey" FOREIGN KEY ("parent_account_id") REFERENCES "public"."chart_of_accounts" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "chart_of_accounts_account_type_check" CHECK ((account_type)::text = ANY ((ARRAY['asset'::character varying, 'liability'::character varying, 'equity'::character varying, 'revenue'::character varying, 'expense'::character varying])::text[]))
);
-- Create index "idx_chart_of_accounts_organization_id" to table: "chart_of_accounts"
CREATE INDEX "idx_chart_of_accounts_organization_id" ON "public"."chart_of_accounts" ("organization_id");
-- Create index "idx_chart_of_accounts_type" to table: "chart_of_accounts"
CREATE INDEX "idx_chart_of_accounts_type" ON "public"."chart_of_accounts" ("account_type");
-- Create "combo_bundle_items" table
CREATE TABLE "public"."combo_bundle_items" (
  "id" serial NOT NULL,
  "combo_bundle_id" integer NOT NULL,
  "menu_item_id" integer NULL,
  "product_id" integer NULL,
  "product_variant_id" integer NULL,
  "item_type" character varying(20) NULL DEFAULT 'menu_item',
  "quantity" numeric(15,3) NULL DEFAULT 1,
  "is_required" boolean NULL DEFAULT true,
  "group_tag" character varying(50) NULL,
  "price_override" numeric(15,2) NULL,
  "display_order" integer NULL DEFAULT 0,
  "metadata" jsonb NULL DEFAULT '{}',
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "combo_bundle_items_combo_bundle_id_fkey" FOREIGN KEY ("combo_bundle_id") REFERENCES "public"."combo_bundles" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "combo_bundle_items_menu_item_id_fkey" FOREIGN KEY ("menu_item_id") REFERENCES "public"."menu_items" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "combo_bundle_items_product_id_fkey" FOREIGN KEY ("product_id") REFERENCES "public"."products" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "combo_bundle_items_product_variant_id_fkey" FOREIGN KEY ("product_variant_id") REFERENCES "public"."product_variants" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "combo_bundle_items_item_type_check" CHECK ((item_type)::text = ANY ((ARRAY['menu_item'::character varying, 'product'::character varying])::text[]))
);
-- Create index "idx_combo_bundle_items_bundle_id" to table: "combo_bundle_items"
CREATE INDEX "idx_combo_bundle_items_bundle_id" ON "public"."combo_bundle_items" ("combo_bundle_id");
-- Create "cost_centers" table
CREATE TABLE "public"."cost_centers" (
  "id" serial NOT NULL,
  "organization_id" integer NOT NULL,
  "code" character varying(50) NOT NULL,
  "name" character varying(100) NOT NULL,
  "dimension" character varying(50) NULL DEFAULT 'general',
  "is_active" boolean NULL DEFAULT true,
  PRIMARY KEY ("id"),
  CONSTRAINT "cost_centers_code_key" UNIQUE ("code"),
  CONSTRAINT "cost_centers_organization_id_fkey" FOREIGN KEY ("organization_id") REFERENCES "public"."organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create "exchange_rates" table
CREATE TABLE "public"."exchange_rates" (
  "id" serial NOT NULL,
  "organization_id" integer NOT NULL,
  "from_currency" character varying(3) NOT NULL,
  "to_currency" character varying(3) NOT NULL,
  "rate_date" date NOT NULL,
  "rate" numeric(18,6) NOT NULL,
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "exchange_rates_organization_id_from_currency_to_currency_ra_key" UNIQUE ("organization_id", "from_currency", "to_currency", "rate_date"),
  CONSTRAINT "exchange_rates_from_currency_fkey" FOREIGN KEY ("from_currency") REFERENCES "public"."currencies" ("code") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "exchange_rates_organization_id_fkey" FOREIGN KEY ("organization_id") REFERENCES "public"."organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "exchange_rates_to_currency_fkey" FOREIGN KEY ("to_currency") REFERENCES "public"."currencies" ("code") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "gl_account_mappings" table
CREATE TABLE "public"."gl_account_mappings" (
  "id" serial NOT NULL,
  "organization_id" integer NOT NULL,
  "mapping_type" character varying(50) NOT NULL,
  "store_id" integer NULL,
  "gl_account_id" integer NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "gl_account_mappings_organization_id_mapping_type_store_id_key" UNIQUE ("organization_id", "mapping_type", "store_id"),
  CONSTRAINT "gl_account_mappings_gl_account_id_fkey" FOREIGN KEY ("gl_account_id") REFERENCES "public"."chart_of_accounts" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "gl_account_mappings_organization_id_fkey" FOREIGN KEY ("organization_id") REFERENCES "public"."organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "gl_account_mappings_store_id_fkey" FOREIGN KEY ("store_id") REFERENCES "public"."stores" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_gl_account_mappings_org_type" to table: "gl_account_mappings"
CREATE INDEX "idx_gl_account_mappings_org_type" ON "public"."gl_account_mappings" ("organization_id", "mapping_type");
-- Create "invoice_status_history" table
CREATE TABLE "public"."invoice_status_history" (
  "id" bigserial NOT NULL,
  "invoice_id" uuid NOT NULL,
  "organization_id" integer NOT NULL,
  "from_status" "public"."invoice_status" NULL,
  "to_status" "public"."invoice_status" NOT NULL,
  "reason" character varying(255) NULL,
  "notes" text NULL,
  "changed_by_user_id" integer NULL,
  "changed_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "invoice_status_history_changed_by_user_id_fkey" FOREIGN KEY ("changed_by_user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "invoice_status_history_invoice_id_fkey" FOREIGN KEY ("invoice_id") REFERENCES "public"."invoices" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "invoice_status_history_organization_id_fkey" FOREIGN KEY ("organization_id") REFERENCES "public"."organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_invoice_status_history_changed_at" to table: "invoice_status_history"
CREATE INDEX "idx_invoice_status_history_changed_at" ON "public"."invoice_status_history" ("changed_at");
-- Create index "idx_invoice_status_history_invoice_id" to table: "invoice_status_history"
CREATE INDEX "idx_invoice_status_history_invoice_id" ON "public"."invoice_status_history" ("invoice_id");
-- Create "journal_entries" table
CREATE TABLE "public"."journal_entries" (
  "id" bigserial NOT NULL,
  "organization_id" integer NOT NULL,
  "entry_number" character varying(50) NOT NULL,
  "posting_date" date NOT NULL DEFAULT CURRENT_DATE,
  "reference_type" character varying(50) NOT NULL,
  "reference_id" character varying(100) NOT NULL,
  "memo" text NULL,
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "journal_entries_entry_number_key" UNIQUE ("entry_number"),
  CONSTRAINT "journal_entries_organization_id_fkey" FOREIGN KEY ("organization_id") REFERENCES "public"."organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_journal_entries_organization_id" to table: "journal_entries"
CREATE INDEX "idx_journal_entries_organization_id" ON "public"."journal_entries" ("organization_id");
-- Create index "idx_journal_entries_posting_date" to table: "journal_entries"
CREATE INDEX "idx_journal_entries_posting_date" ON "public"."journal_entries" ("posting_date");
-- Create index "idx_journal_entries_reference" to table: "journal_entries"
CREATE INDEX "idx_journal_entries_reference" ON "public"."journal_entries" ("reference_type", "reference_id");
-- Create "journal_lines" table
CREATE TABLE "public"."journal_lines" (
  "id" bigserial NOT NULL,
  "journal_id" bigint NOT NULL,
  "account_id" integer NOT NULL,
  "cost_center_id" integer NULL,
  "debit" numeric(15,2) NOT NULL DEFAULT 0.00,
  "credit" numeric(15,2) NOT NULL DEFAULT 0.00,
  "memo" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "journal_lines_account_id_fkey" FOREIGN KEY ("account_id") REFERENCES "public"."chart_of_accounts" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT,
  CONSTRAINT "journal_lines_cost_center_id_fkey" FOREIGN KEY ("cost_center_id") REFERENCES "public"."cost_centers" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "journal_lines_journal_id_fkey" FOREIGN KEY ("journal_id") REFERENCES "public"."journal_entries" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_journal_lines_account_id" to table: "journal_lines"
CREATE INDEX "idx_journal_lines_account_id" ON "public"."journal_lines" ("account_id");
-- Create index "idx_journal_lines_journal_id" to table: "journal_lines"
CREATE INDEX "idx_journal_lines_journal_id" ON "public"."journal_lines" ("journal_id");
-- Create "kiosk_sessions" table
CREATE TABLE "public"."kiosk_sessions" (
  "id" serial NOT NULL,
  "pos_terminal_id" integer NOT NULL,
  "store_id" integer NOT NULL,
  "session_token" character varying(255) NOT NULL,
  "status" character varying(20) NULL DEFAULT 'active',
  "opened_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  "closed_at" timestamp NULL,
  "metadata" jsonb NULL DEFAULT '{}',
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "kiosk_sessions_session_token_key" UNIQUE ("session_token"),
  CONSTRAINT "kiosk_sessions_pos_terminal_id_fkey" FOREIGN KEY ("pos_terminal_id") REFERENCES "public"."pos_terminals" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "kiosk_sessions_store_id_fkey" FOREIGN KEY ("store_id") REFERENCES "public"."stores" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_kiosk_sessions_status" to table: "kiosk_sessions"
CREATE INDEX "idx_kiosk_sessions_status" ON "public"."kiosk_sessions" ("status");
-- Create index "idx_kiosk_sessions_store_id" to table: "kiosk_sessions"
CREATE INDEX "idx_kiosk_sessions_store_id" ON "public"."kiosk_sessions" ("store_id");
-- Create index "idx_kiosk_sessions_terminal_id" to table: "kiosk_sessions"
CREATE INDEX "idx_kiosk_sessions_terminal_id" ON "public"."kiosk_sessions" ("pos_terminal_id");
-- Create index "idx_kiosk_sessions_token" to table: "kiosk_sessions"
CREATE INDEX "idx_kiosk_sessions_token" ON "public"."kiosk_sessions" ("session_token");
-- Create "menu_item_availability_schedules" table
CREATE TABLE "public"."menu_item_availability_schedules" (
  "id" serial NOT NULL,
  "menu_item_id" integer NOT NULL,
  "day_of_week" integer NULL,
  "start_time" time NOT NULL,
  "end_time" time NOT NULL,
  "is_active" boolean NULL DEFAULT true,
  "valid_from" date NULL,
  "valid_to" date NULL,
  "metadata" jsonb NULL DEFAULT '{}',
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "menu_item_availability_schedules_menu_item_id_fkey" FOREIGN KEY ("menu_item_id") REFERENCES "public"."menu_items" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "chk_schedule_times" CHECK (end_time > start_time),
  CONSTRAINT "menu_item_availability_schedules_day_of_week_check" CHECK ((day_of_week >= 0) AND (day_of_week <= 6))
);
-- Create "permissions" table
CREATE TABLE "public"."permissions" (
  "id" serial NOT NULL,
  "name" character varying(100) NOT NULL,
  "code" character varying(50) NOT NULL,
  "description" text NULL,
  "metadata" jsonb NULL DEFAULT '{}',
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "permissions_code_key" UNIQUE ("code")
);
-- Create index "idx_permissions_code" to table: "permissions"
CREATE INDEX "idx_permissions_code" ON "public"."permissions" ("code");
-- Create "menu_permissions" table
CREATE TABLE "public"."menu_permissions" (
  "id" serial NOT NULL,
  "menu_id" integer NOT NULL,
  "permission_id" integer NOT NULL,
  "metadata" jsonb NULL DEFAULT '{}',
  PRIMARY KEY ("id"),
  CONSTRAINT "menu_permissions_menu_id_permission_id_key" UNIQUE ("menu_id", "permission_id"),
  CONSTRAINT "menu_permissions_menu_id_fkey" FOREIGN KEY ("menu_id") REFERENCES "public"."menus" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "menu_permissions_permission_id_fkey" FOREIGN KEY ("permission_id") REFERENCES "public"."permissions" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create "module_permissions" table
CREATE TABLE "public"."module_permissions" (
  "id" serial NOT NULL,
  "module_id" integer NOT NULL,
  "permission_id" integer NOT NULL,
  "metadata" jsonb NULL DEFAULT '{}',
  PRIMARY KEY ("id"),
  CONSTRAINT "module_permissions_module_id_permission_id_key" UNIQUE ("module_id", "permission_id"),
  CONSTRAINT "module_permissions_module_id_fkey" FOREIGN KEY ("module_id") REFERENCES "public"."modules" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "module_permissions_permission_id_fkey" FOREIGN KEY ("permission_id") REFERENCES "public"."permissions" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create "order_fulfillment_items" table
CREATE TABLE "public"."order_fulfillment_items" (
  "id" uuid NOT NULL DEFAULT public.uuid_generate_v4(),
  "fulfillment_id" uuid NOT NULL,
  "order_line_id" uuid NOT NULL,
  "organization_id" integer NOT NULL,
  "quantity_fulfilled" numeric(15,3) NOT NULL,
  "batch_number" character varying(100) NULL,
  "serial_numbers" text[] NULL,
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "order_fulfillment_items_fulfillment_id_fkey" FOREIGN KEY ("fulfillment_id") REFERENCES "public"."order_fulfillments" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "order_fulfillment_items_order_line_id_fkey" FOREIGN KEY ("order_line_id") REFERENCES "public"."sales_order_lines_v2" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "order_fulfillment_items_organization_id_fkey" FOREIGN KEY ("organization_id") REFERENCES "public"."organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "order_fulfillment_items_quantity_fulfilled_check" CHECK (quantity_fulfilled > (0)::numeric)
);
-- Create index "idx_order_fulfillment_items_fulfillment_id" to table: "order_fulfillment_items"
CREATE INDEX "idx_order_fulfillment_items_fulfillment_id" ON "public"."order_fulfillment_items" ("fulfillment_id");
-- Create index "idx_order_fulfillment_items_order_line_id" to table: "order_fulfillment_items"
CREATE INDEX "idx_order_fulfillment_items_order_line_id" ON "public"."order_fulfillment_items" ("order_line_id");
-- Create "order_status_history" table
CREATE TABLE "public"."order_status_history" (
  "id" bigserial NOT NULL,
  "sales_order_id" uuid NOT NULL,
  "organization_id" integer NOT NULL,
  "from_status" "public"."order_status_v2" NULL,
  "to_status" "public"."order_status_v2" NOT NULL,
  "reason" character varying(255) NULL,
  "notes" text NULL,
  "changed_by_user_id" integer NULL,
  "changed_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "order_status_history_changed_by_user_id_fkey" FOREIGN KEY ("changed_by_user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "order_status_history_organization_id_fkey" FOREIGN KEY ("organization_id") REFERENCES "public"."organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "order_status_history_sales_order_id_fkey" FOREIGN KEY ("sales_order_id") REFERENCES "public"."sales_orders_v2" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_order_status_history_changed_at" to table: "order_status_history"
CREATE INDEX "idx_order_status_history_changed_at" ON "public"."order_status_history" ("changed_at");
-- Create index "idx_order_status_history_sales_order_id" to table: "order_status_history"
CREATE INDEX "idx_order_status_history_sales_order_id" ON "public"."order_status_history" ("sales_order_id");
-- Create "partner_addresses" table
CREATE TABLE "public"."partner_addresses" (
  "id" serial NOT NULL,
  "partner_id" integer NOT NULL,
  "address_name" character varying(100) NOT NULL,
  "address_type" character varying(20) NOT NULL,
  "street" text NULL,
  "city" character varying(100) NULL,
  "state" character varying(100) NULL,
  "zip_code" character varying(20) NULL,
  "country_code" character varying(3) NULL DEFAULT 'SA',
  "is_default" boolean NULL DEFAULT false,
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "partner_addresses_partner_id_fkey" FOREIGN KEY ("partner_id") REFERENCES "public"."business_partners" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "partner_addresses_address_type_check" CHECK ((address_type)::text = ANY ((ARRAY['bill_to'::character varying, 'ship_to'::character varying, 'both'::character varying])::text[]))
);
-- Create index "idx_partner_addresses_partner_id" to table: "partner_addresses"
CREATE INDEX "idx_partner_addresses_partner_id" ON "public"."partner_addresses" ("partner_id");
-- Create "partner_contacts" table
CREATE TABLE "public"."partner_contacts" (
  "id" serial NOT NULL,
  "partner_id" integer NOT NULL,
  "first_name" character varying(100) NOT NULL,
  "last_name" character varying(100) NULL,
  "email" character varying(255) NULL,
  "phone" character varying(50) NULL,
  "position" character varying(100) NULL,
  "is_primary" boolean NULL DEFAULT false,
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "partner_contacts_partner_id_fkey" FOREIGN KEY ("partner_id") REFERENCES "public"."business_partners" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_partner_contacts_partner_id" to table: "partner_contacts"
CREATE INDEX "idx_partner_contacts_partner_id" ON "public"."partner_contacts" ("partner_id");
-- Create "role_permissions" table
CREATE TABLE "public"."role_permissions" (
  "id" serial NOT NULL,
  "role_id" integer NOT NULL,
  "permission_id" integer NOT NULL,
  "scope" character varying(50) NULL DEFAULT 'all',
  "metadata" jsonb NULL DEFAULT '{}',
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "role_permissions_role_id_permission_id_key" UNIQUE ("role_id", "permission_id"),
  CONSTRAINT "role_permissions_permission_id_fkey" FOREIGN KEY ("permission_id") REFERENCES "public"."permissions" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "role_permissions_role_id_fkey" FOREIGN KEY ("role_id") REFERENCES "public"."roles" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_role_permissions_permission_id" to table: "role_permissions"
CREATE INDEX "idx_role_permissions_permission_id" ON "public"."role_permissions" ("permission_id");
-- Create index "idx_role_permissions_role_id" to table: "role_permissions"
CREATE INDEX "idx_role_permissions_role_id" ON "public"."role_permissions" ("role_id");
-- Create "sales_order_lines" table
CREATE TABLE "public"."sales_order_lines" (
  "id" serial NOT NULL,
  "sales_order_id" integer NOT NULL,
  "product_id" integer NOT NULL,
  "product_variant_id" integer NULL,
  "quantity" numeric(15,3) NOT NULL,
  "uom_id" integer NULL,
  "unit_price" numeric(15,4) NOT NULL,
  "discount_amount" numeric(15,2) NULL DEFAULT 0,
  "tax_amount" numeric(15,2) NULL DEFAULT 0,
  "subtotal" numeric(15,2) NOT NULL,
  "line_total" numeric(15,2) NULL DEFAULT 0,
  "shipped_quantity" numeric(15,3) NULL DEFAULT 0,
  "line_number" integer NULL,
  "metadata" jsonb NULL DEFAULT '{}',
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "sales_order_lines_product_id_fkey" FOREIGN KEY ("product_id") REFERENCES "public"."products" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "sales_order_lines_product_variant_id_fkey" FOREIGN KEY ("product_variant_id") REFERENCES "public"."product_variants" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "sales_order_lines_sales_order_id_fkey" FOREIGN KEY ("sales_order_id") REFERENCES "public"."sales_orders" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "sales_order_lines_uom_id_fkey" FOREIGN KEY ("uom_id") REFERENCES "public"."units_of_measure" ("id") ON UPDATE NO ACTION ON DELETE SET NULL
);
-- Create index "idx_sales_order_lines_product_id" to table: "sales_order_lines"
CREATE INDEX "idx_sales_order_lines_product_id" ON "public"."sales_order_lines" ("product_id");
-- Create index "idx_sales_order_lines_sales_order_id" to table: "sales_order_lines"
CREATE INDEX "idx_sales_order_lines_sales_order_id" ON "public"."sales_order_lines" ("sales_order_id");
-- Create "sales_return_lines" table
CREATE TABLE "public"."sales_return_lines" (
  "id" serial NOT NULL,
  "return_id" integer NOT NULL,
  "product_id" integer NOT NULL,
  "product_variant_id" integer NULL,
  "original_line_id" integer NULL,
  "quantity" numeric(15,3) NOT NULL,
  "unit_price" numeric(15,4) NOT NULL,
  "refund_amount" numeric(15,2) NOT NULL,
  "return_to_stock" boolean NULL DEFAULT true,
  "serial_number" character varying(100) NULL,
  "batch_number" character varying(100) NULL,
  "condition" character varying(50) NULL DEFAULT 'good',
  "line_number" integer NULL,
  "metadata" jsonb NULL DEFAULT '{}',
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "sales_return_lines_original_line_id_fkey" FOREIGN KEY ("original_line_id") REFERENCES "public"."pos_transaction_lines" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "sales_return_lines_product_id_fkey" FOREIGN KEY ("product_id") REFERENCES "public"."products" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "sales_return_lines_product_variant_id_fkey" FOREIGN KEY ("product_variant_id") REFERENCES "public"."product_variants" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "sales_return_lines_return_id_fkey" FOREIGN KEY ("return_id") REFERENCES "public"."sales_returns" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "sales_return_lines_condition_check" CHECK ((condition)::text = ANY ((ARRAY['good'::character varying, 'damaged'::character varying, 'defective'::character varying, 'opened'::character varying])::text[]))
);
-- Create index "idx_sales_return_lines_return_id" to table: "sales_return_lines"
CREATE INDEX "idx_sales_return_lines_return_id" ON "public"."sales_return_lines" ("return_id");
-- Create "submenu_permissions" table
CREATE TABLE "public"."submenu_permissions" (
  "id" serial NOT NULL,
  "submenu_id" integer NOT NULL,
  "permission_id" integer NOT NULL,
  "metadata" jsonb NULL DEFAULT '{}',
  PRIMARY KEY ("id"),
  CONSTRAINT "submenu_permissions_submenu_id_permission_id_key" UNIQUE ("submenu_id", "permission_id"),
  CONSTRAINT "submenu_permissions_permission_id_fkey" FOREIGN KEY ("permission_id") REFERENCES "public"."permissions" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "submenu_permissions_submenu_id_fkey" FOREIGN KEY ("submenu_id") REFERENCES "public"."submenus" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create "sync_watermarks" table
CREATE TABLE "public"."sync_watermarks" (
  "id" serial NOT NULL,
  "entity_type" character varying(50) NOT NULL,
  "store_id" integer NULL,
  "last_sync_at" timestamptz NOT NULL DEFAULT '1970-01-01 00:00:00+05',
  "metadata" jsonb NULL DEFAULT '{}',
  PRIMARY KEY ("id"),
  CONSTRAINT "sync_watermarks_entity_type_store_id_key" UNIQUE ("entity_type", "store_id"),
  CONSTRAINT "sync_watermarks_store_id_fkey" FOREIGN KEY ("store_id") REFERENCES "public"."stores" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create "uom_packaging_templates" table
CREATE TABLE "public"."uom_packaging_templates" (
  "id" serial NOT NULL,
  "organization_id" integer NOT NULL,
  "uom_id" integer NULL,
  "name" character varying(255) NOT NULL,
  "code" character varying(50) NOT NULL,
  "is_active" boolean NULL DEFAULT true,
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "uom_packaging_templates_code_key" UNIQUE ("code"),
  CONSTRAINT "uom_packaging_templates_organization_id_fkey" FOREIGN KEY ("organization_id") REFERENCES "public"."organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "uom_packaging_templates_uom_id_fkey" FOREIGN KEY ("uom_id") REFERENCES "public"."units_of_measure" ("id") ON UPDATE NO ACTION ON DELETE SET NULL
);
-- Create index "idx_uom_packaging_templates_code" to table: "uom_packaging_templates"
CREATE INDEX "idx_uom_packaging_templates_code" ON "public"."uom_packaging_templates" ("code");
-- Create index "idx_uom_packaging_templates_organization_id" to table: "uom_packaging_templates"
CREATE INDEX "idx_uom_packaging_templates_organization_id" ON "public"."uom_packaging_templates" ("organization_id");
-- Create "uom_packaging_template_levels" table
CREATE TABLE "public"."uom_packaging_template_levels" (
  "id" serial NOT NULL,
  "template_id" integer NOT NULL,
  "level_order" integer NOT NULL,
  "uom_id" integer NOT NULL,
  "multiplier" numeric(15,6) NOT NULL DEFAULT 1,
  PRIMARY KEY ("id"),
  CONSTRAINT "uom_packaging_template_levels_template_id_level_order_key" UNIQUE ("template_id", "level_order"),
  CONSTRAINT "uom_packaging_template_levels_template_id_fkey" FOREIGN KEY ("template_id") REFERENCES "public"."uom_packaging_templates" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "uom_packaging_template_levels_uom_id_fkey" FOREIGN KEY ("uom_id") REFERENCES "public"."units_of_measure" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_uom_pkg_template_levels_template_id" to table: "uom_packaging_template_levels"
CREATE INDEX "idx_uom_pkg_template_levels_template_id" ON "public"."uom_packaging_template_levels" ("template_id");
-- Create index "idx_uom_pkg_template_levels_uom_id" to table: "uom_packaging_template_levels"
CREATE INDEX "idx_uom_pkg_template_levels_uom_id" ON "public"."uom_packaging_template_levels" ("uom_id");
-- Create "user_roles" table
CREATE TABLE "public"."user_roles" (
  "id" serial NOT NULL,
  "user_id" integer NOT NULL,
  "role_id" integer NOT NULL,
  "metadata" jsonb NULL DEFAULT '{}',
  "assigned_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "user_roles_user_id_role_id_key" UNIQUE ("user_id", "role_id"),
  CONSTRAINT "user_roles_role_id_fkey" FOREIGN KEY ("role_id") REFERENCES "public"."roles" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "user_roles_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_user_roles_role_id" to table: "user_roles"
CREATE INDEX "idx_user_roles_role_id" ON "public"."user_roles" ("role_id");
-- Create index "idx_user_roles_user_id" to table: "user_roles"
CREATE INDEX "idx_user_roles_user_id" ON "public"."user_roles" ("user_id");
-- Create "user_store_access" table
CREATE TABLE "public"."user_store_access" (
  "id" serial NOT NULL,
  "user_id" integer NOT NULL,
  "store_id" integer NOT NULL,
  "is_primary" boolean NULL DEFAULT false,
  "metadata" jsonb NULL DEFAULT '{}',
  "granted_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "user_store_access_user_id_store_id_key" UNIQUE ("user_id", "store_id"),
  CONSTRAINT "user_store_access_store_id_fkey" FOREIGN KEY ("store_id") REFERENCES "public"."stores" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "user_store_access_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_user_store_access_store_id" to table: "user_store_access"
CREATE INDEX "idx_user_store_access_store_id" ON "public"."user_store_access" ("store_id");
-- Create index "idx_user_store_access_user_id" to table: "user_store_access"
CREATE INDEX "idx_user_store_access_user_id" ON "public"."user_store_access" ("user_id");
-- Create "zatca_device_configs" table
CREATE TABLE "public"."zatca_device_configs" (
  "id" serial NOT NULL,
  "organization_id" integer NOT NULL,
  "store_id" integer NULL,
  "pos_terminal_id" integer NULL,
  "device_serial" character varying(255) NOT NULL,
  "device_type" character varying(20) NOT NULL,
  "csr_pem" text NULL,
  "private_key_pem" text NOT NULL,
  "compliance_csid" text NULL,
  "production_csid" text NULL,
  "csid_expiry" timestamptz NULL,
  "zatca_env" character varying(20) NULL DEFAULT 'sandbox',
  "is_active" boolean NULL DEFAULT true,
  "is_revoked" boolean NULL DEFAULT false,
  "revoked_at" timestamptz NULL,
  "revoked_reason" text NULL,
  "metadata" jsonb NULL DEFAULT '{}',
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "zatca_device_configs_organization_id_device_serial_key" UNIQUE ("organization_id", "device_serial"),
  CONSTRAINT "zatca_device_configs_organization_id_fkey" FOREIGN KEY ("organization_id") REFERENCES "public"."organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "zatca_device_configs_pos_terminal_id_fkey" FOREIGN KEY ("pos_terminal_id") REFERENCES "public"."pos_terminals" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "zatca_device_configs_store_id_fkey" FOREIGN KEY ("store_id") REFERENCES "public"."stores" ("id") ON UPDATE NO ACTION ON DELETE SET NULL
);
-- Create "zatca_document_chain" table
CREATE TABLE "public"."zatca_document_chain" (
  "id" bigserial NOT NULL,
  "entity_type" character varying(20) NOT NULL,
  "entity_id" text NOT NULL,
  "device_config_id" integer NOT NULL,
  "organization_id" integer NOT NULL,
  "zatca_uuid" uuid NOT NULL DEFAULT public.uuid_generate_v4(),
  "icv" bigint NOT NULL,
  "pih" text NOT NULL,
  "xml_hash" text NOT NULL,
  "zatca_status" "public"."zatca_doc_status" NULL DEFAULT 'pending',
  "zatca_response" jsonb NULL DEFAULT '{}',
  "qr_code_base64" text NULL,
  "signed_xml" text NULL,
  "submitted_at" timestamptz NULL,
  "cleared_at" timestamptz NULL,
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "zatca_document_chain_device_config_id_icv_key" UNIQUE ("device_config_id", "icv"),
  CONSTRAINT "zatca_document_chain_device_config_id_fkey" FOREIGN KEY ("device_config_id") REFERENCES "public"."zatca_device_configs" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "zatca_document_chain_organization_id_fkey" FOREIGN KEY ("organization_id") REFERENCES "public"."organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_zatca_chain_device_icv" to table: "zatca_document_chain"
CREATE INDEX "idx_zatca_chain_device_icv" ON "public"."zatca_document_chain" ("device_config_id", "icv" DESC);
-- Create index "idx_zatca_chain_entity" to table: "zatca_document_chain"
CREATE INDEX "idx_zatca_chain_entity" ON "public"."zatca_document_chain" ("entity_type", "entity_id");
-- Create index "idx_zatca_chain_status" to table: "zatca_document_chain"
CREATE INDEX "idx_zatca_chain_status" ON "public"."zatca_document_chain" ("zatca_status") WHERE (zatca_status = ANY (ARRAY['pending'::public.zatca_doc_status, 'failed'::public.zatca_doc_status]));
-- Create "vw_accounts_payable" view
CREATE VIEW "public"."vw_accounts_payable" (
  "po_id",
  "po_number",
  "organization_id",
  "organization_name",
  "supplier_id",
  "supplier_name",
  "contact_person",
  "email",
  "supplier_payment_terms",
  "store_name",
  "po_date",
  "expected_delivery_date",
  "status",
  "po_total",
  "discount_amount",
  "tax_amount",
  "received_amount",
  "amount_paid_str",
  "created_at"
) AS SELECT po.id AS po_id,
    po.po_number,
    po.organization_id,
    org.name AS organization_name,
    sup.id AS supplier_id,
    sup.name AS supplier_name,
    sup.contact_person,
    sup.email,
    sup.payment_terms AS supplier_payment_terms,
    s.name AS store_name,
    po.po_date,
    po.expected_delivery_date,
    po.status,
    po.total_amount AS po_total,
    po.discount_amount,
    po.tax_amount,
    ( SELECT COALESCE(sum(pol.received_quantity * pol.unit_price), 0::numeric) AS "coalesce"
           FROM public.purchase_order_lines pol
          WHERE pol.purchase_order_id = po.id) AS received_amount,
    po.metadata ->> 'amount_paid'::text AS amount_paid_str,
    po.created_at
   FROM public.purchase_orders po
     JOIN public.organizations org ON org.id = po.organization_id
     JOIN public.suppliers sup ON sup.id = po.supplier_id
     JOIN public.stores s ON s.id = po.store_id
  WHERE po.status::text = ANY (ARRAY['partially_received'::character varying, 'received'::character varying, 'approved'::character varying]::text[])
  ORDER BY po.po_date;
-- Create "vw_active_restaurant_orders" view
CREATE VIEW "public"."vw_active_restaurant_orders" (
  "order_id",
  "order_number",
  "store_id",
  "order_source",
  "order_status",
  "subtotal",
  "tax_amount",
  "total_amount",
  "notes",
  "ordered_at",
  "confirmed_at",
  "table_id",
  "table_number",
  "table_name",
  "table_section",
  "cashier_id",
  "waiter_name",
  "customer_id",
  "customer_name",
  "minutes_since_ordered"
) AS SELECT ro.id AS order_id,
    ro.order_number,
    ro.store_id,
    ro.order_source,
    ro.status AS order_status,
    ro.subtotal,
    ro.tax_amount,
    ro.total_amount,
    ro.notes,
    ro.ordered_at,
    ro.confirmed_at,
    rt.id AS table_id,
    rt.table_number,
    rt.table_name,
    rt.section AS table_section,
    c.id AS cashier_id,
    (u.first_name::text || ' '::text) || u.last_name::text AS waiter_name,
    ro.customer_id,
    cust.name AS customer_name,
    EXTRACT(epoch FROM CURRENT_TIMESTAMP - ro.ordered_at::timestamp with time zone) / 60.0 AS minutes_since_ordered
   FROM public.restaurant_orders ro
     LEFT JOIN public.restaurant_tables rt ON ro.table_id = rt.id
     LEFT JOIN public.cashiers c ON ro.cashier_id = c.id
     LEFT JOIN public.users u ON c.user_id = u.id
     LEFT JOIN public.customers cust ON ro.customer_id = cust.id
  WHERE ro.status::text <> ALL (ARRAY['paid'::character varying, 'voided'::character varying]::text[]);
-- Create "vw_customer_aging_report" view
CREATE VIEW "public"."vw_customer_aging_report" (
  "customer_id",
  "customer_code",
  "customer_name",
  "email",
  "phone",
  "customer_type",
  "credit_limit",
  "outstanding_balance",
  "organization_id",
  "current_amount",
  "overdue_1_30",
  "overdue_31_60",
  "overdue_61_90",
  "overdue_over_90",
  "total_outstanding",
  "overdue_invoice_count",
  "latest_due_date",
  "loyalty_points"
) AS SELECT c.id AS customer_id,
    c.customer_code,
    c.name AS customer_name,
    c.email,
    c.phone,
    c.customer_type,
    c.credit_limit,
    c.outstanding_balance,
    i.organization_id,
    COALESCE(sum(
        CASE
            WHEN i.due_date >= CURRENT_DATE THEN i.balance_due
            ELSE 0::numeric
        END), 0::numeric) AS current_amount,
    COALESCE(sum(
        CASE
            WHEN i.due_date < CURRENT_DATE AND (CURRENT_DATE - i.due_date) <= 30 THEN i.balance_due
            ELSE 0::numeric
        END), 0::numeric) AS overdue_1_30,
    COALESCE(sum(
        CASE
            WHEN (CURRENT_DATE - i.due_date) >= 31 AND (CURRENT_DATE - i.due_date) <= 60 THEN i.balance_due
            ELSE 0::numeric
        END), 0::numeric) AS overdue_31_60,
    COALESCE(sum(
        CASE
            WHEN (CURRENT_DATE - i.due_date) >= 61 AND (CURRENT_DATE - i.due_date) <= 90 THEN i.balance_due
            ELSE 0::numeric
        END), 0::numeric) AS overdue_61_90,
    COALESCE(sum(
        CASE
            WHEN (CURRENT_DATE - i.due_date) > 90 THEN i.balance_due
            ELSE 0::numeric
        END), 0::numeric) AS overdue_over_90,
    COALESCE(sum(
        CASE
            WHEN i.balance_due > 0::numeric THEN i.balance_due
            ELSE 0::numeric
        END), 0::numeric) AS total_outstanding,
    count(
        CASE
            WHEN i.invoice_status = 'overdue'::public.invoice_status THEN 1
            ELSE NULL::integer
        END)::integer AS overdue_invoice_count,
    max(i.due_date) AS latest_due_date,
    c.loyalty_points
   FROM public.customers c
     LEFT JOIN public.invoices i ON i.customer_id = c.id AND (i.invoice_status <> ALL (ARRAY['cancelled'::public.invoice_status, 'draft'::public.invoice_status])) AND i.balance_due > 0::numeric
     LEFT JOIN public.organizations o ON o.id = i.organization_id
  WHERE c.is_active = true
  GROUP BY c.id, c.customer_code, c.name, c.email, c.phone, c.customer_type, c.credit_limit, c.outstanding_balance, c.loyalty_points, i.organization_id
  ORDER BY (COALESCE(sum(
        CASE
            WHEN i.balance_due > 0::numeric THEN i.balance_due
            ELSE 0::numeric
        END), 0::numeric)) DESC;
-- Create "vw_low_stock_alerts" view
CREATE VIEW "public"."vw_low_stock_alerts" (
  "inventory_stock_id",
  "store_id",
  "store_name",
  "store_code",
  "product_id",
  "sku",
  "effective_sku",
  "product_name",
  "variant_name",
  "product_variant_id",
  "category_name",
  "quantity_on_hand",
  "quantity_available",
  "quantity_allocated",
  "reorder_level",
  "reorder_quantity",
  "stock_status",
  "is_out_of_stock",
  "is_low_stock",
  "last_counted_at"
) AS SELECT ist.id AS inventory_stock_id,
    s.id AS store_id,
    s.name AS store_name,
    s.code AS store_code,
    p.id AS product_id,
    p.sku,
    COALESCE(pv.variant_sku, p.sku) AS effective_sku,
    p.name AS product_name,
    COALESCE(pv.variant_name, ''::character varying) AS variant_name,
    pv.id AS product_variant_id,
    pc.name AS category_name,
    ist.quantity_on_hand,
    ist.quantity_available,
    ist.quantity_allocated,
    ist.reorder_level,
    ist.reorder_quantity,
        CASE
            WHEN ist.quantity_available <= 0::numeric THEN 'out_of_stock'::text
            WHEN ist.quantity_available <= COALESCE(ist.reorder_level, 0::numeric) THEN 'low_stock'::text
            WHEN ist.quantity_available <= (COALESCE(ist.reorder_level, 0::numeric) * 1.5) THEN 'near_reorder'::text
            ELSE 'adequate'::text
        END AS stock_status,
    ist.quantity_available <= 0::numeric AS is_out_of_stock,
    ist.quantity_available > 0::numeric AND ist.quantity_available <= COALESCE(ist.reorder_level, 0::numeric) AS is_low_stock,
    ist.last_counted_at
   FROM public.inventory_stock ist
     JOIN public.stores s ON s.id = ist.store_id AND s.is_active = true
     JOIN public.products p ON p.id = ist.product_id AND p.is_active = true AND p.track_inventory = true
     LEFT JOIN public.product_variants pv ON pv.id = ist.product_variant_id
     LEFT JOIN public.product_categories pc ON pc.id = p.category_id
  WHERE ist.quantity_available <= COALESCE(ist.reorder_level, 0::numeric) OR ist.quantity_available <= 0::numeric
  ORDER BY s.name, (
        CASE
            WHEN ist.quantity_available <= 0::numeric THEN 0
            ELSE 1
        END), p.name;
-- Create "vw_pending_purchase_orders" view
CREATE VIEW "public"."vw_pending_purchase_orders" (
  "po_id",
  "po_number",
  "po_date",
  "expected_delivery_date",
  "status",
  "days_overdue",
  "is_overdue",
  "store_id",
  "store_name",
  "supplier_id",
  "supplier_name",
  "contact_person",
  "supplier_email",
  "subtotal",
  "discount_amount",
  "tax_amount",
  "total_amount",
  "outstanding_quantity",
  "created_by_username",
  "approved_by_username",
  "created_at"
) AS SELECT po.id AS po_id,
    po.po_number,
    po.po_date,
    po.expected_delivery_date,
    po.status,
    CURRENT_DATE - po.expected_delivery_date AS days_overdue,
    po.expected_delivery_date < CURRENT_DATE AND (po.status::text <> ALL (ARRAY['received'::character varying, 'cancelled'::character varying, 'closed'::character varying]::text[])) AS is_overdue,
    s.id AS store_id,
    s.name AS store_name,
    sup.id AS supplier_id,
    sup.name AS supplier_name,
    sup.contact_person,
    sup.email AS supplier_email,
    po.subtotal,
    po.discount_amount,
    po.tax_amount,
    po.total_amount,
    ( SELECT COALESCE(sum(pol.quantity - pol.received_quantity), 0::numeric) AS "coalesce"
           FROM public.purchase_order_lines pol
          WHERE pol.purchase_order_id = po.id) AS outstanding_quantity,
    u_created.username AS created_by_username,
    u_approved.username AS approved_by_username,
    po.created_at
   FROM public.purchase_orders po
     JOIN public.stores s ON s.id = po.store_id
     JOIN public.suppliers sup ON sup.id = po.supplier_id
     LEFT JOIN public.users u_created ON u_created.id = po.created_by
     LEFT JOIN public.users u_approved ON u_approved.id = po.approved_by
  WHERE po.status::text <> ALL (ARRAY['received'::character varying, 'cancelled'::character varying, 'closed'::character varying]::text[])
  ORDER BY (po.expected_delivery_date < CURRENT_DATE AND (po.status::text <> ALL (ARRAY['received'::character varying, 'cancelled'::character varying, 'closed'::character varying]::text[]))) DESC, (CURRENT_DATE - po.expected_delivery_date) DESC NULLS LAST, po.expected_delivery_date;
-- Create "vw_pos_categories" view
CREATE VIEW "public"."vw_pos_categories" (
  "category_id",
  "category_code",
  "category_name",
  "parent_category_id",
  "parent_category_name",
  "product_count",
  "in_stock_count",
  "category_metadata"
) AS SELECT pc.id AS category_id,
    pc.code AS category_code,
    pc.name AS category_name,
    pc.parent_category_id,
    pc_parent.name AS parent_category_name,
    count(DISTINCT p.id)::integer AS product_count,
    count(DISTINCT
        CASE
            WHEN inv.quantity_available > 0::numeric THEN p.id
            ELSE NULL::integer
        END)::integer AS in_stock_count,
    pc.metadata AS category_metadata
   FROM public.product_categories pc
     LEFT JOIN public.product_categories pc_parent ON pc.parent_category_id = pc_parent.id
     LEFT JOIN public.products p ON pc.id = p.category_id
     LEFT JOIN public.inventory_stock inv ON p.id = inv.product_id
  GROUP BY pc.id, pc.code, pc.name, pc.parent_category_id, pc_parent.name, pc.metadata
  ORDER BY pc_parent.name NULLS FIRST, pc.name;
-- Create "vw_profit_margin_analysis" view
CREATE VIEW "public"."vw_profit_margin_analysis" (
  "product_id",
  "product_variant_id",
  "sku",
  "product_name",
  "category_name",
  "store_id",
  "store_name",
  "period_month",
  "units_sold",
  "total_revenue",
  "total_cost",
  "gross_profit",
  "gross_margin_pct",
  "total_discounts",
  "total_taxes"
) AS SELECT ptl.product_id,
    ptl.product_variant_id,
    p.sku,
    p.name AS product_name,
    pc.name AS category_name,
    pt.store_id,
    s.name AS store_name,
    date_trunc('month'::text, pt.transaction_date)::date AS period_month,
    sum(ptl.quantity) AS units_sold,
    sum(ptl.line_total) AS total_revenue,
    sum(ptl.cost_price * ptl.quantity) AS total_cost,
    sum(ptl.line_total) - sum(ptl.cost_price * ptl.quantity) AS gross_profit,
        CASE
            WHEN sum(ptl.line_total) > 0::numeric THEN round((sum(ptl.line_total) - sum(ptl.cost_price * ptl.quantity)) / sum(ptl.line_total) * 100::numeric, 2)
            ELSE 0::numeric
        END AS gross_margin_pct,
    sum(ptl.discount_amount) AS total_discounts,
    sum(ptl.tax_amount) AS total_taxes
   FROM public.pos_transaction_lines ptl
     JOIN public.pos_transactions pt ON pt.id = ptl.transaction_id AND pt.status::text = 'completed'::text
     JOIN public.products p ON p.id = ptl.product_id
     LEFT JOIN public.product_variants pv_id ON pv_id.id = ptl.product_variant_id
     LEFT JOIN public.product_categories pc ON pc.id = p.category_id
     JOIN public.stores s ON s.id = pt.store_id
  GROUP BY ptl.product_id, ptl.product_variant_id, p.sku, p.name, pc.name, pt.store_id, s.name, (date_trunc('month'::text, pt.transaction_date))
  ORDER BY (date_trunc('month'::text, pt.transaction_date)::date) DESC, (sum(ptl.line_total) - sum(ptl.cost_price * ptl.quantity)) DESC;
-- Create "vw_user_effective_permissions" view
CREATE VIEW "public"."vw_user_effective_permissions" (
  "user_id",
  "username",
  "email",
  "organization_id",
  "role_id",
  "role_name",
  "role_code",
  "permission_id",
  "permission_name",
  "permission_code",
  "scope",
  "accessible_store_id"
) AS SELECT DISTINCT u.id AS user_id,
    u.username,
    u.email,
    u.organization_id,
    r.id AS role_id,
    r.name AS role_name,
    r.code AS role_code,
    per.id AS permission_id,
    per.name AS permission_name,
    per.code AS permission_code,
    rp.scope,
    usa.store_id AS accessible_store_id
   FROM public.users u
     JOIN public.user_roles ur ON ur.user_id = u.id
     JOIN public.roles r ON r.id = ur.role_id AND r.is_active = true
     JOIN public.role_permissions rp ON rp.role_id = r.id
     JOIN public.permissions per ON per.id = rp.permission_id
     LEFT JOIN public.user_store_access usa ON usa.user_id = u.id
  WHERE u.is_active = true
  ORDER BY u.username, r.name, per.code;
-- Create "vw_waste_daily_summary" view
CREATE VIEW "public"."vw_waste_daily_summary" (
  "store_id",
  "waste_date",
  "waste_source",
  "waste_entries",
  "total_quantity_wasted",
  "total_cost_wasted",
  "avg_cost_per_entry"
) AS SELECT store_id,
    date(wasted_at) AS waste_date,
    waste_source,
    count(*) AS waste_entries,
    sum(quantity) AS total_quantity_wasted,
    sum(total_cost) AS total_cost_wasted,
    avg(total_cost) AS avg_cost_per_entry
   FROM public.waste_logs wl
  GROUP BY store_id, (date(wasted_at)), waste_source;
