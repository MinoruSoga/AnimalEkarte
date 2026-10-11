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
 * @returns {{ status: number|null, stdout: string, stderr: string, error: Error|undefined }}
 */
function runHook(command) {
  const payload = JSON.stringify({ tool_input: { command } });
  const result = spawnSync(process.execPath, [hook], {
    input: payload,
    encoding: 'utf8',
    timeout: 15000,
  });
  return {
    status: result.status,
    stdout: result.stdout || '',
    stderr: result.stderr || '',
    error: result.error,
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
  const { status, stdout, stderr, error } = runHook(c.command);
  const expected = c.expectBlock ? 2 : 0;
  if (error !== undefined || status !== expected) {
    failed += 1;
    console.error(`FAIL ${c.name}: expected status=${expected}, got status=${status} error=${error}`);
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
  const r = runPayload(JSON.stringify({ toolInput: { command: 'git reset --hard' } }));
  if (r.error !== undefined || r.status !== 2) {
    failed += 1;
    console.error(`FAIL grok toolInput camelCase: expected status=2, got status=${r.status} error=${r.error}`);
  } else {
    console.log('ok  grok toolInput camelCase');
  }
}

function runPayload(input) {
  const result = spawnSync(process.execPath, [hook], {
    input,
    encoding: 'utf8',
    timeout: 15000,
  });
  return {
    status: result.status,
    stdout: result.stdout || '',
    stderr: result.stderr || '',
    error: result.error,
  };
}

const SENTINEL = 'SENTINEL-9f4ac2e7-SECRET';

function payloadBytes(command, totalBytes) {
  const prefix = '{"tool_input":{"command":"';
  const suffix = '"}}';
  const pad = totalBytes - Buffer.byteLength(prefix) - Buffer.byteLength(suffix) - Buffer.byteLength(command);
  return prefix + command + 'x'.repeat(pad) + suffix;
}

const strictCases = [
  { name: 'empty stdin', input: '', expect: 2 },
  { name: 'whitespace only', input: ' \n\t\r\n ', expect: 2 },
  { name: 'truncated json', input: '{"tool_input":{"command":"git status"', expect: 2 },
  { name: 'invalid json', input: '{invalid', expect: 2 },
  { name: 'root null', input: 'null', expect: 2 },
  { name: 'root array', input: '[]', expect: 2 },
  { name: 'root number', input: '42', expect: 2 },
  { name: 'root string', input: '"git status"', expect: 2 },
  { name: 'unnamed empty object', input: '{}', expect: 2 },
  { name: 'unnamed container without command', input: '{"tool_input":{}}', expect: 2 },
  { name: 'target without command', input: '{"tool_name":"Bash","tool_input":{}}', expect: 2 },
  { name: 'empty command', input: '{"tool_input":{"command":""}}', expect: 2 },
  { name: 'whitespace command', input: '{"tool_input":{"command":"   "}}', expect: 2 },
  { name: 'whitespace tool name', input: '{"tool_name":"  ","tool_input":{"command":"git status"}}', expect: 2 },
  { name: 'non-string command', input: '{"tool_input":{"command":42}}', expect: 2 },
  { name: 'container array', input: '{"tool_input":[]}', expect: 2 },
  { name: 'container string', input: '{"tool_input":"x"}', expect: 2 },
  { name: 'container null', input: '{"tool_input":null}', expect: 2 },
  { name: 'parameters wrong type', input: '{"parameters":1}', expect: 2 },
  { name: 'conflicting command and cmd', input: '{"tool_input":{"command":"git status","cmd":"git diff"}}', expect: 2 },
  { name: 'conflicting container and top-level command', input: '{"tool_input":{"command":"git status"},"command":"git diff"}', expect: 2 },
  { name: 'conflicting tool names', input: '{"tool_name":"Bash","toolName":"Read","tool_input":{"command":"git status"}}', expect: 2 },
  { name: 'empty tool name', input: '{"tool_name":"","tool_input":{"command":"git status"}}', expect: 2 },
  { name: 'non-string tool name', input: '{"tool_name":1,"tool_input":{"command":"git status"}}', expect: 2 },
  { name: 'claude bash envelope', input: '{"tool_name":"Bash","tool_input":{"command":"git reset --hard"}}', expect: 2 },
  { name: 'grok envelope', input: '{"toolName":"Bash","toolInput":{"command":"rm -rf /"}}', expect: 2 },
  { name: 'parameters envelope', input: '{"parameters":{"command":"git reset --hard"}}', expect: 2 },
  { name: 'top-level command', input: '{"command":"git clean -fd"}', expect: 2 },
  { name: 'cmd alias', input: '{"tool_input":{"cmd":"git reset --hard"}}', expect: 2 },
  { name: 'lowercase bash target', input: '{"tool_name":"bash","tool_input":{"command":"git reset --hard"}}', expect: 2 },
  { name: 'lowercase shell target', input: '{"tool_name":"shell","tool_input":{"command":"git reset --hard"}}', expect: 2 },
  { name: 'Shell target', input: '{"tool_name":"Shell","tool_input":{"command":"git reset --hard"}}', expect: 2 },
  { name: 'run_terminal_command target', input: '{"tool_name":"run_terminal_command","tool_input":{"command":"git push --force"}}', expect: 2 },
  { name: 'non-target without command', input: '{"tool_name":"Read","tool_input":{"file_path":"/tmp/fixture"}}', expect: 0 },
  { name: 'non-target without input container', input: '{"tool_name":"Read"}', expect: 2 },
  { name: 'non-target destructive command denied', input: '{"tool_name":"Read","tool_input":{"command":"git reset --hard"}}', expect: 2 },
  { name: 'non-target benign command allowed', input: '{"tool_name":"Read","tool_input":{"command":"git status"}}', expect: 0 },
  { name: 'benign toolInput envelope', input: '{"toolInput":{"command":"git status"}}', expect: 0 },
  { name: 'benign cmd alias', input: '{"tool_input":{"cmd":"git status"}}', expect: 0 },
  { name: 'benign parameters envelope', input: '{"parameters":{"command":"git status"}}', expect: 0 },
  { name: 'benign top-level command', input: '{"command":"git status"}', expect: 0 },
  { name: 'benign same-value aliases', input: '{"tool_name":"Bash","toolName":"Bash","tool_input":{"command":"git status","cmd":"git status"},"command":"git status"}', expect: 0 },
  { name: 'deny hides command sentinel', input: JSON.stringify({ tool_name: 'Bash', tool_input: { command: `git reset --hard ${SENTINEL}` } }), expect: 2, absent: SENTINEL },
  { name: 'malformed deny hides sentinel', input: `{"tool_input":{"command":"${SENTINEL}`, expect: 2, absent: SENTINEL },
  { name: 'success hides command sentinel', input: JSON.stringify({ tool_name: 'Bash', tool_input: { command: `git status ${SENTINEL}` } }), expect: 0, absent: SENTINEL },
  { name: 'exact byte limit', input: payloadBytes('git status ', 1024 * 1024), expect: 0 },
  { name: 'byte limit plus one', input: payloadBytes('git status ', 1024 * 1024 + 1), expect: 2 },
  { name: 'exact byte limit multibyte', input: payloadBytes('git status あ', 1024 * 1024), expect: 0 },
  { name: 'byte limit plus one multibyte', input: payloadBytes('git status あ', 1024 * 1024 + 1), expect: 2 },
  { name: 'valid json plus oversized whitespace', input: '{"tool_input":{"command":"git status"}}' + ' '.repeat(1024 * 1024), expect: 2 },
  { name: 'invalid utf8', input: Buffer.concat([Buffer.from('{"tool_input":{"command":"git status '), Buffer.from([0xff, 0xfe]), Buffer.from('"}}')]), expect: 2 },
];

for (const c of strictCases) {
  const r = runPayload(c.input);
  if (r.error !== undefined || r.status !== c.expect) {
    failed += 1;
    console.error(`FAIL ${c.name}: expected status=${c.expect}, got status=${r.status} error=${r.error}`);
    console.error(`  stdout=${r.stdout.slice(0, 200)}`);
    console.error(`  stderr=${r.stderr.slice(0, 200)}`);
    continue;
  }
  if (c.expect === 2) {
    try {
      const body = JSON.parse(r.stdout);
      if (body.decision !== 'deny' || body.permissionDecision !== 'deny') {
        throw new Error('missing deny envelope fields');
      }
    } catch (err) {
      failed += 1;
      console.error(`FAIL ${c.name}: deny envelope invalid: ${err}`);
      continue;
    }
  }
  if (c.expect === 0 && (r.stdout !== '' || r.stderr !== '')) {
    failed += 1;
    console.error(`FAIL ${c.name}: success must be silent, got stdout=${r.stdout.length}B stderr=${r.stderr.length}B`);
    continue;
  }
  if (c.absent && (r.stdout.includes(c.absent) || r.stderr.includes(c.absent))) {
    failed += 1;
    console.error(`FAIL ${c.name}: output echoed input content`);
    continue;
  }
  console.log(`ok  ${c.name}`);
}

if (failed > 0) {
  console.error(`\n${failed} failure(s)`);
  process.exit(1);
}
console.log(`\nAll ${cases.length + strictCases.length + 1} checks passed`);
