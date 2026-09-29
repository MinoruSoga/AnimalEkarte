import {
  useCallback,
  useEffect,
  useLayoutEffect,
  useMemo,
  useRef,
  useState,
  type KeyboardEvent,
} from "react";
import Plus from "lucide-react/dist/esm/icons/plus";

import { useSortableList } from "@/hooks/use-sortable-list";
import { useGetCheckupTypeFields } from "@/hooks/use-checkup-fields";
import { C, ICON, STYLE } from "@/lib/design-tokens";

import {
  toCheckupTypeFieldEntry,
  useDeleteCheckupTypeField,
  useReorderCheckupTypeFields,
  type CheckupTypeFieldEntry,
} from "../api/checkup-type-fields";
import { CheckupTypeFieldEditorSession } from "./CheckupTypeFieldDraftPanel";
import { CheckupTypeFieldsTable } from "./CheckupTypeFieldsTable";

// EMR-225: 定期健診パッケージのフィールド定義エディタ（ExamTypeFieldsEditor 同型）。
// checkup_types 一覧レスポンスは fields を埋め込まないため、GET /fields はここで
// 自前取得する（useGetCheckupTypeFields — カルテ画面と query key 共有）。

interface CheckupTypeFieldsEditorProps {
  checkupTypeId: string;
  canCreate: boolean;
  canEdit: boolean;
  canDelete: boolean;
  onDirtyChange?: (dirty: boolean) => void;
}

function handleCheckupTypeFieldsNestedKeyDown(event: KeyboardEvent<HTMLElement>) {
  if (event.key !== "Enter") return;
  const target = event.target;
  if (target instanceof HTMLInputElement && (target.type === "text" || target.type === "number")) {
    event.preventDefault();
    event.stopPropagation();
  }
}

export function CheckupTypeFieldsEditor(props: CheckupTypeFieldsEditorProps) {
  const { checkupTypeId, canCreate, canEdit, canDelete } = props;
  return (
    <CheckupTypeFieldsEditorState
      key={`${checkupTypeId}:${canCreate}:${canEdit}:${canDelete}`}
      {...props}
    />
  );
}

function CheckupTypeFieldsEditorState({
  checkupTypeId,
  canCreate,
  canEdit,
  canDelete,
  onDirtyChange,
}: CheckupTypeFieldsEditorProps) {
  const { data: rows = [], isPending, isError } = useGetCheckupTypeFields(checkupTypeId);
  const items = useMemo(() => rows.map(toCheckupTypeFieldEntry), [rows]);
  const deleteField = useDeleteCheckupTypeField();
  const reorderFields = useReorderCheckupTypeFields();
  const [editingId, setEditingId] = useState<string | "new" | null>(null);
  const [hasDirtyDraft, setHasDirtyDraft] = useState(false);
  const resetOrderRef = useRef<() => void>(() => {});

  const { orderedItems, sensors, handleDragEnd, resetOrder } = useSortableList({
    items,
    onReorder: (ids) => {
      if (!canEdit || hasDirtyDraft) return;
      reorderFields.mutate(
        { checkupTypeId, ids: ids.map(Number) },
        { onError: () => resetOrderRef.current() },
      );
    },
  });

  useLayoutEffect(() => {
    resetOrderRef.current = resetOrder;
  }, [resetOrder]);

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
    (field: CheckupTypeFieldEntry) => {
      if (!canEdit || hasDirtyDraft) return;
      setEditingId(field.id);
    },
    [canEdit, hasDirtyDraft],
  );

  const removeField = useCallback(
    (targetCheckupTypeId: string, fieldId: string) => {
      deleteField.mutate({ checkupTypeId: targetCheckupTypeId, fieldId });
    },
    [deleteField],
  );

  const editingRow =
    editingId === null || editingId === "new"
      ? null
      : (rows.find((row) => String(row.id) === editingId) ?? null);

  return (
    <section
      className={`mt-4 pt-4 ${STYLE.sectionDivider}`}
      aria-label="健診項目設定"
      onKeyDown={handleCheckupTypeFieldsNestedKeyDown}
    >
      <div className="mb-3 flex items-center justify-between gap-2">
        <h3 className={`text-sm font-medium ${C.text}`}>健診項目</h3>
        {canCreate ? (
          <button
            type="button"
            onClick={startCreate}
            aria-label="健診項目を追加"
            disabled={hasDirtyDraft}
            className={`inline-flex min-h-11 items-center gap-1 rounded-xxs px-2 text-sm ${C.textBrand} ${C.hoverBgLight}`}
          >
            <Plus className={ICON.smXs} aria-hidden="true" />
            追加
          </button>
        ) : null}
      </div>

      {isError ? (
        <p role="alert" aria-atomic="true" className={`text-sm ${C.danger}`}>
          健診項目の取得に失敗しました。
        </p>
      ) : isPending ? (
        <p role="status" aria-live="polite" aria-atomic="true" className={`text-sm ${C.text50}`}>
          健診項目を読み込み中です。
        </p>
      ) : (
        <CheckupTypeFieldsTable
          orderedItems={orderedItems}
          sensors={sensors}
          onDragEnd={handleDragEnd}
          canEdit={canEdit}
          canDelete={canDelete}
          hasDirtyDraft={hasDirtyDraft}
          checkupTypeId={checkupTypeId}
          onStartEdit={startEdit}
          onDeleteField={removeField}
        />
      )}

      {editingId !== null ? (
        <CheckupTypeFieldEditorSession
          key={String(editingId)}
          checkupTypeId={checkupTypeId}
          editingId={editingId}
          editingField={editingRow}
          fieldCount={items.length}
          canCreate={canCreate}
          canEdit={canEdit}
          onDirtyChange={handleDirtyChange}
          onClose={() => setEditingId(null)}
        />
      ) : null}
    </section>
  );
}
