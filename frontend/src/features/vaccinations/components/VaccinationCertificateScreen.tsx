import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useCallback, useLayoutEffect, useRef } from "react";
import { Button } from "@/components/ui/button";
import { usePermission } from "@/hooks/use-permission";
import { C } from "@/lib/design-tokens";
import type { CertificateApi } from "../api/certificate-api";
import { CertificateApiError, type CertificateErrorKind } from "../api/certificate-errors";
import { httpCertificateApi } from "../api/certificate-http-client";
import { CertificateHistory } from "./CertificateHistory";
import { CertificateIssuePanel } from "./CertificateIssuePanel";
import { certificateContextQueryKey, certificateQueryOptions } from "../lib/certificate-query";

interface VaccinationCertificateScreenProps {
  vaccinationId: number;
  /**
   * 証明書 API。既定は実 API（契約は selected・runtime 未実装のため、BE 完成までは 404/5xx になりうる）。
   * 契約モックは ../mock/certificate-mock-api を呼び出し側（テスト・Storybook）が明示的に渡す。
   * 実 API の失敗時にモックへ自動で切り替えることはしない。
   */
  api?: CertificateApi;
}

const CONTEXT_ERROR_TEXT: Partial<Record<CertificateErrorKind, string>> = {
  unauthenticated: "ログインの有効期限が切れました。再ログイン後にやり直してください。",
  forbidden: "この接種記録の証明書を閲覧する権限がありません。",
  "not-found": "接種記録が見つかりません。削除された可能性があります。",
  "dependency-changed": "接種記録が訂正または削除されたため、証明書の情報を取得できません。接種記録を開き直してください。",
  "version-conflict": "接種記録が更新されました。再読み込みしてください。",
  "result-unavailable": "証明書の情報を取得できません。時間をおいて再読み込みしてください。",
  "render-unavailable": "証明書の情報を取得できません。時間をおいて再読み込みしてください。",
  malformed: "証明書の情報の形式が想定と異なります。時間をおいて再読み込みしてください。",
};

/**
 * 予防接種証明書の発行・再印刷画面（EMR-487）。
 * 様式の描画は server renderer の PDF で、この画面は PDF を表示・印刷するだけで再描画しない。
 */
export function VaccinationCertificateScreen({
  vaccinationId,
  api = httpCertificateApi,
}: VaccinationCertificateScreenProps) {
  const queryClient = useQueryClient();
  const { canEdit } = usePermission("vaccinations");
  const canEditRef = useRef(canEdit);
  useLayoutEffect(() => {
    canEditRef.current = canEdit;
  }, [canEdit]);
  const getCanEdit = useCallback(() => canEditRef.current, []);

  const context = useQuery({
    queryKey: certificateContextQueryKey(vaccinationId),
    queryFn: ({ signal }) => api.getContext(vaccinationId, signal),
    ...certificateQueryOptions,
  });

  // 発行成功・前提変更の後に context と履歴を取り直す（API 直接の再取得。キャッシュ値を信用しない）。
  const refetchAll = useCallback(() => {
    void queryClient.invalidateQueries({
      predicate: (q) => q.queryKey[0] === "vaccinations" && String(q.queryKey[1]).startsWith("certificate"),
      refetchType: "active",
    });
  }, [queryClient]);

  // 再開の判断用: context を取り直し、取り直した後の can_issue を返す（取得できなければ false）。
  const refetchCanIssue = useCallback(async () => {
    const result = await context.refetch();
    return result.data?.can_issue ?? false;
  }, [context]);

  if (context.isPending) {
    return (
      <p role="status" aria-busy="true" className={`text-sm ${C.text60}`}>
        証明書の情報を読み込んでいます…
      </p>
    );
  }
  // 背景の再取得が失敗しても、取得済みの data があれば画面（発行結果・未確定の操作）を維持する。
  if (context.isError && context.data === undefined) {
    const kind = context.error instanceof CertificateApiError ? context.error.kind : "network";
    return (
      <div role="alert" className="space-y-2">
        <p className={`text-sm ${C.danger}`}>
          {CONTEXT_ERROR_TEXT[kind] ?? "証明書の情報を取得できませんでした。通信状態を確認してください。"}
        </p>
        {kind !== "forbidden" && kind !== "not-found" && kind !== "unauthenticated" ? (
          <Button type="button" variant="outline" onClick={() => void context.refetch()}>
            再読み込み
          </Button>
        ) : null}
      </div>
    );
  }

  const refetchFailedKind =
    context.isError && context.error instanceof CertificateApiError ? context.error.kind : "network";
  return (
    <div className="space-y-8">
      {context.isError ? (
        <div role="alert" className="space-y-2">
          <p className={`text-sm ${C.danger}`}>
            {CONTEXT_ERROR_TEXT[refetchFailedKind] ??
              "証明書の情報の更新に失敗しました。表示は最新ではない可能性があります。"}
          </p>
          <Button type="button" variant="outline" onClick={() => void context.refetch()}>
            再読み込み
          </Button>
        </div>
      ) : null}
      <CertificateIssuePanel
        // 別の接種記録へ移ったら入力・preview を持ち越さない（同一記録の context 更新では保持する）。
        key={vaccinationId}
        api={api}
        vaccinationId={vaccinationId}
        context={context.data}
        getCanEdit={getCanEdit}
        onRefetchContext={refetchAll}
        refetchCanIssue={refetchCanIssue}
      />
      <CertificateHistory key={vaccinationId} api={api} vaccinationId={vaccinationId} context={context.data} onRefetch={refetchAll} getCanEdit={getCanEdit} />
    </div>
  );
}
