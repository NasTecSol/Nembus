#!/usr/bin/env bash
# ==============================================================================
# Script: execute_fingerprint.sh
# Purpose: Build (if needed) and run the DigitalPersona fingerprint microservice
# ==============================================================================

set -euo pipefail

# Determine script directory dynamically
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo "[INFO] DigitalPersona Microservice directory: $SCRIPT_DIR"

# Build binary if not already compiled
if [[ ! -x "./finger" ]]; then
    echo "[INFO] Binary 'finger' not found. Compiling via Makefile..."
    make
fi

# Execute microservice
echo "[INFO] Launching fingerprint server..."
exec ./finger
