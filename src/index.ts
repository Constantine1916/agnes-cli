#!/usr/bin/env node
import { createRequire } from "node:module";
import process from "node:process";
import { Command } from "commander";

import { runAgnesAgent } from "./agent.js";
import { CliInputError, inspectRuntimeConfig, resolveRuntimeConfig } from "./config.js";

const require = createRequire(import.meta.url);
const packageJson = require("../package.json") as { version: string };

type SharedOptions = {
  apiKey?: string;
  apiBaseUrl?: string;
  model?: string;
};

type AskOptions = SharedOptions & {
  image?: string[];
  file?: string[];
};

const program = new Command()
  .name("agnes")
  .description("CLI for Agnes multimodal agent workflows.")
  .version(packageJson.version);

program
  .command("doctor")
  .description("Check local Agnes CLI configuration.")
  .option("-k, --api-key <key>", "API key for the Agnes agent runtime")
  .option("--api-base-url <url>", "optional Agnes agent API base URL")
  .option("-m, --model <model>", "optional model or capability profile")
  .action((options: SharedOptions) => {
    const config = inspectRuntimeConfig(options);

    console.log("Agnes CLI");
    console.log(`API key: ${config.hasApiKey ? "configured" : "missing"}`);
    console.log(`API base URL: ${config.apiBaseUrl ?? "default"}`);
    console.log(`Model: ${config.model ?? "not set"}`);
  });

program
  .command("ask")
  .description("Send a prompt and optional multimodal inputs to Agnes.")
  .argument("<prompt>", "prompt to send to Agnes")
  .option("-k, --api-key <key>", "API key for the Agnes agent runtime")
  .option("--api-base-url <url>", "optional Agnes agent API base URL")
  .option("-m, --model <model>", "optional model or capability profile")
  .option("-i, --image <path...>", "image path(s) to include")
  .option("-f, --file <path...>", "file path(s) to include")
  .action(async (prompt: string, options: AskOptions) => {
    try {
      const config = resolveRuntimeConfig(options);
      const response = await runAgnesAgent({
        prompt,
        images: options.image ?? [],
        files: options.file ?? [],
        config,
      });

      console.log(response.text);
    } catch (error) {
      handleError(error);
    }
  });

program.parseAsync(process.argv).catch(handleError);

function handleError(error: unknown) {
  if (error instanceof CliInputError) {
    console.error(error.message);
    process.exitCode = 1;
    return;
  }

  if (error instanceof Error) {
    console.error(error.message);
    process.exitCode = 1;
    return;
  }

  console.error("Unknown error");
  process.exitCode = 1;
}
