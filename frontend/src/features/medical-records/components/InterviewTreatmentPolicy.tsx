// React/Framework
import { memo } from "react";

// External
import { ChevronRight } from "lucide-react";

// Internal
import { Textarea } from "@/components/ui/textarea";
import { C, ICON } from "@/lib/design-tokens";

interface InterviewTreatmentPolicyProps {
  className?: string;
  treatmentPolicy: string;
  setTreatmentPolicy: (value: string) => void;
}

export const InterviewTreatmentPolicy = memo(function InterviewTreatmentPolicy({
  className,
  treatmentPolicy,
  setTreatmentPolicy,
}: InterviewTreatmentPolicyProps) {
  return (
    <div className={`flex flex-col ${className ?? ""} h-full`}>
      <div className="pb-1.5 shrink-0">
        <h4 className={`text-sm font-bold ${C.text} flex items-center gap-2`}>
          <ChevronRight className={ICON.action} />
          治療方針
          <span className={`text-sm font-normal ${C.text60} ml-auto`}>
            (次工程へ連携)
          </span>
        </h4>
      </div>
      <Textarea
        value={treatmentPolicy}
        onChange={(e) => setTreatmentPolicy(e.target.value)}
        className={`flex-1 resize-none bg-white ${C.borderMedium} text-sm p-3 leading-relaxed font-mono min-h-0`}
      />
    </div>
  );
});
