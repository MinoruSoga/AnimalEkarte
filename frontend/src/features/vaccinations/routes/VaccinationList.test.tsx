import { fireEvent, render, screen } from "@testing-library/react";
import { MemoryRouter, useLocation } from "react-router";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { C } from "@/lib/design-tokens";
import type { VaccinationRecord } from "../api/transforms";
import { VaccinationList } from "./VaccinationList";

const mockUseFilterVaccinations = vi.hoisted(() => vi.fn());

vi.mock("../hooks/use-vaccinations", () => ({
  useFilterVaccinations: mockUseFilterVaccinations,
}));

vi.mock("../api/delete-vaccination", () => ({
  useDeleteVaccination: vi.fn(() => ({ mutate: vi.fn() })),
}));

vi.mock("@/hooks/use-permission", () => ({
  usePermission: vi.fn(() => ({
    canView: true,
    canCreate: true,
    canEdit: true,
    canDelete: true,
  })),
}));

const vaccination: VaccinationRecord = {
  id: "vac-1",
  petId: "pet-1",
  ownerName: "山田太郎",
  petName: "ポチ",
  vaccineId: "vaccine-1",
  vaccineName: "混合ワクチン",
  doctor: "佐藤医師",
  date: "2026-07-13",
  nextDate: "2027-07-13",
  nextScheduleType: undefined,
  lot1: undefined,
  lot2: undefined,
  lot3: undefined,
  lot4: undefined,
  supplemental: undefined,
  remarks: undefined,
};

function LocationProbe() {
  const { pathname } = useLocation();
  return <output data-testid="location">{pathname}</output>;
}

function renderPage() {
  return render(
    <MemoryRouter initialEntries={["/vaccinations"]}>
      <VaccinationList />
      <LocationProbe />
    </MemoryRouter>,
  );
}

beforeEach(() => {
  mockUseFilterVaccinations.mockReturnValue({
    data: [vaccination],
    allVaccinations: [vaccination],
    isLoading: false,
    error: null,
  });
});

describe("VaccinationList row navigation accessibility", () => {
  it("ペット名・実施日を含む44px以上のnative detail linkを行内に表示する", () => {
    renderPage();

    const detailLink = screen.getByRole("link", { name: /ポチ/ });
    expect(detailLink).toHaveAttribute("href", "/vaccinations/vac-1");
    expect(detailLink).toHaveAccessibleName(/ポチ/);
    expect(detailLink).toHaveAccessibleName(/2026-07-13/);
    expect(detailLink).toHaveAccessibleName(/vac-1/);
    expect(detailLink).toHaveClass("min-h-11", "min-w-11");
  });

  it("detail link以外のセルclickでは行遷移しない", () => {
    renderPage();

    fireEvent.click(screen.getByText("山田太郎"));

    expect(screen.getByTestId("location")).toHaveTextContent(/^\/vaccinations$/);
  });

  it("行操作buttonのaccessible nameに予防接種IDを含める", () => {
    renderPage();

    expect(screen.getByRole("button", { name: /vac-1/ })).toBeInTheDocument();
  });
});

describe("VaccinationList 次回予定の期限超過表示", () => {
  beforeEach(() => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-07-23T15:30:00Z"));
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  function renderWithNextDate(nextDate: string) {
    mockUseFilterVaccinations.mockReturnValue({
      data: [{ ...vaccination, nextDate }],
      allVaccinations: [{ ...vaccination, nextDate }],
      isLoading: false,
      error: null,
    });
    renderPage();
  }

  function getNextDateCell() {
    const row = screen.getByText(vaccination.petName).closest("tr");
    return row?.querySelectorAll("td")[4] ?? null;
  }

  it("過去日だけをdanger表現で期限超過表示する", () => {
    renderWithNextDate("2026-07-23");

    const cell = getNextDateCell();
    expect(cell).toHaveClass("font-mono", C.danger);
    expect(cell).toHaveTextContent("2026-07-23");
    expect(cell).toHaveTextContent("（期限超過）");
  });

  it("未来日は現状の通常表示を維持する", () => {
    renderWithNextDate("2026-07-25");

    const cell = getNextDateCell();
    expect(cell).toHaveClass("font-mono", C.text);
    expect(cell).not.toHaveClass(C.danger);
    expect(cell).toHaveTextContent("2026-07-25");
    expect(cell).not.toHaveTextContent("期限超過");
  });

  it("空欄は現状の空セル表示を維持する", () => {
    renderWithNextDate("");

    const cell = getNextDateCell();
    expect(cell).toHaveClass("font-mono", C.text);
    expect(cell).not.toHaveClass(C.danger);
    expect(cell).toHaveTextContent(/^\s*$/);
    expect(cell).not.toHaveTextContent("期限超過");
  });
});
