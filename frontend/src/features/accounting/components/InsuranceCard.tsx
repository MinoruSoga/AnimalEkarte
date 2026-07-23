import { memo } from "react";
import { CreditCard } from "lucide-react";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Switch } from "@/components/ui/switch";
import { C, ICON } from "@/lib/design-tokens";

interface InsuranceCardProps {
  useInsurance: boolean;
  onUseInsuranceChange: (v: boolean) => void;
  insuranceRatio: string;
  onInsuranceRatioChange: (v: string) => void;
  insuranceAmount: number;
}

const INSURANCE_RATIO_ITEMS = (
  <>
    <SelectItem value="0.5">50%</SelectItem>
    <SelectItem value="0.7">70%</SelectItem>
    <SelectItem value="0.9">90%</SelectItem>
    <SelectItem value="1.0">100%</SelectItem>
  </>
);

export const InsuranceCard = memo(function InsuranceCard({
  useInsurance,
  onUseInsuranceChange,
  insuranceRatio,
  onInsuranceRatioChange,
  insuranceAmount,
}: InsuranceCardProps) {
  return (
    <Card>
      <CardHeader className={`py-3 px-4 ${C.bgPage} border-b`}>
        <div className="flex items-center justify-between">
          <CardTitle className="text-sm font-medium flex items-center gap-2">
            <CreditCard className={ICON.action} /> ペット保険（窓口精算）
          </CardTitle>
          <Switch
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
            <Select value={insuranceRatio} onValueChange={onInsuranceRatioChange}>
              <SelectTrigger className="h-11" aria-label="負担割合">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>{INSURANCE_RATIO_ITEMS}</SelectContent>
            </Select>
          </div>
          <div className={`flex justify-between items-center text-sm font-medium ${C.textStatusGreen} ${C.bgStatusGreen} p-2 rounded`}>
            <span>保険負担額（マイナス）</span>
            <span>{insuranceAmount.toLocaleString()} 円</span>
          </div>
        </CardContent>
      ) : null}
    </Card>
  );
});
