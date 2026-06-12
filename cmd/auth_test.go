package cmd

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/Constantine1916/agnes-cli/internal/clierrors"
)

func TestRequireClientMissingKeyHintCoversExistingAndNewUsers(t *testing.T) {
	t.Setenv("AGNES_API_KEY", "")
	t.Setenv("AGNES_CONFIG_DIR", t.TempDir())
	globals = GlobalOptions{}

	_, _, err := requireClient()
	var cliErr *clierrors.CLIError
	if !errors.As(err, &cliErr) {
		t.Fatalf("expected CLIError, got %T: %v", err, err)
	}
	if cliErr.Subtype != "api_key_missing" {
		t.Fatalf("expected api_key_missing, got %q", cliErr.Subtype)
	}
	for _, want := range []string{
		"agnes key set <api-key>",
		"https://platform.agnes-ai.com/",
	} {
		if !strings.Contains(cliErr.Hint, want) {
			t.Fatalf("missing %q in hint: %q", want, cliErr.Hint)
		}
	}
}

func TestPromptForMissingAPIKeyOpensPlatformAfterEnter(t *testing.T) {
	var stderr bytes.Buffer
	var opened string

	err := promptForMissingAPIKeyAndOpen(
		strings.NewReader("\n"),
		&stderr,
		"agnes-cli",
		func(url string) error {
			opened = url
			return nil
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if opened != agnesPlatformURL {
		t.Fatalf("expected opened URL %q, got %q", agnesPlatformURL, opened)
	}
	out := stderr.String()
	for _, want := range []string{
		"If you already have an API key:",
		"agnes-cli key set <api-key>",
		"If you don't have one yet:",
		"https://platform.agnes-ai.com/",
		"Press Enter to open",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in prompt:\n%s", want, out)
		}
	}
}

func TestKeyRegisterOpensPlatform(t *testing.T) {
	var opened string
	oldOpen := openExternalURL
	openExternalURL = func(url string) error {
		opened = url
		return nil
	}
	t.Cleanup(func() {
		openExternalURL = oldOpen
	})

	if err := newKeyRegisterCmd().RunE(nil, nil); err != nil {
		t.Fatal(err)
	}
	if opened != agnesPlatformURL {
		t.Fatalf("expected opened URL %q, got %q", agnesPlatformURL, opened)
	}
}
