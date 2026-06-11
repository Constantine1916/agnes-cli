package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/Constantine1916/agnes-cli/internal/agnes"
	"github.com/Constantine1916/agnes-cli/internal/clierrors"
)

const (
	defaultImageModel = "agnes-image-2.1-flash"
	imageModel20      = "agnes-image-2.0-flash"
)

type imageGenerateOptions struct {
	Prompt string
	Size   string
	Model  string
	Images []string
}

func newImageCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "image",
		Short: "Generate Agnes images",
	}
	cmd.AddCommand(newImageGenerateCmd())
	return cmd
}

func newImageGenerateCmd() *cobra.Command {
	opts := &imageGenerateOptions{}
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate or edit an image and print result URL(s)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runImageGenerate(opts)
		},
	}
	cmd.Flags().StringVar(&opts.Prompt, "prompt", "", "Text prompt or image editing instruction")
	cmd.Flags().StringVar(&opts.Size, "size", "1024x768", "Output image size")
	cmd.Flags().StringVar(&opts.Model, "model", defaultImageModel, "Image model: agnes-image-2.1-flash or agnes-image-2.0-flash")
	cmd.Flags().StringArrayVar(&opts.Images, "image", nil, "Optional image URL, Data URI, or local path (repeatable)")
	return cmd
}

func runImageGenerate(opts *imageGenerateOptions) error {
	if opts.Prompt == "" {
		return clierrors.Validation("missing_required_flag", "missing --prompt", "--prompt", "pass --prompt with the image instruction")
	}
	if opts.Size == "" {
		return clierrors.Validation("missing_required_flag", "missing --size", "--size", "pass --size, for example 1024x768")
	}
	if opts.Model != defaultImageModel && opts.Model != imageModel20 {
		return clierrors.Validation("invalid_model", fmt.Sprintf("unsupported image model %q", opts.Model), "--model", "use agnes-image-2.1-flash or agnes-image-2.0-flash")
	}

	media, err := normalizeInputs(opts.Images)
	if err != nil {
		return err
	}
	extra := map[string]any{"response_format": "url"}
	if len(media) > 0 {
		extra["image"] = media
	}
	request := agnes.ImageRequest{
		Model:     opts.Model,
		Prompt:    opts.Prompt,
		Size:      opts.Size,
		ExtraBody: extra,
	}
	if globals.DryRun {
		return dryRun(map[string]any{
			"method": "POST",
			"url":    resolveBaseURL() + "/v1/images/generations",
			"body":   request,
		})
	}

	client, _, err := requireClient()
	if err != nil {
		return err
	}
	ctx, cancel := contextWithTimeout(360)
	defer cancel()
	response, err := client.GenerateImage(ctx, request)
	if err != nil {
		return err
	}
	if len(response.Data) == 0 {
		return clierrors.API(500, "Agnes image response did not include data")
	}
	for _, item := range response.Data {
		if item.URL != "" {
			fmt.Fprintln(os.Stdout, item.URL)
		}
	}
	return nil
}

func normalizeInputs(inputs []string) ([]string, error) {
	out := make([]string, 0, len(inputs))
	for _, input := range inputs {
		normalized, err := agnes.NormalizeMediaInput(input)
		if err != nil {
			return nil, clierrors.FileIO(err)
		}
		out = append(out, normalized)
	}
	return out, nil
}
