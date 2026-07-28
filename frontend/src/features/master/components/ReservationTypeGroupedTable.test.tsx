import type { ReactNode } from "react";
import { render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { STYLE } from "@/lib/design-tokens";
import { ReservationTypeGroupedTable } from "./ReservationTypeGroupedTable";

const mocks = vi.hoisted(() => ({
  onReorder: undefined as ((ids: string[]) => void) | undefined,
  reorder: vi.fn(),
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
  useSortableList: ({ onReorder }: { onReorder: (ids: string[]) => void }) => {
    mocks.onReorder = onReorder;
    return {
      orderedItems: [],
      sensors: [],
      handleDragStart: vi.fn(),
      handleDragEnd: vi.fn(),
      handleDragCancel: vi.fn(),
      resetOrder: vi.fn(),
    };
  },
}));

vi.mock("../api/reservation-types", () => ({
  useReorderReservationTypes: () => ({ mutate: mocks.reorder }),
}));

vi.mock("./ReservationTypeGroupedTableBody", () => ({
  ReservationTypeGroupedTableBody: () => <tbody />,
}));

describe("ReservationTypeGroupedTable", () => {
  beforeEach(() => {
    mocks.onReorder = undefined;
    vi.clearAllMocks();
  });

  it("狭い表示幅では表を切り取らず、ラッパー内で横スクロールできる", () => {
    render(
      <ReservationTypeGroupedTable
        groups={[]}
        categories={[]}
        onCategoryEdit={vi.fn()}
        onGroupEdit={vi.fn()}
        onCategoryAddInGroup={vi.fn()}
        canEdit
      />,
    );

    const tableWrapper = screen.getByRole("table").parentElement;
    expect(tableWrapper).toHaveClass("overflow-x-auto", "overflow-y-hidden");
    expect(tableWrapper).not.toHaveClass("overflow-hidden");
  });

  it("ヘッダセルはSTYLE.tableHeaderCellの仕様padding(px-4)を保ち、px-3で上書きしない", () => {
    render(
      <ReservationTypeGroupedTable
        groups={[]}
        categories={[]}
        onCategoryEdit={vi.fn()}
        onGroupEdit={vi.fn()}
        onCategoryAddInGroup={vi.fn()}
        canEdit
      />,
    );

    const headerCellClasses = STYLE.tableHeaderCell.split(/\s+/);
    for (const name of ["名称", "備考", "ステータス"]) {
      const th = screen.getByRole("columnheader", { name });
      expect(th).toHaveClass(...headerCellClasses);
      expect(th).not.toHaveClass("px-3");
    }
  });

  it("編集権限がない場合は並び替えmutationを呼ばない", () => {
    render(
      <ReservationTypeGroupedTable
        groups={[]}
        categories={[]}
        onCategoryEdit={vi.fn()}
        onGroupEdit={vi.fn()}
        onCategoryAddInGroup={vi.fn()}
        canEdit={false}
      />,
    );

    expect(mocks.onReorder).toBeDefined();
    mocks.onReorder?.(["2", "1"]);

    expect(mocks.reorder).not.toHaveBeenCalled();
  });
});
