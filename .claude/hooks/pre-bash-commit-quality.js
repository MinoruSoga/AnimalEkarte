#!/usr/bin/env node
/**
 * PreToolUse Hook: Commit quality gate
 *
 * Runs before `git commit` to catch issues that should never be committed:
 *  1. console.log / debugger statements in staged .ts/.tsx/.js/.jsx files
 *  2. Secret patterns (API keys, AWS keys, GitHub PATs, passwords)
 *  3. Commit message format validation (conventional commits)
 *
 * Exit code 2 = block the commit.
 * Exit code 0 = allow.
 */
'use strict';

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
  try {
    const input = JSON.parse(data);
    const command = String(input.tool_input?.command || '');

    // Only trigger on git commit commands
    if (!/\bgit\s+commit\b/.test(command)) {
      process.stdout.write(data);
      process.exit(0);
    }

    const issues = [];
    const projectDir = process.env.CLAUDE_PROJECT_DIR || process.cwd();

    // --- 1. Check staged files for console.log / debugger ---
    let stagedFiles = [];
    try {
      const output = execSync('git diff --cached --name-only --diff-filter=ACM', {
        cwd: projectDir,
        encoding: 'utf8',
        timeout: 10000,
      }).trim();
      stagedFiles = output ? output.split('\n') : [];
    } catch {
      // No staged files or git error — skip
    }

    const jsFiles = stagedFiles.filter(f => /\.(ts|tsx|js|jsx)$/.test(f));

    for (const file of jsFiles) {
      let content;
      try {
        content = execSync(`git show ":${file}"`, {
          cwd: projectDir,
          encoding: 'utf8',
          timeout: 5000,
        });
      } catch {
        continue;
      }

      const lines = content.split('\n');
      lines.forEach((line, idx) => {
        // Skip comments and test files
        const trimmed = line.trim();
        if (trimmed.startsWith('//') || trimmed.startsWith('*')) return;
        if (file.includes('.test.') || file.includes('__tests__')) return;

        if (/\bconsole\.log\b/.test(line)) {
          issues.push(`[console.log] ${file}:${idx + 1}: ${trimmed}`);
        }
        if (/\bdebugger\b/.test(line) && !trimmed.startsWith('//')) {
          issues.push(`[debugger] ${file}:${idx + 1}: ${trimmed}`);
        }
      });
    }

    // --- 2. Check for secrets in ALL staged files ---
    const SECRET_PATTERNS = [
      { pattern: /(?:^|['"`\s=])(?:sk-[a-zA-Z0-9]{20,})/, label: 'OpenAI API key' },
      { pattern: /(?:^|['"`\s=])(?:ghp_[a-zA-Z0-9]{36,})/, label: 'GitHub PAT' },
      { pattern: /(?:^|['"`\s=])(?:ghs_[a-zA-Z0-9]{36,})/, label: 'GitHub App token' },
      { pattern: /(?:^|['"`\s=])(?:AKIA[A-Z0-9]{16})/, label: 'AWS Access Key' },
      { pattern: /(?:^|['"`\s=])(?:sk-ant-[a-zA-Z0-9-]{20,})/, label: 'Anthropic API key' },
      { pattern: /password\s*[:=]\s*['"][^'"]{8,}['"]/, label: 'Hardcoded password' },
      { pattern: /secret\s*[:=]\s*['"][^'"]{8,}['"]/, label: 'Hardcoded secret' },
    ];

    for (const file of stagedFiles) {
      // Skip binary, generated, and lock files
      if (/\.(png|jpg|jpeg|gif|svg|ico|woff|ttf|eot|lock)$/.test(file)) continue;
      if (file.includes('node_modules') || file.includes('vendor')) continue;
      if (file === 'package-lock.json' || file === 'go.sum') continue;

      let content;
      try {
        content = execSync(`git show ":${file}"`, {
          cwd: projectDir,
          encoding: 'utf8',
          timeout: 5000,
        });
      } catch {
        continue;
      }

      for (const { pattern, label } of SECRET_PATTERNS) {
        if (pattern.test(content)) {
          issues.push(`[SECRET] ${file}: Possible ${label} detected`);
        }
      }
    }

    // --- 3. Validate commit message format (conventional commits) ---
    // Extract -m "message" from the command
    const msgMatch = command.match(/-m\s+(?:"([^"]*(?:"[^"]*"[^"]*)*)"|'([^']*)'|(\S+))/);
    if (msgMatch) {
      const msg = (msgMatch[1] || msgMatch[2] || msgMatch[3] || '').split('\n')[0];
      const conventionalPattern = /^(feat|fix|refactor|test|docs|ci|chore|perf|style|build|revert)(\(.+\))?!?:\s.+/;

      if (msg && !conventionalPattern.test(msg) && !msg.startsWith('Merge ')) {
        issues.push(`[commit-msg] Message does not follow conventional commits: "${msg}"`);
        issues.push(`[commit-msg] Expected: type(scope): subject`);
      }
    }

    // --- Report ---
    if (issues.length > 0) {
      const secrets = issues.filter(i => i.startsWith('[SECRET]'));
      const others = issues.filter(i => !i.startsWith('[SECRET]'));

      if (secrets.length > 0) {
        // Secrets are hard blocks
        process.stderr.write(`\n[Hook] BLOCKED: ${secrets.length} secret(s) detected in staged files:\n`);
        secrets.forEach(i => process.stderr.write(`  ${i}\n`));
        process.stderr.write('[Hook] Remove secrets before committing.\n\n');
        process.stdout.write(data);
        process.exit(2);
      }

      if (others.length > 0) {
        // console.log, debugger, commit message — warn but allow
        process.stderr.write(`\n[Hook] WARNING: ${others.length} quality issue(s) found:\n`);
        others.slice(0, 10).forEach(i => process.stderr.write(`  ${i}\n`));
        if (others.length > 10) {
          process.stderr.write(`  ... and ${others.length - 10} more\n`);
        }
        process.stderr.write('\n');
      }
    }
  } catch {
    // Parse error — pass through
  }

  process.stdout.write(data);
  process.exit(0);
});
