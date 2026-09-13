#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SERVICES_DIR="$SCRIPT_DIR/services"

# Load .env without exporting unrelated shell variables.
if [[ -f "$SCRIPT_DIR/.env" ]]; then
  set -a
  # shellcheck disable=SC1091
  source "$SCRIPT_DIR/.env"
  set +a
fi

ALL_SERVICES=()
if [[ -d "$SERVICES_DIR" ]]; then
  for service_dir in "$SERVICES_DIR"/*; do
    [[ -d "$service_dir" ]] && ALL_SERVICES+=("$(basename "$service_dir")")
  done
fi

usage() {
  cat <<EOF
Usage: $0 {up|down|build|setup|migrate|gqlgen} [options]

Commands:
  up          Start services via Docker or Podman Compose
  down        Stop services via Docker or Podman Compose
  build       Build Go binaries and restart selected containers
  setup       Reset containers, run migrations, compose the supergraph, and start services
  migrate     Run migrations only
  gqlgen      Generate GraphQL models and recompose the supergraph

Options:
EOF
  for service in "${ALL_SERVICES[@]}"; do
    printf '  -%-10s Target only %s service\n' "$service" "$service"
  done
  echo
  echo 'Migration forms: migrate [-service] | migrate [-service] VERSION up|down'
  echo 'If no service flag is given, the command applies to all services.'
  exit 1
}

CMD="${1:-}"
[[ -n "$CMD" ]] || usage
shift

TARGETS=()
MIGRATION_VERSION=''
MIGRATION_ACTION=''
for arg in "$@"; do
  if [[ "$CMD" == migrate && "$arg" == '-latest' ]]; then
    echo "-latest is not needed; migrate applies all pending up migrations." >&2
    usage
  elif [[ "$CMD" == migrate && "$arg" != -* && -z "$MIGRATION_VERSION" ]]; then
    MIGRATION_VERSION="$arg"
  elif [[ "$CMD" == migrate && "$arg" != -* && -n "$MIGRATION_VERSION" && -z "$MIGRATION_ACTION" ]]; then
    MIGRATION_ACTION="$arg"
  elif [[ "$arg" == -* ]]; then
    service_name="${arg#-}"
    found=false
    for service in "${ALL_SERVICES[@]}"; do
      if [[ "$service" == "$service_name" ]]; then
        TARGETS+=("$service_name")
        found=true
        break
      fi
    done
    if [[ "$found" == false ]]; then
      echo "Unknown option: $arg" >&2
      usage
    fi
  else
    echo "Unexpected argument: $arg" >&2
    usage
  fi
done

if [[ "$CMD" == migrate && -n "$MIGRATION_VERSION" && ! "$MIGRATION_VERSION" =~ ^[0-9]+$ ]]; then
  echo "Migration version must be numeric." >&2
  usage
fi
if [[ "$CMD" == migrate && -n "$MIGRATION_VERSION" && "$MIGRATION_ACTION" != 'up' && "$MIGRATION_ACTION" != 'down' ]]; then
  echo "Migration action must be 'up' or 'down'." >&2
  usage
fi
if [[ "$CMD" == migrate && -z "$MIGRATION_VERSION" ]]; then
  MIGRATION_ACTION='latest'
fi

[[ ${#TARGETS[@]} -gt 0 ]] || TARGETS=("${ALL_SERVICES[@]}")

case "$CMD" in
  up|down|build|setup|migrate|gqlgen) ;;
  *) echo "Unknown command: $CMD" >&2; usage ;;
esac

prompt_confirm() {
  local prompt_message="$1"
  read -r -p "$prompt_message [Y/n]: " response
  [[ ! "$response" =~ ^[nN]([oO])?$ ]]
}

COMPOSE=()
ensure_container_running() {
  if command -v docker >/dev/null 2>&1 && docker info >/dev/null 2>&1; then
    if docker compose version >/dev/null 2>&1; then
      COMPOSE=(docker compose)
    elif command -v docker-compose >/dev/null 2>&1; then
      COMPOSE=(docker-compose)
    else
      echo 'Docker is running, but Docker Compose is not installed.' >&2
      exit 1
    fi
    echo "Docker is running (${COMPOSE[*]})."
    return
  fi

  if command -v podman >/dev/null 2>&1 && podman info >/dev/null 2>&1; then
    if podman compose version >/dev/null 2>&1; then
      COMPOSE=(podman compose)
    elif command -v podman-compose >/dev/null 2>&1; then
      COMPOSE=(podman-compose)
    else
      echo 'Podman is running, but Podman Compose is not installed.' >&2
      exit 1
    fi
    echo "Podman is running (${COMPOSE[*]})."
    return
  fi

  echo 'No active Docker or Podman runtime found.' >&2
  echo "Start Docker Desktop or run 'podman machine start', then try again." >&2
  exit 1
}

ensure_rover() {
  export PATH="$HOME/.rover/bin:$PATH"
  if command -v rover >/dev/null 2>&1; then
    return
  fi

  echo "Apollo Rover CLI ('rover') is required but was not found."
  if ! prompt_confirm 'Would you like to install Apollo Rover now?'; then
    echo 'Skipping Rover installation. Supergraph composition may fail.'
    return
  fi

  if command -v curl >/dev/null 2>&1; then
    curl -sSL https://rover.apollo.dev/nix/latest | sh
  elif command -v npm >/dev/null 2>&1; then
    npm install -g @apollo/rover
  else
    echo 'Neither curl nor npm is installed; install Rover manually:' >&2
    echo 'https://www.apollographql.com/docs/rover/getting-started' >&2
    exit 1
  fi

  export PATH="$HOME/.rover/bin:$PATH"
  if ! command -v rover >/dev/null 2>&1; then
    echo 'Rover was installed, but is not on PATH. Restart the shell and try again.' >&2
    exit 1
  fi
}

ensure_go() {
  if command -v go >/dev/null 2>&1; then
    return
  fi

  echo "Go is not installed, but is required for '$CMD'." >&2
  if prompt_confirm 'Open the Go download page?'; then
    if command -v open >/dev/null 2>&1; then
      open 'https://go.dev/dl/'
    elif command -v xdg-open >/dev/null 2>&1; then
      xdg-open 'https://go.dev/dl/'
    else
      echo 'Open https://go.dev/dl/ in a browser to install Go.'
    fi
  fi
  exit 1
}

get_db_url() {
  local service="$1"
  local variable_name
  variable_name="$(echo "$service" | tr '[:lower:]' '[:upper:]')DB_URL"
  local service_db_url="${!variable_name:-}"
  echo "${service_db_url:-${DB_URL:-}}"
}

get_db_name() {
  local url_without_query="${1%%\?*}"
  echo "${url_without_query##*/}"
}

wait_for_db() {
  echo "Waiting for database to be ready..."
  until "${COMPOSE[@]}" exec -T db pg_isready -U postgres >/dev/null 2>&1 && \
        "${COMPOSE[@]}" exec -T db psql -U postgres -d postgres -c "SELECT 1;" >/dev/null 2>&1; do
    sleep 1
  done
  echo "Database is ready."
}

ensure_databases() {
  for service in "${TARGETS[@]}"; do
    local service_db_url db_name
    service_db_url="$(get_db_url "$service")"
    db_name="$(get_db_name "$service_db_url")"
    if [[ -n "$db_name" && "$db_name" != postgres ]]; then
      if ! "${COMPOSE[@]}" exec -T db psql -U postgres -d postgres -Atc "SELECT 1 FROM pg_database WHERE datname = '$db_name';" | grep -q 1; then
        "${COMPOSE[@]}" exec -T db psql -U postgres -d postgres -v ON_ERROR_STOP=1 -c "CREATE DATABASE \"$db_name\";" >/dev/null
      fi
    fi
  done
}

ensure_migration_table() {
  local service_db_url="$1"
  "${COMPOSE[@]}" exec -T db psql "$service_db_url" -v ON_ERROR_STOP=1 -c '
    CREATE TABLE IF NOT EXISTS schema_migrations (
      version TEXT PRIMARY KEY,
      dirty BOOLEAN NOT NULL DEFAULT false,
      applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
    );
  ' >/dev/null
}

find_migration() {
  local migration_dir="$1" version="$2" migration
  for migration in "$migration_dir/${version}"_*.up.sql; do
    [[ -f "$migration" ]] && { printf '%s\n' "$migration"; return 0; }
  done
  return 1
}

apply_migration() {
  local service="$1" service_db_url="$2" migration="$3" filename version state
  filename="$(basename "$migration")"
  version="${service}_${filename%%_*}"
  state="$("${COMPOSE[@]}" exec -T db psql "$service_db_url" -Atc "SELECT CASE WHEN dirty THEN 'dirty' ELSE 'clean' END FROM schema_migrations WHERE version = '$version';" | tr -d '\r')"

  if [[ "$state" == 'clean' ]]; then
    echo "  Skipping $filename (already applied)"
    return 0
  fi
  if [[ "$state" == 'dirty' ]]; then
    echo "  ⚠️  $filename is marked dirty; auto-healing..."
    local down_migration="${migration%.up.sql}.down.sql"
    if [[ -f "$down_migration" ]]; then
      echo "  Rolling back $filename using $(basename "$down_migration")..."
      "${COMPOSE[@]}" exec -T db psql "$service_db_url" -v ON_ERROR_STOP=1 -f - < "$down_migration" >/dev/null 2>&1 || true
    else
      echo "  No down migration found for $filename; resetting dirty state directly..."
    fi
    "${COMPOSE[@]}" exec -T db psql "$service_db_url" -v ON_ERROR_STOP=1 -c "DELETE FROM schema_migrations WHERE version = '$version';" >/dev/null
    echo "  Dirty state cleared for $version. Retrying application..."
  fi

  echo "  Applying $filename..."
  "${COMPOSE[@]}" exec -T db psql "$service_db_url" -v ON_ERROR_STOP=1 -c "INSERT INTO schema_migrations (version, dirty) VALUES ('$version', true) ON CONFLICT (version) DO UPDATE SET dirty = true;" >/dev/null
  if ! "${COMPOSE[@]}" exec -T db psql "$service_db_url" -v ON_ERROR_STOP=1 -f - < "$migration" >/dev/null; then
    echo "  ERROR: $filename failed and remains marked dirty" >&2
    return 1
  fi
  "${COMPOSE[@]}" exec -T db psql "$service_db_url" -v ON_ERROR_STOP=1 -c "UPDATE schema_migrations SET dirty = false, applied_at = NOW() WHERE version = '$version';" >/dev/null
}

rollback_migration() {
  local service="$1" service_db_url="$2" migration="$3" filename version state down_migration
  filename="$(basename "$migration")"
  version="${service}_${filename%%_*}"
  state="$("${COMPOSE[@]}" exec -T db psql "$service_db_url" -Atc "SELECT CASE WHEN dirty THEN 'dirty' ELSE 'clean' END FROM schema_migrations WHERE version = '$version';" | tr -d '\r')"
  if [[ "$state" != 'clean' ]]; then
    echo "  ERROR: $filename is not applied cleanly" >&2
    return 1
  fi

  down_migration="${migration%.sql}.down.sql"
  if [[ ! -f "$down_migration" ]]; then
    echo "  ERROR: no rollback file for $filename ($down_migration)" >&2
    return 1
  fi

  echo "  Rolling back $filename..."
  "${COMPOSE[@]}" exec -T db psql "$service_db_url" -v ON_ERROR_STOP=1 -c "UPDATE schema_migrations SET dirty = true WHERE version = '$version';" >/dev/null
  if ! "${COMPOSE[@]}" exec -T db psql "$service_db_url" -v ON_ERROR_STOP=1 -f - < "$down_migration" >/dev/null; then
    echo "  ERROR: rollback for $filename failed and remains marked dirty" >&2
    return 1
  fi
  "${COMPOSE[@]}" exec -T db psql "$service_db_url" -v ON_ERROR_STOP=1 -c "DELETE FROM schema_migrations WHERE version = '$version';" >/dev/null
}

run_migrations() {
  for service in "${TARGETS[@]}"; do
    echo "--- $service migrations ---"
    local migration_dir="$SERVICES_DIR/$service/migrations"
    if [[ ! -d "$migration_dir" ]]; then
      echo "  No migrations directory found for $service"
      continue
    fi

    local service_db_url
    service_db_url="$(get_db_url "$service")"
    ensure_migration_table "$service_db_url"

    if [[ "$CMD" != 'migrate' || -z "$MIGRATION_VERSION" || "$MIGRATION_ACTION" == 'latest' ]]; then
      for migration in "$migration_dir"/*.up.sql; do
        [[ -f "$migration" ]] || continue
        apply_migration "$service" "$service_db_url" "$migration" || return 1
      done
    else
      local migration
      migration="$(find_migration "$migration_dir" "$MIGRATION_VERSION")" || {
        echo "  ERROR: migration version $MIGRATION_VERSION not found for $service" >&2
        return 1
      }
      if [[ "$MIGRATION_ACTION" == 'up' ]]; then
        apply_migration "$service" "$service_db_url" "$migration" || return 1
      else
        rollback_migration "$service" "$service_db_url" "$migration" || return 1
      fi
    fi
  done
}

compose_supergraph() {
  rover supergraph compose \
    --config "$SCRIPT_DIR/supergraph-config.yaml" \
    > "$SCRIPT_DIR/rover/supergraph.graphql"
}

generate_gqlgen() {
  for service in "${TARGETS[@]}"; do
    service_dir="$SERVICES_DIR/$service"
    if [[ -f "$service_dir/gqlgen.yml" ]]; then
      echo "--- Generating gqlgen for $service ---"
      (cd "$service_dir" && go run github.com/99designs/gqlgen generate)
    fi
  done
}

case "$CMD" in
  up|down|build|setup|migrate)
    ensure_container_running
    ;;
esac
case "$CMD" in
  build|setup|gqlgen) ensure_rover ;;
esac
case "$CMD" in
  build|gqlgen) ensure_go ;;
esac

case "$CMD" in
  up)
    "${COMPOSE[@]}" up -d
    ;;
  down)
    "${COMPOSE[@]}" down
    ;;
  build)
    generate_gqlgen
    compose_supergraph
    for service in "${TARGETS[@]}"; do
      echo "Building $service..."
      (cd "$SCRIPT_DIR" && go build -o "bin/$service" "./services/$service")
      "${COMPOSE[@]}" up -d --build "$service"
    done
    "${COMPOSE[@]}" up -d --force-recreate router
    ;;
  setup)
    if [[ -d "$SCRIPT_DIR/.git" && -d "$SCRIPT_DIR/.githooks" ]]; then
      git -C "$SCRIPT_DIR" config core.hooksPath .githooks >/dev/null 2>&1 || true
    fi
    "${COMPOSE[@]}" down -v
    "${COMPOSE[@]}" up -d db redis
    wait_for_db
    ensure_databases
    run_migrations
    compose_supergraph
    "${COMPOSE[@]}" up -d --build
    ;;
  migrate)
    "${COMPOSE[@]}" up -d db
    wait_for_db
    ensure_databases
    run_migrations
    ;;
  gqlgen)
    generate_gqlgen
    compose_supergraph
    ;;
esac
