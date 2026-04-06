#!/usr/bin/env node
/**
 * PreToolUse Hook: Protect linter/formatter/build config files
 *
 * Warns (stderr) when editing config files that control code quality.
 * Steers the agent to fix the code instead of weakening the config.
 *
 * Does NOT block (exit 0) — just warns loudly.
 */
'use strict';

const path = require('path');

const MAX_STDIN = 1024 * 1024;
let data = '';
process.stdin.setEncoding('utf8');

process.stdin.on('data', chunk => {
  if (data.length < MAX_STDIN) {
    data += chunk.substring(0, MAX_STDIN - data.length);
  }
});

process.stdin.on('end', () => {
  try {
    const input = JSON.parse(data);
    const filePath = String(input.tool_input?.file_path || '');
    const basename = path.basename(filePath);

    // Config files that should rarely be modified during normal development
    const PROTECTED_CONFIGS = [
      /^\.eslintrc/,
      /^eslint\.config\./,
      /^tsconfig.*\.json$/,
      /^\.prettierrc/,
      /^prettier\.config\./,
      /^biome\.json$/,
      /^\.golangci\.ya?ml$/,
      /^golangci-lint\.ya?ml$/,
      /^vite\.config\./,
      /^tailwind\.config\./,
    ];

    const isProtected = PROTECTED_CONFIGS.some(p => p.test(basename));

    if (isProtected) {
      process.stderr.write(`\n[Hook] ⚠ CONFIG FILE EDIT: ${filePath}\n`);
      process.stderr.write('[Hook] Are you fixing the code or weakening the config?\n');
      process.stderr.write('[Hook] Prefer fixing code over disabling lint rules.\n\n');
    }
  } catch {
    // Parse error — pass through
  }

  process.stdout.write(data);
  process.exit(0);
});
