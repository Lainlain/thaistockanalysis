#!/bin/bash
# Build script for ThaiStockAnalysis

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🏗️  Building ThaiStockAnalysis${NC}"

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}❌ Go is not installed. Please install Go 1.24.6 or later.${NC}"
    exit 1
fi

# Check Go version
GO_VERSION=$(go version | cut -d' ' -f3)
echo -e "${GREEN}✅ Go version: ${GO_VERSION}${NC}"

# Create bin directory
mkdir -p bin

# Clean previous builds
echo -e "${YELLOW}🧹 Cleaning previous builds...${NC}"
rm -f bin/*

# Install dependencies
echo -e "${YELLOW}📦 Installing dependencies...${NC}"
go mod download
go mod tidy

# Format code
echo -e "${YELLOW}🎨 Formatting code...${NC}"
go fmt ./...

# Run tests
echo -e "${YELLOW}🧪 Running tests...${NC}"
go test ./...

# Build for current platform
echo -e "${YELLOW}🔨 Building for current platform...${NC}"
go build -o bin/thaistockanalysis cmd/server/main.go

# Build for different platforms
echo -e "${YELLOW}🌍 Building for multiple platforms...${NC}"

# Linux
echo -e "${BLUE}Building for Linux...${NC}"
GOOS=linux GOARCH=amd64 go build -o bin/thaistockanalysis-linux cmd/server/main.go

# Windows
echo -e "${BLUE}Building for Windows...${NC}"
GOOS=windows GOARCH=amd64 go build -o bin/thaistockanalysis.exe cmd/server/main.go

# macOS
echo -e "${BLUE}Building for macOS...${NC}"
GOOS=darwin GOARCH=amd64 go build -o bin/thaistockanalysis-macos cmd/server/main.go

# ARM64 builds
echo -e "${BLUE}Building for ARM64...${NC}"
GOOS=linux GOARCH=arm64 go build -o bin/thaistockanalysis-linux-arm64 cmd/server/main.go
GOOS=darwin GOARCH=arm64 go build -o bin/thaistockanalysis-macos-arm64 cmd/server/main.go

# Make binaries executable
chmod +x bin/thaistockanalysis*

# Show build results
echo -e "${GREEN}✅ Build completed successfully!${NC}"
echo -e "${BLUE}📁 Built binaries:${NC}"
ls -la bin/

echo -e "${GREEN}🎉 Ready to deploy!${NC}"
echo -e "${YELLOW}💡 To run locally: ./bin/thaistockanalysis${NC}"
echo -e "${YELLOW}💡 To deploy: Copy the appropriate binary to your server${NC}"
