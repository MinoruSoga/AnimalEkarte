#!/usr/bin/env node
/**
 * Regression tests for pre-write-large-file-block.js
 * 800 lines is a guideline, not a hard limit: the hook warns and never blocks.
 *
 * Run: node --test .claude/hooks/pre-write-large-file-block.test.js
 */
'use strict';

const { describe, it } = require('node:test');
const assert = require('node:assert/strict');
const { spawnSync } = require('child_process');
const path = require('path');

const hookPath = path.join(__dirname, 'pre-write-large-file-block.js');

function runHook(filePath, lineCount) {
  const payload = JSON.stringify({
    tool_input: { file_path: filePath, content: Array(lineCount).fill('x').join('\n') },
  });
  const result = spawnSync(process.execPath, [hookPath], { input: payload, encoding: 'utf8' });
  return { status: result.status, stdout: result.stdout || '', stderr: result.stderr || '', payload };
}

describe('pre-write-large-file-block', () => {
  it('warns but does not block a write over 800 lines', () => {
    const result = runHook('frontend/src/lib/example.ts', 900);
    assert.equal(result.status, 0);
    assert.match(result.stderr, /800-line guideline \(900 lines\)/);
    assert.equal(result.stdout, result.payload);
  });

  it('stays silent at or under 800 lines', () => {
    const result = runHook('frontend/src/lib/example.ts', 800);
    assert.equal(result.status, 0);
    assert.equal(result.stderr, '');
  });

  it('skips generated files', () => {
    const result = runHook('frontend/src/types/generated/models.ts', 5000);
    assert.equal(result.status, 0);
    assert.equal(result.stderr, '');
  });
});
