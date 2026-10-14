#!/usr/bin/env node
import assert from "node:assert/strict";
import { createHash } from "node:crypto";
import { spawnSync } from "node:child_process";
import { readFileSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";
import test from "node:test";

const ROOT = join(dirname(fileURLToPath(import.meta.url)), "..", "..");
const read = (path) => readFileSync(join(ROOT, path), "utf8");

function namedStep(source, name, indent = 6) {
  const lines = source.split("\n");
  const start = lines.findIndex((line) => line === `${" ".repeat(indent)}- name: ${name}`);
  assert.notEqual(start, -1, `required step is absent: ${name}`);
  let end = start + 1;
  for (; end < lines.length; end += 1) {
    if (lines[end].search(/\S/) === indent && lines[end].startsWith(`${" ".repeat(indent)}- `)) break;
  }
  return lines.slice(start, end).join("\n");
}

function job(source, name) {
  const lines = source.split("\n");
  const start = lines.findIndex((line) => line === `  ${name}:`);
  assert.notEqual(start, -1, `required job is absent: ${name}`);
  let end = start + 1;
  for (; end < lines.length; end += 1) {
    if (/^  [^ ].*:$/.test(lines[end])) break;
  }
  return lines.slice(start, end).join("\n");
}

test("harness identity matches the first visible executive demo account", () => {
  const login = read("frontend/src/features/auth/components/LoginForm.tsx");
  const accountsStart = login.indexOf("const DEMO_ACCOUNTS");
  const firstAccount = login.slice(accountsStart).match(/email:\s*"([^"]+)"[\s\S]*?permissionLabel:\s*"([^"]+)"/);
  const harnessEmail = read("infra/scripts/cf-demo-smoke.py").match(/^DEMO_EMAIL = "([^"]+)"$/m);
  assert.ok(firstAccount && harnessEmail, "demo account contract is present");
  assert.ok(harnessEmail[1] === firstAccount[1], "harness uses the first visible demo identity");
  assert.ok(firstAccount[2] === "執行", "first visible demo identity has executive permission label");
});

test("normal cf:smoke alias is auth-only", () => {
  assert.equal(
    JSON.parse(read("package.json")).scripts["cf:smoke"],
    "python3 infra/scripts/cf-demo-smoke.py --auth",
  );
});

test("backend deploy has a required read-only STG demo login smoke and unchanged production job", () => {
  const path = ".github/workflows/backend-deploy.yml";
  const source = read(path);
  const step = namedStep(source, "Verify STG demo login (read-only)");
  const productionJob = job(source, "deploy-production");
  const productionHash = createHash("sha256").update(productionJob).digest("hex");
  assert.equal(productionHash, "6bda632582ba083c0254928da949bc3020dc6bd1b187fe81c823c7dc88ac23de", "production job remains at the reviewed unchanged baseline");
  assert.notEqual(createHash("sha256").update(`${productionJob}\n# modified`).digest("hex"), productionHash, "production job tampering changes the baseline digest");
  const contractSource = read("scripts/checks/cf-demo-smoke.test.mjs");
  assert.ok(
    !contractSource.includes("HEAD" + String.fromCharCode(94)) &&
      !contractSource.includes("git " + "show"),
    "workflow contract is independent of a Git parent in shallow checkout",
  );
  assert.ok(step.includes("SEEDLOGIN_DEMO_PASSWORD: ${{ secrets.SEEDLOGIN_DEMO_PASSWORD }}"), "smoke uses shared seedlogin secret");
  assert.ok(step.includes("WORKER_URL: ${{ env.WORKER_URL }}"), "smoke uses the configured staging worker");
  assert.ok(step.includes("run: python3 infra/scripts/cf-demo-smoke.py --auth"), "smoke invokes auth-only harness");
  assert.ok(!step.includes("continue-on-error"), "required auth smoke cannot be optional");
  assert.ok(!/STG_DEMO_(?:EMAIL|PASSWORD)/.test(step), "smoke has no legacy credential references");
  assert.ok(!step.includes("cf-crud-smoke.sh"), "deploy does not invoke legacy CRUD smoke");
});

test("manual smoke workflow is staging-ref guarded, environment gated, and fixed-origin", () => {
  const source = read(".github/workflows/stg-demo-smoke.yml");
  assert.ok(source.includes("name: STG Demo Smoke"), "workflow name is fixed");
  assert.ok(source.includes("if: github.ref != 'refs/heads/staging'"), "workflow rejects non-staging refs");
  assert.ok(source.includes("ref: ${{ github.sha }}"), "checkout is pinned to triggering SHA");
  assert.ok(source.includes("permissions:\n  contents: read"), "workflow permissions are read-only");
  assert.ok(source.includes("environment: staging"), "workflow uses staging environment approval");
  assert.ok(source.includes("WORKER_URL: https://animalekarte-stg-api.baritech-soga.workers.dev"), "worker origin is fixed");
  assert.ok(source.includes("options: [auth, crud]"), "mode is a two-choice input");
  assert.ok(source.includes("SMOKE_MODE: ${{ inputs.mode }}"), "mode reaches environment through an expression");
  assert.ok(source.includes("SMOKE_ALLOW_FIXTURE_WRITES: ${{ inputs.mode == 'crud' && '1' || '0' }}"), "CRUD writes require explicit mode");
  assert.ok(source.includes("SMOKE_FIXTURE_CLINIC_ID: ${{ inputs.clinic_id }}"), "clinic scope is an explicit input");
  assert.ok(source.includes('run: python3 infra/scripts/cf-demo-smoke.py "--$SMOKE_MODE"'), "mode is quoted as one CLI argument");
  assert.ok(!/https?:\/\/\$\{\{|WORKER_URL:\s*\$\{\{\s*inputs\./.test(source), "workflow has no arbitrary target input");
  assert.ok(!/run:[^\n]*\$\{\{\s*inputs\./.test(source), "inputs are not interpolated into shell source");
});

test("legacy CRUD smoke is blocked before its first network request", () => {
  const source = read("infra/scripts/cf-crud-smoke.sh");
  const guard = 'if [[ "${LEGACY_SMOKE_EXPLICIT_APPROVAL:-}" != "1" ]]; then\n  printf \'%s\\n\' \'BLOCKED legacy smoke can modify existing records; use cf-demo-smoke.py instead\' >&2\n  exit 2\nfi';
  const guardAt = source.indexOf(guard);
  assert.ok(guardAt > source.indexOf("set -uo pipefail"), "legacy guard follows strict mode");
  const firstNetwork = source.indexOf('HEALTH_RAW="$(curl');
  assert.ok(guardAt < firstNetwork, "legacy guard precedes all network calls");
});

test("auth cleanup uses the direct logout route and rejects redirects", () => {
  const source = read("infra/scripts/cf-demo-smoke.py");
  assert.ok(source.includes('self.request("POST", "/api/v1/auth/refresh/logout")'), "cleanup calls the direct logout endpoint");
  assert.ok(source.includes("self.logout_verified = status == 200"), "only HTTP 200 confirms logout");
  assert.ok(source.includes("class NoRedirect(urllib.request.HTTPRedirectHandler):"), "redirect handler is installed");
  assert.ok(source.includes("        return None\n\n\ndef positive_id"), "HTTP redirects are rejected");
  assert.ok(!source.includes('self.request("POST", "/api/v1/logout")'), "legacy redirect endpoint is never called");
});

test("Python fake-HTTP harness suite runs through the supported CI interpreter", () => {
  const result = spawnSync("bash", ["-c", "source scripts/ci/ci-python.sh && ci_py -B infra/scripts/test_cf_demo_smoke.py"], {
    cwd: ROOT,
    encoding: "utf8",
    env: {
      PATH: process.env.PATH,
      HOME: process.env.HOME,
      CI: "1",
      WORKER_URL: "",
      SEEDLOGIN_DEMO_PASSWORD: "",
      STG_DEMO_EMAIL: "",
      STG_DEMO_PASSWORD: "",
      CI_PYTHON: process.env.CI_PYTHON ?? "",
    },
    timeout: 120_000,
  });
  assert.equal(result.status, 0, "loopback unit suite succeeds without inherited live credentials");
  assert.ok(result.stdout.length < 12_000, "unit output is bounded");
  assert.ok(!result.stdout.includes("dummy secret") && !result.stderr.includes("dummy secret"), "secret fixture is not printed");
});
