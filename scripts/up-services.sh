#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

mkdir -p "$ROOT_DIR/data/sqlite" "$ROOT_DIR/data/minio"

if [[ -f "$ROOT_DIR/deploy/.env" ]]; then
  docker compose --env-file "$ROOT_DIR/deploy/.env" -f "$ROOT_DIR/deploy/docker-compose.yml" up -d --build
else
  docker compose -f "$ROOT_DIR/deploy/docker-compose.yml" up -d --build
fi
