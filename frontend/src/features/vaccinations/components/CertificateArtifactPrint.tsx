import { useQuery } from "@tanstack/react-query";
import { Printer } from "lucide-react";
import { useId, useState } from "react";
import { Button } from "@/components/ui/button";
import { C } from "@/lib/design-tokens";
import type { CertificateApi } from "../api/certificate-api";
import type { CertificateErrorKind } from "../api/certificate-errors";
import { useCertificateArtifactContent } from "../hooks/use-certificate-artifact";
import { KIND_LABEL } from "../lib/certificate-inputs";
import { VALIDITY_LABEL } from "../lib/certificate-messages";
import { printPdfBlob } from "../lib/pdf";
import { CertificatePdfFrame } from "./CertificatePdfFrame";
import { certificateEditionQueryKey, certificateQueryOptions } from "../lib/certificate-query";

interface CertificateArtifactPrintProps {
  api: CertificateApi;
  editionId: string;
  artifactId: string;
  /** 例: 「発行しました」「再印刷の準備ができました」 */
  heading: string;
}

const CONTENT_ERROR_TEXT: Partial<Record<CertificateErrorKind, string>> = {
  "dependency-changed": "この版は現在有効ではありません（旧版）。印刷できません。履歴を更新してください。",
  forbidden: "この PDF を取得する権限がありません。",
  unauthenticated: "ログインの有効期限が切れました。再ログイン後にやり直してください。",
  "not-found": "PDF が見つかりません。",
  "render-unavailable": "PDF を取得できません（保存された正本を確認できません）。時間をおいて再取得してください。",
  "result-unavailable": "PDF を取得できません。時間をおいて再取得してください。",
};

/**
 * 成功 → edition 読取 → PDF 取得 → 印刷。印刷に使うのは /content の PDF だけ。
 * 旧版（current 以外）は PDF を取得せず、印刷ボタンも出さない。
 * 再印刷 receipt は「出力準備済み」を表し、紙の印刷成功を保証しない。
 */
export function CertificateArtifactPrint({ api, editionId, artifactId, heading }: CertificateArtifactPrintProps) {
  // 同じ artifact を指す結果を複数同時に開いても id が重ならないよう、インスタンスごとに作る。
  const headingId = `print-${useId()}`;
  const edition = useQuery({
    queryKey: certificateEditionQueryKey(editionId),
    queryFn: ({ signal }) => api.getEdition(editionId, signal),
    ...certificateQueryOptions,
  });
  const isCurrent = edition.data?.validity === "current";
  const content = useCertificateArtifactContent(api, artifactId, isCurrent);
  const [printError, setPrintError] = useState<string | null>(null);
  const [printing, setPrinting] = useState(false);
  const readyBlob = content.state.status === "ready" ? content.state.blob : null;

  async function handlePrint(blob: Blob) {
    if (printing) return;
    setPrintError(null);
    setPrinting(true);
    try {
      await printPdfBlob(blob);
    } catch (error) {
      const timedOut = error instanceof Error && error.message.includes("タイムアウト");
      setPrintError(
        timedOut
          ? "PDF の読み込みがタイムアウトしました。ブラウザの設定（埋め込み表示の制限）を確認し、もう一度お試しください。"
          : "印刷を開始できませんでした。ブラウザの設定を確認してください。",
      );
    } finally {
      setPrinting(false);
    }
  }

  return (
    <section aria-labelledby={headingId} className="space-y-3" aria-busy={edition.isPending || content.state.status === "loading"}>
      <h3 id={headingId} className="text-lg font-semibold">
        {heading}
      </h3>
      {edition.isPending ? (
        <p role="status" className={`text-sm ${C.text60}`}>
          発行内容を確認しています…
        </p>
      ) : null}
      {edition.isError ? (
        <p role="alert" className={`text-sm ${C.danger}`}>
          発行内容を読み取れませんでした。履歴から開き直してください。
        </p>
      ) : null}
      {edition.data ? (
        <p className="text-sm">
          {KIND_LABEL[edition.data.kind]} 第{edition.data.edition}版 ／ {VALIDITY_LABEL[edition.data.validity]}
        </p>
      ) : null}
      {edition.data && !isCurrent ? (
        <p role="alert" className={`text-sm ${C.danger}`}>
          旧版のため印刷できません。最新版を履歴から確認してください。
        </p>
      ) : null}
      {content.state.status === "loading" ? (
        <p role="status" className={`text-sm ${C.text60}`}>
          印刷用 PDF を取得しています…
        </p>
      ) : null}
      {content.state.status === "error" ? (
        <div role="alert" className="space-y-2">
          <p className={`text-sm ${C.danger}`}>
            {CONTENT_ERROR_TEXT[content.state.kind] ?? "PDF を取得できませんでした。"}
          </p>
          {content.state.kind !== "dependency-changed" ? (
            <Button type="button" variant="outline" onClick={content.retry}>
              PDF を再取得
            </Button>
          ) : null}
        </div>
      ) : null}
      {readyBlob !== null ? (
        <div className="space-y-3">
          <CertificatePdfFrame blob={readyBlob} title="印刷する証明書（保存された発行版）" />
          <Button type="button" disabled={printing} onClick={() => void handlePrint(readyBlob)}>
            <Printer aria-hidden="true" />
            {printing ? "印刷を準備中…" : "印刷する"}
          </Button>
          {printError ? (
            <p role="alert" className={`text-sm ${C.danger}`}>
              {printError}
            </p>
          ) : null}
          <p className={`text-xs ${C.text60}`}>
            表示中の PDF は保存された発行版そのものです。印刷ダイアログを開く操作であり、紙の印刷成功を記録するものではありません。
          </p>
        </div>
      ) : null}
    </section>
  );
}
