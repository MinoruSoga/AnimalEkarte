import { useCallback, useEffect, useRef } from "react";

import { useSortableList } from "@/hooks/use-sortable-list";

import {
  type ExaminationTypeMaster,
  useDeleteExaminationTypeField,
  useReorderExaminationTypeFields,
} from "../api/exam-types-master";

export function useExamTypeFieldsList(
  examType: ExaminationTypeMaster,
  canEdit: boolean,
  hasDirtyDraft: boolean,
) {
  const deleteField = useDeleteExaminationTypeField();
  const reorderFields = useReorderExaminationTypeFields();
  const resetOrderRef = useRef<() => void>(() => {});

  const { orderedItems, sensors, handleDragEnd, resetOrder } = useSortableList({
    items: examType.items,
    onReorder: (ids) => {
      if (!canEdit || hasDirtyDraft) return;
      reorderFields.mutate(
        { examTypeId: examType.id, ids: ids.map(Number) },
        { onError: () => resetOrderRef.current() },
      );
    },
  });

  useEffect(() => {
    resetOrderRef.current = resetOrder;
  }, [resetOrder]);

  const removeField = useCallback((examTypeId: string, fieldId: string) => {
    deleteField.mutate({ examTypeId, fieldId });
  }, [deleteField]);

  return { orderedItems, sensors, handleDragEnd, removeField };
}
