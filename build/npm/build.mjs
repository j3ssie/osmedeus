#!/usr/bin/env node
// Stage the @j3ssie/osmedeus npm packages.
//
// Produces, under build/dist-npm/:
//   - osmedeus/              the thin launcher package (@j3ssie/osmedeus@<v>)
//   - osmedeus-<tag>/        4 platform packages (@j3ssie/osmedeus@<v>-<tag>)
//                            each carrying the gzipped binary in vendor/<tag>/
//
// One npm name, version-suffixed platform builds (codex-style): the launcher
// package pulls its platform build in as an aliased optionalDependency, so
// `npm i -g @j3ssie/osmedeus` downloads exactly one ~100MB binary.
//
// The binary is gzipped so each platform package's *unpacked* size stays well
// under npm's per-version ceiling instead of shipping the raw ~350MB Go binary.
// bin/osmedeus.js decompresses it once on first run.
//
// Source binaries come from build/dist-npm-bin/osmedeus_<goos>_<goarch>/osmedeus
// (`make npm-binaries`). Goreleaser output (dist/) works too — the layout and
// the version check both understand it.
//
// Usage:
//   node build/npm/build.mjs [--pack] [--allow-missing=<tag,tag>] [--dist-dir=<path>]
//   OSMEDEUS_VERSION=5.0.3 node build/npm/build.mjs

import { spawnSync } from "node:child_process";
import {
  createWriteStream,
  existsSync,
  readdirSync,
  readFileSync,
  rmSync,
  mkdirSync,
  copyFileSync,
  statSync,
  writeFileSync,
} from "node:fs";
import { Readable } from "node:stream";
import { pipeline } from "node:stream/promises";
import { createGzip } from "node:zlib";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const REPO_ROOT = path.resolve(__dirname, "..", "..");
const OUT_DIR = path.join(REPO_ROOT, "build", "dist-npm");
const LAUNCHER_SRC = path.join(__dirname, "bin", "osmedeus.js");
// Bundle the full project README onto the npm package page.
const README_SRC = path.join(REPO_ROOT, "README.md");
const LICENSE_SRC = path.join(REPO_ROOT, "LICENSE");
const VERSION_GO = path.join(REPO_ROOT, "internal", "core", "constants.go");
// Written by `make npm-binaries`; records the version the binaries were built
// for. See verifyBuiltVersion().
const STAMP_FILE = ".build-version";

const NPM_NAME = "@j3ssie/osmedeus";
const BIN_NAME = "osmedeus";
const LICENSE_ID = "MIT";
const HOMEPAGE = "https://www.osmedeus.org";
const DESCRIPTION =
  "Osmedeus - A Modern Orchestration Engine for Security";
const KEYWORDS = [
  "osmedeus",
  "security",
  "recon",
  "reconnaissance",
  "bug-bounty",
  "pentest",
  "workflow",
  "automation",
  "orchestration",
];
const REPOSITORY = {
  type: "git",
  url: "git+https://github.com/j3ssie/osmedeus.git",
};
const ENGINES = { node: ">=16" };

const PLATFORMS = [
  { tag: "linux-x64", goos: "linux", goarch: "amd64", os: "linux", cpu: "x64" },
  { tag: "linux-arm64", goos: "linux", goarch: "arm64", os: "linux", cpu: "arm64" },
  { tag: "darwin-x64", goos: "darwin", goarch: "amd64", os: "darwin", cpu: "x64" },
  { tag: "darwin-arm64", goos: "darwin", goarch: "arm64", os: "darwin", cpu: "arm64" },
];

const args = process.argv.slice(2);
const doPack = args.includes("--pack");
const allowMissing = new Set(
  (args.find((a) => a.startsWith("--allow-missing=")) || "")
    .split("=")[1]
    ?.split(",")
    .map((s) => s.trim())
    .filter(Boolean) || [],
);
const distDirArg = (args.find((a) => a.startsWith("--dist-dir=")) || "").split("=")[1];
const DIST_DIR = path.resolve(
  REPO_ROOT,
  distDirArg || process.env.OSMEDEUS_NPM_DIST_DIR || path.join("build", "dist-npm-bin"),
);

function fail(msg) {
  console.error(`\x1b[31m[!] ${msg}\x1b[0m`);
  process.exit(1);
}

function info(msg) {
  console.log(`\x1b[36m[*]\x1b[0m ${msg}`);
}

// --- version --------------------------------------------------------------

function deriveBaseVersion() {
  if (process.env.OSMEDEUS_VERSION) {
    return process.env.OSMEDEUS_VERSION.replace(/^v/, "");
  }
  const src = readFileSync(VERSION_GO, "utf8");
  const m = src.match(/^\s*VERSION\s*=\s*"v?([^"]+)"/m);
  if (!m) fail(`Could not parse VERSION from ${VERSION_GO}`);
  return m[1];
}

const baseVersion = deriveBaseVersion();
if (!/^\d+\.\d+\.\d+(?:-[0-9A-Za-z.-]+)?(?:\+[0-9A-Za-z.-]+)?$/.test(baseVersion)) {
  console.warn(
    `\x1b[33m[warn] "${baseVersion}" does not look like a semver string; npm may reject it.\x1b[0m`,
  );
}
const platformVersion = (tag) => `${baseVersion}-${tag}`;

// --- built-version verification -------------------------------------------

// npm versions are immutable, so a stale binary published under a fresh version
// number can never be corrected — only superseded. Refuse to pack unless the
// binaries in DIST_DIR are provably built for the version being published.
//
// Two authoritative sources, in order:
//   1. the .build-version stamp `make npm-binaries` writes next to the binaries
//   2. goreleaser's `osmedeus_<version>_<os>_<arch>.tar.gz` archive names
//
// Do NOT try to read the version out of the binary: a Go binary carries version
// strings from embedded content (docs, presets, UI), so a substring match
// false-positives, and executing it triggers first-run initialization.
function verifyBuiltVersion(expected) {
  if (!existsSync(DIST_DIR)) {
    fail(
      `${DIST_DIR} not found. Build the cross-platform binaries first:\n` +
        `    make npm-binaries`,
    );
  }

  const stampPath = path.join(DIST_DIR, STAMP_FILE);
  if (existsSync(stampPath)) {
    const stamped = readFileSync(stampPath, "utf8").trim().replace(/^v/, "");
    if (stamped !== expected) {
      fail(
        `built-version mismatch: ${DIST_DIR} was built for ${stamped} but this ` +
          `npm publish is version ${expected}. Those binaries are STALE — publishing ` +
          `them would ship a binary whose \`osmedeus version\` reports the wrong number. ` +
          `Run \`make npm-binaries\` to rebuild for ${expected}.`,
      );
    }
    info(`verified ${path.relative(REPO_ROOT, DIST_DIR)}/ binaries were built for ${expected}`);
    return;
  }

  // goreleaser layout: verify against the archive names it wrote.
  const archiveRe = new RegExp(
    `^${BIN_NAME}_(\\d+\\.\\d+\\.\\d+(?:-[0-9A-Za-z.-]+)?)_(?:linux|darwin|windows)_(?:amd64|arm64)\\.tar\\.gz$`,
  );
  const built = new Set();
  for (const f of readdirSync(DIST_DIR)) {
    const m = f.match(archiveRe);
    if (m) built.add(m[1]);
  }
  if (built.size === 0) {
    // Fail closed: warning-and-continuing here is exactly the fail-open that
    // lets stale binaries be repackaged under a new version.
    fail(
      `cannot verify the built version against ${expected}: ${DIST_DIR} has no ` +
        `${STAMP_FILE} stamp and no goreleaser archives to check against. The ` +
        `binaries there are unverifiable and may be STALE. Run \`make npm-binaries\`.`,
    );
  }
  if (built.size !== 1 || !built.has(expected)) {
    fail(
      `built-version mismatch: ${DIST_DIR} was built for [${[...built].sort().join(", ")}] ` +
        `but this npm publish is version ${expected}. Run \`make npm-binaries\` to ` +
        `rebuild for ${expected}.`,
    );
  }
  info(`verified ${path.relative(REPO_ROOT, DIST_DIR)}/ binaries were built for ${expected}`);
}

// --- locate binaries -------------------------------------------------------

// Matches both `make npm-binaries` output (osmedeus_linux_amd64) and
// goreleaser's variant-suffixed dirs (osmedeus_linux_amd64_v1).
function findSourceBinary(goos, goarch) {
  const re = new RegExp(`^${BIN_NAME}_${goos}_${goarch}(?:_.+)?$`);
  const dirs = readdirSync(DIST_DIR)
    .filter((d) => re.test(d))
    .filter((d) => existsSync(path.join(DIST_DIR, d, BIN_NAME)))
    .sort();
  if (dirs.length === 0) return null;
  if (dirs.length > 1) {
    console.warn(
      `\x1b[33m[warn] multiple build dirs for ${goos}/${goarch}: ` +
        `${dirs.join(", ")} — using ${dirs[0]}\x1b[0m`,
    );
  }
  return path.join(DIST_DIR, dirs[0], BIN_NAME);
}

// --- staging --------------------------------------------------------------

function writeJson(file, obj) {
  mkdirSync(path.dirname(file), { recursive: true });
  writeFileSync(file, JSON.stringify(obj, null, 2) + "\n");
}

function humanSize(bytes) {
  return `${(bytes / 1048576).toFixed(0)} MB`;
}

async function gzipBuffer(buf, dest) {
  mkdirSync(path.dirname(dest), { recursive: true });
  await pipeline(
    Readable.from([buf]),
    createGzip({ level: 9 }),
    createWriteStream(dest),
  );
}

async function stagePlatformPackage(p) {
  const src = findSourceBinary(p.goos, p.goarch);
  if (!src) {
    if (allowMissing.has(p.tag)) {
      console.warn(
        `\x1b[33m[warn] skipping ${p.tag}: no binary for ${p.goos}/${p.goarch}\x1b[0m`,
      );
      return false;
    }
    fail(
      `Missing binary for ${p.goos}/${p.goarch} (tag ${p.tag}) in ${DIST_DIR}. ` +
        `Run \`make npm-binaries\` or pass --allow-missing=${p.tag}.`,
    );
  }

  const binBuf = readFileSync(src);
  const pkgDir = path.join(OUT_DIR, `${BIN_NAME}-${p.tag}`);
  const gzPath = path.join(pkgDir, "vendor", p.tag, `${BIN_NAME}.gz`);

  info(`packaging ${p.tag} (${humanSize(binBuf.length)} -> gzip)`);
  await gzipBuffer(binBuf, gzPath);

  writeJson(path.join(pkgDir, "package.json"), {
    name: NPM_NAME,
    version: platformVersion(p.tag),
    description: `${DESCRIPTION} (${p.tag} prebuilt binary)`,
    license: LICENSE_ID,
    homepage: HOMEPAGE,
    repository: REPOSITORY,
    os: [p.os],
    cpu: [p.cpu],
    engines: ENGINES,
    files: ["vendor"],
  });
  if (existsSync(README_SRC)) copyFileSync(README_SRC, path.join(pkgDir, "README.md"));
  if (existsSync(LICENSE_SRC)) copyFileSync(LICENSE_SRC, path.join(pkgDir, "LICENSE"));

  console.log(`    -> ${pkgDir}  (gz ${humanSize(statSync(gzPath).size)})`);
  return true;
}

function stageMainPackage(stagedTags) {
  const pkgDir = path.join(OUT_DIR, BIN_NAME);
  const binDir = path.join(pkgDir, "bin");
  mkdirSync(binDir, { recursive: true });
  copyFileSync(LAUNCHER_SRC, path.join(binDir, `${BIN_NAME}.js`));
  if (existsSync(README_SRC)) copyFileSync(README_SRC, path.join(pkgDir, "README.md"));
  if (existsSync(LICENSE_SRC)) copyFileSync(LICENSE_SRC, path.join(pkgDir, "LICENSE"));

  const optionalDependencies = {};
  for (const p of PLATFORMS) {
    if (!stagedTags.has(p.tag)) continue;
    optionalDependencies[`${NPM_NAME}-${p.tag}`] =
      `npm:${NPM_NAME}@${platformVersion(p.tag)}`;
  }

  writeJson(path.join(pkgDir, "package.json"), {
    name: NPM_NAME,
    version: baseVersion,
    description: DESCRIPTION,
    keywords: KEYWORDS,
    license: LICENSE_ID,
    homepage: HOMEPAGE,
    repository: REPOSITORY,
    type: "module",
    bin: { [BIN_NAME]: `bin/${BIN_NAME}.js` },
    engines: ENGINES,
    files: ["bin"],
    optionalDependencies,
  });
  info(`staged main package -> ${pkgDir}`);
}

function npmPack(pkgDir) {
  const res = spawnSync(
    "npm",
    ["pack", "--json", "--pack-destination", OUT_DIR],
    { cwd: pkgDir, encoding: "utf8" },
  );
  if (res.status !== 0) {
    fail(`npm pack failed in ${pkgDir}:\n${res.stderr || res.stdout}`);
  }
  try {
    const out = JSON.parse(res.stdout);
    const name = out[0]?.filename;
    if (name) console.log(`    -> ${path.join(OUT_DIR, name)}`);
  } catch {
    /* non-fatal: tarball is still written */
  }
}

// --- main -----------------------------------------------------------------

info(`osmedeus npm build — version ${baseVersion}`);
verifyBuiltVersion(baseVersion);
if (existsSync(OUT_DIR)) rmSync(OUT_DIR, { recursive: true, force: true });
mkdirSync(OUT_DIR, { recursive: true });

const stagedTags = new Set();
for (const p of PLATFORMS) {
  if (await stagePlatformPackage(p)) stagedTags.add(p.tag);
}
if (stagedTags.size === 0) fail("No platform packages were staged.");
stageMainPackage(stagedTags);

if (doPack) {
  info("running npm pack on each staged package...");
  npmPack(path.join(OUT_DIR, BIN_NAME));
  for (const tag of stagedTags) npmPack(path.join(OUT_DIR, `${BIN_NAME}-${tag}`));
}

info("done.");
console.log(`\nStaged in ${OUT_DIR}:`);
console.log(`  ${NPM_NAME}@${baseVersion}  (main)`);
for (const tag of stagedTags) {
  console.log(`  ${NPM_NAME}@${platformVersion(tag)}  (${tag})`);
}
console.log(
  `\nVerify:  node ${path.join(OUT_DIR, BIN_NAME, "bin", `${BIN_NAME}.js`)} version`,
);
