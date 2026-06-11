package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/Constantine1916/agnes-cli/internal/agnes"
	"github.com/Constantine1916/agnes-cli/internal/buildinfo"
	"github.com/Constantine1916/agnes-cli/internal/clierrors"
	"github.com/Constantine1916/agnes-cli/internal/config"
)

type GlobalOptions struct {
	APIKey  string
	BaseURL string
	DryRun  bool
}

var globals GlobalOptions

func Execute() int {
	root := newRootCmd()
	if err := root.Execute(); err != nil {
		var cliErr *clierrors.CLIError
		if errors.As(err, &cliErr) {
			clierrors.WriteJSON(os.Stderr, cliErr)
			return cliErr.Exit
		}
		fallback := clierrors.New(clierrors.ExitGeneric, "internal", "unknown", err.Error(), "")
		clierrors.WriteJSON(os.Stderr, fallback)
		return fallback.Exit
	}
	return clierrors.ExitOK
}

func newRootCmd() *cobra.Command {
	globals = GlobalOptions{}
	cmd := &cobra.Command{
		Use:           "agnes",
		Short:         "Agent-friendly CLI for Agnes image and video generation",
		Version:       buildinfo.Version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	cmd.PersistentFlags().StringVar(&globals.APIKey, "api-key", "", "Agnes API key (overrides AGNES_API_KEY and saved key)")
	cmd.PersistentFlags().StringVar(&globals.BaseURL, "base-url", "", "Agnes API base URL (default: AGNES_API_BASE_URL or https://apihub.agnes-ai.com)")
	cmd.PersistentFlags().BoolVar(&globals.DryRun, "dry-run", false, "Print request payload without making API calls")

	cmd.AddCommand(newKeyCmd())
	cmd.AddCommand(newImageCmd())
	cmd.AddCommand(newVideoCmd())
	cmd.AddCommand(newDoctorCmd())
	cmd.AddCommand(newSchemaCmd())
	return cmd
}

func resolveBaseURL() string {
	if globals.BaseURL != "" {
		return globals.BaseURL
	}
	if env := os.Getenv("AGNES_API_BASE_URL"); env != "" {
		return env
	}
	return agnes.DefaultBaseURL
}

func requireClient() (*agnes.Client, config.KeySource, error) {
	key, err := config.ResolveAPIKey(globals.APIKey)
	if err != nil {
		return nil, config.SourceNone, clierrors.FileIO(err)
	}
	if key.Key == "" {
		return nil, config.SourceNone, clierrors.Auth("API key not configured", "pass --api-key, set AGNES_API_KEY, or run: agnes key set <api-key>")
	}
	return agnes.NewClient(resolveBaseURL(), key.Key), key.Source, nil
}

func printJSON(v any) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(v)
}

func dryRun(payload any) error {
	return printJSON(map[string]any{
		"ok":      true,
		"dry_run": true,
		"request": payload,
	})
}

func contextWithTimeout(seconds int) (context.Context, context.CancelFunc) {
	if seconds <= 0 {
		seconds = 600
	}
	return context.WithTimeout(context.Background(), time.Duration(seconds)*time.Second)
}

func printProgress(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}
