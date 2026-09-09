// verify-e2e-page-consumers.test.mjs — TypeScript AST page-consumer contract.
// Fixtures use os.tmpdir(); never write under /app (Docker bind is readonly).

import test from "node:test";
import assert from "node:assert/strict";
import {
  mkdtempSync,
  mkdirSync,
  writeFileSync,
  rmSync,
} from "node:fs";
import { tmpdir } from "node:os";
import path from "node:path";
import { fileURLToPath, pathToFileURL } from "node:url";
import { spawnSync } from "node:child_process";

const SCRIPT = fileURLToPath(new URL("./verify-e2e-page-consumers.mjs", import.meta.url));

function loadModule() {
  return import(pathToFileURL(SCRIPT).href + `?t=${Date.now()}`);
}

function makeFixture(entries) {
  const root = mkdtempSync(path.join(tmpdir(), "ae-e2e-consumers-"));
  const e2eRoot = path.join(root, "e2e");
  mkdirSync(path.join(e2eRoot, "pages"), { recursive: true });
  for (const [rel, body] of Object.entries(entries)) {
    const full = path.join(root, rel);
    mkdirSync(path.dirname(full), { recursive: true });
    writeFileSync(full, body, "utf8");
  }
  return { root, e2eRoot };
}

function runCli(args, cwd) {
  return spawnSync(process.execPath, [SCRIPT, ...args], {
    cwd,
    encoding: "utf8",
    env: { ...process.env, HOME: tmpdir(), TMPDIR: tmpdir(), CI: "true" },
  });
}

function parseStdout(result) {
  const text = (result.stdout || "").trim();
  assert.ok(text, `expected JSON stdout, stderr=${result.stderr}`);
  return JSON.parse(text);
}

test("RED: default+named static import is an exact consumer", async () => {
  const mod = await loadModule();
  const source = 'import Default, { named } from "./pages/accounting-page";\n';
  const refs = mod.collectModuleReferences(source, "e2e/good.spec.ts");
  assert.ok(
    refs.some(
      (ref) =>
        ref.form === "static-import" &&
        ref.specifier === "./pages/accounting-page" &&
        ref.identity === "e2e/pages/accounting-page.ts",
    ),
    `expected default+named static consumer, got ${JSON.stringify(refs)}`,
  );

  const { root } = makeFixture({
    "e2e/pages/accounting-page.ts": "export class AccountingPage {}\n",
    "e2e/good.spec.ts": source,
  });
  try {
    const result = runCli(
      ["--page", "e2e/pages/accounting-page.ts", "--e2e-root", "e2e"],
      root,
    );
    assert.equal(result.status, 0, result.stderr);
    const payload = parseStdout(result);
    assert.equal(payload.ok, true);
    assert.equal(payload.pages[0].consumers.length, 1);
    assert.equal(payload.pages[0].consumers[0].form, "static-import");
    assert.equal(payload.pages[0].blocking.length, 0);
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
});

test("RED: helper URL dynamic import is not an accounting-page alias / does not block", async () => {
  const mod = await loadModule();
  const source =
    'const load = () => import("https://example.test/pages/accounting-page-helper.ts");\n';
  const refs = mod.collectModuleReferences(source, "e2e/bad.spec.ts");
  assert.equal(
    refs.filter((ref) => ref.identity === "e2e/pages/accounting-page.ts").length,
    0,
    `helper URL must not canonicalize to accounting-page, got ${JSON.stringify(refs)}`,
  );

  const { root } = makeFixture({
    "e2e/pages/accounting-page.ts": "export class AccountingPage {}\n",
    "e2e/good.spec.ts":
      'import { AccountingPage } from "./pages/accounting-page";\n',
    "e2e/helper-url.spec.ts": source,
  });
  try {
    const result = runCli(
      ["--page", "e2e/pages/accounting-page.ts", "--e2e-root", "e2e"],
      root,
    );
    assert.equal(result.status, 0, result.stderr);
    const payload = parseStdout(result);
    assert.equal(payload.ok, true);
    assert.equal(payload.pages[0].blocking.length, 0);
    assert.ok(payload.pages[0].consumers.length >= 1);
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
});

test("static import matrix: named/namespace/default/default+named/default+namespace/type-only/multiline/ext/no-ext/require", async () => {
  const mod = await loadModule();
  const cases = [
    ['import { AccountingPage } from "./pages/accounting-page";\n', "static-import"],
    ['import * as Accounting from "./pages/accounting-page";\n', "static-import"],
    ['import AccountingPage from "./pages/accounting-page";\n', "static-import"],
    ['import Default, { named } from "./pages/accounting-page";\n', "static-import"],
    ['import Default, * as NS from "./pages/accounting-page";\n', "static-import"],
    ['import type { AccountingPage } from "./pages/accounting-page";\n', "static-import"],
    [
      'import {\n  AccountingPage,\n} from "./pages/accounting-page";\n',
      "static-import",
    ],
    ['import { AccountingPage } from "./pages/accounting-page.ts";\n', "static-import"],
    ['const { AccountingPage } = require("./pages/accounting-page");\n', "require"],
    ['const page = require("./pages/accounting-page.ts");\n', "require"],
  ];

  for (const [source, form] of cases) {
    const refs = mod.collectModuleReferences(source, "e2e/matrix.spec.ts");
    const hit = refs.find((ref) => ref.identity === "e2e/pages/accounting-page.ts");
    assert.ok(hit, `missing consumer for form=${form}: ${source}`);
    assert.equal(hit.form, form, source);
  }
});

test("unsupported exact-target forms block even when a valid consumer exists", async () => {
  const { root } = makeFixture({
    "e2e/pages/accounting-page.ts": "export class AccountingPage {}\n",
    "e2e/good.spec.ts":
      'import { AccountingPage } from "./pages/accounting-page";\n',
    "e2e/side-effect.spec.ts": 'import "./pages/accounting-page";\n',
    "e2e/dynamic.spec.ts":
      'const load = () => import("./pages/accounting-page");\n',
    "e2e/export-from.spec.ts":
      'export { AccountingPage } from "./pages/accounting-page";\n',
    "e2e/nonliteral.spec.ts": "const load = () => import(moduleName);\n",
  });
  try {
    const result = runCli(
      ["--page", "e2e/pages/accounting-page.ts", "--e2e-root", "e2e"],
      root,
    );
    assert.notEqual(result.status, 0);
    const payload = JSON.parse((result.stdout || result.stderr).match(/\{[\s\S]*\}/)[0]);
    assert.equal(payload.ok, false);
    const forms = new Set(payload.pages[0].blocking.map((item) => item.form));
    assert.ok(forms.has("side-effect-import"), payload.pages[0].blocking);
    assert.ok(forms.has("dynamic-import"), payload.pages[0].blocking);
    assert.ok(forms.has("export-from"), payload.pages[0].blocking);
    assert.ok(
      forms.has("nonliteral-dynamic-import") || forms.has("nonliteral"),
      payload.pages[0].blocking,
    );
    assert.ok(payload.pages[0].consumers.length >= 1);
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
});

test("malformed syntax fails closed", async () => {
  const { root } = makeFixture({
    "e2e/pages/accounting-page.ts": "export class AccountingPage {}\n",
    "e2e/good.spec.ts":
      'import { AccountingPage } from "./pages/accounting-page";\n',
    "e2e/broken.spec.ts": "import { from \"./pages/accounting-page\";\n",
  });
  try {
    const result = runCli(
      ["--page", "e2e/pages/accounting-page.ts", "--e2e-root", "e2e"],
      root,
    );
    assert.notEqual(result.status, 0);
    const raw = `${result.stdout}\n${result.stderr}`;
    assert.match(raw, /\{[\s\S]*\}/);
    const payload = JSON.parse(raw.match(/\{[\s\S]*\}/)[0]);
    assert.equal(payload.ok, false);
    assert.ok(
      payload.pages[0].blocking.some(
        (item) => item.file.includes("broken.spec.ts") && /parse/i.test(item.form),
      ),
      payload.pages[0].blocking,
    );
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
});

test("strings/comments/templates/regex lookalikes are ignored", async () => {
  const mod = await loadModule();
  const source = [
    'const message = "import { AccountingPage } from \'./pages/accounting-page\'";',
    "// import { AccountingPage } from './pages/accounting-page';",
    "/* import { AccountingPage } from './pages/accounting-page' */",
    "const tpl = `import { AccountingPage } from './pages/accounting-page'`;",
    'const pattern = /import { AccountingPage } from ".\\/pages\\/accounting-page"/;',
    "",
  ].join("\n");
  const refs = mod.collectModuleReferences(source, "e2e/lookalike.spec.ts");
  assert.equal(
    refs.filter((ref) => ref.identity === "e2e/pages/accounting-page.ts").length,
    0,
    JSON.stringify(refs),
  );
});

test("alias/canonicalization: relative, /app/e2e, @/e2e, helper, query, basename, traversal", async () => {
  const mod = await loadModule();
  const page = "e2e/pages/accounting-page.ts";

  assert.equal(
    mod.canonicalizeSpecifier("e2e/flow.spec.ts", "./pages/accounting-page"),
    page,
  );
  assert.equal(
    mod.canonicalizeSpecifier("e2e/flow.spec.ts", "./pages/accounting-page.ts"),
    page,
  );
  assert.equal(
    mod.canonicalizeSpecifier("e2e/subdir/nested.spec.ts", "../pages/accounting-page"),
    page,
  );
  assert.equal(
    mod.canonicalizeSpecifier("e2e/flow.spec.ts", "/app/e2e/pages/accounting-page"),
    page,
  );
  assert.equal(
    mod.canonicalizeSpecifier("e2e/flow.spec.ts", "@/e2e/pages/accounting-page"),
    page,
  );
  assert.equal(
    mod.canonicalizeSpecifier("e2e/flow.spec.ts", "@/pages/accounting-page"),
    null,
  );
  assert.equal(
    mod.canonicalizeSpecifier(
      "e2e/flow.spec.ts",
      "https://example.test/pages/accounting-page-helper.ts",
    ),
    null,
  );
  assert.equal(
    mod.canonicalizeSpecifier(
      "e2e/flow.spec.ts",
      "https://example.test/e2e/pages/accounting-page.ts?x=1",
    ),
    null,
  );
  assert.equal(
    mod.canonicalizeSpecifier(
      "e2e/flow.spec.ts",
      "https://user:pass@example.test/e2e/pages/accounting-page.ts",
    ),
    null,
  );
  assert.equal(
    mod.canonicalizeSpecifier(
      "e2e/flow.spec.ts",
      "file:///app/e2e/pages/accounting-page.ts",
    ),
    page,
  );
  assert.equal(
    mod.canonicalizeSpecifier(
      "e2e/flow.spec.ts",
      "file://evil.com/app/e2e/pages/accounting-page.ts",
    ),
    null,
  );
  assert.equal(
    mod.canonicalizeSpecifier(
      "e2e/flow.spec.ts",
      "file://localhost/app/e2e/pages/accounting-page.ts",
    ),
    null,
  );
  assert.equal(
    mod.canonicalizeSpecifier("e2e/flow.spec.ts", "./other/accounting-page"),
    "e2e/other/accounting-page.ts",
  );
  assert.notEqual(
    mod.canonicalizeSpecifier("e2e/flow.spec.ts", "./other/accounting-page"),
    page,
  );
  assert.equal(
    mod.canonicalizeSpecifier("e2e/flow.spec.ts", "./pages/accounting-page-helper"),
    "e2e/pages/accounting-page-helper.ts",
  );
  assert.notEqual(
    mod.canonicalizeSpecifier("e2e/flow.spec.ts", "./pages/accounting-page-helper"),
    page,
  );
  assert.equal(
    mod.canonicalizeSpecifier("e2e/flow.spec.ts", "../pages/accounting-page/../../secret"),
    null,
  );
  assert.equal(
    mod.canonicalizeSpecifier("e2e/flow.spec.ts", ".\\pages\\accounting-page"),
    null,
  );
});

test("verifyPages reports JSON shape with consumers and empty blocking on success", async () => {
  const mod = await loadModule();
  const { root, e2eRoot } = makeFixture({
    "e2e/pages/accounting-page.ts": "export class AccountingPage {}\n",
    "e2e/accounting-flow.spec.ts":
      'import { AccountingPage } from "./pages/accounting-page";\n',
  });
  try {
    const payload = mod.verifyPages({
      pages: ["e2e/pages/accounting-page.ts"],
      e2eRoot,
    });
    assert.equal(payload.ok, true);
    assert.equal(payload.pages.length, 1);
    assert.equal(payload.pages[0].page, "e2e/pages/accounting-page.ts");
    assert.equal(payload.pages[0].consumers[0].file, "e2e/accounting-flow.spec.ts");
    assert.equal(payload.pages[0].consumers[0].form, "static-import");
    assert.equal(
      payload.pages[0].consumers[0].specifier,
      "./pages/accounting-page",
    );
    assert.deepEqual(payload.pages[0].blocking, []);
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
});

test("zero consumers fails closed", async () => {
  const { root } = makeFixture({
    "e2e/pages/orphan-page.ts": "export class Orphan {}\n",
    "e2e/unrelated.spec.ts": 'import { test } from "@playwright/test";\n',
  });
  try {
    const result = runCli(
      ["--page", "e2e/pages/orphan-page.ts", "--e2e-root", "e2e"],
      root,
    );
    assert.notEqual(result.status, 0);
    const payload = JSON.parse((result.stdout || result.stderr).match(/\{[\s\S]*\}/)[0]);
    assert.equal(payload.ok, false);
    assert.equal(payload.pages[0].consumers.length, 0);
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
});
