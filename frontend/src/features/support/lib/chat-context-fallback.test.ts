/**
 * chat-context-fallback.test.ts — UNIT-05 弱ヒットフォールバックの単体テスト
 *
 * buildChatContext の3分岐を直接検証する:
 * - 弱ヒット（pre-gate Fuse 候補あり・根拠ゲート落ち）→ text 空の候補リンク 1〜3 件
 * - 真のノーマッチ（候補 0 件 / 意味語 0 件の早期リターン）→ 空配列
 * - 根拠ゲート通過時 → 候補リンクを混ぜない（text 空の項目は出ない）
 *
 * pre-gate 候補数は検索関数をラップして collect() と同じ重複排除で計測する
 * （内部 Map を直接触らず、本番経路を変えずに観測する）。
 */
import { describe, expect, it, vi } from "vitest";

import { createManualSearcher, screenArticles, workflowArticles } from "@/lib/manual-index";
import type { ManualArticle } from "@/lib/manual-index";

import { buildChatContext, buildKeywordCandidates } from "./chat-context";
import { evalCases } from "./eval-cases";

const articles = [...screenArticles, ...workflowArticles];
const TOP_K = 3;
const EXCERPT_LENGTH = 1800;

/** buildChatContext 内部の candidates Map と同じ category/slug 重複排除で候補数を計測する */
function countPreGateCandidates(question: string): number {
  const search = createManualSearcher(articles);
  const seen = new Set<string>();
  const instrumented = (q: string) => {
    const results = search(q);
    for (const a of results) seen.add(`${a.category}/${a.slug}`);
    return results;
  };
  buildChatContext(instrumented, question, TOP_K, EXCERPT_LENGTH);
  return seen.size;
}

describe("buildChatContext 弱ヒットフォールバック", () => {
  it("弱ヒット: pre-gate 候補があり根拠ゲートを通らない場合、text 空の候補リンクを返す", () => {
    // 計測済み弱ヒット（pre-gate 候補 2 件、根拠ゲート通過 0 件）
    const items = buildChatContext(
      createManualSearcher(articles),
      "今日の天気は？",
      TOP_K,
      EXCERPT_LENGTH,
    );
    expect(items.length).toBeGreaterThanOrEqual(1);
    expect(items.length).toBeLessThanOrEqual(3);
    for (const item of items) {
      // ゲート落ちした抜粋は絶対に出さない（タイトル・リンク識別子のみ）
      expect(item.text).toBe("");
      expect(item.title.length).toBeGreaterThan(0);
      expect(item.slug.length).toBeGreaterThan(0);
      expect(["screens", "workflows"]).toContain(item.category);
    }
  });

  it("弱ヒット候補は最大 3 件に制限される（pre-gate 候補が数十件でも）", () => {
    // 計測済み: "Hôm nay trời có mưa không?" は pre-gate 候補 49 件の弱ヒット
    const items = buildChatContext(
      createManualSearcher(articles),
      "Hôm nay trời có mưa không?",
      TOP_K,
      EXCERPT_LENGTH,
    );
    expect(items).toHaveLength(3);
    for (const item of items) expect(item.text).toBe("");
  });

  it("真のノーマッチ（pre-gate 候補 0 件）→ 空配列（弱ヒット候補は出ない）", () => {
    const items = buildChatContext(
      createManualSearcher(articles),
      "近所でおいしいラーメン屋を教えて",
      TOP_K,
      EXCERPT_LENGTH,
    );
    expect(items).toEqual([]);
  });

  it("意味語 0 件（ストップワードのみ）→ 空配列（groups.length===0 早期リターン維持）", () => {
    const items = buildChatContext(
      createManualSearcher(articles),
      "教えてください",
      TOP_K,
      EXCERPT_LENGTH,
    );
    expect(items).toEqual([]);
  });

  it("意味語のない質問は強い候補が検索可能でも検索関数を呼ばない", () => {
    const search = vi.fn((_query: string) => articles);
    for (const question of ["方法", "cách"]) {
      search.mockClear();
      expect(buildKeywordCandidates(search, question)).toEqual([]);
      expect(search).not.toHaveBeenCalled();
    }
  });

  it("根拠ゲート通過時は弱ヒット候補を混ぜない（text 空の項目は出ない）", () => {
    const items = buildChatContext(
      createManualSearcher(articles),
      "カルテの作成方法は？",
      TOP_K,
      EXCERPT_LENGTH,
    );
    expect(items.length).toBeGreaterThan(0);
    for (const item of items) {
      expect(item.text.length).toBeGreaterThan(0);
    }
  });

  it("expectedEmpty コントロールは pre-gate Fuse 候補が 0 件（真のノーマッチ境界）", () => {
    // UNIT-05 の境界宣言を回帰ガード化: 記事・同義語・ストップワードの変更で
    // 候補が出始めたコントロールは弱ヒットなので expectedEmpty から外す。
    for (const c of evalCases.filter((x) => x.expectedEmpty === true)) {
      expect(
        countPreGateCandidates(c.question),
        `control "${c.question}" に pre-gate 候補 — ノーマッチではなく弱ヒット`,
      ).toBe(0);
    }
  });
});

describe("buildKeywordCandidates slug contract", () => {
  // backend candidateSlugRe (^[A-Za-z0-9][A-Za-z0-9_-]{0,99}$) と同一契約。
  // upsert は slug 非空のみ検証するため、契約外 slug のオーバーライド記事が
  // Fuse 候補に混ざっても 400 "invalid keyword candidate" にならないことを検証。
  const mkArticle = (slug: string): ManualArticle => ({
    category: "screens",
    slug,
    title: slug,
    order: 1,
    section: "s",
    content: "c",
    searchText: slug,
    keywords: [],
    keywords_vi: [],
  });

  it("契約外 slug の記事は候補から除外される", () => {
    const pool = [
      mkArticle("ok-slug"),
      mkArticle("日本語"),
      mkArticle("-draft"),
      mkArticle("a.b"),
      mkArticle("also_ok-1"),
    ];
    const got = buildKeywordCandidates(() => pool, "カルテの作成方法は？");
    expect(got.map((c) => c.slug).sort()).toEqual(["also_ok-1", "ok-slug"]);
  });

  it("全候補が契約外なら空配列（400 にならない）", () => {
    const pool = [mkArticle("日本語"), mkArticle("a.b")];
    expect(buildKeywordCandidates(() => pool, "カルテの作成方法は？")).toEqual([]);
  });
});
