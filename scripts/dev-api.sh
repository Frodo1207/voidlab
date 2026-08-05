#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

mkdir -p "$ROOT_DIR/data/sqlite"

cd "$ROOT_DIR/apps/api"
DB_PATH="$ROOT_DIR/data/sqlite/voidlab.db" \
MINIO_ENDPOINT="http://127.0.0.1:9000" \
go run ./cmd/server
