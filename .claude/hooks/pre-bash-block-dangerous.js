#!/usr/bin/env node
/**
 * PreToolUse Hook: Block destructive shell commands
 *
 * Blocks:
 * - catastrophic system damage (rm -rf /, dd, mkfs, fork bomb, pipe-to-shell)
 * - irreversible git working-tree / history rewrites agents must not auto-run
 *   (git reset --hard, git clean -fdx, force-push, discard-all checkout/restore)
 *
 * Protocol:
 * - Exit 2 = block (Claude Code / Grok)
 * - On deny, emit {"decision":"deny","reason":"..."} for Grok PreToolUse
 * - Fail-open only on parse errors (hook must stay available)
 */
'use strict';

const MAX_STDIN = 1024 * 1024;
let data = '';
process.stdin.setEncoding('utf8');

process.stdin.on('data', (chunk) => {
  if (data.length < MAX_STDIN) {
    data += chunk.substring(0, MAX_STDIN - data.length);
  }
});

/**
 * Extract the shell command from Claude / Grok / Cursor-shaped payloads.
 * @param {Record<string, unknown>} input
 * @returns {string}
 */
function extractCommand(input) {
  const toolInput = input.tool_input || input.toolInput || input.parameters || {};
  if (typeof toolInput === 'object' && toolInput !== null) {
    if (typeof toolInput.command === 'string') return toolInput.command;
    if (typeof toolInput.cmd === 'string') return toolInput.cmd;
  }
  if (typeof input.command === 'string') return input.command;
  return '';
}

/**
 * Patterns that permanently discard uncommitted work or rewrite shared history.
 * Keep these narrow enough that routine git status/diff/add/commit/checkout branch still work.
 */
const DESTRUCTIVE_GIT_PATTERNS = [
  // reset --hard (any ref / bare)
  /\bgit\s+reset\s+(-[a-zA-Z]*H[a-zA-Z]*\b|--hard\b)/,
  // clean that deletes untracked files/dirs
  /\bgit\s+clean\s+[^\n;&|]*-[^\n;&|]*f[^\n;&|]*d/,
  /\bgit\s+clean\s+[^\n;&|]*-[^\n;&|]*d[^\n;&|]*f/,
  /\bgit\s+clean\s+[^\n;&|]*--force\b[^\n;&|]*(-d|--d\b)/,
  // discard entire worktree contents
  /\bgit\s+checkout\s+(?:--\s+)?\.(?:\s|$)/,
  /\bgit\s+restore\s+(?:--(?:worktree|staged|source=\S+)\s+)*\.(?:\s|$)/,
  /\bgit\s+restore\s+[^\n;&|]*--\s+\.(?:\s|$)/,
  // force push family (shared remote rewrite)
  /\bgit\s+push\b[^\n;&|]*--force\b/,
  /\bgit\s+push\b[^\n;&|]*--force-with-lease\b/,
  // short -f must not match substrings inside other flags; allow git push -f / git push origin -f main
  /\bgit\s+push\b[^\n;&|]*?(?:^|[\s])-f(?:\s|$)/,
  // branch delete force
  /\bgit\s+branch\b[^\n;&|]*\s-D\s+/,
];

const SYSTEM_DESTRUCTIVE_PATTERNS = [
  /rm\s+-rf\s+\//,
  /rm\s+-rf\s+~/,
  />\s*\/dev\/sda/,
  /\bmkfs\b/,
  /\bdd\s+if=\/dev\/zero/,
  /:\(\)\{.*\|.*&\}\;:/, // fork bomb
  /\|\s*(sudo\s+)?(ba|z|da)?sh\b/, // pipe-to-shell
  /\b(ba|z|da)?sh\s+<\(\s*(curl|wget)\b/,
  /\b(ba|z|da)?sh\s+-c\s+["']?\$\((curl|wget)\b/,
];

/**
 * @param {string} command
 * @returns {string|null} reason if blocked
 */
function matchDestructive(command) {
  for (const pattern of DESTRUCTIVE_GIT_PATTERNS) {
    if (pattern.test(command)) {
      return (
        'Blocked destructive git command. ' +
        'Do not use git reset --hard, git clean -fd(x), git checkout/restore ., or force-push. ' +
        'Prefer: git fetch + merge/ff-only, path-scoped git restore <path>, or a dedicated git worktree. ' +
        'For parallel agents, use isolated worktrees — never share one working tree.'
      );
    }
  }
  for (const pattern of SYSTEM_DESTRUCTIVE_PATTERNS) {
    if (pattern.test(command)) {
      return 'Blocked destructive system command.';
    }
  }
  return null;
}

/**
 * @param {string} reason
 * @param {string} command
 */
function deny(reason, command) {
  const preview = command.length > 200 ? `${command.slice(0, 200)}…` : command;
  const message = `[Hook] BLOCKED: ${reason}\nCommand: ${preview}\n`;
  process.stderr.write(message);
  // Grok PreToolUse decision envelope (also harmless for Claude Code)
  process.stdout.write(
    JSON.stringify({
      decision: 'deny',
      reason,
      permissionDecision: 'deny',
      permissionDecisionReason: reason,
    })
  );
  process.exit(2);
}

process.stdin.on('end', () => {
  try {
    const input = JSON.parse(data);
    const command = extractCommand(input);
    if (command) {
      const reason = matchDestructive(command);
      if (reason) {
        deny(reason, command);
        return;
      }
    }
  } catch {
    // Parse error — fail-open (pass through)
  }

  process.stdout.write(data);
  process.exit(0);
});
