package cmd

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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

const agnesPlatformURL = "https://platform.agnes-ai.com/"

var globals GlobalOptions
var openExternalURL = openURL

func Execute() int {
	root := newRootCmd()
	if err := root.Execute(); err != nil {
		var cliErr *clierrors.CLIError
		if errors.As(err, &cliErr) {
			if isMissingAPIKey(cliErr) && isInteractiveTerminal() {
				if promptErr := promptForMissingAPIKeyAndOpen(os.Stdin, os.Stderr, invokedCommandName(), openExternalURL); promptErr != nil {
					fmt.Fprintf(os.Stderr, "Could not open Agnes platform: %v\nOpen this URL to get an Agnes API key:\n%s\n", promptErr, agnesPlatformURL)
				}
				return cliErr.Exit
			}
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
		return nil, config.SourceNone, clierrors.Auth("API key not configured", missingAPIKeyHint("agnes"))
	}
	return agnes.NewClient(resolveBaseURL(), key.Key), key.Source, nil
}

func missingAPIKeyHint(commandName string) string {
	if commandName == "" {
		commandName = "agnes"
	}
	return fmt.Sprintf("If you already have an API key, run: %s key set <api-key>. If not, get one at: %s", commandName, agnesPlatformURL)
}

func isMissingAPIKey(err *clierrors.CLIError) bool {
	return err != nil && err.Type == "authentication" && err.Subtype == "api_key_missing"
}

func isInteractiveTerminal() bool {
	return isTerminal(os.Stdin) && isTerminal(os.Stderr)
}

func isTerminal(file *os.File) bool {
	info, err := file.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}

func invokedCommandName() string {
	name := filepath.Base(os.Args[0])
	if name == "" || name == "." {
		return "agnes"
	}
	return name
}

func promptForMissingAPIKeyAndOpen(stdin io.Reader, stderr io.Writer, commandName string, opener func(string) error) error {
	if commandName == "" {
		commandName = "agnes"
	}
	fmt.Fprintln(stderr, "Agnes API key is not configured.")
	fmt.Fprintln(stderr)
	fmt.Fprintln(stderr, "If you already have an API key:")
	fmt.Fprintf(stderr, "  %s key set <api-key>\n", commandName)
	fmt.Fprintln(stderr)
	fmt.Fprintln(stderr, "If you don't have one yet:")
	fmt.Fprintf(stderr, "  %s\n", agnesPlatformURL)
	fmt.Fprintln(stderr)
	fmt.Fprintln(stderr, "Press Enter to open the Agnes platform, or Ctrl+C to cancel.")

	if _, err := bufio.NewReader(stdin).ReadString('\n'); err != nil {
		return err
	}
	return opener(agnesPlatformURL)
}

func openURL(rawURL string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", rawURL).Start()
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", rawURL).Start()
	default:
		return exec.Command("xdg-open", rawURL).Start()
	}
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
