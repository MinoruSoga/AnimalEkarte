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
  /** clinical_plan.physical_exam の単一 owner（BUG-010: 旧 policy/plan 誤配線を修正） */
  physicalExam: string;
  setPhysicalExam: (v: string) => void;
  canEdit: boolean;
}

export const DiagnosisHeaderPhysicalExam = memo(function DiagnosisHeaderPhysicalExam({
  physicalExam,
  setPhysicalExam,
  canEdit,
}: DiagnosisHeaderPhysicalExamProps) {
  return (
    <DiagnosisHeaderSection
      className="col-span-4"
      icon={<Activity className={ICON.action} />}
      title="診察(PE)"
    >
      <Textarea
        value={physicalExam}
        onChange={(e) => setPhysicalExam(e.target.value)}
        aria-label="身体検査所見"
        className={`h-full min-h-0 resize-none rounded-md border ${C.bgWhite} ${C.borderMedium} text-sm p-3 font-mono ${C.focusVisibleRingActionPrimary}`}
        disabled={!canEdit}
      />
    </DiagnosisHeaderSection>
  );
});
