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
  /** BUG-035: 確定済みは権限があっても content attribute で disabled（fieldset 継承だけに依存しない） */
  isFinalized?: boolean;
}

export const InterviewChiefComplaint = memo(function InterviewChiefComplaint({
  className,
  chiefComplaint,
  setChiefComplaint,
  chiefComplaintTypeId,
  setChiefComplaintTypeId,
  templates,
  onInsertTemplate,
  isFinalized = false,
}: InterviewChiefComplaintProps) {
  const navigate = useNavigate();
  const { canEdit } = usePermission("medical-records");
  const fieldsDisabled = !canEdit || isFinalized;
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
        disabled={fieldsDisabled}
      >
        {tmpl.label}
      </Button>
    )),
    [templates, onInsertTemplate, fieldsDisabled]
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
            <Label htmlFor="medical-record-chief-complaint-type" className={`text-sm ${C.text60}`}>主訴区分</Label>
            {!fieldsDisabled ? (
              <button
                type="button"
                onClick={() => navigate(paths.settings.interview.chiefComplaint.getHref())}
                className={`flex min-h-11 min-w-11 items-center gap-1 text-xs ${C.text40} ${C.hoverTextBrand} transition-colors`}
              >
                <Settings className={ICON.xs} />
                マスタ編集
              </button>
            ) : null}
          </div>
          <SearchableSelect
            id="medical-record-chief-complaint-type"
            value={chiefComplaintTypeId ? String(chiefComplaintTypeId) : ""}
            onValueChange={(value) => setChiefComplaintTypeId(value ? Number(value) : null)}
            options={categoryOptions}
            disabled={isLoading || fieldsDisabled}
            placeholder={isLoading ? "読み込み中..." : "選択してください"}
            searchPlaceholder="主訴区分を検索..."
            className={LAYOUT.touch.md}
          />
        </div>

        <div className="space-y-1.5">
          <div className="flex items-center justify-between">
            <span className={`mb-2 flex items-center gap-2 text-sm leading-none select-none ${C.text60}`}>
              定型文挿入
            </span>
            {!fieldsDisabled ? (
              <button
                type="button"
                onClick={() => navigate(paths.settings.interview.interviewTemplate.getHref())}
                className={`flex min-h-11 min-w-11 items-center gap-1 text-xs ${C.text40} ${C.hoverTextBrand} transition-colors`}
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
          <Label htmlFor="medical-record-chief-complaint" className={`text-sm ${C.text60}`}>主訴詳細</Label>
          <CharCountTextarea
            id="medical-record-chief-complaint"
            name="chiefComplaint"
            value={chiefComplaint}
            onChange={setChiefComplaint}
            className="flex-1 min-h-0"
            textareaClassName={`${STYLE.textarea} min-h-0`}
            disabled={fieldsDisabled}
          />
        </div>
      </div>
    </div>
  );
});
