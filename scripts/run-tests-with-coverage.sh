#!/bin/bash
set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🧪 Running tests with coverage...${NC}"
echo ""

# Create test results directory
ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
RESULTS_DIR="$ROOT_DIR/test_results"
mkdir -p "$RESULTS_DIR"

GO_WORK_FILE="$ROOT_DIR/go.work"
if [[ ! -f "$GO_WORK_FILE" ]]; then
    echo -e "${RED}❌ go.work not found at ${GO_WORK_FILE}${NC}"
    exit 1
fi

extract_modules_from_go_work() {
    # Extract paths from the `use ( ... )` section of go.work.
    awk '
        BEGIN { inuse=0 }
        /^use[[:space:]]*\(/ { inuse=1; next }
        inuse && /^\)/ { inuse=0; next }
        inuse {
            gsub(/#.*/, "");
            gsub(/\/\/.*/, "");
            gsub(/^[[:space:]]+|[[:space:]]+$/, "");
            if (length($0) > 0) print $0;
        }
    ' "$GO_WORK_FILE"
}

MODULE_DIRS=()
while IFS= read -r line; do
    MODULE_DIRS+=("$line")
done < <(extract_modules_from_go_work)

if [[ ${#MODULE_DIRS[@]} -eq 0 ]]; then
    echo -e "${RED}❌ No modules found in go.work use() block${NC}"
    exit 1
fi

echo -e "${YELLOW}Workspace modules:${NC}"
for m in "${MODULE_DIRS[@]}"; do
    echo "  - $m"
done
echo ""

failed_modules=()
coverage_summary_tsv="$RESULTS_DIR/coverage_summary.tsv"
: > "$coverage_summary_tsv"

for module_rel in "${MODULE_DIRS[@]}"; do
    module_path="$ROOT_DIR/${module_rel#./}"
    if [[ ! -d "$module_path" ]]; then
        echo -e "${RED}❌ Skipping missing module dir: $module_rel${NC}"
        failed_modules+=("$module_rel")
        continue
    fi

    module_key="${module_rel#./}"
    module_key="${module_key//\//__}"
    module_results="$RESULTS_DIR/$module_key"
    mkdir -p "$module_results"

    echo -e "${YELLOW}Running tests: ${module_rel}${NC}"

    set +e
    (
        cd "$module_path" \
            && go test -v -race -coverprofile="$module_results/coverage.out" -covermode=atomic ./...
    ) 2>&1 | tee "$module_results/test_output.log"
    test_status=${PIPESTATUS[0]}
    set -e

    if [[ $test_status -ne 0 ]]; then
        echo -e "${RED}❌ Tests failed: ${module_rel}${NC}"
        failed_modules+=("$module_rel")
        echo ""
        continue
    fi

    if [[ -f "$module_results/coverage.out" ]]; then
        go tool cover -html="$module_results/coverage.out" -o "$module_results/coverage.html"
        go tool cover -func="$module_results/coverage.out" > "$module_results/coverage.txt"
        cov=$(go tool cover -func="$module_results/coverage.out" | grep total | awk '{print $3}')
        printf "%s\t%s\n" "$module_rel" "$cov" >> "$coverage_summary_tsv"
        echo -e "${GREEN}✅ Coverage (${module_rel}): ${cov}${NC}"
    else
        echo -e "${RED}❌ Coverage profile missing for ${module_rel}${NC}"
        failed_modules+=("$module_rel")
    fi

    echo ""
done

echo -e "${BLUE}📊 Coverage Summary (per module)${NC}"
for module_rel in "${MODULE_DIRS[@]}"; do
    cov=$(awk -v m="$module_rel" 'BEGIN{FS="\t"} $1==m {print $2}' "$coverage_summary_tsv")
    if [[ -n "$cov" ]]; then
        echo -e "${GREEN}${module_rel}: ${cov}${NC}"
    else
        echo -e "${RED}${module_rel}: (no coverage - failed or missing)${NC}"
    fi
done

echo ""
echo -e "${GREEN}✅ Coverage reports generated under:${NC} ${BLUE}test_results/<module>/${NC}"

if [[ ${#failed_modules[@]} -gt 0 ]]; then
    echo ""
    echo -e "${RED}❌ Some modules failed:${NC}"
    for m in "${failed_modules[@]}"; do
        echo "  - $m (see test_results/*/test_output.log)"
    done
    exit 1
fi

echo -e "${GREEN}✅ Done!${NC}"
