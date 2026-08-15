#!/usr/bin/env node
/**
 * Fixture tests for:
 *   - scripts/check-agent-security-policy.sh
 *   - scripts/check-workflow-remote-exec-policy.sh
 *
 * Negative fixtures must fail (exit 1); safe fixtures must pass (exit 0).
 * Docker-free; pure filesystem + bash.
 *
 * Usage: node --test scripts/check-agent-security-policy.test.mjs
 */
import assert from "node:assert/strict";
import { spawnSync } from "node:child_process";
import { mkdtempSync, mkdirSync, writeFileSync, rmSync, chmodSync } from "node:fs";
import { tmpdir } from "node:os";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";
import test from "node:test";

const HERE = dirname(fileURLToPath(import.meta.url));
const AGENT_CHECK = join(HERE, "check-agent-security-policy.sh");
const REMOTE_CHECK = join(HERE, "check-workflow-remote-exec-policy.sh");

function runCheck(script, root) {
  const r = spawnSync("bash", [script, root], {
    encoding: "utf8",
    env: process.env,
  });
  return {
    status: r.status ?? 1,
    stdout: r.stdout ?? "",
    stderr: r.stderr ?? "",
  };
}

function withTempRoot(populate) {
  const root = mkdtempSync(join(tmpdir(), "sec-policy-"));
  try {
    populate(root);
    return root;
  } catch (e) {
    rmSync(root, { recursive: true, force: true });
    throw e;
  }
}

function cleanup(root) {
  rmSync(root, { recursive: true, force: true });
}

// ── agent security policy ──────────────────────────────────────────

test("agent: safe permissions + deny network → PASS", () => {
  const root = withTempRoot((r) => {
    mkdirSync(join(r, ".cursor"), { recursive: true });
    writeFileSync(
      join(r, ".cursor", "permissions.json"),
      JSON.stringify({
        approvalMode: "allowlist",
        mcpAllowlist: ["chrome-devtools"],
        terminalAllowlist: ["git", "bash"],
      }),
    );
    writeFileSync(
      join(r, ".cursor", "sandbox.json"),
      JSON.stringify({
        type: "workspace_readwrite",
        networkPolicy: { default: "deny" },
      }),
    );
  });
  try {
    const { status, stdout } = runCheck(AGENT_CHECK, root);
    assert.equal(status, 0, stdout);
    assert.match(stdout, /PASS/);
  } finally {
    cleanup(root);
  }
});

test("agent: unrestricted approvalMode → FAIL", () => {
  const root = withTempRoot((r) => {
    mkdirSync(join(r, ".cursor"), { recursive: true });
    writeFileSync(
      join(r, ".cursor", "permissions.json"),
      JSON.stringify({
        approvalMode: "unrestricted",
        mcpAllowlist: ["chrome-devtools"],
      }),
    );
    writeFileSync(
      join(r, ".cursor", "sandbox.json"),
      JSON.stringify({ networkPolicy: { default: "deny" } }),
    );
  });
  try {
    const { status, stdout } = runCheck(AGENT_CHECK, root);
    assert.equal(status, 1, stdout);
    assert.match(stdout, /unrestricted/i);
  } finally {
    cleanup(root);
  }
});

test("agent: mcpAllowlist *:* → FAIL", () => {
  const root = withTempRoot((r) => {
    mkdirSync(join(r, ".cursor"), { recursive: true });
    writeFileSync(
      join(r, ".cursor", "permissions.json"),
      JSON.stringify({
        approvalMode: "allowlist",
        mcpAllowlist: ["*:*"],
      }),
    );
    writeFileSync(
      join(r, ".cursor", "sandbox.json"),
      JSON.stringify({ networkPolicy: { default: "deny" } }),
    );
  });
  try {
    const { status, stdout } = runCheck(AGENT_CHECK, root);
    assert.equal(status, 1, stdout);
    assert.match(stdout, /\*:\*/);
  } finally {
    cleanup(root);
  }
});

test("agent: networkPolicy.default allow → FAIL", () => {
  const root = withTempRoot((r) => {
    mkdirSync(join(r, ".cursor"), { recursive: true });
    writeFileSync(
      join(r, ".cursor", "permissions.json"),
      JSON.stringify({
        approvalMode: "allowlist",
        mcpAllowlist: [],
      }),
    );
    writeFileSync(
      join(r, ".cursor", "sandbox.json"),
      JSON.stringify({ networkPolicy: { default: "allow" } }),
    );
  });
  try {
    const { status, stdout } = runCheck(AGENT_CHECK, root);
    assert.equal(status, 1, stdout);
    assert.match(stdout, /networkPolicy\.default is allow/i);
  } finally {
    cleanup(root);
  }
});

test("agent: JSONC comments with unrestricted still → FAIL", () => {
  const root = withTempRoot((r) => {
    mkdirSync(join(r, ".cursor"), { recursive: true });
    writeFileSync(
      join(r, ".cursor", "permissions.json"),
      `{
  // comment
  "approvalMode": "unrestricted",
  "mcpAllowlist": []
}
`,
    );
    writeFileSync(
      join(r, ".cursor", "sandbox.json"),
      `{ "networkPolicy": { "default": "deny" } }`,
    );
  });
  try {
    const { status, stdout } = runCheck(AGENT_CHECK, root);
    assert.equal(status, 1, stdout);
  } finally {
    cleanup(root);
  }
});

// ── remote exec policy ─────────────────────────────────────────────

function seedRemoteRoot(populate) {
  return withTempRoot((r) => {
    mkdirSync(join(r, ".github", "workflows"), { recursive: true });
    // Minimal safe workflow so empty-target failure does not mask cases
    writeFileSync(
      join(r, ".github", "workflows", "noop.yml"),
      "name: noop\non: workflow_dispatch\njobs:\n  a:\n    runs-on: ubuntu-latest\n    steps:\n      - run: echo ok\n",
    );
    populate(r);
  });
}

test("remote-exec: safe Dockerfile + workflow → PASS", () => {
  const root = seedRemoteRoot((r) => {
    writeFileSync(
      join(r, "Dockerfile"),
      "FROM alpine:3.20\nRUN apk add --no-cache curl\n",
    );
  });
  try {
    const { status, stdout } = runCheck(REMOTE_CHECK, root);
    assert.equal(status, 0, stdout);
    assert.match(stdout, /PASS/);
  } finally {
    cleanup(root);
  }
});

test("remote-exec: curl|sh in Dockerfile → FAIL", () => {
  const root = seedRemoteRoot((r) => {
    writeFileSync(
      join(r, "Dockerfile.dev"),
      `FROM alpine
RUN curl -sSfL https://example.com/install.sh | sh
`,
    );
  });
  try {
    const { status, stdout } = runCheck(REMOTE_CHECK, root);
    assert.equal(status, 1, stdout);
    assert.match(stdout, /pipe-to-shell|curl\|sh/i);
  } finally {
    cleanup(root);
  }
});

test("remote-exec: multiline curl | sh continuation → FAIL", () => {
  const root = seedRemoteRoot((r) => {
    writeFileSync(
      join(r, "Dockerfile"),
      `FROM alpine
RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | \\
    sh -s -- -b /usr/local/bin v2.11.4
`,
    );
  });
  try {
    const { status, stdout } = runCheck(REMOTE_CHECK, root);
    assert.equal(status, 1, stdout);
  } finally {
    cleanup(root);
  }
});

test("remote-exec: bash <(curl) in workflow → FAIL", () => {
  const root = seedRemoteRoot((r) => {
    writeFileSync(
      join(r, ".github", "workflows", "bad.yml"),
      `name: bad
on: push
jobs:
  a:
    runs-on: ubuntu-latest
    steps:
      - run: bash <(curl -Ls https://raw.githubusercontent.com/rhysd/actionlint/v1.7.12/scripts/download-actionlint.bash) 1.7.12
`,
    );
  });
  try {
    const { status, stdout } = runCheck(REMOTE_CHECK, root);
    assert.equal(status, 1, stdout);
    assert.match(stdout, /process-substitution|bash <\(curl\)/i);
  } finally {
    cleanup(root);
  }
});

test("remote-exec: pinned tarball download with checksum is allowed → PASS", () => {
  const root = seedRemoteRoot((r) => {
    writeFileSync(
      join(r, ".github", "workflows", "actionlint.yml"),
      `name: actionlint
on: pull_request
jobs:
  actionlint:
    runs-on: ubuntu-latest
    steps:
      - name: Install actionlint
        env:
          ACTIONLINT_VERSION: "1.7.12"
          ACTIONLINT_SHA256: "8aca8db96f1b94770f1b0d72b6dddcb1ebb8123cb3712530b08cc387b349a3d8"
        run: |
          set -euo pipefail
          archive="actionlint_\${ACTIONLINT_VERSION}_linux_amd64.tar.gz"
          curl -fsSL "https://github.com/rhysd/actionlint/releases/download/v\${ACTIONLINT_VERSION}/\${archive}" -o "\${archive}"
          echo "\${ACTIONLINT_SHA256}  \${archive}" | sha256sum -c -
          tar -xzf "\${archive}" actionlint
          ./actionlint
`,
    );
  });
  try {
    const { status, stdout } = runCheck(REMOTE_CHECK, root);
    assert.equal(status, 0, stdout);
  } finally {
    cleanup(root);
  }
});

// Ensure scripts are executable for direct invocation style used in CI docs
test("scripts are present and bash-runnable", () => {
  chmodSync(AGENT_CHECK, 0o755);
  chmodSync(REMOTE_CHECK, 0o755);
  assert.ok(true);
});
