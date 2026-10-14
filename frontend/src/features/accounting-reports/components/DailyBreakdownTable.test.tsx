import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { DailyBreakdownTable } from "./DailyBreakdownTable";
import type { DailyReportDetail } from "../api/get-monthly-report";

const closedDay: DailyReportDetail = {
  date: "2026-06-15",
  weekday: "月",
  amCount: 3,
  amNet: 30000,
  pmCount: 5,
  pmNet: 50000,
  dayNet: 80000,
  refund: 0,
  amClosed: true,
  pmClosed: true,
  isHoliday: false,
};

const openDay: DailyReportDetail = {
  date: "2026-06-16",
  weekday: "火",
  amCount: 2,
  amNet: 20000,
  pmCount: 1,
  pmNet: 10000,
  dayNet: 30000,
  refund: 0,
  amClosed: false,
  pmClosed: false,
  isHoliday: false,
};

const auditDay: DailyReportDetail = {
  date: "2026-06-20",
  weekday: "土",
  amCount: 2,
  amNet: 2200,
  pmCount: 3,
  pmNet: 3300,
  dayNet: 5500,
  refund: 0,
  amClosed: true,
  pmClosed: false,
  isHoliday: false,
};

describe("DailyBreakdownTable drill-down", () => {
  it("formats displayed dates and preserves zero accounting counts", () => {
    render(<DailyBreakdownTable details={[{ ...closedDay, amCount: 0, pmCount: 0 }]} />);
    expect(screen.getByRole("cell", { name: "2026/06/15" })).toHaveClass("whitespace-nowrap");
    expect(screen.getByRole("cell", { name: "0件" })).toHaveClass("whitespace-nowrap");
  });
  it("renders an empty state when there are no details", () => {
    render(<DailyBreakdownTable details={[]} />);
    expect(screen.getByText("日次データがありません")).toBeInTheDocument();
  });

  it("renders a date-cell button on a closed day that drills down with its date", async () => {
    const user = userEvent.setup();
    const onDrillDown = vi.fn();
    render(<DailyBreakdownTable details={[closedDay]} onDrillDown={onDrillDown} />);

    const button = screen.getByRole("button", { name: /2026\/06\/15.*締め/ });
    await user.click(button);

    expect(onDrillDown).toHaveBeenCalledTimes(1);
    expect(onDrillDown).toHaveBeenCalledWith("2026-06-15");
  });

  it("triggers drill-down on Enter key for accessibility", async () => {
    const user = userEvent.setup();
    const onDrillDown = vi.fn();
    render(<DailyBreakdownTable details={[closedDay]} onDrillDown={onDrillDown} />);

    const button = screen.getByRole("button", { name: /2026\/06\/15.*締め/ });
    button.focus();
    await user.keyboard("{Enter}");

    expect(onDrillDown).toHaveBeenCalledWith("2026-06-15");
  });

  it("does not make a day without any closing interactive", () => {
    const onDrillDown = vi.fn();
    render(<DailyBreakdownTable details={[openDay]} onDrillDown={onDrillDown} />);

    expect(screen.queryByRole("button", { name: /2026-06-16/ })).not.toBeInTheDocument();
  });

  it("renders no interactive rows when onDrillDown is omitted", () => {
    render(<DailyBreakdownTable details={[closedDay]} />);

    expect(screen.queryByRole("button")).not.toBeInTheDocument();
  });
});

describe("DailyBreakdownTable 会計件数列と締め表示 (EMR-401)", () => {
  it("午前/午後の件数・売上4列を 会計件数 1列に合算し、日計を保持する", () => {
    render(<DailyBreakdownTable details={[auditDay]} />);

    for (const heading of ["午前件数", "午前売上", "午後件数", "午後売上"]) {
      expect(screen.queryByRole("columnheader", { name: heading })).not.toBeInTheDocument();
    }
    expect(screen.getByRole("columnheader", { name: "会計件数" })).toBeInTheDocument();
    expect(screen.getByText("5件")).toBeInTheDocument();
    expect(screen.getByRole("columnheader", { name: "日計" })).toBeInTheDocument();
    expect(screen.getByText("¥5,500")).toBeInTheDocument();
  });

  it("締め状態をドットではなく 締め済 / 未締め のテキストで表示する", () => {
    render(<DailyBreakdownTable details={[auditDay]} />);

    expect(screen.getByText("締め済")).toBeInTheDocument();
    expect(screen.getByText("未締め")).toBeInTheDocument();
  });

  it("締めのある行の日付ボタンは表示日付（aria-label）と同じ形式でドリルダウンする", async () => {
    const user = userEvent.setup();
    const onDrillDown = vi.fn();
    render(<DailyBreakdownTable details={[auditDay]} onDrillDown={onDrillDown} />);

    await user.click(screen.getByRole("button", { name: /2026\/06\/20.*締め/ }));

    expect(onDrillDown).toHaveBeenCalledWith("2026-06-20");
  });
});
