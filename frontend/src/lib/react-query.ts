import { QueryClient, QueryCache, type DefaultOptions } from "@tanstack/react-query";
import { handleApiError } from "@/lib/handle-api-error";

// FE5-16: 個別 UI で処理済みを明示した query は silentError でスキップできる
declare module "@tanstack/react-query" {
  interface Register {
    queryMeta: {
      silentError?: boolean;
      errorContext?: string;
    };
  }
}

const queryConfig: DefaultOptions = {
  queries: {
    // デフォルト: 中程度の変更頻度（医療記録、検査等）
    staleTime: 5 * 60 * 1000,
    gcTime: 15 * 60 * 1000,
    refetchOnWindowFocus: false,
    retry: 1,
  },
};

export const queryClient = new QueryClient({
  defaultOptions: queryConfig,
  queryCache: new QueryCache({
    onError: (error, query) => {
      if (query.meta?.silentError) return;
      handleApiError(error, query.meta?.errorContext ?? "データ取得");
    },
  }),
});

// リソース別キャッシング戦略
export const QUERY_STALE_TIMES = {
  // セッション: 認証 me 等（FE-RC-082）
  SESSION: 10 * 1000,      // 10秒

  // 常に再取得: preview 系（FE-RC-082）
  NONE: 0,

  // 極短期: 会計の一時集計・提案系（FE3-7 新設。値は既存直値のまま）
  SHORT: 30_000,           // 30秒

  // 短期: lstep配信トリガー系（FE3-7 新設。値は既存直値のまま）
  MINUTE: 60 * 1000,       // 1分

  // 低頻度変更: 飼主・ペット・マスタ等
  STATIC: 30 * 60 * 1000,  // 30分

  // 中程度: 医療記録、検査、会計等
  MEDIUM: 5 * 60 * 1000,   // 5分（デフォルト）

  // 高頻度: リアルタイム予約、Kanban等
  REALTIME: 2 * 60 * 1000, // 2分
};

export const QUERY_GC_TIMES = {
  // 長期保持: マスタデータ等
  LONG: 30 * 60 * 1000,     // 30分

  // 標準保持: ほとんどのデータ
  STANDARD: 15 * 60 * 1000, // 15分

  // 短期保持: 一時的なUI状態
  SHORT: 5 * 60 * 1000,     // 5分
};
