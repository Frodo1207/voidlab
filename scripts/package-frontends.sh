#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUTPUT_DIR="$ROOT_DIR/dist-packages"

mkdir -p "$OUTPUT_DIR"

"$ROOT_DIR/scripts/build-frontends.sh"

tar -czf "$OUTPUT_DIR/web-dist.tar.gz" -C "$ROOT_DIR/apps/web/dist" .
tar -czf "$OUTPUT_DIR/admin-dist.tar.gz" -C "$ROOT_DIR/apps/admin/dist" .

echo "==> web package: $OUTPUT_DIR/web-dist.tar.gz"
echo "==> admin package: $OUTPUT_DIR/admin-dist.tar.gz"
