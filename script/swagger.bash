#!/usr/bin/env bash

set -e

cd "$(dirname "$0")/.."

echo "📖 Generating swagger..."
go run github.com/swaggo/swag/cmd/swag@v1.16.4 init \
  -g cmd/server/main.go \
  -o docs \
  --parseDependency \
  --parseInternal

echo "✅ Swagger updated (docs/)"