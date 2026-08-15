import {
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
  type KeyboardEvent,
} from "react";
import { DndContext, closestCenter } from "@dnd-kit/core";
import { SortableContext, verticalListSortingStrategy } from "@dnd-kit/sortable";
import Pencil from "lucide-react/dist/esm/icons/pencil";
import Plus from "lucide-react/dist/esm/icons/plus";
import Trash2 from "lucide-react/dist/esm/icons/trash-2";

import { DataTable } from "@/components/shared/DataTable/DataTable";
import { SortableDataTableRow } from "@/components/shared/DataTable/SortableDataTableRow";
import { TableCell } from "@/components/ui/table";
import { useSortableList } from "@/hooks/use-sortable-list";
import { C, ICON, STYLE } from "@/lib/design-tokens";

import { useGetAnimalSpecies } from "../api/animal-species";
import {
  type ExaminationTypeField,
  type ExaminationTypeMaster,
  useCreateExaminationTypeField,
  useDeleteExaminationTypeField,
  useReorderExaminationTypeFields,
  useReplaceExamTypeFieldReferenceRanges,
  useUpdateExaminationTypeField,
} from "../api/exam-types-master";
import {
  QUALITATIVE_VALUES,
  buildReferenceRangeRequest,
  toReferenceRangeDraft,
  validateReferenceRangeDrafts,
  type ReferenceRangeDraft,
} from "./exam-type-fields-editor-model";

const FIELD_COLUMNS = [
  { header: "", className: "w-11 px-0" },
  { header: "項目名" },
  { header: "単位", className: "w-[100px]" },
  { header: "操作", className: "w-[96px]", align: "right" as const },
];

interface ExamTypeFieldsEditorProps {
  examType: ExaminationTypeMaster;
  canCreate: boolean;
  canEdit: boolean;
  canDelete: boolean;
  onDirtyChange?: (dirty: boolean) => void;
}

interface FieldDraft {
  name: string;
  inspectionValue: string;
  normalValue: string;
  unit: string;
}

const emptyFieldDraft = (): FieldDraft => ({
  name: "",
  inspectionValue: "",
  normalValue: "",
  unit: "",
});

const fieldToDraft = (field: ExaminationTypeField): FieldDraft => ({
  name: field.name,
  inspectionValue: field.inspectionValue,
  normalValue: field.normalValue,
  unit: field.unit,
});

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
  const createField = useCreateExaminationTypeField();
  const updateField = useUpdateExaminationTypeField();
  const deleteField = useDeleteExaminationTypeField();
  const reorderFields = useReorderExaminationTypeFields();
  const replaceRanges = useReplaceExamTypeFieldReferenceRanges();
  const {
    data: animalSpecies = [],
    isPending,
    isError,
  } = useGetAnimalSpecies();
  const [editingId, setEditingId] = useState<string | "new" | null>(null);
  const [fieldDraft, setFieldDraft] = useState<FieldDraft>(emptyFieldDraft);
  const [rangeDrafts, setRangeDrafts] = useState<ReferenceRangeDraft[]>([]);
  const [error, setError] = useState("");
  const [fieldDirty, setFieldDirty] = useState(false);
  const [rangeDirty, setRangeDirty] = useState(false);
  const hasDirtyDraft = fieldDirty || rangeDirty;
  const resetOrderRef = useRef<() => void>(() => {});

  const { orderedItems, sensors, handleDragEnd, resetOrder } = useSortableList({
    items: examType.items,
    onReorder: (ids) => {
      if (!canEdit || hasDirtyDraft) return;
      reorderFields.mutate(
        {
          examTypeId: examType.id,
          ids: ids.map(Number),
        },
        { onError: () => resetOrderRef.current() },
      );
    },
  });

  useEffect(() => {
    resetOrderRef.current = resetOrder;
  }, [resetOrder]);

  useEffect(() => {
    onDirtyChange?.(hasDirtyDraft);
  }, [hasDirtyDraft, onDirtyChange]);

  useEffect(() => {
    return () => onDirtyChange?.(false);
  }, [onDirtyChange]);

  const editingField = useMemo(
    () => editingId === null || editingId === "new"
      ? null
      : examType.items.find((item) => item.id === editingId) ?? null,
    [editingId, examType.items],
  );

  const startCreate = useCallback(() => {
    if (!canCreate || hasDirtyDraft) return;
    setEditingId("new");
    setFieldDraft(emptyFieldDraft());
    setRangeDrafts([]);
    setError("");
    setFieldDirty(false);
    setRangeDirty(false);
  }, [canCreate, hasDirtyDraft]);

  const startEdit = useCallback((field: ExaminationTypeField) => {
    if (!canEdit || hasDirtyDraft) return;
    setEditingId(field.id);
    setFieldDraft(fieldToDraft(field));
    setRangeDrafts(field.referenceRanges.map(toReferenceRangeDraft));
    setError("");
    setFieldDirty(false);
    setRangeDirty(false);
  }, [canEdit, hasDirtyDraft]);

  const saveField = useCallback(async () => {
    if (!fieldDraft.name.trim()) {
      setError("検査項目名を入力してください");
      return;
    }
    const req = {
      name: fieldDraft.name.trim(),
      inspection_value: fieldDraft.inspectionValue,
      normal_value: fieldDraft.normalValue,
      unit: fieldDraft.unit,
    };
    try {
      if (editingId === "new") {
        if (!canCreate) return;
        await createField.mutateAsync({ examTypeId: examType.id, req });
      } else if (editingId !== null) {
        if (!canEdit) return;
        await updateField.mutateAsync({
          examTypeId: examType.id,
          fieldId: editingId,
          req,
        });
      }
    } catch {
      // The mutation hook owns user-facing error notification.
      return;
    }
    setError("");
    setFieldDirty(false);
    if (editingId === "new") {
      setEditingId(null);
      setRangeDirty(false);
    }
  }, [
    canCreate,
    canEdit,
    createField,
    editingId,
    examType.id,
    fieldDraft,
    updateField,
  ]);

  const saveRanges = useCallback(async () => {
    if (!canEdit || editingId === null || editingId === "new") return;
    const validationError = validateReferenceRangeDrafts(rangeDrafts);
    if (validationError) {
      setError(validationError);
      return;
    }
    try {
      await replaceRanges.mutateAsync({
        examTypeId: examType.id,
        fieldId: editingId,
        ranges: buildReferenceRangeRequest(rangeDrafts),
      });
    } catch {
      // The mutation hook owns user-facing error notification.
      return;
    }
    setError("");
    setRangeDirty(false);
  }, [canEdit, editingId, examType.id, rangeDrafts, replaceRanges]);

  const toggleSpecies = useCallback((speciesId: string) => {
    setRangeDrafts((previous) => {
      if (previous.some((draft) => draft.animalSpeciesId === speciesId)) {
        return previous.filter((draft) => draft.animalSpeciesId !== speciesId);
      }
      return [
        ...previous,
        { animalSpeciesId: speciesId, mode: "numeric", min: "", max: "" },
      ];
    });
    setError("");
    setRangeDirty(true);
  }, []);

  const updateRange = useCallback((
    speciesId: string,
    update: (draft: ReferenceRangeDraft) => ReferenceRangeDraft,
  ) => {
    setRangeDrafts((previous) => previous.map((draft) =>
      draft.animalSpeciesId === speciesId ? update(draft) : draft
    ));
    setError("");
    setRangeDirty(true);
  }, []);

  const handleNestedKeyDown = useCallback((event: KeyboardEvent<HTMLElement>) => {
    if (event.key !== "Enter") return;
    const target = event.target;
    if (
      target instanceof HTMLInputElement &&
      (target.type === "text" || target.type === "number")
    ) {
      event.preventDefault();
      event.stopPropagation();
    }
  }, []);

  return (
    <section
      className={`mt-4 pt-4 ${STYLE.sectionDivider}`}
      aria-label="検査項目設定"
      onKeyDown={handleNestedKeyDown}
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

      <DndContext
        sensors={sensors}
        collisionDetection={closestCenter}
        onDragEnd={handleDragEnd}
      >
        <SortableContext
          items={orderedItems.map((field) => field.id)}
          strategy={verticalListSortingStrategy}
        >
          <DataTable
            columns={FIELD_COLUMNS}
            data={orderedItems}
            emptyMessage="検査項目が登録されていません"
            renderRow={(field) => (
              <SortableDataTableRow
                key={field.id}
                id={field.id}
                dragLabel={`並べ替え: 検査項目 ${field.name} (ID ${field.id})`}
                dragDisabled={!canEdit || hasDirtyDraft}
              >
                <TableCell>{field.name}</TableCell>
                <TableCell>{field.unit || "-"}</TableCell>
                <TableCell className="text-right">
                  {canEdit ? (
                    <button
                      type="button"
                      onClick={() => startEdit(field)}
                      disabled={hasDirtyDraft}
                      aria-label={`編集: 検査項目 ${field.name} (ID ${field.id})`}
                      className={`inline-flex min-h-11 min-w-11 items-center justify-center rounded-xxs ${C.text50} ${C.hoverBgLight}`}
                    >
                      <Pencil className={ICON.smXs} aria-hidden="true" />
                    </button>
                  ) : null}
                  {canDelete ? (
                    <button
                      type="button"
                      onClick={() => {
                        if (hasDirtyDraft) return;
                        deleteField.mutate({
                          examTypeId: examType.id,
                          fieldId: field.id,
                        });
                      }}
                      disabled={hasDirtyDraft}
                      aria-label={`削除: 検査項目 ${field.name} (ID ${field.id})`}
                      className={`inline-flex min-h-11 min-w-11 items-center justify-center rounded-xxs ${C.text50} ${C.hoverTextDanger} ${C.hoverBgLight}`}
                    >
                      <Trash2 className={ICON.smXs} aria-hidden="true" />
                    </button>
                  ) : null}
                </TableCell>
              </SortableDataTableRow>
            )}
          />
        </SortableContext>
      </DndContext>

      {editingId !== null ? (
        <div className={`mt-4 space-y-3 rounded-xs border p-3 ${C.borderLight} ${C.bgPage}`}>
          <h4 className={`text-sm font-medium ${C.text}`}>
            {editingId === "new" ? "検査項目を追加" : "検査項目を編集"}
          </h4>
          <FieldInput
            label="検査項目名"
            value={fieldDraft.name}
            onChange={(value) => {
              setFieldDraft((previous) => ({ ...previous, name: value }));
              setFieldDirty(true);
            }}
          />
          <FieldInput
            label="単位"
            value={fieldDraft.unit}
            onChange={(value) => {
              setFieldDraft((previous) => ({ ...previous, unit: value }));
              setFieldDirty(true);
            }}
          />
          <FieldInput
            label="検査値"
            value={fieldDraft.inspectionValue}
            onChange={(value) => {
              setFieldDraft((previous) => ({
                ...previous,
                inspectionValue: value,
              }));
              setFieldDirty(true);
            }}
          />
          <FieldInput
            label="正常値"
            value={fieldDraft.normalValue}
            onChange={(value) => {
              setFieldDraft((previous) => ({
                ...previous,
                normalValue: value,
              }));
              setFieldDirty(true);
            }}
          />
          {error ? <p role="alert" className={`text-sm ${C.danger}`}>{error}</p> : null}
          <div className="flex justify-end gap-2">
            <button
              type="button"
              onClick={() => {
                setEditingId(null);
                setError("");
                setFieldDirty(false);
                setRangeDirty(false);
              }}
              className={`min-h-11 rounded-xxs px-3 text-sm ${C.text50} ${C.hoverBgLight}`}
            >
              キャンセル
            </button>
            <button
              type="button"
              onClick={saveField}
              className={`min-h-11 rounded-full px-4 text-sm ${C.bgBrand} ${C.textOnBrand}`}
            >
              検査項目情報を保存
            </button>
          </div>

          {editingField !== null ? (
            <div className={`space-y-3 border-t pt-4 ${C.borderLight}`}>
              <h4 className={`text-sm font-medium ${C.text}`}>動物種別の基準範囲</h4>
              {isError ? (
                <p
                  role="alert"
                  aria-atomic="true"
                  className={`text-sm ${C.danger}`}
                >
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
                            onChange={() => toggleSpecies(species.id)}
                          />
                          {species.name}の基準範囲を使用
                        </label>
                        {draft !== undefined ? (
                          <ReferenceRangeInputs
                            speciesName={species.name}
                            draft={draft}
                            onChange={(update) => updateRange(species.id, update)}
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
                      onClick={saveRanges}
                      className={`min-h-11 rounded-full px-4 text-sm ${C.bgBrand} ${C.textOnBrand}`}
                    >
                      基準範囲を保存
                    </button>
                  </div>
                </>
              )}
            </div>
          ) : null}
        </div>
      ) : null}
    </section>
  );
}

function FieldInput({
  label,
  value,
  onChange,
}: {
  label: string;
  value: string;
  onChange: (value: string) => void;
}) {
  return (
    <label className={`block text-sm ${C.text65}`}>
      {label}
      <input
        value={value}
        onChange={(event) => onChange(event.target.value)}
        className={`mt-1 min-h-11 w-full rounded-xs border px-2 ${C.borderMedium} ${C.bgWhite} ${C.text}`}
      />
    </label>
  );
}

function ReferenceRangeInputs({
  speciesName,
  draft,
  onChange,
}: {
  speciesName: string;
  draft: ReferenceRangeDraft;
  onChange: (update: (draft: ReferenceRangeDraft) => ReferenceRangeDraft) => void;
}) {
  const type = draft.mode === "numeric" ? "number" : "text";
  const rangeKind = draft.mode === "numeric" ? "数値" : "定性";
  return (
    <div className="grid grid-cols-[120px_1fr_1fr] gap-2">
      <label className={`text-xs ${C.text65}`}>
        種別
        <select
          aria-label={`${speciesName}の基準範囲種別`}
          value={draft.mode}
          onChange={(event) => onChange((previous) => ({
            ...previous,
            mode: event.target.value === "qualitative" ? "qualitative" : "numeric",
            min: "",
            max: "",
            qualitativeMin: undefined,
            qualitativeMax: undefined,
          }))}
          className={`mt-1 min-h-11 w-full rounded-xs border px-2 ${C.borderMedium} ${C.bgWhite} ${C.text}`}
        >
          <option value="numeric">数値</option>
          <option value="qualitative">定性</option>
        </select>
      </label>
      <RangeBoundInput
        label={`${speciesName}の${rangeKind}下限`}
        type={type}
        value={draft.min}
        onChange={(value) => onChange((previous) => ({ ...previous, min: value }))}
      />
      <RangeBoundInput
        label={`${speciesName}の${rangeKind}上限`}
        type={type}
        value={draft.max}
        onChange={(value) => onChange((previous) => ({ ...previous, max: value }))}
      />
      {draft.mode === "qualitative" ? (
        <p className={`col-span-3 text-xs ${C.text50}`}>
          選択可能: {QUALITATIVE_VALUES.join("、")}
        </p>
      ) : null}
    </div>
  );
}

function RangeBoundInput({
  label,
  type,
  value,
  onChange,
}: {
  label: string;
  type: "number" | "text";
  value: string;
  onChange: (value: string) => void;
}) {
  return (
    <label className={`text-xs ${C.text65}`}>
      {label.endsWith("下限") ? "下限" : "上限"}
      <input
        type={type}
        step={type === "number" ? "any" : undefined}
        aria-label={label}
        list={type === "text" ? "exam-qualitative-values" : undefined}
        value={value}
        onChange={(event) => onChange(event.target.value)}
        className={`mt-1 min-h-11 w-full rounded-xs border px-2 ${C.borderMedium} ${C.bgWhite} ${C.text}`}
      />
    </label>
  );
}
