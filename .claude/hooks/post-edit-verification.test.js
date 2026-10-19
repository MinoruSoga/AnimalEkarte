'use strict';

const { test } = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const { spawnSync } = require('node:child_process');
const { formatEditedGo } = require('./post-edit-format-go.js');

const id = 'a'.repeat(64);
const host = 'unix:///mock/docker.sock';
const operation = args => args[0] === '--host' ? args[2] : args[0];
function fixture(t) {
  const base = fs.mkdtempSync(path.join(fs.realpathSync(os.tmpdir()), 'animal-hook-'));
  t.after(() => fs.rmSync(base, { recursive: true, force: true }));
  const root = path.join(base, 'candidate');
  const backend = path.join(root, 'backend');
  const file = path.join(backend, 'internal', 'example.go');
  fs.mkdirSync(path.dirname(file), { recursive: true });
  fs.writeFileSync(file, 'package example\n');
  const calls = [];
  const mount = { Type: 'bind', Source: backend, Destination: '/app', RW: true };
  const state = { context: 'unix:///mock/docker.sock', ids: id + '\n', mounts: [mount], running: true };
  function run(command, args, options) {
    calls.push({ command, args, options });
    if (args[0] === '--host') { assert.equal(args[1], host); args = args.slice(2); }
    if (args[0] === 'context') return { status: 0, stdout: JSON.stringify([{ Endpoints: { docker: { Host: state.context } } }]) };
    if (args[0] === 'compose') return { status: 0, stdout: state.ids };
    if (operation(args) === 'inspect') return { status: 0, stdout: JSON.stringify([{ Id: id, State: { Running: state.running }, Mounts: state.mounts }]) };
    assert.deepEqual(args, ['exec', '-i', id, 'gofmt', '-w', '/app/internal/example.go']);
    assert.equal(command, 'docker');
    // Mock only: prove the exact fixture is the sole allowed write target.
    fs.writeFileSync(file, 'package example // formatted\n');
    return { status: 0, stdout: '' };
  }
  const input = { tool_name: 'Edit', tool_input: { file_path: file } };
  const invoke = (extra = {}) => formatEditedGo(input, { root, run, environment: {}, ...extra });
  return { base, root, backend, file, calls, state, mount, input, invoke, run };
}

test('formats only the validated exact container and candidate mount', t => {
  const f = fixture(t);
  fs.writeFileSync(path.join(f.root, '.env.local'), 'FIXTURE_ONLY=1\n');
  assert.equal(f.invoke().status, 'formatted');
  assert.match(fs.readFileSync(f.file, 'utf8'), /formatted/);
  assert.deepEqual(f.calls[1].args, ['--host', host, 'compose', '--project-directory', f.root, '--env-file', path.join(f.root, '.env.local'), 'ps', '-q', 'backend']);
  assert.equal(f.calls.at(-1).options.cwd, f.root);
  assert.ok(f.calls.slice(1).every(c => c.args[0] === '--host' && c.args[1] === host));
});

test('normal target without env file uses explicit project directory', t => {
  const f = fixture(t);
  assert.equal(f.invoke().status, 'formatted');
  assert.deepEqual(f.calls[1].args, ['--host', host, 'compose', '--project-directory', f.root, 'ps', '-q', 'backend']);
});

test('later context changes cannot change the verified endpoint', t => {
  const f = fixture(t);
  const run = (command, args, options) => {
    const result = f.run(command, args, options);
    if (operation(args) === 'context') f.state.context = 'ssh://other-daemon';
    return result;
  };
  assert.equal(f.invoke({ run }).status, 'formatted');
  assert.ok(f.calls.slice(1).every(c => c.args[0] === '--host' && c.args[1] === host));
});

for (const kind of ['foreign-worktree', 'outside', 'prefix-sibling']) {
  test(`rejects ${kind} without invoking Docker or touching either tree`, t => {
    const f = fixture(t);
    const dir = kind === 'prefix-sibling' ? f.backend + '-other' : path.join(f.base, kind, 'backend');
    fs.mkdirSync(dir, { recursive: true });
    const foreign = path.join(dir, 'example.go');
    fs.writeFileSync(foreign, 'package untouched\n');
    f.input.tool_input.file_path = foreign;
    assert.equal(f.invoke().status, 'skip');
    assert.equal(f.calls.length, 0);
    assert.equal(fs.readFileSync(foreign, 'utf8'), 'package untouched\n');
    assert.equal(fs.readFileSync(f.file, 'utf8'), 'package example\n');
  });
}

for (const kind of ['file', 'directory', 'backend-root', 'project-root']) {
  test(`rejects ${kind} symlink without invoking Docker`, t => {
    const f = fixture(t);
    let extra = {};
    if (kind === 'file') {
      const link = path.join(f.backend, 'link.go'); fs.symlinkSync(f.file, link); f.input.tool_input.file_path = link;
    } else if (kind === 'directory') {
      const link = path.join(f.backend, 'link'); fs.symlinkSync(path.dirname(f.file), link); f.input.tool_input.file_path = path.join(link, 'example.go');
    } else if (kind === 'project-root') {
      const link = path.join(f.base, 'alias'); fs.symlinkSync(f.root, link); extra = { root: link };
    } else {
      const root = path.join(f.base, 'other'); fs.mkdirSync(root); fs.symlinkSync(f.backend, path.join(root, 'backend')); extra = { root };
      f.input.tool_input.file_path = path.join(root, 'backend', 'internal', 'example.go');
    }
    assert.equal(f.invoke(extra).status, 'skip');
    assert.equal(f.calls.length, 0);
  });
}

for (const kind of ['missing', 'mismatch', 'read-only', 'nested', 'volume', 'stopped', 'multiple', 'empty', 'remote-context']) {
  test(`does not format with ${kind} container/mount evidence`, t => {
    const f = fixture(t);
    if (kind === 'missing') f.state.mounts = [];
    if (kind === 'mismatch') f.mount.Source = f.base;
    if (kind === 'read-only') f.mount.RW = false;
    if (kind === 'nested') f.state.mounts.push({ Destination: '/app/internal', Type: 'volume', RW: true });
    if (kind === 'volume') f.mount.Type = 'volume';
    if (kind === 'stopped') f.state.running = false;
    if (kind === 'multiple') f.state.ids += 'b'.repeat(64) + '\n';
    if (kind === 'empty') f.state.ids = '';
    if (kind === 'remote-context') f.state.context = 'ssh://mock-remote';
    assert.equal(f.invoke().status, 'skip');
    assert.ok(f.calls.every(c => operation(c.args) !== 'exec'));
    assert.equal(fs.readFileSync(f.file, 'utf8'), 'package example\n');
  });
}

for (const environment of [{ DOCKER_HOST: 'unix:///other.sock' }, { DOCKER_CONTEXT: 'other' }]) {
  test(`skips unverified Docker selection override ${Object.keys(environment)[0]}`, t => {
    const f = fixture(t);
    assert.equal(f.invoke({ environment }).status, 'skip');
    assert.equal(f.calls.length, 0);
  });
}

for (const result of [{ status: 1, stderr: 'PRIVATE_MOCK_ERROR' }, { status: null, error: new Error('timeout') }, { status: 0, stdout: 'not-json' }]) {
  test('unknown or failed Docker evidence cannot reach formatting', t => {
    const f = fixture(t);
    assert.equal(f.invoke({ run: () => result }).status, 'skip');
    assert.equal(fs.readFileSync(f.file, 'utf8'), 'package example\n');
  });
}

test('ignores non-edit events, malformed paths and non-Go files', t => {
  const f = fixture(t);
  for (const input of [{}, { tool_name: 'Read', tool_input: { file_path: f.file } }, { tool_name: 'Write', tool_input: { file_path: [] } }, { tool_name: 'Edit', tool_input: { file_path: f.file + '.txt' } }]) {
    assert.notEqual(formatEditedGo(input, { root: f.root, run: f.run, environment: {} }).status, 'formatted');
  }
  assert.equal(f.calls.length, 0);
});

test('rejects hard links without changing either alias', t => {
  const f = fixture(t);
  const alias = path.join(f.base, 'shared.go');
  fs.linkSync(f.file, alias);
  assert.equal(f.invoke().status, 'skip');
  assert.equal(f.calls.length, 0);
  assert.equal(fs.readFileSync(alias, 'utf8'), 'package example\n');
});

test('detects target replacement during Docker inspection', t => {
  const f = fixture(t);
  const run = (command, args, options) => {
    const result = f.run(command, args, options);
    if (operation(args) === 'inspect') {
      const replacement = f.file + '.replacement';
      fs.writeFileSync(replacement, 'package replacement\n');
      fs.renameSync(replacement, f.file);
    }
    return result;
  };
  assert.equal(f.invoke({ run }).reason, 'target-changed-during-verification');
  assert.ok(f.calls.every(c => operation(c.args) !== 'exec'));
  assert.equal(fs.readFileSync(f.file, 'utf8'), 'package replacement\n');
});

test('malformed inspect or identity mismatch cannot reach exec', t => {
  const f = fixture(t);
  for (const stdout of ['not-json', '[]', '{}', JSON.stringify([{ Id: 'b'.repeat(64), State: { Running: true }, Mounts: [f.mount] }])]) {
    const run = (command, args, options) => operation(args) === 'inspect' ? { status: 0, stdout } : f.run(command, args, options);
    assert.equal(f.invoke({ run }).status, 'skip');
  }
  assert.ok(f.calls.every(c => operation(c.args) !== 'exec'));
});

test('formatter failure reports unknown outcome without retry or private errors', t => {
  const f = fixture(t);
  const run = (command, args, options) => operation(args) === 'exec' ? { status: 1, stderr: 'PRIVATE_MOCK_ERROR' } : f.run(command, args, options);
  assert.deepEqual(f.invoke({ run }), { status: 'unknown', reason: 'formatter-result-unknown-no-retry' });
});

for (const name of ['post-edit-format-go.js', 'post-edit-typecheck-ts.js']) {
  test(`${name} rejects invalid UTF-8 without repeating input`, () => {
    const res = spawnSync(process.execPath, [path.join(__dirname, name)], { input: Buffer.from([0xff, 0x7b]), encoding: 'utf8', timeout: 5000 });
    assert.equal(res.status, 0);
    assert.equal(res.stdout, '');
    assert.match(res.stderr, /SKIP: invalid-input/);
  });
  test(`${name} rejects oversized input without running a child process`, t => {
    const f = fixture(t);
    const marker = path.join(f.base, 'process-was-called');
    const preload = path.join(f.base, 'no-process.cjs');
    fs.writeFileSync(preload, `const cp=require('node:child_process');for(const k of ['exec','execSync','execFile','execFileSync','spawn','spawnSync','fork'])cp[k]=()=>{require('node:fs').writeFileSync(${JSON.stringify(marker)},'called');throw Error('forbidden')};\n`);
    const res = spawnSync(process.execPath, ['--require', preload, path.join(__dirname, name)], { input: 'x'.repeat(1024 * 1024 + 1), encoding: 'utf8', maxBuffer: 2 * 1024 * 1024, timeout: 5000 });
    assert.equal(res.status, 0);
    assert.equal(res.stdout, '');
    assert.match(res.stderr, /SKIP: input-limit/);
    assert.equal(fs.existsSync(marker), false);
  });
}

test('Go CLI does not echo payload and skips foreign paths without execution', t => {
  const f = fixture(t);
  const marker = path.join(f.base, 'process-was-called');
  const preload = path.join(f.base, 'no-process.cjs');
  fs.writeFileSync(preload, `require('node:child_process').spawnSync=()=>{require('node:fs').writeFileSync(${JSON.stringify(marker)},'called');throw Error('forbidden')};\n`);
  const payload = JSON.stringify({ tool_name: 'Edit', tool_input: { file_path: path.join(f.base, 'foreign/backend/example.go') } });
  const res = spawnSync(process.execPath, ['--require', preload, path.join(__dirname, 'post-edit-format-go.js')], { input: payload, encoding: 'utf8', timeout: 5000, env: { ...process.env, CLAUDE_PROJECT_DIR: f.root } });
  assert.equal(res.status, 0);
  assert.equal(res.stdout, '');
  assert.match(res.stderr, /SKIP/);
  assert.equal(fs.existsSync(marker), false);
});

test('TS hook only advises scoped verification; never launches a process', t => {
  const f = fixture(t);
  const preload = path.join(f.base, 'no-process.cjs');
  fs.writeFileSync(preload, "const cp=require('node:child_process');for(const k of ['exec','execSync','execFile','execFileSync','spawn','spawnSync','fork'])cp[k]=()=>{throw Error('PROCESS_EXECUTION_FORBIDDEN')};\n");
  const payload = JSON.stringify({ tool_name: 'Edit', tool_input: { file_path: path.join(f.root, 'frontend/src/example.ts') } });
  const res = spawnSync(process.execPath, ['--require', preload, path.join(__dirname, 'post-edit-typecheck-ts.js')], { input: payload, encoding: 'utf8' });
  assert.equal(res.status, 0);
  assert.equal(res.stdout, '');
  assert.match(res.stderr, /SKIP.*scoped verification/);
  assert.doesNotMatch(res.stderr, /PROCESS_EXECUTION_FORBIDDEN|型エラーなし|PASS/);
});
