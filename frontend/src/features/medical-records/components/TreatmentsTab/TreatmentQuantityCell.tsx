import { useEffect, useRef, useState, type KeyboardEvent, type RefObject } from "react";

import { Input } from "@/components/ui/input";
import { TableCell } from "@/components/ui/table";
import { C } from "@/lib/design-tokens";

import type { MedicineDoseContext } from "../../api/medicine-dose-lookup";
import type { Treatment, UpdateTreatmentInput } from "../../types";
import { TreatmentDoseMessages } from "./TreatmentDoseMessages";
import {
  commitTreatmentDeviationReason,
  commitTreatmentQuantity,
  quantityDisplayClassName,
  reduceQuantityEnterKey,
  type QuantityCommitParams,
  type QuantityEnterPhase,
} from "../../lib/treatment-quantity-commit";
import { useTreatmentDoseGate } from "../../hooks/use-treatment-dose-gate";

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
  const enterPhaseRef = useRef<QuantityEnterPhase>("idle");
  const lastDeviationCommitKeyRef = useRef<string | null>(null);

  const resetEnterPhase = () => {
    enterPhaseRef.current = "idle";
  };

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect -- 行データ更新を数量下書きへ同期
    setLocalQuantity(String(treatment.quantity));
    setLocalDeviationReason("");
    setShowDeviationReason(false);
    resetEnterPhase();
    lastDeviationCommitKeyRef.current = null;
  }, [treatment]);

  useEffect(() => {
    if (!isEditing) {
      resetEnterPhase();
    }
  }, [isEditing]);

  const dose = useTreatmentDoseGate(treatment, doseContext, localQuantity, showDeviationReason);

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

  const commitQuantity = () => {
    resetEnterPhase();
    commitTreatmentQuantity(commitParams);
  };
  const commitDeviationReason = () => commitTreatmentDeviationReason(commitParams);

  const handleQuantityKeyDown = (event: KeyboardEvent<HTMLInputElement>) => {
    const isComposing = event.nativeEvent.isComposing || event.nativeEvent.keyCode === 229;
    if (event.key === "Enter" && (event.repeat || isComposing)) {
      return;
    }
    const next = reduceQuantityEnterKey(enterPhaseRef.current, event.key);
    if (event.key === "Enter" || event.key === "Escape") {
      event.preventDefault();
    }
    if (next.action === "none" && next.phase === enterPhaseRef.current) {
      return;
    }
    enterPhaseRef.current = next.phase;
    if (next.action === "commit") {
      commitTreatmentQuantity(commitParams);
      return;
    }
    if (next.action === "cancel") {
      onStopEdit();
    }
  };

  return (
    <TableCell className="w-20 text-right">
      {isEditing ? (
        <Input
          ref={inputRef}
          type="number"
          step="0.1"
          min="0.1"
          value={localQuantity}
          onChange={(e) => {
            resetEnterPhase();
            setLocalQuantity(e.target.value);
          }}
          onBlur={commitQuantity}
          onKeyDown={handleQuantityKeyDown}
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
