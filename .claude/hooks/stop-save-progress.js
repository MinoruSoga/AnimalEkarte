#!/usr/bin/env node
/**
 * Stop Hook: Save session progress for cross-session context persistence
 *
 * After each Claude response, saves a compact progress snapshot to
 * .claude/logs/session-progress.md. The SessionStart hook loads this
 * file so the next session (or post-compaction context) knows where
 * the previous work left off.
 *
 * Saved data:
 *   - Timestamp
 *   - Git branch, uncommitted changes, recent commits
 *   - Transcript path (for detailed context recovery)
 *
 * Always exits 0. Non-blocking writes.
 */
'use strict';

const fs = require('fs');
const path = require('path');
const { execSync } = require('child_process');

const MAX_STDIN = 256 * 1024; // 256KB — we only need metadata, not full response
let data = '';
process.stdin.setEncoding('utf8');

process.stdin.on('data', chunk => {
  if (data.length < MAX_STDIN) {
    data += chunk.substring(0, MAX_STDIN - data.length);
  }
});

process.stdin.on('end', () => {
  // Pass through stdin immediately
  process.stdout.write(data);

  try {
    const projectDir = process.env.CLAUDE_PROJECT_DIR || process.cwd();
    const logsDir = path.join(projectDir, '.claude', 'logs');
    fs.mkdirSync(logsDir, { recursive: true });

    const run = (cmd) => {
      try {
        return execSync(cmd, { cwd: projectDir, encoding: 'utf8', timeout: 5000 }).trim();
      } catch {
        return '';
      }
    };

    // Parse stdin for transcript_path (Stop hooks receive session metadata)
    let transcriptPath = '';
    try {
      const input = JSON.parse(data);
      transcriptPath = input.transcript_path || '';
    } catch {
      // Not JSON or no transcript_path — that's OK
    }

    const timestamp = new Date().toISOString();
    const branch = run('git branch --show-current');
    const status = run('git status --short');
    const lastCommit = run('git log --oneline -1');
    const recentCommits = run('git log --oneline -5');
    const diffFiles = run('git diff --name-only');

    const lines = [
      `# Session Progress`,
      ``,
      `**Last updated:** ${timestamp}`,
      `**Branch:** ${branch}`,
      `**Last commit:** ${lastCommit}`,
    ];

    if (transcriptPath) {
      lines.push(`**Transcript:** ${transcriptPath}`);
    }

    if (status) {
      lines.push(``, `## Uncommitted Changes`, '```', status, '```');
    }

    if (diffFiles) {
      lines.push(``, `## Modified Files`, '```', diffFiles, '```');
    }

    lines.push(``, `## Recent Commits`, '```', recentCommits, '```', '');

    fs.writeFileSync(
      path.join(logsDir, 'session-progress.md'),
      lines.join('\n'),
      'utf8'
    );
  } catch {
    // Silently fail — never block the session
  }

  process.exit(0);
});
