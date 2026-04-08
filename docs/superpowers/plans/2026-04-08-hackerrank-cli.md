# HackerRank CLI Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a general-purpose Go CLI for the HackerRank for Work API with resource-oriented commands, config-file auth, and Homebrew distribution.

**Architecture:** Cobra CLI with resource subcommands (auth, tests, candidates, interviews). Thin HTTP client wrapping the HackerRank v3 API with automatic pagination. Config stored in `~/.config/hackerrank/config.yaml`. Output as table (human) or JSON (machine).

**Tech Stack:** Go 1.22+, cobra, gopkg.in/yaml.v3, goreleaser, golangci-lint

---

## File Map

| File | Responsibility |
|------|---------------|
| `main.go` | Entry point, calls `cmd.Execute()` |
| `cmd/root.go` | Root command, global flags (`--token`, `--output`, `--no-color`), auth resolution |
| `cmd/auth.go` | `auth login`, `auth logout`, `auth status` commands |
| `cmd/tests.go` | `tests list`, `tests get` commands |
| `cmd/candidates.go` | `candidates list`, `candidates get`, `candidates code` commands |
| `cmd/interviews.go` | `interviews list`, `interviews get`, `interviews transcript` commands |
| `internal/api/client.go` | HTTP client, base URL, auth header, request helper, error handling |
| `internal/api/pagination.go` | Generic paginated list fetcher |
| `internal/api/types.go` | Go structs for API response objects (Test, Candidate, Interview, etc.) |
| `internal/config/config.go` | Read/write `~/.config/hackerrank/config.yaml`, token precedence logic |
| `internal/output/output.go` | Table printer and JSON printer |
| `Makefile` | build, test, lint, fmt, vet, setup-hooks targets |
| `hack/ci-checks.sh` | Unified CI/pre-push check script |
| `.githooks/pre-push` | Runs `hack/ci-checks.sh` |
| `.golangci.yml` | Linter configuration |
| `.github/workflows/ci.yml` | CI pipeline |
| `.github/workflows/release.yml` | GoReleaser on tag push |
| `.goreleaser.yaml` | Multi-platform build + homebrew tap |

---

### Task 1: Project Scaffolding

**Files:**
- Create: `main.go`
- Create: `cmd/root.go`
- Create: `go.mod`
- Create: `Makefile`
- Create: `.gitignore`

- [ ] **Step 1: Initialize Go module**

Run:
```bash
cd /path/to/hackerrank-cli
go mod init github.com/jholm117/hackerrank-cli
```

- [ ] **Step 2: Install cobra dependency**

Run:
```bash
go get github.com/spf13/cobra@latest
```

- [ ] **Step 3: Create main.go**

```go
// main.go
package main

import (
	"os"

	"github.com/jholm117/hackerrank-cli/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
```

- [ ] **Step 4: Create cmd/root.go**

```go
// cmd/root.go
package cmd

import (
	"github.com/spf13/cobra"
)

var (
	flagToken  string
	flagOutput string
	flagNoColor bool
)

var rootCmd = &cobra.Command{
	Use:   "hr",
	Short: "CLI for HackerRank for Work API",
	SilenceErrors: true,
	SilenceUsage:  true,
}

func init() {
	rootCmd.PersistentFlags().StringVar(&flagToken, "token", "", "API token (overrides config and env)")
	rootCmd.PersistentFlags().StringVarP(&flagOutput, "output", "o", "", "Output format: table or json")
	rootCmd.PersistentFlags().BoolVar(&flagNoColor, "no-color", false, "Disable color output")
}

func Execute() error {
	return rootCmd.Execute()
}
```

- [ ] **Step 5: Create .gitignore**

```
# .gitignore
/hr
/dist/
.worktrees/
```

- [ ] **Step 6: Create Makefile**

```makefile
# Makefile
BINARY := hr
GO := go

.PHONY: build test lint fmt vet clean setup-hooks

build:
	$(GO) build -o $(BINARY) .

test:
	$(GO) test ./... -v

lint:
	golangci-lint run

fmt:
	$(GO) fmt ./...
	goimports -w .

vet:
	$(GO) vet ./...

clean:
	rm -f $(BINARY)

setup-hooks:
	git config core.hooksPath .githooks
	@echo "Git hooks configured to use .githooks/"
```

- [ ] **Step 7: Verify it builds and runs**

Run:
```bash
go mod tidy && make build && ./hr --help
```
Expected: Help output showing `hr` usage with `--token`, `--output`, `--no-color` flags.

- [ ] **Step 8: Commit**

```bash
git add main.go cmd/root.go go.mod go.sum .gitignore Makefile
git commit -m "init: project scaffolding with root command"
```

---

### Task 2: Config Package

**Files:**
- Create: `internal/config/config.go`
- Create: `internal/config/config_test.go`

- [ ] **Step 1: Write failing test for config read/write**

```go
// internal/config/config_test.go
package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	cfg := &Config{Token: "test-token-123"}
	if err := Save(cfg, path); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.Token != "test-token-123" {
		t.Errorf("got token %q, want %q", loaded.Token, "test-token-123")
	}
}

func TestLoadMissing(t *testing.T) {
	cfg, err := Load("/nonexistent/path/config.yaml")
	if err != nil {
		t.Fatalf("Load missing file should not error: %v", err)
	}
	if cfg.Token != "" {
		t.Errorf("got token %q, want empty", cfg.Token)
	}
}

func TestResolveToken(t *testing.T) {
	tests := []struct {
		name     string
		flag     string
		env      string
		file     string
		want     string
	}{
		{"flag wins", "flag-token", "env-token", "file-token", "flag-token"},
		{"env wins over file", "", "env-token", "file-token", "env-token"},
		{"file fallback", "", "", "file-token", "file-token"},
		{"empty", "", "", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.env != "" {
				t.Setenv("HACKERRANK_API_TOKEN", tt.env)
			}
			cfg := &Config{Token: tt.file}
			got := ResolveToken(tt.flag, cfg)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/config/ -v`
Expected: Compilation error — `config` package doesn't exist yet.

- [ ] **Step 3: Implement config package**

```go
// internal/config/config.go
package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Token string `yaml:"token"`
}

// DefaultPath returns ~/.config/hackerrank/config.yaml
func DefaultPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "hackerrank", "config.yaml")
}

// Load reads config from path. Returns empty config if file doesn't exist.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// Save writes config to path, creating parent directories.
func Save(cfg *Config, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

// ResolveToken returns the token using precedence: flag > env > config file.
func ResolveToken(flag string, cfg *Config) string {
	if flag != "" {
		return flag
	}
	if env := os.Getenv("HACKERRANK_API_TOKEN"); env != "" {
		return env
	}
	return cfg.Token
}
```

- [ ] **Step 4: Install yaml dependency and run tests**

Run:
```bash
go get gopkg.in/yaml.v3@latest
go test ./internal/config/ -v
```
Expected: All 3 tests pass.

- [ ] **Step 5: Commit**

```bash
git add internal/config/ go.mod go.sum
git commit -m "feat: add config package with token precedence"
```

---

### Task 3: API Client

**Files:**
- Create: `internal/api/client.go`
- Create: `internal/api/types.go`
- Create: `internal/api/pagination.go`
- Create: `internal/api/client_test.go`

- [ ] **Step 1: Write failing test for API client**

```go
// internal/api/client_test.go
package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("bad auth header: %s", r.Header.Get("Authorization"))
		}
		if r.URL.Path != "/x/api/v3/tests/123" {
			t.Errorf("bad path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]string{"id": "123", "name": "Test"})
	}))
	defer server.Close()

	c := NewClient("test-token", WithBaseURL(server.URL))
	var result map[string]string
	if err := c.Get("/tests/123", nil, &result); err != nil {
		t.Fatalf("Get: %v", err)
	}
	if result["id"] != "123" {
		t.Errorf("got id %q, want %q", result["id"], "123")
	}
}

func TestClientGetError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message":"Invalid token"}`))
	}))
	defer server.Close()

	c := NewClient("bad-token", WithBaseURL(server.URL))
	var result map[string]string
	err := c.Get("/tests/123", nil, &result)
	if err == nil {
		t.Fatal("expected error for 401")
	}
}

func TestPaginatedList(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		offset := r.URL.Query().Get("offset")
		var resp map[string]interface{}
		if offset == "" || offset == "0" {
			resp = map[string]interface{}{
				"data":  []map[string]string{{"id": "1"}, {"id": "2"}},
				"total": 3,
			}
		} else {
			resp = map[string]interface{}{
				"data":  []map[string]string{{"id": "3"}},
				"total": 3,
			}
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := NewClient("tok", WithBaseURL(server.URL))
	items, err := Paginate[map[string]string](c, "/things", nil)
	if err != nil {
		t.Fatalf("Paginate: %v", err)
	}
	if len(items) != 3 {
		t.Errorf("got %d items, want 3", len(items))
	}
	if callCount != 2 {
		t.Errorf("got %d calls, want 2", callCount)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/api/ -v`
Expected: Compilation error — `api` package doesn't exist.

- [ ] **Step 3: Create API types**

```go
// internal/api/types.go
package api

// Test represents a HackerRank test.
type Test struct {
	ID        string   `json:"id"`
	UniqueID  string   `json:"unique_id"`
	Name      string   `json:"name"`
	Duration  int      `json:"duration"`
	State     string   `json:"state"`
	Draft     bool     `json:"draft"`
	CreatedAt string   `json:"created_at"`
	Questions []string `json:"questions"`
}

// Candidate represents a test candidate.
type Candidate struct {
	ID              string                    `json:"id"`
	Email           string                    `json:"email"`
	FullName        string                    `json:"full_name"`
	Score           float64                   `json:"score"`
	PercentageScore float64                   `json:"percentage_score"`
	Status          int                       `json:"status"`
	AttemptStart    string                    `json:"attempt_starttime"`
	AttemptEnd      string                    `json:"attempt_endtime"`
	AttemptID       string                    `json:"attempt_id"`
	Questions       map[string]QuestionResult `json:"questions"`
	PDFURL          string                    `json:"pdf_url"`
	ReportURL       string                    `json:"report_url"`
}

// QuestionResult holds a candidate's answer and submissions for one question.
type QuestionResult struct {
	Answered    bool         `json:"answered"`
	Answer      Answer       `json:"answer"`
	Score       float64      `json:"score"`
	Submissions []Submission `json:"submissions"`
}

// Answer holds the final code submission.
type Answer struct {
	Code     string `json:"code"`
	Language string `json:"language"`
}

// Submission holds one submission attempt.
type Submission struct {
	ID        int64            `json:"id"`
	Answer    Answer           `json:"answer"`
	Score     float64          `json:"score"`
	IsValid   bool             `json:"is_valid"`
	CreatedAt string           `json:"created_at"`
	Metadata  SubmissionMeta   `json:"metadata"`
}

// SubmissionMeta holds testcase results for a submission.
type SubmissionMeta struct {
	Result         int   `json:"result"`
	TestcaseStatus []int `json:"testcase_status"`
}

// Interview represents a HackerRank interview.
type Interview struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
	URL       string `json:"url"`
}

// Transcript holds interview transcript data.
type Transcript struct {
	Messages []TranscriptMessage `json:"messages"`
}

// TranscriptMessage holds a single message in a transcript.
type TranscriptMessage struct {
	Author    string `json:"author"`
	Email     string `json:"email"`
	Candidate bool   `json:"candidate"`
	Text      string `json:"text"`
	Timestamp string `json:"timestamp"`
}
```

- [ ] **Step 4: Implement the HTTP client**

```go
// internal/api/client.go
package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const defaultBaseURL = "https://www.hackerrank.com"

// Client is an HTTP client for the HackerRank for Work API.
type Client struct {
	token   string
	baseURL string
	http    *http.Client
}

// Option configures a Client.
type Option func(*Client)

// WithBaseURL overrides the default base URL (for testing).
func WithBaseURL(u string) Option {
	return func(c *Client) { c.baseURL = u }
}

// NewClient creates a new API client.
func NewClient(token string, opts ...Option) *Client {
	c := &Client{
		token:   token,
		baseURL: defaultBaseURL,
		http:    &http.Client{},
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

// Get makes a GET request to the API path with optional query params and decodes into dest.
func (c *Client) Get(path string, params url.Values, dest interface{}) error {
	u := c.baseURL + "/x/api/v3" + path
	if params != nil {
		u += "?" + params.Encode()
	}

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	return json.NewDecoder(resp.Body).Decode(dest)
}
```

- [ ] **Step 5: Implement pagination**

```go
// internal/api/pagination.go
package api

import (
	"fmt"
	"net/url"
)

const pageSize = 20

// ListResponse is the generic shape of paginated HackerRank API responses.
type ListResponse[T any] struct {
	Data  []T `json:"data"`
	Total int `json:"total"`
}

// Paginate fetches all pages for a list endpoint.
func Paginate[T any](c *Client, path string, params url.Values) ([]T, error) {
	if params == nil {
		params = url.Values{}
	}

	var all []T
	offset := 0
	for {
		p := url.Values{}
		for k, v := range params {
			p[k] = v
		}
		p.Set("limit", fmt.Sprintf("%d", pageSize))
		p.Set("offset", fmt.Sprintf("%d", offset))

		var resp ListResponse[T]
		if err := c.Get(path, p, &resp); err != nil {
			return nil, err
		}
		all = append(all, resp.Data...)
		if len(all) >= resp.Total {
			break
		}
		offset += pageSize
	}
	return all, nil
}
```

- [ ] **Step 6: Run tests**

Run: `go test ./internal/api/ -v`
Expected: All 3 tests pass.

- [ ] **Step 7: Commit**

```bash
git add internal/api/ go.mod go.sum
git commit -m "feat: add API client with pagination"
```

---

### Task 4: Output Package

**Files:**
- Create: `internal/output/output.go`
- Create: `internal/output/output_test.go`

- [ ] **Step 1: Write failing test for output formatters**

```go
// internal/output/output_test.go
package output

import (
	"bytes"
	"strings"
	"testing"
)

func TestTable(t *testing.T) {
	var buf bytes.Buffer
	w := NewTableWriter(&buf)
	w.SetHeader([]string{"ID", "NAME", "STATE"})
	w.Append([]string{"123", "My Test", "active"})
	w.Append([]string{"456", "Other Test", "draft"})
	w.Render()

	out := buf.String()
	if !strings.Contains(out, "123") {
		t.Errorf("table missing ID 123:\n%s", out)
	}
	if !strings.Contains(out, "My Test") {
		t.Errorf("table missing name:\n%s", out)
	}
}

func TestJSON(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]string{"id": "123"}
	if err := WriteJSON(&buf, data); err != nil {
		t.Fatalf("WriteJSON: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, `"id"`) {
		t.Errorf("json missing id field:\n%s", out)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/output/ -v`
Expected: Compilation error.

- [ ] **Step 3: Implement output package**

```go
// internal/output/output.go
package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

// TableWriter wraps tabwriter for aligned column output.
type TableWriter struct {
	w       *tabwriter.Writer
	headers []string
	rows    [][]string
}

// NewTableWriter creates a table writer that writes to w.
func NewTableWriter(w io.Writer) *TableWriter {
	return &TableWriter{
		w: tabwriter.NewWriter(w, 0, 4, 2, ' ', 0),
	}
}

// SetHeader sets the column headers.
func (t *TableWriter) SetHeader(headers []string) {
	t.headers = headers
}

// Append adds a row.
func (t *TableWriter) Append(row []string) {
	t.rows = append(t.rows, row)
}

// Render writes the table.
func (t *TableWriter) Render() {
	if len(t.headers) > 0 {
		fmt.Fprintln(t.w, strings.Join(t.headers, "\t"))
	}
	for _, row := range t.rows {
		fmt.Fprintln(t.w, strings.Join(row, "\t"))
	}
	t.w.Flush()
}

// WriteJSON writes data as pretty-printed JSON.
func WriteJSON(w io.Writer, data interface{}) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}
```

- [ ] **Step 4: Run tests**

Run: `go test ./internal/output/ -v`
Expected: All 2 tests pass.

- [ ] **Step 5: Commit**

```bash
git add internal/output/
git commit -m "feat: add output package with table and JSON formatters"
```

---

### Task 5: Auth Commands

**Files:**
- Create: `cmd/auth.go`
- Modify: `cmd/root.go` — wire up auth subcommand and token resolution

- [ ] **Step 1: Implement auth commands**

```go
// cmd/auth.go
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/jholm117/hackerrank-cli/internal/config"
	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage authentication",
}

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Save API token to config",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Print("Enter HackerRank API token: ")
		reader := bufio.NewReader(os.Stdin)
		token, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		token = strings.TrimSpace(token)
		if token == "" {
			return fmt.Errorf("token cannot be empty")
		}

		path := config.DefaultPath()
		cfg, err := config.Load(path)
		if err != nil {
			return err
		}
		cfg.Token = token
		if err := config.Save(cfg, path); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "Token saved to %s\n", path)
		return nil
	},
}

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove stored token",
	RunE: func(cmd *cobra.Command, args []string) error {
		path := config.DefaultPath()
		cfg, err := config.Load(path)
		if err != nil {
			return err
		}
		cfg.Token = ""
		if err := config.Save(cfg, path); err != nil {
			return err
		}
		fmt.Fprintln(os.Stderr, "Token removed.")
		return nil
	},
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current auth state",
	RunE: func(cmd *cobra.Command, args []string) error {
		path := config.DefaultPath()
		cfg, err := config.Load(path)
		if err != nil {
			return err
		}
		token := config.ResolveToken(flagToken, cfg)
		if token == "" {
			fmt.Println("Not authenticated. Run: hr auth login")
			return nil
		}
		source := "config file"
		if flagToken != "" {
			source = "--token flag"
		} else if os.Getenv("HACKERRANK_API_TOKEN") != "" {
			source = "HACKERRANK_API_TOKEN env var"
		}
		fmt.Printf("Authenticated via %s\n", source)
		fmt.Printf("Token: %s...%s\n", token[:4], token[len(token)-4:])
		return nil
	},
}

func init() {
	authCmd.AddCommand(authLoginCmd)
	authCmd.AddCommand(authLogoutCmd)
	authCmd.AddCommand(authStatusCmd)
	rootCmd.AddCommand(authCmd)
}
```

- [ ] **Step 2: Add helper to root.go for resolving API client**

Add these imports to `cmd/root.go`'s import block:
```go
"fmt"

"github.com/jholm117/hackerrank-cli/internal/api"
"github.com/jholm117/hackerrank-cli/internal/config"
```

Add this function at the bottom of `cmd/root.go`:
```go
// newClient creates an API client from resolved token.
func newClient() (*api.Client, error) {
	cfg, err := config.Load(config.DefaultPath())
	if err != nil {
		return nil, err
	}
	token := config.ResolveToken(flagToken, cfg)
	if token == "" {
		return nil, fmt.Errorf("not authenticated — run: hr auth login")
	}
	return api.NewClient(token), nil
}
```

- [ ] **Step 3: Build and test manually**

Run:
```bash
make build && ./hr auth status
```
Expected: "Not authenticated. Run: hr auth login"

- [ ] **Step 4: Commit**

```bash
git add cmd/auth.go cmd/root.go
git commit -m "feat: add auth login/logout/status commands"
```

---

### Task 6: Tests Commands

**Files:**
- Create: `cmd/tests.go`

- [ ] **Step 1: Implement tests list and get commands**

```go
// cmd/tests.go
package cmd

import (
	"fmt"
	"os"

	"github.com/jholm117/hackerrank-cli/internal/api"
	"github.com/jholm117/hackerrank-cli/internal/output"
	"github.com/spf13/cobra"
)

var testsCmd = &cobra.Command{
	Use:   "tests",
	Short: "Manage tests",
}

var testsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tests",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}

		tests, err := api.Paginate[api.Test](c, "/tests", nil)
		if err != nil {
			return err
		}

		if flagOutput == "json" {
			return output.WriteJSON(os.Stdout, tests)
		}

		w := output.NewTableWriter(os.Stdout)
		w.SetHeader([]string{"ID", "NAME", "STATE", "DRAFT", "QUESTIONS"})
		for _, t := range tests {
			w.Append([]string{
				t.ID,
				t.Name,
				t.State,
				fmt.Sprintf("%v", t.Draft),
				fmt.Sprintf("%d", len(t.Questions)),
			})
		}
		w.Render()
		return nil
	},
}

var testsGetCmd = &cobra.Command{
	Use:   "get <test-id>",
	Short: "Show test details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}

		var test api.Test
		if err := c.Get("/tests/"+args[0], nil, &test); err != nil {
			return err
		}

		if flagOutput == "table" {
			w := output.NewTableWriter(os.Stdout)
			w.SetHeader([]string{"FIELD", "VALUE"})
			w.Append([]string{"ID", test.ID})
			w.Append([]string{"Name", test.Name})
			w.Append([]string{"State", test.State})
			w.Append([]string{"Draft", fmt.Sprintf("%v", test.Draft)})
			w.Append([]string{"Duration", fmt.Sprintf("%d min", test.Duration)})
			w.Append([]string{"Questions", fmt.Sprintf("%d", len(test.Questions))})
			w.Append([]string{"Created", test.CreatedAt})
			w.Render()
			return nil
		}

		return output.WriteJSON(os.Stdout, test)
	},
}

func init() {
	testsCmd.AddCommand(testsListCmd)
	testsCmd.AddCommand(testsGetCmd)
	rootCmd.AddCommand(testsCmd)
}
```

- [ ] **Step 2: Build and test with real API**

Run:
```bash
make build && ./hr tests list --token "$(security find-generic-password -s hackerrank-api-token -w)"
```
Expected: Table of tests.

- [ ] **Step 3: Test get command**

Run:
```bash
./hr tests get 2309131 --token "$(security find-generic-password -s hackerrank-api-token -w)"
```
Expected: Details for the TAI Technical Take Home test.

- [ ] **Step 4: Commit**

```bash
git add cmd/tests.go
git commit -m "feat: add tests list and get commands"
```

---

### Task 7: Candidates Commands

**Files:**
- Create: `cmd/candidates.go`

- [ ] **Step 1: Implement candidates list, get, and code commands**

```go
// cmd/candidates.go
package cmd

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/jholm117/hackerrank-cli/internal/api"
	"github.com/jholm117/hackerrank-cli/internal/output"
	"github.com/spf13/cobra"
)

var candidatesCmd = &cobra.Command{
	Use:   "candidates",
	Short: "Manage test candidates",
}

var candidatesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List candidates for a test",
	RunE: func(cmd *cobra.Command, args []string) error {
		testID, _ := cmd.Flags().GetString("test")
		if testID == "" {
			return fmt.Errorf("--test flag is required")
		}

		c, err := newClient()
		if err != nil {
			return err
		}

		candidates, err := api.Paginate[api.Candidate](c, "/tests/"+testID+"/candidates", nil)
		if err != nil {
			return err
		}

		if flagOutput == "json" {
			return output.WriteJSON(os.Stdout, candidates)
		}

		w := output.NewTableWriter(os.Stdout)
		w.SetHeader([]string{"ID", "NAME", "EMAIL", "SCORE", "STATUS", "DATE"})
		for _, cand := range candidates {
			w.Append([]string{
				cand.ID,
				cand.FullName,
				cand.Email,
				fmt.Sprintf("%.0f%%", cand.PercentageScore),
				fmt.Sprintf("%d", cand.Status),
				cand.AttemptStart,
			})
		}
		w.Render()
		return nil
	},
}

var candidatesGetCmd = &cobra.Command{
	Use:   "get <test-id> <candidate-id>",
	Short: "Show candidate details with submissions",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}

		params := url.Values{}
		params.Set("additional_fields", "questions,questions.solves,questions.submission_result")

		var cand api.Candidate
		if err := c.Get("/tests/"+args[0]+"/candidates/"+args[1], params, &cand); err != nil {
			return err
		}

		if flagOutput == "table" {
			w := output.NewTableWriter(os.Stdout)
			w.SetHeader([]string{"FIELD", "VALUE"})
			w.Append([]string{"ID", cand.ID})
			w.Append([]string{"Name", cand.FullName})
			w.Append([]string{"Email", cand.Email})
			w.Append([]string{"Score", fmt.Sprintf("%.0f%%", cand.PercentageScore)})
			w.Append([]string{"Started", cand.AttemptStart})
			w.Append([]string{"Ended", cand.AttemptEnd})
			w.Append([]string{"Questions", fmt.Sprintf("%d", len(cand.Questions))})
			w.Render()
			return nil
		}

		return output.WriteJSON(os.Stdout, cand)
	},
}

var candidatesCodeCmd = &cobra.Command{
	Use:   "code <test-id> <candidate-id>",
	Short: "Extract candidate source code",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		saveDir, _ := cmd.Flags().GetString("save")

		c, err := newClient()
		if err != nil {
			return err
		}

		params := url.Values{}
		params.Set("additional_fields", "questions,questions.solves,questions.submission_result")

		var cand api.Candidate
		if err := c.Get("/tests/"+args[0]+"/candidates/"+args[1], params, &cand); err != nil {
			return err
		}

		i := 0
		for qID, q := range cand.Questions {
			i++
			if !q.Answered {
				continue
			}

			lang := q.Answer.Language
			ext := langExtension(lang)
			name := fmt.Sprintf("q%d_%s", i, qID)

			if saveDir != "" {
				filename := name + ext
				path := filepath.Join(saveDir, filename)
				if err := os.MkdirAll(saveDir, 0o755); err != nil {
					return err
				}
				if err := os.WriteFile(path, []byte(q.Answer.Code), 0o644); err != nil {
					return err
				}
				fmt.Fprintf(os.Stderr, "Saved %s\n", path)
			} else {
				header := fmt.Sprintf("## Question %d: %s (%s) — %.0f/%.0f",
					i, qID, lang, q.Score, q.Score)
				fmt.Println(header)
				fmt.Println(strings.Repeat("─", 60))
				fmt.Println(q.Answer.Code)
				fmt.Println()
			}
		}
		return nil
	},
}

func langExtension(lang string) string {
	lang = strings.ToLower(lang)
	switch {
	case strings.Contains(lang, "python"):
		return ".py"
	case strings.Contains(lang, "java") && !strings.Contains(lang, "javascript"):
		return ".java"
	case strings.Contains(lang, "javascript"):
		return ".js"
	case strings.Contains(lang, "typescript"):
		return ".ts"
	case strings.Contains(lang, "go"):
		return ".go"
	case strings.Contains(lang, "ruby"):
		return ".rb"
	case strings.Contains(lang, "rust"):
		return ".rs"
	case strings.Contains(lang, "c++") || strings.Contains(lang, "cpp"):
		return ".cpp"
	case lang == "c":
		return ".c"
	default:
		return ".txt"
	}
}

func init() {
	candidatesListCmd.Flags().String("test", "", "Test ID (required)")
	candidatesListCmd.MarkFlagRequired("test")
	candidatesCodeCmd.Flags().String("save", "", "Directory to save code files")
	candidatesCmd.AddCommand(candidatesListCmd)
	candidatesCmd.AddCommand(candidatesGetCmd)
	candidatesCmd.AddCommand(candidatesCodeCmd)
	rootCmd.AddCommand(candidatesCmd)
}
```

- [ ] **Step 2: Build and test candidates list**

Run:
```bash
make build && ./hr candidates list --test 2309131 --token "$(security find-generic-password -s hackerrank-api-token -w)"
```
Expected: Table with Nicholas Sypherd's candidate entry.

- [ ] **Step 3: Test candidates code**

Run:
```bash
./hr candidates code 2309131 114907465 --token "$(security find-generic-password -s hackerrank-api-token -w)"
```
Expected: Source code printed to stdout with question headers.

- [ ] **Step 4: Test --save flag**

Run:
```bash
./hr candidates code 2309131 114907465 --save /tmp/hr-test --token "$(security find-generic-password -s hackerrank-api-token -w)"
ls /tmp/hr-test/
```
Expected: `.py` files saved to `/tmp/hr-test/`.

- [ ] **Step 5: Commit**

```bash
git add cmd/candidates.go
git commit -m "feat: add candidates list, get, and code commands"
```

---

### Task 8: Interviews Commands

**Files:**
- Create: `cmd/interviews.go`

- [ ] **Step 1: Implement interviews list, get, and transcript commands**

```go
// cmd/interviews.go
package cmd

import (
	"fmt"
	"os"

	"github.com/jholm117/hackerrank-cli/internal/api"
	"github.com/jholm117/hackerrank-cli/internal/output"
	"github.com/spf13/cobra"
)

var interviewsCmd = &cobra.Command{
	Use:   "interviews",
	Short: "Manage interviews",
}

var interviewsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all interviews",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}

		interviews, err := api.Paginate[api.Interview](c, "/interviews", nil)
		if err != nil {
			return err
		}

		if flagOutput == "json" {
			return output.WriteJSON(os.Stdout, interviews)
		}

		w := output.NewTableWriter(os.Stdout)
		w.SetHeader([]string{"ID", "TITLE", "STATUS", "CREATED"})
		for _, iv := range interviews {
			w.Append([]string{iv.ID, iv.Title, iv.Status, iv.CreatedAt})
		}
		w.Render()
		return nil
	},
}

var interviewsGetCmd = &cobra.Command{
	Use:   "get <interview-id>",
	Short: "Show interview details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}

		var iv api.Interview
		if err := c.Get("/interviews/"+args[0], nil, &iv); err != nil {
			return err
		}

		if flagOutput == "table" {
			w := output.NewTableWriter(os.Stdout)
			w.SetHeader([]string{"FIELD", "VALUE"})
			w.Append([]string{"ID", iv.ID})
			w.Append([]string{"Title", iv.Title})
			w.Append([]string{"Status", iv.Status})
			w.Append([]string{"Created", iv.CreatedAt})
			w.Append([]string{"URL", iv.URL})
			w.Render()
			return nil
		}

		return output.WriteJSON(os.Stdout, iv)
	},
}

var interviewsTranscriptCmd = &cobra.Command{
	Use:   "transcript <interview-id>",
	Short: "Get interview transcript",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}

		var transcript api.Transcript
		if err := c.Get("/interviews/"+args[0]+"/transcript", nil, &transcript); err != nil {
			return err
		}

		if flagOutput == "json" {
			return output.WriteJSON(os.Stdout, transcript)
		}

		for _, msg := range transcript.Messages {
			role := msg.Author
			if msg.Candidate {
				role = fmt.Sprintf("%s (candidate)", msg.Author)
			}
			fmt.Printf("[%s] %s:\n%s\n\n", msg.Timestamp, role, msg.Text)
		}
		return nil
	},
}

func init() {
	interviewsCmd.AddCommand(interviewsListCmd)
	interviewsCmd.AddCommand(interviewsGetCmd)
	interviewsCmd.AddCommand(interviewsTranscriptCmd)
	rootCmd.AddCommand(interviewsCmd)
}
```

- [ ] **Step 2: Build and test**

Run:
```bash
make build && ./hr interviews list --token "$(security find-generic-password -s hackerrank-api-token -w)"
```
Expected: Table of interviews (may be empty depending on org data).

- [ ] **Step 3: Commit**

```bash
git add cmd/interviews.go
git commit -m "feat: add interviews list, get, and transcript commands"
```

---

### Task 9: CI, Hooks, and Linting

**Files:**
- Create: `.githooks/pre-push`
- Create: `hack/ci-checks.sh`
- Create: `.golangci.yml`
- Create: `.github/workflows/ci.yml`

- [ ] **Step 1: Create golangci-lint config**

```yaml
# .golangci.yml
version: "2"
linters:
  enable:
    - errcheck
    - govet
    - staticcheck
    - ineffassign
    - misspell
    - gocyclo
    - unused
formatters:
  enable:
    - gofmt
    - goimports
```

- [ ] **Step 2: Create ci-checks script**

```bash
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
```

Make it executable:
```bash
chmod +x hack/ci-checks.sh
```

- [ ] **Step 3: Create pre-push hook**

```bash
#!/usr/bin/env bash
# .githooks/pre-push
exec hack/ci-checks.sh
```

Make it executable:
```bash
chmod +x .githooks/pre-push
```

- [ ] **Step 4: Create CI workflow**

```yaml
# .github/workflows/ci.yml
name: CI

on:
  push:
    branches: [main]
  pull_request:

concurrency:
  group: ci-${{ github.ref }}
  cancel-in-progress: true

jobs:
  lint-and-test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
      - name: Install golangci-lint
        run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
      - name: Run checks
        run: hack/ci-checks.sh --parallel
```

- [ ] **Step 5: Add setup-hooks target to Makefile (already done in Task 1)**

Verify `make setup-hooks` works:
```bash
make setup-hooks
```
Expected: "Git hooks configured to use .githooks/"

- [ ] **Step 6: Run the checks locally**

Run:
```bash
hack/ci-checks.sh
```
Expected: All checks pass.

- [ ] **Step 7: Commit**

```bash
git add .golangci.yml hack/ .githooks/ .github/
git commit -m "ci: add golangci-lint, ci-checks script, pre-push hook, and GitHub Actions"
```

---

### Task 10: GoReleaser and Homebrew

**Files:**
- Create: `.goreleaser.yaml`
- Create: `.github/workflows/release.yml`

- [ ] **Step 1: Create GoReleaser config**

```yaml
# .goreleaser.yaml
version: 2
project_name: hackerrank-cli

builds:
  - main: .
    binary: hr
    env:
      - CGO_ENABLED=0
    goos:
      - darwin
      - linux
    goarch:
      - amd64
      - arm64

archives:
  - format: tar.gz
    name_template: "hr_{{ .Version }}_{{ .Os }}_{{ .Arch }}"

checksum:
  name_template: "checksums.txt"

brews:
  - repository:
      owner: jholm117
      name: homebrew-tap
    directory: Formula
    homepage: "https://github.com/jholm117/hackerrank-cli"
    description: "CLI for HackerRank for Work API"
    install: |
      bin.install "hr"
    test: |
      assert_match "hr", shell_output("#{bin}/hr --help")
```

- [ ] **Step 2: Create release workflow**

```yaml
# .github/workflows/release.yml
name: Release

on:
  push:
    tags:
      - "v*"

permissions:
  contents: write

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
      - uses: goreleaser/goreleaser-action@v6
        with:
          version: latest
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          HOMEBREW_TAP_GITHUB_TOKEN: ${{ secrets.HOMEBREW_TAP_GITHUB_TOKEN }}
```

- [ ] **Step 3: Verify goreleaser config**

Run:
```bash
goreleaser check
```
Expected: Config is valid (or install goreleaser first: `go install github.com/goreleaser/goreleaser/v2@latest`).

- [ ] **Step 4: Commit**

```bash
git add .goreleaser.yaml .github/workflows/release.yml
git commit -m "ci: add goreleaser config and release workflow"
```

---

### Task 11: Final Integration Test and README

**Files:**
- Create: `README.md`

- [ ] **Step 1: Run full test suite**

Run:
```bash
go test ./... -v
```
Expected: All tests pass.

- [ ] **Step 2: Run full CI checks**

Run:
```bash
hack/ci-checks.sh
```
Expected: All checks pass.

- [ ] **Step 3: End-to-end smoke test**

Run:
```bash
make build
./hr auth status
./hr tests list --token "$(security find-generic-password -s hackerrank-api-token -w)"
./hr candidates list --test 2309131 --token "$(security find-generic-password -s hackerrank-api-token -w)"
./hr candidates code 2309131 114907465 --token "$(security find-generic-password -s hackerrank-api-token -w)" | head -20
```
Expected: All commands produce expected output.

- [ ] **Step 4: Create README**

```markdown
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
```

- [ ] **Step 5: Commit**

```bash
git add README.md
git commit -m "docs: add README with install and usage instructions"
```
