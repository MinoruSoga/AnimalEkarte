import { describe, expect, it, vi } from "vitest";
import { render } from "@testing-library/react";

import { ShiftCalendarPage } from "./ShiftCalendarPage";

vi.mock("@/hooks/use-auth", () => ({
  useAuth: () => ({
    hasPermission: () => true,
    isAuthenticated: true,
    isLoading: false,
  }),
}));

vi.mock("../components/ClinicHolidayModal/ClinicHolidayModal", () => ({
  ClinicHolidayModal: () => null,
}));

vi.mock("../api/get-shifts", () => ({
  useGetShifts: () => ({ data: [], isLoading: false, isError: false }),
}));

vi.mock("../api/get-staffs", () => ({
  OCCUPATION_FILTER_ALL: "all",
  filterStaffsByOccupation: <T,>(staffs: readonly T[]) => staffs,
  useGetStaffsForShift: () => ({
    data: [{ id: "s1", name: "スタッフA", occupationId: "1", occupationName: "医師" }],
    isLoading: false,
    isError: false,
  }),
}));

vi.mock("../api/clinic-holidays", () => ({
  useGetClinicHolidays: () => ({ data: [], isLoading: false, isError: false }),
}));

describe("ShiftCalendarPage scrollport", () => {
  it("コンテンツラッパーを伸ばさず、表の overflow-auto だけを両軸スクロールポートにする", () => {
    const { container } = render(<ShiftCalendarPage />);
    const content = container.querySelector(".px-3.py-6");
    expect(content).toBeInstanceOf(HTMLElement);
    if (!(content instanceof HTMLElement)) return;

    expect(content).toHaveClass("min-h-0");
    expect(content).not.toHaveClass("overflow-auto", "overflow-y-auto", "overflow-hidden");

    const scrolls = content.querySelectorAll(".overflow-auto");
    expect(scrolls).toHaveLength(1);
    expect(scrolls[0]).toHaveClass("min-h-0", "flex-1");
    expect(scrolls[0]?.parentElement).toHaveClass("min-h-0", "flex-1");
  });
});
