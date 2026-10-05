/**
 * manual-index.ts — Manual MD ファイルの glob import + frontmatter 解析
 *
 * Vite の import.meta.glob で `?raw` 取得し、frontmatter (---) を解析する。
 * バンドル時に静的解決されるため、ランタイム fetch やネットワーク I/O は発生しない。
 */

import Fuse from "fuse.js";

export type ManualCategory = "screens" | "workflows";

export interface ManualArticle {
  category: ManualCategory;
  slug: string;
  title: string;
  order: number;
  section: string;
  content: string;
  searchText: string;
  /**
   * frontmatter `keywords:` の検索用言い換え語（カンマ区切り）。
   * DB オーバーライド行には存在しないため optional。
   */
  keywords?: string[];
  /**
   * frontmatter `keywords_vi:` のベトナム語検索語（カンマ区切り）。
   * VN の質問を JA 記事へ導くための語彙。DB オーバーライド行には
   * 存在しないため optional。
   */
  keywords_vi?: string[];
}

export interface RawFrontmatter {
  title?: string;
  order?: number;
  section?: string;
  keywords?: string[];
  keywords_vi?: string[];
}

const FRONTMATTER_RE = /^---\n([\s\S]*?)\n---\n([\s\S]*)$/;

/**
 * `keywords:` / `keywords_vi:` の値を分割する。
 * 1 行のカンマ区切り（frontmatter は行単位・最初のコロン分割のため
 * YAML ブロックリストは使えない）。空要素は捨てる。
 */
function splitKeywordList(value: string): string[] {
  return value
    .split(/[,、]/)
    .map((k) => k.trim())
    .filter((k) => k.length > 0);
}

export function parseFrontmatter(raw: string): { meta: RawFrontmatter; body: string } {
  const m = FRONTMATTER_RE.exec(raw);
  if (!m) return { meta: {}, body: raw };
  const [, fmBlock, body] = m;
  const meta: RawFrontmatter = {};
  for (const line of fmBlock.split("\n")) {
    const idx = line.indexOf(":");
    if (idx < 0) continue;
    const key = line.slice(0, idx).trim();
    const value = line
      .slice(idx + 1)
      .trim()
      .replace(/^["']|["']$/g, "");
    if (key === "title") meta.title = value;
    else if (key === "order") meta.order = Number(value);
    else if (key === "section") meta.section = value;
    else if (key === "keywords") {
      meta.keywords = splitKeywordList(value);
    } else if (key === "keywords_vi") {
      meta.keywords_vi = splitKeywordList(value);
    }
  }
  return { meta, body };
}

function pathToSlug(path: string): string {
  const filename = path.split("/").pop() ?? "";
  return filename.replace(/\.md$/, "");
}

// eager: true により ビルド時バンドル。? raw でファイル内容を文字列取得。
// コンテンツコーパスは features/manual が所有。ここは索引・検索機械のみ提供する。
const screenModules = import.meta.glob("../features/manual/content/screens/*.md", {
  eager: true,
  query: "?raw",
  import: "default",
}) as Record<string, string>;

const workflowModules = import.meta.glob("../features/manual/content/workflows/*.md", {
  eager: true,
  query: "?raw",
  import: "default",
}) as Record<string, string>;

function build(category: ManualCategory, modules: Record<string, string>): ManualArticle[] {
  const list: ManualArticle[] = [];
  for (const [path, raw] of Object.entries(modules)) {
    const slug = pathToSlug(path);
    const { meta, body } = parseFrontmatter(raw);
    const title = meta.title ?? slug;
    const order = meta.order ?? 9999;
    const section = meta.section ?? "その他";
    list.push({
      category,
      slug,
      title,
      order,
      section,
      content: body,
      // 検索用: markdown 記号を粗く除去
      searchText: `${title} ${section} ${body.replace(/[#*`>[\]()\-_!]/g, " ")}`,
      keywords: meta.keywords,
      keywords_vi: meta.keywords_vi,
    });
  }
  list.sort((a, b) => a.order - b.order);
  return list;
}

export const screenArticles: ManualArticle[] = build("screens", screenModules);
export const workflowArticles: ManualArticle[] = build("workflows", workflowModules);

const MANUAL_FUSE_OPTIONS = {
  keys: [
    { name: "title", weight: 0.5 },
    { name: "keywords", weight: 0.3 },
    { name: "keywords_vi", weight: 0.3 },
    { name: "section", weight: 0.1 },
    { name: "searchText", weight: 0.1 },
  ],
  threshold: 0.4,
  ignoreLocation: true,
  minMatchCharLength: 2,
  // score は includeScore 指定時のみ返る（採点・ソートに必須）
  includeScore: true,
};

/**
 * Fuse のマッチスコア付き記事。score は 0 が最良・1 が最悪。
 * チャットのコンテキスト構築で弱いヒットを弾く採点に使う。
 */
export interface ScoredManualArticle extends ManualArticle {
  matchScore: number;
}

/**
 * 命令型のマニュアル検索関数を生成する。React の外（イベントハンドラ等）からも
 * 呼べるよう、Fuse インスタンスを一度だけ構築して返す。
 * useManualSearch はこの関数の React ラッパー。
 */
export function createManualSearcher(
  articles: ManualArticle[],
): (query: string) => ScoredManualArticle[] {
  const fuse = new Fuse(articles, MANUAL_FUSE_OPTIONS);
  return (query: string) => {
    const q = query.trim();
    if (q.length === 0) return articles.map((a) => ({ ...a, matchScore: 0 }));
    return fuse.search(q).map((r) => ({ ...r.item, matchScore: r.score ?? 1 }));
  };
}

export function groupBySection(
  articles: ManualArticle[],
): { section: string; items: ManualArticle[] }[] {
  const map = new Map<string, ManualArticle[]>();
  for (const a of articles) {
    const arr = map.get(a.section) ?? [];
    arr.push(a);
    map.set(a.section, arr);
  }
  return Array.from(map.entries()).map(([section, items]) => ({ section, items }));
}

/**
 * DB オーバーライドを既存記事リストに適用する。
 * 同じ category+slug の記事があれば内容を置換、無ければ追加。
 * order でソート。
 */
export interface ArticleOverride {
  category: ManualCategory;
  slug: string;
  title: string;
  order_value: number;
  section: string;
  body_markdown: string;
}

export function applyOverrides(
  base: ManualArticle[],
  overrides: ArticleOverride[],
): ManualArticle[] {
  if (overrides.length === 0) return base;
  const byKey = new Map<string, ManualArticle>();
  for (const a of base) {
    byKey.set(`${a.category}/${a.slug}`, a);
  }
  for (const o of overrides) {
    const key = `${o.category}/${o.slug}`;
    const article: ManualArticle = {
      category: o.category,
      slug: o.slug,
      title: o.title,
      order: o.order_value,
      section: o.section,
      content: o.body_markdown,
      searchText: `${o.title} ${o.section} ${o.body_markdown.replace(/[#*`>[\]()\-_!]/g, " ")}`,
      // DB 行は frontmatter を持たず keywords/keywords_vi を保持できない。
      // 置換対象のバンドル記事がそれらを持つ場合は引き継ぐ
      // （言い換え語の検索カバレッジを DB override で失わないため）。
      // DB だけの新規記事は keywords/keywords_vi 無し（undefined）のまま。
      keywords: byKey.get(key)?.keywords,
      keywords_vi: byKey.get(key)?.keywords_vi,
    };
    byKey.set(key, article);
  }
  const merged = Array.from(byKey.values());
  merged.sort((a, b) => a.order - b.order);
  return merged;
}
