#!/usr/bin/env bash
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../../.." && pwd)"
BRIDGE_DIR="$PROJECT_ROOT/packages/core/bridge"
OUTPUT_BASE="$SCRIPT_DIR/../app/src/main/jniLibs"
mkdir -p "$OUTPUT_BASE"

# Resolve SDK root if not set
if [ -z "$ANDROID_SDK_ROOT" ]; then
    if [ -n "$ANDROID_HOME" ]; then
        ANDROID_SDK_ROOT="$ANDROID_HOME"
    elif [ -f "$SCRIPT_DIR/../local.properties" ]; then
        SDK_FROM_PROP=$(grep '^sdk.dir=' "$SCRIPT_DIR/../local.properties" | cut -d'=' -f2-)
        if [ -n "$SDK_FROM_PROP" ]; then
            ANDROID_SDK_ROOT="$SDK_FROM_PROP"
        fi
    fi
fi

# Resolve NDK home if not set
if [ -z "$ANDROID_NDK_HOME" ] && [ -n "$ANDROID_SDK_ROOT" ] && [ -d "$ANDROID_SDK_ROOT/ndk" ]; then
    ANDROID_NDK_HOME=$(find "$ANDROID_SDK_ROOT/ndk" -maxdepth 1 -mindepth 1 2>/dev/null | sort -V | tail -n 1)
fi

if [ -z "$ANDROID_NDK_HOME" ] || [ ! -d "$ANDROID_NDK_HOME" ]; then
    echo "[GoBuild] WARNING: ANDROID_NDK_HOME not found. Skipping native Android Go build."
    exit 0
fi

if ! command -v go &> /dev/null; then
    echo "[GoBuild] WARNING: 'go' command not found in PATH. Skipping native Android Go build."
    exit 0
fi

HOST_OS="darwin-x86_64"
if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    HOST_OS="linux-x86_64"
fi

TOOLCHAIN="$ANDROID_NDK_HOME/toolchains/llvm/prebuilt/$HOST_OS/bin"
if [ ! -d "$TOOLCHAIN" ]; then
    PREBUILT_DIR=$(find "$ANDROID_NDK_HOME/toolchains/llvm/prebuilt" -maxdepth 1 -mindepth 1 2>/dev/null | head -n 1)
    if [ -n "$PREBUILT_DIR" ] && [ -d "$PREBUILT_DIR/bin" ]; then
        TOOLCHAIN="$PREBUILT_DIR/bin"
    fi
fi

if [ ! -d "$TOOLCHAIN" ]; then
    echo "[GoBuild] WARNING: Toolchain not found at $TOOLCHAIN. Skipping native Android Go build."
    exit 0
fi

API_LEVEL=21

build_abi() {
    local ABI=$1
    local GOARCH=$2
    local GOARM=$3
    local CLANG_TARGET=$4
    local OUT_DIR="$OUTPUT_BASE/$ABI"
    mkdir -p "$OUT_DIR"

    echo "[GoBuild] Building Android Go lib for $ABI..."
    CC="$TOOLCHAIN/${CLANG_TARGET}${API_LEVEL}-clang" \
    CGO_ENABLED=1 \
    GOOS=android \
    GOARCH=$GOARCH \
    GOARM=$GOARM \
    go build -buildmode=c-shared -ldflags="-s -w" \
    -o "$OUT_DIR/libnembus_core.so" "$BRIDGE_DIR"
}

build_abi "arm64-v8a" "arm64" "" "aarch64-linux-android"
build_abi "armeabi-v7a" "arm" "7" "armv7a-linux-androideabi"
build_abi "x86_64" "amd64" "" "x86_64-linux-android"

echo "[GoBuild] Android .so libraries built successfully."