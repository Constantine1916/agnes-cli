package agnes

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGenerateImagePayloadAndURLResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/images/generations" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Fatalf("unexpected auth header: %s", got)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body["model"] != "agnes-image-2.1-flash" || body["prompt"] != "hello" || body["size"] != "1024x768" {
			t.Fatalf("unexpected body: %#v", body)
		}
		extra, ok := body["extra_body"].(map[string]any)
		if !ok || extra["response_format"] != "url" {
			t.Fatalf("missing response_format=url: %#v", body)
		}
		_, _ = w.Write([]byte(`{"created":1780000000,"data":[{"url":"https://example.com/out.png","b64_json":null,"revised_prompt":null}]}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-key")
	resp, err := client.GenerateImage(context.Background(), ImageRequest{
		Model:     "agnes-image-2.1-flash",
		Prompt:    "hello",
		Size:      "1024x768",
		ExtraBody: map[string]any{"response_format": "url"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Data) != 1 || resp.Data[0].URL != "https://example.com/out.png" {
		t.Fatalf("unexpected response: %#v", resp)
	}
}

func TestVideoStatusPrefersVideoIDAndFallsBackToTaskID(t *testing.T) {
	var agnesAPICalls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/agnesapi":
			agnesAPICalls++
			if r.URL.Query().Get("video_id") != "task_1" {
				t.Fatalf("unexpected video id query: %s", r.URL.RawQuery)
			}
			http.NotFound(w, r)
		case "/v1/videos/task_1":
			_, _ = w.Write([]byte(`{"id":"task_1","video_id":"video_1","status":"completed","progress":100,"remixed_from_video_id":"https://example.com/out.mp4"}`))
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-key")
	status, err := client.GetVideoStatus(context.Background(), "task_1", "agnes-video-v2.0")
	if err != nil {
		t.Fatal(err)
	}
	if agnesAPICalls != 1 {
		t.Fatalf("expected agnesapi to be tried once, got %d", agnesAPICalls)
	}
	if status.ResultURL() != "https://example.com/out.mp4" {
		t.Fatalf("unexpected result URL: %#v", status)
	}
}
