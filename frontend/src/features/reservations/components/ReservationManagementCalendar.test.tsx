import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { ReservationManagementCalendar } from "./ReservationManagementCalendar";

function renderCalendar() {
  render(
    <ReservationManagementCalendar
      currentDate={new Date("2026-07-21T00:00:00+09:00")}
      view="week"
      days={5}
      doctorFilter="all"
      sourceFilter="all"
      doctorNames={["山田医師"]}
      appointments={[]}
      activeEntries={[]}
      dynamicColorMap={new Map()}
      canCreate
      onViewChange={vi.fn()}
      onDoctorFilterChange={vi.fn()}
      onSourceFilterChange={vi.fn()}
      onDaysChange={vi.fn()}
      onNavigatePrevious={vi.fn()}
      onNavigateToday={vi.fn()}
      onNavigateNext={vi.fn()}
      onAppointmentClick={vi.fn()}
      onMonthDateClick={vi.fn()}
      onTimeSlotClick={vi.fn()}
      onAppointmentUpdate={vi.fn()}
    />,
  );
}

describe("ReservationManagementCalendar toolbar", () => {
  it("500pxではtoolbarとfilter controlsを縦積み・wrap可能にする", () => {
    renderCalendar();

    expect(screen.getByTestId("reservation-toolbar")).toHaveClass(
      "flex-col",
      "sm:flex-row",
      "items-stretch",
    );
    expect(screen.getByTestId("reservation-toolbar-filters")).toHaveClass(
      "w-full",
      "flex-wrap",
      "sm:w-auto",
    );
  });

  it("予約ソース・担当医・表示切替は44px以上の高さを持つ", () => {
    renderCalendar();

    const controls = screen.getAllByRole("combobox");
    expect(controls).toHaveLength(3);
    controls.forEach((control) => expect(control).toHaveClass("h-11"));
    expect(screen.getByRole("combobox", { name: "予約ソース" })).toBeInTheDocument();
    expect(screen.getByRole("combobox", { name: "担当医絞り込み" })).toBeInTheDocument();
    expect(screen.getByRole("combobox", { name: "カレンダー表示切替" })).toBeInTheDocument();
  });
});
