package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func captureStdout(fn func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	return buf.String()
}

func TestTestsList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/x/api/v3/tests" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]interface{}{
				{"id": "123", "name": "Test One", "state": "active", "draft": false, "questions": []string{"q1"}},
				{"id": "456", "name": "Test Two", "state": "active", "draft": true, "questions": []string{}},
			},
			"total": 2,
		})
	}))
	defer server.Close()

	var execErr error
	out := captureStdout(func() {
		rootCmd.SetArgs([]string{"tests", "list", "--token", "test-tok", "--base-url", server.URL})
		execErr = rootCmd.Execute()
	})

	if execErr != nil {
		t.Fatalf("execute error: %v", execErr)
	}
	if !strings.Contains(out, "Test One") {
		t.Errorf("output missing 'Test One':\n%s", out)
	}
	if !strings.Contains(out, "123") {
		t.Errorf("output missing ID '123':\n%s", out)
	}
}

func TestTestsGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/x/api/v3/tests/123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id": "123", "name": "My Test", "state": "active", "draft": false,
			"duration": 60, "questions": []string{"q1", "q2"}, "created_at": "2026-01-01",
		})
	}))
	defer server.Close()

	var execErr error
	out := captureStdout(func() {
		rootCmd.SetArgs([]string{"tests", "get", "123", "--token", "test-tok", "--base-url", server.URL, "--output", "json"})
		execErr = rootCmd.Execute()
	})

	if execErr != nil {
		t.Fatalf("execute error: %v", execErr)
	}
	if !strings.Contains(out, "My Test") {
		t.Errorf("output missing 'My Test':\n%s", out)
	}
}

func TestTestsListJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]interface{}{
				{"id": "789", "name": "JSON Test", "state": "active", "draft": false, "questions": []string{}},
			},
			"total": 1,
		})
	}))
	defer server.Close()

	var execErr error
	out := captureStdout(func() {
		rootCmd.SetArgs([]string{"tests", "list", "--token", "test-tok", "--base-url", server.URL, "--output", "json"})
		execErr = rootCmd.Execute()
	})

	if execErr != nil {
		t.Fatalf("execute error: %v", execErr)
	}
	if !strings.Contains(out, `"id"`) {
		t.Errorf("json output missing 'id' field:\n%s", out)
	}
	if !strings.Contains(out, "JSON Test") {
		t.Errorf("json output missing 'JSON Test':\n%s", out)
	}
}
