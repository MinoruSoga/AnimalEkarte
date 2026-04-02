// External
import { memo } from "react";
import { C, ICON } from "@/lib/design-tokens";
import { Building2, Calendar } from "lucide-react";

// Internal
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { NotionDatePicker } from "@/components/shared/NotionDatePicker/NotionDatePicker";
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";

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
    <div className={`bg-white rounded-lg shadow-sm border border-[rgba(55,53,47,0.16)] ${H_STYLES.padding.box}`}>
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
          <Label className={`${H_STYLES.text.sm} ${C.text60} mb-1.5 block`}>ケージ・個室</Label>
          <Select 
              value={formData.cageId} 
              onValueChange={(val) => onChange({ cageId: val })}
          >
              <SelectTrigger id="cage_id" className={`h-10 ${H_STYLES.text.base} bg-white border-[rgba(55,53,47,0.16)]`}>
                  <SelectValue placeholder="選択してください" />
              </SelectTrigger>
              <SelectContent>
                  {cageItems.map((cage) => (
                      <SelectItem key={cage.id} value={String(cage.id)}>
                          {cage.name} <span className={`${H_STYLES.text.xs} text-muted-foreground ml-1`}>({cage.description})</span>
                      </SelectItem>
                  ))}
              </SelectContent>
          </Select>
      </div>

      {/* メモ */}
      <div>
        <Label className={`${H_STYLES.text.sm} ${C.text60} mb-1.5 block`}>メモ</Label>
        <Textarea
          id="memo"
          value={formData.memo}
          onChange={(e) => onChange({ memo: e.target.value })}
          placeholder="メモを入力..."
          className={`min-h-[80px] ${H_STYLES.text.base} resize-none bg-white border-[rgba(55,53,47,0.16)] focus-visible:ring-[#2EAADC]`}
        />
      </div>
    </div>
  );
});
