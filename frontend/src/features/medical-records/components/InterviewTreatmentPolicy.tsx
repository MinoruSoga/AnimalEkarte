// React/Framework
import { memo } from "react";

// External
import { ChevronRight } from "lucide-react";

// Internal
import { CharCountTextarea } from "@/components/shared/CharCountTextarea";
import { C, ICON, STYLE } from "@/lib/design-tokens";
import { usePermission } from "@/hooks/use-permission";

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
  const { canEdit } = usePermission("medical-records");
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
      <CharCountTextarea
        value={treatmentPolicy}
        onChange={setTreatmentPolicy}
        className="flex-1 min-h-0"
        textareaClassName={`${STYLE.textarea} min-h-0`}
        disabled={!canEdit}
      />
    </div>
  );
});
