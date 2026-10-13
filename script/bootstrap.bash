#!/usr/bin/env bash

set -e

cd "$(dirname "$0")/.."

if [ -f .env ]; then
    export $(grep -v '^#' .env | xargs)
fi

DB_NAME=${DB_NAME:-motif}
DB_USER=${DB_USER:-postgres}
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
export PGPASSWORD=${DB_PASSWORD:-}

echo "Seeding database '${DB_NAME}' as user '${DB_USER}'..."

psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" \
    -v ON_ERROR_STOP=1 -q -f script/seed.sql

echo "Bootstrap complete."
