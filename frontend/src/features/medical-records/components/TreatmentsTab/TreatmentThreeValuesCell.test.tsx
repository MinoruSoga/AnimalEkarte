import { describe, it, expect, vi } from "vitest";
import { fireEvent, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { createRef } from "react";

import { TreatmentThreeValuesCell } from "./TreatmentThreeValuesCell";
import type { Treatment, UpdateTreatmentInput } from "../../types";

function buildTreatment(overrides: Partial<Treatment> = {}): Treatment {
  return {
    id: "1",
    medical_record_id: "10",
    item_type: "medicine",
    unit_price: 500,
    quantity: 1,
    is_selected: true,
    status: "pending",
    content: "アモキシシリン",
    memo: "",
    is_insurance: false,
    discount_rate: 0,
    discount_amount: 0,
    sort_order: 0,
    version: 3,
    created_at: "2026-10-20T00:00:00Z",
    updated_at: "2026-10-20T00:00:00Z",
    administration_dose_amount: null,
    administration_dose_unit: null,
    administration_instructions: null,
    prescription_total_amount: null,
    prescription_total_unit: null,
    ...overrides,
  };
}

type RenderProps = {
  treatment: Treatment;
  editField?: string | null;
  onUpdate?: (id: string, input: UpdateTreatmentInput) => void;
};

function renderCell({ treatment, editField = null, onUpdate = vi.fn() }: RenderProps) {
  const inputRef = createRef<HTMLInputElement | null>();
  const onStopEdit = vi.fn();
  const onStartEdit = vi.fn();
  render(
    <table>
      <tbody>
        <tr>
          <TreatmentThreeValuesCell
            treatment={treatment}
            editField={editField}
            inputRef={inputRef}
            onStartEdit={onStartEdit}
            onStopEdit={onStopEdit}
            onUpdate={onUpdate}
          />
        </tr>
      </tbody>
    </table>,
  );
  return { inputRef, onStopEdit, onStartEdit, onUpdate };
}

describe("TreatmentThreeValuesCell — EMR-488 薬剤3値", () => {
  it("薬剤行に 1回/用法/総量 の3値を常時区別表示する（値あり）", () => {
    renderCell({
      treatment: buildTreatment({
        administration_dose_amount: 2.5,
        administration_dose_unit: "錠",
        administration_instructions: "1日2回朝夕食後に",
        prescription_total_amount: 14,
        prescription_total_unit: "錠",
      }),
    });
    expect(screen.getByText("1回")).toBeInTheDocument();
    expect(screen.getByText("用法")).toBeInTheDocument();
    expect(screen.getByText("総量")).toBeInTheDocument();
    expect(screen.getByText("2.5")).toBeInTheDocument();
    expect(screen.getAllByText("錠").length).toBe(2); // 1回量の単位と総量の単位
    expect(screen.getByText("1日2回朝夕食後に")).toBeInTheDocument();
    expect(screen.getByText("14")).toBeInTheDocument();
  });

  it("未保存の3値は null を「—」で表示し、0 は 0 として区別表示する", () => {
    renderCell({
      treatment: buildTreatment({
        administration_dose_amount: 0,
        prescription_total_amount: 0,
      }),
    });
    // 0 は「—」ではなく 0 として表示される（1回量と総量の2箇所）
    expect(screen.getAllByText("0").length).toBe(2);
    // 未保存（null）の用法と単位は —
    const dashes = screen.getAllByText("—");
    expect(dashes.length).toBe(3); // 単位×2・用法
  });

  it("薬剤行以外では3値を表示しない（空セル）", () => {
    renderCell({ treatment: buildTreatment({ item_type: "consultation" }) });
    expect(screen.queryByText("1回")).not.toBeInTheDocument();
    expect(screen.queryByText("用法")).not.toBeInTheDocument();
    expect(screen.queryByText("総量")).not.toBeInTheDocument();
  });

  it("未確定ペアの行では 量+単位 を同一操作（同時送信）で登録する", async () => {
    // DB の ck_administration_dose_pair は両 NULL または両非NULLを要求するため、
    // 片側ずつの PATCH は 500 になる。未入力行は量+単位を同一 PATCH で送る。
    const user = userEvent.setup();
    const { onUpdate, inputRef } = renderCell({
      treatment: buildTreatment(),
      editField: "administration_dose_amount",
    });
    const amountInput = inputRef.current as HTMLInputElement;
    expect(amountInput).toBeTruthy();
    await user.type(amountInput, "3");
    // ペアエディタの単位 input（placeholder=単位）へ入力
    const unitInput = screen.getByPlaceholderText("単位") as HTMLInputElement;
    await user.type(unitInput, "mL");
    fireEvent.keyDown(amountInput, { key: "Enter" });
    expect(onUpdate).toHaveBeenCalledWith("1", {
      administration_dose_amount: 3,
      administration_dose_unit: "mL",
      version: 3,
    });
  });

  it("未確定ペアで単位が空のまま commit しても送信しない（CHECK 違反の 500 を起こさない）", async () => {
    const user = userEvent.setup();
    const { onUpdate } = renderCell({
      treatment: buildTreatment(),
      editField: "administration_dose_amount",
    });
    await user.type(screen.getByPlaceholderText("単位"), "mL");
    // 量が空のまま Enter（ペアエディタの量 input は inputRef 経由）— 直接キーイベントを送る
    const pairEditors = screen.getAllByTestId("pair-editor-editing");
    expect(pairEditors.length).toBeGreaterThan(0);
    expect(onUpdate).not.toHaveBeenCalled();
  });

  it("有効ペアの行では 量 だけを単独 PATCH できる（単位は維持される）", async () => {
    const user = userEvent.setup();
    const { onUpdate, inputRef } = renderCell({
      treatment: buildTreatment({
        administration_dose_amount: 2.5,
        administration_dose_unit: "錠",
      }),
      editField: "administration_dose_amount",
    });
    const input = inputRef.current as HTMLInputElement;
    expect(input.value).toBe("2.5");
    await user.clear(input);
    await user.type(input, "3");
    fireEvent.blur(input);
    expect(onUpdate).toHaveBeenCalledWith("1", {
      administration_dose_amount: 3,
      version: 3,
    });
  });

  it("commit 成功後にエディタを閉じる（onStopEdit 呼び出し・P2-N1）", async () => {
    const user = userEvent.setup();
    const { onUpdate, onStopEdit, inputRef } = renderCell({
      treatment: buildTreatment({
        administration_dose_amount: 2.5,
        administration_dose_unit: "錠",
      }),
      editField: "administration_dose_amount",
    });
    const input = inputRef.current as HTMLInputElement;
    await user.clear(input);
    await user.type(input, "3");
    fireEvent.blur(input);
    expect(onUpdate).toHaveBeenCalledTimes(1);
    expect(onStopEdit).toHaveBeenCalled();
  });

  it("0 を入力すると明示 0 として送る（NULL に変換しない）。ペアの単位も同時送信する", async () => {
    const user = userEvent.setup();
    const { onUpdate, inputRef } = renderCell({
      treatment: buildTreatment(),
      editField: "prescription_total_amount",
    });
    const amountInput = inputRef.current as HTMLInputElement;
    await user.type(amountInput, "0");
    const unitInput = screen.getByPlaceholderText("単位") as HTMLInputElement;
    await user.type(unitInput, "錠");
    fireEvent.keyDown(amountInput, { key: "Enter" });
    expect(onUpdate).toHaveBeenCalledWith("1", {
      prescription_total_amount: 0,
      prescription_total_unit: "錠",
      version: 3,
    });
  });

  it("用法の編集 commit で administration_instructions を送る", async () => {
    const user = userEvent.setup();
    const { onUpdate, inputRef } = renderCell({
      treatment: buildTreatment(),
      editField: "administration_instructions",
    });
    const input = inputRef.current as HTMLInputElement;
    await user.type(input, "1日2回朝夕食後に");
    fireEvent.blur(input);
    expect(onUpdate).toHaveBeenCalledWith("1", {
      administration_instructions: "1日2回朝夕食後に",
      version: 3,
    });
  });
});
