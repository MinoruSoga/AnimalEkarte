#!/usr/bin/env node
// PostToolUse Bash: Log PR URL after gh pr create
'use strict';

const MAX_STDIN = 1024 * 1024;
let data = '';
process.stdin.setEncoding('utf8');
process.stdin.on('data', chunk => {
  if (data.length < MAX_STDIN) data += chunk.substring(0, MAX_STDIN - data.length);
});
process.stdin.on('end', () => {
  try {
    const input = JSON.parse(data || '{}');
    const cmd = String(input.tool_input?.command || '');
    const output = String(input.tool_response?.output || input.output || '');

    if (/\bgh\s+pr\s+create\b/.test(cmd)) {
      const m = output.match(/https:\/\/github\.com\/[^\s]+\/pull\/\d+/);
      if (m) {
        process.stderr.write(`[Hook] PR created: ${m[0]}\n`);
        process.stderr.write(`[Hook] To review: gh pr view ${m[0].match(/\/pull\/(\d+)/)?.[1]}\n`);
      }
    }
  } catch {}
  process.stdout.write(data);
  process.exit(0);
});
