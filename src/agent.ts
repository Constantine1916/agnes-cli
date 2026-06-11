import type { RuntimeConfig } from "./config.js";

export type AgnesAgentRequest = {
  prompt: string;
  images: string[];
  files: string[];
  config: RuntimeConfig;
};

export type AgnesAgentResponse = {
  text: string;
};

export async function runAgnesAgent(request: AgnesAgentRequest): Promise<AgnesAgentResponse> {
  const lines = [
    "Agnes CLI is initialized.",
    "The real Agnes multimodal agent adapter is the next integration step.",
    "",
    `Prompt: ${request.prompt}`,
    `Images: ${request.images.length}`,
    `Files: ${request.files.length}`,
    `Model: ${request.config.model ?? "not set"}`,
    `API base URL: ${request.config.apiBaseUrl ?? "default"}`,
  ];

  return {
    text: lines.join("\n"),
  };
}
