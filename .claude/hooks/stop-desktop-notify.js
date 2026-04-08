#!/usr/bin/env node
/**
 * Stop Hook: Desktop notification on macOS
 *
 * Sends a native macOS notification when Claude finishes a response.
 * Useful when running long tasks — you can switch to another window
 * and get notified when Claude is done.
 *
 * Non-blocking (async). Always exits 0.
 */
'use strict';

const { execFile } = require('child_process');

const MAX_STDIN = 1024 * 1024;
let data = '';
process.stdin.setEncoding('utf8');

process.stdin.on('data', chunk => {
  if (data.length < MAX_STDIN) {
    data += chunk.substring(0, MAX_STDIN - data.length);
  }
});

process.stdin.on('end', () => {
  // Pass through immediately — notification is fire-and-forget
  process.stdout.write(data);

  if (process.platform !== 'darwin') {
    process.exit(0);
  }

  const title = 'Claude Code';
  const message = 'Task completed — check your terminal.';

  const script = `display notification "${message}" with title "${title}" sound name "Glass"`;

  execFile('osascript', ['-e', script], { timeout: 5000 }, () => {
    // Ignore errors (e.g., no display, headless)
    process.exit(0);
  });
});
