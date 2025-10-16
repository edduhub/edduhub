#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")"/.. && pwd)"
COMPOSE_FILE="$ROOT_DIR/docker-compose.dev.yml"

echo "Stopping frontend processes..."
pkill -f "bun run dev" || true

echo "Stopping backend processes..."
pkill -f "go run ./server/main.go" || true

echo "Stopping infrastructure containers..."
docker compose -f "$COMPOSE_FILE" down

echo "All services stopped."