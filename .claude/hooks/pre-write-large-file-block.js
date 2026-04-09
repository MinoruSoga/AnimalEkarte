#!/usr/bin/env node
/**
 * PreToolUse Hook: Block creation of files larger than 800 lines
 *
 * Enforces the project rule: "1 file < 500 lines" (hard limit at 800).
 * Exit code 2 = block the tool call.
 */
'use strict';

const MAX_STDIN = 2 * 1024 * 1024; // 2MB for Write content
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
    const content = String(input.tool_input?.content || '');
    const filePath = String(input.tool_input?.file_path || '');

    // Skip generated files, migrations, and test files
    if (/generated|migrations|vendor|node_modules|_test\.go$|\.test\.(ts|tsx)$/.test(filePath)) {
      process.stdout.write(data);
      process.exit(0);
    }

    const lineCount = content.split('\n').length;

    if (lineCount > 800) {
      process.stderr.write(
        `[Hook] BLOCKED: File exceeds 800 lines (${lineCount} lines): ${filePath}\n` +
        '[Hook] Split into smaller, focused modules.\n'
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
