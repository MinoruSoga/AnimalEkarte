/**
 * get-manual-articles.ts — マニュアル記事一覧の取得 API
 *
 * バックエンドにオーバーライド版のマニュアルが保存されていれば取得。
 * 取得失敗時は無視 (empty array)。MD ファイルバンドル版が引き続き利用される。
 */

import { useQuery } from "@tanstack/react-query";

import { axios } from "@/lib/axios";
import { queryKeys } from "@/lib/query-keys";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";

import type { ManualCategory } from "../lib/manual-index";

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

async function getManualArticleOverrides(): Promise<ManualArticleOverride[]> {
  try {
    const { data } = await axios.get<ListResponse>("/v1/manual/articles");
    return data.data;
  } catch {
    // backend 未到達・未認証等の場合は MD ファイル版のみ使用
    return [];
  }
}

export function useGetManualArticleOverrides(enabled: boolean) {
  return useQuery({
    queryKey: queryKeys.manualArticles.all(),
    queryFn: getManualArticleOverrides,
    enabled,
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
}
