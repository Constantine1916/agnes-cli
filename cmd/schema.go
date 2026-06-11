package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/Constantine1916/agnes-cli/internal/clierrors"
	"github.com/Constantine1916/agnes-cli/internal/schema"
)

func newSchemaCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "schema <command>",
		Short: "Print machine-readable command schema",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "image.generate":
				return printJSON(schema.ImageGenerate())
			case "video.generate":
				return printJSON(schema.VideoGenerate())
			default:
				return clierrors.Validation("unknown_schema", fmt.Sprintf("unknown schema %q", args[0]), "command", "use image.generate or video.generate")
			}
		},
	}
}
