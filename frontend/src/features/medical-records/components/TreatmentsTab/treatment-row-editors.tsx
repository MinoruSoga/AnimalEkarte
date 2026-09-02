import { useCallback, useEffect, useState, type KeyboardEvent, type RefObject } from "react";

import { FormFieldError } from "@/components/shared/FormFieldError/FormFieldError";
import { Input } from "@/components/ui/input";
import { TableCell } from "@/components/ui/table";
import { C } from "@/lib/design-tokens";
import { formatCurrency } from "@/lib/format/number";

import type { Treatment, UpdateTreatmentInput } from "../../types";

export function handleTreatmentEditorKeyDown(
  event: KeyboardEvent,
  commit: () => void,
  onStopEdit: () => void,
) {
  if (event.key === "Enter") commit();
  if (event.key === "Escape") onStopEdit();
}

interface TreatmentEditorCellProps {
  treatment: Treatment;
  isEditing: boolean;
  inputRef: RefObject<HTMLInputElement | null>;
  onStartEdit: () => void;
  onStopEdit: () => void;
  onUpdate: (treatmentId: string, input: UpdateTreatmentInput) => void;
}

export function TreatmentContentCell({
  treatment,
  isEditing,
  inputRef,
  onStartEdit,
  onStopEdit,
  onUpdate,
}: TreatmentEditorCellProps) {
  const [localContent, setLocalContent] = useState(treatment.content);

  useEffect(() => {
    setLocalContent(treatment.content);
  }, [treatment]);

  const commitContent = useCallback(() => {
    const trimmed = localContent.trim();
    if (trimmed !== treatment.content) {
      onUpdate(treatment.id, { content: trimmed });
    }
    onStopEdit();
  }, [localContent, treatment.content, treatment.id, onUpdate, onStopEdit]);

  return (
    <TableCell className="min-w-[160px]">
      {isEditing ? (
        <Input
          ref={inputRef}
          value={localContent}
          onChange={(e) => setLocalContent(e.target.value)}
          onBlur={commitContent}
          onKeyDown={(e) => handleTreatmentEditorKeyDown(e, commitContent, onStopEdit)}
          className={`h-8 text-sm px-2 ${C.borderMedium}`}
        />
      ) : (
        <button
          type="button"
          className={`w-full text-left text-sm ${C.text} ${C.hoverBgLight} px-1 py-0.5 rounded-xxs transition-colors`}
          onClick={onStartEdit}
        >
          {treatment.content || <span className={C.text40}>内容を入力</span>}
        </button>
      )}
    </TableCell>
  );
}

export function TreatmentUnitPriceCell({
  treatment,
  isEditing,
  inputRef,
  onStartEdit,
  onStopEdit,
  onUpdate,
}: TreatmentEditorCellProps) {
  const [localUnitPrice, setLocalUnitPrice] = useState(String(treatment.unit_price));
  const [unitPriceError, setUnitPriceError] = useState<string>("");

  useEffect(() => {
    setLocalUnitPrice(String(treatment.unit_price));
  }, [treatment]);

  const commitUnitPrice = useCallback(() => {
    const val = parseFloat(localUnitPrice) || 0;
    if (val < 0) {
      setUnitPriceError("金額は0以上を入力してください");
      return;
    }
    setUnitPriceError("");
    if (val !== treatment.unit_price) {
      onUpdate(treatment.id, { unit_price: val });
    }
    onStopEdit();
  }, [localUnitPrice, treatment.unit_price, treatment.id, onUpdate, onStopEdit]);

  return (
    <TableCell className="w-28 text-right">
      {isEditing ? (
        <>
          <Input
            ref={inputRef}
            type="number"
            min={0}
            value={localUnitPrice}
            onChange={(e) => setLocalUnitPrice(e.target.value)}
            onBlur={commitUnitPrice}
            onKeyDown={(e) => handleTreatmentEditorKeyDown(e, commitUnitPrice, onStopEdit)}
            className={`h-8 text-sm text-right px-2 ${C.borderMedium}`}
          />
          <FormFieldError message={unitPriceError} />
        </>
      ) : (
        <button
          type="button"
          className={`w-full text-right text-sm ${C.text} ${C.hoverBgLight} px-1 py-0.5 rounded-xxs transition-colors font-mono`}
          onClick={onStartEdit}
        >
          {formatCurrency(treatment.unit_price)}
        </button>
      )}
    </TableCell>
  );
}

interface TreatmentDiscountCellProps extends TreatmentEditorCellProps {
  canEditDiscount: boolean;
}

export function TreatmentDiscountCell({
  treatment,
  isEditing,
  inputRef,
  onStartEdit,
  onStopEdit,
  onUpdate,
  canEditDiscount,
}: TreatmentDiscountCellProps) {
  const [localDiscountAmount, setLocalDiscountAmount] = useState(
    String(treatment.discount_amount),
  );
  const [discountAmountError, setDiscountAmountError] = useState<string>("");

  useEffect(() => {
    setLocalDiscountAmount(String(treatment.discount_amount));
  }, [treatment]);

  const commitDiscountAmount = useCallback(() => {
    const val = parseFloat(localDiscountAmount) || 0;
    if (val < 0) {
      setDiscountAmountError("金額は0以上を入力してください");
      return;
    }
    setDiscountAmountError("");
    if (val !== treatment.discount_amount) {
      onUpdate(treatment.id, { discount_amount: val });
    }
    onStopEdit();
  }, [localDiscountAmount, treatment.discount_amount, treatment.id, onUpdate, onStopEdit]);

  return (
    <TableCell className="w-28 text-right">
      {isEditing ? (
        <>
          <Input
            ref={inputRef}
            type="number"
            min={0}
            value={localDiscountAmount}
            onChange={(e) => setLocalDiscountAmount(e.target.value)}
            onBlur={commitDiscountAmount}
            onKeyDown={(e) => handleTreatmentEditorKeyDown(e, commitDiscountAmount, onStopEdit)}
            className={`h-8 text-sm text-right px-2 ${C.borderMedium}`}
          />
          <FormFieldError message={discountAmountError} />
        </>
      ) : (
        <button
          type="button"
          className={`w-full text-right text-sm ${
            treatment.discount_amount > 0 ? C.textDiscount : C.text40
          } ${canEditDiscount ? C.hoverBgLight : ""} px-1 py-0.5 rounded-xxs transition-colors font-mono ${!canEditDiscount ? "cursor-not-allowed opacity-60" : ""}`}
          onClick={() => {
            if (canEditDiscount) onStartEdit();
          }}
          disabled={!canEditDiscount}
          title={!canEditDiscount ? "値引の変更には権限が必要です" : undefined}
        >
          {treatment.discount_amount > 0
            ? `-${formatCurrency(treatment.discount_amount)}`
            : "—"}
        </button>
      )}
    </TableCell>
  );
}

export function TreatmentMemoCell({
  treatment,
  isEditing,
  inputRef,
  onStartEdit,
  onStopEdit,
  onUpdate,
}: TreatmentEditorCellProps) {
  const [localMemo, setLocalMemo] = useState(treatment.memo);

  useEffect(() => {
    setLocalMemo(treatment.memo);
  }, [treatment]);

  const commitMemo = useCallback(() => {
    const trimmed = localMemo.trim();
    if (trimmed !== treatment.memo) {
      onUpdate(treatment.id, { memo: trimmed });
    }
    onStopEdit();
  }, [localMemo, treatment.memo, treatment.id, onUpdate, onStopEdit]);

  return (
    <TableCell className="min-w-[120px]">
      {isEditing ? (
        <Input
          ref={inputRef}
          value={localMemo}
          onChange={(e) => setLocalMemo(e.target.value)}
          onBlur={commitMemo}
          onKeyDown={(e) => handleTreatmentEditorKeyDown(e, commitMemo, onStopEdit)}
          className={`h-8 text-sm px-2 ${C.borderMedium}`}
        />
      ) : (
        <button
          type="button"
          className={`w-full text-left text-sm ${C.text60} ${C.hoverBgLight} px-1 py-0.5 rounded-xxs transition-colors`}
          onClick={onStartEdit}
        >
          {treatment.memo || <span className={C.text30}>メモ</span>}
        </button>
      )}
    </TableCell>
  );
}
