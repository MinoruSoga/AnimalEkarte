#!/usr/bin/env node

/**
 * Coverage ratchet — frontend のカバレッジが記録済みベースラインから低下していないかを判定する
 * （FE-refactor.md R-F5 / backend cmd/coverage-ratchet と対をなす・coverage-policy.md Phase 1 ratchet）。
 *
 * 実行:
 * $ node scripts/coverage-ratchet.mjs --summary coverage/coverage-summary.json --baseline .coverage-baseline
 *
 * coverage-summary.json は vite.config.ts の coverage.reporter に "json-summary" を含めることで
 * `pnpm run test:coverage` 実行時に生成される（istanbul json-summary 形式・total.statements.pct を参照）。
 *
 * ベースラインファイル（.coverage-baseline）は「frontend statements カバレッジ %」を1行の数値で持つ。
 * 未記録（プレースホルダ 0 / ファイル欠落）の場合は warn のみで exit 0 とし、初回 CI 実測での
 * ベースライン記録を促す（coverage-policy.md のベースライン記録手順。backend 側と同じ設計）。
 * ベースラインが記録済みなら現在値が baseline - tolerance を下回ったとき exit 非0（fail）で
 * リファクタによる低下を検出する。
 */

import { readFileSync } from "node:fs";

const DEFAULT_TOLERANCE_PP = 0.5;

function parseArgs(argv) {
  const args = {
    summary: "coverage/coverage-summary.json",
    baseline: ".coverage-baseline",
    tolerance: DEFAULT_TOLERANCE_PP,
    warnOnly: false,
  };
  for (let i = 0; i < argv.length; i += 1) {
    const arg = argv[i];
    if (arg === "--summary") args.summary = argv[++i];
    else if (arg === "--baseline") args.baseline = argv[++i];
    else if (arg === "--tolerance") args.tolerance = Number(argv[++i]);
    else if (arg === "--warn-only") args.warnOnly = true;
  }
  return args;
}

/**
 * parseTotalStatementsPct は vitest の json-summary reporter が出力する
 * coverage-summary.json から total.statements.pct を取り出す。
 */
export function parseTotalStatementsPct(summaryJsonText) {
  const parsed = JSON.parse(summaryJsonText);
  const pct = parsed?.total?.statements?.pct;
  if (typeof pct !== "number" || Number.isNaN(pct)) {
    throw new Error("coverage-summary.json に total.statements.pct が見つかりません");
  }
  return pct;
}

/**
 * readBaseline は baseline ファイルから記録済みカバレッジ % を読む。欠落・空・解析不能なら 0 を返す
 * （未記録扱い＝warn のみ）。
 */
export function readBaseline(fileText) {
  for (const rawLine of fileText.split("\n")) {
    const line = rawLine.trim();
    if (line === "" || line.startsWith("#")) continue;
    const value = Number(line.replace(/%$/, ""));
    if (!Number.isNaN(value)) return value;
  }
  return 0;
}

/**
 * evaluateRatchet は現在値とベースラインを比較する純粋関数。
 *  - baseline <= 0（未記録）: warn のみ、fail=false。初回 CI 実測でのベースライン記録を促す。
 *  - current >= baseline - tolerance: OK。
 *  - current < baseline - tolerance: fail=true（リファクタによるカバレッジ低下）。
 */
export function evaluateRatchet(current, baseline, tolerance) {
  if (baseline <= 0) {
    return {
      fail: false,
      message: `coverage-ratchet: baseline not recorded (current=${current.toFixed(1)}%). Record this value in .coverage-baseline / docs/ops/coverage-policy.md to arm the ratchet.`,
    };
  }
  const floor = baseline - tolerance;
  if (current < floor) {
    return {
      fail: true,
      message: `coverage-ratchet: FAIL — coverage ${current.toFixed(1)}% dropped below baseline ${baseline.toFixed(1)}% (tolerance ${tolerance.toFixed(1)}pp, floor ${floor.toFixed(1)}%). A refactor must not lower coverage; add tests or update the baseline with justification.`,
    };
  }
  return {
    fail: false,
    message: `coverage-ratchet: OK — coverage ${current.toFixed(1)}% >= floor ${floor.toFixed(1)}% (baseline ${baseline.toFixed(1)}%, tolerance ${tolerance.toFixed(1)}pp).`,
  };
}

function main() {
  const args = parseArgs(process.argv.slice(2));

  let current;
  try {
    const summaryText = readFileSync(args.summary, "utf-8");
    current = parseTotalStatementsPct(summaryText);
  } catch (err) {
    console.error(`coverage-ratchet: ${err instanceof Error ? err.message : String(err)}`);
    process.exit(2);
  }

  let baseline = 0;
  try {
    baseline = readBaseline(readFileSync(args.baseline, "utf-8"));
  } catch {
    baseline = 0;
  }

  const result = evaluateRatchet(current, baseline, args.tolerance);
  console.log(result.message);
  if (result.fail && !args.warnOnly) {
    process.exit(1);
  }
}

const isMain = process.argv[1] && import.meta.url === `file://${process.argv[1]}`;
if (isMain) {
  main();
}
