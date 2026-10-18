#!/usr/bin/env node
/** Advisory only. Never executes Docker, a compiler, a runner, or a shell. */
'use strict';

const MAX_STDIN = 1024 * 1024;
let bytes = 0;
let oversized = false;
const chunks = [];
process.stdin.on('data', chunk => {
  bytes += chunk.length;
  if (bytes > MAX_STDIN) oversized = true;
  if (!oversized) chunks.push(chunk);
});
process.stdin.on('end', () => {
  if (oversized) {
    process.stderr.write('[Hook] TS verification SKIP: input-limit.\n');
    return;
  }
  try {
    const raw = new TextDecoder('utf-8', { fatal: true }).decode(Buffer.concat(chunks));
    const input = JSON.parse(raw);
    const file = input?.tool_input?.file_path || input?.tool_input?.path;
    if (['Edit', 'Write', 'MultiEdit'].includes(input?.tool_name) && typeof file === 'string' && /\.(ts|tsx)$/.test(file)) {
      process.stderr.write('[Hook] TS verification SKIP: use scoped verification after confirming target worktree, diff and runner mount; see docs/ops/agent-harness.md. No type check was run.\n');
    }
  } catch {
    process.stderr.write('[Hook] TS verification SKIP: invalid-input.\n');
  }
});
