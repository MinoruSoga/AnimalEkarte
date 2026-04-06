#!/usr/bin/env node
/**
 * PreToolUse Hook: Block --no-verify and --no-gpg-sign flags
 *
 * Prevents bypassing git hooks (pre-commit, commit-msg, pre-push).
 * This is a security-critical hook — if a hook fails, the correct
 * action is to fix the underlying issue, not skip the check.
 *
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

    // Strip quoted strings (commit messages, heredocs) to avoid false positives
    // e.g. git commit -m "added --no-verify blocker" should NOT be blocked
    const commandWithoutQuotes = command
      .replace(/"[^"]*"/g, '""')       // double-quoted strings
      .replace(/'[^']*'/g, "''")       // single-quoted strings
      .replace(/\$\(cat\s+<<[\s\S]*?EOF\s*\)/g, '""');  // heredoc in $(cat <<EOF...EOF)

    const BLOCKED_FLAGS = [
      { pattern: /--no-verify/, label: '--no-verify' },
      { pattern: /--no-gpg-sign/, label: '--no-gpg-sign' },
      // -n is --no-verify for git commit only (NOT git push, where -n is --dry-run)
      { pattern: /\bgit\s+commit\b.*\s-[a-zA-Z]*n\b/, label: '-n flag on git commit (--no-verify shorthand)' },
    ];

    for (const { pattern, label } of BLOCKED_FLAGS) {
      if (pattern.test(commandWithoutQuotes)) {
        process.stderr.write(`[Hook] BLOCKED: "${label}" flag detected in: ${command}\n`);
        process.stderr.write('[Hook] Fix the underlying issue instead of bypassing hooks.\n');
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
