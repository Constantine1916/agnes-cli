package config

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const keychainService = "agnes-cli"

type KeySource string

const (
	SourceFlag  KeySource = "flag"
	SourceEnv   KeySource = "env"
	SourceSaved KeySource = "saved"
	SourceNone  KeySource = "none"
)

type KeyResult struct {
	Key    string
	Source KeySource
}

type fileConfig struct {
	APIKey string `yaml:"api_key,omitempty"`
}

func ResolveAPIKey(flagValue string) (*KeyResult, error) {
	if flagValue != "" {
		return &KeyResult{Key: flagValue, Source: SourceFlag}, nil
	}
	if env := os.Getenv("AGNES_API_KEY"); env != "" {
		return &KeyResult{Key: env, Source: SourceEnv}, nil
	}
	saved, err := LoadSavedAPIKey()
	if err != nil {
		return nil, err
	}
	if saved != "" {
		return &KeyResult{Key: saved, Source: SourceSaved}, nil
	}
	return &KeyResult{Source: SourceNone}, nil
}

func SaveAPIKey(apiKey string) error {
	if apiKey == "" {
		return errors.New("api key is empty")
	}
	if shouldUseKeychain() {
		if err := saveKeychain(apiKey); err == nil {
			_ = removeFileConfig()
			return nil
		}
	}
	return saveFileAPIKey(apiKey)
}

func LoadSavedAPIKey() (string, error) {
	if shouldUseKeychain() {
		if key, err := loadKeychain(); err == nil && key != "" {
			return key, nil
		}
	}
	cfg, err := readFileConfig()
	if err != nil {
		return "", err
	}
	return cfg.APIKey, nil
}

func ClearAPIKey() error {
	_ = clearKeychain()
	return removeFileConfig()
}

func HasSavedAPIKey() (bool, error) {
	key, err := LoadSavedAPIKey()
	if err != nil {
		return false, err
	}
	return key != "", nil
}

func shouldUseKeychain() bool {
	if os.Getenv("AGNES_NO_KEYCHAIN") == "1" || os.Getenv("AGNES_CONFIG_DIR") != "" {
		return false
	}
	if runtime.GOOS == "linux" && os.Getenv("DBUS_SESSION_BUS_ADDRESS") == "" {
		return false
	}
	return runtime.GOOS == "darwin"
}

func ConfigPath() (string, error) {
	if dir := os.Getenv("AGNES_CONFIG_DIR"); dir != "" {
		return filepath.Join(dir, "config.yml"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "agnes", "config.yml"), nil
}

func saveFileAPIKey(apiKey string) error {
	path, err := ConfigPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	return os.WriteFile(path, []byte("api_key: "+quote(apiKey)+"\n"), 0o600)
}

func readFileConfig() (*fileConfig, error) {
	path, err := ConfigPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return &fileConfig{}, nil
	}
	if err != nil {
		return nil, err
	}
	var cfg fileConfig
	cfg.APIKey = parseAPIKey(data)
	return &cfg, nil
}

func removeFileConfig() error {
	path, err := ConfigPath()
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func saveKeychain(apiKey string) error {
	return exec.Command(
		"security",
		"add-generic-password",
		"-a", "api_key",
		"-s", keychainService,
		"-w", apiKey,
		"-U",
	).Run()
}

func loadKeychain() (string, error) {
	out, err := exec.Command(
		"security",
		"find-generic-password",
		"-a", "api_key",
		"-s", keychainService,
		"-w",
	).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func clearKeychain() error {
	return exec.Command(
		"security",
		"delete-generic-password",
		"-a", "api_key",
		"-s", keychainService,
	).Run()
}

func quote(value string) string {
	escaped := strings.ReplaceAll(value, "\\", "\\\\")
	escaped = strings.ReplaceAll(escaped, "\"", "\\\"")
	return "\"" + escaped + "\""
}

func parseAPIKey(data []byte) string {
	for _, line := range bytes.Split(data, []byte("\n")) {
		text := strings.TrimSpace(string(line))
		if !strings.HasPrefix(text, "api_key:") {
			continue
		}
		value := strings.TrimSpace(strings.TrimPrefix(text, "api_key:"))
		value = strings.Trim(value, "\"")
		value = strings.ReplaceAll(value, "\\\"", "\"")
		value = strings.ReplaceAll(value, "\\\\", "\\")
		return value
	}
	return ""
}
