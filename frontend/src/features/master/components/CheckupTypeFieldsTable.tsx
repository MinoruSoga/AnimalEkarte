import { useState } from "react";
import { DndContext, closestCenter, type DragEndEvent } from "@dnd-kit/core";
import { SortableContext, verticalListSortingStrategy } from "@dnd-kit/sortable";
import Pencil from "lucide-react/dist/esm/icons/pencil";
import Trash2 from "lucide-react/dist/esm/icons/trash-2";
import type { useSensors } from "@dnd-kit/core";

import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { DataTable } from "@/components/shared/DataTable/DataTable";
import { SortableDataTableRow } from "@/components/shared/DataTable/SortableDataTableRow";
import { TableCell } from "@/components/ui/table";
import { C, ICON } from "@/lib/design-tokens";

import type { CheckupTypeFieldEntry } from "../api/checkup-type-fields";
import { CHECKUP_FIELD_TYPE_LABELS } from "./checkup-type-fields-editor-model";

// EMR-225: 健診フィールド定義の並べ替え可能テーブル（ExamTypeFieldsTable 同型）。

const FIELD_COLUMNS = [
  { header: "", className: "w-11 px-0" },
  { header: "項目名" },
  { header: "種別", className: "w-[110px]" },
  { header: "単位", className: "w-[80px]" },
  { header: "操作", className: "w-[96px]", align: "right" as const },
];

interface CheckupTypeFieldsTableProps {
  orderedItems: CheckupTypeFieldEntry[];
  sensors: ReturnType<typeof useSensors>;
  onDragEnd: (event: DragEndEvent) => void;
  canEdit: boolean;
  canDelete: boolean;
  hasDirtyDraft: boolean;
  checkupTypeId: string;
  onStartEdit: (field: CheckupTypeFieldEntry) => void;
  onDeleteField: (checkupTypeId: string, fieldId: string) => void;
}

export function CheckupTypeFieldsTable({
  orderedItems,
  sensors,
  onDragEnd,
  canEdit,
  canDelete,
  hasDirtyDraft,
  checkupTypeId,
  onStartEdit,
  onDeleteField,
}: CheckupTypeFieldsTableProps) {
  const [pendingDelete, setPendingDelete] = useState<CheckupTypeFieldEntry | null>(null);

  return (
    <>
      <DndContext sensors={sensors} collisionDetection={closestCenter} onDragEnd={onDragEnd}>
        <SortableContext
          items={orderedItems.map((field) => field.id)}
          strategy={verticalListSortingStrategy}
        >
          <DataTable
            columns={FIELD_COLUMNS}
            data={orderedItems}
            emptyMessage="健診項目が登録されていません"
            renderRow={(field) => (
              <SortableDataTableRow
                key={field.id}
                id={field.id}
                dragLabel={`並べ替え: 健診項目 ${field.name} (ID ${field.id})`}
                dragDisabled={!canEdit || hasDirtyDraft}
              >
                <TableCell>{field.name}</TableCell>
                <TableCell>{CHECKUP_FIELD_TYPE_LABELS[field.fieldType]}</TableCell>
                <TableCell>{field.unit || "-"}</TableCell>
                <TableCell className="text-right">
                  {canEdit ? (
                    <button
                      type="button"
                      onClick={() => onStartEdit(field)}
                      disabled={hasDirtyDraft}
                      aria-label={`編集: 健診項目 ${field.name} (ID ${field.id})`}
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
                        setPendingDelete(field);
                      }}
                      disabled={hasDirtyDraft}
                      aria-label={`削除: 健診項目 ${field.name} (ID ${field.id})`}
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
      <ConfirmDialog
        open={pendingDelete !== null}
        onClose={() => setPendingDelete(null)}
        onConfirm={() => {
          if (pendingDelete === null) return;
          onDeleteField(checkupTypeId, pendingDelete.id);
          setPendingDelete(null);
        }}
        title="健診項目を削除しますか？"
        description={`「${pendingDelete?.name ?? ""}」を削除します。この操作は取り消せません。`}
        confirmLabel="削除"
        variant="destructive"
      />
    </>
  );
}
