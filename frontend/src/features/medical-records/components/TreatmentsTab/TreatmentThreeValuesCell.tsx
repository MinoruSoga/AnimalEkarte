import { memo, useEffect, useId, useState, type RefObject } from "react";

import { TableCell } from "@/components/ui/table";
import { Input } from "@/components/ui/input";
import { C } from "@/lib/design-tokens";

import type { Treatment, UpdateTreatmentInput } from "../../types";

// EMR-488: 薬剤行の3値（1回製剤量・用法・処方総量）を常時区別表示するセル。
// - 薬剤行以外では表示しない（会計数量・単価とは別枠で、mg 計算値(dose_amount_mg)とも混同しない）
// - 未保存は「—」（推測補完しない）。0 は 0 として表示する
// - 量と単位は DB のペア不変条件（001_init.sql ck_administration_dose_pair /
//   ck_prescription_total_pair: 両 NULL または 量非NULL かつ trim 後非空の単位）に従う。
//   片側が欠けた行では量+単位を同一操作で入力させ、単独 commit による CHECK 違反（500）を防ぐ。
//   有効ペアの行では片側だけの変更（単位維持で量だけ変更等）を許可する。
// - PATCH に明示クリア経路はないため、空入力の commit は変更なし（送信しない）
// - 行内編集 Input は label の代わりに aria-label でアクセシブルな名前を与える
//   （AGENTS.md アクセシビリティ規約・Codex review P1）。IME 変換確定の Enter では
//   commit しない（isComposing / keyCode 229 ガード。TreatmentQuantityCell と同一方針）
// - 表示ボタンは dense-table exception（h-9 min-w-11）のターゲットサイズを満たす

export type TreatmentThreeValuesField =
  | "administration_dose_amount"
  | "administration_dose_unit"
  | "administration_instructions"
  | "prescription_total_amount"
  | "prescription_total_unit";

type OnTreatmentUpdate = (
  treatmentId: string,
  input: UpdateTreatmentInput,
) => void | Promise<void>;

interface TreatmentThreeValuesCellProps {
  treatment: Treatment;
  /** 行でアクティブな編集フィールド（TreatmentRow の editField と共有する） */
  editField: string | null;
  inputRef: RefObject<HTMLInputElement | null>;
  onStartEdit: (field: TreatmentThreeValuesField) => void;
  onStopEdit: () => void;
  onUpdate: OnTreatmentUpdate;
}

const AMOUNT_LABEL = "1回";
const USAGE_LABEL = "用法";
const TOTAL_LABEL = "総量";

/** ペアが未確定（量または単位の片側でも欠落）の行。量+単位を同一操作で入力させる。 */
function isPairIncomplete(amount: number | null | undefined, unit: string | null | undefined): boolean {
  return amount == null || unit == null;
}

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

  const adminIncomplete = isPairIncomplete(
    treatment.administration_dose_amount,
    treatment.administration_dose_unit,
  );
  const totalIncomplete = isPairIncomplete(
    treatment.prescription_total_amount,
    treatment.prescription_total_unit,
  );

  return (
    <TableCell className="min-w-[176px]">
      <div className="flex flex-col gap-0.5 text-sm">
        {/* 1回製剤量 + 単位 */}
        <PairValueRow
          label={AMOUNT_LABEL}
          amount={treatment.administration_dose_amount}
          unit={treatment.administration_dose_unit}
          incomplete={adminIncomplete}
          editField={editField}
          inputRef={inputRef}
          onStopEdit={onStopEdit}
          onCommitPair={(amount, unit) => {
            onUpdate(treatment.id, {
              administration_dose_amount: amount,
              administration_dose_unit: unit,
              version: treatment.version,
            });
          }}
          onCommitAmountOnly={(amount) => {
            onUpdate(treatment.id, {
              administration_dose_amount: amount,
              version: treatment.version,
            });
          }}
          onCommitUnitOnly={(unit) => {
            onUpdate(treatment.id, {
              administration_dose_unit: unit,
              version: treatment.version,
            });
          }}
          onStartEdit={onStartEdit}
          amountField="administration_dose_amount"
          unitField="administration_dose_unit"
          amountAriaLabel="1回製剤量"
          unitAriaLabel="1回製剤量の単位"
        />
        {/* 用法（独立した free text） */}
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
              ariaLabel="用法"
            />
          ) : (
            <button
              type="button"
              className={`h-9 min-w-11 flex-1 whitespace-normal break-words px-1 text-left text-sm rounded-xxs transition-colors ${C.hoverBgLight}`}
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
        <PairValueRow
          label={TOTAL_LABEL}
          amount={treatment.prescription_total_amount}
          unit={treatment.prescription_total_unit}
          incomplete={totalIncomplete}
          editField={editField}
          inputRef={inputRef}
          onStopEdit={onStopEdit}
          onCommitPair={(amount, unit) => {
            onUpdate(treatment.id, {
              prescription_total_amount: amount,
              prescription_total_unit: unit,
              version: treatment.version,
            });
          }}
          onCommitAmountOnly={(amount) => {
            onUpdate(treatment.id, {
              prescription_total_amount: amount,
              version: treatment.version,
            });
          }}
          onCommitUnitOnly={(unit) => {
            onUpdate(treatment.id, {
              prescription_total_unit: unit,
              version: treatment.version,
            });
          }}
          onStartEdit={onStartEdit}
          amountField="prescription_total_amount"
          unitField="prescription_total_unit"
          amountAriaLabel="処方総量"
          unitAriaLabel="処方総量の単位"
        />
      </div>
    </TableCell>
  );
});

interface PairValueRowProps {
  label: string;
  amount: number | null | undefined;
  unit: string | null | undefined;
  incomplete: boolean;
  editField: string | null;
  inputRef: RefObject<HTMLInputElement | null>;
  onStopEdit: () => void;
  onCommitPair: (amount: number, unit: string) => void | Promise<void>;
  onCommitAmountOnly: (amount: number) => void | Promise<void>;
  onCommitUnitOnly: (unit: string) => void | Promise<void>;
  onStartEdit: (field: TreatmentThreeValuesField) => void;
  amountField: TreatmentThreeValuesField;
  unitField: TreatmentThreeValuesField;
  amountAriaLabel: string;
  unitAriaLabel: string;
}

/** 量+単位の1行。未確定ペアでは両入力の PairEditor、有効ペアでは片側編集を提供する。 */
const PairValueRow = memo(function PairValueRow({
  label,
  amount,
  unit,
  incomplete,
  editField,
  inputRef,
  onStopEdit,
  onCommitPair,
  onCommitAmountOnly,
  onCommitUnitOnly,
  onStartEdit,
  amountField,
  unitField,
  amountAriaLabel,
  unitAriaLabel,
}: PairValueRowProps) {
  return (
    <div className="flex items-center gap-1">
      <span className={`shrink-0 text-xs ${C.text40}`}>{label}</span>
      {editField === amountField || editField === unitField ? (
        incomplete ? (
          <PairEditor
            inputRef={inputRef}
            initialAmount={amount}
            initialUnit={unit}
            onStopEdit={onStopEdit}
            onCommit={onCommitPair}
            amountAriaLabel={amountAriaLabel}
            unitAriaLabel={unitAriaLabel}
          />
        ) : editField === amountField ? (
          <AmountEditor
            inputRef={inputRef}
            initial={amount}
            onStopEdit={onStopEdit}
            onCommit={onCommitAmountOnly}
            ariaLabel={amountAriaLabel}
          />
        ) : (
          <UnitEditor
            inputRef={inputRef}
            initial={unit}
            onStopEdit={onStopEdit}
            onCommit={onCommitUnitOnly}
            ariaLabel={unitAriaLabel}
          />
        )
      ) : (
        <>
          <button
            type="button"
            className={`h-9 min-w-11 px-1 text-left rounded-xxs transition-colors ${C.hoverBgLight}`}
            onClick={() => onStartEdit(amountField)}
            aria-label={amountAriaLabel}
          >
            {amount == null ? (
              <span className={C.text40}>—</span>
            ) : (
              <span className={C.text}>{amount}</span>
            )}
          </button>
          <button
            type="button"
            className={`h-9 min-w-11 px-1 text-left rounded-xxs transition-colors ${C.hoverBgLight}`}
            onClick={() => onStartEdit(unitField)}
            aria-label={unitAriaLabel}
          >
            {unit == null ? (
              <span className={C.text40}>—</span>
            ) : (
              <span className={C.text40}>{unit}</span>
            )}
          </button>
        </>
      )}
    </div>
  );
});

/** Enter / blur の IME 変換確定を commit に誤変換しないガード（TreatmentQuantityCell と同一方針）。 */
function isComposingKeyEvent(e: React.KeyboardEvent<HTMLInputElement>): boolean {
  const native = e.nativeEvent as KeyboardEvent & { keyCode: number };
  return native.isComposing || native.keyCode === 229;
}

/** 量+単位を同一操作で入力するエディタ（未確定ペア用）。Enter で両方を検証して同時送信する。 */
function PairEditor({
  initialAmount,
  initialUnit,
  inputRef,
  onStopEdit,
  onCommit,
  amountAriaLabel,
  unitAriaLabel,
}: {
  initialAmount: number | null | undefined;
  initialUnit: string | null | undefined;
  inputRef: RefObject<HTMLInputElement | null>;
  onStopEdit: () => void;
  onCommit: (amount: number, unit: string) => void | Promise<void>;
  amountAriaLabel: string;
  unitAriaLabel: string;
}) {
  const [localAmount, setLocalAmount] = useState(initialAmount == null ? "" : String(initialAmount));
  const [localUnit, setLocalUnit] = useState(initialUnit ?? "");
  const [invalid, setInvalid] = useState(false);
  const errorId = useId();

  useEffect(() => {
    inputRef.current?.focus();
    inputRef.current?.select();
  }, [inputRef]);

  const commit = async () => {
    const parsed = parseAmountInput(localAmount);
    const trimmedUnit = localUnit.trim();
    // ペア不変条件: 量は必須（数値）、単位は trim 後非空。片側だけでは送信しない
    if (parsed == null || trimmedUnit === "") {
      setInvalid(true);
      return;
    }
    // 未変更なら PATCH を送らない（既存エディタと同じ規約。二重送信・不要な版加算を防ぐ）
    if (parsed === initialAmount && trimmedUnit === (initialUnit ?? "")) {
      onStopEdit();
      return;
    }
    // PATCH の解決後に閉じる（失敗時は下書きを保持。Codex review minor）
    await onCommit(parsed, trimmedUnit);
    onStopEdit();
  };

  return (
    <div className="flex flex-col" data-testid={`pair-editor-${invalid ? "invalid" : "editing"}`}>
      <div className="flex items-center gap-1">
        <Input
          ref={inputRef}
          value={localAmount}
          inputMode="decimal"
          placeholder="量"
          aria-label={amountAriaLabel}
          aria-invalid={invalid}
          aria-describedby={invalid ? errorId : undefined}
          onChange={(e) => {
            setLocalAmount(e.target.value);
            setInvalid(false);
          }}
          onKeyDown={(e) => {
            if (isComposingKeyEvent(e)) return;
            if (e.key === "Enter") void commit();
            if (e.key === "Escape") onStopEdit();
          }}
          className="h-9 w-14 text-sm px-1"
        />
        <Input
          value={localUnit}
          placeholder="単位"
          aria-label={unitAriaLabel}
          aria-invalid={invalid}
          aria-describedby={invalid ? errorId : undefined}
          onChange={(e) => {
            setLocalUnit(e.target.value);
            setInvalid(false);
          }}
          onKeyDown={(e) => {
            if (isComposingKeyEvent(e)) return;
            if (e.key === "Enter") void commit();
            if (e.key === "Escape") onStopEdit();
          }}
          className="h-9 w-12 text-sm px-1"
        />
      </div>
      {invalid && (
        <p id={errorId} role="alert" className="text-xs text-[#C0392B]">
          量と単位を入力してください（量は数値・単位は空白のみ不可）
        </p>
      )}
    </div>
  );
}

function AmountEditor({
  initial,
  inputRef,
  onStopEdit,
  onCommit,
  ariaLabel,
}: {
  initial: number | null | undefined;
  inputRef: RefObject<HTMLInputElement | null>;
  onStopEdit: () => void;
  onCommit: (value: number) => void | Promise<void>;
  ariaLabel: string;
}) {
  const [local, setLocal] = useState(initial == null ? "" : String(initial));
  const [invalid, setInvalid] = useState(false);
  const errorId = useId();

  useEffect(() => {
    inputRef.current?.focus();
    inputRef.current?.select();
  }, [inputRef]);

  const commit = async () => {
    const parsed = parseAmountInput(local);
    if (parsed === null) {
      setInvalid(true);
      return;
    }
    if (parsed === undefined || parsed === initial) {
      // 空入力は明示クリア経路がないため変更なし。未変更も送らない（BE PATCH 契約・既存エディタ規約）
      onStopEdit();
      return;
    }
    // PATCH の解決後に閉じる（失敗時は下書きを保持。Codex review minor）
    await onCommit(parsed);
    onStopEdit();
  };

  return (
    <div className="flex flex-col">
      <Input
        ref={inputRef}
        value={local}
        inputMode="decimal"
        aria-label={ariaLabel}
        aria-invalid={invalid}
        aria-describedby={invalid ? errorId : undefined}
        onChange={(e) => {
          setLocal(e.target.value);
          setInvalid(false);
        }}
        onBlur={commit}
        onKeyDown={(e) => {
          if (isComposingKeyEvent(e)) return;
          if (e.key === "Enter") void commit();
          if (e.key === "Escape") onStopEdit();
        }}
        className="h-9 w-16 text-sm px-1"
      />
      {invalid && (
        <p id={errorId} role="alert" className="text-xs text-[#C0392B]">
          数値で入力してください
        </p>
      )}
    </div>
  );
}

function UnitEditor({
  initial,
  inputRef,
  onStopEdit,
  onCommit,
  ariaLabel,
}: {
  initial: string | null | undefined;
  inputRef: RefObject<HTMLInputElement | null>;
  onStopEdit: () => void;
  onCommit: (value: string) => void | Promise<void>;
  ariaLabel: string;
}) {
  const [local, setLocal] = useState(initial ?? "");

  useEffect(() => {
    inputRef.current?.focus();
    inputRef.current?.select();
  }, [inputRef]);

  const commit = async () => {
    const text = local.trim();
    if (text === "" || text === (initial ?? "")) {
      // 空入力は明示クリア経路がないため変更なし。未変更も送らない（BE PATCH 契約・既存エディタ規約）
      onStopEdit();
      return;
    }
    // PATCH の解決後に閉じる（失敗時は下書きを保持。Codex review minor）
    await onCommit(text);
    onStopEdit();
  };

  return (
    <Input
      ref={inputRef}
      value={local}
      aria-label={ariaLabel}
      onChange={(e) => setLocal(e.target.value)}
      onBlur={commit}
      onKeyDown={(e) => {
        if (isComposingKeyEvent(e)) return;
        if (e.key === "Enter") void commit();
        if (e.key === "Escape") onStopEdit();
      }}
      className={`h-9 min-w-11 text-sm px-1 ${C.borderMedium}`}
    />
  );
}

function TextEditor({
  initial,
  inputRef,
  onStopEdit,
  onCommit,
  ariaLabel,
}: {
  initial: string | null | undefined;
  inputRef: RefObject<HTMLInputElement | null>;
  onStopEdit: () => void;
  onCommit: (value: string) => void | Promise<void>;
  ariaLabel: string;
}) {
  const [local, setLocal] = useState(initial ?? "");

  useEffect(() => {
    inputRef.current?.focus();
    inputRef.current?.select();
  }, [inputRef]);

  const commit = async () => {
    const text = local.trim();
    if (text === "" || text === (initial ?? "")) {
      // 空入力は明示クリア経路がないため変更なし。未変更も送らない（BE PATCH 契約・既存エディタ規約）
      onStopEdit();
      return;
    }
    // PATCH の解決後に閉じる（失敗時は下書きを保持。Codex review minor）
    await onCommit(text);
    onStopEdit();
  };

  return (
    <Input
      ref={inputRef}
      value={local}
      aria-label={ariaLabel}
      onChange={(e) => setLocal(e.target.value)}
      onBlur={commit}
      onKeyDown={(e) => {
        if (isComposingKeyEvent(e)) return;
        if (e.key === "Enter") void commit();
        if (e.key === "Escape") onStopEdit();
      }}
      className={`h-9 flex-1 min-w-11 text-sm px-1 ${C.borderMedium}`}
    />
  );
}
