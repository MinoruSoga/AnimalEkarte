// React/Framework
import { memo, useState, useCallback, useRef, useEffect, useLayoutEffect } from "react";

// Internal
import { Input } from "@/components/ui/input";
import { FormFieldError } from "@/components/shared/FormFieldError/FormFieldError";
import { C } from "@/lib/design-tokens";

// Relative
import type { Treatment, UpdateTreatmentInput } from "../../types";
import {
  TreatmentInsuranceCell,
  TreatmentRowActions,
  TreatmentSelectionCell,
  TreatmentSubtotalCell,
  TreatmentTypeCell,
} from "./TreatmentRowParts";

// ── Props ─────────────────────────────────────────────────────────────

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
}

// ── Component ─────────────────────────────────────────────────────────

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
}: TreatmentRowProps) {
  // インライン編集用ローカル状態
  const [editField, setEditField] = useState<string | null>(null);
  const [localContent, setLocalContent] = useState(treatment.content);
  const [localUnitPrice, setLocalUnitPrice] = useState(String(treatment.unit_price));
  const [localQuantity, setLocalQuantity] = useState(String(treatment.quantity));
  const [localDiscountAmount, setLocalDiscountAmount] = useState(
    String(treatment.discount_amount)
  );
  const [localMemo, setLocalMemo] = useState(treatment.memo);
  const [unitPriceError, setUnitPriceError] = useState<string>("");
  const [discountAmountError, setDiscountAmountError] = useState<string>("");
  const inputRef = useRef<HTMLInputElement>(null);

  // 外部からの treatment 変更を反映
  useEffect(() => {
    setLocalContent(treatment.content);
    setLocalUnitPrice(String(treatment.unit_price));
    setLocalQuantity(String(treatment.quantity));
    setLocalDiscountAmount(String(treatment.discount_amount));
    setLocalMemo(treatment.memo);
  }, [treatment]);

  // フォーカス時に input を選択
  useEffect(() => {
    if (editField && inputRef.current) {
      inputRef.current.select();
    }
  }, [editField]);

  // 新規追加直後: 数量フィールドへ自動フォーカス
  useLayoutEffect(() => {
    if (autoFocusQuantity) {
      setEditField("quantity");
      onAutoFocusDone?.();
    }
    // autoFocusQuantity は初回マウント時のみ使用するので deps に入れない
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // ── handlers ──

  const handleSelectedChange = useCallback(
    (checked: boolean | "indeterminate") => {
      onUpdate(treatment.id, { is_selected: checked === true });
    },
    [treatment.id, onUpdate]
  );

  const handleInsuranceChange = useCallback(
    (checked: boolean | "indeterminate") => {
      onUpdate(treatment.id, { is_insurance: checked === true });
    },
    [treatment.id, onUpdate]
  );

  const commitContent = useCallback(() => {
    const trimmed = localContent.trim();
    if (trimmed !== treatment.content) {
      onUpdate(treatment.id, { content: trimmed });
    }
    setEditField(null);
  }, [localContent, treatment.content, treatment.id, onUpdate]);

  const commitUnitPrice = useCallback(() => {
    const val = parseFloat(localUnitPrice) || 0;
    // BUG-072: 金額は0以上
    if (val < 0) {
      setUnitPriceError("金額は0以上を入力してください");
      return;
    }
    setUnitPriceError("");
    if (val !== treatment.unit_price) {
      onUpdate(treatment.id, { unit_price: val });
    }
    setEditField(null);
  }, [localUnitPrice, treatment.unit_price, treatment.id, onUpdate]);

  const commitQuantity = useCallback(() => {
    const val = parseFloat(localQuantity) || 1;
    if (val !== treatment.quantity) {
      onUpdate(treatment.id, { quantity: val });
    }
    setEditField(null);
  }, [localQuantity, treatment.quantity, treatment.id, onUpdate]);

  const commitDiscountAmount = useCallback(() => {
    const val = parseFloat(localDiscountAmount) || 0;
    // BUG-072: 値引き金額は0以上
    if (val < 0) {
      setDiscountAmountError("金額は0以上を入力してください");
      return;
    }
    setDiscountAmountError("");
    if (val !== treatment.discount_amount) {
      onUpdate(treatment.id, { discount_amount: val });
    }
    setEditField(null);
  }, [localDiscountAmount, treatment.discount_amount, treatment.id, onUpdate]);

  const commitMemo = useCallback(() => {
    const trimmed = localMemo.trim();
    if (trimmed !== treatment.memo) {
      onUpdate(treatment.id, { memo: trimmed });
    }
    setEditField(null);
  }, [localMemo, treatment.memo, treatment.id, onUpdate]);

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent, commit: () => void) => {
      if (e.key === "Enter") commit();
      if (e.key === "Escape") setEditField(null);
    },
    []
  );

  const handleDelete = useCallback(() => {
    onDelete(treatment.id);
  }, [treatment.id, onDelete]);

  const handleMoveUp = useCallback(() => {
    onMoveUp(treatment.id);
  }, [treatment.id, onMoveUp]);

  const handleMoveDown = useCallback(() => {
    onMoveDown(treatment.id);
  }, [treatment.id, onMoveDown]);

  // ── 小計計算 ──
  const subtotal = treatment.unit_price * treatment.quantity - treatment.discount_amount;

  return (
    <tr
      className={`border-b ${C.borderLight} ${C.hoverBgPageHalf} transition-colors ${
        !treatment.is_selected ? "opacity-50" : ""
      } ${isUpdating ? "pointer-events-none" : ""}`}
    >
      <TreatmentSelectionCell checked={treatment.is_selected} onChange={handleSelectedChange} />
      <TreatmentTypeCell itemType={treatment.item_type} />

      {/* 内容 */}
      <td className="px-3 py-2 min-w-[160px]">
        {editField === "content" ? (
          <Input
            ref={inputRef}
            value={localContent}
            onChange={(e) => setLocalContent(e.target.value)}
            onBlur={commitContent}
            onKeyDown={(e) => handleKeyDown(e, commitContent)}
            className={`h-8 text-sm px-2 ${C.borderMedium}`}
          />
        ) : (
          <button type="button"
            className={`w-full text-left text-sm ${C.text} ${C.hoverBgLight} px-1 py-0.5 rounded-[3px] transition-colors`}
            onClick={() => setEditField("content")}
          >
            {treatment.content || (
              <span className={C.text40}>内容を入力</span>
            )}
          </button>
        )}
      </td>

      <TreatmentInsuranceCell checked={treatment.is_insurance} onChange={handleInsuranceChange} />

      {/* 単価 */}
      <td className="px-3 py-2 w-28 text-right">
        {editField === "unit_price" ? (
          <>
            <Input
              ref={inputRef}
              type="number"
              min={0}
              value={localUnitPrice}
              onChange={(e) => setLocalUnitPrice(e.target.value)}
              onBlur={commitUnitPrice}
              onKeyDown={(e) => handleKeyDown(e, commitUnitPrice)}
              className={`h-8 text-sm text-right px-2 ${C.borderMedium}`}
            />
            <FormFieldError message={unitPriceError} />
          </>
        ) : (
          <button type="button"
            className={`w-full text-right text-sm ${C.text} ${C.hoverBgLight} px-1 py-0.5 rounded-[3px] transition-colors font-mono`}
            onClick={() => setEditField("unit_price")}
          >
            ¥{treatment.unit_price.toLocaleString()}
          </button>
        )}
      </td>

      {/* 数量 */}
      <td className="px-3 py-2 w-20 text-right">
        {editField === "quantity" ? (
          <Input
            ref={inputRef}
            type="number"
            step="0.1"
            min="0.1"
            value={localQuantity}
            onChange={(e) => setLocalQuantity(e.target.value)}
            onBlur={commitQuantity}
            onKeyDown={(e) => handleKeyDown(e, commitQuantity)}
            className={`h-8 text-sm text-right px-2 ${C.borderMedium}`}
          />
        ) : (
          <button type="button"
            className={`w-full text-right text-sm ${C.text} ${C.hoverBgLight} px-1 py-0.5 rounded-[3px] transition-colors`}
            onClick={() => setEditField("quantity")}
          >
            {treatment.quantity}
          </button>
        )}
      </td>

      {/* 値引き */}
      <td className="px-3 py-2 w-28 text-right">
        {editField === "discount_amount" ? (
          <>
            <Input
              ref={inputRef}
              type="number"
              min={0}
              value={localDiscountAmount}
              onChange={(e) => setLocalDiscountAmount(e.target.value)}
              onBlur={commitDiscountAmount}
              onKeyDown={(e) => handleKeyDown(e, commitDiscountAmount)}
              className={`h-8 text-sm text-right px-2 ${C.borderMedium}`}
            />
            <FormFieldError message={discountAmountError} />
          </>
        ) : (
          <button type="button"
            className={`w-full text-right text-sm ${
              treatment.discount_amount > 0 ? C.textDiscount : C.text40
            } ${canEditDiscount ? C.hoverBgLight : ""} px-1 py-0.5 rounded-[3px] transition-colors font-mono ${!canEditDiscount ? "cursor-not-allowed opacity-60" : ""}`}
            onClick={() => { if (canEditDiscount) setEditField("discount_amount"); }}
            disabled={!canEditDiscount}
            title={!canEditDiscount ? "値引の変更には権限が必要です" : undefined}
          >
            {treatment.discount_amount > 0
              ? `-¥${treatment.discount_amount.toLocaleString()}`
              : "—"}
          </button>
        )}
      </td>

      <TreatmentSubtotalCell subtotal={subtotal} />

      {/* メモ */}
      <td className="px-3 py-2 min-w-[120px]">
        {editField === "memo" ? (
          <Input
            ref={inputRef}
            value={localMemo}
            onChange={(e) => setLocalMemo(e.target.value)}
            onBlur={commitMemo}
            onKeyDown={(e) => handleKeyDown(e, commitMemo)}
            className={`h-8 text-sm px-2 ${C.borderMedium}`}
          />
        ) : (
          <button type="button"
            className={`w-full text-left text-sm ${C.text60} ${C.hoverBgLight} px-1 py-0.5 rounded-[3px] transition-colors`}
            onClick={() => setEditField("memo")}
          >
            {treatment.memo || <span className={C.text30}>メモ</span>}
          </button>
        )}
      </td>

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
