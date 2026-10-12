#!/usr/bin/env node
import assert from "node:assert/strict";
import { spawnSync } from "node:child_process";
import { mkdtempSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";
import test from "node:test";

const DIR = dirname(fileURLToPath(import.meta.url));
const BRIDGE = join(DIR, "plane-mcp-bridge.mjs");
const PRELOAD = join(DIR, "plane-mcp-bridge.fake-fetch.mjs");
const KEY = "fixture";
const MARKER = "UPSTREAM-BODY-MARKER-9f3c";

function runBridge(lines, mode) {
  const dir = mkdtempSync(join(tmpdir(), "plane-bridge-"));
  const keyFile = join(dir, "key");
  writeFileSync(keyFile, KEY + "\n", { mode: 0o600 });
  const result = spawnSync(
    process.execPath,
    ["--import", `file://${PRELOAD}`, BRIDGE],
    {
      input: lines.join("\n") + "\n",
      encoding: "utf8",
      env: { ...process.env, PLANE_MCP_KEY_FILE: keyFile, PLANE_MCP_FAKE_FETCH: mode },
      timeout: 15000,
    },
  );
  assert.equal(result.status, 0, `bridge exited ${result.status}: ${result.stderr}`);
  const stdout = result.stdout.trim().split("\n").filter(Boolean);
  return {
    stdout,
    parsed: stdout.map((l) => JSON.parse(l)),
    stderr: result.stderr,
    raw: result.stdout + result.stderr,
  };
}

function request(id) {
  return JSON.stringify({ jsonrpc: "2.0", id, method: "tools/list", params: {} });
}

test("HTTP error statuses emit one -32603 preserving the request id", () => {
  for (const [mode, status] of [["http-401", 401], ["http-502", 502], ["http-429", 429]]) {
    const r = runBridge([request(7)], mode);
    assert.equal(r.parsed.length, 1, `mode ${mode}: ${r.stdout}`);
    const err = r.parsed[0];
    assert.equal(err.jsonrpc, "2.0");
    assert.equal(err.id, 7);
    assert.equal(err.error.code, -32603);
    assert.match(err.error.message, new RegExp(String(status)));
  }
});

test("fetch throw and body-read throw emit sanitized -32603", () => {
  for (const mode of ["fetch-throw", "body-throw", "non-json", "malformed-sse"]) {
    const r = runBridge([request(3)], mode);
    assert.equal(r.parsed.length, 1, `mode ${mode}: ${r.stdout}`);
    assert.equal(r.parsed[0].id, 3);
    assert.equal(r.parsed[0].error.code, -32603);
    assert.equal(typeof r.parsed[0].error.message, "string");
  }
});

test("request ids 0, string and null are preserved on errors", () => {
  for (const id of [0, "req-abc", null]) {
    const r = runBridge([request(id)], "http-401");
    assert.equal(r.parsed.length, 1);
    assert.strictEqual(r.parsed[0].id, id, `id ${JSON.stringify(id)} must round-trip`);
    assert.equal(r.parsed[0].error.code, -32603);
  }
});

test("successful raw JSON and SSE bodies are bridged with the request id", () => {
  const raw = runBridge([request(9)], "raw-json");
  assert.equal(raw.parsed.length, 1);
  assert.deepEqual(raw.parsed[0], { jsonrpc: "2.0", id: 9, result: { ok: true } });

  const sse = runBridge([request(9)], "sse");
  assert.equal(sse.parsed.length, 1);
  assert.deepEqual(sse.parsed[0], { jsonrpc: "2.0", id: 9, result: { ok: true } });
});

test("upstream responses must echo the request id as a valid RPC response", () => {
  for (const mode of [
    "mismatched-id",
    "missing-id",
    "both-result-error",
    "neither-result-error",
    "non-rpc-json",
    "sse-notification-only",
    "sse-mismatched-id",
  ]) {
    const r = runBridge([request(9)], mode);
    assert.equal(r.parsed.length, 1, `mode ${mode}: ${r.stdout}`);
    assert.equal(r.parsed[0].id, 9);
    assert.equal(r.parsed[0].error.code, -32603);
  }
});

test("upstream notifications may accompany the matching SSE response", () => {
  const r = runBridge([request(9)], "sse-notify-then-response");
  assert.equal(r.parsed.length, 2);
  assert.equal(r.parsed[0].method, "notifications/progress");
  assert.deepEqual(r.parsed[1], { jsonrpc: "2.0", id: 9, result: { ok: true } });
});

test("id-less notifications never receive a reply", () => {
  const notification = JSON.stringify({ jsonrpc: "2.0", method: "notifications/initialized" });
  for (const mode of ["http-401", "fetch-throw", "sse", "raw-json"]) {
    const r = runBridge([notification], mode);
    assert.equal(r.stdout.length, 0, `notification under ${mode} emitted output: ${r.stdout}`);
  }
});

test("malformed input line is a -32700 parse error with id null", () => {
  const r = runBridge(["{not valid json"], "raw-json");
  assert.equal(r.parsed.length, 1);
  assert.equal(r.parsed[0].id, null);
  assert.equal(r.parsed[0].error.code, -32700);
});

test("invalid envelopes get -32600 id null and the bridge keeps processing", () => {
  const lines = [
    "null",
    "[]",
    "1",
    "true",
    '"x"',
    JSON.stringify({ jsonrpc: "2.0", method: "notifications/initialized" }),
    JSON.stringify({ jsonrpc: "1.0", id: 1, method: "tools/list" }),
    JSON.stringify({ id: 1, method: "tools/list" }),
    JSON.stringify({ jsonrpc: "2.0", id: 1, method: 5 }),
    JSON.stringify({ jsonrpc: "2.0", id: 1, method: "" }),
    JSON.stringify({ jsonrpc: "2.0", id: 1 }),
    JSON.stringify({ jsonrpc: "2.0", id: {}, method: "tools/list" }),
    JSON.stringify({ jsonrpc: "2.0", id: [1], method: "tools/list" }),
    JSON.stringify({ jsonrpc: "2.0", id: true, method: "tools/list" }),
    request(42),
  ];
  const r = runBridge(lines, "raw-json");
  const errors = r.parsed.filter((m) => m.error);
  const ok = r.parsed.filter((m) => m.result);
  assert.equal(errors.length, 13, JSON.stringify(r.parsed));
  for (const err of errors) {
    assert.equal(err.id, null);
    assert.equal(err.error.code, -32600);
  }
  assert.deepEqual(ok, [{ jsonrpc: "2.0", id: 42, result: { ok: true } }]);
});

test("raw upstream bodies, exceptions and the API key never reach output", () => {
  for (const mode of ["http-401", "http-502", "fetch-throw", "body-throw", "non-json", "malformed-sse"]) {
    const r = runBridge([request(5)], mode);
    assert.ok(!r.raw.includes(MARKER), `mode ${mode} leaked upstream body: ${r.raw}`);
    assert.ok(!r.raw.includes(KEY), `mode ${mode} leaked the API key: ${r.raw}`);
  }
});
