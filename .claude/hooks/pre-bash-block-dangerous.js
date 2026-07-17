#!/usr/bin/env node
/**
 * PreToolUse Hook: Block destructive shell commands
 *
 * Blocks commands that could cause catastrophic damage (rm -rf /, dd, mkfs, fork bomb).
 * Exit code 2 = block the tool call.
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

    const DANGEROUS_PATTERNS = [
      /rm\s+-rf\s+\//,
      /rm\s+-rf\s+~/,
      />\s*\/dev\/sda/,
      /\bmkfs\b/,
      /\bdd\s+if=\/dev\/zero/,
      /:\(\)\{.*\|.*&\}\;:/,  // fork bomb
      /\|\s*(sudo\s+)?(ba|z|da)?sh\b/,  // pipe-to-shell: curl ... | bash, ... | sudo sh
      /\b(ba|z|da)?sh\s+<\(\s*(curl|wget)\b/,  // process substitution: bash <(curl ...)
      /\b(ba|z|da)?sh\s+-c\s+["']?\$\((curl|wget)\b/,  // sh -c "$(curl ...)"
    ];

    for (const pattern of DANGEROUS_PATTERNS) {
      if (pattern.test(command)) {
        process.stderr.write(`[Hook] BLOCKED: Destructive command detected: ${command}\n`);
        process.stdout.write(data);
        process.exit(2);
      }
    }
  } catch {
    // Parse error — pass through
  }

  process.stdout.write(data);
  process.exit(0);
});
