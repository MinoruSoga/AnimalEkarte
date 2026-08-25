import type { ReactNode } from "react";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import type { ReservationType } from "../api/reservation-types";
import type { ReservationTypeGroup } from "../api/reservation-type-groups";
import {
  ReservationTypeGroupHeader,
  ReservationTypeRow,
  ReservationTypeUncategorizedHeader,
} from "./ReservationTypeGroupedTableRows";

vi.mock("@/components/shared/DataTable/SortableDataTableRow", () => ({
  SortableDataTableRow: ({ children }: { children: ReactNode }) => <tr>{children}</tr>,
}));

const group: ReservationTypeGroup = {
  id: "1",
  clinicId: "1",
  name: "診療系",
  color: "#2563eb",
  sortOrder: 1,
  isActive: true,
  createdAt: "2026-01-01T00:00:00Z",
  updatedAt: "2026-01-01T00:00:00Z",
};

const category: ReservationType = {
  id: "10",
  clinicId: "1",
  name: "一般診察",
  color: "#2563eb",
  isActive: true,
  description: "通常診療",
  sortOrder: 1,
  groupId: "1",
  createdAt: "2026-01-01T00:00:00Z",
  updatedAt: "2026-01-01T00:00:00Z",
  reservationDisplayName: "一般診察",
  durationMinutes: 15,
  shortName: "",
  showShortName: false,
  reservationVisible: true,
  reservationComment: "",
  reservationImageUrl: "",
  reservationDayOption: "none",
  isInternal: false,
  category: "general",
  parentId: undefined,
  parentName: undefined,
  isLeaf: true,
  depth: 0,
  childIds: [],
};

describe("ReservationTypeGroupedTableRows", () => {
  it("グループ名・編集・追加ボタンは44px以上の操作領域を保ち、各操作を通知する", async () => {
    const user = userEvent.setup();
    const onGroupEdit = vi.fn();
    const onCategoryAdd = vi.fn();

    render(
      <table>
        <tbody>
          <ReservationTypeGroupHeader
            group={group}
            count={2}
            isCollapsed={false}
            canEdit
            onToggle={vi.fn()}
            onGroupEdit={onGroupEdit}
            onCategoryAdd={onCategoryAdd}
          />
        </tbody>
      </table>,
    );

    const groupNameButton = screen.getByRole("button", { name: "診療系" });
    const editButton = screen.getByRole("button", { name: "編集" });
    const addButton = screen.getByRole("button", { name: "追加" });

    for (const button of [groupNameButton, editButton, addButton]) {
      expect(button).toHaveClass("min-h-11", "min-w-11");
    }

    await user.click(groupNameButton);
    await user.click(editButton);
    await user.click(addButton);

    expect(onGroupEdit).toHaveBeenCalledTimes(2);
    expect(onCategoryAdd).toHaveBeenCalledTimes(1);
  });

  it("未分類の追加ボタンは44px以上の操作領域を保ち、追加操作を通知する", async () => {
    const user = userEvent.setup();
    const onCategoryAdd = vi.fn();

    render(
      <table>
        <tbody>
          <ReservationTypeUncategorizedHeader
            count={1}
            isCollapsed={false}
            canEdit
            onToggle={vi.fn()}
            onCategoryAdd={onCategoryAdd}
          />
        </tbody>
      </table>,
    );

    const addButton = screen.getByRole("button", { name: "追加" });
    expect(addButton).toHaveClass("min-h-11", "min-w-11");

    await user.click(addButton);
    expect(onCategoryAdd).toHaveBeenCalledTimes(1);
  });

  it("ステータスセルはwhitespace-nowrapで有効バッジを1行表示する", () => {
    render(
      <table>
        <tbody>
          <ReservationTypeRow category={category} canEdit onEdit={vi.fn()} />
        </tbody>
      </table>,
    );

    const statusLabel = screen.getByText("有効");
    expect(statusLabel).toBeInTheDocument();
    expect(statusLabel.closest("td")?.className).toContain("whitespace-nowrap");
  });
});
