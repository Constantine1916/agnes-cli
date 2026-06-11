package cmd

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestKeyStatusDoesNotLeakKey(t *testing.T) {
	t.Setenv("AGNES_CONFIG_DIR", t.TempDir())
	t.Setenv("AGNES_API_KEY", "")
	if err := newKeySetCmd().RunE(nil, []string{"secret-key"}); err != nil {
		t.Fatal(err)
	}

	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	runErr := newKeyStatusCmd().RunE(nil, nil)
	_ = w.Close()
	os.Stdout = old
	if runErr != nil {
		t.Fatal(runErr)
	}
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	out := buf.String()
	if strings.Contains(out, "secret-key") {
		t.Fatalf("key status leaked secret: %s", out)
	}
	if !strings.Contains(out, `"status": "configured"`) {
		t.Fatalf("unexpected status output: %s", out)
	}
}
