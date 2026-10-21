import { FilePlus2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group";
import { BADGE, C, ICON } from "@/lib/design-tokens";
import type { CertificateApi } from "../api/certificate-api";
import type { CertificateContext, CertificateKind, CertificateLayout } from "../api/certificate-schemas";
import { useCertificateIssue } from "../hooks/use-certificate-issue";
import { isEditable } from "../lib/certificate-issue-reducer";
import { KIND_LABEL, LAYOUT_LABEL, layoutsForKind, selectableKinds } from "../lib/certificate-inputs";
import { VALIDITY_LABEL } from "../lib/certificate-messages";
import { CertificateArtifactPrint } from "./CertificateArtifactPrint";
import { CertificateInputsForm } from "./CertificateInputsForm";
import { CertificateNotice } from "./CertificateNotice";
import { CertificatePdfFrame } from "./CertificatePdfFrame";

interface CertificateIssuePanelProps {
  api: CertificateApi;
  vaccinationId: number;
  context: CertificateContext;
  /** 現在の vaccinations:edit 権限。mutation 直前にも getCanEdit で再検査する。 */
  getCanEdit: () => boolean;
  onRefetchContext: () => void;
}

const ELIGIBILITY_NOTE = {
  ineligible: "このワクチンは狂犬病予防注射済証の対象として登録されていないため、通常の証明書のみ発行できます。",
  unknown: "このワクチンは狂犬病予防注射済証の対象かどうか未確認のため、通常の証明書のみ発行できます。",
} as const;

export function CertificateIssuePanel({
  api,
  vaccinationId,
  context,
  getCanEdit,
  onRefetchContext,
}: CertificateIssuePanelProps) {
  const flow = useCertificateIssue({ api, vaccinationId, context, getCanEdit, onRefetchContext });
  const { state, fieldErrors } = flow;
  const kinds = selectableKinds(context);
  const editable = isEditable(state.phase);
  const busy = state.phase === "previewing" || state.phase === "issuing";
  const latest = context.latest[state.kind];

  if (!context.can_issue) {
    return (
      <p role="status" className={`rounded-xs border p-3 text-sm ${BADGE.gray}`}>
        この接種記録の証明書を発行する権限がありません。履歴の閲覧のみ行えます。
      </p>
    );
  }
  if (kinds.length === 0) {
    return (
      <p role="alert" className={`text-sm ${C.danger}`}>
        発行できる様式がありません。
      </p>
    );
  }

  if (state.phase === "issued" && state.issued) {
    return (
      <div className="space-y-3">
        {state.issued.replayed ? (
          <p role="status" className={`rounded-xs border p-3 text-sm ${BADGE.blue}`}>
            この発行はすでに完了していました（同じ操作の結果を再表示しています）。
          </p>
        ) : null}
        <CertificateArtifactPrint
          api={api}
          editionId={state.issued.editionId}
          artifactId={state.issued.artifactId}
          heading="発行しました"
        />
        <Button type="button" variant="outline" onClick={flow.startNew}>
          新しく発行する
        </Button>
      </div>
    );
  }

  return (
    <section aria-labelledby="cert-issue-heading" aria-busy={busy} className="space-y-4">
      <h2 id="cert-issue-heading" className="text-lg font-semibold">
        証明書を発行する
      </h2>

      <div role="group" aria-labelledby="cert-kind-label" className="space-y-2">
        <p id="cert-kind-label" className={`text-sm font-medium ${C.text}`}>
          様式
        </p>
        <RadioGroup
          value={state.kind}
          onValueChange={(v) => flow.edit({ kind: v as CertificateKind })}
          disabled={!editable || busy}
        >
          {kinds.map((kind) => (
            <div key={kind} className="flex items-center gap-2">
              <RadioGroupItem value={kind} id={`cert-kind-${kind}`} />
              <Label htmlFor={`cert-kind-${kind}`} className="mb-0">
                {KIND_LABEL[kind]}
              </Label>
            </div>
          ))}
        </RadioGroup>
        {context.eligibility.state !== "eligible" ? (
          <p className={`text-sm ${C.text60}`}>{ELIGIBILITY_NOTE[context.eligibility.state]}</p>
        ) : null}
      </div>

      {state.kind === "normal" ? (
        <div role="group" aria-labelledby="cert-layout-label" className="space-y-2">
          <p id="cert-layout-label" className={`text-sm font-medium ${C.text}`}>
            用紙の向き
          </p>
          <RadioGroup
            value={state.layout}
            onValueChange={(v) => flow.edit({ layout: v as CertificateLayout })}
            disabled={!editable || busy}
          >
            {layoutsForKind("normal").map((layout) => (
              <div key={layout} className="flex items-center gap-2">
                <RadioGroupItem value={layout} id={`cert-layout-${layout}`} />
                <Label htmlFor={`cert-layout-${layout}`} className="mb-0">
                  {LAYOUT_LABEL[layout]}
                </Label>
              </div>
            ))}
          </RadioGroup>
          <p className={`text-xs ${C.text60}`}>
            縦・横は同じ発行内容を別の用紙で出力するだけです。向きを変えるとプレビューを作り直します。
          </p>
        </div>
      ) : null}

      {latest ? (
        <div className={`flex flex-wrap items-center gap-2 text-sm ${C.text60}`}>
          <span>
            最新の発行版: 第{latest.edition}版（{VALIDITY_LABEL[latest.validity]}）
          </span>
          {state.kind === "rabies" ? (
            <Button
              type="button"
              variant="outline"
              size="sm"
              disabled={!editable || busy}
              onClick={() => flow.copyFrom(latest)}
            >
              最新の発行内容を入力欄へコピー
            </Button>
          ) : null}
        </div>
      ) : null}

      {state.kind === "rabies" ? (
        <CertificateInputsForm
          draft={state.draft}
          errors={fieldErrors}
          disabled={!editable || busy}
          onChange={(patch) => flow.edit({ draft: patch })}
        />
      ) : null}

      {state.notice ? (
        <CertificateNotice code={state.notice}>
          {state.phase === "unknown" ? (
            <div className="mt-2">
              <Button type="button" variant="outline" disabled={flow.checking} onClick={() => void flow.recheck()}>
                発行結果を照会する
              </Button>
            </div>
          ) : null}
          {state.phase === "not-issued" ? (
            <div className="mt-2 flex flex-wrap gap-2">
              <Button type="button" onClick={() => void flow.retrySameKey()}>
                同じ内容で再試行
              </Button>
              <Button type="button" variant="outline" onClick={flow.discardAndEdit}>
                内容を直してやり直す
              </Button>
            </div>
          ) : null}
        </CertificateNotice>
      ) : null}

      {state.phase === "editing" || state.phase === "stale" || state.phase === "previewing" ? (
        <Button type="button" onClick={() => void flow.requestPreview()} disabled={busy}>
          <FilePlus2 className={ICON.xs} aria-hidden="true" />
          {state.phase === "previewing" ? "プレビューを作成中…" : "プレビューを作成"}
        </Button>
      ) : null}

      {state.preview && (state.phase === "previewed" || state.phase === "issuing" || state.phase === "unknown" || state.phase === "not-issued") ? (
        <div className="space-y-3">
          <p className={`text-sm ${C.text60}`}>
            確認用のプレビューです（印刷はできません）。発行すると版として保存され、保存された PDF から印刷します。
          </p>
          <CertificatePdfFrame base64={state.preview.pdf_base64} title="証明書の確認プレビュー" />
          {state.phase === "previewed" || state.phase === "issuing" ? (
            <Button type="button" onClick={() => void flow.issue()} disabled={busy || state.phase !== "previewed"}>
              {state.phase === "issuing" ? "発行中…" : "この内容で発行する"}
            </Button>
          ) : null}
        </div>
      ) : null}
    </section>
  );
}
