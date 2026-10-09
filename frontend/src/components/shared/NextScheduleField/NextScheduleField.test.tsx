import { fireEvent, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { NextScheduleField } from "./NextScheduleField";
import { calculateNextDate, resolveScheduleTypeAfterManualDate } from "./next-schedule";

describe("calculateNextDate", () => {
  it("3週後・4週後・1年後を算出する", () => {
    expect(calculateNextDate("2026-07-01", "3weeks")).toBe("2026-07-22");
    expect(calculateNextDate("2026-07-01", "4weeks")).toBe("2026-07-29");
    expect(calculateNextDate("2026-07-01", "1year")).toBe("2027-07-01");
  });

  it("以外（手動）と空基準日は空文字", () => {
    expect(calculateNextDate("2026-07-01", "other")).toBe("");
    expect(calculateNextDate("", "1year")).toBe("");
  });
});

describe("resolveScheduleTypeAfterManualDate", () => {
  it("計算結果と一致すれば種別を維持し、ずれれば other にする", () => {
    expect(resolveScheduleTypeAfterManualDate("2026-07-01", "1year", "2027-07-01")).toBe("1year");
    expect(resolveScheduleTypeAfterManualDate("2026-07-01", "1year", "2026-08-01")).toBe("other");
  });
});

describe("NextScheduleField", () => {
  it("共通ラベルとプルダウン＋日付を描画する", async () => {
    const user = userEvent.setup();
    const onScheduleTypeChange = vi.fn();
    const onNextDateChange = vi.fn();

    render(
      <NextScheduleField
        typeId="next-type"
        dateId="next-date"
        scheduleType="1year"
        nextDate="2027-07-01"
        onScheduleTypeChange={onScheduleTypeChange}
        onNextDateChange={onNextDateChange}
        dateAriaLabel="次回接種予定日"
      />,
    );

    expect(screen.getByText("次回の予定")).toBeInTheDocument();
    await user.click(screen.getByRole("combobox", { name: "次回の予定" }));
    await user.click(screen.getByRole("option", { name: "3週後" }));
    expect(onScheduleTypeChange).toHaveBeenCalledWith("3weeks");

    const nextDate = screen.getByLabelText("次回接種予定日");
    fireEvent.focus(nextDate);
    fireEvent.change(nextDate, { target: { value: "2026/08/13" } });
    fireEvent.blur(nextDate);
    expect(onNextDateChange).toHaveBeenCalledWith("2026-08-13");
  });

  // EMR-319: error があるとき内部 DatePicker に aria-invalid を伝播する。
  it("error 指定時は日付入力に aria-invalid が付きエラー文を表示する", () => {
    render(
      <NextScheduleField
        typeId="next-type"
        dateId="next-date"
        scheduleType="other"
        nextDate=""
        onScheduleTypeChange={vi.fn()}
        onNextDateChange={vi.fn()}
        dateAriaLabel="次回接種予定日"
        error="次回予定日を入力してください"
      />,
    );

    const dateInput = screen.getByLabelText("次回接種予定日");
    expect(dateInput).toHaveAttribute("aria-invalid", "true");
    const alert = screen.getByRole("alert");
    expect(alert).toHaveTextContent("次回予定日を入力してください");
    // PR#614 P3: 入力とエラー文を aria-describedby + 安定 id で関連付ける
    expect(alert).toHaveAttribute("id", "next-date-error");
    expect(dateInput).toHaveAttribute("aria-describedby", "next-date-error");
  });

  it("error なしなら日付入力に aria-invalid は付かずエラー文も出ない", () => {
    render(
      <NextScheduleField
        typeId="next-type"
        dateId="next-date"
        scheduleType="other"
        nextDate=""
        onScheduleTypeChange={vi.fn()}
        onNextDateChange={vi.fn()}
        dateAriaLabel="次回接種予定日"
      />,
    );

    expect(screen.getByLabelText("次回接種予定日")).not.toHaveAttribute("aria-invalid");
    expect(screen.getByLabelText("次回接種予定日")).not.toHaveAttribute("aria-describedby");
    expect(screen.queryByRole("alert")).toBeNull();
  });
});
