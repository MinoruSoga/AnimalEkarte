#!/usr/bin/env node

/**
 * features 配下ファイル名規則 ratchet — `src/features/**` の非コンポーネント `.ts` が
 * kebab-case.ts 命名規則（FE-refactor.md R-F3 / `src/features/CLAUDE.md` 正本）から
 * ドリフトしていないかを判定する。
 *
 * R-F3-S1 で PascalCase `*Model.ts` 53件 + 同種4件を kebab-case にリネーム済みだが、
 * camelCase hook 15件など残存債務が R-F3-S2 まで残るため、baseline は 0 ではなく
 * 「現状件数」で arm する（check-eslint-disable-rationale.mjs と同じ ratchet 方式）。
 *
 * 実行:
 * $ node scripts/check-feature-filename-convention.mjs --baseline .filename-baseline
 *
 * 「違反」の定義: `src/features/**` 配下の `.ts` で、basename（拡張子除く）に
 * `[A-Z]` を含むもの。
 *
 * FE-RC-017 で `.tsx` コンポーネントの逆方向チェックを追加: JSX を export しているのに
 * basename が小文字始まりのファイル（features/CLAUDE.md: JSX export は PascalCase.tsx が正本）。
 * `.filename-baseline` の2行目（1つ目の数値行の次）が `.tsx` ratchet の baseline。
 * JSX export の判定は簡易ヒューリスティック（`<Tag` 形の出現）で、AST 解析ではない。
 *
 * 既知の限界: baseline はリポジトリ全体の合計件数（ファイル単位ではない）。
 * ある1ファイルで違反を1件減らしつつ別ファイルで1件増やすと合計が変わらず検出できない。
 * coverage-ratchet.mjs / check-eslint-disable-rationale.mjs と同じトレードオフ。
 */

import { readFileSync } from "node:fs";
import { readdir } from "node:fs/promises";
import path from "node:path";

const SCAN_ROOT = "src/features";
const SCAN_EXTENSION = ".ts";
const EXCLUDE_DIR_NAMES = new Set(["node_modules", "dist", "generated"]);
const UPPERCASE_RE = /[A-Z]/;

// --- FE-RC-017: .tsx コンポーネントの PascalCase ratchet -----------------------------------
const TSX_SCAN_EXTENSION = ".tsx";
const LOWERCASE_START_RE = /^[a-z]/;
/** JSX を export しているかの簡易検出（tag 形の `<Foo` / `<div` を含むか）。
 *  型定義のみのファイルや純粋 hook（JSX を返さない）は対象外にするための最低限の判定で、
 *  精密な AST 解析ではない（他の ratchet スクリプトと同じ簡易ヒューリスティック方針）。 */
const JSX_TAG_RE = /<[A-Za-z][\w.-]*[\s/>]/;

function parseArgs(argv) {
  const args = { baseline: ".filename-baseline", cwd: process.cwd() };
  for (let i = 0; i < argv.length; i += 1) {
    const arg = argv[i];
    if (arg === "--baseline") args.baseline = argv[++i];
    else if (arg === "--cwd") args.cwd = argv[++i];
  }
  return args;
}

/**
 * isViolation は basename（拡張子除く）が kebab-case 規則に違反しているかを判定する。
 */
export function isViolation(basename) {
  return UPPERCASE_RE.test(basename);
}

/**
 * isTsxViolation は `.tsx` コンポーネントファイルの basename が小文字始まりか判定する
 * （features/CLAUDE.md: JSX を export するファイルは PascalCase.tsx が正本、FE-RC-017）。
 * `fileText` を渡した場合のみ JSX export の簡易判定を行い、渡さない場合は basename のみで判定する。
 */
export function isTsxViolation(basename, fileText) {
  // FE-RC-017 followup2: *.test.tsx / *.spec.tsx はコンポーネントではない。
  // use-*.tsx は JSX を返す hook（features/CLAUDE.md の正当パターン）なので除外。
  if (basename.endsWith(".test") || basename.endsWith(".spec")) return false;
  if (basename.startsWith("use-")) return false;
  if (!LOWERCASE_START_RE.test(basename)) return false;
  if (fileText === undefined) return true;
  return JSX_TAG_RE.test(fileText);
}

async function walk(dir, excludeNames, extensions) {
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
      files.push(...(await walk(full, excludeNames, extensions)));
    } else if (extensions.has(path.extname(entry.name))) {
      files.push(full);
    }
  }
  return files;
}

/**
 * readBaselines は baseline ファイルからコメント/空行を除いた数値行を出現順に全て読む。
 */
export function readBaselines(fileText) {
  const values = [];
  for (const rawLine of fileText.split("\n")) {
    const line = rawLine.trim();
    if (line === "" || line.startsWith("#")) continue;
    const value = Number(line);
    if (!Number.isNaN(value)) values.push(value);
  }
  return values;
}

/**
 * readBaseline は baseline ファイルから記録済み件数を読む（`index` 番目の数値行、0-based）。
 * 欠落・空・解析不能なら 0 を返す。既定の `index=0` は `.ts` kebab-case ratchet（後方互換）。
 * `index=1` は `.tsx` PascalCase ratchet（FE-RC-017）用。
 */
export function readBaseline(fileText, index = 0) {
  return readBaselines(fileText)[index] ?? 0;
}

const DEFAULT_RATCHET_HINT =
  "New non-component .ts files under src/features must use kebab-case filenames.";

/**
 * evaluateRatchet は現在件数とベースラインを比較する純粋関数。
 *  - current > baseline: fail=true（新規違反が混入した）。
 *  - current <= baseline: OK。
 */
export function evaluateRatchet(current, baseline, hint = DEFAULT_RATCHET_HINT) {
  if (current > baseline) {
    return {
      fail: true,
      message: `check-filenames: FAIL — violation count ${current} exceeds baseline ${baseline}. ${hint}`,
    };
  }
  return {
    fail: false,
    message: `check-filenames: OK — violation count ${current} <= baseline ${baseline}.`,
  };
}

async function main() {
  const args = parseArgs(process.argv.slice(2));

  const root = path.join(args.cwd, SCAN_ROOT);
  const allFiles = await walk(
    root,
    EXCLUDE_DIR_NAMES,
    new Set([SCAN_EXTENSION, TSX_SCAN_EXTENSION]),
  );

  const violations = [];
  const tsxViolations = [];
  for (const file of allFiles) {
    if (path.extname(file) === TSX_SCAN_EXTENSION) {
      const basename = path.basename(file, TSX_SCAN_EXTENSION);
      const fileText = readFileSync(file, "utf-8");
      if (isTsxViolation(basename, fileText)) {
        tsxViolations.push(path.relative(args.cwd, file));
      }
      continue;
    }
    const basename = path.basename(file, SCAN_EXTENSION);
    if (isViolation(basename)) {
      violations.push(path.relative(args.cwd, file));
    }
  }

  let baselineText = "";
  try {
    baselineText = readFileSync(path.join(args.cwd, args.baseline), "utf-8");
  } catch {
    baselineText = "";
  }
  const baseline = readBaseline(baselineText, 0);
  const tsxBaseline = readBaseline(baselineText, 1);

  const result = evaluateRatchet(violations.length, baseline);
  console.log(result.message);
  if (violations.length > 0) {
    console.log("kebab-case violations (.ts):");
    for (const file of violations.sort()) {
      console.log(`  ${file}`);
    }
  }

  const tsxResult = evaluateRatchet(
    tsxViolations.length,
    tsxBaseline,
    "New JSX-exporting .tsx files under src/features must use PascalCase filenames (FE-RC-017).",
  );
  console.log(tsxResult.message);
  if (tsxViolations.length > 0) {
    console.log("PascalCase violations (.tsx):");
    for (const file of tsxViolations.sort()) {
      console.log(`  ${file}`);
    }
  }

  if (result.fail || tsxResult.fail) {
    process.exit(1);
  }
}

const isMain = process.argv[1] && import.meta.url === `file://${process.argv[1]}`;
if (isMain) {
  main();
}
