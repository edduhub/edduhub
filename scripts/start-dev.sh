#!/usr/bin/env bash
set -euo pipefail

# NOTE: If you have a local PostgreSQL instance running on port 5432,
# you must stop it first: brew services stop postgresql@14
# Otherwise connections will go to the local instance instead of Docker.

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")"/.. && pwd)"
COMPOSE_FILE="$ROOT_DIR/docker-compose.dev.yml"

POSTGRES_USER="keto"
POSTGRES_PASSWORD="secret"
POSTGRES_DB="keto"
POSTGRES_DSN="postgres://$POSTGRES_USER:$POSTGRES_PASSWORD@127.0.0.1:5432/$POSTGRES_DB?sslmode=disable"
KRATOS_DSN="postgres://kratos:secret@kratos-db:5432/kratos?sslmode=disable"
MIGRATIONS_DIR="$ROOT_DIR/server/db/migrations"

ensure_command() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "ERROR: '$1' is not installed" >&2
    exit 1
  fi
}

ensure_docker_compose() {
  if ! docker compose version >/dev/null 2>&1; then
    echo "ERROR: 'docker compose' is not available" >&2
    exit 1
  fi
}

find_env_file() {
  for candidate in "$ROOT_DIR/server/.env.local" "$ROOT_DIR/server/.env" "$ROOT_DIR/.env"; do
    if [ -f "$candidate" ]; then
      echo "$candidate"
      return
    fi
  done
  echo ""
}

start_containers() {
  echo "Starting infrastructure containers (Postgres, Redis, MinIO, Keto, Kratos DB)..."
  docker compose -f "$COMPOSE_FILE" up -d postgres redis minio keto kratos-db

  echo "Waiting for Postgres to become ready..."
  until docker compose -f "$COMPOSE_FILE" exec -T postgres pg_isready -U keto -d keto >/dev/null 2>&1; do
    sleep 2
  done

  ensure_postgres_bootstrap

  echo "Waiting for Kratos database to become ready..."
  until docker compose -f "$COMPOSE_FILE" exec -T kratos-db pg_isready -U kratos -d kratos >/dev/null 2>&1; do
    sleep 2
  done

  run_kratos_migrations

  echo "Starting Kratos service..."
  docker compose -f "$COMPOSE_FILE" up -d kratos

  echo "Waiting for Kratos to become ready..."
  until curl -sf http://localhost:4434/health/ready >/dev/null 2>&1; do
    sleep 2
  done

  echo "Checking Keto status..."
  if ! docker ps --filter "name=edduhub-keto" --filter "status=running" | grep -q edduhub-keto; then
    echo "WARNING: Keto container is not running. Checking logs..."
    docker logs edduhub-keto 2>&1 | tail -n 10
    echo "Attempting to restart Keto..."
    docker compose -f "$COMPOSE_FILE" restart keto
    sleep 3
  fi

  echo "Infrastructure ready."
}

run_kratos_migrations() {
  echo "Running Kratos SQL migrations..."
  if ! docker compose -f "$COMPOSE_FILE" run --rm --no-deps kratos \
    migrate sql up "$KRATOS_DSN" --yes \
    >/tmp/edduhub-kratos-migrate.log 2>&1; then
    echo "ERROR: Kratos migrations failed. Check /tmp/edduhub-kratos-migrate.log for details." >&2
    tail -n 20 /tmp/edduhub-kratos-migrate.log >&2 || true
    exit 1
  fi
  echo "Kratos migrations applied."
}

ensure_postgres_bootstrap() {
  echo "Ensuring application Postgres role and database exist..."
  
  # The postgres container already has the keto user and database created via environment variables
  # We just need to verify connectivity
  if docker compose -f "$COMPOSE_FILE" exec -T postgres \
    env PGPASSWORD="$POSTGRES_PASSWORD" psql \
      -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c "SELECT 1" >/dev/null 2>&1; then
    echo "Postgres role and database verified."
    return 0
  fi
  
  echo "ERROR: Cannot connect to Postgres database. Check container logs." >&2
  docker logs edduhub-postgres 2>&1 | tail -n 20
  exit 1
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

ensure_command docker
ensure_docker_compose
ensure_command go
ensure_command bun
ensure_command curl

# Check for local PostgreSQL on port 5432
if lsof -i :5432 -sTCP:LISTEN | grep -q postgres 2>/dev/null; then
  echo "WARNING: Local PostgreSQL is running on port 5432."
  echo "This may conflict with the Docker PostgreSQL container."
  echo "To stop it, run: brew services stop postgresql@14"
  echo ""
  read -p "Continue anyway? (y/N) " -n 1 -r
  echo
  if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    exit 1
  fi
fi

start_containers
run_migrations

echo "Starting Go backend..."
# Load environment variables
ENV_FILE="$(find_env_file)"
if [ -n "$ENV_FILE" ]; then
  set -a
  # shellcheck disable=SC1090
  source "$ENV_FILE"
  set +a
else
  echo "WARNING: No environment file found at server/.env.local, server/.env, or .env"
fi
(cd "$ROOT_DIR" && go run ./server/main.go) >/tmp/edduhub-backend.log 2>&1 &
BACKEND_PID=$!

echo "Starting Next.js frontend..."
(cd "$ROOT_DIR/client" && bun install --frozen-lockfile)
# Unset PORT to avoid conflict with backend
(cd "$ROOT_DIR/client" && unset PORT && bun run dev) >/tmp/edduhub-frontend.log 2>&1 &
FRONTEND_PID=$!

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

# Wait for either process to exit
while kill -0 "$BACKEND_PID" 2>/dev/null && kill -0 "$FRONTEND_PID" 2>/dev/null; do
  sleep 1
done
