import { memo, useCallback, useState } from "react";
import { Check, Pencil, X } from "lucide-react";

import { Button } from "@/components/ui/button";
import { TableCell } from "@/components/ui/table";
import { DeleteIconButton } from "@/components/shared/DeleteIconButton/DeleteIconButton";
import { FormFieldError } from "@/components/shared/FormFieldError/FormFieldError";
import { C, ICON, STYLE } from "@/lib/design-tokens";
import { formatJSTDateTimeLocal, jstDateTimeLocalToISOString } from "@/lib/jst-date";
import type { BodyWeightUnit, UpdateVitalInput, Vital } from "../../types";

import {
  ADD_INPUT_CLASS,
  EDIT_INPUT_CLASS,
  displayNum,
  formatRecordedAt,
  parseVitalsNumber,
  toggleWeightValueAndUnit,
  type VitalsAddFormState,
} from "./vitals-tab-table-model";

interface VitalsDisplayRowProps {
  vital: Vital;
  canEdit: boolean;
  canDelete: boolean;
  deletePending: boolean;
  onStartEdit: (vitalId: string) => void;
  onDeleteClick: (vitalId: string) => void;
}

export function VitalsDisplayRow({
  vital,
  canEdit,
  canDelete,
  deletePending,
  onStartEdit,
  onDeleteClick,
}: VitalsDisplayRowProps) {
  return (
    <tr className={`border-b ${C.borderLight} ${C.hoverBgPageHalf} transition-colors h-12`}>
      <TableCell className={C.text}>{formatRecordedAt(vital.recorded_at)}</TableCell>
      <TableCell className={`text-right ${C.text}`}>{displayNum(vital.temperature)}</TableCell>
      <TableCell className={`text-right ${C.text}`}>{displayNum(vital.heart_rate)}</TableCell>
      <TableCell className={`text-right ${C.text}`}>{displayNum(vital.respiration_rate)}</TableCell>
      <TableCell className={`text-right ${C.text}`}>
        {displayNum(vital.weight)}
        <span className={`ml-0.5 text-2xs ${C.text40}`}>{vital.weight_unit}</span>
      </TableCell>
      <TableCell className={C.text60}>{vital.note ? vital.note : "-"}</TableCell>
      <TableCell>
        <div className="flex items-center justify-end gap-1">
          {canEdit ? (
            <button
              type="button"
              onClick={() => onStartEdit(vital.id)}
              className={`${STYLE.iconBtn32} ${C.text60} ${C.hoverText} ${C.hoverBgLight}`}
              title="編集"
            >
              <Pencil className={ICON.xs} />
            </button>
          ) : null}
          {canDelete ? (
            <DeleteIconButton
              onClick={() => onDeleteClick(vital.id)}
              disabled={deletePending}
            />
          ) : null}
        </div>
      </TableCell>
    </tr>
  );
}

interface VitalsEditRowProps {
  vital: Vital;
  onSave: (vitalId: string, input: UpdateVitalInput) => void;
  onCancel: () => void;
  isPending: boolean;
}

function buildEditRowForm(vital: Vital) {
  return {
    recorded_at: vital.recorded_at
      ? formatJSTDateTimeLocal(vital.recorded_at)
      : "",
    temperature: vital.temperature != null ? String(vital.temperature) : "",
    heart_rate: vital.heart_rate != null ? String(vital.heart_rate) : "",
    respiration_rate: vital.respiration_rate != null ? String(vital.respiration_rate) : "",
    weight: vital.weight != null ? String(vital.weight) : "",
    weight_unit: vital.weight_unit ?? "Kg",
    note: vital.note ?? "",
  };
}

export const VitalsEditRow = memo(function VitalsEditRow({
  vital,
  onSave,
  onCancel,
  isPending,
}: VitalsEditRowProps) {
  const [form, setForm] = useState(() => buildEditRowForm(vital));
  const [editFormErrors, setEditFormErrors] = useState<Record<string, string>>({});

  const handleChange = useCallback(
    (field: string, value: string | BodyWeightUnit) => {
      setForm((prev) => ({ ...prev, [field]: value }));
    },
    []
  );

  const handleWeightUnitToggle = useCallback(() => {
    setForm((prev) => {
      const next = toggleWeightValueAndUnit(
        prev.weight,
        prev.weight_unit as BodyWeightUnit
      );
      return { ...prev, weight: next.weight, weight_unit: next.weight_unit };
    });
  }, []);

  const handleSave = useCallback(() => {
    const errors: Record<string, string> = {};
    if (!form.recorded_at) {
      errors.recorded_at = "記録日時は必須です";
    } else {
      const recordedDate = new Date(jstDateTimeLocalToISOString(form.recorded_at));
      if (recordedDate > new Date()) {
        errors.recorded_at = "未来の日時は入力できません";
      }
    }
    const temp = parseVitalsNumber(form.temperature);
    if (temp !== null && (temp < 30 || temp > 45)) {
      errors.temperature = "体温は30〜45℃の範囲で入力してください";
    }
    if (Object.keys(errors).length > 0) {
      setEditFormErrors(errors);
      return;
    }
    setEditFormErrors({});

    onSave(vital.id, {
      recorded_at: jstDateTimeLocalToISOString(form.recorded_at),
      temperature: temp,
      heart_rate: parseVitalsNumber(form.heart_rate),
      respiration_rate: parseVitalsNumber(form.respiration_rate),
      weight: parseVitalsNumber(form.weight),
      weight_unit: form.weight_unit as BodyWeightUnit,
      note: form.note.trim() || null,
    });
  }, [form, onSave, vital.id]);

  return (
    <tr className={`border-b ${C.borderLight} ${C.bgNotice40}`}>
      <TableCell>
        <input
          type="datetime-local"
          value={form.recorded_at}
          onChange={(e) => handleChange("recorded_at", e.target.value)}
          aria-label="記録日時"
          className={EDIT_INPUT_CLASS}
        />
        <FormFieldError message={editFormErrors.recorded_at} />
      </TableCell>
      <TableCell>
        <input
          type="number"
          step="0.1"
          value={form.temperature}
          onChange={(e) => handleChange("temperature", e.target.value)}
          placeholder="-"
          aria-label={`体温 (${formatRecordedAt(vital.recorded_at)})`}
          className={EDIT_INPUT_CLASS}
        />
        <FormFieldError message={editFormErrors.temperature} />
      </TableCell>
      <TableCell>
        <input
          type="number"
          value={form.heart_rate}
          onChange={(e) => handleChange("heart_rate", e.target.value)}
          placeholder="-"
          aria-label={`心拍数 (${formatRecordedAt(vital.recorded_at)})`}
          className={EDIT_INPUT_CLASS}
        />
      </TableCell>
      <TableCell>
        <input
          type="number"
          value={form.respiration_rate}
          onChange={(e) => handleChange("respiration_rate", e.target.value)}
          placeholder="-"
          aria-label={`呼吸数 (${formatRecordedAt(vital.recorded_at)})`}
          className={EDIT_INPUT_CLASS}
        />
      </TableCell>
      <TableCell>
        <div className="flex items-center gap-1">
          <input
            type="number"
            step="0.01"
            value={form.weight}
            onChange={(e) => handleChange("weight", e.target.value)}
            placeholder="-"
            aria-label={`体重 (${formatRecordedAt(vital.recorded_at)})`}
            className={`${EDIT_INPUT_CLASS} text-right`}
          />
          <button
            type="button"
            onClick={handleWeightUnitToggle}
            className={`text-2xs px-1 h-6 rounded border ${C.borderMedium} ${C.bgPage} ${C.hoverBgPage} min-w-[24px]`}
          >
            {form.weight_unit}
          </button>
        </div>
      </TableCell>
      <TableCell>
        <input
          type="text"
          value={form.note}
          onChange={(e) => handleChange("note", e.target.value)}
          placeholder="メモ"
          aria-label={`メモ (${formatRecordedAt(vital.recorded_at)})`}
          className={EDIT_INPUT_CLASS}
        />
      </TableCell>
      <TableCell>
        <div className="flex items-center justify-end gap-1">
          <button
            type="button"
            onClick={handleSave}
            disabled={isPending}
            className={`${STYLE.iconBtn32} ${C.textStatusGreen} ${C.hoverBgStatusGreen}`}
            title="保存"
          >
            <Check className={ICON.xs} />
          </button>
          <button
            type="button"
            onClick={onCancel}
            disabled={isPending}
            className={`${STYLE.iconBtn32} ${C.text60} ${C.hoverBgLight}`}
            title="キャンセル"
          >
            <X className={ICON.xs} />
          </button>
        </div>
      </TableCell>
    </tr>
  );
});

interface VitalsAddRowProps {
  addForm: VitalsAddFormState;
  errors: Record<string, string>;
  isPending: boolean;
  /** 単一フィールドまたは体重単位トグルの原子パッチ（weight + weight_unit）。 */
  onChange: (patch: Partial<VitalsAddFormState>) => void;
  onSubmit: () => void;
  onCancel: () => void;
}

export function VitalsAddRow({
  addForm,
  errors,
  isPending,
  onChange,
  onSubmit,
  onCancel,
}: VitalsAddRowProps) {
  return (
    <div className={`flex flex-wrap items-start gap-2 px-3 py-2 border-t ${C.borderLight} ${C.bgPage30}`}>
      <div className="flex flex-col">
        <input
          autoFocus
          type="datetime-local"
          value={addForm.recorded_at}
          onChange={(e) => onChange({ recorded_at: e.target.value })}
          aria-label="記録日時"
          className={`${ADD_INPUT_CLASS} w-40`}
        />
        <FormFieldError message={errors.recorded_at} />
      </div>
      <div className="flex flex-col">
        <input
          type="number"
          step="0.1"
          value={addForm.temperature}
          onChange={(e) => onChange({ temperature: e.target.value })}
          placeholder="体温"
          aria-label="体温"
          className={`${ADD_INPUT_CLASS} w-20`}
        />
        <FormFieldError message={errors.temperature} />
      </div>
      <input
        type="number"
        value={addForm.heart_rate}
        onChange={(e) => onChange({ heart_rate: e.target.value })}
        placeholder="心拍数"
        aria-label="心拍数"
        className={`${ADD_INPUT_CLASS} w-20`}
      />
      <input
        type="number"
        value={addForm.respiration_rate}
        onChange={(e) => onChange({ respiration_rate: e.target.value })}
        placeholder="呼吸数"
        aria-label="呼吸数"
        className={`${ADD_INPUT_CLASS} w-20`}
      />
      <div className="flex items-center gap-1">
        <input
          type="number"
          step="0.01"
          value={addForm.weight}
          onChange={(e) => onChange({ weight: e.target.value })}
          placeholder="体重"
          aria-label="体重"
          className={`${ADD_INPUT_CLASS} w-20 text-right`}
        />
        <button
          type="button"
          onClick={() =>
            onChange(toggleWeightValueAndUnit(addForm.weight, addForm.weight_unit))
          }
          className={`text-2xs px-1 h-6 rounded border ${C.borderMedium} ${C.bgPage} ${C.hoverBgPage} min-w-[24px]`}
        >
          {addForm.weight_unit}
        </button>
      </div>
      <input
        type="text"
        value={addForm.note}
        onChange={(e) => onChange({ note: e.target.value })}
        onKeyDown={(e) => {
          if (e.key === "Enter") onSubmit();
          if (e.key === "Escape") onCancel();
        }}
        placeholder="メモ"
        aria-label="メモ"
        className={`${ADD_INPUT_CLASS} flex-1 min-w-[120px]`}
      />
      <Button
        size="sm"
        className={`${C.bgBrand} ${C.hoverBgBrand} ${C.hoverTextOnBrand} ${C.textOnBrand} rounded-full border-transparent transition-colors h-8 text-xs px-3`}
        onClick={onSubmit}
        disabled={isPending || !addForm.recorded_at}
      >
        追加
      </Button>
      <Button
        size="sm"
        variant="outline"
        className={`h-8 text-xs px-3 ${C.borderMedium}`}
        onClick={onCancel}
      >
        キャンセル
      </Button>
    </div>
  );
}
