import { useCallback, useMemo } from "react";

import { calculateDose } from "@/lib/medicine-dose";
import {
  buildDoseCalcInput,
  toDoseParamsAuthority,
  useMedicineDoseParams,
  type MedicineDoseContext,
} from "../api/medicine-dose-lookup";
import type { Treatment } from "../types";
import { computeDoseGate, resolveDoseGateSource } from "../components/TreatmentsTab/treatment-row-dose-gate";

export function useTreatmentDoseGate(
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
