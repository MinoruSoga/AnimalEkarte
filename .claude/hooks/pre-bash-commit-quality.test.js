#!/usr/bin/env node
/**
 * Security + regression tests for pre-bash-commit-quality.js
 * Focus: staged-filename shell injection (SEC-CS-F02).
 *
 * Run: node --test .claude/hooks/pre-bash-commit-quality.test.js
 *  or: node .claude/hooks/pre-bash-commit-quality.test.js
 */
'use strict';

const { describe, it, before, after } = require('node:test');
const assert = require('node:assert/strict');
const { spawnSync, execFileSync } = require('child_process');
const fs = require('fs');
const os = require('os');
const path = require('path');

const hookPath = path.join(__dirname, 'pre-bash-commit-quality.js');
const hookSource = fs.readFileSync(hookPath, 'utf8');

/**
 * @param {string} command
 * @param {{ cwd?: string, env?: NodeJS.ProcessEnv }} [opts]
 * @returns {{ status: number|null, stdout: string, stderr: string }}
 */
function runHook(command, opts = {}) {
  const payload = JSON.stringify({ tool_input: { command } });
  const result = spawnSync(process.execPath, [hookPath], {
    input: payload,
    encoding: 'utf8',
    cwd: opts.cwd || process.cwd(),
    env: { ...process.env, ...(opts.env || {}) },
  });
  return {
    status: result.status,
    stdout: result.stdout || '',
    stderr: result.stderr || '',
  };
}

describe('pre-bash-commit-quality source: no shell interpolation of paths', () => {
  it('does not use shell-form execSync for git show with interpolated path', () => {
    // Vulnerable pattern: execSync(`git show ":${file}"`, ...)
    assert.equal(
      /execSync\s*\(\s*[`'"]git\s+show/.test(hookSource),
      false,
      'must not call execSync("git show ...") / execSync(`git show ...`)',
    );
    assert.equal(
      /execSync\s*\(\s*[`'"]git\s+diff\s+--cached/.test(hookSource),
      false,
      'must not call execSync for git diff --cached name listing',
    );
  });

  it('lists staged files via execFileSync argv + NUL split', () => {
    assert.match(
      hookSource,
      /execFileSync\s*\(\s*['"]git['"]\s*,\s*\[[^\]]*--name-only[^\]]*'-z'/,
      'git diff --cached must use execFileSync argv array including -z',
    );
    assert.match(
      hookSource,
      /\.split\s*\(\s*['"]\\0['"]\s*\)/,
      'staged path list must split on NUL (\\0), not newline only',
    );
  });

  it('reads blob content via execFileSync argv (git show :path)', () => {
    // Expect: execFileSync('git', ['show', ':' + file], { ... })
    assert.match(
      hookSource,
      /execFileSync\s*\(\s*['"]git['"]\s*,\s*\[\s*['"]show['"]\s*,\s*['"]:['"]\s*\+\s*file\s*\]/,
      "git show must be argv-based: execFileSync('git', ['show', ':' + file], ...)",
    );
    // Both content-read sites (console.log scan + secrets scan) must use execFileSync
    const showCalls = hookSource.match(
      /execFileSync\s*\(\s*['"]git['"]\s*,\s*\[\s*['"]show['"]\s*,\s*['"]:['"]\s*\+\s*file\s*\]/g,
    );
    assert.ok(showCalls && showCalls.length >= 2, `expected ≥2 argv git-show calls, got ${showCalls ? showCalls.length : 0}`);
  });
});

describe('pre-bash-commit-quality: malicious staged filenames do not execute shell', () => {
  /** @type {string} */
  let tmpDir;
  /** @type {string} */
  let sideEffectPath;

  before(() => {
    tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), 'sec-cs-f02-'));
    // Relative marker only — absolute paths embed '/' and cannot be a single basename.
    sideEffectPath = path.join(tmpDir, 'PWNED');

    execFileSync('git', ['init'], { cwd: tmpDir, encoding: 'utf8' });
    execFileSync('git', ['config', 'user.email', 'test@example.com'], { cwd: tmpDir });
    execFileSync('git', ['config', 'user.name', 'Test'], { cwd: tmpDir });

    // Filenames that would expand / chain if interpolated into a shell string
    // (cwd = tmpDir → touch PWNED would create sideEffectPath).
    const payloads = [
      '$(touch PWNED).js',
      '`touch PWNED`.js',
      '"; touch PWNED; echo ".js',
      'normal-file.js',
    ];

    for (const name of payloads) {
      const full = path.join(tmpDir, name);
      fs.writeFileSync(full, '// safe content\nconst x = 1;\n', 'utf8');
      execFileSync('git', ['add', '--', name], { cwd: tmpDir, encoding: 'utf8' });
    }
  });

  after(() => {
    try {
      fs.rmSync(tmpDir, { recursive: true, force: true });
    } catch {
      // best-effort cleanup
    }
  });

  it('does not create side-effect files when scanning staged malicious names', () => {
    assert.equal(fs.existsSync(sideEffectPath), false, 'precondition: side-effect must not exist yet');

    const { status, stderr } = runHook('git commit -m "chore: test"', {
      cwd: tmpDir,
      env: {
        CLAUDE_PROJECT_DIR: tmpDir,
        SYNC_MIRRORS_DISABLED: '1',
      },
    });

    // Hook should complete (0 warn-allow, or 2 only for secrets/sync — not crash)
    assert.notEqual(status, null, `hook failed to spawn: ${stderr}`);
    assert.ok(status === 0 || status === 2, `unexpected exit ${status}: ${stderr}`);
    assert.equal(
      fs.existsSync(sideEffectPath),
      false,
      'malicious staged filename must not execute shell side effects via git show',
    );
  });

  it('still scans a normal staged .js file (behavior preserved)', () => {
    // Add a console.log in normal-file.js and re-stage
    const normal = path.join(tmpDir, 'normal-file.js');
    fs.writeFileSync(normal, 'console.log("hi");\n', 'utf8');
    execFileSync('git', ['add', '--', 'normal-file.js'], { cwd: tmpDir, encoding: 'utf8' });

    const { status, stderr } = runHook('git commit -m "chore: test"', {
      cwd: tmpDir,
      env: {
        CLAUDE_PROJECT_DIR: tmpDir,
        SYNC_MIRRORS_DISABLED: '1',
      },
    });

    assert.equal(status, 0, `expected warn-allow exit 0, got ${status}`);
    assert.match(stderr, /console\.log/, 'should still warn on console.log in staged JS');
  });
});

// node:test runs describe/it both under `node --test` and direct `node <file>`.
