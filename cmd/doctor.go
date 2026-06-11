package cmd

import (
	"context"
	"net/http"
	"time"

	"github.com/spf13/cobra"

	"github.com/Constantine1916/agnes-cli/internal/config"
)

func newDoctorCmd() *cobra.Command {
	offline := false
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Check Agnes CLI configuration and connectivity",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDoctor(offline)
		},
	}
	cmd.Flags().BoolVar(&offline, "offline", false, "Skip network checks")
	return cmd
}

func runDoctor(offline bool) error {
	key, err := config.ResolveAPIKey(globals.APIKey)
	keyStatus := "missing"
	keySource := config.SourceNone
	if err == nil && key.Key != "" {
		keyStatus = "configured"
		keySource = key.Source
	}
	checks := []map[string]any{
		{"name": "api_key", "status": keyStatus, "source": keySource},
		{"name": "base_url", "status": "configured", "value": resolveBaseURL()},
	}
	if offline {
		checks = append(checks, map[string]any{"name": "endpoint", "status": "skip", "message": "offline mode"})
		return printJSON(map[string]any{"ok": true, "checks": checks})
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, resolveBaseURL(), nil)
	if reqErr != nil {
		checks = append(checks, map[string]any{"name": "endpoint", "status": "fail", "message": reqErr.Error()})
		return printJSON(map[string]any{"ok": true, "checks": checks})
	}
	resp, httpErr := http.DefaultClient.Do(req)
	if httpErr != nil {
		checks = append(checks, map[string]any{"name": "endpoint", "status": "fail", "message": httpErr.Error()})
	} else {
		_ = resp.Body.Close()
		checks = append(checks, map[string]any{"name": "endpoint", "status": "pass", "code": resp.StatusCode})
	}
	return printJSON(map[string]any{"ok": true, "checks": checks})
}
