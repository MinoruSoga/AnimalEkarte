import type { ReactNode, Ref } from "react";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { transformBackendMedicineToFrontend } from "@/lib/transforms/medicine";

import { MEDICINE_STRENGTH_REQUIRED_MESSAGE } from "../lib/medicine-side-panel-model";
import type { MedicineDoseParamsEditorHandle } from "./MedicineDoseParamsEditor";
import { MedicineSidePanelBody } from "./MedicineSidePanelBody";

// 保存導線（MasterSidePanel の保存ボタン）と行レイアウトだけを差し替え、含量欄は実装を使う。
vi.mock("@/components/shared/SidePeek", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/components/shared/SidePeek")>();
  return {
    ...actual,
    MasterSidePanel: ({ children, onSave }: { children: ReactNode; onSave?: () => void }) => (
      <div>
        <button type="button" onClick={() => onSave?.()}>
          保存
        </button>
        {children}
      </div>
    ),
    PropertyRow: ({ children }: { children: ReactNode }) => <div>{children}</div>,
  };
});

vi.mock("./MedicineSidePanelSections", async (importOriginal) => {
  const actual = await importOriginal<typeof import("./MedicineSidePanelSections")>();
  return {
    ...actual,
    MedicineParentCategorySection: () => null,
    MedicinePriceTaxSection: () => null,
    MedicineBasicFlagsSection: () => null,
    MedicineDetailSection: () => null,
  };
});

vi.mock("./MedicineDoseParamsEditor", async () => {
  const { useImperativeHandle } = await import("react");
  return {
    MedicineDoseParamsEditor: ({ ref }: { ref?: Ref<MedicineDoseParamsEditorHandle> }) => {
      useImperativeHandle(ref, () => ({
        saveFilled: async () => true,
        collectFilled: async () => [],
      }));
      return null;
    },
  };
});

const perWeightMedicine = transformBackendMedicineToFrontend({
  id: 5,
  clinic_id: 1,
  name: "薬A",
  price: 100,
  is_active: true,
  description: "",
  dosage_form: "tablet",
  medicine_unit: "per_tablet",
  default_quantity: "1",
  tax_type: "excluded",
  tax_rate: "0.1",
  sort_order: 0,
  is_non_insurance: false,
  calculation_type: "per_weight",
  strength: "50",
  frequency_per_day: 2,
  created_at: "",
  updated_at: "",
});

function renderBody(onSave: () => boolean) {
  render(
    <MedicineSidePanelBody
      selectedMedicine={perWeightMedicine}
      isCategory={false}
      categoryMedicines={[]}
      onCloseEdit={vi.fn()}
      onSave={onSave}
      onDeleteRequest={vi.fn()}
    />,
  );
  return {
    strength: screen.getByLabelText("製品含量(mg/単位)"),
    frequency: screen.getByLabelText("1日投与回数"),
    save: screen.getByRole("button", { name: "保存" }),
  };
}

// SPEC-QA-20261022-447-STRENGTH-CLEAR（EMR-447 Q4）受入 fixture ①④⑥。
describe("MedicineSidePanelBody 製品含量の保存前検証 (EMR-447 Q4)", () => {
  it("① per_weight の含量を空にして保存すると欄に理由を出し、送信せず入力を保持する", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn(() => true);
    const { strength, frequency, save } = renderBody(onSave);

    await user.clear(strength);
    await user.clear(frequency);
    await user.type(frequency, "3");
    await user.click(save);

    expect(await screen.findByRole("alert")).toHaveTextContent(MEDICINE_STRENGTH_REQUIRED_MESSAGE);
    expect(strength).toHaveAttribute("aria-invalid", "true");
    expect(onSave).not.toHaveBeenCalled();
    expect(strength).toHaveValue(null);
    expect(frequency).toHaveValue(3);
  });

  it("含量を入力し直すと理由表示を解除し、入力どおりの含量で保存する", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn(() => true);
    const { strength, save } = renderBody(onSave);

    await user.clear(strength);
    await user.click(save);
    expect(await screen.findByRole("alert")).toBeInTheDocument();

    await user.type(strength, "60");
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
    await user.click(save);

    await waitFor(() => expect(onSave).toHaveBeenCalledTimes(1));
    expect(onSave).toHaveBeenCalledWith(expect.objectContaining({ strength: "60" }));
  });

  it("④ 含量を消して手動→体重換算へ戻すと削除前の値を復元せず、保存を止める", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn(() => true);
    const { strength, save } = renderBody(onSave);

    await user.clear(strength);
    await user.click(screen.getByRole("combobox"));
    await user.click(screen.getByRole("option", { name: "手動" }));
    expect(screen.queryByLabelText("製品含量(mg/単位)")).not.toBeInTheDocument();
    await user.click(screen.getByRole("combobox"));
    await user.click(screen.getByRole("option", { name: "体重換算(mg/kg)" }));

    expect(screen.getByLabelText("製品含量(mg/単位)")).toHaveValue(null);
    await user.click(save);
    expect(await screen.findByRole("alert")).toHaveTextContent(MEDICINE_STRENGTH_REQUIRED_MESSAGE);
    expect(onSave).not.toHaveBeenCalled();
  });

  it("⑥ 保存に失敗しても入力した含量を保持する", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn(() => false);
    const { strength, save } = renderBody(onSave);

    await user.clear(strength);
    await user.type(strength, "60");
    await user.click(save);

    await waitFor(() => expect(onSave).toHaveBeenCalledTimes(1));
    expect(strength).toHaveValue(60);
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
  });
});
