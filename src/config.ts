import process from "node:process";

export class CliInputError extends Error {
  constructor(message: string) {
    super(message);
    this.name = "CliInputError";
  }
}

export type RuntimeConfig = {
  apiKey: string;
  apiBaseUrl?: string;
  model?: string;
};

export type RuntimeConfigOptions = {
  apiKey?: string;
  apiBaseUrl?: string;
  model?: string;
};

export function resolveRuntimeConfig(
  options: RuntimeConfigOptions,
  env: NodeJS.ProcessEnv = process.env,
): RuntimeConfig {
  const apiKey = firstNonEmpty(options.apiKey, env.AGNES_API_KEY, env.OPENAI_API_KEY);

  if (!apiKey) {
    throw new CliInputError(
      "Missing API key. Pass --api-key, set AGNES_API_KEY, or set OPENAI_API_KEY.",
    );
  }

  return {
    apiKey,
    apiBaseUrl: firstNonEmpty(options.apiBaseUrl, env.AGNES_API_BASE_URL),
    model: firstNonEmpty(options.model, env.AGNES_MODEL),
  };
}

export function inspectRuntimeConfig(
  options: RuntimeConfigOptions,
  env: NodeJS.ProcessEnv = process.env,
) {
  return {
    hasApiKey: Boolean(firstNonEmpty(options.apiKey, env.AGNES_API_KEY, env.OPENAI_API_KEY)),
    apiBaseUrl: firstNonEmpty(options.apiBaseUrl, env.AGNES_API_BASE_URL),
    model: firstNonEmpty(options.model, env.AGNES_MODEL),
  };
}

function firstNonEmpty(...values: Array<string | undefined>) {
  return values.find((value) => value !== undefined && value.trim().length > 0);
}
