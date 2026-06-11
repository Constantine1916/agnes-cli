package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/Constantine1916/agnes-cli/internal/agnes"
	"github.com/Constantine1916/agnes-cli/internal/clierrors"
)

const defaultVideoModel = "agnes-video-v2.0"

type videoGenerateOptions struct {
	Prompt       string
	Images       []string
	Mode         string
	Width        int
	Height       int
	NumFrames    int
	FrameRate    int
	Model        string
	Timeout      int
	PollInterval int
}

type videoStatusOptions struct {
	Model        string
	Wait         bool
	Timeout      int
	PollInterval int
}

func newVideoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "video",
		Short: "Generate and inspect Agnes videos",
	}
	cmd.AddCommand(newVideoGenerateCmd())
	cmd.AddCommand(newVideoStatusCmd())
	return cmd
}

func newVideoGenerateCmd() *cobra.Command {
	opts := &videoGenerateOptions{}
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate a video, wait for completion, and print the result URL",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVideoGenerate(opts)
		},
	}
	cmd.Flags().StringVar(&opts.Prompt, "prompt", "", "Text prompt for the video")
	cmd.Flags().StringArrayVar(&opts.Images, "image", nil, "Optional image URL, Data URI, or local path (repeatable)")
	cmd.Flags().StringVar(&opts.Mode, "mode", "", "Optional mode, currently supports keyframes")
	cmd.Flags().IntVar(&opts.Width, "width", 1152, "Video width")
	cmd.Flags().IntVar(&opts.Height, "height", 768, "Video height")
	cmd.Flags().IntVar(&opts.NumFrames, "num-frames", 121, "Frame count; must be <= 441 and match 8n + 1")
	cmd.Flags().IntVar(&opts.FrameRate, "frame-rate", 24, "Frames per second, 1-60")
	cmd.Flags().StringVar(&opts.Model, "model", defaultVideoModel, "Video model")
	cmd.Flags().IntVar(&opts.Timeout, "timeout", 600, "Maximum seconds to wait")
	cmd.Flags().IntVar(&opts.PollInterval, "poll-interval", 5, "Seconds between status checks")
	return cmd
}

func newVideoStatusCmd() *cobra.Command {
	opts := &videoStatusOptions{}
	cmd := &cobra.Command{
		Use:   "status <video_id_or_task_id>",
		Short: "Check video task status and print the URL when complete",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVideoStatus(args[0], opts)
		},
	}
	cmd.Flags().StringVar(&opts.Model, "model", defaultVideoModel, "Video model name for video_id status lookup")
	cmd.Flags().BoolVar(&opts.Wait, "wait", false, "Poll until the task reaches a terminal state")
	cmd.Flags().IntVar(&opts.Timeout, "timeout", 600, "Maximum seconds to wait when --wait is set")
	cmd.Flags().IntVar(&opts.PollInterval, "poll-interval", 5, "Seconds between status checks")
	return cmd
}

func runVideoGenerate(opts *videoGenerateOptions) error {
	if opts.Prompt == "" {
		return clierrors.Validation("missing_required_flag", "missing --prompt", "--prompt", "pass --prompt with the video instruction")
	}
	if err := validateVideoOptions(opts.NumFrames, opts.FrameRate, opts.Mode, len(opts.Images)); err != nil {
		return err
	}
	media, err := normalizeInputs(opts.Images)
	if err != nil {
		return err
	}
	request := buildVideoRequest(opts, media)
	if globals.DryRun {
		return dryRun(map[string]any{
			"method": "POST",
			"url":    resolveBaseURL() + "/v1/videos",
			"body":   request,
		})
	}
	client, _, err := requireClient()
	if err != nil {
		return err
	}
	ctx, cancel := contextWithTimeout(opts.Timeout)
	defer cancel()
	created, err := client.CreateVideo(ctx, request)
	if err != nil {
		return err
	}
	id := created.StableID()
	printProgress("Task submitted: %s", id)
	result, err := pollVideo(ctx, client, id, opts.Model, opts.PollInterval)
	if err != nil {
		return err
	}
	url := result.ResultURL()
	if url == "" {
		return clierrors.API(500, "completed video response did not include a result URL")
	}
	fmt.Fprintln(os.Stdout, url)
	return nil
}

func runVideoStatus(id string, opts *videoStatusOptions) error {
	client, _, err := requireClient()
	if err != nil {
		return err
	}
	if opts.Wait {
		ctx, cancel := contextWithTimeout(opts.Timeout)
		defer cancel()
		result, err := pollVideo(ctx, client, id, opts.Model, opts.PollInterval)
		if err != nil {
			return err
		}
		if url := result.ResultURL(); url != "" {
			fmt.Fprintln(os.Stdout, url)
		}
		return nil
	}
	ctx, cancel := contextWithTimeout(60)
	defer cancel()
	status, err := client.GetVideoStatus(ctx, id, opts.Model)
	if err != nil {
		return err
	}
	printProgress("Status: %s", status.Status)
	printProgress("Progress: %d%%", status.Progress)
	if status.Status == "failed" {
		return clierrors.TaskFailed(status.StableID(), status.FailureReason())
	}
	if url := status.ResultURL(); url != "" {
		fmt.Fprintln(os.Stdout, url)
	}
	return nil
}

func buildVideoRequest(opts *videoGenerateOptions, media []string) agnes.VideoRequest {
	req := agnes.VideoRequest{
		Model:     opts.Model,
		Prompt:    opts.Prompt,
		Mode:      opts.Mode,
		Width:     opts.Width,
		Height:    opts.Height,
		NumFrames: opts.NumFrames,
		FrameRate: opts.FrameRate,
	}
	if len(media) == 1 && opts.Mode == "" {
		req.Image = media[0]
		return req
	}
	if len(media) > 0 || opts.Mode != "" {
		req.ExtraBody = map[string]any{}
	}
	if len(media) > 0 {
		req.ExtraBody["image"] = media
	}
	if opts.Mode != "" {
		req.ExtraBody["mode"] = opts.Mode
	}
	return req
}

func pollVideo(ctx context.Context, client *agnes.Client, id, model string, intervalSeconds int) (*agnes.VideoStatus, error) {
	if intervalSeconds <= 0 {
		intervalSeconds = 5
	}
	interval := time.Duration(intervalSeconds) * time.Second
	var lastProgress = -1
	for {
		status, err := client.GetVideoStatus(ctx, id, model)
		if err != nil {
			return nil, err
		}
		if status.Progress != lastProgress || status.Status == "queued" {
			printProgress("Status: %s", status.Status)
			printProgress("Progress: %d%%", status.Progress)
			lastProgress = status.Progress
		}
		switch status.Status {
		case "completed":
			return status, nil
		case "failed":
			return nil, clierrors.TaskFailed(status.StableID(), status.FailureReason())
		}
		select {
		case <-ctx.Done():
			return nil, clierrors.Timeout(id)
		case <-time.After(interval):
		}
	}
}

func validateVideoOptions(numFrames, frameRate int, mode string, imageCount int) error {
	if numFrames <= 0 || numFrames > 441 || (numFrames-1)%8 != 0 {
		return clierrors.Validation("invalid_num_frames", "--num-frames must be <= 441 and match 8n + 1", "--num-frames", "use values such as 81, 121, 161, 241, or 441")
	}
	if frameRate < 1 || frameRate > 60 {
		return clierrors.Validation("invalid_frame_rate", "--frame-rate must be between 1 and 60", "--frame-rate", "use a value such as 24 or 30")
	}
	if mode != "" && mode != "keyframes" {
		return clierrors.Validation("invalid_mode", fmt.Sprintf("unsupported video mode %q", mode), "--mode", "use --mode keyframes or omit --mode")
	}
	if mode == "keyframes" && imageCount < 2 {
		return clierrors.Validation("invalid_keyframes", "--mode keyframes requires at least two --image values", "--image", "pass starting and ending keyframe images")
	}
	return nil
}
