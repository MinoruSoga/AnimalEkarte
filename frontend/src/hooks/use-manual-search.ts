/**
 * use-manual-search.ts — Manual 全文検索フック
 *
 * Fuse.js を使った fuzzy search。検索クエリ空文字時は全件返却。
 * articles 引数で DB オーバーライド適用済の最新リストを受け取る。
 * 検索実装は @/lib/manual-index の createManualSearcher に集約し、
 * このフックは React 用の memo ラッパーのみを担う。
 */

import { useMemo } from "react";

import { createManualSearcher, type ManualArticle } from "@/lib/manual-index";

export function useManualSearch(query: string, articles: ManualArticle[]): ManualArticle[] {
  const search = useMemo(() => createManualSearcher(articles), [articles]);
  return useMemo(() => search(query), [search, query]);
}
