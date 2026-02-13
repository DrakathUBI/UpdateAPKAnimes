#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")"
mkdir -p dist
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o dist/MacroClone-win-x64.exe .
GOOS=windows GOARCH=386 CGO_ENABLED=0 go build -o dist/MacroClone-win-x86.exe .
GOOS=windows GOARCH=arm64 CGO_ENABLED=0 go build -o dist/MacroClone-win-arm64.exe .
echo "Build concluído em ./dist"
