/**
 * TASK-444-S1: freeze `@/types/generated/models` import sites.
 *
 * - Inventory gate: allowlist file must match the real import-site set exactly
 *   (new unlisted site OR stale allowlist entry → fail closed).
 * - ESLint contract: unlisted importer fails `no-restricted-imports`;
 *   allowlisted importer does not get this boundary error.
 *
 * Run: `docker compose exec -T frontend npx vitest run src/types/generated-model-response-boundary.test.ts`
 */
import { readFileSync, readdirSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { describe, expect, it } from "vitest";
import { loadESLint } from "eslint";

const FRONTEND_ROOT = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "../..");
const ALLOWLIST_PATH = path.join(FRONTEND_ROOT, "generated-models-import-allowlist.json");
const SRC_ROOT = path.join(FRONTEND_ROOT, "src");
// Alias (`@/types/generated/models`) and relative (`./generated/models`, etc.).
const MODELS_IMPORT_RE = /from\s+["'][^"']*generated\/models(?:\.ts)?["']/;
const BOUNDARY_MESSAGE_RE = /TASK-444-S1|generated\/models/;

function collectModelImportSites(dir: string): string[] {
  const results: string[] = [];
  const entries = readdirSync(dir, { withFileTypes: true });
  for (const entry of entries) {
    const full = path.join(dir, entry.name);
    if (entry.isDirectory()) {
      if (entry.name === "node_modules" || entry.name === "dist" || entry.name === "generated") {
        continue;
      }
      results.push(...collectModelImportSites(full));
      continue;
    }
    if (!entry.isFile()) continue;
    if (!/\.(ts|tsx)$/.test(entry.name)) continue;
    // Skip this test file itself (must not import models).
    if (full.endsWith("generated-model-response-boundary.test.ts")) continue;
    const text = readFileSync(full, "utf8");
    if (MODELS_IMPORT_RE.test(text)) {
      results.push(path.relative(FRONTEND_ROOT, full).split(path.sep).join("/"));
    }
  }
  return results.sort();
}

function loadAllowlist(): string[] {
  const raw = JSON.parse(readFileSync(ALLOWLIST_PATH, "utf8")) as unknown;
  if (!Array.isArray(raw) || raw.some((p) => typeof p !== "string")) {
    throw new Error("generated-models-import-allowlist.json must be a JSON string array");
  }
  return [...raw].sort();
}

describe("TASK-444-S1 generated-model response boundary", () => {
  it("allowlist exactly matches measured import sites (fail closed on drift)", () => {
    const allowlist = loadAllowlist();
    const actual = collectModelImportSites(SRC_ROOT);

    const allowSet = new Set(allowlist);
    const actualSet = new Set(actual);

    const unlisted = actual.filter((p) => !allowSet.has(p));
    const stale = allowlist.filter((p) => !actualSet.has(p));

    expect(
      {
        count: actual.length,
        allowlistCount: allowlist.length,
        unlisted,
        stale,
      },
      "import-site set must equal allowlist (new site or stale entry is a failure)",
    ).toEqual({
      count: allowlist.length,
      allowlistCount: allowlist.length,
      unlisted: [],
      stale: [],
    });
  });

  it("unlisted file importing generated/models fails lint (alias + relative RED fixtures)", async () => {
    const ESLint = await loadESLint({ useFlatConfig: true });
    const eslint = new ESLint({ cwd: FRONTEND_ROOT });
    // Build import paths from parts so this test file itself does not become
    // a false-positive inventory hit for the frozen allowlist grep/walk.
    const aliasModule = ["@", "/types/", "generated/", "models"].join("");
    const relativeModule = ["./", "generated/", "models"].join("");
    const cases = [
      {
        label: "alias",
        source: [
          `import type { Pet } from "${aliasModule}";`,
          "export type Task444S1UnlistedAlias = Pet;",
          "",
        ].join("\n"),
        filePath: path.join(FRONTEND_ROOT, "src/types/__task444_s1_unlisted_alias_fixture__.ts"),
      },
      {
        label: "relative",
        source: [
          `import type { Pet } from "${relativeModule}";`,
          "export type Task444S1UnlistedRelative = Pet;",
          "",
        ].join("\n"),
        filePath: path.join(FRONTEND_ROOT, "src/types/__task444_s1_unlisted_relative_fixture__.ts"),
      },
    ] as const;

    for (const c of cases) {
      const [result] = await eslint.lintText(c.source, { filePath: c.filePath });
      const boundaryHits = result.messages.filter(
        (m) => m.ruleId === "no-restricted-imports" && BOUNDARY_MESSAGE_RE.test(m.message),
      );
      expect(
        boundaryHits.length,
        `expected TASK-444-S1 boundary error for ${c.label} import, got: ${JSON.stringify(result.messages)}`,
      ).toBeGreaterThan(0);
    }
  });

  it("allowlisted file may still import models without this boundary error", async () => {
    const ESLint = await loadESLint({ useFlatConfig: true });
    const eslint = new ESLint({ cwd: FRONTEND_ROOT });
    const results = await eslint.lintFiles(["src/types/pet.ts"]);
    expect(results.length).toBe(1);
    const boundaryHits = results[0].messages.filter(
      (m) => m.ruleId === "no-restricted-imports" && BOUNDARY_MESSAGE_RE.test(m.message),
    );
    expect(boundaryHits).toEqual([]);
  });
});
