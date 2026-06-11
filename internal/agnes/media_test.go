package agnes

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNormalizeMediaInputPathBecomesDataURI(t *testing.T) {
	path := filepath.Join(t.TempDir(), "input.png")
	if err := os.WriteFile(path, []byte("png-data"), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := NormalizeMediaInput(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(got, "data:image/png;base64,") {
		t.Fatalf("expected png data URI, got %q", got)
	}
}

func TestNormalizeMediaInputKeepsURLsAndDataURIs(t *testing.T) {
	for _, input := range []string{"https://example.com/a.png", "data:image/png;base64,abc"} {
		got, err := NormalizeMediaInput(input)
		if err != nil {
			t.Fatal(err)
		}
		if got != input {
			t.Fatalf("expected %q, got %q", input, got)
		}
	}
}
