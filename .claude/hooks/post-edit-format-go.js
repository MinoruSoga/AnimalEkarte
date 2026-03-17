#!/usr/bin/env node
/**
 * PostToolUse Hook: Auto-format Go files with gofmt after edits
 *
 * Docker-first: runs gofmt via docker compose exec.
 * Non-blocking — logs errors to stderr but always exits 0.
 */
'use strict';

const { spawnSync } = require('child_process');
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

    if (filePath && /\.go$/.test(filePath)) {
      // Extract path relative to backend/ for Docker container
      const backendIdx = filePath.indexOf('backend/');
      if (backendIdx !== -1) {
        const relativePath = filePath.substring(backendIdx + 'backend/'.length);
        const containerPath = `/app/${relativePath}`;

        const result = spawnSync('docker', [
          'compose', 'exec', '-T', 'backend', 'gofmt', '-w', containerPath
        ], {
          cwd: process.env.CLAUDE_PROJECT_DIR || process.cwd(),
          encoding: 'utf8',
          timeout: 15000,
          stdio: ['pipe', 'pipe', 'pipe']
        });

        if (result.status !== 0 && result.stderr) {
          process.stderr.write(`[Hook] gofmt warning: ${result.stderr.trim()}\n`);
        }
      }
    }
  } catch {
    // Parse error — pass through
  }

  process.stdout.write(data);
  process.exit(0);
});
