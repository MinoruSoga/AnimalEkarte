/**
 * EMR-524: package 選択 → preview → 明示 apply の契約mock picker。
 * runtime 未提供では既定で無効（TREATMENT_PACKAGE_RUNTIME_AVAILABLE=false、
 * 環境変数で自動有効化しない）。選択だけで実施/結果/在庫/請求/証明書は起動しない。
 */
import { useState } from "react";
import { C } from "@/lib/design-tokens";
import { LoadingFallback } from "@/components/shared/DataStates";
import type { TreatmentPackageCandidate } from "../api/treatment-package-applications";
import {
  useTreatmentPackageApply,
  useGetTreatmentPackageCandidates,
  useTreatmentPackagePreview,
} from "../hooks/use-treatment-package-apply";
import {
  TREATMENT_PACKAGE_RUNTIME_AVAILABLE,
  canSubmitApply,
  type OperationScope,
  type TreatmentPackageEditContext,
} from "../lib/treatment-package-operation-state";
import { TreatmentPackagePreview } from "./TreatmentPackagePreview";

export interface TreatmentPackagePickerProps {
  scope: OperationScope;
  editContext: TreatmentPackageEditContext | null;
  canCreate: boolean;
  onClose: () => void;
}

export function TreatmentPackagePicker({
  scope,
  editContext,
  canCreate,
  onClose,
}: TreatmentPackagePickerProps) {
  const list = useGetTreatmentPackageCandidates(scope.clinicId);
  const previewMut = useTreatmentPackagePreview(scope.recordId, scope.clinicId);
  const applyCtl = useTreatmentPackageApply({ scope, editContext });
  const [previewing, setPreviewing] = useState(false);

  if (!TREATMENT_PACKAGE_RUNTIME_AVAILABLE) {
    return (
      <div className={`p-4 text-sm ${C.text60}`} role="note">
        パッケージ適用の実行系は未提供です（契約mock段階）。実 API 有効化は別工程です。
      </div>
    );
  }

  const preview = previewMut.data ?? null;
  const gate = canSubmitApply({
    runtimeAvailable: TREATMENT_PACKAGE_RUNTIME_AVAILABLE,
    editContext,
    preview,
    canCreate,
  });

  return (
    <div className={`flex flex-col gap-3 p-4 ${C.bgWhite}`} aria-label="パッケージ選択">
      {applyCtl.state.kind === "succeeded" ? (
        <div role="status" className="text-sm">
          パッケージを追加しました（application {applyCtl.state.applicationId}）。
          <button type="button" className="ml-2 underline" onClick={onClose}>
            閉じる
          </button>
        </div>
      ) : null}
      {applyCtl.state.kind === "unknown" ? (
        <div role="alert" className={`text-sm ${C.danger}`}>
          結果を確認できません。同じ操作を照会しています（重複追加はしません）。
          <button type="button" className="ml-2 underline" onClick={() => void applyCtl.resolveUnknown()}>
            照会する
          </button>
        </div>
      ) : null}
      {applyCtl.state.kind === "failed" ? (
        <div role="alert" className={`text-sm ${C.danger}`}>
          適用できませんでした（{applyCtl.state.status}）。内容を確認してください。
          <button type="button" className="ml-2 underline" onClick={applyCtl.reset}>
            戻る
          </button>
        </div>
      ) : null}
      {preview === null ? (
        <>
          {list.isPending ? <LoadingFallback /> : null}
          {list.isError ? <p className={`text-sm ${C.danger}`}>候補を取得できませんでした。</p> : null}
          <ul className="flex flex-col gap-1">
            {(list.data ?? [])
              .filter((p) => p.is_active)
              .map((p: TreatmentPackageCandidate) => (
                <li key={p.id}>
                  <button
                    type="button"
                    className="w-full text-left text-sm underline"
                    disabled={previewing}
                    onClick={() => {
                      setPreviewing(true);
                      previewMut.mutate(
                        { packageId: p.id, packageVersion: p.version },
                        { onSettled: () => setPreviewing(false) },
                      );
                    }}
                  >
                    {p.name}（{p.components.length}件）
                  </button>
                </li>
              ))}
          </ul>
        </>
      ) : (
        <TreatmentPackagePreview
          preview={preview}
          canApply={gate.ok}
          applyDisabledReason={gate.ok ? undefined : gate.reason}
          applying={applyCtl.state.kind === "submitting"}
          onApply={() => void applyCtl.submit(preview)}
          onClose={() => previewMut.reset()}
        />
      )}
    </div>
  );
}
