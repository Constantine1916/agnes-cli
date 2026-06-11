package agnes

import (
	"encoding/base64"
	"fmt"
	"mime"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

func NormalizeMediaInput(input string) (string, error) {
	if input == "" {
		return "", nil
	}
	if isRemoteOrData(input) {
		return input, nil
	}
	data, err := os.ReadFile(input)
	if err != nil {
		return "", err
	}
	mimeType := mime.TypeByExtension(strings.ToLower(filepath.Ext(input)))
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	return fmt.Sprintf("data:%s;base64,%s", mimeType, base64.StdEncoding.EncodeToString(data)), nil
}

func isRemoteOrData(input string) bool {
	if strings.HasPrefix(input, "data:") {
		return true
	}
	u, err := url.Parse(input)
	return err == nil && (u.Scheme == "http" || u.Scheme == "https")
}
