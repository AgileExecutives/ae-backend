#!/usr/bin/env bash
set -euo pipefail

# Aggregate test coverage across all Go modules listed in go.work.
# Output:
#   test_results/coverage/all.out   (merged coverprofile)
#   test_results/coverage/all.html  (human-readable HTML report)

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

cd "$REPO_ROOT"

OUT_DIR="$REPO_ROOT/test_results/coverage"
MODULES_DIR="$OUT_DIR/modules"
mkdir -p "$MODULES_DIR"

# Get module dirs from go.work in a way that doesn't require jq.
# go.work format:
#   use (
#     ./base-server
#     ...
#   )
MODULE_DIRS=()
while IFS= read -r module_dir; do
  MODULE_DIRS+=("$module_dir")
done < <(
  awk '
    $1=="use" { inuse=1; next }
    inuse && $1=="(" { next }
    inuse && $1==")" { inuse=0; next }
    inuse { print $1 }
  ' go.work | sed 's/\r$//' | sed '/^$/d'
)

if [[ "${#MODULE_DIRS[@]}" -eq 0 ]]; then
  echo "No modules found in go.work" >&2
  exit 1
fi

echo "Running coverage for ${#MODULE_DIRS[@]} module(s)..."

# Run coverage per module. We write coverprofiles into a single folder, one per module.
# covermode=atomic is safest for concurrency.
for module in "${MODULE_DIRS[@]}"; do
  # Normalize ./foo into foo for a stable filename.
  module_rel="${module#./}"
  module_rel="${module_rel%/}"
  module_dir="$REPO_ROOT/${module_rel}"

  if [[ ! -d "$module_dir" ]]; then
    echo "Skipping missing module dir: $module" >&2
    continue
  fi

  # Replace slashes with __ for filename.
  module_slug="${module_rel//\//__}"
  profile="$MODULES_DIR/${module_slug}.out"

  echo "- $module"
  (
    cd "$module_dir"
    # -coverpkg=./... instruments ALL packages in the module, not just the one
    # being directly tested, so cross-package tests (e.g. tests/ -> services/)
    # contribute coverage to the packages they exercise.
    go test ./... -covermode=atomic -coverprofile="$profile" -coverpkg=./...
  )
done

MERGED="$OUT_DIR/all.out"
HTML="$OUT_DIR/all.html"

# Merge all module profiles.
# (Only include .out files we generated.)
PROFILES=()
while IFS= read -r profile; do
  PROFILES+=("$profile")
done < <(find "$MODULES_DIR" -maxdepth 1 -type f -name '*.out' | sort)

if [[ "${#PROFILES[@]}" -eq 0 ]]; then
  echo "No coverprofiles were generated." >&2
  exit 1
fi

go run ./workspace/cmd/covermerge -out "$MERGED" "${PROFILES[@]}"

go tool cover -func="$MERGED" | tail -n 1

go tool cover -html="$MERGED" -o "$HTML"

echo "Merged profile: $MERGED"
echo "HTML report:    $HTML"
