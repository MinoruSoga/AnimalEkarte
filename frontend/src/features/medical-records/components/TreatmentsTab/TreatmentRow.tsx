// React/Framework
import { memo, useState, useCallback, useMemo, useRef, useEffect, useLayoutEffect } from "react";

// Internal
import { Input } from "@/components/ui/input";
import { TableCell } from "@/components/ui/table";
import { FormFieldError } from "@/components/shared/FormFieldError/FormFieldError";
import { C } from "@/lib/design-tokens";
import { calculateDose } from "@/lib/medicine-dose";
import { formatCurrency } from "@/lib/format/number";

// Relative
import {
  buildDoseCalcInput,
  toDoseParamsAuthority,
  useMedicineDoseParams,
  type MedicineDoseContext,
} from "../../api/medicine-dose-lookup";
import type { Treatment, UpdateTreatmentInput } from "../../types";
import {
  TreatmentInsuranceCell,
  TreatmentRowActions,
  TreatmentSelectionCell,
  TreatmentSubtotalCell,
  TreatmentTypeCell,
} from "./TreatmentRowParts";
import { computeDoseGate, resolveDoseGateSource } from "./treatment-row-dose-gate";

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
  /** #201: 投与量自動計算プレビュー・保存ゲート判定に使うコンテキスト */
  doseContext: MedicineDoseContext;
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
  doseContext,
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

  // #201: 投与量自動計算（薬剤行のみ）。medicine_id が無い/薬剤以外の行は全て null に落ちる
  // （評価情報が無い場合は既存の手動入力・保存挙動を維持する）。
  const isMedicineRow = treatment.item_type === "medicine" && !!treatment.medicine_id;
  const medicine = isMedicineRow
    ? doseContext.medicines?.find((m) => m.id === treatment.medicine_id)
    : undefined;
  const doseParamsQuery = useMedicineDoseParams(
    isMedicineRow ? treatment.medicine_id : undefined
  );
  const doseParamsAuthority = useMemo(
    () =>
      toDoseParamsAuthority(isMedicineRow ? treatment.medicine_id : undefined, {
        data: doseParamsQuery.data,
        isError: doseParamsQuery.isError,
      }),
    [isMedicineRow, treatment.medicine_id, doseParamsQuery.data, doseParamsQuery.isError]
  );
  const doseCalcInput = useMemo(
    () =>
      buildDoseCalcInput(
        medicine,
        doseParamsAuthority.status === "success" ? doseParamsAuthority.params : undefined,
        doseContext.petSpecies,
        doseContext.weightKg
      ),
    [medicine, doseParamsAuthority, doseContext.petSpecies, doseContext.weightKg]
  );
  const doseGateSource = useMemo(
    () => resolveDoseGateSource(doseCalcInput, doseParamsAuthority),
    [doseCalcInput, doseParamsAuthority]
  );
  const dosePreview = useMemo(
    () => (doseCalcInput ? calculateDose(doseCalcInput) : null),
    [doseCalcInput]
  );
  const currentGate = useMemo(
    () => computeDoseGate(doseGateSource, treatment.quantity),
    [doseGateSource, treatment.quantity]
  );
  const isDoseLookupFailed = doseGateSource.kind === "technical_failure";
  // react-review-201 HIGH-2: 警告テキストを aria-describedby で数量セルに紐付けるための安定 id。
  const doseWarningId = `dose-warning-${treatment.id}`;

  // #201: 保存操作で検出した絶対上限超過を、保存値が更新されるまでインライン表示する。
  const [attemptedDoseBlockReason, setAttemptedDoseBlockReason] = useState("");
  const doseBlockReason = attemptedDoseBlockReason || currentGate.blockReason;
  const hasDoseMessage = doseBlockReason !== "" || currentGate.warning !== "none";

  // 外部からの treatment 変更を反映
  useEffect(() => {
    setLocalContent(treatment.content);
    setLocalUnitPrice(String(treatment.unit_price));
    setLocalQuantity(String(treatment.quantity));
    setLocalDiscountAmount(String(treatment.discount_amount));
    setLocalMemo(treatment.memo);
    setAttemptedDoseBlockReason("");
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
    setEditField(null);
    if (val === treatment.quantity) return;

    // #201: マスタの絶対上限超過を物理ブロックする。
    // TASK-025: technical failure も isBlocked として通常保存を止める。
    // 下限未満・推奨値からの乖離・評価情報不足（missing）は保存を継続する。
    const gate = computeDoseGate(doseGateSource, val);
    if (gate.isBlocked) {
      setLocalQuantity(String(treatment.quantity));
      setAttemptedDoseBlockReason(gate.blockReason);
      return;
    }
    setAttemptedDoseBlockReason("");
    onUpdate(treatment.id, { quantity: val });
  }, [localQuantity, treatment.quantity, treatment.id, onUpdate, doseGateSource]);

  const handleRetryDoseParamsLookup = useCallback(() => {
    void doseParamsQuery.refetch();
  }, [doseParamsQuery]);

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
      <TableCell className="min-w-[160px]">
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
            className={`w-full text-left text-sm ${C.text} ${C.hoverBgLight} px-1 py-0.5 rounded-xxs transition-colors`}
            onClick={() => setEditField("content")}
          >
            {treatment.content || (
              <span className={C.text40}>内容を入力</span>
            )}
          </button>
        )}
      </TableCell>

      <TreatmentInsuranceCell checked={treatment.is_insurance} onChange={handleInsuranceChange} />

      {/* 単価 */}
      <TableCell className="w-28 text-right">
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
            className={`w-full text-right text-sm ${C.text} ${C.hoverBgLight} px-1 py-0.5 rounded-xxs transition-colors font-mono`}
            onClick={() => setEditField("unit_price")}
          >
            {formatCurrency(treatment.unit_price)}
          </button>
        )}
      </TableCell>

      {/* 数量 */}
      <TableCell className="w-20 text-right">
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
            aria-label="数量"
            aria-describedby={hasDoseMessage ? doseWarningId : undefined}
            aria-invalid={doseBlockReason !== "" ? true : undefined}
          />
        ) : (
          <button type="button"
            className={`w-full text-right text-sm ${C.hoverBgLight} px-1 py-0.5 rounded-xxs transition-colors ${
              currentGate.warning === "exceeds-max"
                ? C.textRed700
                : currentGate.warning === "below-min"
                  ? C.textWarning
                  : C.text
            }`}
            onClick={() => setEditField("quantity")}
            aria-describedby={hasDoseMessage ? doseWarningId : undefined}
          >
            {treatment.quantity}
          </button>
        )}
        {/* #201: 色だけに依存しない警告表示（react-review-201 HIGH-2）。アイコン/接頭辞テキスト併用 +
            aria-describedby で読み上げ可能にする。 */}
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
                onClick={handleRetryDoseParamsLookup}
                aria-label="投与量パラメータの取得を再試行する"
              >
                再試行する
              </button>
            ) : null}
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
        {/* #201: per_weight 計算対象の薬剤のみ、丸め前/丸め後 mg と推奨値を併記する */}
        {dosePreview?.eligible ? (
          <div className={`text-xs ${C.text40} text-right mt-0.5`}>
            推奨{dosePreview.quantity}（{dosePreview.rawMg.toFixed(1)}→{dosePreview.effectiveDoseMg.toFixed(1)}mg）
          </div>
        ) : null}
        {/* FE-refactor 残件1: 保存済み投与量スナップショットの read-only 表示（監査目的・上の推奨値とは独立）。
            保存値が無い行（medicine 以外・per_weight 計算対象外）は非表示。乖離警告は未実装（要件未確定）。 */}
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
      </TableCell>

      {/* 値引き */}
      <TableCell className="w-28 text-right">
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
            } ${canEditDiscount ? C.hoverBgLight : ""} px-1 py-0.5 rounded-xxs transition-colors font-mono ${!canEditDiscount ? "cursor-not-allowed opacity-60" : ""}`}
            onClick={() => { if (canEditDiscount) setEditField("discount_amount"); }}
            disabled={!canEditDiscount}
            title={!canEditDiscount ? "値引の変更には権限が必要です" : undefined}
          >
            {treatment.discount_amount > 0
              ? `-${formatCurrency(treatment.discount_amount)}`
              : "—"}
          </button>
        )}
      </TableCell>

      <TreatmentSubtotalCell subtotal={subtotal} />

      {/* メモ */}
      <TableCell className="min-w-[120px]">
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
            className={`w-full text-left text-sm ${C.text60} ${C.hoverBgLight} px-1 py-0.5 rounded-xxs transition-colors`}
            onClick={() => setEditField("memo")}
          >
            {treatment.memo || <span className={C.text30}>メモ</span>}
          </button>
        )}
      </TableCell>

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
