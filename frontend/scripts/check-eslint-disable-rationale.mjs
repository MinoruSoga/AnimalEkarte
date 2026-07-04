#!/usr/bin/env node

/**
 * eslint-disable 根拠コメント ratchet — 新規に追加される eslint-disable が
 * 根拠コメント無しで増えていないかを判定する
 * （FE-refactor.md「eslint-disable 根拠コメントの機械強制」・R-F3 の追跡）。
 *
 * R-F3 で既存の eslint-disable 群を監査・分類したが、多くは根拠コメント（`-- 理由`）を
 * 付けていない（レビュー時に個別判断された既知の債務として許容）。この既存債務を
 * 遡って fail させることはせず、frontend/scripts/coverage-ratchet.mjs と同じ
 * ratchet 方式で「件数が baseline を超えて増えていないか」だけを検査する
 * （eslint-comments/require-description 相当をプラグイン依存なしで機械強制）。
 *
 * 実行:
 * $ node scripts/check-eslint-disable-rationale.mjs --baseline .eslint-disable-baseline
 *
 * 「根拠コメント」の定義: eslint-disable / eslint-disable-next-line / eslint-disable-line の
 * コメント行に ` -- ` （半角スペース+ハイフン2つ+半角スペース）以降のテキストが存在すること。
 * 例:
 *   // eslint-disable-next-line react-hooks/exhaustive-deps -- 理由
 *   /* eslint-disable react-hooks/set-state-in-effect -- 理由 * /
 *
 * 既知の限界: baseline はリポジトリ全体の合計件数（ファイル単位ではない）。
 * ある1ファイルで根拠コメント無しの箇所を1件減らしつつ別ファイルで1件増やすと
 * 合計が変わらず検出できない。coverage-ratchet.mjs と同じトレードオフ。
 * 個別の新規混入を確実に捕捉したい場合は per-file baseline への拡張を検討する。
 */

import { readFileSync } from "node:fs";
import { readdir } from "node:fs/promises";
import path from "node:path";

const DISABLE_RE = /eslint-disable(?:-next-line|-line)?\b(?!\S)/;
const ENABLE_RE = /eslint-enable\b/;
const RATIONALE_MARKER = " -- ";

const SCAN_ROOTS = ["src", "liff/src", "line-reserve/src"];
const SCAN_EXTENSIONS = new Set([".ts", ".tsx"]);
const EXCLUDE_DIR_NAMES = new Set(["node_modules", "dist", "generated"]);

function parseArgs(argv) {
  const args = { baseline: ".eslint-disable-baseline", cwd: process.cwd() };
  for (let i = 0; i < argv.length; i += 1) {
    const arg = argv[i];
    if (arg === "--baseline") args.baseline = argv[++i];
    else if (arg === "--cwd") args.cwd = argv[++i];
  }
  return args;
}

/**
 * hasRationale はコメント本文（eslint-disable を含む1行）に根拠テキストがあるかを判定する。
 */
export function hasRationale(line) {
  return line.includes(RATIONALE_MARKER);
}

/**
 * findDisableLines はファイル内容（改行区切り済み）から eslint-disable 系の行を抽出する
 * （eslint-enable の閉じマーカーは対象外）。
 */
export function findDisableLines(lines) {
  const results = [];
  lines.forEach((line, i) => {
    if (ENABLE_RE.test(line)) return;
    if (!DISABLE_RE.test(line)) return;
    results.push({ lineNumber: i + 1, text: line.trim(), hasRationale: hasRationale(line) });
  });
  return results;
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
 *  - current > baseline: fail=true（根拠コメント無し disable が新規に増えた）。
 *  - current <= baseline: OK。
 */
export function evaluateRatchet(current, baseline) {
  if (current > baseline) {
    return {
      fail: true,
      message: `eslint-disable-rationale: FAIL — no-rationale disable count ${current} exceeds baseline ${baseline}. New eslint-disable comments must include a rationale (" -- reason") unless they are pre-existing baseline debt.`,
    };
  }
  return {
    fail: false,
    message: `eslint-disable-rationale: OK — no-rationale disable count ${current} <= baseline ${baseline}.`,
  };
}

async function main() {
  const args = parseArgs(process.argv.slice(2));

  const allFiles = [];
  for (const root of SCAN_ROOTS) {
    const dir = path.join(args.cwd, root);
    allFiles.push(...(await walk(dir, SCAN_EXTENSIONS, EXCLUDE_DIR_NAMES)));
  }

  const noRationale = [];
  for (const file of allFiles) {
    const text = readFileSync(file, "utf-8");
    const lines = text.split("\n");
    for (const entry of findDisableLines(lines)) {
      if (!entry.hasRationale) {
        noRationale.push({ file: path.relative(args.cwd, file), ...entry });
      }
    }
  }

  let baseline = 0;
  try {
    baseline = readBaseline(readFileSync(path.join(args.cwd, args.baseline), "utf-8"));
  } catch {
    baseline = 0;
  }

  const result = evaluateRatchet(noRationale.length, baseline);
  console.log(result.message);
  if (noRationale.length > 0) {
    console.log("no-rationale eslint-disable locations:");
    for (const entry of noRationale) {
      console.log(`  ${entry.file}:${entry.lineNumber}: ${entry.text}`);
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
