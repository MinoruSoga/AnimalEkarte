#!/usr/bin/env node
/**
 * PreCompact Hook: Save working state before context compaction
 *
 * When Claude's context window fills up, it compacts (summarizes) the
 * conversation. This hook saves the current git state and any in-progress
 * information to a file, so the post-compaction context has a recovery point.
 *
 * Output file: .claude/logs/pre-compact-state.md
 * Always exits 0 (non-blocking).
 */
'use strict';

const fs = require('fs');
const path = require('path');
const { execSync } = require('child_process');

const MAX_STDIN = 1024 * 1024;
let data = '';
process.stdin.setEncoding('utf8');

process.stdin.on('data', chunk => {
  if (data.length < MAX_STDIN) {
    data += chunk.substring(0, MAX_STDIN - data.length);
  }
});

process.stdin.on('end', () => {
  process.stdout.write(data);

  try {
    const projectDir = process.env.CLAUDE_PROJECT_DIR || process.cwd();
    const logsDir = path.join(projectDir, '.claude', 'logs');

    // Ensure logs directory exists
    fs.mkdirSync(logsDir, { recursive: true });

    const run = (cmd) => {
      try {
        return execSync(cmd, { cwd: projectDir, encoding: 'utf8', timeout: 10000 }).trim();
      } catch {
        return '(error)';
      }
    };

    const timestamp = new Date().toISOString();
    const branch = run('git branch --show-current');
    const status = run('git status --short');
    const recentCommits = run('git log --oneline -5');
    const diffStat = run('git diff --stat');

    const state = [
      `# Pre-Compaction State`,
      ``,
      `**Saved at:** ${timestamp}`,
      `**Branch:** ${branch}`,
      ``,
      `## Uncommitted Changes`,
      '```',
      status || '(clean)',
      '```',
      ``,
      `## Diff Summary`,
      '```',
      diffStat || '(no changes)',
      '```',
      ``,
      `## Recent Commits`,
      '```',
      recentCommits,
      '```',
      '',
    ].join('\n');

    fs.writeFileSync(path.join(logsDir, 'pre-compact-state.md'), state, 'utf8');
    process.stderr.write(`[Hook] Saved pre-compaction state to .claude/logs/pre-compact-state.md\n`);
  } catch (err) {
    process.stderr.write(`[Hook] Failed to save pre-compaction state: ${err.message}\n`);
  }

  process.exit(0);
});
