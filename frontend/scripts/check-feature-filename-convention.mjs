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
 * 「違反」の定義: `src/features/**` 配下の `.ts`（`.tsx` は対象外 — コンポーネントは
 * PascalCase.tsx が正本）で、basename（拡張子除く）に `[A-Z]` を含むもの。
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

async function walk(dir, excludeNames) {
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
      files.push(...(await walk(full, excludeNames)));
    } else if (path.extname(entry.name) === SCAN_EXTENSION) {
      files.push(full);
    }
  }
  return files;
}

/**
 * readBaseline は baseline ファイルから記録済み件数を読む。欠落・空・解析不能なら 0 を返す。
 */
export function readBaseline(fileText) {
  for (const rawLine of fileText.split("\n")) {
    const line = rawLine.trim();
    if (line === "" || line.startsWith("#")) continue;
    const value = Number(line);
    if (!Number.isNaN(value)) return value;
  }
  return 0;
}

/**
 * evaluateRatchet は現在件数とベースラインを比較する純粋関数。
 *  - current > baseline: fail=true（新規 PascalCase `.ts` が混入した）。
 *  - current <= baseline: OK。
 */
export function evaluateRatchet(current, baseline) {
  if (current > baseline) {
    return {
      fail: true,
      message: `check-filenames: FAIL — violation count ${current} exceeds baseline ${baseline}. New non-component .ts files under src/features must use kebab-case filenames.`,
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
  const allFiles = await walk(root, EXCLUDE_DIR_NAMES);

  const violations = [];
  for (const file of allFiles) {
    const basename = path.basename(file, SCAN_EXTENSION);
    if (isViolation(basename)) {
      violations.push(path.relative(args.cwd, file));
    }
  }

  let baseline = 0;
  try {
    baseline = readBaseline(readFileSync(path.join(args.cwd, args.baseline), "utf-8"));
  } catch {
    baseline = 0;
  }

  const result = evaluateRatchet(violations.length, baseline);
  console.log(result.message);
  if (violations.length > 0) {
    console.log("kebab-case violations:");
    for (const file of violations.sort()) {
      console.log(`  ${file}`);
    }
  }
  if (result.fail) {
    process.exit(1);
  }
}

const isMain = process.argv[1] && import.meta.url === `file://${process.argv[1]}`;
if (isMain) {
  main();
}
