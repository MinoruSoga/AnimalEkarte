/**
 * use-manual-article-overrides.ts — マニュアル記事一覧の取得 API
 *
 * バックエンドにオーバーライド版のマニュアルが保存されていれば取得。
 * 取得失敗（404/401/403/5xx/通信障害/欠落 data 配列）は error として伝播する。
 * 既取得本文を表示し続けず、bundled 本文への黙った fallback は行わない。
 */

import { useQuery } from "@tanstack/react-query";

import { axios } from "@/lib/axios";
import { queryKeys } from "@/lib/query-keys";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";

import type { ManualCategory } from "@/lib/manual-index";

export interface ManualArticleOverride {
  id: number;
  category: ManualCategory;
  slug: string;
  title: string;
  order_value: number;
  section: string;
  body_markdown: string;
  updated_by_staff_id?: number;
  created_at: string;
  updated_at: string;
}

interface ListResponse {
  data: ManualArticleOverride[];
}

// strict 経路のみ: graceful fallback なし。レスポンスの data 配列欠落も失敗扱い。
async function getManualArticleOverridesStrict(): Promise<ManualArticleOverride[]> {
  const { data } = await axios.get<ListResponse>("/v1/manual/articles");
  if (!Array.isArray(data.data)) {
    throw new Error("manual articles response is missing the data array");
  }
  return data.data;
}

export function useGetManualArticleOverrides(enabled: boolean, strict = false) {
  return useQuery({
    // strict は別キーに分離し、ManualHelpTab 等の既存 cache/invalidation と
    // 互換を維持する（param 自体は互換のため残置、取得は常に strict）。
    queryKey: strict
      ? [...queryKeys.manualArticles.all(), "strict"]
      : queryKeys.manualArticles.all(),
    queryFn: getManualArticleOverridesStrict,
    enabled,
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
}
