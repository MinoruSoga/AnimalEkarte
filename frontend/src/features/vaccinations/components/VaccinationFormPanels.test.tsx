import { describe, it, expect, vi } from "vitest";
import { fireEvent, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router";
import type { ComponentProps, ReactNode } from "react";
import type { VaccinationRecord } from "../api/transforms";
import {
  VaccinationFieldsPanel,
  VaccinationHistoryPanel,
} from "./VaccinationFormPanels";

// BUG-401: VaccinationFieldsPanel は以前 VACCINE_TYPE_ITEMS というハードコードの2択
// (value="1" 混合ワクチン / value="2" 狂犬病ワクチン) を描画しており、ワクチンマスタを
// 一切クエリしていなかった。実マスタの id=1/2 は別カテゴリ行（ワクチン犬/ワクチン猫）を
// 指すため、保存される vaccine_id が選択ラベルと一致しないデータ破損が発生していた
// （use-vaccination-form.test.ts の BUG-401 テストで根本原因を検証済み）。
// 本テストはコンポーネントが呼び出し元から渡された vaccineOptions（実マスタ由来）を
// そのまま描画し、旧ハードコード値を含まないことを固定する（このコンポーネントには
// 従来テストが一切無かった — investigation で確認済みのギャップを埋める）。

const noop = () => {};

type PanelProps = ComponentProps<typeof VaccinationFieldsPanel>;

const BASE_PANEL_PROPS = {
  doctorName: "",
  date: "",
  vaccineId: "",
  vaccineOptions: [],
  supplemental: "",
  lot1: "",
  lot2: "",
  lot3: "",
  lot4: "",
  nextScheduleType: "1year",
  nextDate: "",
  remarks: "",
  fieldErrors: {},
  onDateChange: noop,
  onVaccineIdChange: noop,
  onSupplementalChange: noop,
  onLot1Change: noop,
  onLot2Change: noop,
  onLot3Change: noop,
  onLot4Change: noop,
  onNextScheduleTypeChange: noop,
  onNextDateChange: noop,
  onRemarksChange: noop,
  onMarkDirty: noop,
} satisfies PanelProps;

function panel(overrides: Partial<PanelProps> = {}): ReactNode {
  return (
    <MemoryRouter>
      <VaccinationFieldsPanel
        {...BASE_PANEL_PROPS}
        {...overrides}
      />
    </MemoryRouter>
  );
}

function renderPanel(
  vaccineOptions: PanelProps["vaccineOptions"],
  overrides: Partial<PanelProps> = {},
) {
  return render(panel({ ...overrides, vaccineOptions }));
}

describe("VaccinationFieldsPanel — ワクチン選択 (BUG-401)", () => {
  it("編集recordの担当医を読み取り専用の臨床帰属として表示する", () => {
    renderPanel([], { doctorName: "合成監査スタッフ" });

    expect(screen.getByText("担当医")).toBeInTheDocument();
    expect(screen.getByText("合成監査スタッフ")).toBeInTheDocument();
  });

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

  it("編集recordがmaster optionsより先に到着しても既存選択を空値で上書きしない", () => {
    const onVaccineIdChange = vi.fn();
    const onMarkDirty = vi.fn();
    const result = renderPanel([], {
      vaccineId: "14",
      onVaccineIdChange,
      onMarkDirty,
    });

    result.rerender(panel({
      vaccineId: "14",
      vaccineOptions: [{ value: "14", label: "狂犬病ワクチン" }],
      onVaccineIdChange,
      onMarkDirty,
    }));

    expect(screen.getByRole("combobox", { name: /ワクチン/ })).toHaveTextContent("狂犬病ワクチン");
    expect(onVaccineIdChange).not.toHaveBeenCalled();
    expect(onMarkDirty).not.toHaveBeenCalled();
  });

  it("選択・補足・LOT・備考の変更をdirty通知とともに親へ渡す", async () => {
    const user = userEvent.setup();
    const onVaccineIdChange = vi.fn();
    const onSupplementalChange = vi.fn();
    const onLot1Change = vi.fn();
    const onLot2Change = vi.fn();
    const onLot3Change = vi.fn();
    const onLot4Change = vi.fn();
    const onRemarksChange = vi.fn();
    const onMarkDirty = vi.fn();
    renderPanel([{ value: "14", label: "狂犬病ワクチン" }], {
      onVaccineIdChange,
      onSupplementalChange,
      onLot1Change,
      onLot2Change,
      onLot3Change,
      onLot4Change,
      onRemarksChange,
      onMarkDirty,
    });

    await user.click(screen.getByRole("combobox", { name: /ワクチン/ }));
    await user.click(screen.getByRole("option", { name: "狂犬病ワクチン" }));
    fireEvent.change(screen.getByLabelText("補助説明"), { target: { value: "補足" } });
    fireEvent.change(screen.getByLabelText("LOT 1"), { target: { value: "L1" } });
    fireEvent.change(screen.getByLabelText("LOT 2"), { target: { value: "L2" } });
    fireEvent.change(screen.getByLabelText("LOT 3"), { target: { value: "L3" } });
    fireEvent.change(screen.getByLabelText("LOT 4"), { target: { value: "L4" } });
    fireEvent.change(screen.getByLabelText("備考"), { target: { value: "経過観察" } });

    expect(onVaccineIdChange).toHaveBeenCalledWith("14");
    expect(onSupplementalChange).toHaveBeenCalledWith("補足");
    expect(onLot1Change).toHaveBeenCalledWith("L1");
    expect(onLot2Change).toHaveBeenCalledWith("L2");
    expect(onLot3Change).toHaveBeenCalledWith("L3");
    expect(onLot4Change).toHaveBeenCalledWith("L4");
    expect(onRemarksChange).toHaveBeenCalledWith("経過観察");
    expect(onMarkDirty).toHaveBeenCalledTimes(7);
  });

  it("接種日・次回種別・次回日を親へ渡す", async () => {
    const user = userEvent.setup();
    const onDateChange = vi.fn();
    const onNextScheduleTypeChange = vi.fn();
    const onNextDateChange = vi.fn();
    const onMarkDirty = vi.fn();
    renderPanel([], {
      onDateChange,
      onNextScheduleTypeChange,
      onNextDateChange,
      onMarkDirty,
    });

    const date = screen.getByLabelText(/接種日/);
    fireEvent.focus(date);
    fireEvent.change(date, { target: { value: "2026/07/23" } });
    fireEvent.blur(date);

    await user.click(screen.getByRole("combobox", { name: "次回の予定" }));
    await user.click(screen.getByRole("option", { name: "3週後" }));

    const nextDate = screen.getByLabelText("次回接種予定日");
    fireEvent.focus(nextDate);
    fireEvent.change(nextDate, { target: { value: "2026/08/13" } });
    fireEvent.blur(nextDate);

    expect(onDateChange).toHaveBeenCalledWith("2026-07-23");
    expect(onNextScheduleTypeChange).toHaveBeenCalledWith("3weeks");
    expect(onNextDateChange).toHaveBeenCalledWith("2026-08-13");
    expect(onMarkDirty).toHaveBeenCalledTimes(3);
  });
});

const HISTORY_RECORD = {
  id: "vaccination-1",
  petId: "pet-1",
  ownerName: "合成監査飼主",
  petName: "合成監査ペット",
  vaccineId: "vaccine-1",
  vaccineName: "合成監査ワクチン",
  doctor: "合成監査スタッフ",
  date: "2026-07-23",
  nextDate: "2027-07-23",
  nextScheduleType: "1year",
  lot1: "SYN-LOT-1",
  lot2: undefined,
  lot3: undefined,
  lot4: undefined,
  supplemental: "合成補足",
  remarks: "合成備考",
} satisfies VaccinationRecord;

describe("VaccinationHistoryPanel", () => {
  it("履歴を描画し検索・clear・sortを親へ渡す", async () => {
    const user = userEvent.setup();
    const onHistorySearchTermChange = vi.fn();
    const onSortOrderChange = vi.fn();
    const onClear = vi.fn();
    render(
      <VaccinationHistoryPanel
        petHistory={[HISTORY_RECORD]}
        filterStartDate=""
        filterEndDate=""
        historySearchTerm=""
        sortOrder="desc"
        onFilterStartDateChange={noop}
        onFilterEndDateChange={noop}
        onHistorySearchTermChange={onHistorySearchTermChange}
        onSortOrderChange={onSortOrderChange}
        onClear={onClear}
      />,
    );

    expect(screen.getByText("合成監査ワクチン")).toBeInTheDocument();
    expect(screen.getByText(/接種: 2026-07-23/)).toBeInTheDocument();
    expect(screen.getByText(/担当医: 合成監査スタッフ/)).toBeInTheDocument();
    fireEvent.change(screen.getByLabelText("検索単語"), { target: { value: "監査" } });
    await user.click(screen.getByRole("button", { name: "クリア" }));
    await user.click(screen.getByRole("combobox", { name: "並び順" }));
    await user.click(screen.getByRole("option", { name: "古→新" }));

    expect(onHistorySearchTermChange).toHaveBeenCalledWith("監査");
    expect(onClear).toHaveBeenCalledOnce();
    expect(onSortOrderChange).toHaveBeenCalledWith("asc");
  });

  it("履歴0件を明示する", () => {
    render(
      <VaccinationHistoryPanel
        petHistory={[]}
        filterStartDate=""
        filterEndDate=""
        historySearchTerm=""
        sortOrder="desc"
        onFilterStartDateChange={noop}
        onFilterEndDateChange={noop}
        onHistorySearchTermChange={noop}
        onSortOrderChange={noop}
        onClear={noop}
      />,
    );

    expect(screen.getByText("履歴がありません")).toBeInTheDocument();
  });
});
