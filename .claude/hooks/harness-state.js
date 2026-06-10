#!/usr/bin/env node
/**
 * PostToolUse Hook: Harness state tracker
 *
 * ハーネス実行中（.claude/logs/harness-active.json が存在する間）、
 * 各ツール実行後にイテレーション番号と変更ファイルを追記する。
 *
 * harness-active.json の形式:
 * {
 *   "task": "BE-042",
 *   "iteration": 2,
 *   "maxIterations": 3,
 *   "startedAt": "ISO8601",
 *   "changedFiles": ["path/to/file.go"],
 *   "iterationResults": [
 *     { "iteration": 1, "result": "FAIL", "violations": ["..."] }
 *   ]
 * }
 *
 * Always exits 0. Never blocks tool execution.
 */
'use strict';

const fs = require('fs');
const path = require('path');
const { execSync } = require('child_process');

const MAX_STDIN = 512 * 1024;
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
    const statePath = process.env.HARNESS_STATE_PATH
      ? path.resolve(projectDir, process.env.HARNESS_STATE_PATH)
      : path.join(projectDir, '.claude', 'logs', 'harness-active.json');

    // ハーネスが起動していない場合は何もしない
    if (!fs.existsSync(statePath)) {
      process.exit(0);
    }

    const state = JSON.parse(fs.readFileSync(statePath, 'utf8'));

    // Edit/Write ツール実行後に変更ファイルを記録
    let input;
    try {
      input = JSON.parse(data);
    } catch {
      process.exit(0);
    }

    const toolName = input.tool_name || '';
    if (toolName === 'Edit' || toolName === 'Write') {
      const filePath = input.tool_input?.file_path || input.tool_input?.path || '';
      if (filePath) {
        const rel = path.relative(projectDir, filePath);
        if (!state.changedFiles.includes(rel)) {
          state.changedFiles.push(rel);
        }
      }
    }

    // git diff で最新の変更ファイル一覧を同期
    try {
      const diff = execSync('git diff --name-only HEAD', {
        cwd: projectDir,
        encoding: 'utf8',
        timeout: 5000,
      }).trim();
      if (diff) {
        const files = diff.split('\n').filter(Boolean);
        for (const f of files) {
          if (!state.changedFiles.includes(f)) {
            state.changedFiles.push(f);
          }
        }
      }
    } catch {
      // git error — skip
    }

    state.updatedAt = new Date().toISOString();
    fs.writeFileSync(statePath, JSON.stringify(state, null, 2), 'utf8');
  } catch {
    // Silently fail — never block
  }

  process.exit(0);
});
