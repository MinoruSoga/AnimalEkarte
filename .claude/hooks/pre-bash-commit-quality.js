#!/usr/bin/env node
/**
 * PreToolUse Hook: Commit quality gate
 *
 * Runs before `git commit` to catch issues that should never be committed:
 *  1. console.log / debugger statements in staged .ts/.tsx/.js/.jsx files
 *  2. Secret patterns (API keys, AWS keys, GitHub PATs, passwords)
 *  3. Commit message format validation (conventional commits)
 *  4. Mirror sync: if .claude/skills, .claude/commands, or .claude/agents are
 *     staged, regenerate .agents/skills and .codex/agents+commands so the
 *     mirrors never silently drift from the source of truth again (#1/#2).
 *     Disable with SYNC_MIRRORS_DISABLED=1. Failures block the commit and are
 *     logged to .claude/logs/sync-mirrors.log (⑤ 鉄則: 停止手段/失敗通知/監査ログ).
 *
 * Exit code 2 = block the commit.
 * Exit code 0 = allow.
 */
'use strict';

const { execFileSync } = require('child_process');
const fs = require('fs');
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
    const command = String(input.tool_input?.command || '');

    // Only trigger on git commit commands
    if (!/\bgit\s+commit\b/.test(command)) {
      process.stdout.write(data);
      process.exit(0);
    }

    const issues = [];
    const projectDir = process.env.CLAUDE_PROJECT_DIR || process.cwd();

    // Scan the repo the commit actually targets: `git -C <dir>` / `cd <dir> && git commit`
    // aimed at another repo must not be judged by this project's staged files — and must
    // still be secret-scanned against its own staged files.
    let scanDir = projectDir;
    const dirMatch =
      command.match(/\bgit\s+-C\s+(?:"([^"]+)"|'([^']+)'|([^\s;&|]+))/) ||
      command.match(/(?:^|&&|;)\s*cd\s+(?:"([^"]+)"|'([^']+)'|([^\s;&|]+))/);
    if (dirMatch) {
      const rawDir = dirMatch[1] || dirMatch[2] || dirMatch[3] || '';
      const expanded = rawDir.replace(/^~(?=$|\/)/, process.env.HOME || '~');
      if (expanded && fs.existsSync(expanded)) {
        scanDir = path.resolve(expanded);
      }
    }
    const isProjectRepo = scanDir === path.resolve(projectDir);

    // --- 1. Check staged files for console.log / debugger ---
    // argv + NUL delimiters: never shell-interpolate staged paths (SEC-CS-F02)
    let stagedFiles = [];
    try {
      const output = execFileSync(
        'git',
        ['diff', '--cached', '--name-only', '-z', '--diff-filter=ACM'],
        {
          cwd: scanDir,
          encoding: 'utf8',
          timeout: 10000,
        },
      );
      stagedFiles = output ? output.split('\0').filter(Boolean) : [];
    } catch {
      // No staged files or git error — skip
    }

    const jsFiles = stagedFiles.filter(f => /\.(ts|tsx|js|jsx)$/.test(f));

    for (const file of jsFiles) {
      let content;
      try {
        content = execFileSync('git', ['show', ':' + file], {
          cwd: scanDir,
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
      // Skip SQL migration/seed files — credentials in seed data are intentional
      if (/migrations\/.*\.sql$/.test(file)) continue;

      let content;
      try {
        content = execFileSync('git', ['show', ':' + file], {
          cwd: scanDir,
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

    // --- 4. Mirror sync (regenerate .agents/ and .codex/agents+commands) ---
    if (isProjectRepo && process.env.SYNC_MIRRORS_DISABLED !== '1') {
      const mirrorTriggerFiles = stagedFiles.filter(
        f => f.startsWith('.claude/skills/')
          || f.startsWith('.claude/commands/')
          || f.startsWith('.claude/agents/')
          || f.startsWith('.claude/rules/')
          || f.startsWith('.claude/scripts/sync-codex-mirror')
          || f === '.claude/scripts/test_sync_codex_mirror.py'
      );
      if (mirrorTriggerFiles.length > 0) {
        const logPath = path.join(projectDir, '.claude', 'logs', 'sync-mirrors.log');
        const logLine = (msg) => {
          try {
            fs.mkdirSync(path.dirname(logPath), { recursive: true });
            fs.appendFileSync(logPath, `${new Date().toISOString()} ${msg}\n`);
          } catch {
            // logging is best-effort; never let it block the commit path itself
          }
        };
        const scripts = [
          ['.claude/scripts/sync-agents-skills.sh', 'bash'],
          ['.claude/scripts/sync-codex-mirror.sh', 'bash'],
        ];
        let syncFailed = false;
        for (const [rel, shell] of scripts) {
          try {
            const out = execFileSync(shell, [rel], {
              cwd: projectDir,
              encoding: 'utf8',
              timeout: 30000,
            });
            logLine(`OK ${rel} (triggered by: ${mirrorTriggerFiles.join(', ')})\n${out.trim()}`);
          } catch (err) {
            syncFailed = true;
            const out = (err.stdout || '') + (err.stderr || err.message || '');
            logLine(`FAIL ${rel}: ${out.trim()}`);
            issues.push(`[MIRROR-SYNC] ${rel} failed — see .claude/logs/sync-mirrors.log`);
          }
        }
        // .agents/ and .codex/agents|commands are gitignored by design (local
        // generated mirrors) — nothing to `git add`, the regeneration above is
        // enough to keep the working tree in sync.
      }
    }

    // --- Report ---
    if (issues.length > 0) {
      const secrets = issues.filter(i => i.startsWith('[SECRET]'));
      const syncFailures = issues.filter(i => i.startsWith('[MIRROR-SYNC]'));
      const others = issues.filter(i => !i.startsWith('[SECRET]') && !i.startsWith('[MIRROR-SYNC]'));

      if (secrets.length > 0) {
        // Secrets are hard blocks
        process.stderr.write(`\n[Hook] BLOCKED: ${secrets.length} secret(s) detected in staged files:\n`);
        secrets.forEach(i => process.stderr.write(`  ${i}\n`));
        process.stderr.write('[Hook] Remove secrets before committing.\n\n');
        process.stdout.write(data);
        process.exit(2);
      }

      if (syncFailures.length > 0) {
        // Mirror regeneration failing silently is exactly the failure mode this
        // hook exists to prevent (#1) — hard block, don't just warn.
        process.stderr.write(`\n[Hook] BLOCKED: mirror sync failed:\n`);
        syncFailures.forEach(i => process.stderr.write(`  ${i}\n`));
        process.stderr.write('[Hook] Fix the sync script or set SYNC_MIRRORS_DISABLED=1 to bypass intentionally.\n\n');
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
