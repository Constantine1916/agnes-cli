package schema

type CommandSchema struct {
	Command     string         `json:"command"`
	Description string         `json:"description"`
	Endpoint    string         `json:"endpoint"`
	Method      string         `json:"method"`
	Models      []string       `json:"models"`
	Flags       []FlagSchema   `json:"flags"`
	Output      map[string]any `json:"output"`
}

type FlagSchema struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Required    bool     `json:"required,omitempty"`
	Default     any      `json:"default,omitempty"`
	Enum        []string `json:"enum,omitempty"`
	Description string   `json:"description"`
}

func ImageGenerate() CommandSchema {
	return CommandSchema{
		Command:     "agnes image generate",
		Description: "Generate or edit an image with Agnes Image Flash and print result URL(s).",
		Endpoint:    "POST https://apihub.agnes-ai.com/v1/images/generations",
		Method:      "POST",
		Models:      []string{"agnes-image-2.1-flash", "agnes-image-2.0-flash"},
		Flags: []FlagSchema{
			{Name: "--prompt", Type: "string", Required: true, Description: "Text prompt or editing instruction."},
			{Name: "--size", Type: "string", Required: true, Default: "1024x768", Description: "Output image size, for example 1024x768."},
			{Name: "--model", Type: "string", Default: "agnes-image-2.1-flash", Enum: []string{"agnes-image-2.1-flash", "agnes-image-2.0-flash"}, Description: "Agnes image model."},
			{Name: "--image", Type: "string[]", Description: "Optional public URL, Data URI, or local path. Local paths are sent as Data URI."},
		},
		Output: map[string]any{"stdout": "one result URL per line", "stderr": "diagnostics only"},
	}
}

func VideoGenerate() CommandSchema {
	return CommandSchema{
		Command:     "agnes video generate",
		Description: "Create a video task, poll until completion, and print the final video URL.",
		Endpoint:    "POST https://apihub.agnes-ai.com/v1/videos",
		Method:      "POST",
		Models:      []string{"agnes-video-v2.0"},
		Flags: []FlagSchema{
			{Name: "--prompt", Type: "string", Required: true, Description: "Text description of the video."},
			{Name: "--image", Type: "string[]", Description: "Optional public URL, Data URI, or local path for image-to-video, multi-image video, or keyframes."},
			{Name: "--mode", Type: "string", Enum: []string{"keyframes"}, Description: "Optional extra video mode."},
			{Name: "--width", Type: "integer", Default: 1152, Description: "Video width."},
			{Name: "--height", Type: "integer", Default: 768, Description: "Video height."},
			{Name: "--num-frames", Type: "integer", Default: 121, Description: "Frame count. Must be <= 441 and match 8n + 1."},
			{Name: "--frame-rate", Type: "integer", Default: 24, Description: "Frames per second, 1-60."},
		},
		Output: map[string]any{"stdout": "final video URL", "stderr": "task id, status, and progress"},
	}
}
