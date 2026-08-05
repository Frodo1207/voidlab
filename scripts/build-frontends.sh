#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

cd "$ROOT_DIR"

npm run build:web
npm run build:admin

echo "==> web output: $ROOT_DIR/apps/web/dist"
echo "==> admin output: $ROOT_DIR/apps/admin/dist"
