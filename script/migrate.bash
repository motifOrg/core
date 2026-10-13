#!/usr/bin/env bash

set -e

if [ -f .env ]; then
    export $(grep -v '^#' .env | xargs)
fi

if [ -z "$DATABASE_URL" ]; then
    DB_USER=${DB_USER:-postgres}
    DB_PASSWORD=${DB_PASSWORD:-password}
    DB_HOST=${DB_HOST:-localhost}
    DB_PORT=${DB_PORT:-5432}
    DB_NAME=${DB_NAME:-motif}
    DB_SSLMODE=${DB_SSLMODE:-disable}

    DATABASE_URL="postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSLMODE}"
fi

ACTION=${1:-up}

case "$ACTION" in
    up)
        migrate \
            -path internal/migrations \
            -database "$DATABASE_URL" \
            up
        ;;
    down)
        migrate \
            -path internal/migrations \
            -database "$DATABASE_URL" \
            down 1
        ;;
    version)
        migrate \
            -path internal/migrations \
            -database "$DATABASE_URL" \
            version
        ;;
    force)
        if [ -z "$2" ]; then
            echo "Usage: ./script/migrate.bash force <version>"
            exit 1
        fi

        migrate \
            -path internal/migrations \
            -database "$DATABASE_URL" \
            force "$2"
        ;;
    *)
        echo "Usage: ./script/migrate.bash [up|down|version|force]"
        exit 1
        ;;
esac
