import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import type { CarePlanItem } from "../../api/care-plan-items";
import { ItemRow } from "./ItemRow";

const item = {
  id: "synthetic-care-plan",
  hospitalization_id: "synthetic-hospitalization",
  type: "treatment",
  name: "合成監査処置",
  description: "",
  timing: ["morning"],
  status: "active",
  notes: "",
  procedure_id: "synthetic-procedure",
  unit_price: 4_321,
  category: "synthetic",
  sort_order: 1,
  created_at: "2026-07-23T00:00:00+09:00",
  updated_at: "2026-07-23T00:00:00+09:00",
} satisfies CarePlanItem;

describe("ItemRow", () => {
  it("care planの単価を明示する", () => {
    render(<ItemRow item={item} isDeleting={false} />);

    const itemName = screen.getByText("合成監査処置");
    expect(itemName).toHaveClass("min-w-[8rem]");
    expect(itemName.parentElement).toHaveClass("flex-wrap", "sm:flex-nowrap");
    expect(screen.getByText("単価 ￥4,321")).toBeVisible();
  });

  it("44px操作から対象itemを編集・削除できる", async () => {
    const user = userEvent.setup();
    const onEdit = vi.fn();
    const onDelete = vi.fn();
    render(
      <ItemRow
        item={item}
        onEdit={onEdit}
        onDelete={onDelete}
        isDeleting={false}
      />,
    );

    const editButton = screen.getByRole("button", { name: "編集" });
    const deleteButton = screen.getByRole("button", { name: "削除" });
    expect(editButton).toHaveClass("min-h-11", "min-w-11");
    expect(deleteButton).toHaveClass("min-h-11", "min-w-11");

    await user.click(editButton);
    await user.click(deleteButton);
    expect(onEdit).toHaveBeenCalledWith(item.id);
    expect(onDelete).toHaveBeenCalledWith(item.id);
  });
});
