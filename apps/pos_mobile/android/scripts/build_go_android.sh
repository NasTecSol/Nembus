#!/usr/bin/env bash
set -e

PROJECT_ROOT="$(cd "$(dirname "$0")/../../../.." && pwd)"
BRIDGE_DIR="$PROJECT_ROOT/packages/core/bridge"
OUTPUT_BASE="$(cd "$(dirname "$0")/../app/src/main/jniLibs" && pwd)"

if [ -z "$ANDROID_NDK_HOME" ] && [ -n "$ANDROID_SDK_ROOT" ]; then
    ANDROID_NDK_HOME=$(find "$ANDROID_SDK_ROOT/ndk" -maxdepth 1 -mindepth 1 | sort -V | tail -n 1)
fi

if [ -z "$ANDROID_NDK_HOME" ]; then
    echo "[GoBuild] WARNING: ANDROID_NDK_HOME not found. Skipping native Android Go build."
    exit 0
fi

HOST_OS="darwin-x86_64"
if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    HOST_OS="linux-x86_64"
fi

TOOLCHAIN="$ANDROID_NDK_HOME/toolchains/llvm/prebuilt/$HOST_OS/bin"
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