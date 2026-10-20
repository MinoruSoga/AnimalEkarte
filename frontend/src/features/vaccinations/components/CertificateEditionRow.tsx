import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { NativeSelect } from "@/components/ui/select";
import { BADGE, C } from "@/lib/design-tokens";
import { formatDate } from "@/lib/format/date";
import { formatCurrency } from "@/lib/format/number";
import type { CertificateApi } from "../api/certificate-api";
import { CertificateApiError } from "../api/certificate-errors";
import type { CertificateEdition, CertificateLayout, CertificateSavedPreview } from "../api/certificate-schemas";
import { useCertificateReprint } from "../hooks/use-certificate-reprint";
import { LAYOUT_LABEL } from "../lib/certificate-inputs";
import { VALIDITY_LABEL } from "../lib/certificate-messages";
import { CertificateArtifactPrint } from "./CertificateArtifactPrint";
import { CertificateNotice } from "./CertificateNotice";
import { CertificatePdfFrame } from "./CertificatePdfFrame";

interface CertificateEditionRowProps {
  api: CertificateApi;
  edition: CertificateEdition;
  sourceVersion: number;
  onRefetch: () => void;
}

/** 履歴 1 行。現在有効な版だけ保存版プレビュー・再印刷を出す。旧版は閲覧のみ（印刷不可）。 */
export function CertificateEditionRow({ api, edition, sourceVersion, onRefetch }: CertificateEditionRowProps) {
  const isCurrent = edition.validity === "current";
  const [layout, setLayout] = useState<CertificateLayout>(edition.available_layouts[0] ?? "portrait");
  const [saved, setSaved] = useState<CertificateSavedPreview | null>(null);
  const [savedError, setSavedError] = useState<string | null>(null);
  const [savedBusy, setSavedBusy] = useState(false);
  const reprint = useCertificateReprint({ api, edition, sourceVersion, onRefetch });
  const inputs = edition.snapshot.inputs;

  async function showSaved() {
    setSavedBusy(true);
    setSavedError(null);
    try {
      setSaved(await api.previewSaved(String(edition.id), layout));
    } catch (error) {
      setSaved(null);
      const kind = error instanceof CertificateApiError ? error.kind : "network";
      setSavedError(
        kind === "dependency-changed"
          ? "この版は現在有効ではありません。履歴を更新しました。"
          : "保存された版のプレビューを取得できませんでした。",
      );
      if (kind === "dependency-changed") onRefetch();
    } finally {
      setSavedBusy(false);
    }
  }

  function changeLayout(next: CertificateLayout) {
    setLayout(next);
    setSaved(null);
    setSavedError(null);
    reprint.reset();
  }

  return (
    <li className={`space-y-3 rounded-xs border p-3 ${C.borderLight}`}>
      <div className="flex flex-wrap items-center gap-x-4 gap-y-1 text-sm">
        <span className="font-semibold">第{edition.edition}版</span>
        <span>発行 {formatDate(edition.issued_at)}</span>
        <span>接種版 {edition.source_version}</span>
        <span className={`rounded-full border px-2 py-0.5 text-xs font-semibold ${isCurrent ? BADGE.green : BADGE.gray}`}>
          {VALIDITY_LABEL[edition.validity]}
        </span>
        {edition.kind === "rabies" ? (
          <span className={C.text60}>
            合計 {formatCurrency(inputs.total)} ／ 登録手数料 {formatCurrency(inputs.registration_fee)} ／ 注射料{" "}
            {formatCurrency(inputs.injection_fee)} ／ 注射済票費用 {formatCurrency(inputs.tag_fee)}
          </span>
        ) : null}
      </div>

      {!isCurrent ? (
        <p className={`text-sm ${C.text60}`}>旧版のため印刷できません。履歴としての参照のみです。</p>
      ) : (
        <div className="space-y-3">
          <div className="flex flex-wrap items-end gap-2">
            {edition.available_layouts.length > 1 ? (
              <div className="flex flex-col">
                <Label htmlFor={`layout-${edition.id}`}>用紙の向き</Label>
                <NativeSelect
                  id={`layout-${edition.id}`}
                  value={layout}
                  onChange={(e) => changeLayout(e.target.value as CertificateLayout)}
                >
                  {edition.available_layouts.map((l) => (
                    <option key={l} value={l}>
                      {LAYOUT_LABEL[l]}
                    </option>
                  ))}
                </NativeSelect>
              </div>
            ) : null}
            <Button type="button" variant="outline" disabled={savedBusy} onClick={() => void showSaved()}>
              保存版を確認
            </Button>
            <Button
              type="button"
              disabled={reprint.state.phase === "running"}
              onClick={() => void reprint.start(layout)}
            >
              再印刷の準備
            </Button>
          </div>

          {savedError ? (
            <p role="alert" className={`text-sm ${C.danger}`}>
              {savedError}
            </p>
          ) : null}
          {saved ? (
            <div className="space-y-1">
              <p className={`text-sm ${C.text60}`}>保存された版の確認用プレビューです（印刷はできません）。</p>
              <CertificatePdfFrame base64={saved.pdf_base64} title={`第${edition.edition}版の保存プレビュー`} />
            </div>
          ) : null}

          {"notice" in reprint.state ? (
            <CertificateNotice code={reprint.state.notice}>
              {reprint.state.phase === "unknown" ? (
                <div className="mt-2">
                  <Button type="button" variant="outline" onClick={() => void reprint.recheck()}>
                    結果を照会する
                  </Button>
                </div>
              ) : null}
              {reprint.state.phase === "not-issued" ? (
                <div className="mt-2">
                  <Button type="button" onClick={() => void reprint.retrySameKey()}>
                    同じ内容で再試行
                  </Button>
                </div>
              ) : null}
            </CertificateNotice>
          ) : null}
          {reprint.state.phase === "done" ? (
            <CertificateArtifactPrint
              api={api}
              editionId={reprint.state.editionId}
              artifactId={reprint.state.artifactId}
              heading="再印刷の準備ができました"
            />
          ) : null}
        </div>
      )}
    </li>
  );
}
