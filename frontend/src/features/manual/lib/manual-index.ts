/**
 * manual-index.ts — Manual MD ファイルの glob import + frontmatter 解析
 *
 * Vite の import.meta.glob で `?raw` 取得し、frontmatter (---) を解析する。
 * バンドル時に静的解決されるため、ランタイム fetch やネットワーク I/O は発生しない。
 */

export type ManualCategory = "screens" | "workflows";

export interface ManualArticle {
  category: ManualCategory;
  slug: string;
  title: string;
  order: number;
  section: string;
  content: string;
  searchText: string;
}

export interface RawFrontmatter {
  title?: string;
  order?: number;
  section?: string;
}

const FRONTMATTER_RE = /^---\n([\s\S]*?)\n---\n([\s\S]*)$/;

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
  }
  return { meta, body };
}

function pathToSlug(path: string): string {
  const filename = path.split("/").pop() ?? "";
  return filename.replace(/\.md$/, "");
}

// eager: true により ビルド時バンドル。? raw でファイル内容を文字列取得。
const screenModules = import.meta.glob("../content/screens/*.md", {
  eager: true,
  query: "?raw",
  import: "default",
}) as Record<string, string>;

const workflowModules = import.meta.glob("../content/workflows/*.md", {
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
    });
  }
  list.sort((a, b) => a.order - b.order);
  return list;
}

export const screenArticles: ManualArticle[] = build("screens", screenModules);
export const workflowArticles: ManualArticle[] = build("workflows", workflowModules);

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
    };
    byKey.set(key, article);
  }
  const merged = Array.from(byKey.values());
  merged.sort((a, b) => a.order - b.order);
  return merged;
}
