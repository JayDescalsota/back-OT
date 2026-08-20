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

# Docker compose command detection (supports 'docker compose' and 'docker-compose')
if docker compose version >/dev/null 2>&1; then
  DOCKER_COMPOSE="docker compose"
elif command -v docker-compose >/dev/null 2>&1; then
  DOCKER_COMPOSE="docker-compose"
else
  DOCKER_COMPOSE="docker compose"
fi

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

# Ensure Docker daemon is running, try to launch Docker Desktop if not
ensure_docker_running() {
  if ! command -v docker >/dev/null 2>&1; then
    echo "❌ Docker is not installed."
    if prompt_confirm "Would you like to open the Docker Desktop download page?"; then
      if command -v powershell.exe >/dev/null 2>&1; then
        powershell.exe -Command "Start-Process 'https://www.docker.com/products/docker-desktop/'"
      elif command -v open >/dev/null 2>&1; then
        open "https://www.docker.com/products/docker-desktop/"
      elif command -v xdg-open >/dev/null 2>&1; then
        xdg-open "https://www.docker.com/products/docker-desktop/"
      fi
    fi
    echo "Please install Docker and re-run this script."
    exit 1
  fi

  if ! docker info >/dev/null 2>&1; then
    echo "⚠️  Docker daemon is not running. Attempting to start Docker..."
    started=false

    # Windows (Docker Desktop)
    if [ -f "/c/Program Files/Docker/Docker/Docker Desktop.exe" ]; then
      powershell.exe -Command "Start-Process 'C:\Program Files\Docker\Docker\Docker Desktop.exe'" >/dev/null 2>&1 &
      started=true
    elif [ -n "$PROGRAMFILES" ] && [ -f "$PROGRAMFILES/Docker/Docker/Docker Desktop.exe" ]; then
      powershell.exe -Command "Start-Process '$PROGRAMFILES\Docker\Docker\Docker Desktop.exe'" >/dev/null 2>&1 &
      started=true
    elif command -v powershell.exe >/dev/null 2>&1; then
      powershell.exe -Command "Start-Process 'Docker Desktop'" >/dev/null 2>&1 || true
      started=true
    # macOS
    elif [ -d "/Applications/Docker.app" ]; then
      open -a Docker
      started=true
    # Linux systemd
    elif command -v systemctl >/dev/null 2>&1; then
      sudo systemctl start docker || true
      started=true
    fi

    echo "⏳ Waiting for Docker daemon to become ready..."
    local retries=45
    while ! docker info >/dev/null 2>&1; do
      sleep 2
      retries=$((retries - 1))
      if [ $retries -le 0 ]; then
        echo "❌ Docker daemon failed to start in time. Please launch Docker Desktop manually and re-run."
        exit 1
      fi
      printf "."
    done
    echo ""
    echo "✅ Docker is running."
  fi
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
    up|down|setup|migrate)
      ensure_docker_running
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
        $DOCKER_COMPOSE exec -T db psql -U postgres -c "CREATE DATABASE \"$DB_NAME\";" >/dev/null 2>&1 || true
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
          $DOCKER_COMPOSE exec -T db psql "$SVC_DB_URL" -f - < "$f" >/dev/null || \
          echo "  WARN: could not run $(basename "$f") — check DB_URL"
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
        $DOCKER_COMPOSE exec -T db psql -U postgres -c "CREATE DATABASE \"$DB_NAME\";" >/dev/null 2>&1 || true
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
          $DOCKER_COMPOSE exec -T db psql "$SVC_DB_URL" -f - < "$f" >/dev/null || \
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
