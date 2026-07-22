import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router";
import { VaccinationFieldsPanel } from "./VaccinationFormPanels";

// BUG-401: VaccinationFieldsPanel は以前 VACCINE_TYPE_ITEMS というハードコードの2択
// (value="1" 混合ワクチン / value="2" 狂犬病ワクチン) を描画しており、ワクチンマスタを
// 一切クエリしていなかった。実マスタの id=1/2 は別カテゴリ行（ワクチン犬/ワクチン猫）を
// 指すため、保存される vaccine_id が選択ラベルと一致しないデータ破損が発生していた
// （use-vaccination-form.test.ts の BUG-401 テストで根本原因を検証済み）。
// 本テストはコンポーネントが呼び出し元から渡された vaccineOptions（実マスタ由来）を
// そのまま描画し、旧ハードコード値を含まないことを固定する（このコンポーネントには
// 従来テストが一切無かった — investigation で確認済みのギャップを埋める）。

const noop = () => {};

function renderPanel(vaccineOptions: { value: string; label: string }[]) {
  return render(
    <MemoryRouter>
      <VaccinationFieldsPanel
        date=""
        vaccineId=""
        vaccineOptions={vaccineOptions}
        supplemental=""
        lot1=""
        lot2=""
        lot3=""
        lot4=""
        nextScheduleType="1year"
        nextDate=""
        remarks=""
        fieldErrors={{}}
        onDateChange={noop}
        onVaccineIdChange={noop}
        onSupplementalChange={noop}
        onLot1Change={noop}
        onLot2Change={noop}
        onLot3Change={noop}
        onLot4Change={noop}
        onNextScheduleTypeChange={noop}
        onNextDateChange={noop}
        onRemarksChange={noop}
        onMarkDirty={noop}
      />
    </MemoryRouter>,
  );
}

describe("VaccinationFieldsPanel — ワクチン選択 (BUG-401)", () => {
  it("LOT番号はmobileで1列、sm以上で2列になる", () => {
    renderPanel([]);

    const grid = screen.getByPlaceholderText("LOT 1番号").parentElement?.parentElement;
    expect(grid).toHaveClass("grid-cols-1", "sm:grid-cols-2");
    expect(grid).not.toHaveClass("grid-cols-2");
  });

  it("呼び出し元から渡された vaccineOptions（実マスタ由来）を選択肢として描画する", async () => {
    const user = userEvent.setup();
    renderPanel([
      { value: "11", label: "混合ワクチン5種（犬）" },
      { value: "14", label: "狂犬病ワクチン" },
    ]);

    await user.click(screen.getByRole("combobox", { name: /ワクチン/ }));

    expect(await screen.findByRole("option", { name: "混合ワクチン5種（犬）" })).toBeInTheDocument();
    expect(await screen.findByRole("option", { name: "狂犬病ワクチン" })).toBeInTheDocument();
  });

  it("旧ハードコードのラベル（value='1'→混合ワクチン, value='2'→狂犬病ワクチン のダミー2択）を描画しない", async () => {
    const user = userEvent.setup();
    // 実マスタが返す本物のオプションのみを渡す。旧ハードコードの短縮ラベル「混合ワクチン」
    // （5種等の修飾なし）は実マスタに存在しないため、描画されてはならない。
    renderPanel([{ value: "14", label: "狂犬病ワクチン" }]);

    await user.click(screen.getByRole("combobox", { name: /ワクチン/ }));

    expect(screen.queryByRole("option", { name: "混合ワクチン" })).not.toBeInTheDocument();
  });

  it("vaccineOptions が空のときも選択欄自体はエラーなく描画される（マスタ取得中/未取得の状態）", () => {
    renderPanel([]);
    expect(screen.getByRole("combobox", { name: /ワクチン/ })).toBeInTheDocument();
  });
});
