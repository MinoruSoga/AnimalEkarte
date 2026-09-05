#!/usr/bin/env node
import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";
import test from "node:test";

const ROOT = join(dirname(fileURLToPath(import.meta.url)), "..");
const PNPM_VERSION = "10.15.0";

function read(relativePath) {
  return readFileSync(join(ROOT, relativePath), "utf8");
}

function setupPnpmBlocks(workflow) {
  const lines = workflow.split("\n");
  const blocks = [];
  for (let index = 0; index < lines.length; index += 1) {
    if (!lines[index].includes("uses: pnpm/action-setup@")) continue;
    const indent = lines[index].search(/\S/);
    const block = [lines[index]];
    for (let next = index + 1; next < lines.length; next += 1) {
      const nextIndent = lines[next].search(/\S/);
      if (nextIndent !== -1 && nextIndent < indent) break;
      block.push(lines[next]);
    }
    blocks.push(block.join("\n"));
  }
  return blocks;
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

test("AgentShield fail gate treats every AGENTS.md as agent configuration", () => {
  const workflow = read(".github/workflows/security-scan.yml");
  assert.match(workflow, /^\s+- ['"]\*\*\/AGENTS\.md['"]\s*$/m);
});

test("Docker, packageManager declarations, and CI use pnpm 10.15.0", () => {
  assert.match(
    read("frontend/Dockerfile.dev"),
    new RegExp(
      `npm install -g pnpm@${PNPM_VERSION.replaceAll(".", "\\.")}(?:\\s|$)`,
    ),
  );

  for (const packageJson of ["package.json", "frontend/package.json"]) {
    const manifest = JSON.parse(read(packageJson));
    assert.equal(manifest.packageManager, `pnpm@${PNPM_VERSION}`, packageJson);
  }

  const workflowPaths = [
    ".github/workflows/backend-deploy.yml",
    ".github/workflows/ci.yml",
    ".github/workflows/frontend-deploy.yml",
    ".github/workflows/performance-tests.yml",
  ];
  let setupCount = 0;
  for (const workflowPath of workflowPaths) {
    const blocks = setupPnpmBlocks(read(workflowPath));
    setupCount += blocks.length;
    for (const block of blocks) {
      assert.match(block, /^\s+version: 10\.15\.0\s*$/m, workflowPath);
    }
  }
  assert.equal(
    setupCount,
    7,
    "expected every pnpm/action-setup use to be covered",
  );
});

test("frontend deploy bakes VERCEL_ENV into the prebuilt Vite bundle", () => {
  const workflow = read(".github/workflows/frontend-deploy.yml");
  assert.match(
    workflow,
    /VERCEL_ENV="\$ENV" VITE_VERCEL_ENV="\$ENV" pnpm --dir frontend build/,
  );
  assert.match(workflow, /if \[ "\$ENV" = "preview" \]/);
  assert.match(
    workflow,
    /vercel alias set "\$DEPLOY_URL" stg\.noah-karte\.com/,
  );
});

test("frontend audit treats registry audit endpoint timeouts as unavailable", () => {
  const workflow = read(".github/workflows/ci.yml");
  assert.match(workflow, /ERR_PNPM_AUDIT_BAD_RESPONSE\|ERR_SOCKET_TIMEOUT/);
});

test("frontend pnpm install policy remains explicit", () => {
  const manifest = JSON.parse(read("frontend/package.json"));
  assert.deepEqual(manifest.pnpm?.onlyBuiltDependencies, [
    "@swc/core",
    "esbuild",
    "msw",
  ]);
});

test("backend Compose receives APP_ENV with the development default", () => {
  const backend = yamlBlock(read("docker-compose.yml"), "backend", 2);
  assert.match(backend, /^\s+APP_ENV: \$\{APP_ENV:-development\}\s*$/m);
});

test("E2E job uses the test-only synthetic login and runs only the auth smoke spec", () => {
  const e2e = workflowJob(read(".github/workflows/e2e.yml"), "e2e");
  assert.match(e2e, /^\s+APP_ENV: test\s*$/m);
  assert.match(
    e2e,
    /^\s+E2E_LOGIN_EMAIL: stg-staff-10000021@example\.test\s*$/m,
  );
  assert.match(e2e, /^\s+E2E_LOGIN_PASSWORD: password\s*$/m);

  const run = namedStep(e2e, "Run Playwright E2E");
  assert.match(
    run,
    /^\s+run: \.\/scripts\/run-e2e\.sh e2e\/auth-flows\.spec\.ts\s*$/m,
  );
  assert.doesNotMatch(
    e2e,
    /(?:echo|printf).*E2E_LOGIN|E2E_LOGIN.*(?:echo|printf)/,
  );
});

test("auth smoke keeps the successful-login response and authenticated-home assertions", () => {
  const authSpec = read("frontend/e2e/auth-flows.spec.ts");
  assert.match(authSpec, /test\("\/login — 有効な認証情報でログインできる"/);
  assert.match(authSpec, /expect\(loginResponse\.status\(\)\)\.toBe\(200\)/);
  assert.match(authSpec, /expect\(loginPage\.homeHeading\(\)\)\.toBeVisible/);
});

test("local load job and both k6 scripts use the dedicated fail-closed login variables", () => {
  const loadJob = workflowJob(
    read(".github/workflows/performance-tests.yml"),
    "load-test",
  );
  assert.match(loadJob, /^\s+APP_ENV: test\s*$/m);
  assert.doesNotMatch(loadJob, /STG_DEMO/);

  for (const stepName of [
    "Run k6 API endpoints load test",
    "Run k6 spike test",
  ]) {
    const step = namedStep(loadJob, stepName);
    assert.match(step, /^\s+LOAD_TEST_LOGIN_EMAIL:/m, stepName);
    assert.match(step, /^\s+LOAD_TEST_LOGIN_PASSWORD:/m, stepName);
  }

  for (const scriptPath of [
    "load-tests/k6-api-endpoints.js",
    "load-tests/k6-spike-test.js",
  ]) {
    const script = read(scriptPath);
    assert.match(script, /__ENV\.LOAD_TEST_LOGIN_EMAIL/);
    assert.match(script, /__ENV\.LOAD_TEST_LOGIN_PASSWORD/);
    assert.match(
      script,
      /if \(!LOAD_TEST_LOGIN_EMAIL \|\| !LOAD_TEST_LOGIN_PASSWORD\)/,
    );
    assert.doesNotMatch(script, /STG_DEMO/);
    assert.match(script, /loginRes\.status !== 200/);
    assert.match(script, /Set-Cookie/);
    assert.match(script, /handleSummary/);
  }
});

test("E2E and local load jobs each own always-run volume cleanup", () => {
  const e2e = workflowJob(read(".github/workflows/e2e.yml"), "e2e");
  const performanceWorkflow = read(".github/workflows/performance-tests.yml");
  const loadJob = workflowJob(performanceWorkflow, "load-test");
  const summaryJob = workflowJob(performanceWorkflow, "summary");

  for (const job of [e2e, loadJob]) {
    const cleanup = namedStep(job, "Stop app stack");
    assert.match(cleanup, /^\s+if: always\(\)\s*$/m);
    assert.match(cleanup, /^\s+run: docker compose down -v\s*$/m);
  }

  assert.doesNotMatch(summaryJob, /docker compose down -v/);
});

test("clinical-plan inventory matches the bounded current PATCH contract", () => {
  const inventory = read("docs/ops/testing/scenarios/FORM-FIELD-INVENTORY.md");
  const heading =
    "### medical-record-clinical-plan PATCH — `/api/v1/medical-records/:id/clinical-plan`";
  const start = inventory.indexOf(heading);
  assert.notEqual(start, -1, "missing bounded clinical-plan inventory heading");
  const nextHeading = inventory.indexOf("\n### ", start + heading.length);
  const section = inventory.slice(
    start,
    nextHeading === -1 ? undefined : nextHeading,
  );

  const saveAction = read(
    "frontend/src/features/medical-records/hooks/use-medical-record-save-action.ts",
  );
  const clinicalPlanApi = read(
    "frontend/src/features/medical-records/api/clinical-plan.ts",
  );
  const clinicalPlanRequest = read(
    "backend/internal/medicalrecord/clinical_plan_request.go",
  );
  const routes = read("backend/internal/medicalrecord/routes_records.go");
  const treatmentPlanPayload = saveAction.match(
    /const treatmentPlanPayload = \{[\s\S]*?\n\s*\};/,
  );
  assert.notEqual(treatmentPlanPayload, null, "missing treatmentPlanPayload");

  const mappings = [
    [
      "physicalExam",
      "physical_exam",
      /physical_exam:\s*snapshot\.physicalExam/,
    ],
    [
      "diagnosis1CategoryId",
      "diagnosis_type_id",
      /diagnosis_type_id:\s*snapshot\.diagnosis1CategoryId \?\? undefined/,
    ],
    [
      "diagnosis1NameId",
      "diagnosis_name_id",
      /diagnosis_name_id:\s*snapshot\.diagnosis1NameId \?\? undefined/,
    ],
    [
      "diagnosis2CategoryId",
      "diagnosis_2_type_id",
      /diagnosis_2_type_id:\s*snapshot\.diagnosis2CategoryId/,
    ],
    [
      "diagnosis2NameId",
      "diagnosis_2_name_id",
      /diagnosis_2_name_id:\s*snapshot\.diagnosis2NameId/,
    ],
    [
      "assessment",
      "diagnosis_details",
      /diagnosis_details:\s*snapshot\.assessment/,
    ],
    ["plan", "treatment_policy", /treatment_policy:\s*snapshot\.plan/],
    [
      "existingClinicalPlanVersion",
      "version",
      /version:\s*snapshot\.existingClinicalPlanVersion/,
    ],
  ];

  for (const [uiState, wireKey, savePayload] of mappings) {
    const inventoryRow = new RegExp(
      `\\|\\s*${uiState}\\s*\\|\\s*${wireKey}\\s*\\|`,
    );
    assert.match(section, inventoryRow, `${uiState} inventory mapping`);
    assert.match(
      treatmentPlanPayload[0],
      savePayload,
      `${uiState} save payload mapping`,
    );
    assert.match(
      clinicalPlanApi,
      new RegExp(`\\b${wireKey}\\?:`),
      `${wireKey} frontend API input`,
    );
    assert.match(
      clinicalPlanRequest,
      new RegExp(`json:"${wireKey}"`),
      `${wireKey} Go request tag`,
    );
  }

  assert.match(section, /version.*S.*CAS.*user-entered/i);
  assert.match(section, /Owner: clinical_plan PATCH child resource\./);
  for (const staleKey of [
    "soap_s",
    "soap_o",
    "soap_a",
    "soap_p",
    "diagnosis3_type",
    "diagnosis3_name",
  ]) {
    assert.doesNotMatch(section, new RegExp(`\\b${staleKey}\\b`));
  }

  assert.match(
    clinicalPlanApi,
    /\.patch[\s\S]*`\/v1\/medical-records\/\$\{medicalRecordId\}\/clinical-plan`/,
  );
  assert.match(
    routes,
    /records\.PATCH\("\/:id\/clinical-plan",[^\n]*h\.clinicalPlan\.UpdateClinicalPlan\)/,
  );
  assert.match(
    saveAction,
    /updateTreatmentPlanMutation\.mutateAsync\(treatmentPlanPayload\)/,
  );
});
