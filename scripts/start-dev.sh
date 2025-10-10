#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")"/.. && pwd)"
COMPOSE_FILE="$ROOT_DIR/docker-compose.dev.yml"

POSTGRES_DSN="postgres://keto:secret@localhost:5432/keto?sslmode=disable"
MIGRATIONS_DIR="$ROOT_DIR/server/db/migrations"

ensure_command() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "ERROR: '$1' is not installed" >&2
    exit 1
  fi
}

start_containers() {
  echo "Starting infrastructure containers (Postgres, Redis, MinIO, Keto, Kratos)..."
  docker compose -f "$COMPOSE_FILE" up -d postgres redis minio keto kratos kratos-db

  echo "Waiting for Postgres to become ready..."
  until docker compose -f "$COMPOSE_FILE" exec -T postgres pg_isready -U keto -d keto >/dev/null 2>&1; do
    sleep 2
  done

  echo "Infrastructure ready."
}

run_migrations() {
  echo "Running database migrations..."
  if command -v migrate >/dev/null 2>&1; then
    migrate -path "$MIGRATIONS_DIR" -database "$POSTGRES_DSN" up
  else
    (cd "$ROOT_DIR/server" && go run github.com/golang-migrate/migrate/v4/cmd/migrate@v4.17.1 \
      -path db/migrations -database "$POSTGRES_DSN" up)
  fi
  echo "Migrations applied."
}

start_backend() {
  echo "Starting Go backend..."
  (cd "$ROOT_DIR" && go run ./server/main.go) >/tmp/edduhub-backend.log 2>&1 &
  BACKEND_PID=$!
  echo $BACKEND_PID
}

start_frontend() {
  echo "Starting Next.js frontend..."
  if [ ! -d "$ROOT_DIR/client/node_modules" ]; then
    (cd "$ROOT_DIR/client" && npm install)
  fi
  (cd "$ROOT_DIR/client" && npm run dev) >/tmp/edduhub-frontend.log 2>&1 &
  FRONTEND_PID=$!
  echo $FRONTEND_PID
}

ensure_command docker
ensure_command go
ensure_command npm

start_containers
run_migrations

BACKEND_PID=$(start_backend)
FRONTEND_PID=$(start_frontend)

echo "Backend logs: tail -f /tmp/edduhub-backend.log"
echo "Frontend logs: tail -f /tmp/edduhub-frontend.log"
echo "Press Ctrl+C to stop services."

cleanup() {
  echo "Stopping backend and frontend..."
  kill "$BACKEND_PID" "$FRONTEND_PID" >/dev/null 2>&1 || true
  wait "$BACKEND_PID" "$FRONTEND_PID" 2>/dev/null || true
  echo "Containers remain running. Use 'docker compose -f $COMPOSE_FILE down' to stop them."
}

trap cleanup EXIT

wait -n "$BACKEND_PID" "$FRONTEND_PID"
