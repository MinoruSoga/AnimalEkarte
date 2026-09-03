import { useCallback, useEffect, useState, type KeyboardEvent } from "react";
import Plus from "lucide-react/dist/esm/icons/plus";

import { C, ICON, STYLE } from "@/lib/design-tokens";

import { useGetAnimalSpecies } from "../api/animal-species";
import type { ExaminationTypeField, ExaminationTypeMaster } from "../api/exam-types-master";
import { ExamTypeFieldEditorSession } from "./ExamTypeFieldDraftPanel";
import { ExamTypeFieldsTable } from "./ExamTypeFieldsTable";
import { useExamTypeFieldsList } from "../hooks/use-exam-type-fields-list";

interface ExamTypeFieldsEditorProps {
  examType: ExaminationTypeMaster;
  canCreate: boolean;
  canEdit: boolean;
  canDelete: boolean;
  onDirtyChange?: (dirty: boolean) => void;
}

function handleExamTypeFieldsNestedKeyDown(event: KeyboardEvent<HTMLElement>) {
  if (event.key !== "Enter") return;
  const target = event.target;
  if (target instanceof HTMLInputElement && (target.type === "text" || target.type === "number")) {
    event.preventDefault();
    event.stopPropagation();
  }
}

export function ExamTypeFieldsEditor(props: ExamTypeFieldsEditorProps) {
  const { examType, canCreate, canEdit, canDelete } = props;
  return (
    <ExamTypeFieldsEditorState
      key={`${examType.id}:${canCreate}:${canEdit}:${canDelete}`}
      {...props}
    />
  );
}

function ExamTypeFieldsEditorState({
  examType,
  canCreate,
  canEdit,
  canDelete,
  onDirtyChange,
}: ExamTypeFieldsEditorProps) {
  const { data: animalSpecies = [], isPending, isError } = useGetAnimalSpecies();
  const [editingId, setEditingId] = useState<string | "new" | null>(null);
  const [hasDirtyDraft, setHasDirtyDraft] = useState(false);
  const { orderedItems, sensors, handleDragEnd, removeField } = useExamTypeFieldsList(
    examType,
    canEdit,
    hasDirtyDraft,
  );

  const handleDirtyChange = useCallback(
    (dirty: boolean) => {
      setHasDirtyDraft(dirty);
      onDirtyChange?.(dirty);
    },
    [onDirtyChange],
  );

  useEffect(() => {
    return () => onDirtyChange?.(false);
  }, [onDirtyChange]);

  const startCreate = useCallback(() => {
    if (!canCreate || hasDirtyDraft) return;
    setEditingId("new");
  }, [canCreate, hasDirtyDraft]);

  const startEdit = useCallback(
    (field: ExaminationTypeField) => {
      if (!canEdit || hasDirtyDraft) return;
      setEditingId(field.id);
    },
    [canEdit, hasDirtyDraft],
  );

  const editingField =
    editingId === null || editingId === "new"
      ? null
      : (examType.items.find((item) => item.id === editingId) ?? null);

  return (
    <section
      className={`mt-4 pt-4 ${STYLE.sectionDivider}`}
      aria-label="検査項目設定"
      onKeyDown={handleExamTypeFieldsNestedKeyDown}
    >
      <div className="mb-3 flex items-center justify-between gap-2">
        <h3 className={`text-sm font-medium ${C.text}`}>検査項目</h3>
        {canCreate ? (
          <button
            type="button"
            onClick={startCreate}
            aria-label="検査項目を追加"
            disabled={hasDirtyDraft}
            className={`inline-flex min-h-11 items-center gap-1 rounded-xxs px-2 text-sm ${C.textBrand} ${C.hoverBgLight}`}
          >
            <Plus className={ICON.smXs} aria-hidden="true" />
            追加
          </button>
        ) : null}
      </div>

      <ExamTypeFieldsTable
        orderedItems={orderedItems}
        sensors={sensors}
        onDragEnd={handleDragEnd}
        canEdit={canEdit}
        canDelete={canDelete}
        hasDirtyDraft={hasDirtyDraft}
        examTypeId={examType.id}
        onStartEdit={startEdit}
        onDeleteField={removeField}
      />

      {editingId !== null ? (
        <ExamTypeFieldEditorSession
          key={String(editingId)}
          examTypeId={examType.id}
          editingId={editingId}
          editingField={editingField}
          canCreate={canCreate}
          canEdit={canEdit}
          animalSpecies={animalSpecies}
          isPending={isPending}
          isError={isError}
          onDirtyChange={handleDirtyChange}
          onClose={() => setEditingId(null)}
        />
      ) : null}
    </section>
  );
}
