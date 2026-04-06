#!/usr/bin/env node
/**
 * PostToolUse Hook: Warn when a file exceeds 500 lines after edit
 *
 * Complements pre-write-large-file-block.js (which only fires on Write).
 * This hook catches gradual file growth through successive Edit operations.
 *
 * Thresholds:
 *   > 500 lines: warning (soft limit from project rules)
 *   > 800 lines: strong warning (hard limit)
 *
 * Does NOT block (exit 0) — the edit already happened.
 */
'use strict';

const fs = require('fs');

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

    if (!filePath) {
      process.stdout.write(data);
      process.exit(0);
    }

    // Skip generated files, migrations, vendor, node_modules, lock files
    if (/generated|migrations|vendor|node_modules|\.lock$|lock\.json$/.test(filePath)) {
      process.stdout.write(data);
      process.exit(0);
    }

    let content;
    try {
      content = fs.readFileSync(filePath, 'utf8');
    } catch {
      process.stdout.write(data);
      process.exit(0);
    }

    const lineCount = content.split('\n').length;

    if (lineCount > 800) {
      process.stderr.write(
        `\n[Hook] ⚠ FILE TOO LARGE: ${filePath} is ${lineCount} lines (hard limit: 800)\n` +
        '[Hook] Split this file into smaller, focused modules immediately.\n\n'
      );
    } else if (lineCount > 500) {
      process.stderr.write(
        `[Hook] WARNING: ${filePath} is ${lineCount} lines (soft limit: 500). Consider splitting.\n`
      );
    }
  } catch {
    // Parse error — pass through
  }

  process.stdout.write(data);
  process.exit(0);
});
