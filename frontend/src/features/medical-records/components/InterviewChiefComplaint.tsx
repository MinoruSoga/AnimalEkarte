// React/Framework
import { memo, useMemo } from "react";
import { useNavigate } from "react-router";

// External
import { FileText, Settings } from "lucide-react";

// Internal
import { paths } from "@/config/paths";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { CharCountTextarea } from "@/components/shared/CharCountTextarea";
import { SearchableSelect, type SearchableSelectOption } from "@/components/ui/searchable-select";
import { C, LAYOUT, ICON, STYLE } from "@/lib/design-tokens";
import { usePermission } from "@/hooks/use-permission";

// Relative
import { useGetChiefComplaintTypes } from "../api/get-chief-complaint-types";

interface InterviewChiefComplaintProps {
  className?: string;
  chiefComplaint: string;
  setChiefComplaint: (value: string) => void;
  chiefComplaintTypeId: number | null;
  setChiefComplaintTypeId: (id: number | null) => void;
  templates: { label: string; text: string }[];
  onInsertTemplate: (text: string) => void;
}

export const InterviewChiefComplaint = memo(function InterviewChiefComplaint({
  className,
  chiefComplaint,
  setChiefComplaint,
  chiefComplaintTypeId,
  setChiefComplaintTypeId,
  templates,
  onInsertTemplate,
}: InterviewChiefComplaintProps) {
  const navigate = useNavigate();
  const { canEdit } = usePermission("medical-records");
  const { data: categories = [], isLoading } = useGetChiefComplaintTypes();

  // SearchableSelect 用に選択肢を {value,label} 形へ変換(参照安定のため memo 化)
  const categoryOptions = useMemo<SearchableSelectOption[]>(
    () => categories.map((category) => ({ value: String(category.id), label: category.name })),
    [categories]
  );
  const templateButtons = useMemo(
    () => templates.map((tmpl) => (
      <Button
        key={tmpl.label}
        variant="outline"
        size="sm"
        className={`${LAYOUT.touch.md} text-sm px-3 ${C.bgWhite} ${C.hoverBgPage} ${C.text60} ${C.borderMedium}`}
        onClick={() => onInsertTemplate(tmpl.text)}
        disabled={!canEdit}
      >
        {tmpl.label}
      </Button>
    )),
    [templates, onInsertTemplate, canEdit]
  );

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
            {canEdit ? (
              <button
                type="button"
                onClick={() => navigate(paths.settings.interview.chiefComplaint.getHref())}
                className={`text-xs ${C.text40} ${C.hoverTextAccent} transition-colors flex items-center gap-1`}
              >
                <Settings className={ICON.xs} />
                マスタ編集
              </button>
            ) : null}
          </div>
          <SearchableSelect
            value={chiefComplaintTypeId ? String(chiefComplaintTypeId) : ""}
            onValueChange={(value) => setChiefComplaintTypeId(value ? Number(value) : null)}
            options={categoryOptions}
            disabled={isLoading || !canEdit}
            placeholder={isLoading ? "読み込み中..." : "選択してください"}
            searchPlaceholder="主訴区分を検索..."
            className={LAYOUT.touch.md}
          />
        </div>

        <div className="space-y-1.5">
          <div className="flex items-center justify-between">
            <Label className={`text-sm ${C.text60}`}>定型文挿入</Label>
            {canEdit ? (
              <button
                type="button"
                onClick={() => navigate(paths.settings.interview.interviewTemplate.getHref())}
                className={`text-xs ${C.text40} ${C.hoverTextAccent} transition-colors flex items-center gap-1`}
              >
                <Settings className={ICON.xs} />
                マスタ編集
              </button>
            ) : null}
          </div>
          <div className="flex flex-wrap gap-2">
            {templateButtons}
          </div>
        </div>

        <div className="flex-1 flex flex-col gap-1.5 min-h-0">
          <Label className={`text-sm ${C.text60}`}>主訴詳細</Label>
          <CharCountTextarea
            value={chiefComplaint}
            onChange={setChiefComplaint}
            className="flex-1 min-h-0"
            textareaClassName={`${STYLE.textarea} min-h-0`}
            disabled={!canEdit}
          />
        </div>
      </div>
    </div>
  );
});
