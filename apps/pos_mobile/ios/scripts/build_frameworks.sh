#!/usr/bin/env bash
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
IOS_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
PROJECT_ROOT="$(cd "$IOS_DIR/../../.." && pwd)"
BRIDGE_DIR="$PROJECT_ROOT/packages/core/bridge"
RUNNER_DIR="$IOS_DIR/Runner"
FRAMEWORKS_DIR="$IOS_DIR/Frameworks"
MIN_VERSION="14.0"

# Locate Go binary in standard paths if not found in current PATH (essential for Xcode build phases)
if ! command -v go &> /dev/null; then
  for p in /opt/homebrew/bin /usr/local/bin /usr/local/go/bin "$HOME/go/bin" "$HOME/.go/bin"; do
    if [ -x "$p/go" ]; then
      export PATH="$p:$PATH"
      break
    fi
  done
fi

if ! command -v go &> /dev/null; then
  echo "❌ Error: 'go' binary not found. Please ensure Go is installed and available in PATH." >&2
  exit 1
fi

FORCE_BUILD=0
for arg in "$@"; do
  if [ "$arg" == "--force" ] || [ "$arg" == "-f" ]; then
    FORCE_BUILD=1
    break
  fi
done

# Smart incremental build check: skip recompilation during Xcode / Flutter build compile if sources haven't changed
XCFRAMEWORK_OUTPUT="$FRAMEWORKS_DIR/nembus_core.xcframework"
if [ "$FORCE_BUILD" -eq 0 ] && [ -d "$XCFRAMEWORK_OUTPUT" ]; then
  NEWER_SOURCE=$(find "$PROJECT_ROOT/packages/core" -name "*.go" -newer "$XCFRAMEWORK_OUTPUT" 2>/dev/null | head -n 1)
  if [ -z "$NEWER_SOURCE" ] && [ -f "$RUNNER_DIR/libnembus_core.h" ]; then
    echo "✅ [Nembus Core] Framework is up to date with Go source. Skipping recompilation."
    exit 0
  fi
  echo "🔄 [Nembus Core] Detected modified Go source ($NEWER_SOURCE). Rebuilding framework..."
fi

echo "================================================================="
echo " Building Nembus Core Framework for iOS (Device + Simulator) "
echo "================================================================="

mkdir -p "$FRAMEWORKS_DIR/device" "$FRAMEWORKS_DIR/simulator" "$RUNNER_DIR"

SDK_DEVICE_PATH=$(xcrun --sdk iphoneos --show-sdk-path)
CLANG_DEVICE=$(xcrun --sdk iphoneos --find clang)
SDK_SIM_PATH=$(xcrun --sdk iphonesimulator --show-sdk-path)
CLANG_SIM=$(xcrun --sdk iphonesimulator --find clang)

# 1. Compile Go C-archive for Device (arm64)
# Note: DO NOT use -tags netgo for iOS as it breaks DNS resolution on physical devices
echo "📦 [1/4] Compiling Go C-archive for iOS Device (arm64)..."
CGO_FLAGS_DEVICE="-isysroot $SDK_DEVICE_PATH -miphoneos-version-min=$MIN_VERSION -arch arm64 -g"
CGO_ENABLED=1 \
GOOS=ios \
GOARCH=arm64 \
CC="$CLANG_DEVICE" \
CGO_CFLAGS="$CGO_FLAGS_DEVICE" \
CGO_LDFLAGS="$CGO_FLAGS_DEVICE" \
go build -buildmode=c-archive \
  -o "$FRAMEWORKS_DIR/device/libnembus_core.a" "$BRIDGE_DIR"

cp "$FRAMEWORKS_DIR/device/libnembus_core.h" "$RUNNER_DIR/libnembus_core.h"

# 2. Compile Go C-archive for Simulator (arm64 + x86_64)
echo "📦 [2/4] Compiling Go C-archive for iOS Simulator (arm64 + x86_64)..."
SIM_TMP=$(mktemp -d)

# 2a. Simulator arm64 (Apple Silicon)
CGO_FLAGS_SIM_ARM64="-isysroot $SDK_SIM_PATH -miphonesimulator-version-min=$MIN_VERSION -arch arm64 -g"
CGO_ENABLED=1 GOOS=ios GOARCH=arm64 CC="$CLANG_SIM" \
CGO_CFLAGS="$CGO_FLAGS_SIM_ARM64" CGO_LDFLAGS="$CGO_FLAGS_SIM_ARM64" \
go build -buildmode=c-archive \
  -o "$SIM_TMP/sim_arm64.a" "$BRIDGE_DIR"

# 2b. Simulator x86_64 (Intel)
CGO_FLAGS_SIM_X86="-isysroot $SDK_SIM_PATH -miphonesimulator-version-min=$MIN_VERSION -arch x86_64 -g"
CGO_ENABLED=1 GOOS=ios GOARCH=amd64 CC="$CLANG_SIM" \
CGO_CFLAGS="$CGO_FLAGS_SIM_X86" CGO_LDFLAGS="$CGO_FLAGS_SIM_X86" \
go build -buildmode=c-archive \
  -o "$SIM_TMP/sim_x86.a" "$BRIDGE_DIR"

# Merge simulator archives into fat archive
lipo -create "$SIM_TMP/sim_arm64.a" "$SIM_TMP/sim_x86.a" -output "$FRAMEWORKS_DIR/simulator/libnembus_core.a"
cp "$SIM_TMP/sim_arm64.h" "$FRAMEWORKS_DIR/simulator/libnembus_core.h"

# 3. Create Dynamic Framework bundles for Device and Simulator
echo "🔨 [3/4] Linking Dynamic .framework bundles..."

# 3a. Device Dynamic Framework
DEV_FW_DIR="$FRAMEWORKS_DIR/device/nembus_core.framework"
DEV_DSYM="$FRAMEWORKS_DIR/device/nembus_core.framework.dSYM"
rm -rf "$DEV_FW_DIR" "$DEV_DSYM"
mkdir -p "$DEV_FW_DIR/Headers" "$DEV_FW_DIR/Modules"

$CLANG_DEVICE -dynamiclib \
  -isysroot "$SDK_DEVICE_PATH" \
  -miphoneos-version-min=$MIN_VERSION \
  -arch arm64 -g \
  -install_name @rpath/nembus_core.framework/nembus_core \
  -Wl,-all_load \
  -framework CoreFoundation -framework Security \
  -lpthread -lresolv \
  -o "$DEV_FW_DIR/nembus_core" \
  "$FRAMEWORKS_DIR/device/libnembus_core.a"

cp "$RUNNER_DIR/libnembus_core.h" "$DEV_FW_DIR/Headers/libnembus_core.h"
cp "$RUNNER_DIR/libnembus_core.h" "$DEV_FW_DIR/Headers/nembus_core.h"

cat << 'EOF' > "$DEV_FW_DIR/Modules/module.modulemap"
framework module nembus_core {
  umbrella header "nembus_core.h"
  export *
  module * { export * }
}
EOF

cat << EOF > "$DEV_FW_DIR/Info.plist"
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>CFBundleDevelopmentRegion</key>
	<string>en</string>
	<key>CFBundleExecutable</key>
	<string>nembus_core</string>
	<key>CFBundleIdentifier</key>
	<string>com.nastecsol.nembus-core</string>
	<key>CFBundleInfoDictionaryVersion</key>
	<string>6.0</string>
	<key>CFBundleName</key>
	<string>nembus_core</string>
	<key>CFBundlePackageType</key>
	<string>FMWK</string>
	<key>CFBundleShortVersionString</key>
	<string>1.0.0</string>
	<key>CFBundleVersion</key>
	<string>1</string>
	<key>MinimumOSVersion</key>
	<string>$MIN_VERSION</string>
</dict>
</plist>
EOF

# Generate dSYM for Device
echo "🔍 Extracting dSYM for Device Framework..."
dsymutil "$DEV_FW_DIR/nembus_core" -o "$DEV_DSYM"

# 3b. Simulator Dynamic Framework (fat arm64 + x86_64 dylib)
SIM_FW_DIR="$FRAMEWORKS_DIR/simulator/nembus_core.framework"
SIM_DSYM="$FRAMEWORKS_DIR/simulator/nembus_core.framework.dSYM"
rm -rf "$SIM_FW_DIR" "$SIM_DSYM"
mkdir -p "$SIM_FW_DIR/Headers" "$SIM_FW_DIR/Modules"

SIM_DYLIB_TMP=$(mktemp -d)

$CLANG_SIM -dynamiclib \
  -isysroot "$SDK_SIM_PATH" \
  -miphonesimulator-version-min=$MIN_VERSION \
  -arch arm64 -g \
  -install_name @rpath/nembus_core.framework/nembus_core \
  -Wl,-all_load \
  -framework CoreFoundation -framework Security \
  -lpthread -lresolv \
  -o "$SIM_DYLIB_TMP/sim_arm64.dylib" \
  "$SIM_TMP/sim_arm64.a"

$CLANG_SIM -dynamiclib \
  -isysroot "$SDK_SIM_PATH" \
  -miphonesimulator-version-min=$MIN_VERSION \
  -arch x86_64 -g \
  -install_name @rpath/nembus_core.framework/nembus_core \
  -Wl,-all_load \
  -framework CoreFoundation -framework Security \
  -lpthread -lresolv \
  -o "$SIM_DYLIB_TMP/sim_x86.dylib" \
  "$SIM_TMP/sim_x86.a"

lipo -create "$SIM_DYLIB_TMP/sim_arm64.dylib" "$SIM_DYLIB_TMP/sim_x86.dylib" -output "$SIM_FW_DIR/nembus_core"

cp "$RUNNER_DIR/libnembus_core.h" "$SIM_FW_DIR/Headers/libnembus_core.h"
cp "$RUNNER_DIR/libnembus_core.h" "$SIM_FW_DIR/Headers/nembus_core.h"
cp "$DEV_FW_DIR/Modules/module.modulemap" "$SIM_FW_DIR/Modules/module.modulemap"
cp "$DEV_FW_DIR/Info.plist" "$SIM_FW_DIR/Info.plist"

# Generate dSYM for Simulator
echo "🔍 Extracting dSYM for Simulator Framework..."
dsymutil "$SIM_FW_DIR/nembus_core" -o "$SIM_DSYM"
rm -rf "$SIM_DYLIB_TMP" "$SIM_TMP"

# 4. Generate multi-platform Apple XCFramework with embedded debug symbols
echo "🚀 [4/4] Generating multi-platform nembus_core.xcframework with dSYMs..."
rm -rf "$XCFRAMEWORK_OUTPUT"

xcodebuild -create-xcframework \
  -framework "$DEV_FW_DIR" \
  -debug-symbols "$DEV_DSYM" \
  -framework "$SIM_FW_DIR" \
  -debug-symbols "$SIM_DSYM" \
  -output "$XCFRAMEWORK_OUTPUT"

# Backward compatibility copies
mkdir -p "$RUNNER_DIR/device" "$RUNNER_DIR/simulator"
cp -R "$DEV_FW_DIR" "$RUNNER_DIR/device/"
cp -R "$SIM_FW_DIR" "$RUNNER_DIR/simulator/"
cp -R "$DEV_DSYM" "$RUNNER_DIR/device/"
cp -R "$DEV_DSYM" "$RUNNER_DIR/"
cp "$FRAMEWORKS_DIR/device/libnembus_core.a" "$RUNNER_DIR/device/libnembus_core.a"
cp "$FRAMEWORKS_DIR/simulator/libnembus_core.a" "$RUNNER_DIR/simulator/libnembus_core.a"
cp "$FRAMEWORKS_DIR/device/libnembus_core.a" "$RUNNER_DIR/libnembus_core.a"

echo "================================================================="
echo " ✅ Framework build completed successfully!"
echo "   - Device:    $DEV_FW_DIR"
echo "   - Simulator: $SIM_FW_DIR"
echo "   - Universal: $XCFRAMEWORK_OUTPUT"
echo "================================================================="
