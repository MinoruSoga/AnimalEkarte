/**
 * design-system-audit — docs/spec/ui-design-compliance.md §1 の機械判定。
 *
 * 判定内容:
 *   C1 — legacy accent 色（`C.accent` / `#0075DE` / `#2383E2`）禁止。
 *   C3 — hex 直書き（文字列/テンプレートリテラル内）禁止。
 *   C5 — Primary CTA の `colorVariant` は `"brand"` のみ。
 *   C6 — rgba()/rgb()/hsla()/hsl() の直値禁止（doc §1 の臨床安全 C6a とは ID 分離: 本スクリプトは C6b）。
 *   C7 — PageLayout maxWidth の生値禁止（`max-w-full` / `max-w-[Npx]`）。トークン経由のみ。
 *   C8 — `src/features/<feat>/routes/<file>.tsx` が PageLayout / Master*Page / allowlist のいずれか。
 *   C9 — `rounded(-[trbl]{1,2})?-[Npx]` 任意値禁止（トークン `rounded-xxs/xs/...` へ）。
 *
 * C2/C4（PageLayout 使用）は C8 で routes 配下を機械化。新規リーフを allowlist に載せる場合は
 * C8_ALLOWLIST と docs/spec/ui-design-compliance.md §2 を同一コミットで更新する。
 *
 * strict fail（1件でも検出したら exit 1）。正当例外は allowlist 追記のみ。
 *
 * Exit: 0=OK / 1=違反 / 2=実行エラー
 * 実行: node scripts/design-system-audit.mjs [--cwd frontend]
 *       docker compose exec frontend pnpm design-audit
 */

import { readFileSync } from "node:fs";
import { readdir } from "node:fs/promises";
import path from "node:path";

const SCAN_ROOTS = ["src", path.join("liff", "src"), path.join("line-reserve", "src")];
const SCAN_EXTENSIONS = new Set([".ts", ".tsx"]);
const EXCLUDE_DIR_NAMES = new Set(["node_modules", "dist", "generated"]);
const LEAF_DIR_NAMES = new Set(["routes", "pages"]);

// design-tokens.ts は色定数の定義正本のため C3/C6 の allowlist。
const DESIGN_TOKENS_REL_PATH = path.join("src", "lib", "design-tokens.ts");
// 実行時に動的生成する rgba() のみ扱う（JSDoc で根拠を記載済み）ため C6 の allowlist。
const COLOR_MAP_REL_PATH = path.join("src", "hooks", "use-reservation-type-color-map.ts");

const C1_RE = /C\.accent\b|#0075DE|#2383E2/;
const C3_RE = /['"`]#[0-9A-Fa-f]{3,8}['"`]/;
const C5_RE = /colorVariant="([a-zA-Z]+)"/g;
const C5_BRAND_VALUE = "brand";
const C6_RE = /rgba?\(|hsla?\(/;
const C7_RE = /maxWidth=["']max-w-(full|\[[0-9]+px\])["']/;
const C9_RE = /rounded(?:-[trbl]{1,2})?-\[[0-9]+px\]/

/** C8: PageLayout/Master shell を持たない正当な routes ファイル（basename 15件） */
export const C8_ALLOWLIST = new Set([
  "Login.tsx",
  "ForgotPasswordPage.tsx",
  "ResetPasswordPage.tsx",
  "OwnerReport.tsx",
  "Reception.tsx",
  "ManualPage.tsx",
  "ReservationManagement.tsx",
  "CheckupSyncPage.tsx",
  "ClinicMasterSettings.tsx",
  "LineReservationSlotsSettings.tsx",
  "TrimmingLazyModals.tsx",
  "ReceptionLazyModals.tsx",
  "LstepDeliveryMonitorPageParts.tsx",
  "LstepDeliveryMonitorLogsTable.tsx",
  "medical-records-columns.tsx",
]);
const C8_SHELL_RE = /<PageLayout|Master(?:CRUD|Tab|List)Page/;

function parseArgs(argv) {
  const args = { cwd: process.cwd() };
  for (let i = 0; i < argv.length; i += 1) {
    const arg = argv[i];
    if (arg === "--cwd") args.cwd = argv[++i];
  }
  return args;
}

/**
 * isTestFile はファイル名が `*.test.*` パターンか判定する（C3/C6 除外用）。
 */
export function isTestFile(fileName) {
  return /\.test\./.test(fileName);
}

/**
 * isLeafRouteFile は相対パスが routes/ または pages/ ディレクトリ配下かを判定する
 * （`--glob '**\/routes/**' --glob '**\/pages/**'` 相当）。
 * FE3-2 以降 collectViolations 自体はこの判定で走査対象を絞らないが、
 * ユーティリティとして引き続き提供する。
 */
export function isLeafRouteFile(relPath) {
  const segments = relPath.split(path.sep);
  return segments.some((seg) => LEAF_DIR_NAMES.has(seg));
}

/**
 * isDesignTokensFile は色定数カタログ本体（C3/C6 allowlist）かを判定する。
 */
export function isDesignTokensFile(relPath) {
  return relPath === DESIGN_TOKENS_REL_PATH;
}

/**
 * isColorMapFile は実行時動的生成 rgba() の唯一の許容ファイル（C6 allowlist）かを判定する。
 */
export function isColorMapFile(relPath) {
  return relPath === COLOR_MAP_REL_PATH;
}

/**
 * checkC1 はテキストから legacy accent 参照を行単位で検出する。
 */
export function checkC1(text) {
  const violations = [];
  text.split("\n").forEach((line, i) => {
    if (C1_RE.test(line)) {
      violations.push({ lineNumber: i + 1, text: line.trim() });
    }
  });
  return violations;
}

/**
 * checkC3 はテキストから引用符で囲まれた hex 直書きリテラルを行単位で検出する。
 * 引用符無しの `#158` のような issue 番号コメントは対象外（誤検知しない）。
 */
export function checkC3(text) {
  const violations = [];
  text.split("\n").forEach((line, i) => {
    if (C3_RE.test(line)) {
      violations.push({ lineNumber: i + 1, text: line.trim() });
    }
  });
  return violations;
}

/**
 * checkC5 はテキストから `colorVariant="..."` の値が "brand" 以外の箇所を検出する。
 */
export function checkC5(text) {
  const violations = [];
  text.split("\n").forEach((line, i) => {
    const re = new RegExp(C5_RE.source, "g");
    let match;
    while ((match = re.exec(line)) !== null) {
      if (match[1] !== C5_BRAND_VALUE) {
        violations.push({ lineNumber: i + 1, text: line.trim(), value: match[1] });
      }
    }
  });
  return violations;
}

/**
 * checkC6 はテキストから rgba()/rgb()/hsla()/hsl() の直値を行単位で検出する（FE3-2 新設）。
 */
export function checkC6(text) {
  const violations = [];
  text.split("\n").forEach((line, i) => {
    if (C6_RE.test(line)) {
      violations.push({ lineNumber: i + 1, text: line.trim() });
    }
  });
  return violations;
}


/**
 * checkC7 は maxWidth="max-w-full" / maxWidth="max-w-[Npx]" の生値を検出する（FE8-5）。
 */
export function checkC7(text) {
  const violations = [];
  text.split("\n").forEach((line, i) => {
    if (C7_RE.test(line)) {
      violations.push({ lineNumber: i + 1, text: line.trim() });
    }
  });
  return violations;
}

/**
 * checkC8 は routes ファイル本文が PageLayout / Master*Page / allowlist のいずれかを満たすか判定する。
 * 違反時は 1 件の擬似 violation を返す（ファイル単位）。
 */
export function checkC8(fileName, text) {
  if (C8_ALLOWLIST.has(fileName)) return [];
  if (C8_SHELL_RE.test(text)) return [];
  return [{ lineNumber: 1, text: "PageLayout / Master*Page / C8 allowlist のいずれも無し" }];
}

/**
 * checkC9 は rounded-[Npx] / rounded-t-[Npx] 等の任意値角丸を検出する（FE8-5）。
 */
export function checkC9(text) {
  const violations = [];
  text.split("\n").forEach((line, i) => {
    if (C9_RE.test(line)) {
      violations.push({ lineNumber: i + 1, text: line.trim() });
    }
  });
  return violations;
}

async function walk(dir, exts, excludeNames) {
  let entries;
  try {
    entries = await readdir(dir, { withFileTypes: true });
  } catch {
    return [];
  }
  const files = [];
  for (const entry of entries) {
    if (excludeNames.has(entry.name)) continue;
    const full = path.join(dir, entry.name);
    if (entry.isDirectory()) {
      files.push(...(await walk(full, exts, excludeNames)));
    } else if (exts.has(path.extname(entry.name))) {
      files.push(full);
    }
  }
  return files;
}

/**
 * collectViolations は cwd 配下の SCAN_ROOTS（src・liff/src・line-reserve/src）を走査し、
 * C1/C3/C5/C6 違反を集計する純粋寄りの関数。
 * ファイル I/O のみ副作用を持ち、判定ロジック自体は checkC1/checkC3/checkC5/checkC6 に委譲する。
 */
export async function collectViolations(cwd) {
  const result = { c1: [], c3: [], c5: [], c6: [], c7: [], c8: [], c9: [] };

  for (const scanRoot of SCAN_ROOTS) {
    const root = path.join(cwd, scanRoot);
    const allFiles = await walk(root, SCAN_EXTENSIONS, EXCLUDE_DIR_NAMES);

    for (const file of allFiles) {
      const relPath = path.relative(cwd, file);
      const isTest = isTestFile(path.basename(file));
      const text = readFileSync(file, "utf-8");

      if (!isDesignTokensFile(relPath)) {
        for (const v of checkC1(text)) {
          result.c1.push({ file: relPath, ...v });
        }
      }
      if (!isTest && !isDesignTokensFile(relPath)) {
        for (const v of checkC3(text)) {
          result.c3.push({ file: relPath, ...v });
        }
      }
      if (!isTest) {
        for (const v of checkC5(text)) {
          result.c5.push({ file: relPath, ...v });
        }
      }
      if (!isTest && !isDesignTokensFile(relPath) && !isColorMapFile(relPath)) {
        for (const v of checkC6(text)) {
          result.c6.push({ file: relPath, ...v });
        }
      }
      if (!isTest) {
        for (const v of checkC7(text)) {
          result.c7.push({ file: relPath, ...v });
        }
        for (const v of checkC9(text)) {
          result.c9.push({ file: relPath, ...v });
        }
      }

      // C8: src/features/<feat>/routes/<file>.tsx のみ（ネスト無し・.test 除外）
      const parts = relPath.split(path.sep);
      if (
        parts[0] === "src" &&
        parts[1] === "features" &&
        parts[3] === "routes" &&
        parts.length === 5 &&
        path.extname(parts[4]) === ".tsx" &&
        !isTest
      ) {
        for (const v of checkC8(parts[4], text)) {
          result.c8.push({ file: relPath, ...v });
        }
      }
    }
  }
  return result;
}

function printGroup(label, violations) {
  console.log(`design-system-audit: ${label} — ${violations.length} 件`);
  for (const v of violations) {
    console.log(`  ${v.file}:${v.lineNumber}: ${v.text}`);
  }
}

async function main() {
  const args = parseArgs(process.argv.slice(2));

  let result;
  try {
    result = await collectViolations(args.cwd);
  } catch (err) {
    console.error(`design-system-audit: 実行エラー — ${err instanceof Error ? err.message : String(err)}`);
    process.exit(2);
    return;
  }

  printGroup("C1 legacy accent", result.c1);
  printGroup("C3 hex 直書き", result.c3);
  printGroup("C5 非 brand colorVariant", result.c5);
  printGroup("C6 rgba/hsla 直値", result.c6);
  printGroup("C7 maxWidth 生値", result.c7);
  printGroup("C8 PageLayout 未使用 routes", result.c8);
  printGroup("C9 rounded 任意値", result.c9);

  const total = result.c1.length + result.c3.length + result.c5.length + result.c6.length
    + result.c7.length + result.c8.length + result.c9.length;
  if (total > 0) {
    console.log(`design-system-audit: FAIL — ${total} 件の違反`);
    process.exit(1);
  }
  console.log("design-system-audit: PASS — 違反 0 件");
}

const isMain = process.argv[1] && import.meta.url === `file://${process.argv[1]}`;
if (isMain) {
  main();
}
