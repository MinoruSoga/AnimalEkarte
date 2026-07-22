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
});
