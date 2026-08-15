// React/Framework
import { C, ICON } from "@/lib/design-tokens";
import { memo } from "react";

// External
import { FileText } from "lucide-react";

// Internal
import { ScrollArea } from "@/components/ui/scroll-area";

// Relative
import { DiagnosisHeaderSection } from "./DiagnosisHeaderSection";

interface DiagnosisHeaderChiefComplaintProps {
  content?: string;
}

export const DiagnosisHeaderChiefComplaint = memo(function DiagnosisHeaderChiefComplaint({
  content,
}: DiagnosisHeaderChiefComplaintProps) {
  return (
    <DiagnosisHeaderSection
      className="col-span-3"
      icon={<FileText className={ICON.action} />}
      title="問診・主訴"
    >
      <ScrollArea
        className={`h-full min-h-0 rounded-md border ${C.borderMedium} ${C.bgPage} p-3 text-sm ${C.text}`}
      >
        {content ? (
          <div className="whitespace-pre-wrap leading-relaxed font-mono">
            {content}
          </div>
        ) : (
          <div className={`flex items-center justify-center h-full ${C.text40} italic`}>
            主訴の入力はありません
          </div>
        )}
      </ScrollArea>
    </DiagnosisHeaderSection>
  );
});
