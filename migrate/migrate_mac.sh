#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")"/.. && pwd)"
ENV_FILE="$ROOT_DIR/.env"

if [[ ! -f "$ENV_FILE" ]]; then
  echo "env file not found: $ENV_FILE" >&2
  exit 1
fi

set -o allexport
# shellcheck disable=SC1090
source "$ENV_FILE"
set +o allexport

if ! command -v goose >/dev/null 2>&1; then
  echo "goose binary not found; install with: go install github.com/pressly/goose/v3/cmd/goose@latest" >&2
  exit 1
fi

build_dsn() {
  local user="$1" pass="$2" host="$3" port="$4" db="$5"
  printf "postgres://%s:%s@%s:%s/%s?sslmode=disable" "$user" "$pass" "$host" "$port" "$db"
}

run_migration() {
  local name="$1" dir="$2" dsn="$3"
  if [[ ! -d "$dir" ]]; then
    echo "skip $name: dir not found ($dir)" >&2
    return
  fi

  if [[ -z "$(find "$dir" -maxdepth 1 -type f -name '*.sql' -print -quit)" ]]; then
    echo "skip $name: no migration files in $dir" >&2
    return
  fi

  echo "==> running goose up for $name"
  goose -dir "$dir" postgres "$dsn" up
}

run_migration "init-scenario" \
  "$ROOT_DIR/SAGA/init_scenario_api/migrations" \
  "$(build_dsn \
      "${INIT_SCENARIO_POSTGRES_USER:-postgres}" \
      "${INIT_SCENARIO_POSTGRES_PASSWORD:-postgres}" \
      "${INIT_SCENARIO_POSTGRES_HOST:-localhost}" \
      "${INIT_SCENARIO_POSTGRES_PORT:-5432}" \
      "${INIT_SCENARIO_POSTGRES_DB:-postgres}")"

run_migration "runner" \
  "$ROOT_DIR/RTSP_PROCESSING/runner/migrations" \
  "$(build_dsn \
      "${RUNNER_POSTGRES_USER:-postgres}" \
      "${RUNNER_POSTGRES_PASSWORD:-postgres}" \
      "${RUNNER_POSTGRES_HOST:-localhost}" \
      "${RUNNER_POSTGRES_PORT:-5432}" \
      "${RUNNER_POSTGRES_DB:-postgres}")"

run_migration "runner-scheduler" \
  "$ROOT_DIR/SAGA/runner_scheduler/migrations" \
  "$(build_dsn \
      "${RUNNER_SCHEDULER_POSTGRES_USER:-postgres}" \
      "${RUNNER_SCHEDULER_POSTGRES_PASSWORD:-postgres}" \
      "${RUNNER_SCHEDULER_POSTGRES_HOST:-localhost}" \
      "${RUNNER_SCHEDULER_POSTGRES_MASTER_PORT:-5432}" \
      "${RUNNER_SCHEDULER_POSTGRES_DB:-runner_scheduler}")"

