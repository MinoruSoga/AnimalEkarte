import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { MedicineCalculationTypeNone } from "@/types/generated/models";
import type { Medicine } from "@/types";
import { MedicineCategoryHeaderRow } from "./MedicineTableRows";

const CATEGORY = {
  id: "medicine-group-1",
  name: "内用薬",
  parentId: undefined,
  dosageForm: "tablet",
  medicineUnit: "錠",
  price: 0,
  defaultQuantity: 1,
  inventoryId: undefined,
  description: "",
  isActive: true,
  sortOrder: 1,
  taxType: "excluded",
  taxRate: 0.1,
  isNonInsurance: false,
  createdAt: "",
  updatedAt: "",
  calculationType: MedicineCalculationTypeNone,
  strength: undefined,
  frequencyPerDay: undefined,
  defaultDurationDays: undefined,
  doseParams: [],
} satisfies Medicine;

describe("MedicineCategoryHeaderRow", () => {
  it("グループ切替はlabelとglyphを維持したまま44px以上の操作領域を持つ", async () => {
    const user = userEvent.setup();
    const onToggleGroup = vi.fn();
    const onEdit = vi.fn();

    render(
      <table>
        <tbody>
          <MedicineCategoryHeaderRow
            parentId={CATEGORY.id}
            header={CATEGORY}
            itemCount={27}
            isCollapsed={false}
            canCreate={false}
            canEdit={false}
            onToggleGroup={onToggleGroup}
            onEdit={onEdit}
            onCreate={vi.fn()}
          />
        </tbody>
      </table>,
    );

    const toggle = screen.getByRole("button", { name: /内用薬/ });
    expect(toggle).toHaveClass("min-h-11", "min-w-11");
    expect(toggle).toHaveTextContent("内用薬");
    expect(toggle.querySelector("svg")).toBeInTheDocument();

    await user.click(toggle);

    expect(onToggleGroup).toHaveBeenCalledWith(CATEGORY.id);
    expect(onEdit).not.toHaveBeenCalled();
  });

  it("行は非interactiveで、編集権限時だけ固有名の44px buttonを表示する", async () => {
    const user = userEvent.setup();
    const onEdit = vi.fn();
    const { rerender } = render(
      <table>
        <tbody>
          <MedicineCategoryHeaderRow
            parentId={CATEGORY.id}
            header={CATEGORY}
            itemCount={27}
            isCollapsed={false}
            canCreate={false}
            canEdit={false}
            onToggleGroup={vi.fn()}
            onEdit={onEdit}
            onCreate={vi.fn()}
          />
        </tbody>
      </table>,
    );

    const categoryRow = screen.getByText("内用薬").closest("tr");
    expect(categoryRow).not.toBeNull();
    await user.click(categoryRow!);
    expect(onEdit).not.toHaveBeenCalled();
    expect(
      screen.queryByRole("button", { name: "詳細: 薬剤カテゴリ 内用薬 (ID medicine-group-1)" }),
    ).not.toBeInTheDocument();

    rerender(
      <table>
        <tbody>
          <MedicineCategoryHeaderRow
            parentId={CATEGORY.id}
            header={CATEGORY}
            itemCount={27}
            isCollapsed={false}
            canCreate={false}
            canEdit
            onToggleGroup={vi.fn()}
            onEdit={onEdit}
            onCreate={vi.fn()}
          />
        </tbody>
      </table>,
    );

    const editButton = screen.getByRole("button", {
      name: "詳細: 薬剤カテゴリ 内用薬 (ID medicine-group-1)",
    });
    expect(editButton.tagName).toBe("BUTTON");
    expect(editButton).toHaveClass("min-h-11", "min-w-11");
    await user.click(editButton);
    expect(onEdit).toHaveBeenCalledWith(CATEGORY);
  });
});
