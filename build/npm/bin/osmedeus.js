#!/usr/bin/env node
// Unified entry point for the Osmedeus CLI distributed via npm.
//
// The platform-specific binary is shipped *gzipped* inside an optional
// dependency package (one of @j3ssie/osmedeus-<tag>). On first run we
// decompress it once into a version-scoped cache directory and then exec it,
// forwarding args, stdio, signals and the exit status.

import { spawn } from "node:child_process";
import { createReadStream, createWriteStream, existsSync, statSync } from "node:fs";
import { mkdir, rename, chmod, unlink } from "node:fs/promises";
import { createRequire } from "node:module";
import { pipeline } from "node:stream/promises";
import { createGunzip } from "node:zlib";
import os from "node:os";
import path from "node:path";
import { fileURLToPath } from "node:url";

// __dirname / require equivalents in ESM.
const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const require = createRequire(import.meta.url);

// Local alias (npm install path) -> nothing else; the underlying package
// published to npm is always @j3ssie/osmedeus with a version suffix.
const PLATFORM_PACKAGE_BY_TAG = {
  "linux-x64": "@j3ssie/osmedeus-linux-x64",
  "linux-arm64": "@j3ssie/osmedeus-linux-arm64",
  "darwin-x64": "@j3ssie/osmedeus-darwin-x64",
  "darwin-arm64": "@j3ssie/osmedeus-darwin-arm64",
};

const { platform, arch } = process;

function detectPackageManager() {
  const userAgent = process.env.npm_config_user_agent || "";
  if (/\bbun\//.test(userAgent)) return "bun";
  const execPath = process.env.npm_execpath || "";
  if (execPath.includes("bun")) return "bun";
  return userAgent ? "npm" : null;
}

function reinstallHint() {
  return detectPackageManager() === "bun"
    ? "bun install -g @j3ssie/osmedeus@latest"
    : "npm install -g @j3ssie/osmedeus@latest";
}

let platformTag = null;
switch (platform) {
  case "linux":
  case "android":
    if (arch === "x64") platformTag = "linux-x64";
    else if (arch === "arm64") platformTag = "linux-arm64";
    break;
  case "darwin":
    if (arch === "x64") platformTag = "darwin-x64";
    else if (arch === "arm64") platformTag = "darwin-arm64";
    break;
  default:
    break;
}

if (!platformTag) {
  throw new Error(
    `Unsupported platform: ${platform} (${arch}). ` +
      `Osmedeus npm builds cover linux/darwin on x64/arm64. ` +
      `See https://docs.osmedeus.org for other install options.`,
  );
}

const platformPackage = PLATFORM_PACKAGE_BY_TAG[platformTag];

// Resolve the gzipped binary: prefer the installed optional-dependency
// package, fall back to a local vendor/ tree (used by `npm pack` testing and
// in-repo dev before publishing).
const gzName = "osmedeus.gz";
const localVendorRoot = path.join(__dirname, "..", "vendor");
const localGzPath = path.join(localVendorRoot, platformTag, gzName);

let gzPath = null;
try {
  const pkgJsonPath = require.resolve(`${platformPackage}/package.json`);
  gzPath = path.join(path.dirname(pkgJsonPath), "vendor", platformTag, gzName);
} catch {
  if (existsSync(localGzPath)) {
    gzPath = localGzPath;
  }
}

if (!gzPath || !existsSync(gzPath)) {
  throw new Error(
    `Missing platform package ${platformPackage} (the osmedeus binary for ` +
      `${platformTag} was not installed). This usually means npm skipped ` +
      `optional dependencies. Reinstall Osmedeus:\n    ${reinstallHint()}`,
  );
}

// Version-scoped cache path so upgrades never exec a stale binary, and so the
// extraction directory is always writable by the running user even when the
// package itself was installed into a root-owned global prefix.
const pkgVersion = require("../package.json").version;
const osmedeusHome =
  process.env.OSMEDEUS_NPM_HOME || path.join(os.homedir(), ".osmedeus");
const binaryDir = path.join(osmedeusHome, "npm-bin", pkgVersion, platformTag);
const binaryPath = path.join(binaryDir, "osmedeus");

async function ensureBinary() {
  if (existsSync(binaryPath) && statSync(binaryPath).size > 0) {
    return;
  }

  await mkdir(binaryDir, { recursive: true });

  // Decompress to a unique temp file, then atomically rename into place so
  // concurrent first-runs and interrupted extractions can never leave a
  // truncated binary at the final path.
  const tmpPath = path.join(
    binaryDir,
    `osmedeus.${process.pid}.${Date.now()}.tmp`,
  );

  try {
    await pipeline(
      createReadStream(gzPath),
      createGunzip(),
      createWriteStream(tmpPath, { mode: 0o755 }),
    );
    await chmod(tmpPath, 0o755);
    await rename(tmpPath, binaryPath);
  } catch (err) {
    await unlink(tmpPath).catch(() => {});
    // Another racing process may have completed the extraction already.
    if (existsSync(binaryPath) && statSync(binaryPath).size > 0) {
      return;
    }
    throw new Error(
      `Failed to unpack the osmedeus binary for ${platformTag}: ${err.message}\n` +
        `Try reinstalling:\n    ${reinstallHint()}`,
    );
  }
}

await ensureBinary();

// Use an asynchronous spawn (not spawnSync) so Node can respond to signals
// (e.g. Ctrl-C / SIGINT) while the native binary runs, forward them to the
// child, and mirror the child's termination reason in the parent.
const child = spawn(binaryPath, process.argv.slice(2), {
  stdio: "inherit",
  env: process.env,
});

child.on("error", (err) => {
  // eslint-disable-next-line no-console
  console.error(err);
  process.exit(1);
});

const forwardSignal = (signal) => {
  if (child.killed) return;
  try {
    child.kill(signal);
  } catch {
    /* ignore */
  }
};

["SIGINT", "SIGTERM", "SIGHUP"].forEach((sig) => {
  process.on(sig, () => forwardSignal(sig));
});

const childResult = await new Promise((resolve) => {
  child.on("exit", (code, signal) => {
    if (signal) {
      resolve({ type: "signal", signal });
    } else {
      resolve({ type: "code", exitCode: code ?? 1 });
    }
  });
});

if (childResult.type === "signal") {
  // Re-emit the same signal so the parent terminates with 128 + n semantics.
  process.kill(process.pid, childResult.signal);
} else {
  process.exit(childResult.exitCode);
}
