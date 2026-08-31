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
      if (nextIndent !== -1 && nextIndent <= indent) break;
      block.push(lines[next]);
    }
    blocks.push(block.join("\n"));
  }
  return blocks;
}

test("AgentShield fail gate treats every AGENTS.md as agent configuration", () => {
  const workflow = read(".github/workflows/security-scan.yml");
  assert.match(workflow, /^\s+- ['"]\*\*\/AGENTS\.md['"]\s*$/m);
});

test("Docker, packageManager declarations, and CI use pnpm 10.15.0", () => {
  assert.match(
    read("frontend/Dockerfile.dev"),
    new RegExp(`npm install -g pnpm@${PNPM_VERSION.replaceAll(".", "\\.")}(?:\\s|$)`),
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
  assert.equal(setupCount, 7, "expected every pnpm/action-setup use to be covered");
});

test("frontend pnpm install policy remains explicit", () => {
  const manifest = JSON.parse(read("frontend/package.json"));
  assert.deepEqual(manifest.pnpm?.onlyBuiltDependencies, ["@swc/core", "esbuild", "msw"]);
});
