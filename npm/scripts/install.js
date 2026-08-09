#!/usr/bin/env node
'use strict';

/**
 * Downloads the PortMaster binary for the current platform from the GitHub
 * release matching this package's version, verifies it against
 * SHA256SUMS.txt, and places it in npm/vendor/.
 *
 * Set PORTMASTER_BINARY_MIRROR to use a mirror of the GitHub releases
 * (useful in restricted networks).
 */

const { createHash } = require('node:crypto');
const { chmodSync, mkdirSync, writeFileSync } = require('node:fs');
const path = require('node:path');
const os = require('node:os');

const pkg = require('../package.json');

const GOOS = { win32: 'windows', darwin: 'darwin', linux: 'linux' }[process.platform];
const GOARCH = { x64: 'amd64', arm64: 'arm64' }[os.arch()];
const EXT = process.platform === 'win32' ? '.exe' : '';

if (!GOOS || !GOARCH) {
  console.error(
    `portmaster-cli: platform ${process.platform}-${os.arch()} is not supported. ` +
      'Supported: Windows/macOS/Linux on amd64 and arm64.'
  );
  process.exit(1);
}

const asset = `portmaster-${GOOS}-${GOARCH}${EXT}`;
const base =
  process.env.PORTMASTER_BINARY_MIRROR ||
  `https://github.com/RichardFlp/portmaster/releases/download/v${pkg.version}`;

const target = path.join(__dirname, '..', 'vendor', `portmaster${EXT}`);

async function download(url) {
  const res = await fetch(url, { redirect: 'follow' });
  if (!res.ok) {
    throw new Error(`GET ${url} -> ${res.status} ${res.statusText}`);
  }
  return Buffer.from(await res.arrayBuffer());
}

async function main() {
  mkdirSync(path.dirname(target), { recursive: true });

  const [binary, sumsText] = await Promise.all([
    download(`${base}/${asset}`),
    download(`${base}/SHA256SUMS.txt`),
  ]);

  const sums = new Map(
    sumsText
      .toString('utf8')
      .split(/\r?\n/)
      .filter(Boolean)
      .map((line) => {
        const [hash, name] = line.trim().split(/\s+/);
        return [name, hash.toLowerCase()];
      })
  );

  const expected = sums.get(asset);
  if (!expected) {
    throw new Error(`no checksum for ${asset} found in SHA256SUMS.txt`);
  }
  const actual = createHash('sha256').update(binary).digest('hex');
  if (actual !== expected) {
    throw new Error(
      `checksum mismatch for ${asset}\n  expected: ${expected}\n  actual:   ${actual}`
    );
  }

  writeFileSync(target, binary);
  if (EXT !== '.exe') chmodSync(target, 0o755);

  console.log(`portmaster-cli: installed ${asset} (v${pkg.version})`);
}

main().catch((err) => {
  console.error(`portmaster-cli: install failed: ${err.message}`);
  process.exit(1);
});
