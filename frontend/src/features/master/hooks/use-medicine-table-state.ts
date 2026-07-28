import { useCallback, useDeferredValue, useMemo, useState } from "react";
import { flushSync } from "react-dom";
import type { DragEndEvent } from "@dnd-kit/core";
import type { UseMutationResult } from "@tanstack/react-query";
import type { ActiveFilter } from "@/components/shared/PropertyFilter/types";
import { handleApiError } from "@/lib/handle-api-error";
import { useSortableList } from "@/hooks/use-sortable-list";
import type { ReorderMedicinesRequest, UpdateMedicineRequest } from "@/types/medicine";
import type { Medicine } from "@/types";
import {
  applyMedicineCategoryOverrides,
  groupFilteredMedicines,
  resolveMedicineDrag,
} from "../routes/medicine-settings-model";

interface UseMedicineTableStateOptions {
  medicines: Medicine[];
  reorderMutation: UseMutationResult<void, Error, ReorderMedicinesRequest>;
  updateMutation: UseMutationResult<Medicine, Error, { id: string; req: UpdateMedicineRequest }>;
  canEdit: boolean;
}

export function useMedicineTableState({
  medicines,
  reorderMutation,
  updateMutation,
  canEdit,
}: UseMedicineTableStateOptions) {
  const [searchTerm, setSearchTerm] = useState("");
  const [activeFilters, setActiveFilters] = useState<ActiveFilter[]>([]);
  const [collapsedGroups, setCollapsedGroups] = useState<Set<string>>(new Set());
  const [overrideCategories, setOverrideCategories] = useState<Map<string, string | undefined>>(
    new Map(),
  );

  const {
    orderedItems: sortedMedicines,
    sensors,
    activeId,
    handleDragStart,
    handleDragCancel,
    handleDragEnd: handleFlatSortDragEnd,
    resetOrder,
  } = useSortableList({
    items: medicines,
    onReorder: (newIds) => {
      if (!canEdit) return;
      reorderMutation.mutate(
        { ids: newIds.map(Number) },
        { onSuccess: resetOrder },
      );
    },
  });

  const orderedMedicines = useMemo(() => {
    return applyMedicineCategoryOverrides(sortedMedicines, overrideCategories);
  }, [sortedMedicines, overrideCategories]);

  const medicinesById = useMemo(
    () => new Map(medicines.map((medicine) => [medicine.id, medicine])),
    [medicines],
  );

  const orderedMedicinesById = useMemo(
    () => new Map(orderedMedicines.map((medicine) => [medicine.id, medicine])),
    [orderedMedicines],
  );

  const deferredSearch = useDeferredValue(searchTerm);

  const { groupedMedicines, ungroupedMedicines, totalCount } = useMemo(() => {
    return groupFilteredMedicines({
      orderedMedicines,
      activeFilters,
      searchTerm: deferredSearch,
      medicinesById,
    });
  }, [orderedMedicines, activeFilters, deferredSearch, medicinesById]);

  const toggleGroup = useCallback((key: string) => {
    setCollapsedGroups((prev) => {
      const next = new Set(prev);
      if (next.has(key)) {
        next.delete(key);
      } else {
        next.add(key);
      }
      return next;
    });
  }, []);

  const handleDragEnd = useCallback(
    (event: DragEndEvent) => {
      if (!canEdit) return;
      const { active, over } = event;
      if (!over || active.id === over.id) return;

      const activeItemId = String(active.id);
      const overItemId = String(over.id);

      const dragResolution = resolveMedicineDrag({
        activeItemId,
        overItemId,
        orderedMedicinesById,
      });

      if (dragResolution.type === "none") return;

      if (dragResolution.type === "move-category") {
        flushSync(() => {
          setOverrideCategories((prev) => {
            const next = new Map(prev);
            next.set(dragResolution.activeItemId, dragResolution.overParentId ?? undefined);
            return next;
          });
        });

        const clearOptimistic = () => {
          setOverrideCategories((prev) => {
            const next = new Map(prev);
            next.delete(dragResolution.activeItemId);
            return next;
          });
        };

        updateMutation.mutate(
          { id: dragResolution.activeItemId, req: dragResolution.updateRequest },
          {
            onSuccess: clearOptimistic,
            onError: (error: unknown) => {
              handleApiError(error, "カテゴリの変更");
              clearOptimistic();
            },
          },
        );
        return;
      }

      handleFlatSortDragEnd(event);
    },
    [canEdit, orderedMedicinesById, updateMutation, handleFlatSortDragEnd],
  );

  return {
    searchTerm,
    setSearchTerm,
    activeFilters,
    setActiveFilters,
    totalCount,
    groupedMedicines,
    ungroupedMedicines,
    collapsedGroups,
    orderedMedicinesById,
    sensors,
    activeId,
    handleDragStart,
    handleDragEnd,
    handleDragCancel,
    toggleGroup,
  };
}
