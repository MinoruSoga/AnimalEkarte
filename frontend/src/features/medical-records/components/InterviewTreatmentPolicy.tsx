// React/Framework
import { memo } from "react";

// External
import { ChevronRight } from "lucide-react";

// Internal
import { Textarea } from "@/components/ui/textarea";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { C } from "@/lib/design-tokens";

interface InterviewTreatmentPolicyProps {
  treatmentPolicy: string;
  setTreatmentPolicy: (value: string) => void;
}

export const InterviewTreatmentPolicy = memo(function InterviewTreatmentPolicy({
  treatmentPolicy,
  setTreatmentPolicy,
}: InterviewTreatmentPolicyProps) {
  return (
    <Card className="flex-1 flex flex-col min-h-0 border-none shadow-none bg-transparent gap-0">
      <CardHeader className="p-0 pb-1.5">
        <CardTitle className={`text-sm font-bold ${C.text} flex items-center gap-2`}>
          <ChevronRight className="size-4" />
          治療方針
          <span className={`text-sm font-normal ${C.text60} ml-auto`}>
            (次工程へ連携)
          </span>
        </CardTitle>
      </CardHeader>
      <CardContent className="p-0 flex-1 flex flex-col gap-2 min-h-0">
        <div className="flex-1 flex flex-col gap-1 min-h-0">
          <Textarea
            value={treatmentPolicy}
            onChange={(e) => setTreatmentPolicy(e.target.value)}
            className={`flex-1 resize-none bg-white ${C.borderMedium} text-sm p-3 leading-relaxed font-mono`}
          />
        </div>
      </CardContent>
    </Card>
  );
});
