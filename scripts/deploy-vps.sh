#!/usr/bin/env sh
set -eu

APP_DIR="${APP_DIR:-/home/alpardfm/apps/haze-api}"
COMPOSE="docker-compose -f docker-compose.prod.yml"

cd "$APP_DIR"

if [ ! -f .env.vps ]; then
  echo "missing .env.vps in $APP_DIR" >&2
  exit 1
fi

cp .env.vps .env

if [ -f docs/openapi.yaml ] && command -v sudo >/dev/null 2>&1 && sudo -n true 2>/dev/null; then
  sudo install -d -m 755 /var/www/haze-api
  sudo install -m 644 docs/openapi.yaml /var/www/haze-api/openapi.yaml
fi

$COMPOSE build api
$COMPOSE up -d postgres
$COMPOSE run --rm migrate
$COMPOSE run --rm seed-admin
$COMPOSE up -d api reminder-worker status-worker
$COMPOSE ps
