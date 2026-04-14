#!/usr/bin/env sh
set -eu

APP_DIR="${APP_DIR:-/home/alpardfm/apps/haze-api}"
COMPOSE_PROJECT_NAME="${COMPOSE_PROJECT_NAME:-haze-api}"
COMPOSE="docker-compose -f docker-compose.prod.yml"
COMPOSE_DOCKER_CLI_BUILD=0
DOCKER_BUILDKIT=0
export COMPOSE_PROJECT_NAME COMPOSE_DOCKER_CLI_BUILD DOCKER_BUILDKIT

cd "$APP_DIR"

if [ ! -f .env.vps ]; then
  echo "missing .env.vps in $APP_DIR" >&2
  exit 1
fi

cp .env.vps .env

if [ -f docs/openapi.yaml ] && command -v sudo >/dev/null 2>&1 && sudo -n true 2>/dev/null; then
  sudo install -d -m 755 /var/www/haze-api
  sudo install -m 644 docs/openapi.yaml /var/www/haze-api/openapi.yaml

  if [ -f docs/swagger.html ]; then
    sudo install -m 644 docs/swagger.html /var/www/haze-api/swagger.html
  fi
fi

$COMPOSE build api
$COMPOSE up -d postgres
$COMPOSE run --rm migrate
$COMPOSE run --rm seed-admin

# docker-compose v1 can fail with KeyError 'ContainerConfig' when recreating
# containers against newer Docker image metadata. Remove only stateless app
# containers first; keep postgres and its named volume intact.
for service in api reminder-worker status-worker migrate seed-admin; do
  ids="$(docker ps -aq \
    --filter "label=com.docker.compose.project=$COMPOSE_PROJECT_NAME" \
    --filter "label=com.docker.compose.service=$service")"

  if [ -n "$ids" ]; then
    docker rm -f $ids
  fi
done

$COMPOSE up -d --no-deps --force-recreate api reminder-worker status-worker
$COMPOSE ps
