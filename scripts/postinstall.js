#!/usr/bin/env node

const crypto = require("crypto");
const fs = require("fs");
const os = require("os");
const path = require("path");

const AdmZip = require("adm-zip");
const tar = require("tar");

const rootDir = path.resolve(__dirname, "..");
const pkg = require(path.join(rootDir, "package.json"));
const bundledAssetsDir = path.join(rootDir, pkg.agnes?.bundledAssetsDir || "npm-bundles");
const vendorDir = path.join(rootDir, "vendor");
const tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), "agnes-npm-"));

const SUPPORTED_TARGETS = {
  "darwin-arm64": { os: "darwin", arch: "arm64", ext: "tar.gz", bin: "agnes" },
  "darwin-x64": { os: "darwin", arch: "amd64", ext: "tar.gz", bin: "agnes" },
  "linux-arm64": { os: "linux", arch: "arm64", ext: "tar.gz", bin: "agnes" },
  "linux-x64": { os: "linux", arch: "amd64", ext: "tar.gz", bin: "agnes" },
  "win32-x64": { os: "windows", arch: "amd64", ext: "zip", bin: "agnes.exe" }
};

async function main() {
  try {
    if (process.env.AGNES_SKIP_POSTINSTALL === "1") {
      log("skip postinstall because AGNES_SKIP_POSTINSTALL=1");
      return;
    }

    const target = resolveTarget();
    const version = pkg.version;
    const projectName = pkg.agnes?.projectName || "agnes";
    const assetName = `${projectName}_${version}_${target.os}_${target.arch}.${target.ext}`;
    const archivePath = path.join(tmpDir, assetName);
    const extractDir = path.join(tmpDir, "extract");
    const sources = resolveReleaseSources();

    fs.rmSync(vendorDir, { recursive: true, force: true });
    fs.mkdirSync(vendorDir, { recursive: true });
    fs.mkdirSync(extractDir, { recursive: true });

    const expectedSha = await resolveExpectedSha(sources, assetName);
    await materializeArchive(sources, assetName, archivePath, expectedSha);
    await extractArchive(archivePath, extractDir, target.ext);

    const extractedBinary = findFileRecursive(extractDir, target.bin);
    if (!extractedBinary) {
      throw new Error(`failed to locate ${target.bin} in extracted archive`);
    }

    const finalBinary = path.join(vendorDir, target.bin);
    fs.copyFileSync(extractedBinary, finalBinary);
    if (process.platform !== "win32") {
      fs.chmodSync(finalBinary, 0o755);
    }
    log(`installed ${target.bin}`);
  } finally {
    fs.rmSync(tmpDir, { recursive: true, force: true });
  }
}

function resolveTarget() {
  const key = `${process.platform}-${process.arch}`;
  const target = SUPPORTED_TARGETS[key];
  if (!target) {
    throw new Error(`unsupported platform: ${process.platform}/${process.arch}`);
  }
  return target;
}

function resolveReleaseSources() {
  const sources = [];
  if (fs.existsSync(bundledAssetsDir)) {
    sources.push({ name: "bundled npm assets", type: "local", basePath: bundledAssetsDir });
  }
  sources.push({ name: "GitHub Release", type: "remote", baseUrl: resolveReleaseBaseUrl() });
  return sources;
}

function resolveReleaseBaseUrl() {
  if (process.env.AGNES_RELEASE_BASE_URL) {
    return stripTrailingSlash(process.env.AGNES_RELEASE_BASE_URL);
  }
  const template = pkg.agnes?.releaseBaseUrlTemplate;
  if (!template) {
    throw new Error("missing agnes.releaseBaseUrlTemplate in package.json");
  }
  return stripTrailingSlash(
    template
      .replaceAll("{version}", pkg.version)
      .replaceAll("{tag}", `v${pkg.version}`)
  );
}

async function resolveExpectedSha(sources, assetName) {
  const errors = [];
  for (const source of sources) {
    try {
      const checksumText = source.type === "local"
        ? fs.readFileSync(path.join(source.basePath, "SHA256SUMS"), "utf8")
        : await fetchText(`${source.baseUrl}/SHA256SUMS`);
      const expectedSha = parseChecksum(checksumText, assetName);
      if (!expectedSha) {
        throw new Error(`checksum for ${assetName} not found in SHA256SUMS`);
      }
      log(`using checksums from ${source.name}`);
      return expectedSha;
    } catch (err) {
      errors.push(`${source.name}: ${err.message}`);
    }
  }
  throw new Error(`failed to resolve checksums for ${assetName}\n${errors.join("\n")}`);
}

async function materializeArchive(sources, assetName, archivePath, expectedSha) {
  const errors = [];
  for (const source of sources) {
    try {
      if (source.type === "local") {
        fs.copyFileSync(path.join(source.basePath, assetName), archivePath);
      } else {
        await downloadFile(`${source.baseUrl}/${assetName}`, archivePath);
      }

      const actualSha = sha256File(archivePath);
      if (actualSha !== expectedSha) {
        throw new Error(`checksum mismatch for ${assetName}`);
      }
      log(`using ${assetName} from ${source.name}`);
      return;
    } catch (err) {
      fs.rmSync(archivePath, { force: true });
      errors.push(`${source.name}: ${err.message}`);
    }
  }
  throw new Error(`failed to obtain ${assetName}\n${errors.join("\n")}`);
}

function stripTrailingSlash(value) {
  return value.replace(/\/+$/, "");
}

async function fetchText(url) {
  const response = await fetch(url);
  if (!response.ok) {
    throw new Error(`failed to download ${url}: HTTP ${response.status}`);
  }
  return response.text();
}

async function downloadFile(url, destination) {
  const response = await fetch(url);
  if (!response.ok) {
    throw new Error(`failed to download ${url}: HTTP ${response.status}`);
  }
  const buffer = Buffer.from(await response.arrayBuffer());
  fs.writeFileSync(destination, buffer);
}

function parseChecksum(text, assetName) {
  for (const line of text.split(/\r?\n/)) {
    const match = line.trim().match(/^([a-fA-F0-9]{64})\s+\*?(.+)$/);
    if (match && path.basename(match[2]) === assetName) {
      return match[1].toLowerCase();
    }
  }
  return null;
}

function sha256File(filePath) {
  const hash = crypto.createHash("sha256");
  hash.update(fs.readFileSync(filePath));
  return hash.digest("hex");
}

async function extractArchive(archivePath, extractDir, ext) {
  if (ext === "zip") {
    const zip = new AdmZip(archivePath);
    zip.extractAllTo(extractDir, true);
    return;
  }
  await tar.x({ file: archivePath, cwd: extractDir });
}

function findFileRecursive(dir, filename) {
  for (const entry of fs.readdirSync(dir, { withFileTypes: true })) {
    const fullPath = path.join(dir, entry.name);
    if (entry.isDirectory()) {
      const found = findFileRecursive(fullPath, filename);
      if (found) return found;
    } else if (entry.name === filename) {
      return fullPath;
    }
  }
  return null;
}

function log(message) {
  console.error(`[agnes-cli] ${message}`);
}

main().catch((err) => {
  console.error(`[agnes-cli] ${err.message}`);
  process.exit(1);
});
