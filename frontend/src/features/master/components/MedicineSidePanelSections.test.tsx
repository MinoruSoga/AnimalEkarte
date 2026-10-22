import { useState } from "react";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import type { MedicineFormData } from "../lib/medicine-side-panel-model";
import { MedicineBasicFlagsSection, MedicinePriceTaxSection } from "./MedicineSidePanelSections";

// Radix Select の開閉に依存せず、onChange 結線（EMR-536 の編集 intent）だけを検証する。
// PropertyRow の <label> が子 button の accessible name になる（"税率 税率を8%にする"）ため、
// 取得は部分一致の正規表現で行う。
vi.mock("@/components/shared/TaxRateSelector/TaxRateSelector", () => ({
  TaxRateSelector: ({ onChange }: { onChange: (value: number) => void }) => (
    <button type="button" onClick={() => onChange(0.08)}>
      税率を8%にする
    </button>
  ),
}));

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

/**
 * EMR-536: 税率セレクタの onChange だけが taxRateTouched を立てる。
 * 他項目の編集では立たない（未操作の保存 exact 税率は PATCH 省略で保持される）。
 */
describe("MedicinePriceTaxSection 税率の編集 intent (EMR-536)", () => {
  function Stateful() {
    const [formData, setFormData] = useState(INITIAL);
    return (
      <>
        <MedicinePriceTaxSection
          formData={formData}
          isCategory={false}
          setFormDataDirty={setFormData}
        />
        <output data-testid="state">
          {JSON.stringify([formData.taxRate, formData.taxRateTouched])}
        </output>
      </>
    );
  }

  it("税率セレクタ操作で taxRate と taxRateTouched=true を設定する", async () => {
    const user = userEvent.setup();
    render(<Stateful />);
    expect(screen.getByTestId("state")).toHaveTextContent("[0.1,null]");

    await user.click(screen.getByRole("button", { name: /税率を8%にする/ }));

    expect(screen.getByTestId("state")).toHaveTextContent("[0.08,true]");
  });
});
