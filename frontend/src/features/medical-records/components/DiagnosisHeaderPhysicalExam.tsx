// React/Framework
import { C, ICON } from "@/lib/design-tokens";
import { memo } from "react";

// External
import { Activity } from "lucide-react";

// Internal
import { Textarea } from "@/components/ui/textarea";

// Relative
import { DiagnosisHeaderSection } from "./DiagnosisHeaderSection";

interface DiagnosisHeaderPhysicalExamProps {
  policy: string;
  setPolicy: (v: string) => void;
  canEdit: boolean;
}

export const DiagnosisHeaderPhysicalExam = memo(function DiagnosisHeaderPhysicalExam({
  policy,
  setPolicy,
  canEdit,
}: DiagnosisHeaderPhysicalExamProps) {
  return (
    <DiagnosisHeaderSection
      className="col-span-4"
      icon={<Activity className={ICON.action} />}
      title="診察(PE)"
    >
      <Textarea
        value={policy}
        onChange={(e) => setPolicy(e.target.value)}
        aria-label="診察所見・方針"
        className={`h-full min-h-0 resize-none rounded-md border ${C.bgWhite} ${C.borderMedium} text-sm p-3 font-mono ${C.focusVisibleRingActionPrimary}`}
        disabled={!canEdit}
      />
    </DiagnosisHeaderSection>
  );
});
