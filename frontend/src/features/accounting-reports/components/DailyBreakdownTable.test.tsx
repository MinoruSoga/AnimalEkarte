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

describe("DailyBreakdownTable drill-down", () => {
  it("renders an empty state when there are no details", () => {
    render(<DailyBreakdownTable details={[]} />);
    expect(screen.getByText("日次データがありません")).toBeInTheDocument();
  });

  it("makes a closed-day row a button that drills down with its date when onDrillDown is provided", async () => {
    const user = userEvent.setup();
    const onDrillDown = vi.fn();
    render(<DailyBreakdownTable details={[closedDay]} onDrillDown={onDrillDown} />);

    const row = screen.getByRole("button", { name: /2026-06-15.*締め/ });
    await user.click(row);

    expect(onDrillDown).toHaveBeenCalledTimes(1);
    expect(onDrillDown).toHaveBeenCalledWith("2026-06-15");
  });

  it("triggers drill-down on Enter key for accessibility", async () => {
    const user = userEvent.setup();
    const onDrillDown = vi.fn();
    render(<DailyBreakdownTable details={[closedDay]} onDrillDown={onDrillDown} />);

    const row = screen.getByRole("button", { name: /2026-06-15.*締め/ });
    row.focus();
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
