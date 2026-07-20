// scripts/design-system-audit.test.mjs
//
// design-system-audit.mjs の回帰テスト。Node 組み込みの test runner を使う
// （check-eslint-disable-rationale.test.mjs と同じ no-new-dependency 方針）。
//
// 実行: node --test scripts/design-system-audit.test.mjs

import test from "node:test";
import assert from "node:assert/strict";
import { mkdtempSync, mkdirSync, writeFileSync, rmSync } from "node:fs";
import { tmpdir } from "node:os";
import path from "node:path";
import { execFileSync } from "node:child_process";
import {
  isTestFile,
  isLeafRouteFile,
  isDesignTokensFile,
  isColorMapFile,
  checkC1,
  checkC3,
  checkC5,
  checkC6,
  checkC7,
  checkC8,
  checkC9,
  C8_ALLOWLIST,
  collectViolations,
} from "./design-system-audit.mjs";

test("isTestFile: *.test.tsx を検出する", () => {
  assert.equal(isTestFile("Foo.test.tsx"), true);
  assert.equal(isTestFile("Foo.test.ts"), true);
  assert.equal(isTestFile("Foo.tsx"), false);
});

test("isLeafRouteFile: routes/ または pages/ 配下を検出する", () => {
  assert.equal(isLeafRouteFile(path.join("accounting", "routes", "AccountingList.tsx")), true);
  assert.equal(isLeafRouteFile(path.join("lstep", "pages", "CheckupSyncPage.tsx")), true);
  assert.equal(isLeafRouteFile(path.join("accounting", "components", "Foo.tsx")), false);
});

test("checkC1: C.accent / #0075DE / #2383E2 を検出する", () => {
  const text = [
    'const x = <div className={C.accent} />;',
    'const y = "#0075DE";',
    'const z = "#2383E2";',
    'const ok = <div className={C.text} />;',
  ].join("\n");
  const violations = checkC1(text);
  assert.equal(violations.length, 3);
  assert.equal(violations[0].lineNumber, 1);
  assert.equal(violations[1].lineNumber, 2);
  assert.equal(violations[2].lineNumber, 3);
});

test("checkC1: legacy accent が無ければ 0 件", () => {
  const text = 'const ok = <div className={C.brand} />;';
  assert.equal(checkC1(text).length, 0);
});

test("checkC1: 単語境界判定により C.accentBrand は誤検知しない（FE3-2）", () => {
  const text = 'className={`size-4 cursor-pointer ${C.accentBrand}`}';
  assert.equal(checkC1(text).length, 0);
});

test("checkC3: 引用符付き hex リテラルを検出する", () => {
  const text = [
    "const bad = '#FF0000';",
    'const alsoBad = "#ABC";',
    "const templated = `#123456`;",
  ].join("\n");
  const violations = checkC3(text);
  assert.equal(violations.length, 3);
});

test("checkC3: 引用符無しの issue 番号コメント（#158 等）は誤検知しない", () => {
  const text = [
    "// #158 で追加された対応",
    "// see #2383 for details",
    "const ok = C.text;",
  ].join("\n");
  assert.equal(checkC3(text).length, 0);
});

test("checkC5: colorVariant=\"brand\" 以外を検出する", () => {
  const text = [
    '<PrimaryButton colorVariant="brand">保存</PrimaryButton>',
    '<PrimaryButton colorVariant="default">保存</PrimaryButton>',
  ].join("\n");
  const violations = checkC5(text);
  assert.equal(violations.length, 1);
  assert.equal(violations[0].value, "default");
  assert.equal(violations[0].lineNumber, 2);
});

test("checkC5: 全て brand なら 0 件", () => {
  const text = '<SubmitButton colorVariant="brand">締める</SubmitButton>';
  assert.equal(checkC5(text).length, 0);
});

test("checkC6: rgba()/rgb()/hsla()/hsl() の直値を検出する（FE3-2 新設）", () => {
  const text = [
    'hover:bg-[rgba(242,241,238,0.5)]',
    'backgroundColor: `rgb(${r}, ${g}, ${b})`,',
    'border: "1px solid hsla(0, 0%, 0%, 0.1)"',
    'color: "hsl(200, 50%, 50%)"',
    'const ok = C.text;',
  ].join("\n");
  const violations = checkC6(text);
  assert.equal(violations.length, 4);
});

test("checkC6: rgba/hsla が無ければ 0 件", () => {
  const text = 'const ok = <div className={C.text} />;';
  assert.equal(checkC6(text).length, 0);
});

test("isDesignTokensFile: src/lib/design-tokens.ts のみ true", () => {
  assert.equal(isDesignTokensFile(path.join("src", "lib", "design-tokens.ts")), true);
  assert.equal(isDesignTokensFile(path.join("src", "lib", "utils.ts")), false);
});

test("isColorMapFile: src/hooks/use-reservation-type-color-map.ts のみ true", () => {
  assert.equal(isColorMapFile(path.join("src", "hooks", "use-reservation-type-color-map.ts")), true);
  assert.equal(isColorMapFile(path.join("src", "hooks", "use-other.ts")), false);
});

// --- 統合テスト: 一時フィクスチャディレクトリで collectViolations / CLI 実行を検証 ---
// 本番 src/features ツリーは一切変更しない。

function makeFixture() {
  const root = mkdtempSync(path.join(tmpdir(), "design-audit-fixture-"));
  const routesDir = path.join(root, "src", "features", "widget", "routes");
  mkdirSync(routesDir, { recursive: true });
  return { root, routesDir };
}

test("collectViolations: クリーンな fixture は 0 件", async () => {
  const { root, routesDir } = makeFixture();
  try {
    writeFileSync(
      path.join(routesDir, "WidgetPage.tsx"),
      [
        'import { C } from "@/lib/design-tokens";',
        "// #158 で追加",
        'export function WidgetPage() {',
        '  return <PrimaryButton colorVariant="brand" className={C.text}>OK</PrimaryButton>;',
        "}",
      ].join("\n"),
    );
    const result = await collectViolations(root);
    assert.equal(result.c1.length, 0);
    assert.equal(result.c3.length, 0);
    assert.equal(result.c5.length, 0);
    assert.equal(result.c6.length, 0);
    assert.equal(result.c7.length, 0);
    assert.equal(result.c9.length, 0);
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
});

test("collectViolations: 意図的違反 fixture は各カテゴリで検出される", async () => {
  const { root, routesDir } = makeFixture();
  try {
    writeFileSync(
      path.join(routesDir, "BrokenPage.tsx"),
      [
        'import { C } from "@/lib/design-tokens";',
        'const legacy = C.accent;',
        'const hex = "#FF0000";',
        'const raw = "rgba(0,0,0,0.5)";',
        'export function BrokenPage() {',
        '  return <PrimaryButton colorVariant="default" className="rounded-[4px]" maxWidth="max-w-full">NG</PrimaryButton>;',
        "}",
      ].join("\n"),
    );
    const result = await collectViolations(root);
    assert.equal(result.c1.length, 1);
    assert.equal(result.c3.length, 1);
    assert.equal(result.c5.length, 1);
    assert.equal(result.c6.length, 1);
    assert.equal(result.c7.length, 1);
    assert.equal(result.c9.length, 1);
    // BrokenPage は PageLayout 無し・allowlist 外 → C8
    assert.equal(result.c8.length, 1);
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
});

test("collectViolations: C3/C5/C6 は *.test.* を除外する", async () => {
  const { root, routesDir } = makeFixture();
  try {
    writeFileSync(
      path.join(routesDir, "WidgetPage.test.tsx"),
      [
        'const hex = "#FF0000"; const raw = "rgba(0,0,0,0.5)"; // test file — C3/C6 から除外される',
        'render(<PrimaryButton colorVariant="default">opt-out検証</PrimaryButton>); // C5 も除外される',
      ].join("\n"),
    );
    const result = await collectViolations(root);
    assert.equal(result.c3.length, 0);
    assert.equal(result.c5.length, 0);
    assert.equal(result.c6.length, 0);
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
});

test("collectViolations（FE3-2）: components/ 配下も走査対象になる（旧 routes/pages 限定を撤廃）", async () => {
  const root = mkdtempSync(path.join(tmpdir(), "design-audit-fixture-"));
  try {
    const componentsDir = path.join(root, "src", "features", "widget", "components");
    mkdirSync(componentsDir, { recursive: true });
    writeFileSync(path.join(componentsDir, "Widget.tsx"), 'const legacy = C.accent;');
    const result = await collectViolations(root);
    assert.equal(result.c1.length, 1);
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
});

test("collectViolations（FE3-2）: liff/src・line-reserve/src も走査対象になる", async () => {
  const root = mkdtempSync(path.join(tmpdir(), "design-audit-fixture-"));
  try {
    const liffDir = path.join(root, "liff", "src", "pages");
    const lineReserveDir = path.join(root, "line-reserve", "src", "pages");
    mkdirSync(liffDir, { recursive: true });
    mkdirSync(lineReserveDir, { recursive: true });
    writeFileSync(path.join(liffDir, "LiffPage.tsx"), 'const legacy = C.accent;');
    writeFileSync(path.join(lineReserveDir, "ReservePage.tsx"), 'const hex = "#FF0000";');
    const result = await collectViolations(root);
    assert.equal(result.c1.length, 1);
    assert.equal(result.c3.length, 1);
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
});

test("collectViolations（FE3-2）: design-tokens.ts は C1/C3/C6 の allowlist（legacy トークン定義元は自己検出しない）", async () => {
  const root = mkdtempSync(path.join(tmpdir(), "design-audit-fixture-"));
  try {
    const libDir = path.join(root, "src", "lib");
    mkdirSync(libDir, { recursive: true });
    writeFileSync(
      path.join(libDir, "design-tokens.ts"),
      'export const C = { text: "#000000", accent: "#0075DE" }; export const P = { h: "rgba(0,0,0,0.5)" };',
    );
    const result = await collectViolations(root);
    assert.equal(result.c1.length, 0);
    assert.equal(result.c3.length, 0);
    assert.equal(result.c6.length, 0);
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
});

test("collectViolations（FE3-2）: use-reservation-type-color-map.ts は C6 の allowlist", async () => {
  const root = mkdtempSync(path.join(tmpdir(), "design-audit-fixture-"));
  try {
    const hooksDir = path.join(root, "src", "hooks");
    mkdirSync(hooksDir, { recursive: true });
    writeFileSync(
      path.join(hooksDir, "use-reservation-type-color-map.ts"),
      'const style = { backgroundColor: `rgba(${r},${g},${b},0.1)` };',
    );
    const result = await collectViolations(root);
    assert.equal(result.c6.length, 0);
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
});

test("CLI: 違反ありの fixture は exit code 1、file:line を stdout に出す", () => {
  const { root, routesDir } = makeFixture();
  try {
    writeFileSync(
      path.join(routesDir, "BrokenPage.tsx"),
      'const legacy = C.accent;',
    );
    const scriptPath = path.join(import.meta.dirname, "design-system-audit.mjs");
    assert.throws(() => {
      execFileSync("node", [scriptPath, "--cwd", root], { encoding: "utf-8" });
    }, (err) => {
      assert.equal(err.status, 1);
      assert.match(err.stdout, /BrokenPage\.tsx:1/);
      assert.match(err.stdout, /FAIL/);
      return true;
    });
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
});

test("CLI: 違反なしの fixture は exit code 0", () => {
  const { root, routesDir } = makeFixture();
  try {
    writeFileSync(
      path.join(routesDir, "CleanPage.tsx"),
      'export function CleanPage() { return <PageLayout title="ok">x</PageLayout>; }',
    );
    const scriptPath = path.join(import.meta.dirname, "design-system-audit.mjs");
    const stdout = execFileSync("node", [scriptPath, "--cwd", root], { encoding: "utf-8" });
    assert.match(stdout, /PASS/);
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
});


test("checkC7: maxWidth 生値を検出する", () => {
  const text = [
    'maxWidth="max-w-full"',
    'maxWidth="max-w-[1400px]"',
    'maxWidth={LAYOUT.pageContentMaxWidth.full}',
    'maxWidth="max-w-3xl"',
  ].join("\n");
  const violations = checkC7(text);
  assert.equal(violations.length, 2);
});

test("checkC9: rounded 任意値を検出する", () => {
  const text = [
    'className="rounded-[4px]"',
    'className="rounded-t-[3px]"',
    'className="rounded-xs"',
  ].join("\n");
  const violations = checkC9(text);
  assert.equal(violations.length, 2);
});

test("checkC8: allowlist / PageLayout / 欠落を判定する", () => {
  assert.equal(C8_ALLOWLIST.size, 15);
  assert.equal(checkC8("Login.tsx", "export function Login() { return null; }").length, 0);
  assert.equal(checkC8("FooPage.tsx", "return <PageLayout title=\"x\">").length, 0);
  assert.equal(checkC8("FooPage.tsx", "return <MasterListPage />").length, 0);
  assert.equal(checkC8("FooPage.tsx", "export function Foo() { return null; }").length, 1);
});

test("collectViolations（FE8-5）: C8 は allowlist basename を除外する", async () => {
  const root = mkdtempSync(path.join(tmpdir(), "design-audit-fixture-"));
  try {
    const routesDir = path.join(root, "src", "features", "auth", "routes");
    mkdirSync(routesDir, { recursive: true });
    writeFileSync(path.join(routesDir, "Login.tsx"), "export function Login() { return null; }");
    writeFileSync(path.join(routesDir, "OrphanPage.tsx"), "export function OrphanPage() { return null; }");
    const result = await collectViolations(root);
    assert.equal(result.c8.length, 1);
    assert.match(result.c8[0].file, /OrphanPage\.tsx$/);
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
});
