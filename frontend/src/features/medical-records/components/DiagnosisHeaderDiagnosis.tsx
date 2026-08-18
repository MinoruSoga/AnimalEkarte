// React/Framework
import { C, ICON } from "@/lib/design-tokens";
import { memo, useMemo } from "react";
import { FormFieldError } from "@/components/shared/FormFieldError";

// External
import { ChevronRight } from "lucide-react";

// Internal
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { SearchableSelect, type SearchableSelectOption } from "@/components/ui/searchable-select";

// Relative
import { useGetDiagnosisTypes, useGetDiagnosisNames } from "../api/get-diagnosis-options";
import { DiagnosisHeaderSection } from "./DiagnosisHeaderSection";

type SelectedDiagnosisOption = { id: string | number; name: string } | null | undefined;

interface DiagnosisHeaderDiagnosisProps {
  diagnosisDetails: string;
  setDiagnosisDetails: (v: string) => void;
  diagnosis1CategoryId?: number | null;
  setDiagnosis1CategoryId?: (id: number | null) => void;
  diagnosis1NameId?: number | null;
  setDiagnosis1NameId?: (id: number | null) => void;
  diagnosis2CategoryId?: number | null;
  setDiagnosis2CategoryId?: (id: number | null) => void;
  diagnosis2NameId?: number | null;
  setDiagnosis2NameId?: (id: number | null) => void;
  canEdit: boolean;
  diagnosis1NameIdError?: string | null;
  selectedDiagnosisType?: SelectedDiagnosisOption;
  selectedDiagnosisName?: SelectedDiagnosisOption;
  selectedDiagnosis2Type?: SelectedDiagnosisOption;
  selectedDiagnosis2Name?: SelectedDiagnosisOption;
}

function mergeSelectedOption(
  options: SearchableSelectOption[],
  selected: SelectedDiagnosisOption,
): SearchableSelectOption[] {
  if (!selected) return options;
  const value = String(selected.id);
  if (options.some((option) => option.value === value)) return options;
  return [{ value, label: selected.name }, ...options];
}

export const DiagnosisHeaderDiagnosis = memo(function DiagnosisHeaderDiagnosis({
  diagnosisDetails,
  setDiagnosisDetails,
  diagnosis1CategoryId,
  setDiagnosis1CategoryId,
  diagnosis1NameId,
  setDiagnosis1NameId,
  diagnosis2CategoryId,
  setDiagnosis2CategoryId,
  diagnosis2NameId,
  setDiagnosis2NameId,
  canEdit,
  diagnosis1NameIdError,
  selectedDiagnosisType,
  selectedDiagnosisName,
  selectedDiagnosis2Type,
  selectedDiagnosis2Name,
}: DiagnosisHeaderDiagnosisProps) {
  const { data: categories = [], isLoading: isTypesLoading } = useGetDiagnosisTypes();
  const { data: names1 = [], isLoading: isNames1Loading } = useGetDiagnosisNames(diagnosis1CategoryId);
  const { data: names2 = [], isLoading: isNames2Loading } = useGetDiagnosisNames(diagnosis2CategoryId);

  // SearchableSelect 用に選択肢を {value,label} 形へ変換(参照安定のため memo 化)
  const categoryOptions = useMemo<SearchableSelectOption[]>(
    () => mergeSelectedOption(
      mergeSelectedOption(
        categories.map((cat) => ({ value: String(cat.id), label: cat.name })),
        selectedDiagnosisType,
      ),
      selectedDiagnosis2Type,
    ),
    [categories, selectedDiagnosisType, selectedDiagnosis2Type],
  );
  const names1Options = useMemo<SearchableSelectOption[]>(
    () => mergeSelectedOption(
      names1.map((name) => ({ value: String(name.id), label: name.name })),
      selectedDiagnosisName,
    ),
    [names1, selectedDiagnosisName],
  );
  const names2Options = useMemo<SearchableSelectOption[]>(
    () => mergeSelectedOption(
      names2.map((name) => ({ value: String(name.id), label: name.name })),
      selectedDiagnosis2Name,
    ),
    [names2, selectedDiagnosis2Name],
  );

  const controls = (
    <div className="flex flex-col gap-2">
      <div className="flex flex-col">
        <div className="flex items-center gap-2">
          <Label className={`w-10 shrink-0 text-sm font-medium ${C.text60} mb-0`}>
            診断1
          </Label>
          <SearchableSelect
            value={diagnosis1CategoryId ? String(diagnosis1CategoryId) : ""}
            onValueChange={(value) => {
              setDiagnosis1CategoryId?.(value ? Number(value) : null);
              setDiagnosis1NameId?.(null);
            }}
            options={categoryOptions}
            disabled={isTypesLoading || !canEdit}
            placeholder={isTypesLoading ? "読み込み中..." : "カテゴリを選択"}
            searchPlaceholder="カテゴリを検索..."
            className="flex-1"
          />
          <SearchableSelect
            value={diagnosis1NameId ? String(diagnosis1NameId) : ""}
            onValueChange={(value) => setDiagnosis1NameId?.(value ? Number(value) : null)}
            options={names1Options}
            disabled={isNames1Loading || !diagnosis1CategoryId || !canEdit}
            placeholder={isNames1Loading ? "読み込み中..." : "病名を選択"}
            searchPlaceholder="病名を検索..."
            className="flex-1"
            ariaInvalid={Boolean(diagnosis1NameIdError)}
          />
        </div>
        <div className="h-5 [&>p]:mt-0">
          <FormFieldError message={diagnosis1NameIdError} />
        </div>
      </div>

      <div className="flex items-center gap-2">
        <Label className={`w-10 shrink-0 text-sm font-medium ${C.text60} mb-0`}>
          診断2
        </Label>
        <SearchableSelect
          value={diagnosis2CategoryId ? String(diagnosis2CategoryId) : ""}
          onValueChange={(value) => {
            setDiagnosis2CategoryId?.(value ? Number(value) : null);
            setDiagnosis2NameId?.(null);
          }}
          options={categoryOptions}
          disabled={isTypesLoading || !canEdit}
          placeholder={isTypesLoading ? "読み込み中..." : "カテゴリを選択"}
          searchPlaceholder="カテゴリを検索..."
          className="flex-1"
        />
        <SearchableSelect
          value={diagnosis2NameId ? String(diagnosis2NameId) : ""}
          onValueChange={(value) => setDiagnosis2NameId?.(value ? Number(value) : null)}
          options={names2Options}
          disabled={isNames2Loading || !diagnosis2CategoryId || !canEdit}
          placeholder={isNames2Loading ? "読み込み中..." : "病名を選択"}
          searchPlaceholder="病名を検索..."
          className="flex-1"
        />
      </div>
    </div>
  );

  return (
    <DiagnosisHeaderSection
      className="col-span-5"
      icon={<ChevronRight className={ICON.action} />}
      title="診断"
      controls={controls}
    >
      <Textarea
        value={diagnosisDetails}
        onChange={(e) => setDiagnosisDetails(e.target.value)}
        aria-label="診断詳細"
        className={`h-full min-h-0 resize-none rounded-md border ${C.bgWhite} ${C.borderMedium} text-sm p-3 font-mono ${C.focusVisibleRingActionPrimary}`}
        disabled={!canEdit}
      />
    </DiagnosisHeaderSection>
  );
});
