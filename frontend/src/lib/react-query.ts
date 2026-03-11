import { QueryClient, type DefaultOptions } from "@tanstack/react-query";

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
});

// リソース別キャッシング戦略
export const QUERY_STALE_TIMES = {
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
