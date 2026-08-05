#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUTPUT_DIR="$ROOT_DIR/dist-packages"

mkdir -p "$OUTPUT_DIR"

"$ROOT_DIR/scripts/build-api-image.sh"

docker save -o "$OUTPUT_DIR/voidlab-api.tar" voidlab-api:latest

echo "==> docker image package: $OUTPUT_DIR/voidlab-api.tar"
