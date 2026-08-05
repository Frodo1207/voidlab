#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

docker build \
  -t voidlab-api:latest \
  -f "$ROOT_DIR/apps/api/Dockerfile" \
  "$ROOT_DIR"
