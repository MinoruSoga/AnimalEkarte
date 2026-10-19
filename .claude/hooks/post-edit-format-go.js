#!/usr/bin/env node
/** Docker-only formatter: unknown target or mount evidence means no formatting.
 * Does not start containers, install dependencies, or apply migrations.
 * Local path checks reduce races; they do not provide OS-level isolation or
 * eliminate replacement between the final check and container execution.
 */
'use strict';

const fs = require('node:fs');
const path = require('node:path');
const { spawnSync } = require('node:child_process');
const MAX_STDIN = 1024 * 1024;

function targetSnapshot(root, filePath) {
  const project = path.resolve(root);
  if (fs.realpathSync(project) !== project) throw new Error('symlinked-project');
  const backend = path.join(project, 'backend');
  const file = path.resolve(project, filePath);
  const relative = path.relative(backend, file);
  if (!relative || relative === '..' || relative.startsWith('..' + path.sep) || path.isAbsolute(relative)) {
    throw new Error('outside-backend');
  }
  const parts = [project, backend];
  let current = backend;
  for (const part of relative.split(path.sep)) {
    current = path.join(current, part);
    parts.push(current);
  }
  const stamps = parts.map((entry, index) => {
    const stat = fs.lstatSync(entry);
    const last = index === parts.length - 1;
    if (stat.isSymbolicLink() || fs.realpathSync(entry) !== entry || (last ? !stat.isFile() || stat.nlink !== 1 : !stat.isDirectory())) {
      throw new Error('unsafe-file-kind');
    }
    return [entry, stat.dev, stat.ino, stat.mode, stat.nlink, stat.size, stat.mtimeMs, stat.ctimeMs];
  });
  return { project, backend, relative, stamps: JSON.stringify(stamps) };
}

function formatEditedGo(input, { root = process.env.CLAUDE_PROJECT_DIR || process.cwd(), run = spawnSync, environment = process.env } = {}) {
  if (!input || !['Edit', 'Write', 'MultiEdit'].includes(input.tool_name)) return { status: 'ignored' };
  const filePath = input.tool_input?.file_path;
  if (typeof filePath !== 'string' || !filePath.endsWith('.go') || /[\x00-\x1f\x7f]/.test(filePath)) return { status: 'ignored' };
  let executionStarted = false;
  try {
    const before = targetSnapshot(root, filePath);
    // Context inspection does not prove the endpoint selected by env overrides.
    if (['DOCKER_HOST', 'DOCKER_CONTEXT', 'DOCKER_TLS_VERIFY', 'DOCKER_CERT_PATH'].some(key => environment[key])) {
      return { status: 'skip', reason: 'unverified-docker-override' };
    }
    let verifiedHost;
    const call = (args, timeout = 5000) => {
      const selected = verifiedHost ? ['--host', verifiedHost, ...args] : args;
      const result = run('docker', selected, { cwd: before.project, encoding: 'utf8', timeout, maxBuffer: MAX_STDIN, stdio: ['pipe', 'pipe', 'pipe'] });
      if (result.error || result.status !== 0 || typeof result.stdout !== 'string') throw new Error('docker-evidence-or-execution-failed');
      return result.stdout;
    };
    const contexts = JSON.parse(call(['context', 'inspect']));
    if (!Array.isArray(contexts) || contexts.length !== 1 || !/^unix:\/\/\//.test(contexts[0]?.Endpoints?.docker?.Host || '')) {
      return { status: 'skip', reason: 'unverified-local-docker-endpoint' };
    }
    // Pin every subsequent operation even if another session switches context.
    verifiedHost = contexts[0].Endpoints.docker.Host;
    const compose = ['compose', '--project-directory', before.project];
    const envFile = path.join(before.project, '.env.local');
    if (fs.existsSync(envFile)) {
      if (!fs.lstatSync(envFile).isFile() || fs.realpathSync(envFile) !== envFile) return { status: 'skip', reason: 'unsafe-compose-env-path' };
      compose.push('--env-file', envFile);
    }
    const ids = call([...compose, 'ps', '-q', 'backend']).trim().split(/\s+/);
    if (ids.length !== 1 || !/^[a-f0-9]{64}$/.test(ids[0])) return { status: 'skip', reason: 'unknown-or-multiple-containers' };
    const id = ids[0];
    const inspected = JSON.parse(call(['inspect', id]));
    if (!Array.isArray(inspected) || inspected.length !== 1 || inspected[0]?.Id !== id || inspected[0]?.State?.Running !== true) {
      return { status: 'skip', reason: 'unknown-or-stopped-container' };
    }
    const mounts = inspected[0].Mounts;
    if (!Array.isArray(mounts)) return { status: 'skip', reason: 'unknown-mounts' };
    const app = mounts.filter(mount => mount?.Destination === '/app');
    if (app.length !== 1 || app[0].Type !== 'bind' || app[0].RW !== true || app[0].Source !== before.backend || fs.realpathSync(app[0].Source) !== before.backend || mounts.some(mount => typeof mount?.Destination !== 'string' || mount.Destination.startsWith('/app/'))) {
      return { status: 'skip', reason: 'unverified-backend-bind-mount' };
    }
    if (targetSnapshot(root, filePath).stamps !== before.stamps) return { status: 'skip', reason: 'target-changed-during-verification' };
    const destination = '/app/' + before.relative.split(path.sep).join('/');
    executionStarted = true;
    call(['exec', '-i', id, 'gofmt', '-w', destination], 15000);
    return { status: 'formatted' };
  } catch {
    // Never echo Docker output, file paths, input, or credentials into a hook log.
    return executionStarted
      ? { status: 'unknown', reason: 'formatter-result-unknown-no-retry' }
      : { status: 'skip', reason: 'target-or-docker-evidence-unavailable' };
  }
}

if (require.main === module) {
  const chunks = [];
  let bytes = 0;
  let oversized = false;
  process.stdin.on('data', chunk => {
    bytes += chunk.length;
    if (bytes > MAX_STDIN) oversized = true;
    if (!oversized) chunks.push(chunk);
  });
  process.stdin.on('end', () => {
    if (oversized) {
      process.stderr.write('[Hook] Go format SKIP: input-limit.\n');
      return;
    }
    let result;
    try {
      const raw = new TextDecoder('utf-8', { fatal: true }).decode(Buffer.concat(chunks));
      result = formatEditedGo(JSON.parse(raw));
    } catch { result = { status: 'skip', reason: 'invalid-input' }; }
    if (result.status === 'skip') process.stderr.write(`[Hook] Go format SKIP: ${result.reason}.\n`);
    if (result.status === 'unknown') process.stderr.write(`[Hook] Go format UNKNOWN: ${result.reason}; inspect the owned file diff before any retry.\n`);
  });
}

module.exports = { formatEditedGo };
