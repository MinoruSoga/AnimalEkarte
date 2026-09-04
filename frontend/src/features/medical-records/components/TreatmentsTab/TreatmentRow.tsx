import { memo, useState, useCallback, useRef, useEffect, useLayoutEffect } from "react";

import { C } from "@/lib/design-tokens";

import type { MedicineDoseContext } from "../../api/medicine-dose-lookup";
import type { Treatment, UpdateTreatmentInput } from "../../types";
import {
  TreatmentInsuranceCell,
  TreatmentRowActions,
  TreatmentSelectionCell,
  TreatmentSubtotalCell,
  TreatmentTypeCell,
} from "./TreatmentRowParts";
import {
  TreatmentContentCell,
  TreatmentDiscountCell,
  TreatmentMemoCell,
  TreatmentUnitPriceCell,
} from "./TreatmentRowEditors";
import { TreatmentQuantityCell } from "./TreatmentQuantityCell";

interface TreatmentRowProps {
  treatment: Treatment;
  isFirst: boolean;
  isLast: boolean;
  onUpdate: (treatmentId: string, input: UpdateTreatmentInput) => void;
  onDelete: (treatmentId: string) => void;
  onMoveUp: (treatmentId: string) => void;
  onMoveDown: (treatmentId: string) => void;
  isUpdating: boolean;
  canDelete?: boolean;
  /** BUG-372: 割引権限（値引額編集可否） */
  canEditDiscount?: boolean;
  /** 新規追加直後に数量フィールドへ自動フォーカスする */
  autoFocusQuantity?: boolean;
  /** autoFocusQuantity 完了後に親へ通知するコールバック */
  onAutoFocusDone?: () => void;
  /** #201: 投与量自動計算プレビュー・保存ゲート判定に使うコンテキスト */
  doseContext: MedicineDoseContext;
}

export const TreatmentRow = memo(function TreatmentRow({
  treatment,
  isFirst,
  isLast,
  onUpdate,
  onDelete,
  onMoveUp,
  onMoveDown,
  isUpdating,
  canDelete = true,
  canEditDiscount = true,
  autoFocusQuantity = false,
  onAutoFocusDone,
  doseContext,
}: TreatmentRowProps) {
  const [editField, setEditField] = useState<string | null>(null);
  const inputRef = useRef<HTMLInputElement>(null);
  const subtotal = treatment.unit_price * treatment.quantity - treatment.discount_amount;

  useEffect(() => {
    if (editField && inputRef.current) {
      inputRef.current.select();
    }
  }, [editField]);

  useLayoutEffect(() => {
    if (autoFocusQuantity) {
      setEditField("quantity");
      onAutoFocusDone?.();
    }
    // autoFocusQuantity は初回マウント時のみ使用するので deps に入れない
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const handleSelectedChange = useCallback(
    (checked: boolean | "indeterminate") => {
      onUpdate(treatment.id, { is_selected: checked === true });
    },
    [treatment.id, onUpdate],
  );

  const handleInsuranceChange = useCallback(
    (checked: boolean | "indeterminate") => {
      onUpdate(treatment.id, { is_insurance: checked === true });
    },
    [treatment.id, onUpdate],
  );

  const startEdit = useCallback((field: string) => setEditField(field), []);
  const stopEdit = useCallback(() => setEditField(null), []);
  const handleDelete = useCallback(() => onDelete(treatment.id), [treatment.id, onDelete]);
  const handleMoveUp = useCallback(() => onMoveUp(treatment.id), [treatment.id, onMoveUp]);
  const handleMoveDown = useCallback(() => onMoveDown(treatment.id), [treatment.id, onMoveDown]);

  const editor = {
    treatment,
    inputRef,
    onStopEdit: stopEdit,
    onUpdate,
  };

  return (
    <tr
      className={`border-b ${C.borderLight} ${C.hoverBgPageHalf} transition-colors ${
        !treatment.is_selected ? "opacity-50" : ""
      } ${isUpdating ? "pointer-events-none" : ""}`}
    >
      <TreatmentSelectionCell checked={treatment.is_selected} onChange={handleSelectedChange} />
      <TreatmentTypeCell itemType={treatment.item_type} />
      <TreatmentContentCell
        {...editor}
        isEditing={editField === "content"}
        onStartEdit={() => startEdit("content")}
      />
      <TreatmentInsuranceCell checked={treatment.is_insurance} onChange={handleInsuranceChange} />
      <TreatmentUnitPriceCell
        {...editor}
        isEditing={editField === "unit_price"}
        onStartEdit={() => startEdit("unit_price")}
      />
      <TreatmentQuantityCell
        treatment={treatment}
        doseContext={doseContext}
        isEditing={editField === "quantity"}
        inputRef={inputRef}
        onStartEdit={() => startEdit("quantity")}
        onStopEdit={stopEdit}
        onUpdate={onUpdate}
      />
      <TreatmentDiscountCell
        {...editor}
        isEditing={editField === "discount_amount"}
        onStartEdit={() => startEdit("discount_amount")}
        canEditDiscount={canEditDiscount}
      />
      <TreatmentSubtotalCell subtotal={subtotal} />
      <TreatmentMemoCell
        {...editor}
        isEditing={editField === "memo"}
        onStartEdit={() => startEdit("memo")}
      />
      <TreatmentRowActions
        isFirst={isFirst}
        isLast={isLast}
        canDelete={canDelete}
        onMoveUp={handleMoveUp}
        onMoveDown={handleMoveDown}
        onDelete={handleDelete}
      />
    </tr>
  );
});
