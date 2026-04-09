# Changelog

## v0.1.2 (2026-04-08)

- Add `interviews code` command — extracts source code from interview pads
  via undocumented `/api/interviews/{id}/recordings/code` endpoint
- Add `GetRaw` method to API client for non-v3 endpoints

## v0.1.1 (2026-04-08)

- Fix: mask token input in `auth login` (no longer echoes to terminal)
- Add auth command tests (login, logout, status)
- Add `--base-url` hidden flag for test mocking

## v0.1.0 (2026-04-08)

Initial release.

- Commands: `auth`, `tests`, `candidates`, `interviews`
  - `auth login/logout/status` — token management
  - `tests list/get` — browse tests
  - `candidates list/get/code` — list candidates, view details, extract source code
  - `interviews list/get/transcript` — interview management
- Config-file auth with precedence: `--token` flag > `HACKERRANK_API_TOKEN` env > `~/.config/hackerrank/config.yaml`
- Table and JSON output formats (`--output json`)
- `candidates code --save <dir>` writes individual source files
- 24 tests across 4 packages
- GoReleaser + Homebrew tap (`brew install jholm117/tap/hr`)
- CI with golangci-lint, go vet, go test
- Pre-push hooks matching CI checks
