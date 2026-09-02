import { useCallback, useEffect, useMemo, useRef, useState, type RefObject } from "react";

import { Input } from "@/components/ui/input";
import { TableCell } from "@/components/ui/table";
import { C } from "@/lib/design-tokens";
import { calculateDose } from "@/lib/medicine-dose";

import {
  buildDoseCalcInput,
  toDoseParamsAuthority,
  useMedicineDoseParams,
  type MedicineDoseContext,
} from "../../api/medicine-dose-lookup";
import type { Treatment, UpdateTreatmentInput } from "../../types";
import { handleTreatmentEditorKeyDown } from "./treatment-row-editors";
import {
  computeDoseGate,
  resolveDoseGateSource,
  type DoseGateResult,
  type DoseGateSource,
} from "./treatment-row-dose-gate";

interface QuantityCommitParams {
  localQuantity: string;
  localDeviationReason: string;
  lastDeviationCommitKeyRef: { current: string | null };
  setLocalQuantity: (value: string) => void;
  setShowDeviationReason: (value: boolean) => void;
  setLocalDeviationReason: (value: string) => void;
  onStopEdit: () => void;
  treatment: Treatment;
  doseGateSource: DoseGateSource;
  onUpdate: (treatmentId: string, input: UpdateTreatmentInput) => void;
}

function commitTreatmentQuantity(params: QuantityCommitParams) {
  const val = parseFloat(params.localQuantity) || 1;
  const gate = computeDoseGate(params.doseGateSource, val);
  if (gate.isBlocked) {
    params.setLocalQuantity(String(params.treatment.quantity));
    params.setShowDeviationReason(false);
    params.onStopEdit();
    return;
  }

  if (gate.requiresDeviationReason) {
    const reason = params.localDeviationReason.trim();
    if (!reason) {
      params.setShowDeviationReason(true);
      params.onStopEdit();
      return;
    }
    const commitKey = `${val}\0${reason}`;
    if (params.lastDeviationCommitKeyRef.current === commitKey) {
      params.onStopEdit();
      return;
    }
    params.lastDeviationCommitKeyRef.current = commitKey;
    params.setShowDeviationReason(false);
    params.onStopEdit();
    params.onUpdate(params.treatment.id, { quantity: val, dose_deviation_reason: reason });
    return;
  }

  params.setShowDeviationReason(false);
  params.setLocalDeviationReason("");
  params.lastDeviationCommitKeyRef.current = null;
  params.onStopEdit();
  if (val === params.treatment.quantity) return;
  params.onUpdate(params.treatment.id, { quantity: val });
}

function commitTreatmentDeviationReason(params: QuantityCommitParams) {
  const val = parseFloat(params.localQuantity) || 1;
  const gate = computeDoseGate(params.doseGateSource, val);
  if (gate.isBlocked) return;
  if (!gate.requiresDeviationReason) {
    params.setShowDeviationReason(false);
    params.setLocalDeviationReason("");
    params.lastDeviationCommitKeyRef.current = null;
    return;
  }
  const reason = params.localDeviationReason.trim();
  if (!reason) {
    params.setShowDeviationReason(true);
    return;
  }
  const commitKey = `${val}\0${reason}`;
  if (params.lastDeviationCommitKeyRef.current === commitKey) return;
  params.lastDeviationCommitKeyRef.current = commitKey;
  params.setShowDeviationReason(false);
  params.onUpdate(params.treatment.id, { quantity: val, dose_deviation_reason: reason });
}

function quantityDisplayClassName(
  currentWarning: DoseGateResult["warning"],
  pendingWarning: DoseGateResult["warning"],
  needsDeviationReasonUI: boolean,
): string {
  if (currentWarning === "exceeds-max" || pendingWarning === "exceeds-max") {
    return C.textRed700;
  }
  if (
    currentWarning === "below-min" ||
    pendingWarning === "below-min" ||
    needsDeviationReasonUI
  ) {
    return C.textWarning;
  }
  return C.text;
}

function useTreatmentDoseGate(
  treatment: Treatment,
  doseContext: MedicineDoseContext,
  localQuantity: string,
  showDeviationReason: boolean,
) {
  const isMedicineRow = treatment.item_type === "medicine" && !!treatment.medicine_id;
  const medicine = isMedicineRow
    ? doseContext.medicines?.find((m) => m.id === treatment.medicine_id)
    : undefined;
  const doseParamsQuery = useMedicineDoseParams(
    isMedicineRow ? treatment.medicine_id : undefined,
  );
  const doseParamsAuthority = useMemo(
    () =>
      toDoseParamsAuthority(isMedicineRow ? treatment.medicine_id : undefined, {
        data: doseParamsQuery.data,
        isError: doseParamsQuery.isError,
      }),
    [isMedicineRow, treatment.medicine_id, doseParamsQuery.data, doseParamsQuery.isError],
  );
  const doseCalcInput = useMemo(
    () =>
      buildDoseCalcInput(
        medicine,
        doseParamsAuthority.status === "success" ? doseParamsAuthority.params : undefined,
        doseContext.petSpecies,
        doseContext.weightKg,
      ),
    [medicine, doseParamsAuthority, doseContext.petSpecies, doseContext.weightKg],
  );
  const doseGateSource = useMemo(
    () => resolveDoseGateSource(doseCalcInput, doseParamsAuthority),
    [doseCalcInput, doseParamsAuthority],
  );
  const dosePreview = useMemo(
    () => (doseCalcInput ? calculateDose(doseCalcInput) : null),
    [doseCalcInput],
  );
  const currentGate = useMemo(
    () => computeDoseGate(doseGateSource, treatment.quantity),
    [doseGateSource, treatment.quantity],
  );
  const pendingQty = parseFloat(localQuantity) || treatment.quantity;
  const pendingGate = useMemo(
    () => computeDoseGate(doseGateSource, pendingQty),
    [doseGateSource, pendingQty],
  );
  const quantityDirty = pendingQty !== treatment.quantity;
  const needsDeviationReasonUI =
    showDeviationReason || (quantityDirty && pendingGate.requiresDeviationReason);
  const doseBlockReason = pendingGate.blockReason;
  const doseWarningText =
    doseBlockReason ||
    (needsDeviationReasonUI
      ? pendingGate.reason || "用量が推奨域から逸脱しています"
      : currentGate.warning !== "none"
        ? currentGate.warning === "exceeds-max"
          ? "上限超過"
          : "下限未満"
        : "");

  const handleRetryDoseParamsLookup = useCallback(() => {
    void doseParamsQuery.refetch();
  }, [doseParamsQuery]);

  return {
    doseGateSource,
    dosePreview,
    currentGate,
    pendingGate,
    isDoseLookupFailed: doseGateSource.kind === "technical_failure",
    doseWarningId: `dose-warning-${treatment.id}`,
    doseBlockReason,
    needsDeviationReasonUI,
    doseWarningText,
    hasDoseMessage: doseWarningText !== "" || needsDeviationReasonUI,
    handleRetryDoseParamsLookup,
  };
}

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

function TreatmentDoseMessages({
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
          推奨{dosePreview.quantity}（{dosePreview.rawMg.toFixed(1)}→{dosePreview.effectiveDoseMg.toFixed(1)}mg）
        </div>
      ) : null}
      {treatment.dose_amount_mg != null ? (
        <div
          className={`text-xs ${C.text40} text-right mt-0.5`}
          title={
            treatment.dose_weight_source
              ? `体重出典: ${treatment.dose_weight_source}`
              : undefined
          }
        >
          保存時{treatment.dose_amount_mg}{treatment.dose_amount_unit ?? "mg"}
          {treatment.dose_weight_kg != null ? `（体重${treatment.dose_weight_kg}kg）` : ""}
        </div>
      ) : null}
    </>
  );
}

interface TreatmentQuantityCellProps {
  treatment: Treatment;
  doseContext: MedicineDoseContext;
  isEditing: boolean;
  inputRef: RefObject<HTMLInputElement | null>;
  onStartEdit: () => void;
  onStopEdit: () => void;
  onUpdate: (treatmentId: string, input: UpdateTreatmentInput) => void;
}

export function TreatmentQuantityCell({
  treatment,
  doseContext,
  isEditing,
  inputRef,
  onStartEdit,
  onStopEdit,
  onUpdate,
}: TreatmentQuantityCellProps) {
  const [localQuantity, setLocalQuantity] = useState(String(treatment.quantity));
  const [localDeviationReason, setLocalDeviationReason] = useState("");
  const [showDeviationReason, setShowDeviationReason] = useState(false);
  const lastDeviationCommitKeyRef = useRef<string | null>(null);

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect -- 行データ更新を数量下書きへ同期
    setLocalQuantity(String(treatment.quantity));
    setLocalDeviationReason("");
    setShowDeviationReason(false);
    lastDeviationCommitKeyRef.current = null;
  }, [treatment]);

  const dose = useTreatmentDoseGate(
    treatment,
    doseContext,
    localQuantity,
    showDeviationReason,
  );

  const commitParams: QuantityCommitParams = {
    localQuantity,
    localDeviationReason,
    lastDeviationCommitKeyRef,
    setLocalQuantity,
    setShowDeviationReason,
    setLocalDeviationReason,
    onStopEdit,
    treatment,
    doseGateSource: dose.doseGateSource,
    onUpdate,
  };

  const commitQuantity = () => commitTreatmentQuantity(commitParams);
  const commitDeviationReason = () => commitTreatmentDeviationReason(commitParams);

  return (
    <TableCell className="w-20 text-right">
      {isEditing ? (
        <Input
          ref={inputRef}
          type="number"
          step="0.1"
          min="0.1"
          value={localQuantity}
          onChange={(e) => setLocalQuantity(e.target.value)}
          onBlur={commitQuantity}
          onKeyDown={(e) => handleTreatmentEditorKeyDown(e, commitQuantity, onStopEdit)}
          className={`h-8 text-sm text-right px-2 ${C.borderMedium}`}
          aria-label="数量"
          aria-describedby={dose.hasDoseMessage ? dose.doseWarningId : undefined}
          aria-invalid={dose.doseBlockReason !== "" ? true : undefined}
        />
      ) : (
        <button
          type="button"
          className={`w-full text-right text-sm ${C.hoverBgLight} px-1 py-0.5 rounded-xxs transition-colors ${quantityDisplayClassName(
            dose.currentGate.warning,
            dose.pendingGate.warning,
            dose.needsDeviationReasonUI,
          )}`}
          onClick={onStartEdit}
          aria-describedby={dose.hasDoseMessage ? dose.doseWarningId : undefined}
        >
          {showDeviationReason ? localQuantity : treatment.quantity}
        </button>
      )}
      <TreatmentDoseMessages
        treatment={treatment}
        doseWarningId={dose.doseWarningId}
        doseBlockReason={dose.doseBlockReason}
        isDoseLookupFailed={dose.isDoseLookupFailed}
        needsDeviationReasonUI={dose.needsDeviationReasonUI}
        doseWarningText={dose.doseWarningText}
        currentGate={dose.currentGate}
        dosePreview={dose.dosePreview}
        localDeviationReason={localDeviationReason}
        onDeviationReasonChange={setLocalDeviationReason}
        onCommitDeviationReason={commitDeviationReason}
        onRetryDoseParamsLookup={dose.handleRetryDoseParamsLookup}
      />
    </TableCell>
  );
}
