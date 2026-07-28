import type { ReactNode } from "react";
import { render } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { TreatmentPlanTabContent } from "./TreatmentPlanTabContent";

const mocks = vi.hoisted(() => ({
  reorderCallbacks: [] as Array<(ids: string[]) => void>,
}));

vi.mock("@dnd-kit/core", () => ({
  DndContext: ({ children }: { children: ReactNode }) => <>{children}</>,
  closestCenter: vi.fn(),
}));

vi.mock("@dnd-kit/sortable", () => ({
  SortableContext: ({ children }: { children: ReactNode }) => <>{children}</>,
  verticalListSortingStrategy: vi.fn(),
}));

vi.mock("@/hooks/use-sortable-list", () => ({
  useSortableList: ({
    items,
    onReorder,
  }: {
    items: Array<{ id: string }>;
    onReorder: (ids: string[]) => void;
  }) => {
    mocks.reorderCallbacks.push(onReorder);
    return {
      orderedItems: items,
      sensors: [],
      handleDragEnd: vi.fn(),
    };
  },
}));

vi.mock("@/components/shared/DataTable/DataTable", () => ({
  DataTable: () => null,
  DESIGN_TABLE_HEADER_ROW: "",
  DESIGN_TABLE_HEADER_CELL: "",
}));

vi.mock("@/components/shared/PropertyFilter/PropertyFilter", () => ({
  PropertyFilter: () => null,
}));

function renderContent(canEdit: boolean, onReorder: (ids: number[]) => void) {
  render(
    <TreatmentPlanTabContent
      data={[]}
      entityLabel="診察"
      emptyMessage="診察が登録されていません"
      searchPlaceholder="診察名で検索..."
      onReorder={onReorder}
      onEditTargetChange={vi.fn()}
      canEdit={canEdit}
    />,
  );
}

function callLatestReorder() {
  const callback = mocks.reorderCallbacks.at(-1);
  expect(callback).toBeDefined();
  callback?.(["2", "1"]);
}

describe("TreatmentPlanTabContent reorder permission guard", () => {
  beforeEach(() => {
    mocks.reorderCallbacks.length = 0;
  });

  it("does not forward reorder when edit permission is absent", () => {
    const onReorder = vi.fn();
    renderContent(false, onReorder);
    callLatestReorder();
    expect(onReorder).not.toHaveBeenCalled();
  });

  it("forwards numeric IDs when edit permission is present", () => {
    const onReorder = vi.fn();
    renderContent(true, onReorder);
    callLatestReorder();
    expect(onReorder).toHaveBeenCalledWith([2, 1]);
  });
});
