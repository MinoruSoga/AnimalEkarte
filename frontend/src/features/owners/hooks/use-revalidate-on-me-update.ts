import { useContext, useEffect, useLayoutEffect, useRef } from "react";
import { QueryClientContext, hashKey } from "@tanstack/react-query";
import { ME_QUERY_KEY } from "@/lib/query-keys";

const ME_QUERY_HASH = hashKey(ME_QUERY_KEY);

/**
 * EMR-463: 全体 `ME_QUERY_KEY` のデータが更新されたら（refreshPermissions・login 後の再取得等）
 * owners loader を再検証させる。
 *
 * cached profile は「再検証のきっかけ」にだけ使い、現在許可の根拠には使わない
 * （許可集合は loader が owners 専有 adapter で GET /me を取り直して決める）。
 * QueryClientContext を直接引くのは、provider の無い単体 render でも落とさないため
 * （use-copy-treatment-details.ts と同じ流儀）。
 */
export function useRevalidateOnMeUpdate(revalidate: () => void, enabled = true) {
  const queryClient = useContext(QueryClientContext);
  const revalidateRef = useRef(revalidate);
  useLayoutEffect(() => {
    revalidateRef.current = revalidate;
  }, [revalidate]);

  useEffect(() => {
    if (!enabled || !queryClient) return undefined;
    const cache = queryClient.getQueryCache();
    // 購読開始時点の値を基準にし、mount 直後に既存 cache だけで再検証しない。
    let lastUpdatedAt =
      cache.find({ queryKey: ME_QUERY_KEY, exact: true })?.state.dataUpdatedAt ?? 0;
    return cache.subscribe((event) => {
      if (event.type !== "updated") return;
      if (event.query.queryHash !== ME_QUERY_HASH) return;
      const updatedAt = event.query.state.dataUpdatedAt;
      // 取得失敗・fetch 開始等の data を伴わない更新では再検証しない。
      if (updatedAt === 0 || updatedAt === lastUpdatedAt) return;
      lastUpdatedAt = updatedAt;
      revalidateRef.current();
    });
  }, [queryClient, enabled]);
}
