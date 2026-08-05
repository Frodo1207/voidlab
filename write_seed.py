with open("scripts/seed-local-data.sh", "w") as f:
    f.write("""#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
DB_PATH="${1:-$ROOT_DIR/apps/api/data/sqlite/voidlab.db}"

if ! command -v sqlite3 >/dev/null 2>&1; then
  echo "sqlite3 is required to seed local data" >&2
  exit 1
fi

if [ ! -f "$DB_PATH" ]; then
  echo "Database not found: $DB_PATH" >&2
  echo "Start the API once to bootstrap the local SQLite schema." >&2
  exit 1
fi

require_table() {
  local table_name="$1"
  if ! sqlite3 "$DB_PATH" "SELECT 1 FROM sqlite_master WHERE type = 'table' AND name = '$table_name';" | grep -q 1; then
    echo "Missing table '$table_name' in $DB_PATH" >&2
    echo "Start the API once so migrations can create the latest schema." >&2
    exit 1
  fi
}

require_table "articles"
require_table "events"
require_table "builders"
require_table "knowledge_spaces"
require_table "knowledge_entries"
require_table "kwith open("scripts/seed-local-data.sh", "w") as f:
 es    f.write("""#!/usr/bin/env bash

set -euo pipebu
set -euo pipefail

ROOT_DIR="$(cl-k
ROOT_DIR="$(cd obaDB_PATH="${1:-$ROOT_DIR/apps/api/data/sqlit"$
if ! command -v sqlite3 >/dev/null 2>&1; then
  echo "s ar  echo "sqlite3 is required to seed local daud  exit 1
fi

if [ ! -f "$DB_PATH" ]; then
  echo "e,fi

if _u
l,   echo "Database not found:_a  echo "Start the API once to bootstrap " exit 1