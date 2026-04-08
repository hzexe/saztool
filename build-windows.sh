#!/usr/bin/env bash
set -euo pipefail

ROOT=$(cd "$(dirname "$0")" && pwd)
cd "$ROOT"
mkdir -p bin

GOOS=windows GOARCH=amd64 go build -o ./bin/saztool-windows-amd64.exe ./cmd/saztool
GOOS=windows GOARCH=arm64 go build -o ./bin/saztool-windows-arm64.exe ./cmd/saztool

echo "Built:"
ls -lh ./bin/saztool-windows-*.exe
