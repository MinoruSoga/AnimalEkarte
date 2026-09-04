import { DndContext, closestCenter, type DragEndEvent } from "@dnd-kit/core";
import { SortableContext, verticalListSortingStrategy } from "@dnd-kit/sortable";
import Pencil from "lucide-react/dist/esm/icons/pencil";
import Trash2 from "lucide-react/dist/esm/icons/trash-2";
import type { useSensors } from "@dnd-kit/core";

import { DataTable } from "@/components/shared/DataTable/DataTable";
import { SortableDataTableRow } from "@/components/shared/DataTable/SortableDataTableRow";
import { TableCell } from "@/components/ui/table";
import { C, ICON } from "@/lib/design-tokens";

import type { ExaminationTypeField } from "../api/exam-types-master";

const FIELD_COLUMNS = [
  { header: "", className: "w-11 px-0" },
  { header: "項目名" },
  { header: "単位", className: "w-[100px]" },
  { header: "操作", className: "w-[96px]", align: "right" as const },
];

interface ExamTypeFieldsTableProps {
  orderedItems: ExaminationTypeField[];
  sensors: ReturnType<typeof useSensors>;
  onDragEnd: (event: DragEndEvent) => void;
  canEdit: boolean;
  canDelete: boolean;
  hasDirtyDraft: boolean;
  examTypeId: string;
  onStartEdit: (field: ExaminationTypeField) => void;
  onDeleteField: (examTypeId: string, fieldId: string) => void;
}

export function ExamTypeFieldsTable({
  orderedItems,
  sensors,
  onDragEnd,
  canEdit,
  canDelete,
  hasDirtyDraft,
  examTypeId,
  onStartEdit,
  onDeleteField,
}: ExamTypeFieldsTableProps) {
  return (
    <DndContext sensors={sensors} collisionDetection={closestCenter} onDragEnd={onDragEnd}>
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
                    onClick={() => onStartEdit(field)}
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
                      onDeleteField(examTypeId, field.id);
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
  );
}
