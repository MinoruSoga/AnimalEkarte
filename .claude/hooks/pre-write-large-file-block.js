#!/usr/bin/env node
/**
 * PreToolUse Hook: Warn when a Write creates a file larger than 800 lines
 *
 * 800 lines is a maintainability guideline, not a hard limit (documented
 * exceptions such as frontend/src/lib/design-tokens.ts exist), so this hook
 * warns and never blocks. The file name is kept for .claude/settings.json.
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
        `[Hook] WARNING: File exceeds the 800-line guideline (${lineCount} lines): ${filePath}\n` +
        '[Hook] Consider splitting unless it is a documented exception.\n'
      );
    }
  } catch {
    // Parse error — pass through
  }

  process.stdout.write(data);
  process.exit(0);
});
