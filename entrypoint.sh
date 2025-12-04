#!/bin/bash
set -e

echo "Running database migrations..."
greener-migration -u "$GREENER_DATABASE_URL"

echo "Starting services..."
exec "$@"
