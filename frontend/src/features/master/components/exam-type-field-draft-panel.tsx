import type { AnimalSpecies } from "../api/animal-species";
import { C } from "@/lib/design-tokens";

import { QUALITATIVE_VALUES, type ReferenceRangeDraft } from "./exam-type-fields-editor-model";
import { FieldInput, ReferenceRangeInputs } from "./exam-type-field-editors";
import { useExamTypeFieldSession } from "./use-exam-type-field-session";
import type { ExaminationTypeField } from "../api/exam-types-master";

interface ExamTypeFieldDraft {
  name: string;
  inspectionValue: string;
  normalValue: string;
  unit: string;
}

interface ExamTypeFieldDraftFormProps {
  editingId: string | "new";
  fieldDraft: ExamTypeFieldDraft;
  error: string;
  onFieldChange: (patch: Partial<ExamTypeFieldDraft>) => void;
  onCancel: () => void;
  onSaveField: () => void;
}

export function ExamTypeFieldDraftForm({
  editingId,
  fieldDraft,
  error,
  onFieldChange,
  onCancel,
  onSaveField,
}: ExamTypeFieldDraftFormProps) {
  return (
    <>
      <h4 className={`text-sm font-medium ${C.text}`}>
        {editingId === "new" ? "検査項目を追加" : "検査項目を編集"}
      </h4>
      <FieldInput
        label="検査項目名"
        value={fieldDraft.name}
        onChange={(value) => onFieldChange({ name: value })}
      />
      <FieldInput
        label="単位"
        value={fieldDraft.unit}
        onChange={(value) => onFieldChange({ unit: value })}
      />
      <FieldInput
        label="検査値"
        value={fieldDraft.inspectionValue}
        onChange={(value) => onFieldChange({ inspectionValue: value })}
      />
      <FieldInput
        label="正常値"
        value={fieldDraft.normalValue}
        onChange={(value) => onFieldChange({ normalValue: value })}
      />
      {error ? <p role="alert" className={`text-sm ${C.danger}`}>{error}</p> : null}
      <div className="flex justify-end gap-2">
        <button
          type="button"
          onClick={onCancel}
          className={`min-h-11 rounded-xxs px-3 text-sm ${C.text50} ${C.hoverBgLight}`}
        >
          キャンセル
        </button>
        <button
          type="button"
          onClick={onSaveField}
          className={`min-h-11 rounded-full px-4 text-sm ${C.bgBrand} ${C.textOnBrand}`}
        >
          検査項目情報を保存
        </button>
      </div>
    </>
  );
}

interface ExamTypeFieldReferenceRangesProps {
  animalSpecies: AnimalSpecies[];
  isPending: boolean;
  isError: boolean;
  rangeDrafts: ReferenceRangeDraft[];
  onToggleSpecies: (speciesId: string) => void;
  onUpdateRange: (
    speciesId: string,
    update: (draft: ReferenceRangeDraft) => ReferenceRangeDraft,
  ) => void;
  onSaveRanges: () => void;
}

export function ExamTypeFieldReferenceRanges({
  animalSpecies,
  isPending,
  isError,
  rangeDrafts,
  onToggleSpecies,
  onUpdateRange,
  onSaveRanges,
}: ExamTypeFieldReferenceRangesProps) {
  return (
    <div className={`space-y-3 border-t pt-4 ${C.borderLight}`}>
      <h4 className={`text-sm font-medium ${C.text}`}>動物種別の基準範囲</h4>
      {isError ? (
        <p role="alert" aria-atomic="true" className={`text-sm ${C.danger}`}>
          動物種の取得に失敗したため、基準範囲を設定できません。
        </p>
      ) : isPending ? (
        <p
          role="status"
          aria-live="polite"
          aria-atomic="true"
          className={`text-sm ${C.text50}`}
        >
          動物種を読み込み中です。基準範囲はまだ設定できません。
        </p>
      ) : animalSpecies.length === 0 ? (
        <p
          role="status"
          aria-live="polite"
          aria-atomic="true"
          className={`text-sm ${C.text50}`}
        >
          動物種マスタが登録されていないため、基準範囲を設定できません。
        </p>
      ) : (
        <>
          {animalSpecies.map((species) => {
            const draft = rangeDrafts.find(
              (item) => item.animalSpeciesId === species.id,
            );
            return (
              <div key={species.id} className={`rounded-xs border p-2 ${C.borderLight} ${C.bgWhite}`}>
                <label className={`flex min-h-11 items-center gap-2 text-sm ${C.text}`}>
                  <input
                    type="checkbox"
                    checked={draft !== undefined}
                    onChange={() => onToggleSpecies(species.id)}
                  />
                  {species.name}の基準範囲を使用
                </label>
                {draft !== undefined ? (
                  <ReferenceRangeInputs
                    speciesName={species.name}
                    draft={draft}
                    onChange={(update) => onUpdateRange(species.id, update)}
                  />
                ) : null}
              </div>
            );
          })}
          <datalist id="exam-qualitative-values">
            {QUALITATIVE_VALUES.map((item) => <option key={item} value={item} />)}
          </datalist>
          <div className="flex justify-end">
            <button
              type="button"
              onClick={onSaveRanges}
              className={`min-h-11 rounded-full px-4 text-sm ${C.bgBrand} ${C.textOnBrand}`}
            >
              基準範囲を保存
            </button>
          </div>
        </>
      )}
    </div>
  );
}

interface ExamTypeFieldEditorSessionProps {
  examTypeId: string;
  editingId: string | "new";
  editingField: ExaminationTypeField | null;
  canCreate: boolean;
  canEdit: boolean;
  animalSpecies: AnimalSpecies[];
  isPending: boolean;
  isError: boolean;
  onDirtyChange?: (dirty: boolean) => void;
  onClose: () => void;
}

export function ExamTypeFieldEditorSession({
  examTypeId,
  editingId,
  editingField,
  canCreate,
  canEdit,
  animalSpecies,
  isPending,
  isError,
  onDirtyChange,
  onClose,
}: ExamTypeFieldEditorSessionProps) {
  const session = useExamTypeFieldSession({
    examTypeId,
    editingId,
    editingField,
    canCreate,
    canEdit,
    onDirtyChange,
    onClose,
  });

  return (
    <div className={`mt-4 space-y-3 rounded-xs border p-3 ${C.borderLight} ${C.bgPage}`}>
      <ExamTypeFieldDraftForm
        editingId={editingId}
        fieldDraft={session.fieldDraft}
        error={session.error}
        onFieldChange={session.patchFieldDraft}
        onCancel={session.cancelEdit}
        onSaveField={() => {
          void session.saveField();
        }}
      />
      {editingField !== null ? (
        <ExamTypeFieldReferenceRanges
          animalSpecies={animalSpecies}
          isPending={isPending}
          isError={isError}
          rangeDrafts={session.rangeDrafts}
          onToggleSpecies={session.toggleSpecies}
          onUpdateRange={session.updateRange}
          onSaveRanges={() => {
            void session.saveRanges();
          }}
        />
      ) : null}
    </div>
  );
}
