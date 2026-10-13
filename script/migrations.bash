#!/bin/bash

set -e

if [ -z "$1" ]; then
    echo "Usage: ./script/migrations.bash <name>"
    exit 1
fi

NAME=$(echo "$1" | tr ' ' '_' | tr '[:upper:]' '[:lower:]')

mkdir -p internal/migrations

LAST=$(find internal/migrations -name "*.up.sql" \
    | sed -E 's#.*/([0-9]+)_.*#\1#' \
    | sort -n \
    | tail -1)

if [ -z "$LAST" ]; then
    NEXT="000001"
else
    NEXT=$(printf "%06d" $((10#$LAST + 1)))
fi

UP="internal/migrations/${NEXT}_${NAME}.up.sql"
DOWN="internal/migrations/${NEXT}_${NAME}.down.sql"

touch "$UP" "$DOWN"

echo "Created:"
echo "  $UP"
echo "  $DOWN"