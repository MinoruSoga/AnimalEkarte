#!/usr/bin/env node
/**
 * PostToolUse Hook: Scoped golangci-lint after Go file edits
 *
 * Docker-first: runs golangci-lint scoped to the edited file's package,
 * via `docker compose exec` against the already-running backend container
 * (same execution model as post-edit-format-go.js). Skips silently when
 * the backend service is not up — this hook never starts containers.
 *
 * Non-blocking — always exits 0. Findings go to stderr only.
 *
 * Why scoped, not `./...`: full-repo golangci-lint is on the
 * auto-execution-prohibited list (.claude/CLAUDE.md) because of output
 * size/runtime. `--max-same-issues 0 --max-issues-per-linter 0` disables
 * the default result caps that previously hid issues beyond the 10th of
 * the same kind in this repo (a 27-issue backlog once slipped past CI
 * because of this exact cap).
 */
'use strict';

const fs = require('fs');
const { spawnSync } = require('child_process');

const MAX_STDIN = 1024 * 1024;
const DOCKER_PS_TIMEOUT_MS = 5000;
const LINT_SPAWN_TIMEOUT_MS = 30000;
const LINT_INTERNAL_TIMEOUT = '25s';
const MAX_OUTPUT_CHARS = 4000;

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
    runScopedLint(filePath);
  } catch {
    // Parse error — pass through
  }

  process.stdout.write(data);
  process.exit(0);
});

function runScopedLint(filePath) {
  if (!filePath || !/\.go$/.test(filePath)) return;
  if (!fs.existsSync(filePath)) return;

  // Anchored on a path separator (or string start) so an ancestor directory
  // that merely ends in "backend" (e.g. "mybackend/AnimalEkarte/backend/...")
  // can't be mistaken for the repo's backend/ root.
  const backendMatch = /(?:^|\/)backend\/(.+)$/.exec(filePath);
  if (!backendMatch) return;

  const relativePath = backendMatch[1];
  const lastSlash = relativePath.lastIndexOf('/');
  const relativeDir = lastSlash === -1 ? '' : relativePath.slice(0, lastSlash);
  // Root-level backend/*.go files (e.g. doc.go, tools.go) must NOT fall back
  // to "./..." — that is the exact full-repo recursive scan this hook exists
  // to avoid (.claude/CLAUDE.md auto-execution-prohibited list). "." lints
  // only the root package, non-recursively.
  const target = relativeDir === '' ? '.' : `./${relativeDir}/...`;

  if (!isBackendRunning()) return;

  const result = spawnSync('docker', [
    'compose', 'exec', '-T', 'backend',
    'golangci-lint', 'run',
    '--timeout', LINT_INTERNAL_TIMEOUT,
    '--max-same-issues', '0',
    '--max-issues-per-linter', '0',
    target,
  ], {
    cwd: process.env.CLAUDE_PROJECT_DIR || process.cwd(),
    encoding: 'utf8',
    timeout: LINT_SPAWN_TIMEOUT_MS,
    stdio: ['pipe', 'pipe', 'pipe'],
  });

  if (result.error && result.error.code === 'ETIMEDOUT') {
    process.stderr.write(
      `\n[Hook] golangci-lint timed out scanning ${target} (skipped, non-blocking)\n` +
      `[Hook] Run manually: docker compose exec backend golangci-lint run ${target}\n\n`
    );
    return;
  }

  if (result.status !== 0) {
    const output = (result.stdout || result.stderr || '').trim();
    if (output) {
      const truncated = output.length > MAX_OUTPUT_CHARS
        ? `${output.slice(0, MAX_OUTPUT_CHARS)}\n... (truncated)`
        : output;
      process.stderr.write(
        `\n[Hook] golangci-lint found issues in ${target}:\n${truncated}\n` +
        `[Hook] Full scope: docker compose exec backend golangci-lint run ${target}\n\n`
      );
    }
  }
}

function isBackendRunning() {
  try {
    const check = spawnSync('docker', ['compose', 'ps', 'backend', '--format', 'json'], {
      encoding: 'utf8',
      timeout: DOCKER_PS_TIMEOUT_MS,
      stdio: ['pipe', 'pipe', 'pipe'],
    });
    return check.status === 0 && Boolean(check.stdout && check.stdout.trim());
  } catch {
    return false;
  }
}
