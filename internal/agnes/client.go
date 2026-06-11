package agnes

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Constantine1916/agnes-cli/internal/buildinfo"
	"github.com/Constantine1916/agnes-cli/internal/clierrors"
)

const DefaultBaseURL = "https://apihub.agnes-ai.com"

type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func NewClient(baseURL, apiKey string) *Client {
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 60 * time.Second},
	}
}

func (c *Client) do(ctx context.Context, method, path string, body any, out any) error {
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", buildinfo.UserAgent())
	req.Header.Set("X-Source", "cli")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return clierrors.Network(err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		return clierrors.API(resp.StatusCode, parseErrorMessage(resp.StatusCode, respBody))
	}
	if out == nil {
		return nil
	}
	if len(respBody) == 0 {
		return nil
	}
	if err := json.Unmarshal(respBody, out); err != nil {
		return fmt.Errorf("decode Agnes response: %w", err)
	}
	return nil
}

func parseErrorMessage(status int, body []byte) string {
	var obj map[string]any
	if json.Unmarshal(body, &obj) == nil {
		for _, key := range []string{"message", "error", "detail"} {
			if v, ok := obj[key].(string); ok && v != "" {
				return fmt.Sprintf("Agnes API HTTP %d: %s", status, v)
			}
		}
	}
	if len(body) > 0 {
		return fmt.Sprintf("Agnes API HTTP %d: %s", status, string(body))
	}
	return fmt.Sprintf("Agnes API HTTP %d", status)
}

type ImageRequest struct {
	Model     string         `json:"model"`
	Prompt    string         `json:"prompt"`
	Size      string         `json:"size"`
	ExtraBody map[string]any `json:"extra_body"`
}

type ImageResponse struct {
	Created int64 `json:"created,omitempty"`
	Data    []struct {
		URL           string  `json:"url"`
		B64JSON       *string `json:"b64_json"`
		RevisedPrompt *string `json:"revised_prompt"`
	} `json:"data"`
}

func (c *Client) GenerateImage(ctx context.Context, req ImageRequest) (*ImageResponse, error) {
	var out ImageResponse
	if err := c.do(ctx, http.MethodPost, "/v1/images/generations", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

type VideoRequest struct {
	Model     string         `json:"model"`
	Prompt    string         `json:"prompt"`
	Image     string         `json:"image,omitempty"`
	Mode      string         `json:"mode,omitempty"`
	Height    int            `json:"height,omitempty"`
	Width     int            `json:"width,omitempty"`
	NumFrames int            `json:"num_frames,omitempty"`
	FrameRate int            `json:"frame_rate,omitempty"`
	ExtraBody map[string]any `json:"extra_body,omitempty"`
}

type VideoStatus struct {
	ID                 string `json:"id"`
	TaskID             string `json:"task_id,omitempty"`
	VideoID            string `json:"video_id,omitempty"`
	Object             string `json:"object,omitempty"`
	Model              string `json:"model,omitempty"`
	Status             string `json:"status"`
	Progress           int    `json:"progress,omitempty"`
	CreatedAt          int64  `json:"created_at,omitempty"`
	Seconds            string `json:"seconds,omitempty"`
	Size               string `json:"size,omitempty"`
	URL                string `json:"url,omitempty"`
	RemixedFromVideoID string `json:"remixed_from_video_id,omitempty"`
	Error              any    `json:"error,omitempty"`
}

func (c *Client) CreateVideo(ctx context.Context, req VideoRequest) (*VideoStatus, error) {
	var out VideoStatus
	if err := c.do(ctx, http.MethodPost, "/v1/videos", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) GetVideoStatus(ctx context.Context, id string, model string) (*VideoStatus, error) {
	status, err := c.getVideoByVideoID(ctx, id, model)
	if err == nil {
		return status, nil
	}
	if ce, ok := err.(*clierrors.CLIError); !ok || ce.Code != http.StatusNotFound {
		return nil, err
	}
	var legacy VideoStatus
	if legacyErr := c.do(ctx, http.MethodGet, "/v1/videos/"+url.PathEscape(id), nil, &legacy); legacyErr != nil {
		return nil, err
	}
	return &legacy, nil
}

func (c *Client) getVideoByVideoID(ctx context.Context, videoID string, model string) (*VideoStatus, error) {
	q := url.Values{}
	q.Set("video_id", videoID)
	if model != "" {
		q.Set("model_name", model)
	}
	var out VideoStatus
	if err := c.do(ctx, http.MethodGet, "/agnesapi?"+q.Encode(), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (v *VideoStatus) ResultURL() string {
	if v == nil {
		return ""
	}
	if v.URL != "" {
		return v.URL
	}
	return v.RemixedFromVideoID
}

func (v *VideoStatus) StableID() string {
	if v == nil {
		return ""
	}
	if v.VideoID != "" {
		return v.VideoID
	}
	if v.TaskID != "" {
		return v.TaskID
	}
	return v.ID
}

func (v *VideoStatus) FailureReason() string {
	if v == nil || v.Error == nil {
		return "unknown error"
	}
	switch val := v.Error.(type) {
	case string:
		return val
	case map[string]any:
		for _, key := range []string{"message", "error_message", "error"} {
			if s, ok := val[key].(string); ok && s != "" {
				return s
			}
		}
	}
	return fmt.Sprintf("%v", v.Error)
}
