package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveAPIKeyPrecedence(t *testing.T) {
	t.Setenv("AGNES_CONFIG_DIR", t.TempDir())
	t.Setenv("AGNES_API_KEY", "env-key")
	if err := SaveAPIKey("saved-key"); err != nil {
		t.Fatal(err)
	}

	got, err := ResolveAPIKey("flag-key")
	if err != nil {
		t.Fatal(err)
	}
	if got.Key != "flag-key" || got.Source != SourceFlag {
		t.Fatalf("flag precedence mismatch: %#v", got)
	}

	got, err = ResolveAPIKey("")
	if err != nil {
		t.Fatal(err)
	}
	if got.Key != "env-key" || got.Source != SourceEnv {
		t.Fatalf("env precedence mismatch: %#v", got)
	}
}

func TestSavedAPIKeyFileFallback(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("AGNES_CONFIG_DIR", dir)
	t.Setenv("AGNES_API_KEY", "")

	if err := SaveAPIKey("saved-key"); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "config.yml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "api_key:") {
		t.Fatalf("config file did not include api_key: %s", string(data))
	}

	got, err := ResolveAPIKey("")
	if err != nil {
		t.Fatal(err)
	}
	if got.Key != "saved-key" || got.Source != SourceSaved {
		t.Fatalf("saved key mismatch: %#v", got)
	}
}
