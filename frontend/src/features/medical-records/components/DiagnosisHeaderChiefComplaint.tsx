// React/Framework
import { C, ICON } from "@/lib/design-tokens";
import { memo } from "react";

// External
import { FileText } from "lucide-react";

// Internal
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { ScrollArea } from "@/components/ui/scroll-area";

interface DiagnosisHeaderChiefComplaintProps {
  content?: string;
}

export const DiagnosisHeaderChiefComplaint = memo(function DiagnosisHeaderChiefComplaint({ 
  content 
}: DiagnosisHeaderChiefComplaintProps) {
  return (
    <div className="col-span-3 flex flex-col min-h-0">
      <Card className="flex-1 flex flex-col min-h-0 border-none shadow-none bg-transparent">
        <CardHeader className="p-0 pb-2">
          <CardTitle className={`text-sm font-bold ${C.text} flex items-center gap-2`}>
            <FileText className={ICON.action} />
            問診・主訴
          </CardTitle>
        </CardHeader>
        <CardContent className="p-0 flex-1 min-h-0">
          <ScrollArea className={`h-full rounded-md border ${C.borderMedium} ${C.bgPage} p-3 text-sm ${C.text}`}>
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
        </CardContent>
      </Card>
    </div>
  );
});
