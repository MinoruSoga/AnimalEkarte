/**
 * retrieval-eval.test.ts — マニュアル検索の recall@3 評価 + 回帰アサート
 *
 * eval-cases.ts の各ケースに対し、本番と同じ経路
 * （buildChatContext + createManualSearcher + バンドル済み記事）で top-3 を取り、
 * 期待記事の有無を採点して per-case 結果とサマリを console に出力する。
 *
 * UNIT-02 以降のアサート:
 * - JA パラフレーズ集合は recall@3 = 100%（全ケース HIT）
 * - expectedEmpty コントロールはコンテキストが必ず空（フォールバック維持）
 * - UNIT-03 以降: VN ケース集合も recall@3 = 100%（keywords_vi + VN クエリ処理）、
 *   かつ全記事が keywords_vi を持つ（frontmatter 記載の防御的確認）
 * - UNIT-05 以降: expectedWeak コントロールは弱ヒット候補が text 空で
 *   1〜3 件返る（参照リンクのフォールバック）
 */
import { beforeAll, describe, expect, it } from "vitest";

import { createManualSearcher, screenArticles, workflowArticles } from "@/lib/manual-index";

import { buildChatContext } from "./chat-context";
import { evalCases, type EvalCase } from "./eval-cases";

/** 本番のチャットコンテキスト構築と同じパラメータ */
const TOP_K = 3;
const EXCERPT_LENGTH = 1800;

const articles = [...screenArticles, ...workflowArticles];
const articleKeys = new Set(articles.map((a) => `${a.category}/${a.slug}`));

describe("eval fixture", () => {
  it("各ケースが question・lang・expectedSlugs|expectedEmpty|expectedWeak（排他）を持つ", () => {
    expect(evalCases.length).toBeGreaterThan(0);
    for (const c of evalCases) {
      expect(c.question.length).toBeGreaterThan(0);
      expect(["ja", "vi"]).toContain(c.lang);
      const hasSlugs = c.expectedSlugs !== undefined && c.expectedSlugs.length > 0;
      const hasEmpty = c.expectedEmpty === true;
      const hasWeak = c.expectedWeak === true;
      expect([hasSlugs, hasEmpty, hasWeak].filter(Boolean)).toHaveLength(1);
    }
  });

  it("expectedSlugs はカテゴリ修飾付きで記事インデックスに実在する", () => {
    for (const c of evalCases) {
      for (const slug of c.expectedSlugs ?? []) {
        expect(slug).toMatch(/^(screens|workflows)\/[^/]+$/);
        expect(articleKeys.has(slug)).toBe(true);
      }
    }
  });

  it("件数下限（JA>=8・VN>=6・control>=3）と報告済みケースを満たす", () => {
    const ja = evalCases.filter((c) => c.lang === "ja" && c.expectedSlugs !== undefined);
    const vi = evalCases.filter((c) => c.lang === "vi" && c.expectedSlugs !== undefined);
    const controls = evalCases.filter((c) => c.expectedEmpty === true);
    expect(ja.length).toBeGreaterThanOrEqual(8);
    expect(vi.length).toBeGreaterThanOrEqual(6);
    expect(controls.length).toBeGreaterThanOrEqual(3);
    expect(
      evalCases.some(
        (c) =>
          c.question === "カルテの作成方法は？" &&
          c.expectedSlugs?.includes("screens/04-medical-records"),
      ),
    ).toBe(true);
  });
});

interface ScoredCase {
  readonly question: string;
  readonly lang: EvalCase["lang"];
  readonly isControl: boolean;
  readonly isWeak: boolean;
  readonly expectedLabel: string;
  readonly got: readonly string[];
  /** 返ったコンテキスト項目がすべて text 空（弱ヒット候補の必須条件） */
  readonly allTextEmpty: boolean;
  readonly hit: boolean;
}

function scoreCases(): ScoredCase[] {
  const search = createManualSearcher(articles);
  return evalCases.map((c) => {
    const items = buildChatContext(search, c.question, TOP_K, EXCERPT_LENGTH);
    const got = items.map((item) => `${item.category}/${item.slug}`);
    const isControl = c.expectedEmpty === true;
    const isWeak = c.expectedWeak === true;
    const allTextEmpty = items.every((item) => item.text === "");
    const hit = isControl
      ? got.length === 0
      : isWeak
        ? got.length >= 1 && got.length <= TOP_K && allTextEmpty
        : (c.expectedSlugs ?? []).some((slug) => got.includes(slug));
    return {
      question: c.question,
      lang: c.lang,
      isControl,
      isWeak,
      expectedLabel: isControl ? "(empty)" : isWeak ? "(weak)" : (c.expectedSlugs ?? []).join(","),
      got,
      allTextEmpty,
      hit,
    };
  });
}

describe("recall@3 scorer", () => {
  // scoreCases() は全ケース分の Fuse 検索で数秒かかるため、採点は describe ごとに
  // 1 回だけ行い、各テストは結果のフィルタのみ行う
  // （テストごとの再採点は vitest デフォルト 5s タイムアウトを超過する）。
  let results: ScoredCase[] = [];
  beforeAll(() => {
    results = scoreCases();
  }, 120000);

  it("全ケースを採点して per-case 結果とレポートを出力する", () => {
    console.log("\n=== per-case recall@3 ===");
    for (const r of results) {
      console.log(
        `[eval] ${r.hit ? "HIT " : "MISS"} ${r.lang} "${r.question}" ` +
          `expected=${r.expectedLabel} got=[${r.got.join(", ")}]`,
      );
    }

    const report = (label: string, rs: readonly ScoredCase[]): string => {
      const hits = rs.filter((r) => r.hit).length;
      const pct = rs.length === 0 ? "n/a" : `${((hits / rs.length) * 100).toFixed(1)}%`;
      return `[eval] recall@${TOP_K} ${label}: ${hits}/${rs.length} (${pct})`;
    };
    console.log("=== recall report ===");
    console.log(
      report(
        "ja",
        results.filter((r) => r.lang === "ja" && !r.isControl && !r.isWeak),
      ),
    );
    console.log(
      report(
        "vi",
        results.filter((r) => r.lang === "vi" && !r.isControl && !r.isWeak),
      ),
    );
    console.log(
      report(
        "controls",
        results.filter((r) => r.isControl),
      ),
    );
    console.log(
      report(
        "weak",
        results.filter((r) => r.isWeak),
      ),
    );
    console.log(report("all", results));

    // スイート整合性: 全ケースが採点された
    expect(results).toHaveLength(evalCases.length);
  });

  it("JA パラフレーズ集合は recall@3 = 100%（全ケース HIT）", () => {
    for (const r of results.filter((r) => r.lang === "ja" && !r.isControl && !r.isWeak)) {
      expect(r.hit, `"${r.question}" -> [${r.got.join(", ")}]`).toBe(true);
    }
  });

  it("VN ケース集合は recall@3 = 100%（全ケース HIT）", () => {
    for (const r of results.filter((r) => r.lang === "vi" && !r.isControl && !r.isWeak)) {
      expect(r.hit, `"${r.question}" -> [${r.got.join(", ")}]`).toBe(true);
    }
  });

  it("全記事が keywords_vi を持つ（frontmatter 記載の防御的確認）", () => {
    for (const a of articles) {
      expect(
        a.keywords_vi?.length ?? 0,
        `${a.category}/${a.slug} に keywords_vi が無い`,
      ).toBeGreaterThan(0);
    }
  });

  it("expectedEmpty コントロールはコンテキストが空（フォールバック維持）", () => {
    for (const r of results.filter((r) => r.isControl)) {
      expect(r.got, `control "${r.question}" leaked [${r.got.join(", ")}]`).toHaveLength(0);
    }
  });

  it("expectedWeak ケースは text 空の候補リンクを 1〜3 件返す（弱ヒットフォールバック）", () => {
    for (const r of results.filter((r) => r.isWeak)) {
      expect(
        r.got.length,
        `weak "${r.question}" -> [${r.got.join(", ")}]（1〜3 件の候補が期待値）`,
      ).toBeGreaterThanOrEqual(1);
      expect(r.got.length).toBeLessThanOrEqual(TOP_K);
      expect(r.allTextEmpty, `weak "${r.question}" がゲート落ちした抜粋をリーク`).toBe(true);
    }
  });
});
