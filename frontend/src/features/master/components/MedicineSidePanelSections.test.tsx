import { useState } from "react";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it } from "vitest";

import type { MedicineFormData } from "../lib/medicine-side-panel-model";
import { MedicineBasicFlagsSection } from "./MedicineSidePanelSections";

const INITIAL: MedicineFormData = {
  name: "",
  parentId: "",
  dosageForm: "tablet",
  medicineUnit: "per_tablet",
  price: 0,
  description: "",
  isActive: true,
  taxType: "included",
  taxRate: 0.1,
  isNonInsurance: false,
  calculationType: "none",
  strength: "",
  frequencyPerDay: "",
  defaultDurationDays: "",
};

/**
 * EMR-411 C-2 (Q2): 保険対象外トグルは yes/no 表示（はい=true / いいえ=false）にし、
 * 現在値を aria-pressed で公開する。aria-label は維持。
 */
describe("MedicineBasicFlagsSection 保険対象外トグル (EMR-411 C-2)", () => {
  function Stateful({ initial }: { initial: MedicineFormData }) {
    const [formData, setFormData] = useState(initial);
    return <MedicineBasicFlagsSection formData={formData} setFormDataDirty={setFormData} />;
  }

  it("いいえ→はい→いいえ と aria-pressed が連動する", async () => {
    const user = userEvent.setup();
    render(<Stateful initial={INITIAL} />);

    const toggle = screen.getByRole("button", { name: "保険対象外を切り替え" });
    expect(toggle).toHaveAttribute("aria-pressed", "false");
    expect(toggle).toHaveTextContent("いいえ");

    await user.click(toggle);
    expect(toggle).toHaveAttribute("aria-pressed", "true");
    expect(toggle).toHaveTextContent("はい");

    await user.click(toggle);
    expect(toggle).toHaveAttribute("aria-pressed", "false");
    expect(toggle).toHaveTextContent("いいえ");
  });

  it("初期値が対象外(true)の場合は「はい」と pressed=true で表示する", () => {
    render(<Stateful initial={{ ...INITIAL, isNonInsurance: true }} />);

    const toggle = screen.getByRole("button", { name: "保険対象外を切り替え" });
    expect(toggle).toHaveAttribute("aria-pressed", "true");
    expect(toggle).toHaveTextContent("はい");
  });
});
