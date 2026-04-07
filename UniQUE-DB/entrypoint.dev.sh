#!/bin/sh
set -eu

MIGRATE_BIN=/usr/local/bin/migrate
MIGRATIONS_DIR=${MIGRATIONS_DIR:-/migrations}

# If no action provided, default to 'up'. Use 'set --' to preserve multiple args.
if [ "$#" -eq 0 ]; then
  set -- up
fi

if [ -z "${DATABASE_URL:-}" ]; then
  echo "DATABASE_URL is required, e.g. mysql://user:pass@tcp(host:3306)/dbname?multiStatements=true"
  exit 1
fi

MAX_RETRIES=${MAX_RETRIES:-60}
SLEEP=${SLEEP:-2}
i=0

until "$MIGRATE_BIN" -path "$MIGRATIONS_DIR" -database "$DATABASE_URL" "$@"; do
  i=$((i+1))
  echo "migrate failed or DB not ready, retrying in ${SLEEP}s... ($i/$MAX_RETRIES)"
  if [ "$i" -ge "$MAX_RETRIES" ]; then
    echo "max retries reached, aborting"
    exit 1
  fi
  sleep "$SLEEP"
done

echo "migrations applied"
exit 0
