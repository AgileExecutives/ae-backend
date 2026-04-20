#!/usr/bin/env bash
set -euo pipefail

# Run all tests across the Go workspace from the repo root.
# -count=1 disables test caching so the run is always visible.
# -v shows per-module progress from the workspace runner.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

cd "$REPO_ROOT"

go test -count=1 -v ./...
