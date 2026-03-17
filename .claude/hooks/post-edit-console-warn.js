#!/usr/bin/env node
/**
 * PostToolUse Hook: Warn about console.log statements after edits
 *
 * Enforces the project rule: "console.log 放置禁止".
 * Warns via stderr — does not block.
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

    if (filePath && /\.(ts|tsx|js|jsx)$/.test(filePath)) {
      let content;
      try {
        content = fs.readFileSync(filePath, 'utf8');
      } catch {
        process.stdout.write(data);
        process.exit(0);
      }

      const lines = content.split('\n');
      const matches = [];

      lines.forEach((line, idx) => {
        if (/console\.log/.test(line)) {
          matches.push(`  ${idx + 1}: ${line.trim()}`);
        }
      });

      if (matches.length > 0) {
        process.stderr.write(`[Hook] WARNING: console.log found in ${filePath}\n`);
        matches.slice(0, 5).forEach(m => process.stderr.write(m + '\n'));
        process.stderr.write('[Hook] Remove console.log before committing.\n');
      }
    }
  } catch {
    // Parse error — pass through
  }

  process.stdout.write(data);
  process.exit(0);
});
