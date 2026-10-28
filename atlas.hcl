# Atlas Configuration for Nembus Monorepo
# Docs: https://atlasgo.io/concepts/config-file

variable "master_db_url" {
  type        = string
  default     = getenv("MASTER_DB_URL") != "" ? getenv("MASTER_DB_URL") : "postgres://root:nastecsol@localhost:5432/masterDB?sslmode=disable"
  description = "Connection URL for the target database"
}

variable "stg_db_url" {
  type        = string
  default     = getenv("STG_DB_URL") != "" ? getenv("STG_DB_URL") : "postgres://nembus_admin_user:your-password-here@localhost:5432/qitaf2?sslmode=disable"
  description = "Connection URL for the staging database"
}

variable "dev_db_url" {
  type        = string
  default     = getenv("ATLAS_DEV_URL") != "" ? getenv("ATLAS_DEV_URL") : "docker://postgres/16/dev"
  description = "Connection URL for Atlas dev database used during calculation and linting"
}

# Environment for Core Database Package
env "local" {
  src = [
    "file://packages/core/db/schema",
    "file://packages/core/db/views_functions"
  ]
  url = var.master_db_url
  dev = var.dev_db_url
  migration {
    dir    = "file://packages/core/db/migrations"
    format = atlas
  }
  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
}

# Environment for Cloud Server
env "cloud" {
  src = [
    "file://packages/core/db/schema",
    "file://packages/core/db/views_functions"
  ]
  url = var.master_db_url
  dev = var.dev_db_url
  migration {
    dir    = "file://apps/cloud-server/migrations"
    format = atlas
  }
}

# Environment for POS Client
env "pos" {
  src = [
    "file://packages/core/db/schema",
    "file://packages/core/db/views_functions"
  ]
  url = var.master_db_url
  dev = var.dev_db_url
  migration {
    dir    = "file://apps/pos-client/migrations"
    format = atlas
  }
}

variable "supabase_db_url" {
  type        = string
  default     = getenv("SUPABASE_DB_URL") != "" ? getenv("SUPABASE_DB_URL") : ""
  description = "Connection URL for Supabase tenant database"
}

env "stg" {
  src = [
    "file://packages/core/db/schema",
    "file://packages/core/db/views_functions"
  ]
  url = var.stg_db_url
  dev = var.dev_db_url
  migration {
    dir    = "file://packages/core/db/migrations"
    format = atlas
  }
}

env "supabase" {
  src = [
    "file://packages/core/db/schema",
    "file://packages/core/db/views_functions"
  ]
  url = var.supabase_db_url
  dev = var.dev_db_url
  migration {
    dir    = "file://packages/core/db/migrations"
    format = atlas
  }
}

variable "aws_rds_db_url" {
  type        = string
  default     = getenv("AWS_RDS_DB_URL") != "" ? getenv("AWS_RDS_DB_URL") : "postgresql://postgres:Nembus-2026@nembus-pg-db.cf86sus6qqzy.us-east-2.rds.amazonaws.com:5432/postgres?sslmode=require"
  description = "Connection URL for AWS RDS PostgreSQL database"
}

env "aws_rds" {
  src = [
    "file://packages/core/db/schema",
    "file://packages/core/db/views_functions"
  ]
  url = var.aws_rds_db_url
  dev = var.dev_db_url
  migration {
    dir    = "file://packages/core/db/migrations"
    format = atlas
  }
}

