// React/Framework
import { memo } from "react";
import { useNavigate } from "react-router";

// External
import { FileText, Settings } from "lucide-react";

// Internal
import { paths } from "@/config/paths";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { CharCountTextarea } from "@/components/shared/CharCountTextarea";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { C, LAYOUT, ICON, STYLE } from "@/lib/design-tokens";

// Relative
import { useGetChiefComplaintCategories } from "../api/get-chief-complaint-categories";

interface InterviewChiefComplaintProps {
  className?: string;
  chiefComplaint: string;
  setChiefComplaint: (value: string) => void;
  chiefComplaintCategoryId: number | null;
  setChiefComplaintCategoryId: (id: number | null) => void;
  templates: { label: string; text: string }[];
  onInsertTemplate: (text: string) => void;
}

export const InterviewChiefComplaint = memo(function InterviewChiefComplaint({
  className,
  chiefComplaint,
  setChiefComplaint,
  chiefComplaintCategoryId,
  setChiefComplaintCategoryId,
  templates,
  onInsertTemplate,
}: InterviewChiefComplaintProps) {
  const navigate = useNavigate();
  const { data: categories = [], isLoading } = useGetChiefComplaintCategories();

  return (
    <div className={`flex flex-col ${className ?? ""}`}>
      <div className="pb-1.5">
        <h4 className={`text-sm font-bold ${C.text} flex items-center gap-2`}>
          <FileText className={ICON.action} />
          主訴情報
        </h4>
      </div>
      <div className="flex-1 flex flex-col gap-2 min-h-0">
        <div className="space-y-1.5">
          <div className="flex items-center justify-between">
            <Label className={`text-sm ${C.text60}`}>主訴区分</Label>
            <button
              onClick={() => navigate(paths.settings.interview.chiefComplaint.getHref())}
              className={`text-xs ${C.text40} ${C.hoverTextAccent} transition-colors flex items-center gap-1`}
              type="button"
            >
              <Settings className={ICON.xs} />
              マスタ編集
            </button>
          </div>
          <Select
            value={chiefComplaintCategoryId ? String(chiefComplaintCategoryId) : ""}
            onValueChange={(value) => setChiefComplaintCategoryId(value ? Number(value) : null)}
            disabled={isLoading}
          >
            <SelectTrigger className={`w-full ${LAYOUT.touch.md} bg-white ${C.borderMedium} text-sm ${C.text}`}>
              <SelectValue placeholder={isLoading ? "読み込み中..." : "選択してください"} />
            </SelectTrigger>
            <SelectContent>
              {categories.map((category) => (
                <SelectItem key={category.id} value={String(category.id)}>
                  {category.name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        <div className="space-y-1.5">
          <div className="flex items-center justify-between">
            <Label className={`text-sm ${C.text60}`}>定型文挿入</Label>
            <button
              onClick={() => navigate(paths.settings.interview.interviewTemplate.getHref())}
              className={`text-xs ${C.text40} ${C.hoverTextAccent} transition-colors flex items-center gap-1`}
              type="button"
            >
              <Settings className={ICON.xs} />
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
          <CharCountTextarea
            value={chiefComplaint}
            onChange={setChiefComplaint}
            className="flex-1 min-h-0"
            textareaClassName={`${STYLE.textarea} min-h-0`}
          />
        </div>
      </div>
    </div>
  );
});
