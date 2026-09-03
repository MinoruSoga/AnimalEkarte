import { useEffect, useRef, useState, type RefObject } from "react";

import { Input } from "@/components/ui/input";
import { TableCell } from "@/components/ui/table";
import { C } from "@/lib/design-tokens";

import type { MedicineDoseContext } from "../../api/medicine-dose-lookup";
import type { Treatment, UpdateTreatmentInput } from "../../types";
import { handleTreatmentEditorKeyDown } from "./treatment-row-editors";
import { TreatmentDoseMessages } from "./treatment-dose-messages";
import {
  commitTreatmentDeviationReason,
  commitTreatmentQuantity,
  quantityDisplayClassName,
  type QuantityCommitParams,
} from "./treatment-quantity-commit";
import { useTreatmentDoseGate } from "./use-treatment-dose-gate";

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
