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
  echo "  up          Start services via compose (Docker or Podman)"
  echo "  down        Stop services via compose (Docker or Podman)"
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

# Helper to prompt user (defaults to Yes)
prompt_confirm() {
  local prompt_msg="$1"
  read -r -p "$prompt_msg [Y/n]: " response
  case "$response" in
    [nN][oO]|[nN])
      return 1
      ;;
    *)
      return 0
      ;;
  esac
}

# Detect which container runtime is actively running (Docker or Podman)
DOCKER_COMPOSE=""
ensure_container_running() {
  # 1. Check if Docker is actively running
  if command -v docker >/dev/null 2>&1 && docker info >/dev/null 2>&1; then
    if docker compose version >/dev/null 2>&1; then
      DOCKER_COMPOSE="docker compose"
    elif command -v docker-compose >/dev/null 2>&1; then
      DOCKER_COMPOSE="docker-compose"
    else
      DOCKER_COMPOSE="docker compose"
    fi
    echo "✅ Docker is running ($DOCKER_COMPOSE)."
    return 0
  fi

  # 2. Check if Podman is actively running
  if command -v podman >/dev/null 2>&1 && podman info >/dev/null 2>&1; then
    if podman compose version >/dev/null 2>&1; then
      DOCKER_COMPOSE="podman compose"
    elif command -v podman-compose >/dev/null 2>&1; then
      DOCKER_COMPOSE="podman-compose"
    else
      DOCKER_COMPOSE="podman compose"
    fi
    echo "✅ Podman is running ($DOCKER_COMPOSE)."
    return 0
  fi

  # 3. Neither is running
  echo "❌ No active container runtime found."
  echo "   Please start Docker Desktop or run 'podman machine start', then re-run this script."
  exit 1
}

# Ensure Apollo Rover CLI is installed
ensure_rover() {
  # Add default Rover install paths to current PATH if not present
  export PATH="$HOME/.rover/bin:$USERPROFILE/.rover/bin:$LOCALAPPDATA/Rover/bin:$PATH"

  if ! command -v rover >/dev/null 2>&1; then
    echo "⚠️  Apollo Rover CLI ('rover') is required but not found."
    if prompt_confirm "Would you like to install Apollo Rover now?"; then
      echo "Installing Apollo Rover..."
      if command -v powershell.exe >/dev/null 2>&1; then
        powershell.exe -NoProfile -ExecutionPolicy Bypass -Command "iwr 'https://rover.apollo.dev/win/latest' | iex"
      elif command -v curl >/dev/null 2>&1; then
        curl -sSL https://rover.apollo.dev/nix/latest | sh
      elif command -v npm >/dev/null 2>&1; then
        npm install -g @apollo/rover
      else
        echo "❌ Neither PowerShell, curl, nor npm found to install rover."
        echo "Please install Rover manually: https://www.apollographql.com/docs/rover/getting-started"
        exit 1
      fi

      export PATH="$HOME/.rover/bin:$USERPROFILE/.rover/bin:$LOCALAPPDATA/Rover/bin:$PATH"
      if ! command -v rover >/dev/null 2>&1; then
        echo "Rover was installed. You may need to restart your terminal if it is not immediately recognized."
      else
        echo "✅ Apollo Rover installed successfully."
      fi
    else
      echo "Skipping Rover installation. Note that supergraph composition may fail without it."
    fi
  fi
}

# Ensure Go is installed
ensure_go() {
  if ! command -v go >/dev/null 2>&1; then
    echo "❌ Go is not installed, but is required for '$CMD'."
    if prompt_confirm "Would you like to open the Go download page?"; then
      if command -v powershell.exe >/dev/null 2>&1; then
        powershell.exe -Command "Start-Process 'https://go.dev/dl/'"
      elif command -v open >/dev/null 2>&1; then
        open "https://go.dev/dl/"
      elif command -v xdg-open >/dev/null 2>&1; then
        xdg-open "https://go.dev/dl/"
      fi
    fi
    echo "Please install Go and re-run this script."
    exit 1
  fi
}

# Check prerequisites based on the command being run
check_prerequisites() {
  case "$CMD" in
    up|down|build|setup|migrate)
      ensure_container_running
      ;;
  esac

  case "$CMD" in
    setup|gqlgen)
      ensure_rover
      ;;
  esac

  case "$CMD" in
    build|gqlgen)
      ensure_go
      ;;
  esac
}

# Run prerequisite checks before executing commands
check_prerequisites

case "$CMD" in
  up)
    $DOCKER_COMPOSE up -d
    ;;
  down)
    $DOCKER_COMPOSE down
    ;;
  build)
    for svc in "${TARGETS[@]}"; do
      echo "Building $svc..."
      (cd "$SCRIPT_DIR" && go build -o "bin/$svc" "./services/$svc")
      echo "Rebuilding and restarting docker container for $svc..."
      $DOCKER_COMPOSE up -d --build "$svc"
    done
    ;;
  setup)
    # Configure git hooks if in a git repository
    if [ -d "$SCRIPT_DIR/.git" ] && [ -d "$SCRIPT_DIR/.githooks" ]; then
      git config core.hooksPath .githooks >/dev/null 2>&1 || true
    fi

    echo "Stopping and removing all services..."
    $DOCKER_COMPOSE down -v

    echo "Starting database and redis..."
    $DOCKER_COMPOSE up -d db redis

    echo "Waiting for database to be ready..."
    until $DOCKER_COMPOSE exec -T db pg_isready -U postgres >/dev/null 2>&1; do
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
        if ! $DOCKER_COMPOSE exec -T db psql -U postgres -d postgres -Atc "SELECT 1 FROM pg_database WHERE datname = '$DB_NAME';" | grep -q 1; then
          $DOCKER_COMPOSE exec -T db psql -U postgres -d postgres -v ON_ERROR_STOP=1 -c "CREATE DATABASE \"$DB_NAME\";" >/dev/null
        fi
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

        $DOCKER_COMPOSE exec -T db psql "$SVC_DB_URL" -v ON_ERROR_STOP=1 -c '
          CREATE TABLE IF NOT EXISTS schema_migrations (
            version TEXT PRIMARY KEY,
            dirty BOOLEAN NOT NULL DEFAULT false,
            applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
          );
        ' >/dev/null
        
        for f in "$migration_dir"/*.up.sql; do
          [ -f "$f" ] || continue
          filename="$(basename "$f")"
          version="${filename%%_*}"
          state="$($DOCKER_COMPOSE exec -T db psql "$SVC_DB_URL" -Atc "SELECT CASE WHEN dirty THEN 'dirty' ELSE 'clean' END FROM schema_migrations WHERE version = '$version';" | tr -d '\r')"
          if [ "$state" = "clean" ]; then
            echo "  Skipping $filename (already applied)"
            continue
          fi
          if [ "$state" = "dirty" ]; then
            echo "  ERROR: $filename is marked dirty; repair it before retrying" >&2
            exit 1
          fi
          echo "  Applying $filename..."
          $DOCKER_COMPOSE exec -T db psql "$SVC_DB_URL" -v ON_ERROR_STOP=1 -c "INSERT INTO schema_migrations (version, dirty) VALUES ('$version', true);" >/dev/null
          if ! $DOCKER_COMPOSE exec -T db psql "$SVC_DB_URL" -v ON_ERROR_STOP=1 -f - < "$f" >/dev/null; then
            echo "  ERROR: $filename failed and remains marked dirty" >&2
            exit 1
          fi
          $DOCKER_COMPOSE exec -T db psql "$SVC_DB_URL" -v ON_ERROR_STOP=1 -c "UPDATE schema_migrations SET dirty = false, applied_at = NOW() WHERE version = '$version';" >/dev/null
        done
      else
        echo "  No migrations directory found for $svc"
      fi
    done

    echo "Composing Apollo Federation supergraph..."
    rover supergraph compose --config "$SCRIPT_DIR/supergraph-config.yaml" > "$SCRIPT_DIR/rover/supergraph.graphql"

    echo "Building and starting all services..."
    $DOCKER_COMPOSE up -d --build
    ;;
  migrate)
    echo "Starting database..."
    $DOCKER_COMPOSE up -d db

    echo "Waiting for database to be ready..."
    until $DOCKER_COMPOSE exec -T db pg_isready -U postgres >/dev/null 2>&1; do
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
        if ! $DOCKER_COMPOSE exec -T db psql -U postgres -d postgres -Atc "SELECT 1 FROM pg_database WHERE datname = '$DB_NAME';" | grep -q 1; then
          $DOCKER_COMPOSE exec -T db psql -U postgres -d postgres -v ON_ERROR_STOP=1 -c "CREATE DATABASE \"$DB_NAME\";" >/dev/null
        fi
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

        $DOCKER_COMPOSE exec -T db psql "$SVC_DB_URL" -v ON_ERROR_STOP=1 -c '
          CREATE TABLE IF NOT EXISTS schema_migrations (
            version TEXT PRIMARY KEY,
            dirty BOOLEAN NOT NULL DEFAULT false,
            applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
          );
        ' >/dev/null
        
        for f in "$migration_dir"/*.up.sql; do
          [ -f "$f" ] || continue
          filename="$(basename "$f")"
          version="${filename%%_*}"
          state="$($DOCKER_COMPOSE exec -T db psql "$SVC_DB_URL" -Atc "SELECT CASE WHEN dirty THEN 'dirty' ELSE 'clean' END FROM schema_migrations WHERE version = '$version';" | tr -d '\r')"
          if [ "$state" = "clean" ]; then
            echo "  Skipping $filename (already applied)"
            continue
          fi
          if [ "$state" = "dirty" ]; then
            echo "  ERROR: $filename is marked dirty; repair it before retrying" >&2
            exit 1
          fi
          echo "  Applying $filename..."
          $DOCKER_COMPOSE exec -T db psql "$SVC_DB_URL" -v ON_ERROR_STOP=1 -c "INSERT INTO schema_migrations (version, dirty) VALUES ('$version', true);" >/dev/null
          if ! $DOCKER_COMPOSE exec -T db psql "$SVC_DB_URL" -v ON_ERROR_STOP=1 -f - < "$f" >/dev/null; then
            echo "  ERROR: $filename failed and remains marked dirty" >&2
            exit 1
          fi
          $DOCKER_COMPOSE exec -T db psql "$SVC_DB_URL" -v ON_ERROR_STOP=1 -c "UPDATE schema_migrations SET dirty = false, applied_at = NOW() WHERE version = '$version';" >/dev/null
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
