/**
 * 証明書の読取 query 共通設定。snapshot には飼主・ペットの個人情報が入るため、
 * 画面を離れたらキャッシュを破棄し再利用しない（staleTime 0 / gcTime 0）。PDF Blob は載せない。
 */
export const certificateQueryOptions = {
  staleTime: 0,
  gcTime: 0,
  refetchOnMount: "always" as const,
  // 自動再試行しない: 401/403/404/409 は結果が変わらず、503 は職員の「再読み込み」で明示的に取り直す。
  // 一時的な通信断の再試行は共有 axios（GET のみ）が担う。
  retry: false,
  meta: { silentError: true },
};

export const certificateContextQueryKey = (vaccinationId: number) =>
  ["vaccinations", "certificate-context", vaccinationId] as const;

export const certificateEditionQueryKey = (editionId: string) =>
  ["vaccinations", "certificate-edition", editionId] as const;

/** 履歴 1 ページ分の key。source_version を含め、接種訂正後の旧ページを再利用しない。 */
export const certificateHistoryPageQueryKey = (
  vaccinationId: number,
  kind: string,
  page: number,
  sourceVersion: number,
) => ["vaccinations", "certificate-history", vaccinationId, kind, page, sourceVersion] as const;
