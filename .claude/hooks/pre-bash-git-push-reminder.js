#!/usr/bin/env node
/**
 * PreToolUse Hook: Remind to review changes before git push
 *
 * Warns (via stderr) when a git push command is detected.
 * Does NOT block — just reminds.
 */
'use strict';

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
    const command = String(input.tool_input?.command || '');

    if (/\bgit\s+push\b/.test(command)) {
      process.stderr.write('[Hook] git push detected. Ensure all changes are reviewed and tests pass before pushing.\n');
    }
  } catch {
    // Parse error — pass through
  }

  process.stdout.write(data);
  process.exit(0);
});
