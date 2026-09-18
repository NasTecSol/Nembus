#!/usr/bin/env bash
set -e

PROJECT_ROOT="$(cd "$(dirname "$0")/../../../.." && pwd)"
BRIDGE_DIR="$PROJECT_ROOT/packages/core/bridge"
RUNNER_DIR="$(cd "$(dirname "$0")/../Runner" && pwd)"

SDK_NAME="${SDK_NAME:-iphoneos}"
MIN_VERSION="14.0"

if [[ "$SDK_NAME" == *"simulator"* ]]; then
    SDK_PATH=$(xcrun --sdk iphonesimulator --show-sdk-path)
    CLANG=$(xcrun --sdk iphonesimulator --find clang)
    CGO_FLAGS="-isysroot $SDK_PATH -miphonesimulator-version-min=$MIN_VERSION -arch arm64"
    echo "[GoBuild] Compiling for iOS Simulator (arm64)..."
else
    SDK_PATH=$(xcrun --sdk iphoneos --show-sdk-path)
    CLANG=$(xcrun --sdk iphoneos --find clang)
    CGO_FLAGS="-isysroot $SDK_PATH -miphoneos-version-min=$MIN_VERSION -arch arm64"
    echo "[GoBuild] Compiling for iOS Device (arm64)..."
fi

CGO_ENABLED=1 \
GOOS=ios \
GOARCH=arm64 \
CC="$CLANG" \
CGO_CFLAGS="$CGO_FLAGS" \
CGO_LDFLAGS="$CGO_FLAGS" \
go build -tags netgo -buildmode=c-archive -ldflags="-s -w" \
-o "$RUNNER_DIR/libnembus_core.a" "$BRIDGE_DIR"

echo "[GoBuild] iOS Static Archive generated: $RUNNER_DIR/libnembus_core.a"