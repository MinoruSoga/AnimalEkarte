#!/usr/bin/env node
/**
 * PreToolUse Hook: Block shell reads of .env files
 *
 * The permissions deny list (Read(./.env) etc.) only governs the Read tool.
 * Shell utilities in the allow list (cat, grep, head, sed, ...) could read
 * .env files anyway. This hook closes that gap: any Bash command that
 * references a .env file AND contains a read/copy utility is blocked.
 * Template files (.env.example / .env.sample / .env.template / .env.dist)
 * are exempt. Exit code 2 = block the tool call.
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

// .env file reference: ".env" preceded by start / whitespace / quote / = / path
// separator, optionally followed by a suffix (.local, .production, ...).
const ENV_FILE_RE = /(^|[\s'"=/])\.env(\.[\w.-]+)?(?=$|[\s'"`;|&)<>])/g;
const TEMPLATE_SUFFIX_RE = /^\.(example|sample|template|dist)\b/;

// Utilities that can expose file contents (read, copy, edit, source).
const READ_UTIL_RE = new RegExp(
  '(^|[\\s;|&`$(])(' +
  ['cat', 'head', 'tail', 'less', 'more', 'grep', 'rg', 'sed', 'awk', 'cut',
    'sort', 'uniq', 'tr', 'strings', 'xxd', 'od', 'hexdump', 'base64', 'tee',
    'cp', 'mv', 'install', 'rsync', 'source', 'vim', 'vi', 'nano', 'emacs',
    'code', 'open', 'python3?', 'node', 'perl', 'ruby'].join('|') +
  ')\\b'
);
const DOT_SOURCE_RE = /(^|[;|&])\s*\.\s+\S*\.env/;

process.stdin.on('end', () => {
  try {
    const input = JSON.parse(data);
    const command = String(input.tool_input?.command || '');

    const refs = [];
    let match;
    while ((match = ENV_FILE_RE.exec(command)) !== null) {
      const suffix = match[2] || '';
      if (!TEMPLATE_SUFFIX_RE.test(suffix)) {
        refs.push(`.env${suffix}`);
      }
    }

    const canExpose = READ_UTIL_RE.test(command) || DOT_SOURCE_RE.test(command);
    if (refs.length > 0 && canExpose) {
      process.stderr.write(
        `[Hook] BLOCKED: shell read of ${refs.join(', ')} — .env contents must not enter the session. ` +
        'Use ${VAR:+set} to check presence without leaking values.\n'
      );
      process.stdout.write(data);
      process.exit(2);
    }
  } catch {
    // Parse error — pass through
  }

  process.stdout.write(data);
  process.exit(0);
});
