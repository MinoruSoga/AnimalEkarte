import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import type { ReservationTypeGroup } from "../api/reservation-type-groups";
import {
  ReservationTypeGroupHeader,
  ReservationTypeUncategorizedHeader,
} from "./ReservationTypeGroupedTableRows";

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
});
