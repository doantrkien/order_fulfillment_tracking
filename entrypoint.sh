#!/bin/sh
set -e

echo "Running migrations..."
/app_migrate

echo "Starting application..."
exec /app_service
