import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import type { Medicine } from "@/types";
import { MedicineDeleteDialog } from "./MedicineDeleteDialog";

const medicine: Medicine = {
  id: "42",
  parentId: undefined,
  dosageForm: undefined,
  medicineUnit: undefined,
  inventoryId: undefined,
  name: "合成薬剤",
  price: 500,
  defaultQuantity: 1,
  description: "",
  isActive: true,
  sortOrder: 1,
  taxType: "included",
  taxRate: 0.1,
  isNonInsurance: false,
  calculationType: "none",
  strength: undefined,
  frequencyPerDay: undefined,
  defaultDurationDays: undefined,
  doseParams: [],
  createdAt: "2026-10-14T00:00:00Z",
  updatedAt: "2026-10-14T00:00:00Z",
};

describe("MedicineDeleteDialog", () => {
  it("cancel closes without confirming and names the selected medicine", async () => {
    const user = userEvent.setup();
    const onClose = vi.fn();
    const onConfirm = vi.fn();
    render(
      <MedicineDeleteDialog open medicine={medicine} onClose={onClose} onConfirm={onConfirm} />,
    );

    expect(screen.getByRole("alertdialog")).toHaveTextContent("「合成薬剤」を削除します");
    await user.click(screen.getByRole("button", { name: "キャンセル" }));
    expect(onClose).toHaveBeenCalledOnce();
    expect(onConfirm).not.toHaveBeenCalled();
  });

  it("only confirms when the destructive action is selected", async () => {
    const user = userEvent.setup();
    const onClose = vi.fn();
    const onConfirm = vi.fn();
    render(
      <MedicineDeleteDialog open medicine={medicine} onClose={onClose} onConfirm={onConfirm} />,
    );

    await user.click(screen.getByRole("button", { name: "削除" }));
    expect(onConfirm).toHaveBeenCalledOnce();
    expect(onClose).toHaveBeenCalledOnce();
  });
});
