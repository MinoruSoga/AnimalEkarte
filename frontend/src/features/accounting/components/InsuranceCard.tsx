import { memo } from "react";
import { CreditCard } from "lucide-react";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Switch } from "@/components/ui/switch";
import { C, ICON } from "@/lib/design-tokens";

interface InsuranceCardProps {
  useInsurance: boolean;
  onUseInsuranceChange: (v: boolean) => void;
  insuranceRatio: string;
  onInsuranceRatioChange: (v: string) => void;
  insuranceAmount: number;
  /** EMR-63: 既存会計の編集時、未送信の保険フィールドが既存値を保持することを明示する */
  showPreservedNote?: boolean;
}

const NEW_INSURANCE_RATIO_ITEMS = [
  { value: "0.5", label: "50%" },
  { value: "0.7", label: "70%" },
] as const;

const LEGACY_INSURANCE_RATIO_LABELS: Record<string, string> = {
  "0.9": "90%",
  "1.0": "100%",
  "1": "100%",
};

function insuranceRatioSelectItems(currentValue: string) {
  const items = NEW_INSURANCE_RATIO_ITEMS.map((item) => (
    <SelectItem key={item.value} value={item.value}>
      {item.label}
    </SelectItem>
  ));
  const isNewChoice = NEW_INSURANCE_RATIO_ITEMS.some((item) => item.value === currentValue);
  if (!isNewChoice) {
    const label = LEGACY_INSURANCE_RATIO_LABELS[currentValue];
    if (label !== undefined) {
      items.push(
        <SelectItem key={currentValue} value={currentValue}>
          {label}
        </SelectItem>,
      );
    }
  }
  return items;
}

export const InsuranceCard = memo(function InsuranceCard({
  useInsurance,
  onUseInsuranceChange,
  insuranceRatio,
  onInsuranceRatioChange,
  insuranceAmount,
  showPreservedNote = false,
}: InsuranceCardProps) {
  return (
    <Card>
      <CardHeader className={`py-3 px-4 ${C.bgPage} border-b`}>
        <div className="flex items-center justify-between">
          <CardTitle className="text-sm font-medium flex items-center gap-2">
            <CreditCard className={ICON.action} /> ペット保険（窓口精算）
          </CardTitle>
          {/* form action 完了時に React が form.reset() を実行し、Radix は mount 時点の
              checked を初期値として復元する（initialCheckedStateRef）。既存会計では mount 時
              hasInsurance=false → 同期 effect で true になるため、key で remount して
              reset の復元先を常に現在値へ揃える。なければ確定済み修正の確認ダイアログ後に
              保険が false へ巻き戻り OFF として送信される（S15 手順6 で実測）。 */}
          <Switch
            key={useInsurance ? "on" : "off"}
            checked={useInsurance}
            onCheckedChange={onUseInsuranceChange}
            aria-label="ペット保険を利用"
          />
        </div>
      </CardHeader>
      {useInsurance ? (
        <CardContent className="p-4 space-y-4">
          <div className="space-y-2">
            <Label className="text-xs">負担割合（保険会社が支払う割合）</Label>
            {/* Switch と同じ reset 巻き戻し対策: Select も mount 時 value を初期値として
                form.reset() で復元する（initialValueRef）。比率変更後の submit で
                mount 値へ戻らないよう key で remount する。 */}
            <Select
              key={insuranceRatio}
              value={insuranceRatio}
              onValueChange={onInsuranceRatioChange}
            >
              <SelectTrigger className="h-11" aria-label="負担割合">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>{insuranceRatioSelectItems(insuranceRatio)}</SelectContent>
            </Select>
          </div>
          <div
            className={`flex justify-between items-center text-sm font-medium ${C.textStatusGreen} ${C.bgStatusGreen} p-2 rounded`}
          >
            <span>保険負担額</span>
            {/* EMR-62: insuranceAmount は正の magnitude で保存・表示する */}
            <span>{Math.abs(insuranceAmount).toLocaleString()} 円</span>
          </div>
          {showPreservedNote ? (
            <p className={`text-xs ${C.text60}`}>保険情報は既存の値が保持されます</p>
          ) : null}
        </CardContent>
      ) : null}
    </Card>
  );
});
