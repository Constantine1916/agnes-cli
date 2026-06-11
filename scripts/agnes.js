#!/usr/bin/env node

const fs = require("fs");
const path = require("path");
const { spawnSync } = require("child_process");

const root = path.resolve(__dirname, "..");
const vendorBinary = path.join(root, "vendor", process.platform === "win32" ? "agnes.exe" : "agnes");
const localBinary = path.join(root, "bin", process.platform === "win32" ? "agnes.exe" : "agnes");

const binary = fs.existsSync(vendorBinary)
  ? vendorBinary
  : fs.existsSync(localBinary)
    ? localBinary
    : null;

if (!binary) {
  console.error("agnes binary not found. Build it with: go build -o bin/agnes .");
  process.exit(1);
}

const result = spawnSync(binary, process.argv.slice(2), { stdio: "inherit" });
if (result.error) {
  console.error(result.error.message);
  process.exit(1);
}
process.exit(result.status ?? 0);
