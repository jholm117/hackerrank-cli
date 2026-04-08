#!/usr/bin/env bash
# hack/ci-checks.sh — single source of truth for CI and pre-push checks
set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

run_tidy_check() {
    echo "==> go mod tidy check..."
    cp go.sum go.sum.bak
    go mod tidy
    if ! diff -q go.sum go.sum.bak > /dev/null 2>&1; then
        echo -e "${RED}FAIL: go.mod/go.sum not tidy${NC}"
        rm go.sum.bak
        return 1
    fi
    rm go.sum.bak
    echo -e "${GREEN}OK${NC}"
}

run_lint() {
    echo "==> golangci-lint..."
    golangci-lint run
    echo -e "${GREEN}OK${NC}"
}

run_test() {
    echo "==> go test..."
    go test ./... -v
    echo -e "${GREEN}OK${NC}"
}

run_vet() {
    echo "==> go vet..."
    go vet ./...
    echo -e "${GREEN}OK${NC}"
}

if [[ "${1:-}" == "--parallel" ]]; then
    run_tidy_check

    pids=()
    run_lint & pids+=($!)
    run_test & pids+=($!)
    run_vet & pids+=($!)

    failed=0
    for pid in "${pids[@]}"; do
        if ! wait "$pid"; then
            failed=1
        fi
    done

    if [[ $failed -ne 0 ]]; then
        echo -e "${RED}Some checks failed${NC}"
        exit 1
    fi
else
    run_tidy_check
    run_vet
    run_lint
    run_test
fi

echo -e "${GREEN}All checks passed${NC}"
