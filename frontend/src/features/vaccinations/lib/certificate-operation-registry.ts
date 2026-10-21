import type { QueryClient } from "@tanstack/react-query";
import type { IssuePending } from "./certificate-issue-reducer";
import { createReprintStore, type ReprintStore } from "./certificate-reprint-store";

/**
 * 未確定の発行・再印刷の operation_key を、画面の unmount を越えて「メモリ上だけ」で保持する登録簿。
 * 画面から離れて戻ったときに key を失うと、先の POST が commit 済みでも別 key で二重発行しうるため。
 *
 * - 永続化しない: localStorage / sessionStorage / URL / IndexedDB には書かない（リロードで消える。
 *   リロードを越える保持が必要かは仕様判断待ち）。
 * - 範囲: 接種記録（vaccinationId）ごと。この feature 専用で、他 feature からは参照しない。
 * - 寿命: QueryClient の query cache に載せる。ログアウト・医院切替の `queryClient.clear()` で一緒に消え、
 *   別ユーザーへ未確定の操作を持ち越さない。unmount 後も gc されないよう gcTime は無限にする。
 * - 保持するのは key と request（入力欄の値を含むが、飼主・ペットの個人情報や PDF は含まない）。
 *   確認用 preview（個人情報を含む snapshot / PDF）は持たない。
 */
const REGISTRY_QUERY_KEY = ["vaccinations", "operation-registry"] as const;

export interface CertificateOperationRegistry {
  /** 発行（POST /certificate-editions）の未確定操作。接種記録ごとに高々 1 つ。 */
  readonly issuePending: Map<number, IssuePending>;
  /** 再印刷の store（edition ごとの key と進行状態）。 */
  readonly reprintStores: Map<number, ReprintStore>;
}

export function certificateOperationRegistry(queryClient: QueryClient): CertificateOperationRegistry {
  const existing = queryClient.getQueryData<CertificateOperationRegistry>(REGISTRY_QUERY_KEY);
  if (existing !== undefined) return existing;
  queryClient.setQueryDefaults(REGISTRY_QUERY_KEY, { gcTime: Infinity, structuralSharing: false });
  const created: CertificateOperationRegistry = { issuePending: new Map(), reprintStores: new Map() };
  queryClient.setQueryData(REGISTRY_QUERY_KEY, created);
  return created;
}

/** 接種記録の再印刷 store。なければ作る（同じ接種記録に戻ると同じ store を返す）。 */
export function reprintStoreFor(registry: CertificateOperationRegistry, vaccinationId: number): ReprintStore {
  const existing = registry.reprintStores.get(vaccinationId);
  if (existing !== undefined) return existing;
  const created = createReprintStore();
  registry.reprintStores.set(vaccinationId, created);
  return created;
}
