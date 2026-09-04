import { Input } from "@/components/ui/input";
import { C } from "@/lib/design-tokens";
import { calculateDose } from "@/lib/medicine-dose";
import type { Treatment } from "../../types";
import type { DoseGateResult } from "../../lib/treatment-row-dose-gate";

interface TreatmentDoseMessagesProps {
  treatment: Treatment;
  doseWarningId: string;
  doseBlockReason: string;
  isDoseLookupFailed: boolean;
  needsDeviationReasonUI: boolean;
  doseWarningText: string;
  currentGate: DoseGateResult;
  dosePreview: ReturnType<typeof calculateDose> | null;
  localDeviationReason: string;
  onDeviationReasonChange: (value: string) => void;
  onCommitDeviationReason: () => void;
  onRetryDoseParamsLookup: () => void;
}

export function TreatmentDoseMessages({
  treatment,
  doseWarningId,
  doseBlockReason,
  isDoseLookupFailed,
  needsDeviationReasonUI,
  doseWarningText,
  currentGate,
  dosePreview,
  localDeviationReason,
  onDeviationReasonChange,
  onCommitDeviationReason,
  onRetryDoseParamsLookup,
}: TreatmentDoseMessagesProps) {
  return (
    <>
      {doseBlockReason ? (
        <div
          id={doseWarningId}
          role="alert"
          className={`text-xs text-right mt-0.5 ${C.textRed700}`}
        >
          <div>⚠ {doseBlockReason}</div>
          {isDoseLookupFailed ? (
            <button
              type="button"
              className={`mt-0.5 text-xs underline ${C.textRed700}`}
              onClick={onRetryDoseParamsLookup}
              aria-label="投与量パラメータの取得を再試行する"
            >
              再試行する
            </button>
          ) : null}
        </div>
      ) : needsDeviationReasonUI && doseWarningText ? (
        <div
          id={doseWarningId}
          role="alert"
          className={`text-xs text-right mt-0.5 ${C.textWarning}`}
        >
          ⚠ {doseWarningText}
        </div>
      ) : currentGate.warning !== "none" ? (
        <div
          id={doseWarningId}
          role="alert"
          className={`text-xs text-right mt-0.5 ${
            currentGate.warning === "exceeds-max" ? C.textRed700 : C.textWarning
          }`}
        >
          {currentGate.warning === "exceeds-max" ? "⚠ 上限超過" : "⚠ 下限未満"}
        </div>
      ) : null}
      {needsDeviationReasonUI && !doseBlockReason ? (
        <div className="mt-1 text-right">
          <label className="sr-only" htmlFor={`dose-deviation-reason-${treatment.id}`}>
            用量逸脱の理由
          </label>
          <Input
            id={`dose-deviation-reason-${treatment.id}`}
            value={localDeviationReason}
            onChange={(e) => onDeviationReasonChange(e.target.value)}
            onBlur={onCommitDeviationReason}
            onKeyDown={(e) => {
              if (e.key === "Enter") {
                e.preventDefault();
                onCommitDeviationReason();
              }
            }}
            placeholder="逸脱理由（必須）"
            maxLength={500}
            className={`h-8 text-xs px-2 ${C.borderMedium}`}
            aria-label="用量逸脱の理由"
            aria-required={true}
          />
        </div>
      ) : null}
      {dosePreview?.eligible ? (
        <div className={`text-xs ${C.text40} text-right mt-0.5`}>
          推奨{dosePreview.quantity}（{dosePreview.rawMg.toFixed(1)}→
          {dosePreview.effectiveDoseMg.toFixed(1)}mg）
        </div>
      ) : null}
      {treatment.dose_amount_mg != null ? (
        <div
          className={`text-xs ${C.text40} text-right mt-0.5`}
          title={
            treatment.dose_weight_source ? `体重出典: ${treatment.dose_weight_source}` : undefined
          }
        >
          保存時{treatment.dose_amount_mg}
          {treatment.dose_amount_unit ?? "mg"}
          {treatment.dose_weight_kg != null ? `（体重${treatment.dose_weight_kg}kg）` : ""}
        </div>
      ) : null}
    </>
  );
}
