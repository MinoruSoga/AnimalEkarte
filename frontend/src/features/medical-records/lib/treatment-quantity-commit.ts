import { C } from "@/lib/design-tokens";
import type { Treatment, UpdateTreatmentInput } from "../types";
import {
  computeDoseGate,
  type DoseGateResult,
  type DoseGateSource,
} from "./treatment-row-dose-gate";

export interface QuantityCommitParams {
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

export function commitTreatmentQuantity(params: QuantityCommitParams) {
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

export function commitTreatmentDeviationReason(params: QuantityCommitParams) {
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

export function quantityDisplayClassName(
  currentWarning: DoseGateResult["warning"],
  pendingWarning: DoseGateResult["warning"],
  needsDeviationReasonUI: boolean,
): string {
  if (currentWarning === "exceeds-max" || pendingWarning === "exceeds-max") {
    return C.textRed700;
  }
  if (currentWarning === "below-min" || pendingWarning === "below-min" || needsDeviationReasonUI) {
    return C.textWarning;
  }
  return C.text;
}
