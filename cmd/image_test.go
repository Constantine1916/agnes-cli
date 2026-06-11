package cmd

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func captureStdout(t *testing.T, fn func() error) (string, error) {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	runErr := fn()
	_ = w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	return buf.String(), runErr
}

func TestImageGenerateDryRunDoesNotRequireAPIKey(t *testing.T) {
	t.Setenv("AGNES_API_KEY", "")
	t.Setenv("AGNES_CONFIG_DIR", t.TempDir())
	globals = GlobalOptions{DryRun: true}
	out, err := captureStdout(t, func() error {
		return runImageGenerate(&imageGenerateOptions{
			Prompt: "hello",
			Size:   "1024x768",
			Model:  defaultImageModel,
		})
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, `"dry_run": true`) || !strings.Contains(out, `"response_format": "url"`) {
		t.Fatalf("unexpected dry-run output: %s", out)
	}
}
