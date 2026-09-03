// External
import { memo } from "react";
import { C, ICON } from "@/lib/design-tokens";
import { Building2, Calendar, ShieldCheck } from "lucide-react";

// Internal
import { Label } from "@/components/ui/label";
import { Input } from "@/components/ui/input";
import { Checkbox } from "@/components/ui/checkbox";
import { MasterLink } from "@/components/shared/MasterLink";
import { Textarea } from "@/components/ui/textarea";
import { DatePicker } from "@/components/shared/DatePicker/DatePicker";
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group";
import { SearchableSelect } from "@/components/ui/searchable-select";
import { FormFieldError } from "@/components/shared/FormFieldError";
import { NextScheduleField, calculateNextDate } from "@/components/shared/NextScheduleField";

// Relative
import { H_STYLES } from "../lib/styles";

// Types
import type { MasterItem } from "@/types";
import type { HospitalizationFormData } from "../types";

interface HospitalizationBasicInfoProps {
  formData: HospitalizationFormData;
  onChange: (updates: Partial<HospitalizationFormData>) => void;
  cageItems: MasterItem[];
  // rerender-memo: 呼び出し側で毎レンダー新規オブジェクトを作らないよう
  // プリミティブ値で受け取る（`fieldErrors={{ cage_id: ... }}` を廃止）。
  cageIdError?: string;
}

export const HospitalizationBasicInfo = memo(function HospitalizationBasicInfo({ formData, onChange, cageItems, cageIdError }: HospitalizationBasicInfoProps) {
  return (
    <div className={`${C.bgWhite} rounded-lg border ${C.borderMedium} ${H_STYLES.padding.box}`}>
      <h2 className={`${H_STYLES.text.base} font-bold mb-3 flex items-center gap-2 ${C.text}`}>
        <Building2 className={`${ICON.action} ${C.text60}`} />
        基本情報
      </h2>
      
      {/* 入院タイプ */}
      <div className="mb-3">
        <Label className={`${H_STYLES.text.sm} ${C.text60} mb-1.5 block`}>入院タイプ</Label>
        <RadioGroup
          value={formData.hospitalizationType}
          onValueChange={(val) =>
            onChange({ hospitalizationType: val })
          }
          className="flex gap-4"
          id="hospitalization_type"
        >
          <div className="flex items-center gap-2 cursor-pointer">
            <RadioGroupItem
              value="入院"
              id="type-hospitalization"
              aria-label="入院タイプ: 入院"
              className={`${C.text}`}
            />
            <Label htmlFor="type-hospitalization" className={`inline-flex min-h-11 min-w-11 items-center ${H_STYLES.text.base} ${C.text} cursor-pointer`}>入院</Label>
          </div>
          <div className="flex items-center gap-2 cursor-pointer">
            <RadioGroupItem
              value="ホテル"
              id="type-hotel"
              aria-label="入院タイプ: ホテル"
              className={`${C.text}`}
            />
            <Label htmlFor="type-hotel" className={`inline-flex min-h-11 min-w-11 items-center ${H_STYLES.text.base} ${C.text} cursor-pointer`}>ホテル</Label>
          </div>
        </RadioGroup>
      </div>

      {/* 期間 */}
      <div className="mb-3">
        <Label className={`${H_STYLES.text.sm} ${C.text60} mb-1.5 block flex items-center gap-2`}>
          <Calendar className={ICON.action} />
          期間
        </Label>
        <div className={`flex flex-col items-stretch ${H_STYLES.gap.default}`}>
          <Label htmlFor="start_date" className="sr-only">開始日</Label>
          <DatePicker
            id="start_date"
            value={formData.displayDate}
            onChange={(v) => onChange({ displayDate: v })}
            placeholder="開始日"
            className="flex-1"
          />
          <span className={`text-center text-sm ${C.text40}`} aria-hidden="true">〜</span>
          <Label htmlFor="end_date" className="sr-only">終了日</Label>
          <DatePicker
            id="end_date"
            value={formData.endDate ?? ""}
            onChange={(v) => onChange({ endDate: v })}
            placeholder="終了日"
            className="flex-1"
          />
        </div>
      </div>

      {/* ケージ/個室（BUG-037: 必須） */}
      <div className="mb-3">
          <div className="flex items-center justify-between mb-1.5">
            <Label htmlFor="cage_id" className={`${H_STYLES.text.sm} ${C.text60}`}>
              ケージ・個室<span className="text-destructive ml-0.5" aria-hidden="true">*</span>
            </Label>
            <MasterLink category="cage" label="編集" className="text-2xs" />
          </div>
          <SearchableSelect
              id="cage_id"
              value={formData.cageId}
              onValueChange={(val) => onChange({ cageId: val })}
              options={cageItems.map((cage) => ({
                value: String(cage.id),
                label: cage.description ? `${cage.name}（${cage.description}）` : cage.name,
              }))}
              placeholder="選択してください"
              searchPlaceholder="ケージを検索..."
              ariaInvalid={Boolean(cageIdError)}
          />
          <FormFieldError message={cageIdError} />
      </div>

      {/* 保険 */}
      <div className="mb-3">
        <div className="flex items-center gap-2 mb-1.5">
          <Checkbox
            id="is_insurance"
            aria-label="保険適用"
            touchTarget
            checked={formData.isInsurance}
            onCheckedChange={(checked) =>
              onChange({
                isInsurance: checked === true,
                insuranceCompanyName: checked === true ? formData.insuranceCompanyName : "",
                insuranceNumber: checked === true ? formData.insuranceNumber : "",
              })
            }
          />
          <Label
            htmlFor="is_insurance"
            className={`${H_STYLES.text.sm} ${C.text} cursor-pointer flex items-center gap-1`}
          >
            <ShieldCheck className={`${ICON.action} ${C.text60}`} />
            保険適用
          </Label>
        </div>
        {formData.isInsurance ? (
          <div className={`flex flex-col ${H_STYLES.gap.default} pl-6`}>
            <div>
              <Label htmlFor="insurance_company_name" className={`${H_STYLES.text.xs} ${C.text60} mb-1 block`}>
                保険会社名
              </Label>
              <Input
                id="insurance_company_name"
                value={formData.insuranceCompanyName}
                onChange={(e) => onChange({ insuranceCompanyName: e.target.value })}
                placeholder="保険会社名を入力..."
                className={`h-9 ${H_STYLES.text.base} ${C.bgWhite} ${C.borderMedium}`}
                maxLength={100}
              />
            </div>
            <div>
              <Label htmlFor="insurance_number" className={`${H_STYLES.text.xs} ${C.text60} mb-1 block`}>
                保険番号
              </Label>
              <Input
                id="insurance_number"
                value={formData.insuranceNumber}
                onChange={(e) => onChange({ insuranceNumber: e.target.value })}
                placeholder="保険番号を入力..."
                className={`h-9 ${H_STYLES.text.base} ${C.bgWhite} ${C.borderMedium}`}
                maxLength={50}
              />
            </div>
          </div>
        ) : null}
      </div>

      <NextScheduleField
        typeId="hospitalization-next-schedule"
        dateId="hospitalization-next-date"
        scheduleType={formData.nextVisit ? "other" : "4weeks"}
        nextDate={formData.nextVisit}
        onScheduleTypeChange={(value) => {
          const calculated = calculateNextDate(formData.displayDate, value);
          onChange({ nextVisit: calculated || formData.nextVisit });
        }}
        onNextDateChange={(value) => onChange({ nextVisit: value })}
      />

      {/* メモ */}
      <div>
        <Label htmlFor="memo" className={`${H_STYLES.text.sm} ${C.text60} mb-1.5 block`}>メモ</Label>
        <Textarea
          id="memo"
          value={formData.memo}
          onChange={(e) => onChange({ memo: e.target.value })}
          placeholder="メモを入力..."
          className={`min-h-[80px] ${H_STYLES.text.base} resize-none ${C.bgWhite} ${C.borderMedium} ${C.focusVisibleRingActionPrimary}`}
        />
      </div>
    </div>
  );
});
