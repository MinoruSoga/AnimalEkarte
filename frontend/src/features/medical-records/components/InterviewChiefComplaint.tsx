// React/Framework
import { memo } from "react";
import { useNavigate } from "react-router";

// External
import { FileText, Settings } from "lucide-react";

// Internal
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { C, LAYOUT } from "@/lib/design-tokens";

interface InterviewChiefComplaintProps {
  chiefComplaint: string;
  setChiefComplaint: (value: string) => void;
  templates: { label: string; text: string }[];
  onInsertTemplate: (text: string) => void;
}

export const InterviewChiefComplaint = memo(function InterviewChiefComplaint({
  chiefComplaint,
  setChiefComplaint,
  templates,
  onInsertTemplate,
}: InterviewChiefComplaintProps) {
  const navigate = useNavigate();

  return (
    <Card className="flex-1 flex flex-col min-h-0 border-none shadow-none bg-transparent gap-0">
      <CardHeader className="p-0 pb-1.5">
        <CardTitle className={`text-sm font-bold ${C.text} flex items-center gap-2`}>
          <FileText className="size-4" />
          主訴情報
        </CardTitle>
      </CardHeader>
      <CardContent className="p-0 flex-1 flex flex-col gap-2 min-h-0">
        <div className="space-y-1.5">
          <div className="flex items-center justify-between">
            <Label className={`text-sm ${C.text60}`}>主訴区分</Label>
            <button
              onClick={() => navigate("/settings/interview?tab=chief_complaint")}
              className={`text-xs ${C.text40} ${C.hoverTextAccent} transition-colors flex items-center gap-1`}
              type="button"
            >
              <Settings className="size-3" />
              マスタ編集
            </button>
          </div>
          <Select>
            <SelectTrigger className={`w-full ${LAYOUT.touch.md} bg-white ${C.borderMedium} text-sm ${C.text}`}>
              <SelectValue placeholder="選択してください" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="checkup">定期検診</SelectItem>
              <SelectItem value="sick">傷病</SelectItem>
              <SelectItem value="prevention">予防</SelectItem>
            </SelectContent>
          </Select>
        </div>

        <div className="space-y-1.5">
          <div className="flex items-center justify-between">
            <Label className={`text-sm ${C.text60}`}>定型文挿入</Label>
            <button
              onClick={() => navigate("/settings/interview?tab=interview_template")}
              className={`text-xs ${C.text40} ${C.hoverTextAccent} transition-colors flex items-center gap-1`}
              type="button"
            >
              <Settings className="size-3" />
              マスタ編集
            </button>
          </div>
          <div className="flex flex-wrap gap-2">
            {templates.map((tmpl) => (
              <Button
                key={tmpl.label}
                variant="outline"
                size="sm"
                className={`${LAYOUT.touch.md} text-sm px-3 bg-white ${C.hoverBgPage} ${C.text60} ${C.borderMedium}`}
                onClick={() => onInsertTemplate(tmpl.text)}
              >
                {tmpl.label}
              </Button>
            ))}
          </div>
        </div>

        <div className="flex-1 flex flex-col gap-1.5 min-h-0">
          <Label className={`text-sm ${C.text60}`}>主訴詳細</Label>
          <Textarea
            value={chiefComplaint}
            onChange={(e) => setChiefComplaint(e.target.value)}
            className={`flex-1 resize-none bg-white ${C.borderMedium} text-sm p-3 leading-relaxed font-mono`}
          />
        </div>
      </CardContent>
    </Card>
  );
});
