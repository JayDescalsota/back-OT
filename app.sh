#!/bin/bash
set -e

# Find all services dynamically
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Load environment variables from .env
if [ -f "$SCRIPT_DIR/.env" ]; then
  export $(grep -v '^#' "$SCRIPT_DIR/.env" | xargs)
fi

SERVICES_DIR="$SCRIPT_DIR/services"
ALL_SERVICES=()

if [ -d "$SERVICES_DIR" ]; then
  for d in "$SERVICES_DIR"/*; do
    if [ -d "$d" ]; then
      ALL_SERVICES+=("$(basename "$d")")
    fi
  done
fi

# Helper to extract database name from connection string
get_db_name() {
  local url="$1"
  local url_no_query="${url%%\?*}"
  echo "${url_no_query##*/}"
}

usage() {
  echo "Usage: $0 {up|down|build|setup|migrate|gqlgen} [options]"
  echo ""
  echo "Commands:"
  echo "  up          Start services via docker-compose"
  echo "  down        Stop services via docker-compose"
  echo "  build       Build Go binaries"
  echo "  setup       Run migrations and seed data"
  echo "  migrate     Run migrations only"
  echo "  gqlgen      Generate GraphQL models and recompose supergraph"
  echo ""
  echo "Options:"
  for s in "${ALL_SERVICES[@]}"; do
    printf "  -%-10s Target only %s service\n" "$s" "$s"
  done
  echo ""
  echo "If no service flag is given, applies to all services."
  exit 1
}

CMD="$1"
if [ -z "$CMD" ]; then
  usage
fi
shift

TARGETS=()
for arg in "$@"; do
  if [[ "$arg" == -* ]]; then
    svc_name="${arg#-}"
    found=false
    for s in "${ALL_SERVICES[@]}"; do
      if [ "$s" = "$svc_name" ]; then
        TARGETS+=("$svc_name")
        found=true
        break
      fi
    done
    if [ "$found" = false ]; then
      echo "Unknown option: $arg"
      usage
    fi
  fi
done

if [ ${#TARGETS[@]} -eq 0 ]; then
  TARGETS=("${ALL_SERVICES[@]}")
fi

case "$CMD" in
  up|down|build|setup|migrate|gqlgen) ;;
  *) echo "Unknown command: $CMD"; usage ;;
esac

case "$CMD" in
  up)
    docker-compose up -d
    ;;
  down)
    docker-compose down
    ;;
  build)
    for svc in "${TARGETS[@]}"; do
      echo "Building $svc..."
      (cd "$SCRIPT_DIR" && go build -o "bin/$svc" "./services/$svc")
      echo "Rebuilding and restarting docker container for $svc..."
      docker-compose up -d --build "$svc"
    done
    ;;
  setup)
    echo "Stopping and removing all services..."
    docker compose down -v

    echo "Starting database and redis..."
    docker-compose up -d db redis

    echo "Waiting for database to be ready..."
    until docker-compose exec -T db pg_isready -U postgres >/dev/null 2>&1; do
      sleep 1
    done

    # Ensure databases exist for target services dynamically
    for svc in "${TARGETS[@]}"; do
      VAR_NAME="$(echo "$svc" | tr '[:lower:]' '[:upper:]')DB_URL"
      SVC_DB_URL="${!VAR_NAME}"
      if [ -z "$SVC_DB_URL" ]; then
        SVC_DB_URL="$DB_URL"
      fi
      DB_NAME=$(get_db_name "$SVC_DB_URL")
      if [ -n "$DB_NAME" ] && [ "$DB_NAME" != "postgres" ]; then
        docker-compose exec -T db psql -U postgres -c "CREATE DATABASE \"$DB_NAME\";" >/dev/null 2>&1 || true
      fi
    done

    echo "Running migrations..."
    for svc in "${TARGETS[@]}"; do
      echo "--- $svc migrations ---"
      migration_dir="$SERVICES_DIR/$svc/migrations"
      if [ -d "$migration_dir" ]; then
        # Resolve DB URL for this service
        VAR_NAME="$(echo "$svc" | tr '[:lower:]' '[:upper:]')DB_URL"
        SVC_DB_URL="${!VAR_NAME}"
        if [ -z "$SVC_DB_URL" ]; then
          SVC_DB_URL="$DB_URL"
        fi
        
        for f in "$migration_dir"/*.up.sql; do
          [ -f "$f" ] || continue
          echo "  Applying $(basename "$f")..."
          docker-compose exec -T db psql "$SVC_DB_URL" -f - < "$f" >/dev/null || \
          echo "  WARN: could not run $(basename "$f") — check DB_URL"
        done
      else
        echo "  No migrations directory found for $svc"
      fi
    done

    echo "Composing Apollo Federation supergraph..."
    rover supergraph compose --config "$SCRIPT_DIR/supergraph-config.yaml" > "$SCRIPT_DIR/rover/supergraph.graphql"

    echo "Building and starting all services..."
    docker-compose up -d --build
    ;;
  migrate)
    echo "Starting database..."
    docker-compose up -d db

    echo "Waiting for database to be ready..."
    until docker-compose exec -T db pg_isready -U postgres >/dev/null 2>&1; do
      sleep 1
    done

    # Ensure databases exist for target services dynamically
    for svc in "${TARGETS[@]}"; do
      VAR_NAME="$(echo "$svc" | tr '[:lower:]' '[:upper:]')DB_URL"
      SVC_DB_URL="${!VAR_NAME}"
      if [ -z "$SVC_DB_URL" ]; then
        SVC_DB_URL="$DB_URL"
      fi
      DB_NAME=$(get_db_name "$SVC_DB_URL")
      if [ -n "$DB_NAME" ] && [ "$DB_NAME" != "postgres" ]; then
        docker-compose exec -T db psql -U postgres -c "CREATE DATABASE \"$DB_NAME\";" >/dev/null 2>&1 || true
      fi
    done

    echo "Running migrations..."
    for svc in "${TARGETS[@]}"; do
      echo "--- $svc migrations ---"
      migration_dir="$SERVICES_DIR/$svc/migrations"
      if [ -d "$migration_dir" ]; then
        # Resolve DB URL for this service
        VAR_NAME="$(echo "$svc" | tr '[:lower:]' '[:upper:]')DB_URL"
        SVC_DB_URL="${!VAR_NAME}"
        if [ -z "$SVC_DB_URL" ]; then
          SVC_DB_URL="$DB_URL"
        fi
        
        for f in "$migration_dir"/*.up.sql; do
          [ -f "$f" ] || continue
          echo "  Applying $(basename "$f")..."
          docker-compose exec -T db psql "$SVC_DB_URL" -f - < "$f" >/dev/null || \
          echo "  WARN: could not run $(basename "$f") — check DB_URL"
        done
      else
        echo "  No migrations directory found for $svc"
      fi
    done
    ;;
  gqlgen)
    echo "Running gqlgen generate..."
    for svc in "${TARGETS[@]}"; do
      svc_dir="$SERVICES_DIR/$svc"
      if [ -f "$svc_dir/gqlgen.yml" ]; then
        echo "--- Generating gqlgen for $svc ---"
        (cd "$svc_dir" && go run github.com/99designs/gqlgen generate)
      fi
    done

    echo "Composing Apollo Federation supergraph..."
    rover supergraph compose --config "$SCRIPT_DIR/supergraph-config.yaml" > "$SCRIPT_DIR/rover/supergraph.graphql"
    ;;
esac
