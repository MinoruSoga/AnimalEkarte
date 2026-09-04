import { act, renderHook } from "@testing-library/react";
import type { DragEndEvent } from "@dnd-kit/core";
import type { UseMutationResult } from "@tanstack/react-query";
import { describe, expect, it, vi } from "vitest";
import type { Medicine } from "@/types";
import type { ReorderMedicinesRequest, UpdateMedicineRequest } from "@/types/medicine";
import { useMedicineTableState } from "./use-medicine-table-state";

function makeMedicine(overrides: Partial<Medicine> = {}): Medicine {
  return {
    id: "1",
    name: "抗生剤A",
    parentId: "10",
    dosageForm: "tablet",
    medicineUnit: "per_tablet",
    price: 1000,
    defaultQuantity: 1,
    inventoryId: undefined,
    description: "",
    isActive: true,
    sortOrder: 0,
    taxType: "excluded",
    taxRate: 0.1,
    isNonInsurance: false,
    calculationType: "none",
    strength: undefined,
    frequencyPerDay: undefined,
    defaultDurationDays: undefined,
    doseParams: [],
    createdAt: "2026-01-01T00:00:00Z",
    updatedAt: "2026-01-01T00:00:00Z",
    ...overrides,
  };
}

function dragEvent(activeId: string, overId: string): DragEndEvent {
  return {
    active: { id: activeId },
    over: { id: overId },
  } as unknown as DragEndEvent;
}

function renderState({ canEdit, medicines }: { canEdit: boolean; medicines: Medicine[] }) {
  const reorder = vi.fn();
  const update = vi.fn();
  const reorderMutation = { mutate: reorder } as unknown as UseMutationResult<
    void,
    Error,
    ReorderMedicinesRequest
  >;
  const updateMutation = { mutate: update } as unknown as UseMutationResult<
    Medicine,
    Error,
    { id: string; req: UpdateMedicineRequest }
  >;

  const hook = renderHook(() =>
    useMedicineTableState({ medicines, reorderMutation, updateMutation, canEdit }),
  );

  return { ...hook, reorder, update };
}

describe("useMedicineTableState reorder permission guard", () => {
  it("blocks same-category reorder mutation when edit permission is absent", () => {
    const first = makeMedicine({ id: "1", parentId: "10" });
    const second = makeMedicine({ id: "2", parentId: "10" });
    const { result, reorder } = renderState({ canEdit: false, medicines: [first, second] });

    act(() => result.current.handleDragEnd(dragEvent("1", "2")));

    expect(reorder).not.toHaveBeenCalled();
  });

  it("blocks category-change update mutation when edit permission is absent", () => {
    const first = makeMedicine({ id: "1", parentId: "10" });
    const second = makeMedicine({ id: "2", parentId: "20" });
    const { result, update } = renderState({ canEdit: false, medicines: [first, second] });

    act(() => result.current.handleDragEnd(dragEvent("1", "2")));

    expect(update).not.toHaveBeenCalled();
  });

  it("keeps same-category reorder enabled when edit permission is present", () => {
    const first = makeMedicine({ id: "1", parentId: "10" });
    const second = makeMedicine({ id: "2", parentId: "10" });
    const { result, reorder } = renderState({ canEdit: true, medicines: [first, second] });

    act(() => result.current.handleDragEnd(dragEvent("1", "2")));

    expect(reorder).toHaveBeenCalledWith(
      { ids: [2, 1] },
      expect.objectContaining({ onSuccess: expect.any(Function) }),
    );
  });
});
