import React from "react";
import { Card, CardContent, CardHeader, CardTitle } from "../../../components/ui/card";
import { Label } from "../../../components/ui/label";
import { Textarea } from "../../../components/ui/textarea";
import { ChevronRight } from "lucide-react";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "../../../components/ui/select";

interface DiagnosisHeaderDiagnosisProps {
  diagnosisDetails: string;
  setDiagnosisDetails: (v: string) => void;
}

export const DiagnosisHeaderDiagnosis = React.memo(function DiagnosisHeaderDiagnosis({
  diagnosisDetails,
  setDiagnosisDetails,
}: DiagnosisHeaderDiagnosisProps) {
  return (
    <div className="col-span-5 flex flex-col min-h-0">
      <Card className="flex-1 flex flex-col min-h-0 border-none shadow-none bg-transparent">
        <CardHeader className="p-0 pb-2">
          <CardTitle className="text-sm font-bold text-[#37352F] flex items-center gap-2">
            <ChevronRight className="size-4" />
            診断
          </CardTitle>
        </CardHeader>
        <CardContent className="p-0 flex-1 flex flex-col gap-2 min-h-0">
          <div className="space-y-2">
            <div className="flex items-center gap-2">
              <Label className="w-10 shrink-0 text-sm font-medium text-[#37352F]/60">
                診断1
              </Label>
              <Select defaultValue="digestive">
                <SelectTrigger className="flex-1 bg-white border-[rgba(55,53,47,0.16)] h-10 text-sm">
                  <SelectValue placeholder="選択してください" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="digestive">消化器疾患</SelectItem>
                </SelectContent>
              </Select>
              <Select defaultValue="esophagitis">
                <SelectTrigger className="flex-1 bg-white border-[rgba(55,53,47,0.16)] h-10 text-sm">
                  <SelectValue placeholder="選択してください" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="esophagitis">食道炎</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div className="flex items-center gap-2">
              <Label className="w-10 shrink-0 text-sm font-medium text-[#37352F]/60">
                診断2
              </Label>
              <Select defaultValue="unselected">
                <SelectTrigger className="flex-1 bg-white border-[rgba(55,53,47,0.16)] h-10 text-sm">
                  <SelectValue placeholder="選択してください" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="unselected">未選択</SelectItem>
                </SelectContent>
              </Select>
              <Select disabled>
                <SelectTrigger className="flex-1 bg-gray-50 border-[rgba(55,53,47,0.16)] h-10 text-sm">
                  <SelectValue placeholder="" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="none">なし</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>

          <div className="flex-1 flex flex-col min-h-0">
            <Label className="text-sm text-[#37352F]/60 mb-1.5">
              診断詳細
            </Label>
            <Textarea
              value={diagnosisDetails}
              onChange={(e) => setDiagnosisDetails(e.target.value)}
              className="flex-1 resize-none bg-white border-[rgba(55,53,47,0.16)] text-sm p-3 font-mono focus-visible:ring-[#2EAADC]"
            />
          </div>
        </CardContent>
      </Card>
    </div>
  );
});
