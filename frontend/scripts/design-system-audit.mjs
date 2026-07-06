#!/usr/bin/env node

/**
 * design-system-audit — docs/UI_DESIGN_COMPLIANCE.md §1 の C1/C3/C5 判定を機械化する。
 *
 * 2026-07-06 UI 準拠監査で C2/C4（PageLayout 未使用）5件を grep ベースで手動発見したが、
 * 再発防止のため機械化できる判定（C1/C3/C5）だけを CI 前に実行できる形で固定する。
 * C2/C4（leaf component の PageLayout 使用追跡）は動的解決が複雑なため対象外 —
 * 引き続き docs/UI_DESIGN_COMPLIANCE.md §2 の手動更新で追跡する。
 *
 * 判定内容（docs/UI_DESIGN_COMPLIANCE.md §1 の grep コマンドと同一パターン）:
 *   C1 — legacy accent 色（`C.accent` / `#0075DE` / `#2383E2`）は brand teal 系のみ許容のため禁止
 *   C3 — route 表面での hex 直書き（文字列/テンプレートリテラル内の `#RGB`〜`#RRGGBBAA`）禁止。
 *        issue 番号コメント（`#158` 等）は引用符で囲まれないため誤検知しない。
 *   C5 — Primary CTA の `colorVariant` は `"brand"` のみ許容（`"default"` 等は違反）
 *
 * 対象範囲: `src/features/**\/routes/**` および `src/features/**\/pages/**`（.ts/.tsx）。
 * C3 のみ `*.test.*` を除外する（docs/UI_DESIGN_COMPLIANCE.md §1 の rg コマンドに準拠）。
 *
 * 現状（2026-07-06 監査時点）は C1/C3/C5 いずれも違反 0 件のため baseline ratchet ではなく
 * strict fail（1件でも検出したら exit 1）を採用する。新規に non-brand / legacy accent /
 * hex 直書きが混入した時点で失敗させ、baseline 更新という「既知債務化」の抜け道を作らない。
 *
 * Exit code:
 *   0 — 違反なし
 *   1 — 違反あり（file:line を stdout に列挙）
 *   2 — 実行エラー（想定外の例外）
 *
 * 実行:
 * $ node scripts/design-system-audit.mjs
 * $ docker compose exec frontend pnpm design-audit
 */

import { readFileSync } from "node:fs";
import { readdir } from "node:fs/promises";
import path from "node:path";

const SCAN_ROOT = "src/features";
const SCAN_EXTENSIONS = new Set([".ts", ".tsx"]);
const EXCLUDE_DIR_NAMES = new Set(["node_modules", "dist", "generated"]);
const LEAF_DIR_NAMES = new Set(["routes", "pages"]);

const C1_RE = /C\.accent|#0075DE|#2383E2/;
const C3_RE = /['"`]#[0-9A-Fa-f]{3,8}['"`]/;
const C5_RE = /colorVariant="([a-zA-Z]+)"/g;
const C5_BRAND_VALUE = "brand";

function parseArgs(argv) {
  const args = { cwd: process.cwd() };
  for (let i = 0; i < argv.length; i += 1) {
    const arg = argv[i];
    if (arg === "--cwd") args.cwd = argv[++i];
  }
  return args;
}

/**
 * isTestFile はファイル名が `*.test.*` パターンか判定する（C3 除外用）。
 */
export function isTestFile(fileName) {
  return /\.test\./.test(fileName);
}

/**
 * isLeafRouteFile は相対パスが routes/ または pages/ ディレクトリ配下かを判定する
 * （`--glob '**\/routes/**' --glob '**\/pages/**'` 相当）。
 */
export function isLeafRouteFile(relPath) {
  const segments = relPath.split(path.sep);
  return segments.some((seg) => LEAF_DIR_NAMES.has(seg));
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
 * collectViolations は cwd 配下の SCAN_ROOT を走査し、C1/C3/C5 違反を集計する純粋寄りの関数。
 * ファイル I/O のみ副作用を持ち、判定ロジック自体は checkC1/checkC3/checkC5 に委譲する。
 */
export async function collectViolations(cwd) {
  const root = path.join(cwd, SCAN_ROOT);
  const allFiles = await walk(root, SCAN_EXTENSIONS, EXCLUDE_DIR_NAMES);

  const result = { c1: [], c3: [], c5: [] };
  for (const file of allFiles) {
    const relPath = path.relative(cwd, file);
    if (!isLeafRouteFile(path.relative(root, file))) continue;

    const text = readFileSync(file, "utf-8");

    for (const v of checkC1(text)) {
      result.c1.push({ file: relPath, ...v });
    }
    if (!isTestFile(path.basename(file))) {
      for (const v of checkC3(text)) {
        result.c3.push({ file: relPath, ...v });
      }
    }
    for (const v of checkC5(text)) {
      result.c5.push({ file: relPath, ...v });
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
  printGroup("C3 route表面 hex 直書き", result.c3);
  printGroup("C5 非 brand colorVariant", result.c5);

  const total = result.c1.length + result.c3.length + result.c5.length;
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
