import { memo, useEffect, useState, type RefObject } from "react";

import { TableCell } from "@/components/ui/table";
import { Input } from "@/components/ui/input";
import { C } from "@/lib/design-tokens";

import type { Treatment, UpdateTreatmentInput } from "../../types";

// EMR-488: 薬剤行の3値（1回製剤量・用法・処方総量）を常時区別表示するセル。
// - 薬剤行以外では表示しない（会計数量・単価とは別枠で、mg 計算値(dose_amount_mg)とも混同しない）
// - 未保存は「—」（推測補完しない）。0 は 0 として表示する
// - PATCH に明示クリア経路はないため、空入力の commit は変更なし（送信しない）

export type TreatmentThreeValuesField =
  | "administration_dose_amount"
  | "administration_dose_unit"
  | "administration_instructions"
  | "prescription_total_amount"
  | "prescription_total_unit";

interface TreatmentThreeValuesCellProps {
  treatment: Treatment;
  /** 行でアクティブな編集フィールド（TreatmentRow の editField と共有する） */
  editField: string | null;
  inputRef: RefObject<HTMLInputElement | null>;
  onStartEdit: (field: TreatmentThreeValuesField) => void;
  onStopEdit: () => void;
  onUpdate: (treatmentId: string, input: UpdateTreatmentInput) => void;
}

const AMOUNT_LABEL = "1回";
const USAGE_LABEL = "用法";
const TOTAL_LABEL = "総量";

/** 数値入力の commit 値。空は未送信（変更なし）、不正数値は null（コミット中止）。 */
function parseAmountInput(raw: string): number | null | undefined {
  const text = raw.trim();
  if (text === "") return undefined;
  if (!/^\d+(\.\d+)?$/.test(text)) return null;
  return Number(text);
}

export const TreatmentThreeValuesCell = memo(function TreatmentThreeValuesCell({
  treatment,
  editField,
  inputRef,
  onStartEdit,
  onStopEdit,
  onUpdate,
}: TreatmentThreeValuesCellProps) {
  if (treatment.item_type !== "medicine") {
    return <TableCell className="min-w-[176px]" />;
  }
  return (
    <TableCell className="min-w-[176px]">
      <div className="flex flex-col gap-0.5 text-sm">
        {/* 1回製剤量 + 単位 */}
        <div className="flex items-center gap-1">
          <span className={`shrink-0 text-xs ${C.text40}`}>{AMOUNT_LABEL}</span>
          {editField === "administration_dose_amount" ? (
            <AmountEditor
              inputRef={inputRef}
              initial={treatment.administration_dose_amount}
              onStopEdit={onStopEdit}
              onCommit={(value) => {
                onUpdate(treatment.id, {
                  administration_dose_amount: value,
                  version: treatment.version,
                });
              }}
            />
          ) : (
            <button
              type="button"
              className={`h-6 min-w-11 px-1 text-left rounded-xxs transition-colors ${C.hoverBgLight}`}
              onClick={() => onStartEdit("administration_dose_amount")}
            >
              {treatment.administration_dose_amount == null ? (
                <span className={C.text40}>—</span>
              ) : (
                <span className={C.text}>{treatment.administration_dose_amount}</span>
              )}
            </button>
          )}
          {editField === "administration_dose_unit" ? (
            <TextEditor
              inputRef={inputRef}
              initial={treatment.administration_dose_unit}
              onStopEdit={onStopEdit}
              onCommit={(value) => {
                onUpdate(treatment.id, {
                  administration_dose_unit: value,
                  version: treatment.version,
                });
              }}
            />
          ) : (
            <button
              type="button"
              className={`h-6 min-w-8 px-1 text-left rounded-xxs transition-colors ${C.hoverBgLight}`}
              onClick={() => onStartEdit("administration_dose_unit")}
            >
              {treatment.administration_dose_unit == null ? (
                <span className={C.text40}>—</span>
              ) : (
                <span className={C.text40}>{treatment.administration_dose_unit}</span>
              )}
            </button>
          )}
        </div>
        {/* 用法 */}
        <div className="flex items-center gap-1">
          <span className={`shrink-0 text-xs ${C.text40}`}>{USAGE_LABEL}</span>
          {editField === "administration_instructions" ? (
            <TextEditor
              inputRef={inputRef}
              initial={treatment.administration_instructions}
              onStopEdit={onStopEdit}
              onCommit={(value) => {
                onUpdate(treatment.id, {
                  administration_instructions: value,
                  version: treatment.version,
                });
              }}
            />
          ) : (
            <button
              type="button"
              className={`h-6 flex-1 min-w-0 whitespace-normal break-words px-1 text-left text-sm rounded-xxs transition-colors ${C.hoverBgLight}`}
              onClick={() => onStartEdit("administration_instructions")}
            >
              {treatment.administration_instructions == null ? (
                <span className={C.text40}>—</span>
              ) : (
                <span className={C.text}>{treatment.administration_instructions}</span>
              )}
            </button>
          )}
        </div>
        {/* 処方総量 + 単位 */}
        <div className="flex items-center gap-1">
          <span className={`shrink-0 text-xs ${C.text40}`}>{TOTAL_LABEL}</span>
          {editField === "prescription_total_amount" ? (
            <AmountEditor
              inputRef={inputRef}
              initial={treatment.prescription_total_amount}
              onStopEdit={onStopEdit}
              onCommit={(value) => {
                onUpdate(treatment.id, {
                  prescription_total_amount: value,
                  version: treatment.version,
                });
              }}
            />
          ) : (
            <button
              type="button"
              className={`h-6 min-w-11 px-1 text-left rounded-xxs transition-colors ${C.hoverBgLight}`}
              onClick={() => onStartEdit("prescription_total_amount")}
            >
              {treatment.prescription_total_amount == null ? (
                <span className={C.text40}>—</span>
              ) : (
                <span className={C.text}>{treatment.prescription_total_amount}</span>
              )}
            </button>
          )}
          {editField === "prescription_total_unit" ? (
            <TextEditor
              inputRef={inputRef}
              initial={treatment.prescription_total_unit}
              onStopEdit={onStopEdit}
              onCommit={(value) => {
                onUpdate(treatment.id, {
                  prescription_total_unit: value,
                  version: treatment.version,
                });
              }}
            />
          ) : (
            <button
              type="button"
              className={`h-6 min-w-8 px-1 text-left rounded-xxs transition-colors ${C.hoverBgLight}`}
              onClick={() => onStartEdit("prescription_total_unit")}
            >
              {treatment.prescription_total_unit == null ? (
                <span className={C.text40}>—</span>
              ) : (
                <span className={C.text40}>{treatment.prescription_total_unit}</span>
              )}
            </button>
          )}
        </div>
      </div>
    </TableCell>
  );
});

function AmountEditor({
  initial,
  inputRef,
  onStopEdit,
  onCommit,
}: {
  initial: number | null | undefined;
  inputRef: RefObject<HTMLInputElement | null>;
  onStopEdit: () => void;
  onCommit: (value: number) => void;
}) {
  const [local, setLocal] = useState(initial == null ? "" : String(initial));
  const [invalid, setInvalid] = useState(false);

  useEffect(() => {
    inputRef.current?.focus();
    inputRef.current?.select();
  }, [inputRef]);

  const commit = () => {
    const parsed = parseAmountInput(local);
    if (parsed === null) {
      setInvalid(true);
      return;
    }
    if (parsed === undefined) {
      // 空入力は明示クリア経路がないため変更なし（BE PATCH 契約）
      onStopEdit();
      return;
    }
    onCommit(parsed);
  };

  return (
    <Input
      ref={inputRef}
      value={local}
      inputMode="decimal"
      onChange={(e) => {
        setLocal(e.target.value);
        setInvalid(false);
      }}
      onBlur={commit}
      onKeyDown={(e) => {
        if (e.key === "Enter") commit();
        if (e.key === "Escape") onStopEdit();
      }}
      aria-invalid={invalid}
      className={`h-6 w-16 text-sm px-1 ${invalid ? C.borderDanger : C.borderMedium}`}
    />
  );
}

function TextEditor({
  initial,
  inputRef,
  onStopEdit,
  onCommit,
}: {
  initial: string | null | undefined;
  inputRef: RefObject<HTMLInputElement | null>;
  onStopEdit: () => void;
  onCommit: (value: string) => void;
}) {
  const [local, setLocal] = useState(initial ?? "");

  useEffect(() => {
    inputRef.current?.focus();
    inputRef.current?.select();
  }, [inputRef]);

  const commit = () => {
    const text = local.trim();
    if (text === "") {
      // 空入力は明示クリア経路がないため変更なし（BE PATCH 契約）
      onStopEdit();
      return;
    }
    onCommit(text);
  };

  return (
    <Input
      ref={inputRef}
      value={local}
      onChange={(e) => setLocal(e.target.value)}
      onBlur={commit}
      onKeyDown={(e) => {
        if (e.key === "Enter") commit();
        if (e.key === "Escape") onStopEdit();
      }}
      className={`h-6 flex-1 min-w-0 text-sm px-1 ${C.borderMedium}`}
    />
  );
}
