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
