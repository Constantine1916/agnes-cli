package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/Constantine1916/agnes-cli/internal/clierrors"
	"github.com/Constantine1916/agnes-cli/internal/config"
)

func newKeyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "key",
		Short: "Manage the local Agnes API key",
	}
	cmd.AddCommand(newKeySetCmd())
	cmd.AddCommand(newKeyStatusCmd())
	cmd.AddCommand(newKeyClearCmd())
	return cmd
}

func newKeySetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set <api-key>",
		Short: "Save an Agnes API key locally",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := config.SaveAPIKey(args[0]); err != nil {
				return clierrors.FileIO(err)
			}
			fmt.Fprintln(os.Stderr, "Agnes API key saved")
			return nil
		},
	}
}

func newKeyStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show whether an Agnes API key is configured",
		RunE: func(cmd *cobra.Command, args []string) error {
			resolved, err := config.ResolveAPIKey(globals.APIKey)
			if err != nil {
				return clierrors.FileIO(err)
			}
			status := "missing"
			if resolved.Key != "" {
				status = "configured"
			}
			return printJSON(map[string]any{
				"ok":     true,
				"status": status,
				"source": resolved.Source,
			})
		},
	}
}

func newKeyClearCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "clear",
		Short: "Remove the saved Agnes API key",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := config.ClearAPIKey(); err != nil {
				return clierrors.FileIO(err)
			}
			fmt.Fprintln(os.Stderr, "Saved Agnes API key cleared")
			return nil
		},
	}
}
