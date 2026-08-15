import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import type { TreatmentItem } from "@/lib/transforms/treatment";
import { TreatmentItemSidePanel } from "./TreatmentItemSidePanel";

function renderPanel({
  item = null,
  showAnesthesia = true,
  onSave = vi.fn(),
}: {
  item?: TreatmentItem | null;
  showAnesthesia?: boolean;
  onSave?: ReturnType<typeof vi.fn>;
} = {}) {
  render(
    <TreatmentItemSidePanel
      item={item}
      parentCandidates={[]}
      hasChildren={false}
      onClose={vi.fn()}
      onSave={onSave}
      showAnesthesia={showAnesthesia}
    />,
  );
  return { onSave };
}

describe("TreatmentItemSidePanel procedure anesthesia (BUG-028)", () => {
  it("renders a visible anesthesia select with four options for new procedure", async () => {
    const user = userEvent.setup();
    renderPanel({ showAnesthesia: true });

    const select = screen.getByRole("combobox", { name: "麻酔区分" });
    expect(select).toBeInTheDocument();

    await user.click(select);
    expect(screen.getByRole("option", { name: "麻酔なし" })).toBeInTheDocument();
    expect(screen.getByRole("option", { name: "局所麻酔" })).toBeInTheDocument();
    expect(screen.getByRole("option", { name: "鎮静" })).toBeInTheDocument();
    expect(screen.getByRole("option", { name: "全身麻酔" })).toBeInTheDocument();
  });

  it("defaults anesthesia to 麻酔なし and sends selected value on save", async () => {
    const user = userEvent.setup();
    const { onSave } = renderPanel({ showAnesthesia: true });

    const title = document.getElementById("master-title");
    expect(title).not.toBeNull();
    await user.type(title!, "V04処置テスト");

    await user.clear(screen.getByLabelText("単価(税込)"));
    await user.type(screen.getByLabelText("単価(税込)"), "500");

    await user.click(screen.getByRole("combobox", { name: "麻酔区分" }));
    await user.click(screen.getByRole("option", { name: "全身麻酔" }));

    await user.click(screen.getByRole("button", { name: "保存" }));

    expect(onSave).toHaveBeenCalledTimes(1);
    expect(onSave).toHaveBeenCalledWith(
      expect.objectContaining({
        name: "V04処置テスト",
        price: 500,
        anesthesia: "general",
      }),
    );
  });

  it("initializes anesthesia from existing item and sends change on edit save", async () => {
    const user = userEvent.setup();
    const item: TreatmentItem = {
      id: "100",
      name: "既存処置",
      price: 1000,
      isActive: true,
      description: "",
      sortOrder: 1,
      taxType: "excluded",
      taxRate: 0.1,
      anesthesia: "sedation",
    };
    const { onSave } = renderPanel({ item, showAnesthesia: true });

    // Select trigger should reflect existing value (鎮静)
    expect(screen.getByRole("combobox", { name: "麻酔区分" })).toHaveTextContent("鎮静");

    await user.click(screen.getByRole("combobox", { name: "麻酔区分" }));
    await user.click(screen.getByRole("option", { name: "局所麻酔" }));
    await user.click(screen.getByRole("button", { name: "保存" }));

    expect(onSave).toHaveBeenCalledWith(
      expect.objectContaining({
        name: "既存処置",
        anesthesia: "local",
      }),
    );
  });

  it("hides anesthesia field when showAnesthesia is false", () => {
    renderPanel({ showAnesthesia: false });
    expect(screen.queryByRole("combobox", { name: "麻酔区分" })).not.toBeInTheDocument();
  });

  it("shows Japanese field error and does not call onSave when price is negative", async () => {
    const user = userEvent.setup();
    const { onSave } = renderPanel({ showAnesthesia: true });

    const title = document.getElementById("master-title");
    if (title) {
      await user.type(title, "V04処置テスト");
    }

    const priceInput = screen.getByLabelText("単価(税込)");
    await user.clear(priceInput);
    await user.type(priceInput, "-100");

    await user.click(screen.getByRole("button", { name: "保存" }));

    expect(screen.getByText("金額は0以上を入力してください")).toBeInTheDocument();
    expect(priceInput).toHaveAttribute("aria-invalid", "true");
    expect(onSave).not.toHaveBeenCalled();
  });
});
