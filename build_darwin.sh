#!/bin/bash

set -e

APP_NAME="qtray"
TEMPLATE_DIR="build/darwin"
TEMP_DIR="temp_build"
OUTPUT_ZIP="${APP_NAME}_darwin.zip"

echo "Building for macOS..."

# cleanup
rm -rf "${TEMP_DIR}"
rm -f "${OUTPUT_ZIP}"

# template
cp -r "${TEMPLATE_DIR}" "${TEMP_DIR}"

# build
CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build \
    -ldflags='-s -w' \
    -o "${TEMP_DIR}/${APP_NAME}.app/Contents/MacOS/${APP_NAME}" .

# zip
cd "${TEMP_DIR}"
zip -r "../${OUTPUT_ZIP}" "${APP_NAME}.app"
cd ..

# cleanup
rm -rf "${TEMP_DIR}"

echo "✅ Build completed: ${OUTPUT_ZIP}"
