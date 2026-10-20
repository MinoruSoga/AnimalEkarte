import { useQuery } from "@tanstack/react-query";
import { useState } from "react";
import { Button } from "@/components/ui/button";
import { C } from "@/lib/design-tokens";
import type { CertificateApi } from "../api/certificate-api";
import { CertificateApiError } from "../api/certificate-errors";
import type { CertificateContext, CertificateKind } from "../api/certificate-schemas";
import { KIND_LABEL } from "../lib/certificate-inputs";
import { CertificateEditionRow } from "./CertificateEditionRow";
import { createReprintStore } from "../lib/certificate-reprint-store";
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
  // 様式タブ・ページ送り・context 更新で行が unmount されても、未確定の再印刷 key を保持する。
  const [reprintStore] = useState(createReprintStore);
  const kinds: CertificateKind[] = (["normal", "rabies"] as const).filter(
    (k) => context.eligibility.allowed_kinds.includes(k) || context.latest[k] !== null,
  );
  const [kind, setKind] = useState<CertificateKind>(kinds[0] ?? "normal");
  const [page, setPage] = useState(1);
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
                setKind(k);
                setPage(1);
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
        <p role="alert" className={`text-sm ${C.danger}`}>
          {status === "forbidden"
            ? "履歴を閲覧する権限がありません。"
            : status === "not-found"
              ? "接種記録が見つかりません。"
              : "履歴を取得できませんでした。"}
        </p>
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
          <Button type="button" size="sm" variant="outline" disabled={page <= 1} onClick={() => setPage((p) => p - 1)}>
            前へ
          </Button>
          <span>
            {page} / {totalPages} ページ
          </span>
          <Button type="button" size="sm" variant="outline" disabled={page >= totalPages} onClick={() => setPage((p) => p + 1)}>
            次へ
          </Button>
        </div>
      ) : null}
    </section>
  );
}
