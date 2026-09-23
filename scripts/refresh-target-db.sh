#!/usr/bin/env bash
# ==============================================================================
# Refresh Migrated SAP Domains from STG DB -> Target DB
# ==============================================================================

set -e

SOURCE_URL="${SOURCE_DB_URL:-${STG_DATABASE_URL:-${DATABASE_URL:-}}}"
TARGET_URL="${TARGET_DB_URL:-${PROD_DATABASE_URL:-}}"
ORG_ID="${ORG_ID:-0}"
DOMAINS="${DOMAINS:-all}"
MODE="${MODE:-truncate_copy}"
DRY_RUN="${DRY_RUN:-false}"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CLOUD_SERVER_DIR="${SCRIPT_DIR}/../apps/cloud-server"

echo "================================================================"
echo "       NEMBUS: Refresh Migrated Domains -> Target Database      "
echo "================================================================"

cd "${CLOUD_SERVER_DIR}"

CMD_ARGS=("-source-url=${SOURCE_URL}" "-target-url=${TARGET_URL}" "-org-id=${ORG_ID}" "-domains=${DOMAINS}" "-mode=${MODE}")

if [ "${DRY_RUN}" = "true" ] || [ "${1}" = "--dry-run" ]; then
    CMD_ARGS+=("-dry-run=true")
fi

go run ./cmd/sync-target-db "${CMD_ARGS[@]}"
