#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

echo "==> install frontend dependencies"
npm install --prefix "$ROOT_DIR"

echo "==> tidy go module"
cd "$ROOT_DIR/apps/api"
go mod tidy

echo "==> bootstrap complete"
