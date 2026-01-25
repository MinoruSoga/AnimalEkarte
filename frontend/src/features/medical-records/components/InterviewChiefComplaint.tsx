// React/Framework
import React from "react";

// External
import { FileText } from "lucide-react";

// Internal
import { Button } from "../../../components/ui/button";
import { Label } from "../../../components/ui/label";
import { Textarea } from "../../../components/ui/textarea";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "../../../components/ui/select";
import { Card, CardContent, CardHeader, CardTitle } from "../../../components/ui/card";

interface InterviewChiefComplaintProps {
  chiefComplaint: string;
  setChiefComplaint: (value: string) => void;
  templates: { label: string; text: string }[];
  onInsertTemplate: (text: string) => void;
}

export const InterviewChiefComplaint = React.memo(function InterviewChiefComplaint({
  chiefComplaint,
  setChiefComplaint,
  templates,
  onInsertTemplate,
}: InterviewChiefComplaintProps) {
  return (
    <Card className="flex-1 flex flex-col min-h-0 border-none shadow-none bg-transparent">
      <CardHeader className="p-0 pb-2">
        <CardTitle className="text-sm font-bold text-[#37352F] flex items-center gap-2">
          <FileText className="size-4" />
          主訴情報
        </CardTitle>
      </CardHeader>
      <CardContent className="p-0 flex-1 flex flex-col gap-2 min-h-0">
        <div className="space-y-1.5">
          <Label className="text-sm text-[#37352F]/60">主訴区分</Label>
          <Select>
            <SelectTrigger className="w-full h-10 bg-white border-[rgba(55,53,47,0.16)] text-sm text-[#37352F]">
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
          <Label className="text-sm text-[#37352F]/60">定型文挿入</Label>
          <div className="flex flex-wrap gap-2">
            {templates.map((tmpl) => (
              <Button
                key={tmpl.label}
                variant="outline"
                size="sm"
                className="h-10 text-sm px-3 bg-white hover:bg-slate-50 text-slate-600 border-[rgba(55,53,47,0.16)]"
                onClick={() => onInsertTemplate(tmpl.text)}
              >
                {tmpl.label}
              </Button>
            ))}
          </div>
        </div>

        <div className="flex-1 flex flex-col gap-1.5 min-h-0">
          <Label className="text-sm text-[#37352F]/60">主訴詳細</Label>
          <Textarea
            value={chiefComplaint}
            onChange={(e) => setChiefComplaint(e.target.value)}
            className="flex-1 resize-none bg-white border-[rgba(55,53,47,0.16)] focus-visible:ring-[#2EAADC] text-sm p-3 leading-relaxed font-mono"
          />
        </div>
      </CardContent>
    </Card>
  );
});
