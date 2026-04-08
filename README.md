# hackerrank-cli

CLI for the [HackerRank for Work API](https://www.hackerrank.com/work/apidocs).

## Install

```bash
brew install jholm117/tap/hr
```

Or download from [Releases](https://github.com/jholm117/hackerrank-cli/releases).

## Auth

```bash
hr auth login    # prompts for API token
hr auth status   # show current auth
```

Generate a token at HackerRank Settings → API.

You can also set `HACKERRANK_API_TOKEN` or pass `--token`.

## Usage

```bash
# List tests
hr tests list

# List candidates for a test
hr candidates list --test <test-id>

# Get candidate source code
hr candidates code <test-id> <candidate-id>

# Save code to files
hr candidates code <test-id> <candidate-id> --save ./submissions

# List interviews
hr interviews list

# Get interview transcript
hr interviews transcript <interview-id>
```

All commands support `--output json` for machine-readable output.

## Development

```bash
make build       # build binary
make test        # run tests
make lint        # run golangci-lint
make setup-hooks # install pre-push hook
```
