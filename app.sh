#!/bin/bash
set -e

# Find all services dynamically
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SERVICES_DIR="$SCRIPT_DIR/services"

# Load environment variables from .env
if [ -f "$SCRIPT_DIR/.env" ]; then
  export $(grep -v '^#' "$SCRIPT_DIR/.env" | xargs)
fi

ALL_SERVICES=()
if [ -d "$SERVICES_DIR" ]; then
  for d in "$SERVICES_DIR"/*; do
    if [ -d "$d" ]; then
      ALL_SERVICES+=("$(basename "$d")")
    fi
  done
fi

usage() {
  cat <<EOF
Usage: $0 {up|down|build|setup|migrate|gqlgen} [options]

Commands:
  up          Start services via compose (Docker or Podman)
  down        Stop services via compose (Docker or Podman)
  build       Build Go binaries and restart selected containers
  setup       Reset containers, run migrations, compose supergraph, and start services
  migrate     Run migrations only
  gqlgen      Generate GraphQL models and recompose supergraph

Options:
EOF
  for s in "${ALL_SERVICES[@]}"; do
    printf "  -%-10s Target only %s service\n" "$s" "$s"
  done
  echo ""
  echo "Migration forms: migrate [-service] | migrate [-service] VERSION up|down"
  echo "If no service flag is given, applies to all services."
  exit 1
}

CMD="$1"
if [ -z "$CMD" ]; then
  usage
fi
shift

TARGETS=()
MIGRATION_VERSION=""
MIGRATION_ACTION=""

for arg in "$@"; do
  if [ "$CMD" = "migrate" ] && [ "$arg" = "-latest" ]; then
    echo "-latest is not needed; migrate applies all pending up migrations." >&2
    usage
  elif [ "$CMD" = "migrate" ] && [[ "$arg" != -* ]] && [ -z "$MIGRATION_VERSION" ]; then
    MIGRATION_VERSION="$arg"
  elif [ "$CMD" = "migrate" ] && [[ "$arg" != -* ]] && [ -n "$MIGRATION_VERSION" ] && [ -z "$MIGRATION_ACTION" ]; then
    MIGRATION_ACTION="$arg"
  elif [[ "$arg" == -* ]]; then
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
      echo "Unknown option: $arg" >&2
      usage
    fi
  else
    echo "Unexpected argument: $arg" >&2
    usage
  fi
done

if [ "$CMD" = "migrate" ] && [ -n "$MIGRATION_VERSION" ] && [[ ! "$MIGRATION_VERSION" =~ ^[0-9]+$ ]]; then
  echo "Migration version must be numeric." >&2
  usage
fi

if [ "$CMD" = "migrate" ] && [ -n "$MIGRATION_VERSION" ] && [ "$MIGRATION_ACTION" != "up" ] && [ "$MIGRATION_ACTION" != "down" ]; then
  echo "Migration action must be 'up' or 'down'." >&2
  usage
fi

if [ "$CMD" = "migrate" ] && [ -z "$MIGRATION_VERSION" ]; then
  MIGRATION_ACTION="latest"
fi

if [ ${#TARGETS[@]} -eq 0 ]; then
  TARGETS=("${ALL_SERVICES[@]}")
fi

case "$CMD" in
  up|down|build|setup|migrate|gqlgen) ;;
  *) echo "Unknown command: $CMD" >&2; usage ;;
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

# Ensure Apollo Rover CLI is installed (supports Windows PowerShell, macOS open, Linux curl/npm)
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

# Ensure Go is installed (cross-platform prompt)
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

# Resolve database URL for a service
get_db_url() {
  local svc="$1"
  local var_name
  var_name="$(echo "$svc" | tr '[:lower:]' '[:upper:]')DB_URL"
  local svc_db_url="${!var_name}"
  if [ -z "$svc_db_url" ]; then
    svc_db_url="$DB_URL"
  fi
  echo "$svc_db_url"
}

# Helper to extract database name from connection string
get_db_name() {
  local url="$1"
  local url_no_query="${url%%\?*}"
  echo "${url_no_query##*/}"
}

wait_for_db() {
  echo "Waiting for database to be ready..."
  until $DOCKER_COMPOSE exec -T db pg_isready -U postgres >/dev/null 2>&1 && \
        $DOCKER_COMPOSE exec -T db psql -U postgres -d postgres -c "SELECT 1;" >/dev/null 2>&1; do
    sleep 1
  done
  echo "✅ Database is ready."
}

ensure_databases() {
  for svc in "${TARGETS[@]}"; do
    local svc_db_url db_name
    svc_db_url="$(get_db_url "$svc")"
    db_name="$(get_db_name "$svc_db_url")"
    if [ -n "$db_name" ] && [ "$db_name" != "postgres" ]; then
      if ! $DOCKER_COMPOSE exec -T db psql -U postgres -d postgres -Atc "SELECT 1 FROM pg_database WHERE datname = '$db_name';" | grep -q 1; then
        $DOCKER_COMPOSE exec -T db psql -U postgres -d postgres -v ON_ERROR_STOP=1 -c "CREATE DATABASE \"$db_name\";" >/dev/null
      fi
    fi
  done
}

ensure_migration_table() {
  local svc_db_url="$1"
  $DOCKER_COMPOSE exec -T db psql "$svc_db_url" -v ON_ERROR_STOP=1 -c '
    CREATE TABLE IF NOT EXISTS schema_migrations (
      version TEXT PRIMARY KEY,
      dirty BOOLEAN NOT NULL DEFAULT false,
      applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
    );
  ' >/dev/null
}

find_migration() {
  local migration_dir="$1"
  local version="$2"
  for migration in "$migration_dir/${version}"_*.up.sql; do
    if [ -f "$migration" ]; then
      echo "$migration"
      return 0
    fi
  done
  return 1
}

apply_migration() {
  local svc="$1"
  local svc_db_url="$2"
  local migration="$3"
  local filename version state down_migration

  filename="$(basename "$migration")"
  version="${svc}_${filename%%_*}"
  state="$($DOCKER_COMPOSE exec -T db psql "$svc_db_url" -Atc "SELECT CASE WHEN dirty THEN 'dirty' ELSE 'clean' END FROM schema_migrations WHERE version = '$version';" | tr -d '\r')"

  if [ "$state" = "clean" ]; then
    echo "  Skipping $filename (already applied)"
    return 0
  fi

  if [ "$state" = "dirty" ]; then
    echo "  ⚠️  $filename is marked dirty; auto-healing..."
    down_migration="${migration%.up.sql}.down.sql"
    if [ -f "$down_migration" ]; then
      echo "  Rolling back $filename using $(basename "$down_migration")..."
      $DOCKER_COMPOSE exec -T db psql "$svc_db_url" -v ON_ERROR_STOP=1 -f - < "$down_migration" >/dev/null 2>&1 || true
    else
      echo "  No down migration found for $filename; resetting dirty state directly..."
    fi
    $DOCKER_COMPOSE exec -T db psql "$svc_db_url" -v ON_ERROR_STOP=1 -c "DELETE FROM schema_migrations WHERE version = '$version';" >/dev/null
    echo "  Dirty state cleared for $version. Retrying application..."
  fi

  echo "  Applying $filename..."
  $DOCKER_COMPOSE exec -T db psql "$svc_db_url" -v ON_ERROR_STOP=1 -c "INSERT INTO schema_migrations (version, dirty) VALUES ('$version', true) ON CONFLICT (version) DO UPDATE SET dirty = true;" >/dev/null
  if ! $DOCKER_COMPOSE exec -T db psql "$svc_db_url" -v ON_ERROR_STOP=1 -f - < "$migration" >/dev/null; then
    echo "  ERROR: $filename failed and remains marked dirty" >&2
    return 1
  fi
  $DOCKER_COMPOSE exec -T db psql "$svc_db_url" -v ON_ERROR_STOP=1 -c "UPDATE schema_migrations SET dirty = false, applied_at = NOW() WHERE version = '$version';" >/dev/null
}

rollback_migration() {
  local svc="$1"
  local svc_db_url="$2"
  local migration="$3"
  local filename version state down_migration

  filename="$(basename "$migration")"
  version="${svc}_${filename%%_*}"
  state="$($DOCKER_COMPOSE exec -T db psql "$svc_db_url" -Atc "SELECT CASE WHEN dirty THEN 'dirty' ELSE 'clean' END FROM schema_migrations WHERE version = '$version';" | tr -d '\r')"

  if [ "$state" != "clean" ]; then
    echo "  ERROR: $filename is not applied cleanly" >&2
    return 1
  fi

  down_migration="${migration%.sql}.down.sql"
  if [ ! -f "$down_migration" ]; then
    echo "  ERROR: no rollback file for $filename ($down_migration)" >&2
    return 1
  fi

  echo "  Rolling back $filename..."
  $DOCKER_COMPOSE exec -T db psql "$svc_db_url" -v ON_ERROR_STOP=1 -c "UPDATE schema_migrations SET dirty = true WHERE version = '$version';" >/dev/null
  if ! $DOCKER_COMPOSE exec -T db psql "$svc_db_url" -v ON_ERROR_STOP=1 -f - < "$down_migration" >/dev/null; then
    echo "  ERROR: rollback for $filename failed and remains marked dirty" >&2
    return 1
  fi
  $DOCKER_COMPOSE exec -T db psql "$svc_db_url" -v ON_ERROR_STOP=1 -c "DELETE FROM schema_migrations WHERE version = '$version';" >/dev/null
}

run_migrations() {
  for svc in "${TARGETS[@]}"; do
    echo "--- $svc migrations ---"
    local migration_dir="$SERVICES_DIR/$svc/migrations"
    if [ ! -d "$migration_dir" ]; then
      echo "  No migrations directory found for $svc"
      continue
    fi

    local svc_db_url
    svc_db_url="$(get_db_url "$svc")"
    ensure_migration_table "$svc_db_url"

    if [ "$CMD" != "migrate" ] || [ -z "$MIGRATION_VERSION" ] || [ "$MIGRATION_ACTION" = "latest" ]; then
      for migration in "$migration_dir"/*.up.sql; do
        [ -f "$migration" ] || continue
        apply_migration "$svc" "$svc_db_url" "$migration" || return 1
      done
    else
      local migration
      migration="$(find_migration "$migration_dir" "$MIGRATION_VERSION")" || {
        echo "  ERROR: migration version $MIGRATION_VERSION not found for $svc" >&2
        return 1
      }
      if [ "$MIGRATION_ACTION" = "up" ]; then
        apply_migration "$svc" "$svc_db_url" "$migration" || return 1
      else
        rollback_migration "$svc" "$svc_db_url" "$migration" || return 1
      fi
    fi
  done
}

compose_supergraph() {
  rover supergraph compose --config "$SCRIPT_DIR/supergraph-config.yaml" > "$SCRIPT_DIR/rover/supergraph.graphql"
}

generate_gqlgen() {
  for svc in "${TARGETS[@]}"; do
    local svc_dir="$SERVICES_DIR/$svc"
    if [ -f "$svc_dir/gqlgen.yml" ]; then
      echo "--- Generating gqlgen for $svc ---"
      (cd "$svc_dir" && go run github.com/99designs/gqlgen generate)
    fi
  done
}

# Check prerequisites based on the command being run
check_prerequisites() {
  case "$CMD" in
    up|down|build|setup|migrate)
      ensure_container_running
      ;;
  esac

  case "$CMD" in
    build|setup|gqlgen)
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
    generate_gqlgen
    compose_supergraph
    for svc in "${TARGETS[@]}"; do
      echo "Building $svc..."
      (cd "$SCRIPT_DIR" && go build -o "bin/$svc" "./services/$svc")
      $DOCKER_COMPOSE up -d --build "$svc"
    done
    $DOCKER_COMPOSE up -d --force-recreate router
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

    wait_for_db

    ensure_databases
    run_migrations
    compose_supergraph

    echo "Building and starting all services..."
    $DOCKER_COMPOSE up -d --build
    ;;
  migrate)
    echo "Starting database..."
    $DOCKER_COMPOSE up -d db

    wait_for_db

    ensure_databases
    run_migrations
    ;;
  gqlgen)
    generate_gqlgen
    compose_supergraph
    ;;
esac
