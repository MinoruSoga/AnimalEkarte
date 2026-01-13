import { Building2, Calendar } from "lucide-react";
import { Label } from "../../../components/ui/label";
import { Input } from "../../../components/ui/input";
import { Textarea } from "../../../components/ui/textarea";
import { RadioGroup, RadioGroupItem } from "../../../components/ui/radio-group";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "../../../components/ui/select";
import { MasterItem } from "../../../types";
import { HospitalizationFormData } from "../types";
import { H_STYLES } from "../styles";

interface HospitalizationBasicInfoProps {
  formData: HospitalizationFormData;
  onChange: (updates: Partial<HospitalizationFormData>) => void;
  cageItems: MasterItem[];
}

export const HospitalizationBasicInfo = ({ formData, onChange, cageItems }: HospitalizationBasicInfoProps) => {
  return (
    <div className={`bg-white rounded-lg shadow-sm border border-[rgba(55,53,47,0.16)] ${H_STYLES.padding.box}`}>
      <h2 className={`${H_STYLES.text.base} font-bold mb-3 flex items-center gap-2 text-[#37352F]`}>
        <Building2 className="h-3.5 w-3.5 text-[#37352F]/60" />
        基本情報
      </h2>
      
      {/* 入院タイプ */}
      <div className="mb-3">
        <Label className={`${H_STYLES.text.sm} text-[#37352F]/60 mb-1.5 block`}>入院タイプ</Label>
        <RadioGroup
          value={formData.hospitalizationType}
          onValueChange={(val) =>
            onChange({ hospitalizationType: val })
          }
          className="flex gap-4"
        >
          <div className="flex items-center gap-2 cursor-pointer">
            <RadioGroupItem
              value="入院"
              id="type-hospitalization"
              className="text-[#37352F]"
            />
            <Label htmlFor="type-hospitalization" className={`${H_STYLES.text.base} text-[#37352F] cursor-pointer`}>入院</Label>
          </div>
          <div className="flex items-center gap-2 cursor-pointer">
            <RadioGroupItem
              value="ホテル"
              id="type-hotel"
              className="text-[#37352F]"
            />
            <Label htmlFor="type-hotel" className={`${H_STYLES.text.base} text-[#37352F] cursor-pointer`}>ホテル</Label>
          </div>
        </RadioGroup>
      </div>

      {/* 期間 */}
      <div className="mb-3">
        <Label className={`${H_STYLES.text.sm} text-[#37352F]/60 mb-1.5 block flex items-center gap-2`}>
          <Calendar className="h-4 w-4" />
          期間
        </Label>
        <div className={`flex items-center ${H_STYLES.gap.default}`}>
          <Input
            type="date"
            value={formData.displayDate}
            onChange={(e) => onChange({ displayDate: e.target.value })}
            className={`flex-1 h-10 ${H_STYLES.text.base} bg-white border-[rgba(55,53,47,0.16)] focus-visible:ring-[#2EAADC]`}
          />
          <span className="text-[#37352F]/40 text-sm">〜</span>
          <Input
            type="date"
            className={`flex-1 h-10 ${H_STYLES.text.base} bg-white border-[rgba(55,53,47,0.16)] focus-visible:ring-[#2EAADC]`}
          />
        </div>
      </div>

      {/* ケージ/個室 */}
      <div className="mb-3">
          <Label className={`${H_STYLES.text.sm} text-[#37352F]/60 mb-1.5 block`}>ケージ・個室</Label>
          <Select 
              value={formData.cageId} 
              onValueChange={(val) => onChange({ cageId: val })}
          >
              <SelectTrigger className={`h-10 ${H_STYLES.text.base} bg-white border-[rgba(55,53,47,0.16)]`}>
                  <SelectValue placeholder="選択してください" />
              </SelectTrigger>
              <SelectContent>
                  {cageItems.map((cage) => (
                      <SelectItem key={cage.id} value={cage.id}>
                          {cage.name} <span className={`${H_STYLES.text.xs} text-muted-foreground ml-1`}>({cage.description})</span>
                      </SelectItem>
                  ))}
              </SelectContent>
          </Select>
      </div>

      {/* メモ */}
      <div>
        <Label className={`${H_STYLES.text.sm} text-[#37352F]/60 mb-1.5 block`}>メモ</Label>
        <Textarea
          value={formData.memo}
          onChange={(e) => onChange({ memo: e.target.value })}
          placeholder="メモを入力..."
          className={`min-h-[80px] ${H_STYLES.text.base} resize-none bg-white border-[rgba(55,53,47,0.16)] focus-visible:ring-[#2EAADC]`}
        />
      </div>
    </div>
  );
};
