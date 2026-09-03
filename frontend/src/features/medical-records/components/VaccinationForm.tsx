// React/Framework
import { memo } from "react";

// Internal
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { SearchableSelect } from "@/components/ui/searchable-select";
import { DatePicker } from "@/components/shared/DatePicker";
import { FormFieldError } from "@/components/shared/FormFieldError";
import { MasterLink } from "@/components/shared/MasterLink";
import { NextScheduleField } from "@/components/shared/NextScheduleField";
import { SubmitButton } from "@/components/shared/Form/SubmitButton";
import { C } from "@/lib/design-tokens";

interface VaccineOption {
  value: string;
  label: string;
}

interface VaccinationFormProps {
  vaccineOptions: VaccineOption[];
  vaccineName: string;
  setVaccineName: (v: string) => void;
  date: string;
  setDate: (v: string) => void;
  supplemental: string;
  setSupplemental: (v: string) => void;
  lot1: string;
  setLot1: (v: string) => void;
  lot2: string;
  setLot2: (v: string) => void;
  lot3: string;
  setLot3: (v: string) => void;
  lot4: string;
  setLot4: (v: string) => void;
  nextScheduleType: string;
  setNextScheduleType: (v: string) => void;
  nextDate: string;
  setNextDate: (v: string) => void;
  remarks: string;
  setRemarks: (v: string) => void;
  /** BUG-015: 必須未選択時のインラインエラー（独立フォームと同文言） */
  fieldErrors?: Record<string, string>;
}

export const VaccinationForm = memo(function VaccinationForm({
  vaccineOptions,
  vaccineName,
  setVaccineName,
  date,
  setDate,
  supplemental,
  setSupplemental,
  lot1,
  setLot1,
  lot2,
  setLot2,
  lot3,
  setLot3,
  lot4,
  setLot4,
  nextScheduleType,
  setNextScheduleType,
  nextDate,
  setNextDate,
  remarks,
  setRemarks,
  fieldErrors = {},
}: VaccinationFormProps) {
  return (
    <div className="col-span-1 flex flex-col gap-4 lg:col-span-3">
      {/* Row 1: Name and Date */}
      <div className="grid w-full grid-cols-1 gap-4 sm:grid-cols-2">
        <div className="flex flex-col gap-1.5">
          <div className="flex items-center justify-between">
            <Label className={`text-sm font-medium ${C.text60}`}>予防接種名</Label>
            <MasterLink category="vaccine" label="編集" className="text-2xs" />
          </div>
          <SearchableSelect
            value={vaccineName}
            onValueChange={setVaccineName}
            options={vaccineOptions}
            placeholder="ワクチンを選択"
            searchPlaceholder="ワクチンを検索..."
            ariaInvalid={Boolean(fieldErrors.vaccineId)}
            ariaDescribedBy={fieldErrors.vaccineId ? "mr-vaccine-error" : undefined}
          />
          <FormFieldError id="mr-vaccine-error" message={fieldErrors.vaccineId} />
        </div>
        <div className="flex flex-col gap-1.5">
          <Label className={`text-sm font-medium ${C.text60}`}>予防接種日</Label>
          <DatePicker value={date} onChange={setDate} />
          <FormFieldError id="mr-vaccine-date-error" message={fieldErrors.date} />
        </div>
      </div>

      {/* Supplemental */}
      <div className="flex flex-col gap-1.5">
        <Label className={`text-sm font-medium ${C.text60}`}>補助説明</Label>
        <Input
          value={supplemental}
          onChange={(e) => setSupplemental(e.target.value)}
          className={`${C.bgWhite} ${C.borderMedium} h-10 text-sm ${C.text}`}
        />
      </div>

      {/* LOT Numbers */}
      <div className="grid w-full grid-cols-1 gap-4 sm:grid-cols-2">
        <div className="flex flex-col gap-1.5">
          <Label className={`text-sm font-medium ${C.text60}`}>LOT1</Label>
          <Input
            value={lot1}
            onChange={(e) => setLot1(e.target.value)}
            className={`${C.bgWhite} ${C.borderMedium} h-10 text-sm ${C.text}`}
          />
        </div>
        <div className="flex flex-col gap-1.5">
          <Label className={`text-sm font-medium ${C.text60}`}>LOT2</Label>
          <Input
            value={lot2}
            onChange={(e) => setLot2(e.target.value)}
            className={`${C.bgWhite} ${C.borderMedium} h-10 text-sm ${C.text}`}
          />
        </div>
        <div className="flex flex-col gap-1.5">
          <Label className={`text-sm font-medium ${C.text60}`}>LOT3</Label>
          <Input
            value={lot3}
            onChange={(e) => setLot3(e.target.value)}
            className={`${C.bgWhite} ${C.borderMedium} h-10 text-sm ${C.text}`}
          />
        </div>
        <div className="flex flex-col gap-1.5">
          <Label className={`text-sm font-medium ${C.text60}`}>LOT4</Label>
          <Input
            value={lot4}
            onChange={(e) => setLot4(e.target.value)}
            className={`${C.bgWhite} ${C.borderMedium} h-10 text-sm ${C.text}`}
          />
        </div>
      </div>

      <NextScheduleField
        typeId="mr-vaccination-next-schedule"
        dateId="mr-vaccination-next-date"
        scheduleType={nextScheduleType}
        nextDate={nextDate}
        onScheduleTypeChange={setNextScheduleType}
        onNextDateChange={setNextDate}
        dateAriaLabel="次回接種予定日"
      />

      {/* Remarks */}
      <div className="flex flex-col gap-1.5 flex-1 min-h-0">
        <Label className={`text-sm font-medium ${C.text60}`}>備考</Label>
        <Textarea
          value={remarks}
          onChange={(e) => setRemarks(e.target.value)}
          className={`flex-1 resize-none ${C.bgWhite} ${C.borderMedium} p-3 text-sm ${C.text} leading-relaxed`}
        />
      </div>

      {/* Save Button — React 19 useFormStatus 経由で親 <form> の pending を自動反映 */}
      <div className="pt-2">
        <SubmitButton colorVariant="primary" loadingText="登録中...">
          接種記録を追加
        </SubmitButton>
      </div>
    </div>
  );
});
