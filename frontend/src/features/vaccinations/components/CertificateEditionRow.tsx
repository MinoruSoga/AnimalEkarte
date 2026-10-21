import { useRef, useState } from "react";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { BADGE, C } from "@/lib/design-tokens";
import { formatDate } from "@/lib/format/date";
import { formatCurrency } from "@/lib/format/number";
import type { CertificateApi } from "../api/certificate-api";
import { CertificateApiError } from "../api/certificate-errors";
import type { CertificateEdition, CertificateLayout, CertificateSavedPreview } from "../api/certificate-schemas";
import { useCertificateRecovery } from "../hooks/use-certificate-recovery";
import { isReprintUnresolved, useCertificateReprint } from "../hooks/use-certificate-reprint";
import type { ReprintStore } from "../lib/certificate-reprint-store";
import { LAYOUT_LABEL } from "../lib/certificate-inputs";
import { VALIDITY_LABEL } from "../lib/certificate-messages";
import { CertificateArtifactPrint } from "./CertificateArtifactPrint";
import { CertificateNotice } from "./CertificateNotice";
import { CertificatePdfFrame } from "./CertificatePdfFrame";
import { CertificateRecoveryPanel } from "./CertificateRecoveryPanel";

interface CertificateEditionRowProps {
  api: CertificateApi;
  edition: CertificateEdition;
  sourceVersion: number;
  onRefetch: () => void;
  /** 再印刷できる権限（context.can_issue）。false なら再印刷の導線を無効にする。 */
  canReprint: boolean;
  /** 再印刷 POST の直前に最新の edit 権限を返す。 */
  getCanEdit: () => boolean;
  /** 未確定の再印刷 key を行の外で保持する store（行が unmount されても失わない）。 */
  reprintStore: ReprintStore;
}

/** 履歴 1 行。現在有効な版だけ保存版プレビュー・再印刷を出す。旧版は閲覧のみ（印刷不可）。 */
export function CertificateEditionRow({
  api,
  edition,
  sourceVersion,
  onRefetch,
  canReprint,
  getCanEdit,
  reprintStore,
}: CertificateEditionRowProps) {
  const isCurrent = edition.validity === "current";
  const [layout, setLayout] = useState<CertificateLayout>(edition.available_layouts[0] ?? "portrait");
  const [saved, setSaved] = useState<CertificateSavedPreview | null>(null);
  const [savedError, setSavedError] = useState<string | null>(null);
  const [savedBusy, setSavedBusy] = useState(false);
  // 再 mount / reload 後に、この版の未確定の再印刷を発見する（現在有効な版・再印刷できる権限のときだけ）。
  const recovery = useCertificateRecovery({
    api,
    target: { intent: "reprint", vaccinationId: edition.vaccination_id, editionId: edition.id },
    enabled: isCurrent && canReprint,
  });
  const reprint = useCertificateReprint({
    api,
    edition,
    sourceVersion,
    onRefetch,
    getCanEdit,
    store: reprintStore,
    ensureRecoveryClear: recovery.ensureClear,
    onBoundSuccess: recovery.adoptSuccess,
  });
  const shownLayout = reprint.activeLayout ?? layout;
  // 再印刷の結果が未確定の間は、同じ key の照会・再試行だけを許す（別操作・向き変更を止める）。
  const reprintLocked = isReprintUnresolved(reprint.state) || reprint.hasPending;
  // 保存版プレビューの応答が、向き変更後の画面に載らないようにする。
  const savedRequestRef = useRef(0);
  const inputs = edition.snapshot.inputs;

  async function showSaved() {
    savedRequestRef.current += 1;
    const requestId = savedRequestRef.current;
    setSavedBusy(true);
    setSavedError(null);
    try {
      const result = await api.previewSaved(String(edition.id), shownLayout);
      if (requestId === savedRequestRef.current) setSaved(result);
    } catch (error) {
      if (requestId !== savedRequestRef.current) return;
      setSaved(null);
      const kind = error instanceof CertificateApiError ? error.kind : "network";
      setSavedError(
        kind === "dependency-changed"
          ? "この版は現在有効ではありません。履歴を更新しました。"
          : "保存された版のプレビューを取得できませんでした。",
      );
      if (kind === "dependency-changed") onRefetch();
    } finally {
      if (requestId === savedRequestRef.current) setSavedBusy(false);
    }
  }

  function changeLayout(next: CertificateLayout) {
    if (reprintLocked) return;
    savedRequestRef.current += 1;
    setSavedBusy(false);
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
                <Select
                  value={shownLayout}
                  disabled={reprintLocked}
                  onValueChange={(v) => changeLayout(v as CertificateLayout)}
                >
                  <SelectTrigger id={`layout-${edition.id}`} className="min-w-40">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {edition.available_layouts.map((l) => (
                      <SelectItem key={l} value={l}>
                        {LAYOUT_LABEL[l]}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            ) : null}
            <Button type="button" variant="outline" disabled={savedBusy} onClick={() => void showSaved()}>
              保存版を確認
            </Button>
            <Button
              type="button"
              disabled={reprintLocked || !canReprint || !recovery.canStart}
              onClick={() => void reprint.start(shownLayout)}
            >
              再印刷の準備
            </Button>
          </div>

          <CertificateRecoveryPanel
            api={api}
            intent="reprint"
            recovery={recovery}
            ownPendingKey={reprintStore.get(edition.id).pending?.request.operation_key ?? null}
          />

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
                  <Button
                    type="button"
                    variant="outline"
                    disabled={reprint.checking}
                    onClick={() => void reprint.recheck()}
                  >
                    結果を照会する
                  </Button>
                </div>
              ) : null}
              {reprint.state.phase === "blocked" && reprint.hasPending ? (
                <div className="mt-2">
                  <Button type="button" variant="outline" onClick={() => void reprint.resumeAfterBlocked()}>
                    権限を再確認して照会する
                  </Button>
                </div>
              ) : null}
              {reprint.state.phase === "not-issued" ? (
                <div className="mt-2">
                  <div className="flex flex-wrap gap-2">
                    <Button type="button" onClick={() => void reprint.retrySameKey()}>
                      同じ内容で再試行
                    </Button>
                    <Button type="button" variant="outline" onClick={() => reprint.discard()}>
                      この再印刷をやめる
                    </Button>
                  </div>
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
