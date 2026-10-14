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
 * - Fail-closed on oversized, malformed or conflicting input; deny reasons
 *   never echo command text or raw stdin
 */
'use strict';
const { writeSync } = require('fs');
const { literalDeviceWrite } = require('./literal-device-write.js');

const MAX_STDIN = 1024 * 1024;
const chunks = [];
let size = 0;

/**
 * @param {string} reason
 * @returns {void}
 */
function deny(reason) {
  process.stderr.write(`[Hook] BLOCKED: ${reason}\n`);
  // Grok PreToolUse decision envelope (also harmless for Claude Code)
  writeSync(1,
    JSON.stringify({
      decision: 'deny',
      reason,
      permissionDecision: 'deny',
      permissionDecisionReason: reason,
    })
  );
  process.exit(2);
}

process.stdin.on('data', (chunk) => {
  size += chunk.length;
  if (size > MAX_STDIN) {
    deny('Blocked oversized hook input.');
    return;
  }
  chunks.push(chunk);
});

process.stdin.on('error', () => {
  deny('Blocked unreadable hook input.');
});

function plainObject(value) {
  return value !== null && typeof value === 'object' && !Array.isArray(value);
}

/**
 * Extract the shell command and tool name from Claude / Grok / Cursor-shaped payloads.
 * @param {Record<string, unknown>} input
 * @returns {{ command: string, toolName: string|null, hasContainer: boolean, error: string|null }}
 */
function inspectEnvelope(input) {
  const containers = [];
  for (const key of ['tool_input', 'toolInput', 'parameters']) {
    if (key in input) {
      if (!plainObject(input[key])) {
        return { command: '', toolName: null, hasContainer: false, error: 'Blocked malformed hook input.' };
      }
      containers.push(input[key]);
    }
  }
  const commands = [];
  for (const container of containers) {
    for (const key of ['command', 'cmd']) {
      if (key in container) commands.push(container[key]);
    }
  }
  if ('command' in input) commands.push(input.command);
  for (const value of commands) {
    if (typeof value !== 'string' || !value.trim()) {
      return { command: '', toolName: null, hasContainer: containers.length > 0, error: 'Blocked malformed hook input.' };
    }
  }
  for (let i = 1; i < commands.length; i++) {
    if (commands[i] !== commands[0]) {
      return { command: '', toolName: null, hasContainer: containers.length > 0, error: 'Blocked conflicting hook input.' };
    }
  }
  const names = [];
  for (const key of ['tool_name', 'toolName']) {
    if (key in input) names.push(input[key]);
  }
  for (const value of names) {
    if (typeof value !== 'string' || !value.trim()) {
      return { command: '', toolName: null, hasContainer: containers.length > 0, error: 'Blocked malformed hook input.' };
    }
  }
  for (let i = 1; i < names.length; i++) {
    if (names[i] !== names[0]) {
      return { command: '', toolName: null, hasContainer: containers.length > 0, error: 'Blocked conflicting hook input.' };
    }
  }
  return { command: commands[0] || '', toolName: names[0] || null, hasContainer: containers.length > 0, error: null };
}

const TARGET_TOOLS = new Set(['Bash', 'bash', 'Shell', 'shell', 'run_terminal_command']);

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
  if (literalDeviceWrite(command)) return 'Blocked literal raw-device write (redirect, tee or dd).';
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

process.stdin.on('end', () => {
  let text;
  try {
    text = new TextDecoder('utf-8', { fatal: true }).decode(Buffer.concat(chunks));
  } catch {
    deny('Blocked malformed hook input.');
    return;
  }
  if (!text.trim()) {
    deny('Blocked malformed hook input.');
    return;
  }
  let input;
  try {
    input = JSON.parse(text);
  } catch {
    deny('Blocked malformed hook input.');
    return;
  }
  if (!plainObject(input)) {
    deny('Blocked malformed hook input.');
    return;
  }
  const { command, toolName, hasContainer, error } = inspectEnvelope(input);
  if (error) {
    deny(error);
    return;
  }
  if (!command) {
    if (toolName && !TARGET_TOOLS.has(toolName)) {
      if (hasContainer) {
        process.exit(0);
      }
      deny('Blocked malformed hook input.');
      return;
    }
    deny('Blocked shell event without a command.');
    return;
  }
  const reason = matchDestructive(command);
  if (reason) {
    deny(reason);
    return;
  }
  process.exit(0);
});
