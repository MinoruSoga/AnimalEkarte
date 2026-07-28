import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import {
  BreakHoursEditor,
  ClosedDatesEditor,
  FieldRow,
  WeekdayHoursEditor,
} from "./LineReservationSettingsFormSections";

describe("LineReservationSettingsFormSections responsive layout", () => {
  it("FieldRowはmobileで1列、sm以上でラベルと入力の2列になる", () => {
    render(
      <FieldRow label="営業時間">
        <input aria-label="設定値" />
      </FieldRow>,
    );

    const row = screen.getByText("営業時間").parentElement;
    expect(row).toHaveClass("grid-cols-1", "sm:grid-cols-[200px_1fr]");
    expect(row).not.toHaveClass("grid-cols-[200px_1fr]");
  });

  it("休憩時間はmobileで縦積み・入力全幅、sm以上で横並びに戻る", () => {
    render(
      <BreakHoursEditor
        value={[{ start: "1200", end: "1300" }]}
        onChange={vi.fn()}
      />,
    );

    const inputs = screen.getAllByDisplayValue(/12:00|13:00/);
    const row = inputs[0].parentElement;
    expect(row).toHaveClass("flex-col", "items-stretch", "sm:flex-row", "sm:items-center");
    for (const input of inputs) {
      expect(input).toHaveClass("w-full", "sm:max-w-[120px]");
      expect(input).not.toHaveClass("max-w-[120px]");
    }
    expect(screen.getByRole("button", { name: "削除" })).toHaveClass("w-full", "sm:w-auto");
  });

  it("特定定休日はmobileで縦積み・入力全幅、sm以上で横並びに戻る", () => {
    render(<ClosedDatesEditor value={["2026-07-21"]} onChange={vi.fn()} />);

    const input = screen.getByDisplayValue("2026-07-21");
    expect(input.parentElement).toHaveClass(
      "flex-col",
      "items-stretch",
      "sm:flex-row",
      "sm:items-center",
    );
    expect(input).toHaveClass("w-full", "sm:max-w-[160px]");
    expect(input).not.toHaveClass("max-w-[160px]");
    expect(screen.getByRole("button", { name: "削除" })).toHaveClass("w-full", "sm:w-auto");
  });

  it("曜日別営業時間はmobileで縦積み・入力全幅、sm以上で横並びに戻る", () => {
    render(
      <WeekdayHoursEditor
        defaultHours={{ start: "0900", end: "1900" }}
        value={{}}
        onChange={vi.fn()}
      />,
    );

    const inputs = screen.getAllByDisplayValue(/09:00|19:00/);
    expect(inputs).toHaveLength(14);
    expect(inputs[0].parentElement).toHaveClass(
      "flex-col",
      "items-stretch",
      "sm:flex-row",
      "sm:items-center",
    );
    for (const input of inputs) {
      expect(input).toHaveClass("w-full", "sm:max-w-[120px]");
      expect(input).not.toHaveClass("max-w-[120px]");
    }
  });
});
