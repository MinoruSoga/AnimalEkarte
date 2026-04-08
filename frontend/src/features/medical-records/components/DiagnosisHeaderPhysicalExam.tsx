// React/Framework
import { C, ICON } from "@/lib/design-tokens";
import { memo } from "react";

// External
import { Activity } from "lucide-react";

// Internal
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";

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
    <div className="col-span-4 flex flex-col min-h-0">
      <Card className="flex-1 flex flex-col min-h-0 border-none shadow-none bg-transparent">
        <CardHeader className="p-0 pb-2">
          <CardTitle className={`text-sm font-bold ${C.text} flex items-center gap-2`}>
            <Activity className={ICON.action} />
            診察(PE)
          </CardTitle>
        </CardHeader>
        <CardContent className="p-0 flex-1 flex flex-col min-h-0">
          <Label className={`text-sm ${C.text60} mb-1.5`}>方針</Label>
          <Textarea
            value={policy}
            onChange={(e) => setPolicy(e.target.value)}
            className={`flex-1 resize-none bg-white ${C.borderMedium} text-sm p-3 font-mono ${C.focusRingMedicalBlue}`}
            disabled={!canEdit}
          />
        </CardContent>
      </Card>
    </div>
  );
});
