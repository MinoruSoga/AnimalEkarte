import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import { C } from "@/lib/design-tokens";
import type { CertificateApi } from "../api/certificate-api";
import { CertificateApiError } from "../api/certificate-errors";
import type { CertificateContext, CertificateKind } from "../api/certificate-schemas";
import { KIND_LABEL } from "../lib/certificate-inputs";
import { CertificateEditionRow } from "./CertificateEditionRow";
import { certificateOperationRegistry, releaseReprintStore, reprintStoreFor, retainReprintStore } from "../lib/certificate-operation-registry";
import { certificateHistoryPageQueryKey, certificateQueryOptions } from "../lib/certificate-query";

interface CertificateHistoryProps {
  api: CertificateApi;
  vaccinationId: number;
  context: CertificateContext;
  onRefetch: () => void;
  getCanEdit: () => boolean;
}

const PAGE_SIZE = 20;

/** 発行履歴（edition 降順）。旧版は参照のみで、印刷導線を持たない。 */
export function CertificateHistory({ api, vaccinationId, context, onRefetch, getCanEdit }: CertificateHistoryProps) {
  // 様式タブ・ページ送り・context 更新、さらに画面の離脱・再 mount を越えて、未確定の再印刷 key を保持する
  // （メモリ上のみ。接種記録ごとの store を QueryClient 単位の登録簿から得る）。
  const queryClient = useQueryClient();
  const [reprintStore] = useState(() => reprintStoreFor(certificateOperationRegistry(queryClient), vaccinationId));
  // 離れるときは、未確定の key を持つ store だけを残す（StrictMode の模擬 unmount でも再登録する）。
  useEffect(() => {
    const registry = certificateOperationRegistry(queryClient);
    retainReprintStore(registry, vaccinationId, reprintStore);
    return () => releaseReprintStore(registry, vaccinationId, reprintStore);
  }, [queryClient, vaccinationId, reprintStore]);
  const kinds: CertificateKind[] = (["normal", "rabies"] as const).filter(
    (k) => context.eligibility.allowed_kinds.includes(k) || context.latest[k] !== null,
  );
  // 様式と頁は一緒に持つ。context の更新で選択中の様式が選べなくなったら、effect を使わず先頭の様式・1 頁目へ導出する
  // （選択肢が 1 つになると切替ボタンも消えるため、残った選択を持ち越すと戻れなくなる）。
  const [selection, setSelection] = useState<{ kind: CertificateKind; page: number }>({
    kind: kinds[0] ?? "normal",
    page: 1,
  });
  const { kind, page } = kinds.includes(selection.kind) ? selection : { kind: kinds[0] ?? "normal", page: 1 };
  const history = useQuery({
    queryKey: certificateHistoryPageQueryKey(vaccinationId, kind, page, context.source_version),
    queryFn: ({ signal }) => api.listEditions(vaccinationId, { kind, page, limit: PAGE_SIZE }, signal),
    ...certificateQueryOptions,
  });

  const totalPages = history.data ? Math.max(1, Math.ceil(history.data.total / history.data.limit)) : 1;
  const status = history.error instanceof CertificateApiError ? history.error.kind : null;

  return (
    <section aria-labelledby="cert-history-heading" className="space-y-3" aria-busy={history.isPending}>
      <h2 id="cert-history-heading" className="text-lg font-semibold">
        発行履歴
      </h2>
      {kinds.length > 1 ? (
        <div role="group" aria-label="様式" className="flex gap-2">
          {kinds.map((k) => (
            <Button
              key={k}
              type="button"
              size="sm"
              variant={k === kind ? "default" : "outline"}
              aria-pressed={k === kind}
              onClick={() => {
                setSelection({ kind: k, page: 1 });
              }}
            >
              {KIND_LABEL[k]}
            </Button>
          ))}
        </div>
      ) : null}

      {history.isPending ? (
        <p role="status" className={`text-sm ${C.text60}`}>
          履歴を読み込んでいます…
        </p>
      ) : null}
      {history.isError ? (
        <div role="alert" className="space-y-2">
          <p className={`text-sm ${C.danger}`}>
            {status === "forbidden"
              ? "履歴を閲覧する権限がありません。"
              : status === "not-found"
                ? "接種記録が見つかりません。"
                : "履歴を取得できませんでした。"}
          </p>
          {status !== "forbidden" && status !== "not-found" && status !== "unauthenticated" ? (
            <Button type="button" size="sm" variant="outline" onClick={() => void history.refetch()}>
              履歴を再読み込み
            </Button>
          ) : null}
        </div>
      ) : null}
      {history.data && history.data.data.length === 0 ? (
        <p className={`text-sm ${C.text60}`}>
          {KIND_LABEL[kind]}の発行履歴はありません。上のフォームから発行すると、ここに版として残ります。
        </p>
      ) : null}
      {history.data && history.data.data.length > 0 ? (
        <ul className="space-y-3">
          {history.data.data.map((edition) => (
            <CertificateEditionRow
              key={edition.id}
              api={api}
              edition={edition}
              sourceVersion={context.source_version}
              onRefetch={onRefetch}
              canReprint={context.can_issue}
              getCanEdit={getCanEdit}
              reprintStore={reprintStore}
            />
          ))}
        </ul>
      ) : null}
      {totalPages > 1 ? (
        <div className="flex items-center gap-2 text-sm">
          <Button type="button" size="sm" variant="outline" disabled={page <= 1} onClick={() => setSelection({ kind, page: page - 1 })}>
            前へ
          </Button>
          <span>
            {page} / {totalPages} ページ
          </span>
          <Button type="button" size="sm" variant="outline" disabled={page >= totalPages} onClick={() => setSelection({ kind, page: page + 1 })}>
            次へ
          </Button>
        </div>
      ) : null}
    </section>
  );
}
