# HackerRank CLI Design Spec

## Overview

General-purpose CLI for the HackerRank for Work API. Resource-oriented command structure following modern CLI conventions (`gh`, `kubectl`). Written in Go, distributed via Homebrew tap.

## Command Structure

```
hr <resource> <action> [args] [flags]
```

### Auth

```
hr auth login          # Prompt for API token, save to config
hr auth status         # Show current auth state
hr auth logout         # Remove stored token
```

### Tests

```
hr tests list                              # List all tests
hr tests get <id>                          # Show test details
```

### Candidates

```
hr candidates list --test <id>             # List candidates for a test
hr candidates get <test-id> <candidate-id> # Full candidate detail
hr candidates code <test-id> <candidate-id> # Extract source code per question
```

### Interviews

```
hr interviews list                         # List interviews
hr interviews get <id>                     # Show interview details
hr interviews transcript <id>             # Get interview transcript
```

### Global Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--token` | Override API token | - |
| `--output json\|table` | Output format | `table` for lists, `json` for detail |
| `--no-color` | Disable color | false |

## Auth Precedence

1. `--token` flag (highest)
2. `HACKERRANK_API_TOKEN` env var
3. Config file `~/.config/hackerrank/config.yaml`

`hr auth login` prompts for the token and writes it to the config file.

## Project Layout

```
hackerrank-cli/
├── cmd/
│   ├── root.go           # Root command, global flags, auth
│   ├── auth.go           # auth login/logout/status
│   ├── tests.go          # tests list/get
│   ├── candidates.go     # candidates list/get/code
│   └── interviews.go     # interviews list/get/transcript
├── internal/
│   ├── api/
│   │   └── client.go     # HTTP client, pagination, error handling
│   ├── config/
│   │   └── config.go     # Config file read/write (~/.config/hackerrank/)
│   └── output/
│       └── output.go     # Table and JSON formatters
├── main.go
├── go.mod
├── Makefile
├── .goreleaser.yaml
└── .github/
    └── workflows/
        └── ci.yml
```

## API Client

- Base URL: `https://www.hackerrank.com/x/api/v3/`
- Auth: `Authorization: Bearer <token>` header
- Pagination: automatic for list endpoints using `offset`/`limit` params
- Error handling: structured errors from API response bodies

### Key Endpoints

| Command | Endpoint | Notes |
|---------|----------|-------|
| `tests list` | `GET /tests?limit=&offset=` | Paginated |
| `tests get` | `GET /tests/{id}` | |
| `candidates list` | `GET /tests/{test_id}/candidates?limit=&offset=` | Paginated |
| `candidates get` | `GET /tests/{test_id}/candidates/{id}?additional_fields=questions,questions.solves,questions.submission_result` | Includes scores, submissions |
| `candidates code` | Same as `candidates get` | Extracts `answer.code` from each question |
| `interviews list` | `GET /interviews` | |
| `interviews get` | `GET /interviews/{id}` | |
| `interviews transcript` | `GET /interviews/{id}/transcript` | |

## Output

### Table format (default for list commands)

Human-readable columns. Example for `hr tests list`:

```
ID        NAME                                          STATE   CANDIDATES
2324754   Scale AI Sr. SRE Hiring Test - Python          active  0
2309131   Scale AI TAI Technical Take Home               active  1
2233043   [UR] Scale AI ML Take Home 2025 V2             active  703
```

### JSON format

Raw API JSON piped to stdout. Useful for `jq` pipelines.

### Code output (`candidates code`)

Default: prints source code to stdout with headers per question.

```
## Question 1: The Galactic Trade Routes (Python 3) — 65/65
────────────────────────────────────────────────────────────
import sys
...

## Question 2: Data Center Module Decommissioning (Python 3) — 80/80
────────────────────────────────────────────────────────────
#!/bin/python3
...
```

`--save <dir>` flag writes individual files: `q1_galactic_trade_routes.py`, `q2_data_center_module_decommissioning.py`.

## Release & Distribution

- **GoReleaser**: builds darwin-arm64/amd64 + linux-arm64/amd64
- **GitHub Releases**: binaries + checksums on tag push
- **Homebrew**: auto-update formula in `jholm117/homebrew-tap`

## CI & Hooks

### Git Hooks

`.githooks/pre-push` runs `hack/ci-checks.sh` — same checks as CI (single source of truth). Install via `make setup-hooks` which sets `git config core.hooksPath .githooks`.

### CI (GitHub Actions)

`.github/workflows/ci.yml` — triggers on push to main and PRs. Cancels in-progress builds for the same ref.

Runs `hack/ci-checks.sh --parallel`:
1. `go mod tidy` check (ensures go.mod/go.sum are clean)
2. In parallel: `golangci-lint`, `go test`, `go vet`

### Linting

`.golangci.yml` with: errcheck, govet, staticcheck, ineffassign, misspell, gocyclo, unused. Formatters: gofmt + goimports.

## Dependencies

- `github.com/spf13/cobra` — CLI framework
- `github.com/fatih/color` — terminal colors
- `gopkg.in/yaml.v3` — config file parsing
- Standard library `net/http`, `encoding/json` for API calls
