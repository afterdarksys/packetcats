#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
APP_NAME="packetcats"
OUTPUT_DIR="bin"
CMD_PATH="./cmd/packetcats"

# Print colored message
print_message() {
    local color=$1
    local message=$2
    echo -e "${color}${message}${NC}"
}

# Check if Go is installed
if ! command -v go &> /dev/null; then
    print_message "$RED" "Error: Go is not installed. Please install Go 1.25.2 or later."
    exit 1
fi

# Get Go version
GO_VERSION=$(go version | awk '{print $3}')
print_message "$GREEN" "Using $GO_VERSION"

# Clean previous builds
print_message "$YELLOW" "Cleaning previous builds..."
rm -rf "$OUTPUT_DIR"
mkdir -p "$OUTPUT_DIR"

# Download dependencies
print_message "$YELLOW" "Downloading dependencies..."
go mod download

# Run tests (if any exist)
if ls *_test.go &> /dev/null || find . -name "*_test.go" -not -path "./vendor/*" | grep -q .; then
    print_message "$YELLOW" "Running tests..."
    go test ./...
fi

# Build for current platform
print_message "$YELLOW" "Building $APP_NAME..."
go build -o "$OUTPUT_DIR/$APP_NAME" "$CMD_PATH"

# Make the binary executable
chmod +x "$OUTPUT_DIR/$APP_NAME"

# Print success message
print_message "$GREEN" "✓ Build successful!"
print_message "$GREEN" "Binary location: $OUTPUT_DIR/$APP_NAME"

# Show binary info
if [ -f "$OUTPUT_DIR/$APP_NAME" ]; then
    SIZE=$(ls -lh "$OUTPUT_DIR/$APP_NAME" | awk '{print $5}')
    print_message "$GREEN" "Binary size: $SIZE"
fi

print_message "$YELLOW" "\nTo run the application:"
print_message "$NC" "  ./$OUTPUT_DIR/$APP_NAME --help"
