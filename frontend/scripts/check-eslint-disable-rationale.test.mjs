// scripts/check-eslint-disable-rationale.test.mjs
//
// check-eslint-disable-rationale.mjs の回帰テスト。Node 組み込みの test runner を使う
// （新規 pnpm 依存を追加しない — FE-refactor.md「eslint-disable 根拠コメントの機械強制」の
// 「既存/no-dependency harness」要件）。
//
// 実行: node --test scripts/check-eslint-disable-rationale.test.mjs

import test from "node:test";
import assert from "node:assert/strict";
import {
  hasRationale,
  findDisableLines,
  readBaseline,
  evaluateRatchet,
} from "./check-eslint-disable-rationale.mjs";

test("hasRationale: ' -- ' マーカー付きは true", () => {
  assert.equal(
    hasRationale("  // eslint-disable-next-line react-hooks/exhaustive-deps -- 理由あり"),
    true,
  );
});

test("hasRationale: マーカー無しは false", () => {
  assert.equal(hasRationale("  // eslint-disable-next-line react-hooks/exhaustive-deps"), false);
});

test("hasRationale: ハイフン単体（'-'）はマーカーとして誤検出しない", () => {
  assert.equal(
    hasRationale("  // eslint-disable-next-line no-control-regex - not a real rationale marker"),
    false,
  );
});

test("findDisableLines: line-comment / block-comment 両方を検出する", () => {
  const lines = [
    "function f() {",
    "  // eslint-disable-next-line react-hooks/exhaustive-deps",
    "  doSomething();",
    "  /* eslint-disable react-hooks/set-state-in-effect */",
    "  setState(1);",
    "  /* eslint-enable react-hooks/set-state-in-effect */",
    "}",
  ];
  const found = findDisableLines(lines);
  assert.equal(found.length, 2);
  assert.equal(found[0].lineNumber, 2);
  assert.equal(found[0].hasRationale, false);
  assert.equal(found[1].lineNumber, 4);
  assert.equal(found[1].hasRationale, false);
});

test("findDisableLines: eslint-enable は disable としてカウントしない", () => {
  const lines = ["/* eslint-enable react-hooks/set-state-in-effect */"];
  assert.equal(findDisableLines(lines).length, 0);
});

test("findDisableLines: 根拠コメント付きは hasRationale=true で検出する", () => {
  const lines = ["  // eslint-disable-next-line no-control-regex -- NULL バイト除去のため"];
  const found = findDisableLines(lines);
  assert.equal(found.length, 1);
  assert.equal(found[0].hasRationale, true);
});

test("readBaseline: 数値行を読む", () => {
  assert.equal(readBaseline("# comment\n22\n"), 22);
});

test("readBaseline: 空・コメントのみなら 0", () => {
  assert.equal(readBaseline("# comment only\n"), 0);
  assert.equal(readBaseline(""), 0);
});

test("evaluateRatchet: current <= baseline は PASS", () => {
  const result = evaluateRatchet(22, 22);
  assert.equal(result.fail, false);
});

test("evaluateRatchet: current < baseline も PASS（減少は許容）", () => {
  const result = evaluateRatchet(10, 22);
  assert.equal(result.fail, false);
});

test("evaluateRatchet: current > baseline は FAIL（新規混入検出）", () => {
  const result = evaluateRatchet(23, 22);
  assert.equal(result.fail, true);
  assert.match(result.message, /FAIL/);
});
