import { DatePicker } from "@/components/shared/DatePicker/DatePicker";
import { FormFieldError } from "@/components/shared/FormFieldError/FormFieldError";
import { NextScheduleField } from "@/components/shared/NextScheduleField";
import { Label } from "@/components/ui/label";
import { SearchableSelect } from "@/components/ui/searchable-select";
import { Textarea } from "@/components/ui/textarea";
import { C } from "@/lib/design-tokens";
import { toJSTWallDate } from "@/lib/jst-date";
import type { StaffItem } from "@/hooks/use-staffs";
import type { CheckupTypeItem } from "@/hooks/use-treatment-master";
import type { CheckupTypeFieldRow } from "../api/get-checkup-type-fields";
import { DynamicCheckupFields, type CheckupFieldValue } from "../components/DynamicCheckupFields";
import type { CheckupFormState } from "../hooks/use-checkup-form-model";
import { RequiredMark } from "@/components/shared/RequiredMark";

interface CheckupFieldsPanelProps {
  canSubmit: boolean;
  form: CheckupFormState;
  fieldErrors: Record<string, string>;
  checkupTypes: CheckupTypeItem[];
  staffs: StaffItem[];
  checkupFields: CheckupTypeFieldRow[];
  fieldValues: Record<number, CheckupFieldValue>;
  onDateChange: (value: string) => void;
  onCheckupTypeIdChange: (value: string) => void;
  onNextScheduleTypeChange: (value: string) => void;
  onNextDateChange: (value: string) => void;
  onDoctorIdChange: (value: string) => void;
  onResultChange: (value: string) => void;
  onFieldValueChange: (fieldId: number, value: CheckupFieldValue) => void;
}

export function CheckupFieldsPanel({
  canSubmit,
  form,
  fieldErrors,
  checkupTypes,
  staffs,
  checkupFields,
  fieldValues,
  onDateChange,
  onCheckupTypeIdChange,
  onNextScheduleTypeChange,
  onNextDateChange,
  onDoctorIdChange,
  onResultChange,
  onFieldValueChange,
}: CheckupFieldsPanelProps) {
  return (
    <fieldset
      aria-label="定期健診入力"
      disabled={!canSubmit}
      className={`lg:col-span-3 ${C.bgWhite} p-6 rounded-lg border ${C.borderLight} space-y-6 min-w-0`}
    >
      <div className="space-y-2">
        <Label htmlFor="checkup-date">
          実施日
          <RequiredMark />
        </Label>
        <DatePicker
          id="checkup-date"
          value={form.date}
          onChange={onDateChange}
          ariaRequired
          ariaInvalid={Boolean(fieldErrors.date)}
          ariaDescribedBy={fieldErrors.date ? "checkup-date-error" : undefined}
          disabledDays={{ after: toJSTWallDate(new Date()) }}
        />
        <FormFieldError id="checkup-date-error" message={fieldErrors.date} />
      </div>

      <div className="space-y-2">
        <Label htmlFor="checkup-type-select">
          健診種別
          <RequiredMark />
        </Label>
        <SearchableSelect
          id="checkup-type-select"
          value={form.checkupTypeId}
          onValueChange={onCheckupTypeIdChange}
          options={checkupTypes.map((checkupType) => ({
            value: String(checkupType.id),
            label: checkupType.name,
          }))}
          placeholder="選択してください"
          searchPlaceholder="健診種別を検索..."
          ariaInvalid={Boolean(fieldErrors.checkupTypeId)}
          ariaRequired
          ariaDescribedBy={fieldErrors.checkupTypeId ? "checkup-type-error" : undefined}
        />
        <FormFieldError id="checkup-type-error" message={fieldErrors.checkupTypeId} />
      </div>

      <NextScheduleField
        typeId="checkup-next-schedule"
        dateId="checkup-next-date"
        scheduleType={form.nextScheduleType}
        nextDate={form.nextDate}
        onScheduleTypeChange={onNextScheduleTypeChange}
        onNextDateChange={onNextDateChange}
      />

      <div className="space-y-2">
        <Label htmlFor="checkup-doctor-select">担当医</Label>
        <SearchableSelect
          id="checkup-doctor-select"
          value={form.doctorId}
          onValueChange={onDoctorIdChange}
          options={staffs
            .filter((staff) => staff.isActive && staff.staffType === "doctor")
            .map((staff) => ({ value: staff.id, label: staff.name }))}
          placeholder="選択してください"
          searchPlaceholder="担当医を検索..."
        />
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
        <Label htmlFor="checkup-result">結果・所見</Label>
        <Textarea
          id="checkup-result"
          value={form.result}
          onChange={(event) => onResultChange(event.target.value)}
          placeholder="健診結果・所見を入力"
          className="min-h-[120px]"
        />
      </div>
    </fieldset>
  );
}
