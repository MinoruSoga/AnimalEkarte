// scripts/design-system-audit.test.mjs
//
// design-system-audit.mjs の回帰テスト。Node 組み込みの test runner を使う
// （check-eslint-disable-rationale.test.mjs と同じ no-new-dependency 方針）。
//
// 実行（cwd 非依存）:
//   node --test scripts/design-system-audit.test.mjs
//   node --test frontend/scripts/design-system-audit.test.mjs  （make ci / リポジトリ直下）

import test from "node:test";
import assert from "node:assert/strict";
import {
  mkdtempSync,
  mkdirSync,
  readFileSync,
  readdirSync,
  writeFileSync,
  rmSync,
} from "node:fs";
import { tmpdir } from "node:os";
import path from "node:path";
import { execFileSync } from "node:child_process";
import * as audit from "./design-system-audit.mjs";
import {
  isTestFile,
  isLeafRouteFile,
  isDesignTokensFile,
  isBrandTokensFile,
  isColorMapFile,
  checkC1,
  checkC3,
  checkC5,
  checkC6,
  checkC7,
  checkC8,
  checkC9,
  checkC10,
  checkC11,
  checkC12,
  checkC13,
  checkC14,
  C8_ALLOWLIST,
  C8_PAGE_ALLOWLIST,
  C8_ROUTE_HELPER_ALLOWLIST,
  collectViolations,
} from "./design-system-audit.mjs";

/** パッケージルート。make ci はリポジトリ直下でこのファイルを実行するため cwd に依存しない。 */
const FRONTEND_ROOT = path.join(import.meta.dirname, "..");

test("production source は brand focus ring を再び半透明化しない", () => {
  const pendingDirectories = [path.join(FRONTEND_ROOT, "src")];

  while (pendingDirectories.length > 0) {
    const directory = pendingDirectories.pop();
    assert.ok(directory);

    for (const entry of readdirSync(directory, { withFileTypes: true })) {
      const entryPath = path.join(directory, entry.name);
      if (entry.isDirectory()) {
        pendingDirectories.push(entryPath);
      } else if (
        /\.(?:css|ts|tsx)$/.test(entry.name) &&
        !/\.test\.(?:ts|tsx)$/.test(entry.name)
      ) {
        assert.doesNotMatch(
          readFileSync(entryPath, "utf-8"),
          /\b(?:outline|ring)-ring\/50\b/,
          entryPath,
        );
      }
    }
  }
});

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

test("globals.css は brand と semantic primary を同じ teal 値で保持する", () => {
  const source = readFileSync(
    path.join(FRONTEND_ROOT, "src", "styles", "globals.css"),
    "utf-8",
  );

  assert.match(source, /--brand:\s*#038b94;/i);
  assert.match(source, /--brand-active:\s*#027078;/i);
  assert.match(source, /--primary:\s*#038b94;/i);
  assert.match(source, /--primary-active:\s*#027078;/i);
  assert.match(source, /--primary-foreground:\s*#ffffff;/i);
  assert.match(source, /--ring:\s*#038b94;/i);
  assert.match(source, /--sidebar-primary:\s*#038b94;/i);
  assert.match(source, /--sidebar-primary-foreground:\s*#ffffff;/i);
  assert.match(source, /--sidebar-ring:\s*#038b94;/i);
  assert.match(source, /--shadow-focus-primary:\s*0 0 0 2px #038b94;/i);
});

test("checkC1: C.accent / 旧 accent #2383E2 を検出し、brand/primary teal と臨床 blue を許可する", () => {
  const text = [
    'const x = <div className={C.accent} />;',
    'const y = "#0075DE";',
    'const z = "#2383E2";',
    'const w = "#005BAB";',
    'const ok = "bg-[#038B94] hover:bg-[#027078]";',
  ].join("\n");
  const violations = checkC1(text);
  assert.equal(violations.length, 2);
  assert.equal(violations[0].lineNumber, 1);
  assert.equal(violations[1].lineNumber, 3);
});

test("checkC1: brand/primary teal と臨床 blue は legacy accent として扱わない", () => {
  const text = 'const ok = "bg-[#038B94] hover:bg-[#027078] bg-[#0075DE] hover:bg-[#005BAB]";';
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

test("checkC3: Tailwind arbitrary color 内の hex を検出する", () => {
  const text = [
    'className="bg-[#ff0000]"',
    'className="placeholder:text-[#A39E98]"',
    'className="hover:border-[#ABC]/50"',
  ].join("\n");

  const violations = checkC3(text);
  assert.equal(violations.length, 3);
  assert.deepEqual(violations.map(({ lineNumber }) => lineNumber), [1, 2, 3]);
});

test("checkC3: 引用符無しの issue 番号コメント（#158 等）は誤検知しない", () => {
  const text = [
    "// #158 で追加された対応",
    "// see #2383 for details",
    "const ok = C.text;",
  ].join("\n");
  assert.equal(checkC3(text).length, 0);
});

test("checkC5: 未定義の colorVariant を検出する", () => {
  const text = [
    '<PrimaryButton colorVariant="brand">保存</PrimaryButton>',
    '<PrimaryButton colorVariant="primary">保存</PrimaryButton>',
    '<PrimaryButton colorVariant="default">保存</PrimaryButton>',
    '<PrimaryButton colorVariant="accent">保存</PrimaryButton>',
  ].join("\n");
  const violations = checkC5(text);
  assert.equal(violations.length, 1);
  assert.equal(violations[0].value, "accent");
  assert.equal(violations[0].lineNumber, 4);
});

test("checkC5: primary・brand・default は 0 件", () => {
  const text = [
    '<SubmitButton colorVariant="primary">保存</SubmitButton>',
    '<SubmitButton colorVariant="brand">ログイン</SubmitButton>',
    '<SubmitButton colorVariant="default">互換</SubmitButton>',
  ].join("\n");
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

test("isBrandTokensFile: src/shared-liff/brand-tokens.ts のみ true", () => {
  assert.equal(
    isBrandTokensFile(path.join("src", "shared-liff", "brand-tokens.ts")),
    true,
  );
  assert.equal(
    isBrandTokensFile(path.join("src", "features", "widget", "brand-tokens.ts")),
    false,
  );
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
        '  return <PrimaryButton colorVariant="accent" className="rounded-[4px]" maxWidth="max-w-full">NG</PrimaryButton>;',
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

test("collectViolations: 本体 route の named color と非仕様 spacing を C15/C16 に集計する", async () => {
  const { root, routesDir } = makeFixture();
  try {
    writeFileSync(
      path.join(routesDir, "WidgetPage.tsx"),
      [
        "export function WidgetPage() {",
        '  return <PageLayout title="ok">',
        '    <div className="bg-white p-5">surface</div>',
        '    <span className="text-white gap-5">label</span>',
        '    <div className="bg-black/30">overlay</div>',
        "  </PageLayout>;",
        "}",
      ].join("\n"),
    );

    const result = await collectViolations(root);
    assert.ok(Array.isArray(result.c15), "C15 named color の集計結果が必要");
    assert.equal(result.c15.length, 3);
    assert.ok(Array.isArray(result.c16), "C16 非仕様 spacing の集計結果が必要");
    assert.equal(result.c16.length, 2);
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
});

test("collectViolations: CSS の legacy accent は C1、直接 shadow は C17 に集計する", async () => {
  const root = mkdtempSync(path.join(tmpdir(), "design-audit-fixture-"));
  try {
    const featureStylesDir = path.join(root, "src", "features", "widget", "styles");
    const globalStylesDir = path.join(root, "src", "styles");
    mkdirSync(featureStylesDir, { recursive: true });
    mkdirSync(globalStylesDir, { recursive: true });
    writeFileSync(
      path.join(featureStylesDir, "widget.css"),
      [
        ".card { box-shadow: 0 1px 2px currentColor; }",
        ".icon { filter: drop-shadow(0 2px 4px currentColor); }",
      ].join("\n"),
    );
    writeFileSync(
      path.join(globalStylesDir, "globals.css"),
      [
        ":root {",
        "  --legacy-accent: #2383E2;",
        "  --shadow-level1: 0 1px 2px currentColor;",
        "  --shadow-level2: 0 4px 18px currentColor;",
        "}",
      ].join("\n"),
    );

    const result = await collectViolations(root);
    assert.equal(result.c1.length, 1);
    assert.deepEqual(
      result.c1.map(({ file, lineNumber }) => ({ file, lineNumber })),
      [{ file: path.join("src", "styles", "globals.css"), lineNumber: 2 }],
    );
    assert.ok(Array.isArray(result.c17), "C17 CSS shadow の集計結果が必要");
    assert.equal(result.c17.length, 2);
    assert.deepEqual(
      result.c17.map(({ file, lineNumber }) => ({ file, lineNumber })),
      [
        { file: path.join("src", "features", "widget", "styles", "widget.css"), lineNumber: 1 },
        { file: path.join("src", "features", "widget", "styles", "widget.css"), lineNumber: 2 },
      ],
    );
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

test("collectViolations: design-tokens.ts も C1 の legacy accent 禁止対象にする", async () => {
  const root = mkdtempSync(path.join(tmpdir(), "design-audit-fixture-"));
  try {
    const libDir = path.join(root, "src", "lib");
    mkdirSync(libDir, { recursive: true });
    writeFileSync(
      path.join(libDir, "design-tokens.ts"),
      'export const C = { legacyAccent: "#2383E2" }; export const P = { h: "rgba(0,0,0,0.5)" };',
    );
    const result = await collectViolations(root);
    assert.equal(result.c1.length, 1);
    assert.equal(result.c3.length, 0);
    assert.equal(result.c6.length, 0);
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
});

test("collectViolations: shared-liff/brand-tokens.ts の hex は C3 対象外", async () => {
  const root = mkdtempSync(path.join(tmpdir(), "design-audit-fixture-"));
  try {
    const brandDir = path.join(root, "src", "shared-liff");
    mkdirSync(brandDir, { recursive: true });
    writeFileSync(
      path.join(brandDir, "brand-tokens.ts"),
      "export const NOAH_BRAND_COLORS = { teal: '#008B94' };",
    );
    const spoofDir = path.join(root, "src", "features", "widget");
    mkdirSync(spoofDir, { recursive: true });
    writeFileSync(
      path.join(spoofDir, "brand-tokens.ts"),
      "export const FAKE = { teal: '#008B94' };",
    );
    const result = await collectViolations(root);
    assert.equal(result.c3.length, 1);
    assert.match(result.c3[0].file, /features/);
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

test("checkC10: Tailwind 既定影と shadow 任意値を検出し、level トークンは許可する", () => {
  const text = [
    'className="shadow-sm"',
    'className="data-[variant=outline]:shadow-xs"',
    'className="shadow-[0_1px_2px_rgba(0,0,0,0.1)]"',
    'className="shadow-level1"',
    'className="shadow-level2"',
    'className="shadow-btn shadow-none"',
    'className="drop-shadow-md"',
  ].join("\n");
  const violations = checkC10(text);
  assert.equal(violations.length, 4);
  assert.deepEqual(violations.map(({ lineNumber }) => lineNumber), [1, 2, 3, 7]);
});

test("checkC11: font-size 任意値を検出し、ロールクラスは許可する", () => {
  const text = [
    'className="text-[11px]"',
    'className="text-[0.6875rem]"',
    'className="text-2xs"',
    'className="text-sm text-[24px]"',
  ].join("\n");
  const violations = checkC11(text);
  assert.equal(violations.length, 3);
});

test("checkC12: 非仕様サイズ段を本体で検出し、別アプリは除外する", () => {
  const text = [
    'className="text-lg"',
    'className="text-2xl"',
    'className="text-9xl"',
    'className="text-xl text-heading-1 text-heading-2 text-heading-3"',
  ].join("\n");

  assert.equal(checkC12(text, path.join("src", "features", "widget", "Widget.tsx")).length, 3);
  assert.equal(checkC12(text, path.join("liff", "src", "pages", "Widget.tsx")).length, 0);
  assert.equal(checkC12(text, path.join("src", "shared-liff", "Widget.tsx")).length, 0);
  assert.equal(
    checkC12(text, path.join("src", "features", "shared-liff-tools", "Widget.tsx")).length,
    3,
  );
});

test("checkC13: ink 黒アルファを検出し、4段トークンは許可する", () => {
  const text = [
    'className="text-[#000000]/60"',
    'className="placeholder:text-[rgba(0,0,0,0.3)]"',
    'className={C.text60}',
  ].join("\n");

  assert.equal(checkC13(text).length, 2);
});

test("checkC14: built-in・任意値・負値・shorthand の tracking を本体で検出し、別アプリと PrintArea を除外する", () => {
  const text = [
    'className="tracking-wide"',
    'className="uppercase tracking-wider"',
    'className="tracking-widest text-2xs"',
    'className="tracking-tight"',
    'className="tracking-normal"',
    'className="tracking-[0.12em]"',
    'className="tracking-[var(--tracking-compact)]"',
    'className="tracking-[var(--tracking-compact-sm)]"',
    'className="tracking-[var(--tracking-compact-xs)]"',
    'className="-tracking-wide"',
    'className="tracking-(--custom-letter-spacing)"',
  ].join("\n");

  const violations = checkC14(text, path.join("src", "features", "widget", "Widget.tsx"));
  assert.equal(violations.length, 11);
  assert.equal(checkC14(text, path.join("liff", "src", "pages", "Widget.tsx")).length, 0);
  assert.equal(checkC14(text, path.join("line-reserve", "src", "pages", "Widget.tsx")).length, 0);
  assert.equal(checkC14(text, path.join("src", "shared-liff", "Widget.tsx")).length, 0);
  assert.equal(
    checkC14(text, path.join("src", "features", "shared-liff-tools", "Widget.tsx")).length,
    11,
  );
  assert.equal(
    checkC14(
      text,
      path.join("src", "features", "examinations", "components", "ExaminationPrintArea.tsx"),
    ).length,
    0,
  );
});

test("checkC15: 本体 routes/pages の直接 named color を検出し、トークンと別アプリは許可する", () => {
  assert.equal(typeof audit.checkC15, "function");

  const text = [
    'className="bg-white"',
    'className="text-white"',
    'className="fixed inset-0 bg-black/30"',
    'className="text-black"',
    'className="bg-black"',
    'className="border-white"',
    'className="hover:border-black/50"',
    'className="placeholder:text-black"',
    'className={`${C.bgWhite} ${C.textWhite} ${C.bgOverlay}`}',
  ].join("\n");
  const mainRoute = path.join("src", "features", "widget", "routes", "WidgetPage.tsx");
  const mainPage = path.join("src", "features", "widget", "pages", "WidgetPage.tsx");

  assert.equal(audit.checkC15(text, mainRoute).length, 8);
  assert.equal(audit.checkC15(text, mainPage).length, 8);
  assert.equal(
    audit.checkC15(text, path.join("src", "features", "widget", "components", "Widget.tsx")).length,
    0,
  );
  assert.equal(audit.checkC15(text, path.join("liff", "src", "pages", "Widget.tsx")).length, 0);
  assert.equal(audit.checkC15(text, path.join("line-reserve", "src", "pages", "Widget.tsx")).length, 0);
});

test("checkC16: DESIGN.md の spacing scale にない p/m/gap/space の *-5 を本体で検出する", () => {
  assert.equal(typeof audit.checkC16, "function");

  const text = [
    'className="p-5"',
    'className="px-5 py-5"',
    'className="mt-5 mb-5"',
    'className="gap-5 space-y-5"',
    'className="-m-5"',
    'className="-mx-5 -mt-5"',
    'className="p-[20px]"',
    'className="gap-[1.25rem]"',
    'className="space-y-[20px]"',
    'className="p-4 px-6 gap-4 size-5 w-5 h-5"',
  ].join("\n");
  const mainComponent = path.join("src", "features", "widget", "components", "Widget.tsx");

  assert.equal(audit.checkC16(text, mainComponent).length, 9);
  assert.equal(audit.checkC16(text, path.join("liff", "src", "pages", "Widget.tsx")).length, 0);
  assert.equal(audit.checkC16(text, path.join("src", "shared-liff", "Widget.tsx")).length, 0);
  assert.equal(
    audit.checkC16(
      text,
      path.join("src", "features", "medical-records", "components", "MedicalRecordPrintView.tsx"),
    ).length,
    0,
  );
});

test("checkC17: CSS の直接 box-shadow / drop-shadow を検出し --shadow-* token 定義は許可する", () => {
  assert.equal(typeof audit.checkC17, "function");

  const css = [
    ".clean { filter: blur(2px); box-sizing: border-box; }",
    ".card { box-shadow: 0 1px 2px currentColor; }",
    ".icon { filter: drop-shadow(0 2px 4px currentColor); }",
  ].join("\n");
  const globalsCss = [
    ":root {",
    "  --shadow-level1: 0 1px 2px currentColor;",
    "  --shadow-level2: 0 4px 18px currentColor;",
    "}",
    ".raw { box-shadow: 0 1px 2px currentColor; }",
  ].join("\n");

  const violations = audit.checkC17(
    css,
    path.join("src", "features", "widget", "styles", "widget.css"),
  );
  assert.equal(violations.length, 2);
  assert.deepEqual(violations.map(({ lineNumber }) => lineNumber), [2, 3]);

  const globalsViolations = audit.checkC17(
    globalsCss,
    path.join("src", "styles", "globals.css"),
  );
  assert.equal(globalsViolations.length, 1);
  assert.equal(globalsViolations[0].lineNumber, 5);
  assert.equal(
    audit.checkC17(css, path.join("liff", "src", "styles", "widget.css")).length,
    0,
  );
});

test("checkC18: TableCell / TableHead の非仕様 override を検出し正確な行番号を返す", () => {
  assert.equal(typeof audit.checkC18, "function");

  const text = [
    "<Table>",
    "  <TableHeader>",
    "    <TableRow>",
    '      <TableHead className="text-base">A</TableHead>',
    '      <TableHead className="text-xs">B</TableHead>',
    '      <TableHead className="font-medium">C</TableHead>',
    '      <TableHead className="py-1">D</TableHead>',
    '      <TableHead className="py-2">E</TableHead>',
    '      <TableHead className="py-2.5">F</TableHead>',
    "      <TableHead",
    '        className="p-0"',
    "      >G</TableHead>",
    "    </TableRow>",
    "  </TableHeader>",
    "  <TableBody>",
    "    <TableRow>",
    '      <TableCell className="text-base">A</TableCell>',
    '      <TableCell className="py-1">B</TableCell>',
    '      <TableCell className="py-2">C</TableCell>',
    '      <TableCell className="py-2.5">D</TableCell>',
    "      <TableCell",
    '        className="p-0"',
    "      >E</TableCell>",
    "    </TableRow>",
    "  </TableBody>",
    "</Table>",
  ].join("\n");

  const violations = audit.checkC18(
    text,
    path.join("src", "features", "widget", "components", "WidgetTable.tsx"),
  );
  assert.equal(violations.length, 12);
  assert.deepEqual(
    violations.map(({ lineNumber }) => lineNumber),
    [4, 5, 6, 7, 8, 9, 11, 17, 18, 19, 20, 22],
  );
});

test("checkC18: multiline JSX の className 式内にある override の実在行を返す", () => {
  assert.equal(typeof audit.checkC18, "function");

  const text = [
    "<TableCell",
    "  colSpan={2}",
    "  className={cn(",
    "    STYLE.tableCell,",
    '    "py-2",',
    "  )}",
    ">",
    "  value",
    "</TableCell>",
  ].join("\n");

  const violations = audit.checkC18(
    text,
    path.join("src", "features", "widget", "components", "WidgetTable.tsx"),
  );
  assert.equal(violations.length, 1);
  assert.equal(violations[0].lineNumber, 5);
});

test("checkC18: className より前の arrow function で opening tag の走査を打ち切らない", () => {
  const text = [
    "<TableCell",
    "  onClick={(event) => event.stopPropagation()}",
    '  className="py-2"',
    ">",
    "  value",
    "</TableCell>",
  ].join("\n");

  const violations = audit.checkC18(
    text,
    path.join("src", "features", "widget", "components", "WidgetTable.tsx"),
  );
  assert.equal(violations.length, 1);
  assert.equal(violations[0].lineNumber, 3);
});

test("checkC18: comment・string・prop内の子JSX classNameはprimitive overrideとして扱わない", () => {
  const text = [
    '// <TableCell className="py-1">comment</TableCell>',
    'const fixture = `<TableCell className="py-1">string</TableCell>`;',
    'const multilineFixture = `',
    '<TableCell className="py-1">multiline string</TableCell>',
    '`;',
    '<TableCell render={() => <span className="py-1">child</span>}>value</TableCell>',
  ].join("\n");

  const violations = audit.checkC18(
    text,
    path.join("src", "features", "widget", "components", "WidgetTable.tsx"),
  );
  assert.equal(violations.length, 0);
});

test("checkC18: header/body の標準 typography より大きい・小さい文字段を禁止する", () => {
  const text = [
    '<TableHead className="text-sm">header</TableHead>',
    '<TableCell className="text-xs">dense body</TableCell>',
    '<TableCell className="text-2xs">smaller body</TableCell>',
  ].join("\n");

  const violations = audit.checkC18(
    text,
    path.join("src", "features", "widget", "components", "WidgetTable.tsx"),
  );
  assert.deepEqual(violations.map(({ lineNumber }) => lineNumber), [1, 2, 3]);
});

test("checkC18: TableHead/TableCell に逆側の STYLE tokenを渡す迂回を禁止する", () => {
  const text = [
    "<TableHead className={STYLE.tableCell}>header</TableHead>",
    "<TableHead className={STYLE.tableCellMono}>header</TableHead>",
    "<TableHead className={STYLE.tableCellMuted}>header</TableHead>",
    "<TableCell className={STYLE.tableHeaderCell}>body</TableCell>",
  ].join("\n");

  const violations = audit.checkC18(
    text,
    path.join("src", "features", "widget", "components", "WidgetTable.tsx"),
  );
  assert.deepEqual(violations.map(({ lineNumber }) => lineNumber), [1, 2, 3, 4]);
});

test("checkC18: 標準py-3以外のvertical paddingを禁止し、明示した空stateだけを許可する", () => {
  const text = [
    '<TableCell className="py-0">body</TableCell>',
    '<TableCell className="py-4">body</TableCell>',
    '<TableCell className="py-[10px]">body</TableCell>',
    '<TableHead className="py-6">header</TableHead>',
    '<TableCell colSpan={5} className="py-8 text-center">集計値</TableCell>',
    "<TableCell",
    "  colSpan={5}",
    '  className="py-12 text-center"',
    ">集計値</TableCell>",
  ].join("\n");

  const violations = audit.checkC18(
    text,
    path.join("src", "features", "widget", "components", "WidgetTable.tsx"),
  );
  assert.deepEqual(violations.map(({ lineNumber }) => lineNumber), [1, 2, 3, 4, 5, 8]);
});

test("checkC18: padding shorthand・片側padding・誤ったtypography roleを禁止する", () => {
  const text = [
    '<TableCell className="p-4">body</TableCell>',
    '<TableCell className="pt-2">body</TableCell>',
    '<TableCell className="pb-[10px]">body</TableCell>',
    '<TableCell className="text-xl">body</TableCell>',
    '<TableCell className="font-bold">body</TableCell>',
    '<TableHead className="text-xl">header</TableHead>',
    '<TableHead className="font-bold">header</TableHead>',
    '<TableCell className="text-heading-3">body</TableCell>',
    '<TableHead className="text-heading-1">header</TableHead>',
  ].join("\n");

  const violations = audit.checkC18(
    text,
    path.join("src", "features", "widget", "components", "WidgetTable.tsx"),
  );
  assert.deepEqual(violations.map(({ lineNumber }) => lineNumber), [1, 2, 3, 4, 5, 6, 7, 8, 9]);
});

test("checkC18: 標準 table class・class なし・明示した空 state の追加余白は許可する", () => {
  assert.equal(typeof audit.checkC18, "function");

  const text = [
    "<TableHead>名前</TableHead>",
    '<TableHead className={STYLE.tableHeaderCell}>名前</TableHead>',
    '<TableHead className="text-2xs font-semibold px-4 py-3">名前</TableHead>',
    "<TableCell>ポチ</TableCell>",
    '<TableCell className={STYLE.tableCell}>ポチ</TableCell>',
    '<TableCell className="text-sm font-normal px-4 py-3">ポチ</TableCell>',
    '<TableCell data-empty-state colSpan={5} className="py-6 text-center">データなし</TableCell>',
    '<TableCell><span className="text-base py-2">子要素</span></TableCell>',
    '<TableCell aria-label="text-base py-2" data-density="py-0">属性値</TableCell>',
    '<TableHead aria-description="text-sm py-6">属性値</TableHead>',
    '<DataTableCell className="text-base py-2 p-0">別 component</DataTableCell>',
  ].join("\n");
  const mainComponent = path.join("src", "features", "widget", "components", "WidgetTable.tsx");

  assert.equal(audit.checkC18(text, mainComponent).length, 0);
  assert.equal(
    audit.checkC18(
      '<TableCell className="text-base py-2 p-0">TS file</TableCell>',
      path.join("src", "features", "widget", "components", "table-source.ts"),
    ).length,
    0,
  );
});

test("collectViolations: Table primitive 呼び出し側 override を C18 に集計する", async () => {
  const root = mkdtempSync(path.join(tmpdir(), "design-audit-fixture-"));
  try {
    const componentsDir = path.join(root, "src", "features", "widget", "components");
    mkdirSync(componentsDir, { recursive: true });
    writeFileSync(
      path.join(componentsDir, "BrokenTable.tsx"),
      [
        'import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";',
        "export function BrokenTable() {",
        "  return (",
        "    <Table>",
        "      <TableHeader>",
        "        <TableRow>",
        '          <TableHead className="text-xs">名前</TableHead>',
        "        </TableRow>",
        "      </TableHeader>",
        "      <TableBody>",
        "        <TableRow>",
        "          <TableCell",
        '            className="py-2.5"',
        "          >ポチ</TableCell>",
        "        </TableRow>",
        "      </TableBody>",
        "    </Table>",
        "  );",
        "}",
      ].join("\n"),
    );
    writeFileSync(
      path.join(componentsDir, "CleanTable.tsx"),
      [
        "export function CleanTable() {",
        '  return <TableCell className="text-sm font-normal px-4 py-3">ポチ</TableCell>;',
        "}",
      ].join("\n"),
    );
    writeFileSync(
      path.join(componentsDir, "BrokenTable.test.tsx"),
      'render(<TableCell className="text-base p-0">fixture test</TableCell>);',
    );

    const result = await collectViolations(root);
    assert.ok(Array.isArray(result.c18), "C18 table override の集計結果が必要");
    assert.equal(result.c18.length, 2);
    assert.deepEqual(
      result.c18.map(({ file, lineNumber }) => ({ file, lineNumber })),
      [
        {
          file: path.join("src", "features", "widget", "components", "BrokenTable.tsx"),
          lineNumber: 7,
        },
        {
          file: path.join("src", "features", "widget", "components", "BrokenTable.tsx"),
          lineNumber: 13,
        },
      ],
    );
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
});

test("checkC18: raw th/td の非仕様 typography・vertical padding・STYLE cross-use を検出する", () => {
  const text = [
    '<th className="text-sm">サイズ</th>',
    '<th className="font-medium">weight</th>',
    '<th className="py-2">縦padding</th>',
    "<th className={STYLE.tableCell}>cross</th>",
    '<td className="text-xs">サイズ</td>',
    '<td className="text-2xs">サイズ</td>',
    '<td className="font-bold">weight</td>',
    '<td className="py-2.5">縦padding</td>',
    '<td className="p-0">shorthand</td>',
    "<td className={STYLE.tableHeaderCell}>cross</td>",
    "<td",
    '  className="py-1"',
    ">multiline</td>",
  ].join("\n");

  const violations = audit.checkC18(
    text,
    path.join("src", "features", "widget", "components", "RawTable.tsx"),
  );
  assert.deepEqual(
    violations.map(({ lineNumber }) => lineNumber),
    [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11],
  );
});

test("checkC18: raw th/td 限定で横padding（px-4以外）と背景再指定を検出する", () => {
  const text = [
    "<th className={`${STYLE.tableHeaderCell} px-3`}>横padding上書き</th>",
    '<td className="px-3 py-3">横padding</td>',
    '<td className="pr-3">片側横padding</td>',
    '<td className="pl-10">片側横padding</td>',
    '<td className="px-[10px]">任意値横padding</td>',
    "<th className={`${C.bgLight} text-left`}>背景</th>",
    '<td className="bg-white">背景</td>',
    '<td className="px-4 py-3">仕様値は許可</td>',
    '<TableCell className="px-3">primitive の横paddingは対象外（基底styleがpx-4を保証）</TableCell>',
  ].join("\n");

  const violations = audit.checkC18(
    text,
    path.join("src", "features", "widget", "components", "RawTable.tsx"),
  );
  assert.deepEqual(
    violations.map(({ lineNumber }) => lineNumber),
    [1, 2, 3, 4, 5, 6, 7, 8],
  );
});

test("checkC18: 仕様準拠の raw th/td（STYLE token・仕様値literal・構造セル・空state）は許可する", () => {
  const text = [
    "<th className={`${STYLE.tableHeaderCell} text-left w-48`}>種別</th>",
    '<th className="text-2xs font-semibold px-4 py-3">名前</th>',
    "<td className={STYLE.tableCell}>ポチ</td>",
    "<td className={`${STYLE.tableCell} font-mono`}>0123</td>",
    "<td className={`${STYLE.tableCellMuted} text-right`}>備考</td>",
    "<td className={`text-sm font-normal px-4 py-3 ${C.text}`}>値</td>",
    "<td colSpan={6} className={STYLE.tableEmpty}>データなし</td>",
    '<td data-empty-state colSpan={5} className="py-12 text-center">データなし</td>',
    '<th data-c18-structural-cell className="w-11 px-0" />',
    '<th data-c18-structural-cell className="w-20" />',
    '<td className={STYLE.tableCell}><span className="text-base py-2 px-2">子要素</span></td>',
    '<thead className="text-base">',
    '<tr className="text-xs py-2">',
  ].join("\n");

  const violations = audit.checkC18(
    text,
    path.join("src", "features", "widget", "components", "RawTable.tsx"),
  );
  assert.deepEqual(violations, []);
});

test("checkC18: raw cellのmarker・variant・Tailwind v4 shorthand回避をfail closedする", () => {
  const text = [
    "<th>unstyled</th>",
    '<th className="w-11 px-0" />',
    "<td className={`${STYLE.tableCell} ps-3`}>logical padding</td>",
    '<th data-c18-structural-cell className="bg-red-500 px-3 text-base" />',
    '<th className="sm:text-2xs sm:font-semibold sm:px-4 sm:py-3">variant only</th>',
    '<td className="text-sm font-normal px-(--dense) py-3">v4 shorthand</td>',
    '<td className="text-sm font-normal px-4 py-(--dense)">v4 shorthand</td>',
  ].join("\n");

  const violations = audit.checkC18(
    text,
    path.join("src", "features", "widget", "components", "RawTable.tsx"),
  );
  assert.deepEqual(violations.map(({ lineNumber }) => lineNumber), [1, 2, 3, 4, 5, 6, 7]);
});

test("checkC18: raw cell allowlist（primitive正本/print帳票/ManualContent/owner-report密度/nested構造）と別アプリを除外する", () => {
  const rawViolation = '<td className="py-2 pr-3 text-xs">dense</td>';
  const exemptPaths = [
    path.join("src", "components", "ui", "table.tsx"),
    path.join("src", "features", "accounting", "components", "AccountingDocument.tsx"),
    path.join("src", "features", "accounting", "components", "DailyAccountingTabParts.tsx"),
    path.join("src", "features", "manual", "components", "ManualContent.tsx"),
    path.join("src", "features", "owner-report", "components", "HistoryTable.tsx"),
    path.join("src", "features", "owner-report", "components", "TreatmentHistorySection.tsx"),
    path.join("src", "features", "owner-report", "components", "TrimmingHistorySection.tsx"),
    path.join("src", "features", "owner-report", "components", "VaccinationHistorySection.tsx"),
    path.join("src", "features", "owner-report", "components", "ClinicalHistoryMatrix.tsx"),
    path.join("src", "features", "master", "components", "ReservationTypeGroupedTableRows.tsx"),
    path.join("src", "features", "accounting-reports", "components", "MonthlyReportPrintArea.tsx"),
    path.join("src", "features", "cash-register", "components", "ClosePrintArea.tsx"),
    path.join("src", "features", "medical-records", "components", "MedicalRecordPrintView.tsx"),
    path.join("liff", "src", "pages", "PetHealthPage.tsx"),
  ];
  for (const exemptPath of exemptPaths) {
    assert.equal(audit.checkC18(rawViolation, exemptPath).length, 0, exemptPath);
  }
  assert.equal(
    audit.checkC18(
      rawViolation,
      path.join("src", "features", "widget", "components", "Widget.tsx"),
    ).length,
    1,
  );
});

test("checkC18: raw cell allowlist ファイルでも Table primitive override は検出する", () => {
  const violations = audit.checkC18(
    '<TableCell className="py-2">x</TableCell>',
    path.join("src", "features", "manual", "components", "ManualContent.tsx"),
  );
  assert.equal(violations.length, 1);
});

test("collectViolations: raw th/td override を C18 に集計し allowlist は除外する", async () => {
  const root = mkdtempSync(path.join(tmpdir(), "design-audit-fixture-"));
  try {
    const componentsDir = path.join(root, "src", "features", "widget", "components");
    const manualDir = path.join(root, "src", "features", "manual", "components");
    mkdirSync(componentsDir, { recursive: true });
    mkdirSync(manualDir, { recursive: true });
    writeFileSync(
      path.join(componentsDir, "RawBrokenTable.tsx"),
      [
        "export function RawBrokenTable() {",
        "  return (",
        "    <table>",
        "      <thead>",
        "        <tr>",
        '          <th className="py-2">名前</th>',
        "        </tr>",
        "      </thead>",
        "      <tbody>",
        "        <tr>",
        '          <td className="px-3">ポチ</td>',
        "        </tr>",
        "      </tbody>",
        "    </table>",
        "  );",
        "}",
      ].join("\n"),
    );
    writeFileSync(
      path.join(manualDir, "ManualContent.tsx"),
      'export function ManualContent() { return <td className="py-2">doc</td>; }',
    );

    const result = await collectViolations(root);
    assert.deepEqual(
      result.c18.map(({ file, lineNumber }) => ({ file, lineNumber })),
      [
        {
          file: path.join("src", "features", "widget", "components", "RawBrokenTable.tsx"),
          lineNumber: 6,
        },
        {
          file: path.join("src", "features", "widget", "components", "RawBrokenTable.tsx"),
          lineNumber: 11,
        },
      ],
    );
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
});

test("checkC19: DataTableRow / SortableDataTableRow の single-line onClick を検出する", () => {
  assert.equal(typeof audit.checkC19, "function");

  const text = [
    "<DataTableRow onClick={onEdit}><TableCell>通常行</TableCell></DataTableRow>",
    "<SortableDataTableRow id={item.id} onClick={() => onEdit(item)}>並べ替え行</SortableDataTableRow>",
  ].join("\n");

  const violations = audit.checkC19(
    text,
    path.join("src", "features", "widget", "components", "WidgetRows.tsx"),
  );
  assert.equal(violations.length, 2);
  assert.deepEqual(violations.map(({ lineNumber }) => lineNumber), [1, 2]);
  assert.match(violations[0].text, /DataTableRow/);
  assert.match(violations[1].text, /SortableDataTableRow/);
});

test("checkC19: raw TableRow / tr の行全体 onClick も検出する", () => {
  assert.equal(typeof audit.checkC19, "function");

  const text = [
    "<TableRow onClick={onEdit}><TableCell>通常行</TableCell></TableRow>",
    "<tr",
    "  data-highlighted={isHighlighted}",
    "  onClick={() => onSelect(item)}",
    ">",
    "  <td><button onClick={onAction}>子操作</button></td>",
    "</tr>",
    "<TableRow><TableCell><button onClick={onAction}>安全な子操作</button></TableCell></TableRow>",
  ].join("\n");

  const violations = audit.checkC19(
    text,
    path.join("src", "features", "widget", "components", "RawWidgetRows.tsx"),
  );
  assert.equal(violations.length, 2);
  assert.deepEqual(violations.map(({ lineNumber }) => lineNumber), [1, 4]);
  assert.match(violations[0].text, /TableRow/);
  assert.match(violations[1].text, /onClick/);
});

test("checkC19: multiline opening tag と arrow handler の onClick 実在行を返す", () => {
  assert.equal(typeof audit.checkC19, "function");

  const text = [
    "<DataTableRow",
    "  key={item.id}",
    "  onClick={() => onEdit(item)}",
    ">",
    "  <TableCell>通常行</TableCell>",
    "</DataTableRow>",
    "<SortableDataTableRow",
    "  id={item.id}",
    "  onClick={(event) => {",
    "    event.preventDefault();",
    "    onEdit(item);",
    "  }}",
    ">",
    "  並べ替え行",
    "</SortableDataTableRow>",
  ].join("\n");

  const violations = audit.checkC19(
    text,
    path.join("src", "features", "widget", "components", "WidgetRows.tsx"),
  );
  assert.equal(violations.length, 2);
  assert.deepEqual(violations.map(({ lineNumber }) => lineNumber), [3, 9]);
});

test("checkC19: test file・comment・string 内の見かけ上の row onClick は除外する", () => {
  assert.equal(typeof audit.checkC19, "function");

  const source = [
    "// <DataTableRow onClick={onEdit}>comment</DataTableRow>",
    "/* <SortableDataTableRow onClick={() => onEdit(item)}>comment</SortableDataTableRow> */",
    "const singleQuoted = '<DataTableRow onClick={onEdit}>fixture</DataTableRow>';",
    "const doubleQuoted = \"<SortableDataTableRow onClick={onEdit}>fixture</SortableDataTableRow>\";",
    "const template = `<DataTableRow onClick={() => onEdit(item)}>fixture</DataTableRow>`;",
    "<DataTableRow data-description=\"onClick={onEdit}\"><TableCell>安全な行</TableCell></DataTableRow>",
  ].join("\n");
  const componentPath = path.join("src", "features", "widget", "components", "WidgetRows.tsx");
  const testPath = path.join("src", "features", "widget", "components", "WidgetRows.test.tsx");

  assert.equal(audit.checkC19(source, componentPath).length, 0);
  assert.equal(
    audit.checkC19("<DataTableRow onClick={onEdit}>test</DataTableRow>", testPath).length,
    0,
  );
});

test("collectViolations: row-level onClick を C19 に集計し test file は除外する", async () => {
  const root = mkdtempSync(path.join(tmpdir(), "design-audit-fixture-"));
  try {
    const componentsDir = path.join(root, "src", "features", "widget", "components");
    mkdirSync(componentsDir, { recursive: true });
    writeFileSync(
      path.join(componentsDir, "BrokenRows.tsx"),
      [
        "export function BrokenRows() {",
        "  return <>",
        "    <DataTableRow onClick={onEdit}>通常行</DataTableRow>",
        "    <SortableDataTableRow",
        "      id={item.id}",
        "      onClick={() => onEdit(item)}",
        "    >並べ替え行</SortableDataTableRow>",
        "  </>;",
        "}",
      ].join("\n"),
    );
    writeFileSync(
      path.join(componentsDir, "BrokenRows.test.tsx"),
      "render(<DataTableRow onClick={onEdit}>fixture test</DataTableRow>);",
    );

    const result = await collectViolations(root);
    assert.ok(Array.isArray(result.c19), "C19 row-level onClick の集計結果が必要");
    assert.deepEqual(
      result.c19.map(({ file, lineNumber }) => ({ file, lineNumber })),
      [
        {
          file: path.join("src", "features", "widget", "components", "BrokenRows.tsx"),
          lineNumber: 3,
        },
        {
          file: path.join("src", "features", "widget", "components", "BrokenRows.tsx"),
          lineNumber: 6,
        },
      ],
    );
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
});

test("CLI: tracking 直書きは C14 として exit code 1 になる", () => {
  const { root, routesDir } = makeFixture();
  try {
    writeFileSync(
      path.join(routesDir, "WidgetPage.tsx"),
      'export function WidgetPage() { return <PageLayout title="ok" className="tracking-wide">x</PageLayout>; }',
    );
    const scriptPath = path.join(import.meta.dirname, "design-system-audit.mjs");
    assert.throws(() => {
      execFileSync("node", [scriptPath, "--cwd", root], { encoding: "utf-8" });
    }, (err) => {
      assert.equal(err.status, 1);
      assert.match(err.stdout, /C14 tracking 直書き/);
      assert.match(err.stdout, /WidgetPage\.tsx:1/);
      assert.match(err.stdout, /FAIL — 1 件/);
      return true;
    });
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
});

test("checkC8: allowlist は basename ではなく feature 相対パスだけを許可する", () => {
  const loginPath = path.join("src", "features", "auth", "routes", "Login.tsx");
  const spoofedLoginPath = path.join("src", "features", "widget", "routes", "Login.tsx");
  const fooPath = path.join("src", "features", "widget", "routes", "FooPage.tsx");

  assert.equal(C8_PAGE_ALLOWLIST.size, 9);
  assert.equal(C8_ROUTE_HELPER_ALLOWLIST.size, 35);
  assert.equal(C8_ALLOWLIST.size, 44);
  assert.ok([...C8_ALLOWLIST].every((relPath) => relPath.includes(path.sep)));
  assert.equal(checkC8(loginPath, "export function Login() { return null; }").length, 0);
  assert.equal(checkC8(spoofedLoginPath, "export function Login() { return null; }").length, 1);
  assert.equal(checkC8(fooPath, "return <PageLayout title=\"x\">").length, 0);
  assert.equal(checkC8(fooPath, "return <MasterListPage />").length, 0);
  assert.equal(checkC8(fooPath, "export function Foo() { return null; }").length, 1);
});

test("collectViolations（FE8-5）: C8 allowlist と同名の別 feature route は除外しない", async () => {
  const root = mkdtempSync(path.join(tmpdir(), "design-audit-fixture-"));
  try {
    const authRoutesDir = path.join(root, "src", "features", "auth", "routes");
    const widgetRoutesDir = path.join(root, "src", "features", "widget", "routes");
    mkdirSync(authRoutesDir, { recursive: true });
    mkdirSync(widgetRoutesDir, { recursive: true });
    writeFileSync(path.join(authRoutesDir, "Login.tsx"), "export function Login() { return null; }");
    writeFileSync(path.join(widgetRoutesDir, "Login.tsx"), "export function Login() { return null; }");
    const result = await collectViolations(root);
    assert.equal(result.c8.length, 1);
    assert.equal(result.c8[0].file, path.join("src", "features", "widget", "routes", "Login.tsx"));
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
});
