import React from "react";
import { Card, CardContent, CardHeader, CardTitle } from "../../../components/ui/card";
import { Label } from "../../../components/ui/label";
import { Textarea } from "../../../components/ui/textarea";
import { Activity } from "lucide-react";

interface DiagnosisHeaderPhysicalExamProps {
  policy: string;
  setPolicy: (v: string) => void;
}

export const DiagnosisHeaderPhysicalExam = React.memo(function DiagnosisHeaderPhysicalExam({
  policy,
  setPolicy,
}: DiagnosisHeaderPhysicalExamProps) {
  return (
    <div className="col-span-4 flex flex-col min-h-0">
      <Card className="flex-1 flex flex-col min-h-0 border-none shadow-none bg-transparent">
        <CardHeader className="p-0 pb-2">
          <CardTitle className="text-sm font-bold text-[#37352F] flex items-center gap-2">
            <Activity className="size-4" />
            診察(PE)
          </CardTitle>
        </CardHeader>
        <CardContent className="p-0 flex-1 flex flex-col min-h-0">
          <Label className="text-sm text-[#37352F]/60 mb-1.5">方針</Label>
          <Textarea
            value={policy}
            onChange={(e) => setPolicy(e.target.value)}
            className="flex-1 resize-none bg-white border-[rgba(55,53,47,0.16)] text-sm p-3 font-mono focus-visible:ring-[#2EAADC]"
          />
        </CardContent>
      </Card>
    </div>
  );
});
