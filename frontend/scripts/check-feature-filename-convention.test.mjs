// scripts/check-feature-filename-convention.test.mjs
//
// FE-RC-017 followup2: isTsxViolation の *.test/*.spec / use-* 除外回帰。
// 実行: node --test scripts/check-feature-filename-convention.test.mjs

import test from "node:test";
import assert from "node:assert/strict";
import { isTsxViolation } from "./check-feature-filename-convention.mjs";

const jsxSample = 'export function Foo() { return <div className="x" />; }';

test("isTsxViolation: *.test basename は JSX があっても false", () => {
  assert.equal(isTsxViolation("foo.test", jsxSample), false);
});

test("isTsxViolation: *.spec basename は JSX があっても false", () => {
  assert.equal(isTsxViolation("foo.spec", jsxSample), false);
});

test("isTsxViolation: use-* basename は JSX があっても false", () => {
  assert.equal(isTsxViolation("use-reception-column-view", jsxSample), false);
});

test("isTsxViolation: 小文字始まり + JSX は true", () => {
  assert.equal(isTsxViolation("foo-bar", jsxSample), true);
});

test("isTsxViolation: PascalCase は false", () => {
  assert.equal(isTsxViolation("FooBar", jsxSample), false);
});
