// React/Framework
import { memo, useState, useCallback, useRef, useEffect } from "react";

// External
import { ChevronUp, ChevronDown, Shield } from "lucide-react";
import { toast } from "sonner";

// Internal
import { Button } from "@/components/ui/button";
import { DeleteIconButton } from "@/components/shared/DeleteIconButton/DeleteIconButton";
import { Checkbox } from "@/components/ui/checkbox";
import { Input } from "@/components/ui/input";
import { C, BADGE, ICON } from "@/lib/design-tokens";

// Relative
import type { Treatment, TreatmentItemType, UpdateTreatmentInput } from "../../types";

// ── 静的JSX (モジュール定数に巻き上げ) ────────────────────────────────

const ITEM_TYPE_LABELS: Record<TreatmentItemType, string> = {
  consultation: "診療",
  procedure: "処置",
  medicine: "薬品",
  other: "その他",
};

const ITEM_TYPE_BADGE: Record<TreatmentItemType, string> = {
  consultation: BADGE.blue,
  procedure:    BADGE.purple,
  medicine:     BADGE.green,
  other:        BADGE.muted,
};

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

  // ── handlers ──

  const handleSelectedChange = useCallback(
    (checked: boolean | "indeterminate") => {
      onUpdate(treatment.id, { selected: checked === true });
    },
    [treatment.id, onUpdate]
  );

  const handleInsuranceChange = useCallback(
    (checked: boolean | "indeterminate") => {
      onUpdate(treatment.id, { insurance: checked === true });
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
      toast.error("金額は0以上を入力してください");
      setLocalUnitPrice(String(treatment.unit_price));
      setEditField(null);
      return;
    }
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
      toast.error("金額は0以上を入力してください");
      setLocalDiscountAmount(String(treatment.discount_amount));
      setEditField(null);
      return;
    }
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

  const badgeClass = ITEM_TYPE_BADGE[treatment.item_type] ?? BADGE.muted;
  const typeLabel = ITEM_TYPE_LABELS[treatment.item_type] ?? treatment.item_type;

  return (
    <tr
      className={`border-b ${C.borderLight} ${C.hoverBgPageHalf} transition-colors ${
        !treatment.selected ? "opacity-50" : ""
      } ${isUpdating ? "pointer-events-none" : ""}`}
    >
      {/* チェックボックス (selected) */}
      <td className="px-3 py-2 w-10 text-center">
        <Checkbox
          checked={treatment.selected}
          onCheckedChange={handleSelectedChange}
          className="data-[state=checked]:bg-[#2383E2] data-[state=checked]:border-[#2383E2]"
        />
      </td>

      {/* 種別バッジ */}
      <td className="px-3 py-2 w-24">
        <span
          className={`inline-flex items-center h-[22px] px-2 text-xs font-medium rounded border ${badgeClass}`}
        >
          {typeLabel}
        </span>
      </td>

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
          <button
            className={`w-full text-left text-sm ${C.text} ${C.hoverBgLight} px-1 py-0.5 rounded-[3px] transition-colors`}
            onClick={() => setEditField("content")}
          >
            {treatment.content || (
              <span className={C.text40}>内容を入力</span>
            )}
          </button>
        )}
      </td>

      {/* 保険アイコン */}
      <td className="px-3 py-2 w-16 text-center">
        <Checkbox
          checked={treatment.insurance}
          onCheckedChange={handleInsuranceChange}
          className="data-[state=checked]:bg-[#038B94] data-[state=checked]:border-[#038B94]"
        />
        {treatment.insurance ? (
          <Shield className={`${ICON.xs} mt-0.5 mx-auto ${C.textStatusGreen}`} />
        ) : null}
      </td>

      {/* 単価 */}
      <td className="px-3 py-2 w-28 text-right">
        {editField === "unit_price" ? (
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
        ) : (
          <button
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
          <button
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
        ) : (
          <button
            className={`w-full text-right text-sm ${
              treatment.discount_amount > 0 ? C.textDiscount : C.text40
            } ${C.hoverBgLight} px-1 py-0.5 rounded-[3px] transition-colors font-mono`}
            onClick={() => setEditField("discount_amount")}
          >
            {treatment.discount_amount > 0
              ? `-¥${treatment.discount_amount.toLocaleString()}`
              : "—"}
          </button>
        )}
      </td>

      {/* 小計 */}
      <td className="px-3 py-2 w-28 text-right">
        <span className={`text-sm font-medium ${C.text} font-mono`}>
          ¥{subtotal.toLocaleString()}
        </span>
      </td>

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
          <button
            className={`w-full text-left text-sm ${C.text60} ${C.hoverBgLight} px-1 py-0.5 rounded-[3px] transition-colors`}
            onClick={() => setEditField("memo")}
          >
            {treatment.memo || <span className={C.text30}>メモ</span>}
          </button>
        )}
      </td>

      {/* 並び替え & 削除 */}
      <td className="px-2 py-2 w-28">
        <div className="flex items-center gap-0.5 justify-end">
          <Button
            variant="ghost"
            size="icon"
            className={`size-7 ${C.text40} ${C.hoverText} disabled:opacity-20`}
            onClick={handleMoveUp}
            disabled={isFirst}
            title="上に移動"
          >
            <ChevronUp className={`${ICON.xs}`} />
          </Button>
          <Button
            variant="ghost"
            size="icon"
            className={`size-7 ${C.text40} ${C.hoverText} disabled:opacity-20`}
            onClick={handleMoveDown}
            disabled={isLast}
            title="下に移動"
          >
            <ChevronDown className={`${ICON.xs}`} />
          </Button>
          <DeleteIconButton
            onClick={handleDelete}
            className="size-7"
          />
        </div>
      </td>
    </tr>
  );
});
