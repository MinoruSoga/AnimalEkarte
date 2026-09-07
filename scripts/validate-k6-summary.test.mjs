#!/usr/bin/env node
import assert from "node:assert/strict";
import { mkdtempSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { spawnSync } from "node:child_process";
import { dirname } from "node:path";
import { fileURLToPath } from "node:url";
import test from "node:test";

import { metricActivity, validateSummary } from "./validate-k6-summary.mjs";

const SCRIPT = join(dirname(fileURLToPath(import.meta.url)), "validate-k6-summary.mjs");

test("metricActivity reads flat k6 summary-export fields first", () => {
  assert.equal(metricActivity({ count: 5755, values: { count: 0 } }), 5755);
  assert.equal(metricActivity({ passes: 10, fails: 2 }), 12);
  assert.equal(metricActivity({ passes: 1 }), 1);
  assert.equal(metricActivity({ fails: 3 }), 3);
});

test("metricActivity keeps nested values.* as a fallback", () => {
  assert.equal(metricActivity({ values: { count: 1918 } }), 1918);
  assert.equal(metricActivity({ values: { passes: 4, fails: 1 } }), 5);
});

test("metricActivity is fail-closed for missing or non-numeric activity", () => {
  assert.equal(metricActivity(undefined), 0);
  assert.equal(metricActivity({}), 0);
  assert.equal(metricActivity({ values: {} }), 0);
  assert.equal(metricActivity({ count: "5755" }), 0);
});

test("validateSummary accepts a flat artifact and rejects a zeroed one", () => {
  const ok = validateSummary({
    metrics: {
      http_reqs: { count: 5755 },
      iterations: { count: 1918 },
      checks: { passes: 10, fails: 0 },
      successful_logins: { count: 1 },
    },
  });
  assert.equal(ok.failed, false);
  assert.deepEqual(ok.report, [
    "http_reqs=5755",
    "iterations=1918",
    "checks=10",
    "successful_logins=1",
  ]);

  const zeroed = validateSummary({
    metrics: {
      http_reqs: { values: {} },
      iterations: {},
      checks: {},
      successful_logins: { count: 0 },
    },
  });
  assert.equal(zeroed.failed, true);
});

test("CLI validates fixture files and does not print secret-like keys", () => {
  const dir = mkdtempSync(join(tmpdir(), "k6-summary-"));
  const path = join(dir, "summary.json");
  writeFileSync(
    path,
    JSON.stringify({
      metrics: {
        http_reqs: { count: 825 },
        iterations: { count: 824 },
        checks: { passes: 8, fails: 0 },
        successful_logins: { count: 1 },
        password: { count: 1 },
        cookie: { count: 1 },
      },
    }),
  );

  const result = spawnSync(process.execPath, [SCRIPT, path], { encoding: "utf8" });
  assert.equal(result.status, 0, result.stderr);
  assert.match(result.stdout, /http_reqs=825/);
  assert.doesNotMatch(result.stdout, /password/);
  assert.doesNotMatch(result.stdout, /cookie/);
});
