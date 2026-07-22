import { fireEvent, render, screen } from "@testing-library/react";
import { MemoryRouter, useLocation } from "react-router";
import { beforeEach, describe, expect, it, vi } from "vitest";
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
