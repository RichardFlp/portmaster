#!/usr/bin/env node
'use strict';

const { spawn } = require('node:child_process');
const { existsSync } = require('node:fs');
const path = require('node:path');

const ext = process.platform === 'win32' ? '.exe' : '';
const binary = path.join(__dirname, '..', 'vendor', `portmaster${ext}`);

if (!existsSync(binary)) {
  console.error(
    'portmaster: binary is missing. Re-run `npm install -g @richardflp/portmaster`' +
      ' (or `npm rebuild -g @richardflp/portmaster`) to download it.'
  );
  process.exit(1);
}

const child = spawn(binary, process.argv.slice(2), { stdio: 'inherit' });

child.on('error', (err) => {
  console.error(`portmaster: failed to launch binary: ${err.message}`);
  process.exit(1);
});

child.on('exit', (code, signal) => {
  if (signal) {
    process.kill(process.pid, signal);
    return;
  }
  process.exit(code ?? 0);
});
