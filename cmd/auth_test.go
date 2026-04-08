package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jholm117/hackerrank-cli/internal/config"
)

func TestAuthLoginSavesToken(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	// Override config path and token reader for testing
	origConfigPath := configPath
	origReadToken := readToken
	defer func() {
		configPath = origConfigPath
		readToken = origReadToken
	}()
	configPath = path
	readToken = func() (string, error) {
		return "test-secret-token", nil
	}

	// Capture stdout to verify token is NOT printed
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	rootCmd.SetArgs([]string{"auth", "login"})
	err := rootCmd.Execute()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	stdout := buf.String()

	if err != nil {
		t.Fatalf("execute error: %v", err)
	}

	// Token must NOT appear in stdout
	if strings.Contains(stdout, "test-secret-token") {
		t.Error("token was printed to stdout — must be masked")
	}

	// Token must be saved to config
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.Token != "test-secret-token" {
		t.Errorf("got token %q, want %q", cfg.Token, "test-secret-token")
	}
}

func TestAuthLoginRejectsEmpty(t *testing.T) {
	origReadToken := readToken
	defer func() { readToken = origReadToken }()
	readToken = func() (string, error) {
		return "", nil
	}

	rootCmd.SetArgs([]string{"auth", "login"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error for empty token")
	}
}

func TestAuthStatusNotAuthenticated(t *testing.T) {
	origConfigPath := configPath
	defer func() { configPath = origConfigPath }()
	configPath = filepath.Join(t.TempDir(), "config.yaml")

	// Clear any flag state
	flagToken = ""

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	rootCmd.SetArgs([]string{"auth", "status"})
	err := rootCmd.Execute()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	stdout := buf.String()

	if err != nil {
		t.Fatalf("execute error: %v", err)
	}
	if !strings.Contains(stdout, "Not authenticated") {
		t.Errorf("expected 'Not authenticated', got: %s", stdout)
	}
}

func TestAuthStatusWithToken(t *testing.T) {
	origConfigPath := configPath
	defer func() { configPath = origConfigPath; flagToken = "" }()

	dir := t.TempDir()
	configPath = filepath.Join(dir, "config.yaml")

	// Save a token to config
	cfg := &config.Config{Token: "abcd1234efgh5678"}
	config.Save(cfg, configPath)

	flagToken = ""

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	rootCmd.SetArgs([]string{"auth", "status"})
	err := rootCmd.Execute()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	stdout := buf.String()

	if err != nil {
		t.Fatalf("execute error: %v", err)
	}
	if !strings.Contains(stdout, "Authenticated") {
		t.Errorf("expected 'Authenticated', got: %s", stdout)
	}
	// Should show masked token, NOT the full token
	if strings.Contains(stdout, "abcd1234efgh5678") {
		t.Error("full token displayed — should be masked")
	}
}

func TestAuthLogout(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	origConfigPath := configPath
	defer func() { configPath = origConfigPath }()
	configPath = path

	// Save a token first
	cfg := &config.Config{Token: "to-be-removed"}
	config.Save(cfg, path)

	rootCmd.SetArgs([]string{"auth", "logout"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("execute error: %v", err)
	}

	// Verify token was removed
	cfg, _ = config.Load(path)
	if cfg.Token != "" {
		t.Errorf("token not removed: %q", cfg.Token)
	}
}
