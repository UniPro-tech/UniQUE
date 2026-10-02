#!/bin/sh
set -eu

: "${DATABASE_URL:?DATABASE_URL is required}"
MIGRATIONS_DIR=${MIGRATIONS_DIR:-/migrations}
MAX_RETRIES=${MAX_RETRIES:-60}
SLEEP=${SLEEP:-2}

if [ "$#" -eq 0 ]; then
  set -- up
fi

if [ "$1" = wait ]; then
  # Use the same schema target as the migration Job, including on upgrades.
  target=$(find "$MIGRATIONS_DIR" -name '*.up.sql' -exec basename {} \; | cut -d_ -f1 | sort -n | tail -1)
  if [ -z "$target" ]; then
    echo "No migrations found" >&2
    exit 1
  fi
fi

i=0
while [ "$i" -lt "$MAX_RETRIES" ]; do
  if [ "$1" = wait ]; then
    # migrate prints the version to stderr and appends '(dirty)' on failure.
    # Only a clean version at or beyond this image's target can start the app.
    if version=$(migrate -path "$MIGRATIONS_DIR" -database "$DATABASE_URL" version 2>&1); then
      case "$version" in
        ''|*[!0-9]*) ;;
        *) if [ "$version" -ge "$target" ]; then exit 0; fi ;;
      esac
    fi
  elif migrate -path "$MIGRATIONS_DIR" -database "$DATABASE_URL" "$@"; then
    exit 0
  fi
  i=$((i + 1))
  echo "Waiting for database migration ($i/$MAX_RETRIES)" >&2
  sleep "$SLEEP"
done

echo "Database migration did not become ready" >&2
exit 1
