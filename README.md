# agnes-cli

Agent-friendly Go CLI for Agnes image and video generation.

The CLI is designed for agents first: stable commands, API-key based auth,
URL-only result output on stdout, and diagnostics/progress on stderr.

## Install From Source

```bash
go build -o bin/agnes .
./bin/agnes doctor --offline
```

## Release

Releases are tag-driven. A tag like `v0.1.0` builds macOS, Linux, and Windows
binaries with GoReleaser, uploads them to GitHub Releases, then publishes the
npm wrapper package with npm Trusted Publishing.

Before the first npm publish, configure npm Trusted Publishing for this package:

1. Open the `agnes-cli` package settings on npmjs.com.
2. Add GitHub Actions as a Trusted Publisher.
3. Use repository `Constantine1916/agnes-cli`.
4. Use workflow filename `release.yml`.
5. Leave environment blank unless this workflow later adds a GitHub Environment.
6. Under allowed actions, select `npm publish`.

No `NPM_TOKEN` GitHub secret is required.

```bash
npm version patch --no-git-tag-version
git add package.json
git commit -m "Release v0.1.1"
git tag v0.1.1
git push origin main --tags
```

After publish, users install with:

```bash
npm install -g agnes-cli
agnes --version
```

## API Key

No login is required. Configure a key with one of:

```bash
agnes key set "$AGNES_API_KEY"
AGNES_API_KEY=... agnes image generate --prompt "A product photo" --size 1024x768
agnes --api-key ... video status video_xxx
```

Key precedence:

1. `--api-key`
2. `AGNES_API_KEY`
3. Saved key from `agnes key set`

## Commands

```bash
agnes key set <api-key>
agnes key status
agnes key clear

agnes image generate \
  --prompt "A clean product photo of a glass cube" \
  --size 1024x768

agnes image generate \
  --prompt "Make it cinematic" \
  --image ./input.png \
  --model agnes-image-2.1-flash

agnes video generate \
  --prompt "A cat walking on the beach at sunset" \
  --num-frames 121 \
  --frame-rate 24

agnes video generate \
  --prompt "Smooth transition between keyframes" \
  --image ./start.png \
  --image ./end.png \
  --mode keyframes

agnes video status <video_id_or_task_id>
agnes schema image.generate
agnes schema video.generate
agnes doctor
```

Use `--dry-run` to inspect a request payload without calling Agnes:

```bash
agnes --dry-run image generate --prompt "A futuristic city" --size 1024x768
```

## Output Contract

Successful generation commands print only result URL(s) to stdout. Progress,
task ids, and diagnostics are written to stderr. Errors are JSON envelopes on
stderr with stable `type`, `subtype`, `message`, and `hint` fields.
