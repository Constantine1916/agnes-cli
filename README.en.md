# agnes-cli

[中文](https://github.com/Constantine1916/agnes-cli/blob/main/README.md) | **English**

Agent-friendly Agnes multimodal CLI for calling Agnes image and video generation APIs with an API key.

The CLI is designed for humans and AI agents: stable commands, no login flow, API-key injection through flags or environment variables, final result URL(s) on stdout, and progress or diagnostics on stderr.

## Installation

```bash
npm install -g @agnes-ai/agnes-cli
agnes --version
agnes doctor --offline
```

You can also build from source:

```bash
go build -o bin/agnes .
./bin/agnes doctor --offline
```

## API Key

No login is required. API key precedence:

1. `--api-key`
2. `AGNES_API_KEY`
3. Saved key from `agnes key set`

Save an API key:

```bash
agnes key set "$AGNES_API_KEY"
agnes key status
```

Inject a key for one command:

```bash
AGNES_API_KEY="your_api_key" agnes image generate \
  --prompt "A clean product photo of a glass cube" \
  --size 1024x768
```

Clear the saved key:

```bash
agnes key clear
```

## Image Generation

```bash
agnes image generate \
  --prompt "A clean product photo of a glass cube" \
  --size 1024x768
```

Use image inputs:

```bash
agnes image generate \
  --prompt "Make it cinematic" \
  --image ./input.png \
  --model agnes-image-2.1-flash
```

Supported image models:

- `agnes-image-2.1-flash`, default
- `agnes-image-2.0-flash`

Local image paths are converted to Data URIs before the request is sent.

## Video Generation

```bash
agnes video generate \
  --prompt "A cat walking on the beach at sunset" \
  --num-frames 121 \
  --frame-rate 24
```

Keyframe mode:

```bash
agnes video generate \
  --prompt "Smooth transition between keyframes" \
  --image ./start.png \
  --image ./end.png \
  --mode keyframes
```

Check task status:

```bash
agnes video status <video_id_or_task_id>
```

## Agent-Friendly Output Contract

Successful generation commands print only final result URL(s) to stdout:

```text
https://example.com/result.png
```

Progress, task IDs, and diagnostics are written to stderr. Errors are JSON envelopes on stderr with stable `type`, `subtype`, `message`, and `hint` fields.

## Dry Run

Use `--dry-run` to inspect the request payload without making a network call:

```bash
agnes --dry-run image generate \
  --prompt "A futuristic city" \
  --size 1024x768
```

## Schema

Expose command schemas for agents:

```bash
agnes schema image.generate
agnes schema video.generate
```

## Common Commands

```bash
agnes key set <api-key>
agnes key status
agnes key clear

agnes image generate --prompt "..." --size 1024x768
agnes video generate --prompt "..." --num-frames 121 --frame-rate 24
agnes video status <video_id_or_task_id>
agnes doctor
```

## npm Package Contents

The npm package is `@agnes-ai/agnes-cli`. It contains the README, minimal installer scripts, and prebuilt release archives under `npm-bundles/`; the Go source code is open-sourced in this GitHub repository, not shipped in the npm package.

## Release

Releases are tag-driven. Pushing a tag like `v0.1.0` triggers GitHub Actions to:

1. Validate version metadata and run tests
2. Build macOS, Linux, and Windows binaries with GoReleaser
3. Upload GitHub Release assets and `SHA256SUMS`
4. Bundle release assets into the npm package
5. Publish to npm through npm Trusted Publishing

```bash
npm version patch --no-git-tag-version
git add package.json package-lock.json internal/buildinfo/buildinfo.go
git commit -m "Release v0.1.1"
git tag v0.1.1
git push origin main --tags
```
