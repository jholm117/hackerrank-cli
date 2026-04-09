package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestCandidatesList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/x/api/v3/tests/100/candidates" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]interface{}{
				{
					"id": "1001", "full_name": "Alice Smith", "email": "alice@example.com",
					"score": 95.0, "percentage_score": 95.0, "status": 7,
					"attempt_starttime": "2026-01-01T10:00:00+0000", "questions": map[string]float64{},
				},
			},
			"total": 1,
		})
	}))
	defer server.Close()

	var execErr error
	out := captureStdout(func() {
		rootCmd.SetArgs([]string{"candidates", "list", "--test", "100", "--token", "tok", "--base-url", server.URL})
		execErr = rootCmd.Execute()
	})

	if execErr != nil {
		t.Fatalf("execute error: %v", execErr)
	}
	if !strings.Contains(out, "Alice Smith") {
		t.Errorf("missing candidate name:\n%s", out)
	}
	if !strings.Contains(out, "1001") {
		t.Errorf("missing candidate ID:\n%s", out)
	}
}

func TestCandidatesListRequiresTestFlag(t *testing.T) {
	rootCmd.SetArgs([]string{"candidates", "list", "--token", "tok", "--base-url", "http://localhost"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error when --test not provided")
	}
}

func TestCandidatesSearch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/x/api/v3/tests/100/candidates/search" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("search") != "alice" {
			t.Errorf("unexpected search param: %s", r.URL.Query().Get("search"))
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]interface{}{
				{
					"id": "1001", "full_name": "Alice Smith", "email": "alice@example.com",
					"score": 95.0, "percentage_score": 95.0, "status": 7,
					"attempt_starttime": "2026-01-01T10:00:00+0000", "questions": map[string]float64{},
				},
			},
			"total": 1,
		})
	}))
	defer server.Close()

	var execErr error
	out := captureStdout(func() {
		rootCmd.SetArgs([]string{"candidates", "search", "--test", "100", "--query", "alice", "--token", "tok", "--base-url", server.URL})
		execErr = rootCmd.Execute()
	})

	if execErr != nil {
		t.Fatalf("execute error: %v", execErr)
	}
	if !strings.Contains(out, "Alice Smith") {
		t.Errorf("missing candidate name:\n%s", out)
	}
	if !strings.Contains(out, "alice@example.com") {
		t.Errorf("missing email:\n%s", out)
	}
}

func TestCandidatesSearchRequiresFlags(t *testing.T) {
	rootCmd.SetArgs([]string{"candidates", "search", "--token", "tok", "--base-url", "http://localhost"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error when --test and --query not provided")
	}
}

func TestCandidatesCode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id": "1001", "full_name": "Alice", "email": "a@b.com",
			"score": 100.0, "percentage_score": 100.0, "status": 7,
			"attempt_starttime": "2026-01-01", "attempt_endtime": "2026-01-01",
			"questions": map[string]interface{}{
				"q1": map[string]interface{}{
					"answered": true,
					"answer":   map[string]interface{}{"code": "print('hello')", "language": "python3"},
					"score":    50.0,
				},
			},
		})
	}))
	defer server.Close()

	var execErr error
	out := captureStdout(func() {
		rootCmd.SetArgs([]string{"candidates", "code", "100", "1001", "--token", "tok", "--base-url", server.URL})
		execErr = rootCmd.Execute()
	})

	if execErr != nil {
		t.Fatalf("execute error: %v", execErr)
	}
	if !strings.Contains(out, "print('hello')") {
		t.Errorf("missing code in output:\n%s", out)
	}
	if !strings.Contains(out, "python3") {
		t.Errorf("missing language in output:\n%s", out)
	}
}

func TestCandidatesCodeSave(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id": "1001", "full_name": "Alice", "email": "a@b.com",
			"score": 100.0, "percentage_score": 100.0, "status": 7,
			"questions": map[string]interface{}{
				"q1": map[string]interface{}{
					"answered": true,
					"answer":   map[string]interface{}{"code": "def solve(): pass", "language": "python3"},
					"score":    50.0,
				},
			},
		})
	}))
	defer server.Close()

	dir := t.TempDir()
	rootCmd.SetArgs([]string{"candidates", "code", "100", "1001", "--token", "tok", "--base-url", server.URL, "--save", dir})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("execute error: %v", err)
	}

	entries, _ := os.ReadDir(dir)
	if len(entries) == 0 {
		t.Fatal("no files saved")
	}
	found := false
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".py") {
			found = true
			data, _ := os.ReadFile(dir + "/" + e.Name())
			if !strings.Contains(string(data), "def solve()") {
				t.Errorf("saved file missing code: %s", string(data))
			}
		}
	}
	if !found {
		t.Errorf("no .py file found in %v", entries)
	}
}
