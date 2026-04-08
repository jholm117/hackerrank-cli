package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

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
