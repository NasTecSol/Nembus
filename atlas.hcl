# Atlas Configuration for Nembus
# Docs: https://atlasgo.io/concepts/config-file

variable "master_db_url" {
  type        = string
  default     = getenv("MASTER_DB_URL") != "" ? getenv("MASTER_DB_URL") : "postgres://root:nastecsol@localhost:5432/masterDB?sslmode=disable"
  description = "Connection URL for the master database"
}

variable "stg_db_url" {
  type        = string
  default     = getenv("STG_DB_URL") != "" ? getenv("STG_DB_URL") : "postgres://root:nastecsol@localhost:5432/stg?sslmode=disable"
  description = "Connection URL for the staging database"
}

variable "dev_db_url" {
  type        = string
  default     = getenv("ATLAS_DEV_URL") != "" ? getenv("ATLAS_DEV_URL") : "docker://postgres/16/dev"
  description = "Connection URL for Atlas dev database used during schema calculation and diffing"
}

env "local" {
  src = "file://packages/core/db/schema"
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

env "stg" {
  src = "file://packages/core/db/schema"
  url = var.stg_db_url
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

env "cloud" {
  src = "file://packages/core/db/schema"
  url = var.master_db_url
  dev = var.dev_db_url
  migration {
    dir    = "file://packages/core/db/migrations"
    format = atlas
  }
}
