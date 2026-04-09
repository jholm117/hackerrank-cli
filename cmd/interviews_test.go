package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestInterviewsSearch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/x/api/v3/interviews/search" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("query") != "shubhadeep" {
			t.Errorf("expected query=shubhadeep, got %q", r.URL.Query().Get("query"))
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]interface{}{
				{"id": "8126945", "title": "shubhadeep.0708@gmail.com", "status": "ended", "created_at": "2026-03-26"},
			},
			"total": 1,
		})
	}))
	defer server.Close()

	var execErr error
	out := captureStdout(func() {
		rootCmd.SetArgs([]string{"interviews", "search", "--query", "shubhadeep", "--token", "tok", "--base-url", server.URL})
		execErr = rootCmd.Execute()
	})

	if execErr != nil {
		t.Fatalf("execute error: %v", execErr)
	}
	if !strings.Contains(out, "shubhadeep.0708@gmail.com") {
		t.Errorf("missing interview:\n%s", out)
	}
	if !strings.Contains(out, "8126945") {
		t.Errorf("missing interview ID:\n%s", out)
	}
}

func TestInterviewsListSortOrder(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sort := r.URL.Query().Get("sort")
		order := r.URL.Query().Get("order")
		if sort != "created_at" {
			t.Errorf("expected sort=created_at, got %q", sort)
		}
		if order != "desc" {
			t.Errorf("expected order=desc, got %q", order)
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]interface{}{
				{"id": "iv3", "title": "newest@test.com", "status": "ended", "created_at": "2026-04-08"},
				{"id": "iv2", "title": "older@test.com", "status": "ended", "created_at": "2026-04-07"},
			},
			"total": 2,
		})
	}))
	defer server.Close()

	var execErr error
	out := captureStdout(func() {
		rootCmd.SetArgs([]string{"interviews", "list", "--token", "tok", "--base-url", server.URL, "--sort", "created_at", "--order", "desc"})
		execErr = rootCmd.Execute()
	})

	if execErr != nil {
		t.Fatalf("execute error: %v", execErr)
	}
	if !strings.Contains(out, "newest@test.com") {
		t.Errorf("missing newest interview:\n%s", out)
	}
}

func TestInterviewsList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/x/api/v3/interviews" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]interface{}{
				{"id": "iv1", "title": "Backend Interview", "status": "completed", "created_at": "2026-01-01"},
				{"id": "iv2", "title": "Frontend Interview", "status": "scheduled", "created_at": "2026-01-02"},
			},
			"total": 2,
		})
	}))
	defer server.Close()

	var execErr error
	out := captureStdout(func() {
		rootCmd.SetArgs([]string{"interviews", "list", "--token", "tok", "--base-url", server.URL})
		execErr = rootCmd.Execute()
	})

	if execErr != nil {
		t.Fatalf("execute error: %v", execErr)
	}
	if !strings.Contains(out, "Backend Interview") {
		t.Errorf("missing interview title:\n%s", out)
	}
	if !strings.Contains(out, "iv1") {
		t.Errorf("missing interview ID:\n%s", out)
	}
}

func TestInterviewsGetJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/x/api/v3/interviews/iv1" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id": "iv1", "title": "Backend Interview", "status": "completed",
			"created_at": "2026-01-01", "url": "https://hackerrank.com/interview/iv1",
		})
	}))
	defer server.Close()

	var execErr error
	out := captureStdout(func() {
		rootCmd.SetArgs([]string{"interviews", "get", "iv1", "--token", "tok", "--base-url", server.URL, "--output", "json"})
		execErr = rootCmd.Execute()
	})

	if execErr != nil {
		t.Fatalf("execute error: %v", execErr)
	}
	if !strings.Contains(out, "Backend Interview") {
		t.Errorf("missing interview title:\n%s", out)
	}
}

func TestInterviewsCode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/interviews/iv1/recordings/code" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"questions": []map[string]interface{}{
					{
						"qtype":    "code",
						"question": "<h3>Task Prioritizer</h3><p>Implement a priority queue</p>",
						"runs": []map[string]interface{}{
							{"code": "def solve():\n    return 42", "lang": "python3"},
						},
					},
				},
			},
		})
	}))
	defer server.Close()

	flagOutput = ""

	var execErr error
	out := captureStdout(func() {
		rootCmd.SetArgs([]string{"interviews", "code", "iv1", "--token", "tok", "--base-url", server.URL})
		execErr = rootCmd.Execute()
	})

	if execErr != nil {
		t.Fatalf("execute error: %v", execErr)
	}
	if !strings.Contains(out, "def solve()") {
		t.Errorf("missing code in output:\n%s", out)
	}
	if !strings.Contains(out, "Task Prioritizer") {
		t.Errorf("missing question title:\n%s", out)
	}
	if !strings.Contains(out, "python3") {
		t.Errorf("missing language:\n%s", out)
	}
}

func TestInterviewsTranscript(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/x/api/v3/interviews/iv1/transcript" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"messages": []map[string]interface{}{
				{"author": "Interviewer", "email": "int@co.com", "candidate": false, "text": "Tell me about yourself", "timestamp": "10:00"},
				{"author": "Candidate", "email": "cand@co.com", "candidate": true, "text": "I am a developer", "timestamp": "10:01"},
			},
		})
	}))
	defer server.Close()

	// Reset flagOutput in case a prior test set --output json
	flagOutput = ""

	var execErr error
	out := captureStdout(func() {
		rootCmd.SetArgs([]string{"interviews", "transcript", "iv1", "--token", "tok", "--base-url", server.URL})
		execErr = rootCmd.Execute()
	})

	if execErr != nil {
		t.Fatalf("execute error: %v", execErr)
	}
	if !strings.Contains(out, "Tell me about yourself") {
		t.Errorf("missing interviewer message:\n%s", out)
	}
	if !strings.Contains(out, "(candidate)") {
		t.Errorf("missing candidate label:\n%s", out)
	}
}
