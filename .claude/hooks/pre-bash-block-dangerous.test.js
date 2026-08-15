#!/usr/bin/env node
/**
 * Unit checks for pre-bash-block-dangerous.js (no test runner required).
 * Run: node .claude/hooks/pre-bash-block-dangerous.test.js
 */
'use strict';

const { spawnSync } = require('child_process');
const path = require('path');

const hook = path.join(__dirname, 'pre-bash-block-dangerous.js');

/**
 * @param {string} command
 * @returns {{ status: number|null, stdout: string, stderr: string }}
 */
function runHook(command) {
  const payload = JSON.stringify({ tool_input: { command } });
  const result = spawnSync(process.execPath, [hook], {
    input: payload,
    encoding: 'utf8',
  });
  return {
    status: result.status,
    stdout: result.stdout || '',
    stderr: result.stderr || '',
  };
}

/** @type {Array<{ name: string, command: string, expectBlock: boolean }>} */
const cases = [
  { name: 'reset --hard HEAD', command: 'git reset --hard HEAD', expectBlock: true },
  { name: 'reset --hard origin/main', command: 'git reset --hard origin/main', expectBlock: true },
  { name: 'compound with reset --hard', command: 'git fetch origin main && git reset --hard origin/main', expectBlock: true },
  { name: 'clean -fd', command: 'git clean -fd', expectBlock: true },
  { name: 'clean -fdx', command: 'git clean -fdx', expectBlock: true },
  { name: 'checkout -- .', command: 'git checkout -- .', expectBlock: true },
  { name: 'restore .', command: 'git restore .', expectBlock: true },
  { name: 'push --force', command: 'git push --force origin main', expectBlock: true },
  { name: 'push -f', command: 'git push -f origin main', expectBlock: true },
  { name: 'push --force-with-lease', command: 'git push --force-with-lease', expectBlock: true },
  { name: 'rm -rf /', command: 'rm -rf /', expectBlock: true },
  // allowed
  { name: 'status', command: 'git status -sb', expectBlock: false },
  { name: 'diff', command: 'git diff', expectBlock: false },
  { name: 'checkout branch', command: 'git checkout main', expectBlock: false },
  { name: 'switch branch', command: 'git switch -c feature/x', expectBlock: false },
  { name: 'restore single path', command: 'git restore backend/go.mod', expectBlock: false },
  { name: 'reset soft (allowed)', command: 'git reset --soft HEAD~1', expectBlock: false },
  { name: 'reset mixed path', command: 'git reset HEAD -- file.go', expectBlock: false },
  { name: 'pull ff-only', command: 'git pull --ff-only', expectBlock: false },
  { name: 'merge', command: 'git merge origin/main', expectBlock: false },
];

let failed = 0;
for (const c of cases) {
  const { status, stdout, stderr } = runHook(c.command);
  const blocked = status === 2;
  const ok = blocked === c.expectBlock;
  if (!ok) {
    failed += 1;
    console.error(`FAIL ${c.name}: expected block=${c.expectBlock}, got status=${status}`);
    console.error(`  stdout=${stdout.slice(0, 200)}`);
    console.error(`  stderr=${stderr.slice(0, 200)}`);
    continue;
  }
  if (c.expectBlock) {
    try {
      const body = JSON.parse(stdout);
      if (body.decision !== 'deny') {
        failed += 1;
        console.error(`FAIL ${c.name}: deny JSON missing decision=deny`);
      }
    } catch (err) {
      failed += 1;
      console.error(`FAIL ${c.name}: deny JSON parse error: ${err}`);
    }
  }
  console.log(`ok  ${c.name}`);
}

// Grok camelCase toolInput
{
  const payload = JSON.stringify({ toolInput: { command: 'git reset --hard' } });
  const result = spawnSync(process.execPath, [hook], { input: payload, encoding: 'utf8' });
  if (result.status !== 2) {
    failed += 1;
    console.error('FAIL grok toolInput camelCase: expected block');
  } else {
    console.log('ok  grok toolInput camelCase');
  }
}

if (failed > 0) {
  console.error(`\n${failed} failure(s)`);
  process.exit(1);
}
console.log(`\nAll ${cases.length + 1} checks passed`);
