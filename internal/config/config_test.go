package config

import (
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
		name string
		flag string
		env  string
		file string
		want string
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
