# agnes-cli

Command-line interface for Agnes multimodal agent workflows.

This repository is the CLI home for Agnes. The intended auth model is simple:
users do not need to log in. They provide an API key with `--api-key` or
`AGNES_API_KEY`, and the CLI passes requests to the Agnes agent runtime.

## Status

Initial repository scaffold:

- TypeScript CLI package
- `agnes` and `agnes-cli` binary names
- API key based runtime configuration
- Placeholder adapter for the upcoming Agnes multimodal agent integration

## Development

```bash
npm install
npm run build
npm run dev -- doctor
npm run dev -- ask "summarize this image" --api-key "$AGNES_API_KEY" --image ./example.png
```

## Configuration

The CLI reads configuration in this order:

1. Command-line flags
2. Environment variables

Supported values:

- `--api-key` or `AGNES_API_KEY`
- `OPENAI_API_KEY` as a compatibility fallback
- `--api-base-url` or `AGNES_API_BASE_URL`
- `--model` or `AGNES_MODEL`

## Next step

Wire `src/agent.ts` to the real Agnes multimodal agent runtime so prompts,
images, and files can be sent through the same terminal command.
