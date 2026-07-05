#!/usr/bin/env node
/**
 * scripts/check-design-primary-cta.mjs
 *
 * DESIGN.md primary CTA への旧 accent 再混入を検出する（Wave 0–7 完了レポート P1 ガード）。
 *
 * 検査対象:
 *   frontend/src/features/**
 *   frontend/src/components/shared/**
 *
 * 対象コンポーネント:
 *   SubmitButton / PrimaryButton / <Button type="submit"
 *
 * 禁止トークン（className 等に明示的に旧 accent CTA を指定）:
 *   C.bgAccent, STYLE.confirmPrimary, btnAccent, sidePeekSaveBtn
 *
 * SubmitButton / PrimaryButton は既定で DESIGN.md button-primary（brand blue + pill）を
 * 適用するため、`colorVariant="default"` は旧 accent CTA への opt-out（再混入）とみなして検出する。
 *
 * 許容（検出しない）:
 *   design-tokens.ts（STYLE 定義本体）
 *   *.test.tsx / *.test.ts（fixture）
 *   コメント行・JSX コメント
 *   非 CTA の C.accent / C.bgAccentLight 等（本スクリプトは上記4トークンのみ）
 *   SubmitButton.tsx / PrimaryButton.tsx（default variant 実装）
 *
 * Usage: node scripts/check-design-primary-cta.mjs [--root <repo-root>]
 * Exit: 0 = clean, 1 = violation
 */

import { readFileSync, readdirSync, statSync } from "node:fs";
import path from "node:path";

const FORBIDDEN = [
  { id: "C.bgAccent", re: /\bC\.bgAccent\b/ },
  { id: "STYLE.confirmPrimary", re: /\bSTYLE\.confirmPrimary\b/ },
  { id: "btnAccent", re: /\b(btnAccent|STYLE\.btnAccent)\b/ },
  { id: "sidePeekSaveBtn", re: /\bsidePeekSaveBtn\b/ },
];

// SubmitButton/PrimaryButton は既定で brand（button-primary）を適用するため、
// colorVariant="default" は旧 accent CTA への明示的な opt-out（再混入）として扱う。
const FORBIDDEN_COLOR_VARIANT_DEFAULT = {
  id: 'colorVariant="default"',
  re: /colorVariant\s*=\s*["']default["']/,
};

const SCAN_REL_DIRS = [
  "frontend/src/features",
  "frontend/src/components/shared",
];

const EXCLUDE_FILE = /(?:\.test\.(?:tsx?|ts)$|design-tokens\.ts$|SubmitButton\.tsx$|PrimaryButton\.tsx$)/;

const BUTTON_TAG_RE = /<(SubmitButton|PrimaryButton|Button)\b[\s\S]*?(?:\/>|<\/\1>)/g;

function parseArgs(argv) {
  let root = path.resolve(path.dirname(new URL(import.meta.url).pathname), "..");
  if (process.platform === "win32" && root.startsWith("/")) {
    // noop — macOS/linux only in this repo
  }
  for (let i = 2; i < argv.length; i += 1) {
    if (argv[i] === "--root") {
      root = path.resolve(argv[++i]);
    }
  }
  return { root };
}

function walk(dir, out = []) {
  for (const name of readdirSync(dir)) {
    if (name === "node_modules" || name === "generated") continue;
    const full = path.join(dir, name);
    const st = statSync(full);
    if (st.isDirectory()) {
      walk(full, out);
    } else if (/\.tsx?$/.test(name)) {
      out.push(full);
    }
  }
  return out;
}

function stripJsxComments(text) {
  return text
    .replace(/\{\/\*[\s\S]*?\*\/\}/g, "")
    .replace(/\/\*[\s\S]*?\*\//g, "")
    .replace(/\/\/[^\n]*/g, "");
}

function isPrimaryCtaBlock(block) {
  if (block.startsWith("<SubmitButton") || block.startsWith("<PrimaryButton")) {
    return true;
  }
  if (block.startsWith("<Button") && /type\s*=\s*["']submit["']/.test(block)) {
    return true;
  }
  return false;
}

function findViolations(content, filePath) {
  const hits = [];
  const strippedFileComments = content.replace(/^\s*\/\/.*$/gm, "");
  let match;
  BUTTON_TAG_RE.lastIndex = 0;
  while ((match = BUTTON_TAG_RE.exec(strippedFileComments)) !== null) {
    const block = match[0];
    if (!isPrimaryCtaBlock(block)) continue;
    const inspect = stripJsxComments(block);
    for (const rule of FORBIDDEN) {
      if (rule.re.test(inspect)) {
        hits.push({ file: filePath, rule: rule.id, snippet: block.slice(0, 120).replace(/\s+/g, " ") });
      }
    }
    const isSubmitOrPrimaryButton = block.startsWith("<SubmitButton") || block.startsWith("<PrimaryButton");
    if (isSubmitOrPrimaryButton && FORBIDDEN_COLOR_VARIANT_DEFAULT.re.test(inspect)) {
      hits.push({
        file: filePath,
        rule: FORBIDDEN_COLOR_VARIANT_DEFAULT.id,
        snippet: block.slice(0, 120).replace(/\s+/g, " "),
      });
    }
  }
  return hits;
}

function main() {
  const { root } = parseArgs(process.argv);
  const allHits = [];

  for (const rel of SCAN_REL_DIRS) {
    const dir = path.join(root, rel);
    try {
      statSync(dir);
    } catch {
      console.error(`FAIL  scan directory not found: ${dir}`);
      process.exit(1);
    }
    for (const file of walk(dir)) {
      if (EXCLUDE_FILE.test(file)) continue;
      const relFile = path.relative(root, file);
      const content = readFileSync(file, "utf8");
      allHits.push(...findViolations(content, relFile));
    }
  }

  if (allHits.length === 0) {
    console.log("PASS  no primary CTA accent reintroduction detected");
    process.exit(0);
  }

  console.error(`FAIL  ${allHits.length} primary CTA accent violation(s):`);
  for (const hit of allHits) {
    console.error(`  - ${hit.file} [${hit.rule}] ${hit.snippet}…`);
  }
  process.exit(1);
}

main();
