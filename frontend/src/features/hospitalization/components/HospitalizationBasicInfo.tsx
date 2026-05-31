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
import { NotionDatePicker } from "@/components/shared/NotionDatePicker/NotionDatePicker";
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group";
import { SearchableSelect } from "@/components/ui/searchable-select";

// Relative
import { H_STYLES } from "../styles";

// Types
import type { MasterItem } from "@/types";
import type { HospitalizationFormData } from "../types";

interface HospitalizationBasicInfoProps {
  formData: HospitalizationFormData;
  onChange: (updates: Partial<HospitalizationFormData>) => void;
  cageItems: MasterItem[];
}

export const HospitalizationBasicInfo = memo(function HospitalizationBasicInfo({ formData, onChange, cageItems }: HospitalizationBasicInfoProps) {
  return (
    <div className={`${C.bgWhite} rounded-lg shadow-sm border ${C.borderMedium} ${H_STYLES.padding.box}`}>
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
              className={`${C.text}`}
            />
            <Label htmlFor="type-hospitalization" className={`${H_STYLES.text.base} ${C.text} cursor-pointer`}>入院</Label>
          </div>
          <div className="flex items-center gap-2 cursor-pointer">
            <RadioGroupItem
              value="ホテル"
              id="type-hotel"
              className={`${C.text}`}
            />
            <Label htmlFor="type-hotel" className={`${H_STYLES.text.base} ${C.text} cursor-pointer`}>ホテル</Label>
          </div>
        </RadioGroup>
      </div>

      {/* 期間 */}
      <div className="mb-3">
        <Label className={`${H_STYLES.text.sm} ${C.text60} mb-1.5 block flex items-center gap-2`}>
          <Calendar className={ICON.action} />
          期間
        </Label>
        <div className={`flex items-center ${H_STYLES.gap.default}`}>
          <NotionDatePicker
            id="start_date"
            value={formData.displayDate}
            onChange={(v) => onChange({ displayDate: v })}
            placeholder="開始日"
            className="flex-1"
          />
          <span className={`${C.text40} text-sm`}>〜</span>
          <NotionDatePicker
            id="end_date"
            value={formData.endDate ?? ""}
            onChange={(v) => onChange({ endDate: v })}
            placeholder="終了日"
            className="flex-1"
          />
        </div>
      </div>

      {/* ケージ/個室 */}
      <div className="mb-3">
          <div className="flex items-center justify-between mb-1.5">
            <Label className={`${H_STYLES.text.sm} ${C.text60}`}>ケージ・個室</Label>
            <MasterLink category="cage" label="編集" className="text-[11px]" />
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
          />
      </div>

      {/* 保険 */}
      <div className="mb-3">
        <div className="flex items-center gap-2 mb-1.5">
          <Checkbox
            id="is_insurance"
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

      {/* メモ */}
      <div>
        <Label className={`${H_STYLES.text.sm} ${C.text60} mb-1.5 block`}>メモ</Label>
        <Textarea
          id="memo"
          value={formData.memo}
          onChange={(e) => onChange({ memo: e.target.value })}
          placeholder="メモを入力..."
          className={`min-h-[80px] ${H_STYLES.text.base} resize-none ${C.bgWhite} ${C.borderMedium} ${C.focusRingMedicalBlue}`}
        />
      </div>
    </div>
  );
});
