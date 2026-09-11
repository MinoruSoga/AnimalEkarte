// verify-e2e-page-consumers.test.mjs — TypeScript AST page-consumer contract.
// Fixtures use os.tmpdir(); never write under /app (Docker bind is readonly).

import test from "node:test";
import assert from "node:assert/strict";
import {
  mkdtempSync,
  mkdirSync,
  writeFileSync,
  symlinkSync,
  unlinkSync,
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

test("static import matrix: named/namespace/default/default+named/default+namespace/type-only/multiline/ext/no-ext only", async () => {
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
  ];

  for (const [source, form] of cases) {
    const refs = mod.collectModuleReferences(source, "e2e/matrix.spec.ts");
    const hit = refs.find((ref) => ref.identity === "e2e/pages/accounting-page.ts");
    assert.ok(hit, `missing consumer for form=${form}: ${source}`);
    assert.equal(hit.form, form, source);
  }
});

test("RED: require is never a happy consumer; exact-target require blocks; unrelated stays unrelated", async () => {
  const mod = await loadModule();
  const requireSources = [
    'const { AccountingPage } = require("./pages/accounting-page");\n',
    'const page = require("./pages/accounting-page.ts");\n',
  ];
  for (const source of requireSources) {
    const refs = mod.collectModuleReferences(source, "e2e/matrix.spec.ts");
    const hit = refs.find((ref) => ref.identity === "e2e/pages/accounting-page.ts");
    assert.ok(hit, `expected require identity match for blocking: ${source}`);
    assert.equal(hit.form, "require", source);
  }

  const { root } = makeFixture({
    "e2e/pages/accounting-page.ts": "export class AccountingPage {}\n",
    "e2e/good.spec.ts":
      'import { AccountingPage } from "./pages/accounting-page";\n',
    "e2e/require-exact.spec.ts":
      'const { AccountingPage } = require("./pages/accounting-page");\n',
    "e2e/require-unrelated.spec.ts":
      'const other = require("./pages/other-page");\n',
  });
  try {
    const result = runCli(
      ["--page", "e2e/pages/accounting-page.ts", "--e2e-root", "e2e"],
      root,
    );
    assert.notEqual(result.status, 0);
    const payload = JSON.parse((result.stdout || result.stderr).match(/\{[\s\S]*\}/)[0]);
    assert.equal(payload.ok, false);
    assert.ok(
      payload.pages[0].blocking.some((item) => item.form === "require"),
      payload.pages[0].blocking,
    );
    assert.equal(
      payload.pages[0].consumers.some((item) => item.form === "require"),
      false,
    );
    assert.equal(
      payload.pages[0].blocking.some((item) =>
        item.file.includes("require-unrelated.spec.ts"),
      ),
      false,
      "unrelated require must stay unrelated",
    );
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
});

test("RED: shadowed require sole-consumer must not ok:true", async () => {
  const mod = await loadModule();
  const source = [
    "function require(name: string) { return { name }; }",
    'const x = require("./pages/accounting-page");',
    "",
  ].join("\n");
  const refs = mod.collectModuleReferences(source, "e2e/shadowed.spec.ts");
  assert.ok(
    refs.some(
      (ref) =>
        ref.form === "require" &&
        ref.identity === "e2e/pages/accounting-page.ts",
    ),
    `shadowed require still exact-target form=require for blocking, got ${JSON.stringify(refs)}`,
  );

  const { root } = makeFixture({
    "e2e/pages/accounting-page.ts": "export class AccountingPage {}\n",
    "e2e/shadowed.spec.ts": source,
  });
  try {
    const result = runCli(
      ["--page", "e2e/pages/accounting-page.ts", "--e2e-root", "e2e"],
      root,
    );
    assert.notEqual(result.status, 0);
    const payload = JSON.parse((result.stdout || result.stderr).match(/\{[\s\S]*\}/)[0]);
    assert.equal(payload.ok, false);
    assert.equal(payload.pages[0].consumers.length, 0);
    assert.ok(
      payload.pages[0].blocking.some((item) => item.form === "require"),
      payload.pages[0].blocking,
    );
  } finally {
    rmSync(root, { recursive: true, force: true });
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

test("alias/canonicalization: relative ok; @/e2e unrelated; /app+file identity only; helper/query/credentials null", async () => {
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
  // @/* maps to src/* only — never an E2E page identity.
  assert.equal(
    mod.canonicalizeSpecifier("e2e/flow.spec.ts", "@/e2e/pages/accounting-page"),
    null,
  );
  assert.equal(
    mod.canonicalizeSpecifier("e2e/flow.spec.ts", "@/pages/accounting-page"),
    null,
  );
  // Absolute /app and file URL keep identity for exact-target blocking, not consumers.
  assert.equal(
    mod.canonicalizeSpecifier("e2e/flow.spec.ts", "/app/e2e/pages/accounting-page"),
    page,
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

test("RED: absolute and file-URL exact-target static imports block, never consume", async () => {
  const { root } = makeFixture({
    "e2e/pages/accounting-page.ts": "export class AccountingPage {}\n",
    "e2e/good.spec.ts":
      'import { AccountingPage } from "./pages/accounting-page";\n',
    "e2e/absolute.spec.ts":
      'import { AccountingPage } from "/app/e2e/pages/accounting-page";\n',
    "e2e/fileurl.spec.ts":
      'import { AccountingPage } from "file:///app/e2e/pages/accounting-page.ts";\n',
    "e2e/alias.spec.ts":
      'import { AccountingPage } from "@/e2e/pages/accounting-page";\n',
  });
  try {
    const result = runCli(
      ["--page", "e2e/pages/accounting-page.ts", "--e2e-root", "e2e"],
      root,
    );
    assert.notEqual(result.status, 0);
    const payload = JSON.parse((result.stdout || result.stderr).match(/\{[\s\S]*\}/)[0]);
    assert.equal(payload.ok, false);
    const blockingForms = new Set(payload.pages[0].blocking.map((item) => item.form));
    assert.ok(
      blockingForms.has("absolute-static-import"),
      payload.pages[0].blocking,
    );
    assert.ok(
      payload.pages[0].blocking.some((item) => item.file.includes("absolute.spec.ts")),
      payload.pages[0].blocking,
    );
    assert.ok(
      payload.pages[0].blocking.some((item) => item.file.includes("fileurl.spec.ts")),
      payload.pages[0].blocking,
    );
    assert.equal(
      payload.pages[0].consumers.some((item) =>
        item.specifier.startsWith("/app/") || item.specifier.startsWith("file:"),
      ),
      false,
    );
    assert.equal(
      payload.pages[0].consumers.some((item) => item.specifier.startsWith("@/")),
      false,
    );
    // @/e2e is unrelated (tsconfig @/* → src/*), so it must not alone create a page hit.
    assert.equal(
      payload.pages[0].blocking.some((item) => item.file.includes("alias.spec.ts")),
      false,
      "unrelated @/e2e must not block as exact-target page",
    );
    assert.ok(payload.pages[0].consumers.length >= 1);
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
});

test("RED: filesystem fail-closed on missing page, bad root, and unreadable scan path", async () => {
  const mod = await loadModule();
  const { root } = makeFixture({
    "e2e/pages/accounting-page.ts": "export class AccountingPage {}\n",
    "e2e/good.spec.ts":
      'import { AccountingPage } from "./pages/accounting-page";\n',
  });
  try {
    const missing = runCli(
      ["--page", "e2e/pages/missing-page.ts", "--e2e-root", "e2e"],
      root,
    );
    assert.notEqual(missing.status, 0);
    const missingPayload = JSON.parse(
      (missing.stdout || missing.stderr).match(/\{[\s\S]*\}/)[0],
    );
    assert.equal(missingPayload.ok, false);
    assert.match(String(missingPayload.error || ""), /missing/i);

    const badRoot = runCli(
      ["--page", "e2e/pages/accounting-page.ts", "--e2e-root", "e2e-missing"],
      root,
    );
    assert.notEqual(badRoot.status, 0);

    // Root bypasses mode bits; prove scan fail-closed via unreadable path instead.
    assert.throws(
      () => mod.listSpecFiles(path.join(root, "e2e-does-not-exist")),
      /unreadable/i,
    );
    const scanPayload = mod.verifyPages({
      pages: ["e2e/pages/accounting-page.ts"],
      e2eRoot: path.join(root, "e2e"),
      listSpecs: () => {
        throw new Error(`unreadable e2e scan path: ${path.join(root, "e2e")}`);
      },
    });
    assert.equal(scanPayload.ok, false);
    assert.match(String(scanPayload.error || ""), /unreadable/i);
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
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

test("RED: symlink .spec.ts with unsupported dynamic import must not ok:true", async () => {
  const mod = await loadModule();
  const { root, e2eRoot } = makeFixture({
    "e2e/pages/accounting-page.ts": "export class AccountingPage {}\n",
    "e2e/good.spec.ts":
      'import { AccountingPage } from "./pages/accounting-page";\n',
    "outside/evil.ts":
      'const load = () => import("./pages/accounting-page");\n',
  });
  try {
    symlinkSync(
      path.join(root, "outside/evil.ts"),
      path.join(e2eRoot, "evil.spec.ts"),
    );
    assert.throws(() => mod.listSpecFiles(e2eRoot), /symlink/i);
    const payload = mod.verifyPages({
      pages: ["e2e/pages/accounting-page.ts"],
      e2eRoot,
    });
    assert.equal(payload.ok, false, JSON.stringify(payload));
    assert.match(String(payload.error || ""), /symlink/i);
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
});

test("RED: symlink directory and symlink e2e root fail closed", async () => {
  const mod = await loadModule();
  const { root, e2eRoot } = makeFixture({
    "e2e/pages/accounting-page.ts": "export class AccountingPage {}\n",
    "e2e/good.spec.ts":
      'import { AccountingPage } from "./pages/accounting-page";\n',
    "outside/dir/nested.spec.ts":
      'import { AccountingPage } from "../pages/accounting-page";\n',
  });
  try {
    symlinkSync(path.join(root, "outside/dir"), path.join(e2eRoot, "linked-dir"));
    assert.throws(() => mod.listSpecFiles(e2eRoot), /symlink/i);
    const dirPayload = mod.verifyPages({
      pages: ["e2e/pages/accounting-page.ts"],
      e2eRoot,
    });
    assert.equal(dirPayload.ok, false);
    assert.match(String(dirPayload.error || ""), /symlink/i);

    const linkRootParent = mkdtempSync(path.join(tmpdir(), "ae-e2e-linkroot-"));
    const realE2e = path.join(linkRootParent, "real-e2e");
    const linkedE2e = path.join(linkRootParent, "e2e");
    mkdirSync(path.join(realE2e, "pages"), { recursive: true });
    writeFileSync(
      path.join(realE2e, "pages/accounting-page.ts"),
      "export class AccountingPage {}\n",
      "utf8",
    );
    writeFileSync(
      path.join(realE2e, "good.spec.ts"),
      'import { AccountingPage } from "./pages/accounting-page";\n',
      "utf8",
    );
    symlinkSync(realE2e, linkedE2e);
    try {
      assert.throws(() => mod.listSpecFiles(linkedE2e), /symlink/i);
      const rootPayload = mod.verifyPages({
        pages: ["e2e/pages/accounting-page.ts"],
        e2eRoot: linkedE2e,
      });
      assert.equal(rootPayload.ok, false);
      assert.match(String(rootPayload.error || ""), /symlink/i);
    } finally {
      rmSync(linkRootParent, { recursive: true, force: true });
    }
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
});

test("RED: outside-root enumerated path via custom listSpecs fails closed", async () => {
  const mod = await loadModule();
  const { root, e2eRoot } = makeFixture({
    "e2e/pages/accounting-page.ts": "export class AccountingPage {}\n",
    "e2e/good.spec.ts":
      'import { AccountingPage } from "./pages/accounting-page";\n',
    "outside/forged.spec.ts":
      'const load = () => import("./pages/accounting-page");\n',
  });
  try {
    const good = path.join(e2eRoot, "good.spec.ts");
    const outside = path.join(root, "outside/forged.spec.ts");
    // Outside enumerated path must fail closed even when a good consumer exists.
    const payload = mod.verifyPages({
      pages: ["e2e/pages/accounting-page.ts"],
      e2eRoot,
      listSpecs: () => [good, outside],
    });
    assert.equal(payload.ok, false, JSON.stringify(payload));
    assert.match(String(payload.error || ""), /escape|noncanonical|outside|root/i);
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
});

test("RED: raw /./ absolute spelling via custom listSpecs fails closed", async () => {
  const mod = await loadModule();
  const { root, e2eRoot } = makeFixture({
    "e2e/pages/accounting-page.ts": "export class AccountingPage {}\n",
    "e2e/good.spec.ts":
      'import { AccountingPage } from "./pages/accounting-page";\n',
  });
  try {
    const dotted = e2eRoot + "/./good.spec.ts";
    assert.notEqual(path.resolve(dotted), dotted);
    assert.equal(path.relative(e2eRoot, dotted), "good.spec.ts");
    const payload = mod.verifyPages({
      pages: ["e2e/pages/accounting-page.ts"],
      e2eRoot,
      listSpecs: () => [dotted],
    });
    assert.equal(payload.ok, false, JSON.stringify(payload));
    assert.match(
      String(payload.error || ""),
      /noncanonical|raw|spelling|absolute|canonical/i,
    );
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
});

test("RED: control and whitespace spec paths fail before identity authority", async () => {
  const mod = await loadModule();
  const { root, e2eRoot } = makeFixture({
    "e2e/pages/accounting-page.ts": "export class AccountingPage {}\n",
  });
  try {
    for (const unsafe of ["\n", "\x01", "\x7f", " "]) {
      const absolutePath = path.join(e2eRoot, `bad${unsafe}name.spec.ts`);
      writeFileSync(
        absolutePath,
        'import { AccountingPage } from "./pages/accounting-page";\n',
      );
      const payload = mod.verifyPages({
        pages: ["e2e/pages/accounting-page.ts"],
        e2eRoot,
        listSpecs: () => [absolutePath],
      });
      assert.equal(payload.ok, false, JSON.stringify(payload));
      assert.match(String(payload.error || ""), /control|whitespace|canonical/i);
    }
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
});

test("RED: custom listSpecs symlink path is rejected by descriptor-safe open", async () => {
  const mod = await loadModule();
  const { root, e2eRoot } = makeFixture({
    "e2e/pages/accounting-page.ts": "export class AccountingPage {}\n",
    "e2e/good.spec.ts":
      'import { AccountingPage } from "./pages/accounting-page";\n',
    "outside/valid-import.ts":
      'import { AccountingPage } from "./pages/accounting-page";\n',
  });
  try {
    const good = path.join(e2eRoot, "good.spec.ts");
    const outside = path.join(root, "outside/valid-import.ts");
    unlinkSync(good);
    symlinkSync(outside, good);
    const payload = mod.verifyPages({
      pages: ["e2e/pages/accounting-page.ts"],
      e2eRoot,
      listSpecs: () => [good],
    });
    assert.equal(payload.ok, false, JSON.stringify(payload));
    assert.match(
      String(payload.error || ""),
      /symlink|ELOOP|nofollow|trusted|open|regular/i,
    );
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
});

test("RED: mid-loop swap to outside symlink before trusted read fails closed", async () => {
  const mod = await loadModule();
  const { root, e2eRoot } = makeFixture({
    "e2e/pages/accounting-page.ts": "export class AccountingPage {}\n",
    "e2e/good.spec.ts":
      'import { AccountingPage } from "./pages/accounting-page";\n',
    "outside/valid-import.ts":
      'import { AccountingPage } from "./pages/accounting-page";\n',
  });
  try {
    const good = path.join(e2eRoot, "good.spec.ts");
    const outside = path.join(root, "outside/valid-import.ts");
    let swapped = false;
    const payload = mod.verifyPages({
      pages: ["e2e/pages/accounting-page.ts"],
      e2eRoot,
      listSpecs: () => [good],
      beforeTrustedRead: () => {
        unlinkSync(good);
        symlinkSync(outside, good);
        swapped = true;
      },
    });
    assert.equal(swapped, true);
    assert.equal(payload.ok, false, JSON.stringify(payload));
    assert.match(
      String(payload.error || ""),
      /symlink|ELOOP|nofollow|trusted|open|regular|escape/i,
    );
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
});

test("real-file relative static-import consumers remain ok after symlink policy", async () => {
  const mod = await loadModule();
  const { root, e2eRoot } = makeFixture({
    "e2e/pages/accounting-page.ts": "export class AccountingPage {}\n",
    "e2e/good.spec.ts":
      'import { AccountingPage } from "./pages/accounting-page";\n',
    "e2e/subdir/nested.spec.ts":
      'import { AccountingPage } from "../pages/accounting-page";\n',
  });
  try {
    const listed = mod.listSpecFiles(e2eRoot);
    assert.equal(listed.length, 2);
    const payload = mod.verifyPages({
      pages: ["e2e/pages/accounting-page.ts"],
      e2eRoot,
    });
    assert.equal(payload.ok, true, JSON.stringify(payload));
    assert.equal(payload.pages[0].blocking.length, 0);
    assert.ok(payload.pages[0].consumers.length >= 2);
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
});
