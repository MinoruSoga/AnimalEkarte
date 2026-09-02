import { memo, useCallback, useState, type ChangeEvent } from "react";
import { Check, Pencil, X } from "lucide-react";

import { Button } from "@/components/ui/button";
import { TableCell } from "@/components/ui/table";
import { DeleteIconButton } from "@/components/shared/DeleteIconButton/DeleteIconButton";
import { FormFieldError } from "@/components/shared/FormFieldError/FormFieldError";
import { DatePicker } from "@/components/shared/DatePicker/DatePicker";
import { NextScheduleField, calculateNextDate } from "@/components/shared/NextScheduleField";
import { CheckupAlertBadge } from "@/components/shared/CheckupAlertBadge/CheckupAlertBadge";
import { Label } from "@/components/ui/label";
import { C, ICON, STYLE } from "@/lib/design-tokens";
import type { StaffItem } from "@/hooks/use-staffs";
import type { CheckupTypeItem } from "@/hooks/use-treatment-master";
import type { CheckupTypeFieldRow } from "@/hooks/use-checkup-fields";
import {
  DynamicCheckupFields,
  type CheckupFieldValue,
} from "@/components/shared/DynamicCheckupFields/DynamicCheckupFields";

import type { Checkup, UpdateCheckupInput } from "../../api/checkups";
import type { AddCheckupFormState } from "./checkups-tab-table-model";

interface CheckupEditRowProps {
  checkup: Checkup;
  onSave: (checkupId: string, input: UpdateCheckupInput) => void;
  onCancel: () => void;
  isPending: boolean;
  checkupTypes: CheckupTypeItem[];
  staffs: StaffItem[];
}

export const CheckupEditRow = memo(function CheckupEditRow({
  checkup,
  onSave,
  onCancel,
  isPending,
  checkupTypes,
  staffs,
}: CheckupEditRowProps) {
  const [form, setForm] = useState<UpdateCheckupInput>({
    checkup_type_id: Number(checkup.checkup_type_id),
    date: checkup.date,
    next_date: checkup.next_date ?? "",
    doctor_id: checkup.doctor_id ? Number(checkup.doctor_id) : null,
    result: checkup.result,
  });

  const handleChange = useCallback(
    (field: keyof UpdateCheckupInput, value: string | number | null) => {
      setForm((prev) => ({ ...prev, [field]: value }));
    },
    [],
  );

  const handleSave = useCallback(() => {
    const payload: UpdateCheckupInput = { ...form };
    if (form.doctor_id === null) {
      payload.doctor_id_clear = true;
    }
    onSave(checkup.id, payload);
  }, [checkup.id, form, onSave]);

  return (
    <tr className={`border-b ${C.borderLight} ${C.bgNotice40}`}>
      <TableCell>
        <DatePicker
          value={form.date ?? ""}
          onChange={(value) => handleChange("date", value)}
          placeholder="日付"
          className="h-8 w-full"
        />
      </TableCell>
      <TableCell>
        <CheckupTypeSelect
          value={form.checkup_type_id ?? ""}
          checkupTypes={checkupTypes}
          onChange={(value) => handleChange("checkup_type_id", value ? Number(value) : null)}
        />
      </TableCell>
      <TableCell>
        <DatePicker
          value={form.next_date ?? ""}
          onChange={(value) => handleChange("next_date", value || null)}
          placeholder="次回日"
          className="h-8 w-full"
        />
      </TableCell>
      <TableCell>
        <StaffSelect
          value={form.doctor_id ?? ""}
          staffs={staffs}
          emptyLabel="-"
          onChange={(value) => handleChange("doctor_id", value ? Number(value) : null)}
        />
      </TableCell>
      <TableCell>
        <input
          type="text"
          value={form.result ?? ""}
          onChange={(e) => handleChange("result", e.target.value)}
          placeholder="結果を入力..."
          aria-label={`結果 (${checkup.date})`}
          className={`h-8 text-sm border ${C.borderMedium} rounded-xxs px-2 ${C.bgWhite} ${C.text} outline-none ${C.focusBorderAccent} w-full`}
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

interface CheckupDisplayRowProps {
  checkup: Checkup;
  canEdit: boolean;
  canDelete: boolean;
  isFinalized: boolean;
  deletePending: boolean;
  onStartEdit: (checkupId: string) => void;
  onDelete: (checkupId: string) => void;
}

export function CheckupDisplayRow({
  checkup,
  canEdit,
  canDelete,
  isFinalized,
  deletePending,
  onStartEdit,
  onDelete,
}: CheckupDisplayRowProps) {
  return (
    <tr className={`border-b ${C.borderLight} ${C.hoverBgPageHalf} transition-colors h-12`}>
      <TableCell className={C.text}>{checkup.date}</TableCell>
      <TableCell className={C.text}>
        {checkup.checkup_type?.name ?? checkup.checkup_type_id}
      </TableCell>
      <TableCell className={C.text60}>
        {checkup.next_date ? (
          <div className="flex items-center gap-1.5">
            <span>{checkup.next_date}</span>
            <CheckupAlertBadge nextDate={checkup.next_date} />
          </div>
        ) : (
          "-"
        )}
      </TableCell>
      <TableCell className={C.text60}>
        {checkup.doctor?.name ?? (checkup.doctor_id ? checkup.doctor_id : "-")}
      </TableCell>
      <TableCell className={C.text}>{checkup.result}</TableCell>
      <TableCell>
        <div className="flex items-center justify-end gap-1">
          {canEdit && !isFinalized ? (
            <button
              type="button"
              onClick={() => onStartEdit(checkup.id)}
              className={`${STYLE.iconBtn32} ${C.text60} ${C.hoverText} ${C.hoverBgLight}`}
              title="編集"
            >
              <Pencil className={ICON.xs} />
            </button>
          ) : null}
          {canDelete && !isFinalized ? (
            <DeleteIconButton
              onClick={() => onDelete(checkup.id)}
              disabled={deletePending}
            />
          ) : null}
        </div>
      </TableCell>
    </tr>
  );
}

interface CheckupAddRowProps {
  addForm: AddCheckupFormState;
  errors: Record<string, string>;
  checkupTypes: CheckupTypeItem[];
  staffs: StaffItem[];
  isPending: boolean;
  checkupFields: CheckupTypeFieldRow[];
  fieldValues: Record<number, CheckupFieldValue>;
  onChange: (field: keyof AddCheckupFormState, value: string) => void;
  onFieldValueChange: (fieldId: number, value: CheckupFieldValue) => void;
  onSubmit: () => void;
  onCancel: () => void;
}

export function CheckupAddRow({
  addForm,
  errors,
  checkupTypes,
  staffs,
  isPending,
  checkupFields,
  fieldValues,
  onChange,
  onFieldValueChange,
  onSubmit,
  onCancel,
}: CheckupAddRowProps) {
  return (
    <div className={`${C.bgWhite} border-t ${C.borderLight} p-6 space-y-6`}>
      <div className="grid grid-cols-1 gap-6 md:grid-cols-2">
        <div className="space-y-2">
          <Label htmlFor="checkup-tab-date">実施日</Label>
          <DatePicker
            id="checkup-tab-date"
            value={addForm.date}
            onChange={(value) => {
              onChange("date", value);
              const calculated = calculateNextDate(value, addForm.next_schedule_type);
              if (calculated) onChange("next_date", calculated);
            }}
            placeholder="日付"
            className="w-full min-w-[10rem]"
          />
          <FormFieldError message={errors.date} />
        </div>
        <div className="space-y-2">
          <Label htmlFor="checkup-tab-type">健診種別</Label>
          <CheckupTypeSelect
            value={addForm.checkup_type_id}
            checkupTypes={checkupTypes}
            emptyLabel="選択"
            onChange={(value) => onChange("checkup_type_id", value)}
          />
          <FormFieldError message={errors.checkup_type_id} />
        </div>
      </div>
      <div className="grid grid-cols-1 gap-6 md:grid-cols-2">
        <NextScheduleField
          typeId="checkup-tab-next-schedule"
          dateId="checkup-tab-next-date"
          scheduleType={addForm.next_schedule_type}
          nextDate={addForm.next_date}
          onScheduleTypeChange={(value) => {
            onChange("next_schedule_type", value);
            const calculated = calculateNextDate(addForm.date, value);
            if (calculated) onChange("next_date", calculated);
          }}
          onNextDateChange={(value) => onChange("next_date", value)}
        />
        <div className="space-y-2">
          <Label htmlFor="checkup-tab-doctor">担当医</Label>
          <StaffSelect
            value={addForm.doctor_id}
            staffs={staffs}
            emptyLabel="担当医"
            onChange={(value) => onChange("doctor_id", value)}
          />
        </div>
      </div>
      {checkupFields.length > 0 ? (
        <div className={`border-t ${C.borderLight} pt-6`}>
          <h2 className={`mb-4 text-sm font-medium ${C.text}`}>健診項目</h2>
          <DynamicCheckupFields
            fields={checkupFields}
            values={fieldValues}
            onChange={onFieldValueChange}
          />
        </div>
      ) : null}
      <div className="space-y-2">
        <Label htmlFor="checkup-tab-result">結果・所見</Label>
        <input
          id="checkup-tab-result"
          autoFocus
          type="text"
          placeholder="結果を入力..."
          value={addForm.result}
          onChange={(e) => onChange("result", e.target.value)}
          onKeyDown={(e) => {
            if (e.key === "Enter") onSubmit();
            if (e.key === "Escape") onCancel();
          }}
          aria-label="結果"
          className={`h-10 w-full text-sm border ${C.borderMedium} rounded-xxs px-2 ${C.bgWhite} ${C.text} outline-none ${C.focusBorderAccent}`}
        />
      </div>
      <div className="flex justify-end gap-2">
        <Button
          size="sm"
          variant="outline"
          className={`h-10 text-sm px-3 ${C.borderMedium}`}
          onClick={onCancel}
        >
          キャンセル
        </Button>
        <Button
          size="sm"
          className={`${C.bgBrand} ${C.hoverBgBrand} ${C.hoverTextOnBrand} ${C.textOnBrand} rounded-full border-transparent transition-colors h-10 text-sm px-4`}
          onClick={onSubmit}
          disabled={isPending || !addForm.date || !addForm.checkup_type_id}
        >
          追加
        </Button>
      </div>
    </div>
  );
}

function CheckupTypeSelect({
  value,
  checkupTypes,
  onChange,
  emptyLabel = "選択してください",
}: {
  value: string | number;
  checkupTypes: CheckupTypeItem[];
  onChange: (value: string) => void;
  emptyLabel?: string;
}) {
  return (
    <select
      value={value}
      onChange={(e: ChangeEvent<HTMLSelectElement>) => onChange(e.target.value)}
      className={`h-8 text-sm border ${C.borderMedium} rounded-xxs px-2 ${C.bgWhite} ${C.text} outline-none ${C.focusBorderAccent} w-32`}
    >
      <option value="">{emptyLabel}</option>
      {checkupTypes.map((type) => (
        <option key={type.id} value={type.id}>
          {type.name}
        </option>
      ))}
    </select>
  );
}

function StaffSelect({
  value,
  staffs,
  onChange,
  emptyLabel,
}: {
  value: string | number;
  staffs: StaffItem[];
  onChange: (value: string) => void;
  emptyLabel: string;
}) {
  return (
    <select
      value={value}
      onChange={(e: ChangeEvent<HTMLSelectElement>) => onChange(e.target.value)}
      className={`h-8 text-sm border ${C.borderMedium} rounded-xxs px-2 ${C.bgWhite} ${C.text} outline-none ${C.focusBorderAccent} w-32`}
    >
      <option value="">{emptyLabel}</option>
      {staffs.map((staff) => (
        <option key={staff.id} value={staff.id}>
          {staff.name}
        </option>
      ))}
    </select>
  );
}
