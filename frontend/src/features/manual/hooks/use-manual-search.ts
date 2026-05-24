/**
 * use-manual-search.ts — Manual 全文検索フック
 *
 * Fuse.js を使った fuzzy search。検索クエリ空文字時は全件返却。
 * articles 引数で DB オーバーライド適用済の最新リストを受け取る。
 */

import Fuse from "fuse.js";
import { useMemo } from "react";

import type { ManualArticle } from "../lib/manual-index";

const FUSE_OPTIONS = {
  keys: [
    { name: "title", weight: 0.6 },
    { name: "section", weight: 0.2 },
    { name: "searchText", weight: 0.2 },
  ],
  threshold: 0.4,
  ignoreLocation: true,
  minMatchCharLength: 2,
};

export function useManualSearch(query: string, articles: ManualArticle[]): ManualArticle[] {
  const fuse = useMemo(() => new Fuse(articles, FUSE_OPTIONS), [articles]);

  return useMemo(() => {
    const q = query.trim();
    if (q.length === 0) return articles;
    return fuse.search(q).map((r) => r.item);
  }, [articles, fuse, query]);
}
