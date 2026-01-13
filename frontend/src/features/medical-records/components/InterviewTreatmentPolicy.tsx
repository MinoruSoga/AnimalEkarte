import React from "react";
import { Textarea } from "../../../components/ui/textarea";
import { ChevronRight } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "../../../components/ui/card";

interface InterviewTreatmentPolicyProps {
  treatmentPolicy: string;
  setTreatmentPolicy: (value: string) => void;
}

export const InterviewTreatmentPolicy = React.memo(function InterviewTreatmentPolicy({
  treatmentPolicy,
  setTreatmentPolicy,
}: InterviewTreatmentPolicyProps) {
  return (
    <Card className="flex-1 flex flex-col min-h-0 border-none shadow-none bg-transparent">
      <CardHeader className="p-0 pb-2">
        <CardTitle className="text-sm font-bold text-[#37352F] flex items-center gap-2">
          <ChevronRight className="size-4" />
          治療方針
          <span className="text-sm font-normal text-muted-foreground ml-auto">
            (次工程へ連携)
          </span>
        </CardTitle>
      </CardHeader>
      <CardContent className="p-0 flex-1 flex flex-col gap-2 min-h-0">
        <div className="flex-1 flex flex-col gap-1 min-h-0">
          <Textarea
            value={treatmentPolicy}
            onChange={(e) => setTreatmentPolicy(e.target.value)}
            className="flex-1 resize-none bg-white border-[rgba(55,53,47,0.16)] focus-visible:ring-[#2EAADC] text-sm p-3 leading-relaxed font-mono"
          />
        </div>
      </CardContent>
    </Card>
  );
});
