#!/usr/bin/env node
import assert from "node:assert/strict";
import { spawnSync } from "node:child_process";
import { readFileSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";
import test from "node:test";

const ROOT = join(dirname(fileURLToPath(import.meta.url)), "..", "..");
const HELPER = join(ROOT, "scripts/checks/check-lab-device-secret.mjs");
const WORKFLOW_PATH = ".github/workflows/backend-deploy.yml";
const REQUIRED_SECRET = "LAB_DEVICE_AGENT_CONSUMER_TOKEN";
const MISSING_MESSAGE = `Missing required remote secret: ${REQUIRED_SECRET}\n`;
const GENERIC_MESSAGE = "Unable to verify required remote secret names\n";
const MINIMAL_ENV = { PATH: `${dirname(process.execPath)}:/usr/bin:/bin` };
const PRESENT_LIST = JSON.stringify([
  { name: REQUIRED_SECRET, type: "secret_text" },
  { name: "OTHER_SECRET", type: "secret_text" },
]);

function read(relativePath) {
  return readFileSync(join(ROOT, relativePath), "utf8");
}

function yamlBlock(source, key, indent) {
  const lines = source.split("\n");
  const start = lines.findIndex(
    (line) => line === `${" ".repeat(indent)}${key}:`,
  );
  assert.notEqual(start, -1, `missing ${key} block`);
  const block = [lines[start]];
  for (let index = start + 1; index < lines.length; index += 1) {
    const line = lines[index];
    const contentIndent = line.search(/\S/);
    if (contentIndent !== -1 && contentIndent <= indent) break;
    block.push(line);
  }
  return block.join("\n");
}

function workflowJob(workflow, name) {
  const jobs = yamlBlock(workflow, "jobs", 0);
  return yamlBlock(jobs, name, 2);
}

function namedStep(job, name) {
  const lines = job.split("\n");
  const start = lines.findIndex((line) => line === `      - name: ${name}`);
  assert.notEqual(start, -1, `missing ${name} step`);
  const block = [lines[start]];
  for (let index = start + 1; index < lines.length; index += 1) {
    const line = lines[index];
    const indent = line.search(/\S/);
    if (indent === 6 && line.startsWith("      - ")) break;
    block.push(line);
  }
  return block.join("\n");
}

function extractRunBlock(step) {
  const lines = step.split("\n");
  const start = lines.findIndex((line) => /^\s+run: \|\s*$/.test(line));
  assert.notEqual(start, -1, "step must use a block-scalar run");
  const runIndent = lines[start].search(/\S/);
  const body = [];
  for (let index = start + 1; index < lines.length; index += 1) {
    const line = lines[index];
    const indent = line.search(/\S/);
    if (indent !== -1 && indent <= runIndent) break;
    body.push(line);
  }
  const nonEmpty = body.filter((line) => line.trim() !== "");
  assert.ok(nonEmpty.length > 0, "run block must not be empty");
  const contentIndent = Math.min(
    ...nonEmpty.map((line) => line.search(/\S/)),
  );
  const extracted = body.map((line) => line.slice(contentIndent));
  while (extracted.length > 0 && extracted.at(-1).trim() === "") {
    extracted.pop();
  }
  return extracted;
}

function preflightRunBlock() {
  return "set -euo pipefail\nnpx --no-install wrangler secret list | node ../scripts/checks/check-lab-device-secret.mjs";
}

function runHelper(stdin) {
  return spawnSync(process.execPath, [HELPER], {
    input: stdin,
    encoding: "utf8",
  });
}

function runPreflight(producerStdout, producerCode) {
  assert.ok(
    !producerStdout.includes("'"),
    "synthetic producer output must stay single-quote free",
  );
  const script = [
    "npx() {",
    `  printf '%s\\n' '${producerStdout}'`,
    `  return ${producerCode}`,
    "}",
    preflightRunBlock(),
    "printf 'DEPLOY_REACHED\\n'",
  ].join("\n");
  return spawnSync("bash", ["-c", script], {
    cwd: join(ROOT, "backend"),
    env: MINIMAL_ENV,
    encoding: "utf8",
  });
}

test("normal deploy omits the optional legacy agent secret check and preserves deploy guards", () => {
  const staging = workflowJob(read(WORKFLOW_PATH), "deploy");
  assert.doesNotMatch(staging, /Verify lab device consumer secret|check-lab-device-secret\.mjs|LAB_DEVICE_AGENT_CONSUMER_TOKEN/);
  const credentials = namedStep(staging, "Verify Cloudflare credentials");
  const credentialsIndex = staging.indexOf("      - name: Verify Cloudflare credentials");
  const credentialsRun = extractRunBlock(credentials).join("\n");
  const deployIndex = staging.indexOf(
    "      - name: Deploy Worker and Container",
  );
  const migrationIndex = staging.indexOf("      - name: Run database migration");
  const healthIndex = staging.indexOf("      - name: Wait for /health (post-migrate)");
  assert.match(credentials, /CLOUDFLARE_API_TOKEN: \$\{\{ secrets\.CLOUDFLARE_API_TOKEN \}\}/);
  assert.match(credentialsRun, /if \[ -z "\$\{CLOUDFLARE_API_TOKEN\}" \]/);
  assert.match(credentialsRun, /pnpm exec wrangler whoami/);
  assert.ok(credentialsIndex !== -1 && credentialsIndex < deployIndex);
  assert.ok(deployIndex < migrationIndex && migrationIndex < healthIndex);
  const manifest = read("backend/wrangler.jsonc");
  const required = manifest.match(/"required"\s*:\s*\[([\s\S]*?)\]/)?.[1] ?? "";
  assert.doesNotMatch(required, /"LAB_DEVICE_AGENT_CONSUMER_TOKEN"/);
});

test("legacy-agent preflight fixture remains fail-closed for helper behavior", () => {
  assert.equal(preflightRunBlock(), "set -euo pipefail\nnpx --no-install wrangler secret list | node ../scripts/checks/check-lab-device-secret.mjs");
});

test("normal Board source uses the browser receiver without legacy consumer or loopback paths", () => {
  const source = read("frontend/src/features/lab-device/routes/LabDeviceBoard.tsx");
  assert.match(source, /useLabDeviceBrowserReceive/);
  assert.doesNotMatch(source, /useGetLabDeviceAgentConsumer|useLabDeviceAgentListen|127\.0\.0\.1|localhost/);
});

test("production deploy job remains byte-for-byte unchanged", () => {
  const base = spawnSync("git", ["show", `HEAD:${WORKFLOW_PATH}`], {
    cwd: ROOT,
    encoding: "utf8",
  });
  assert.equal(base.status, 0, base.stderr.trim());
  assert.equal(
    workflowJob(read(WORKFLOW_PATH), "deploy-production"),
    workflowJob(base.stdout, "deploy-production"),
  );
});

test("legacy-agent preflight run block reaches the marker only when the secret is listed", () => {
  const ok = runPreflight(PRESENT_LIST, 0);
  assert.equal(ok.status, 0);
  assert.equal(ok.stdout, "DEPLOY_REACHED\n");
  assert.equal(ok.stderr, "");

  const missing = runPreflight(
    JSON.stringify([{ name: "OTHER_SECRET", type: "secret_text" }]),
    0,
  );
  assert.notEqual(missing.status, 0);
  assert.equal(missing.stdout, "");
  assert.equal(missing.stderr, MISSING_MESSAGE);
});

test("legacy-agent preflight run block stops when the producer exits nonzero", () => {
  const failed = runPreflight(PRESENT_LIST, 17);
  assert.equal(failed.status, 17);
  assert.equal(failed.stdout, "");
});

test("legacy-agent preflight run block stops on empty or malformed producer output", () => {
  for (const [index, producerStdout] of ["", "not json"].entries()) {
    const result = runPreflight(producerStdout, 0);
    assert.notEqual(result.status, 0, `case ${index}`);
    assert.equal(result.stdout, "", `case ${index}`);
    assert.equal(result.stderr, GENERIC_MESSAGE, `case ${index}`);
  }
});

test("helper exits cleanly when the required secret name is present", () => {
  const result = runHelper(PRESENT_LIST);
  assert.equal(result.status, 0);
  assert.equal(result.stdout, "");
  assert.equal(result.stderr, "");
});

test("helper reports the missing secret name", () => {
  const cases = [
    "[]",
    JSON.stringify([{ name: "OTHER_SECRET", type: "secret_text" }]),
    JSON.stringify([{ name: "lab_device_agent_consumer_token" }]),
    JSON.stringify([{ name: ` ${REQUIRED_SECRET}` }]),
    JSON.stringify([{ name: `${REQUIRED_SECRET} ` }]),
  ];
  for (const [index, stdin] of cases.entries()) {
    const result = runHelper(stdin);
    assert.equal(result.status, 1, `case ${index}`);
    assert.equal(result.stdout, "", `case ${index}`);
    assert.equal(result.stderr, MISSING_MESSAGE, `case ${index}`);
  }
});

test("helper rejects malformed and non-list payloads without echoing input", () => {
  const sentinel = "SYNTHETIC_SENTINEL_8f4c2b";
  const cases = [
    `[{"name":"${sentinel}",]`,
    "",
    "   \n\t ",
    "{}",
    "null",
    `"${REQUIRED_SECRET}"`,
    "42",
    "true",
  ];
  for (const [index, stdin] of cases.entries()) {
    const result = runHelper(stdin);
    assert.equal(result.status, 1, `case ${index}`);
    assert.equal(result.stdout, "", `case ${index}`);
    assert.equal(result.stderr, GENERIC_MESSAGE, `case ${index}`);
    assert.ok(!result.stderr.includes(sentinel), `case ${index}`);
    assert.ok(!result.stdout.includes(sentinel), `case ${index}`);
  }
});

test("helper rejects entries that are not name-bearing objects", () => {
  const cases = [
    "[null]",
    "[{}]",
    "[1]",
    `[["${REQUIRED_SECRET}"]]`,
    `[{"name":1}]`,
    `[{"name":null}]`,
    `[{"name":${JSON.stringify(REQUIRED_SECRET)}},null]`,
    `[{"name":${JSON.stringify(REQUIRED_SECRET)}},{}]`,
  ];
  for (const [index, stdin] of cases.entries()) {
    const result = runHelper(stdin);
    assert.equal(result.status, 1, `case ${index}`);
    assert.equal(result.stdout, "", `case ${index}`);
    assert.equal(result.stderr, GENERIC_MESSAGE, `case ${index}`);
  }
});
