#!/bin/bash
set -e

echo "Building C-Archive framework for iOS..."

# Create temp build dir
BUILD_DIR="build/ios_cgo"
rm -rf "$BUILD_DIR"
mkdir -p "$BUILD_DIR/sim_arm64" "$BUILD_DIR/sim_x86" "$BUILD_DIR/device_arm64"

SIM_SDK=$(xcrun -sdk iphonesimulator --show-sdk-path)
DEV_SDK=$(xcrun -sdk iphoneos --show-sdk-path)
CLANG=$(xcrun -sdk iphoneos -find clang)

# 1. iOS Simulator arm64
CGO_ENABLED=1 GOOS=ios GOARCH=arm64 \
CC="$CLANG" \
CGO_CFLAGS="-isysroot $SIM_SDK -miphonesimulator-version-min=13.0 -arch arm64" \
CGO_LDFLAGS="-isysroot $SIM_SDK -miphonesimulator-version-min=13.0 -arch arm64" \
go build -buildmode=c-archive -o "$BUILD_DIR/sim_arm64/libnembus_mobile.a" ../../packages/core/bridge

# 2. iOS Simulator x86_64
CGO_ENABLED=1 GOOS=ios GOARCH=amd64 \
CC="$CLANG" \
CGO_CFLAGS="-isysroot $SIM_SDK -miphonesimulator-version-min=13.0 -arch x86_64" \
CGO_LDFLAGS="-isysroot $SIM_SDK -miphonesimulator-version-min=13.0 -arch x86_64" \
go build -buildmode=c-archive -o "$BUILD_DIR/sim_x86/libnembus_mobile.a" ../../packages/core/bridge

# Combine simulator architectures into fat archive
mkdir -p "$BUILD_DIR/sim_fat"
lipo -create "$BUILD_DIR/sim_arm64/libnembus_mobile.a" "$BUILD_DIR/sim_x86/libnembus_mobile.a" -output "$BUILD_DIR/sim_fat/libnembus_mobile.a"

# 3. iOS Device arm64
CGO_ENABLED=1 GOOS=ios GOARCH=arm64 \
CC="$CLANG" \
CGO_CFLAGS="-isysroot $DEV_SDK -miphoneos-version-min=13.0 -arch arm64" \
CGO_LDFLAGS="-isysroot $DEV_SDK -miphoneos-version-min=13.0 -arch arm64" \
go build -buildmode=c-archive -o "$BUILD_DIR/device_arm64/libnembus_mobile.a" ../../packages/core/bridge

# Modulemap content
MODULEMAP='framework module "nembus_mobile" {
    header "nembus_mobile.h"
    export *
}'

# 4. Create Framework bundle structure for Simulator
SIM_FW="$BUILD_DIR/nembus_mobile.framework"
mkdir -p "$SIM_FW/Headers" "$SIM_FW/Modules"
cp "$BUILD_DIR/sim_fat/libnembus_mobile.a" "$SIM_FW/nembus_mobile"
cp "$BUILD_DIR/sim_arm64/libnembus_mobile.h" "$SIM_FW/Headers/nembus_mobile.h"
echo "$MODULEMAP" > "$SIM_FW/Modules/module.modulemap"

# Create Framework bundle structure for Device
DEV_FW="$BUILD_DIR/dev/nembus_mobile.framework"
mkdir -p "$DEV_FW/Headers" "$DEV_FW/Modules"
cp "$BUILD_DIR/device_arm64/libnembus_mobile.a" "$DEV_FW/nembus_mobile"
cp "$BUILD_DIR/device_arm64/libnembus_mobile.h" "$DEV_FW/Headers/nembus_mobile.h"
echo "$MODULEMAP" > "$DEV_FW/Modules/module.modulemap"

# 5. Create final xcframework
rm -rf ios/Frameworks/nembus_mobile.xcframework
mkdir -p ios/Frameworks
xcodebuild -create-xcframework \
  -framework "$SIM_FW" \
  -framework "$DEV_FW" \
  -output ios/Frameworks/nembus_mobile.xcframework

echo "Successfully created ios/Frameworks/nembus_mobile.xcframework!"
