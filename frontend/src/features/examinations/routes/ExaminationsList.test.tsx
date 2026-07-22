import { fireEvent, render, screen } from "@testing-library/react";
import { MemoryRouter, useLocation } from "react-router";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { ExaminationRecord } from "../api/transforms";
import { ExaminationsList } from "./ExaminationsList";

const mockUseFilterExaminationRecords = vi.hoisted(() => vi.fn());

vi.mock("../hooks/use-examination-records", () => ({
  useFilterExaminationRecords: mockUseFilterExaminationRecords,
}));

vi.mock("@/hooks/use-permission", () => ({
  usePermission: vi.fn(() => ({
    canView: true,
    canCreate: true,
    canEdit: true,
    canDelete: true,
  })),
}));

const examination: ExaminationRecord = {
  id: "exam-1",
  date: "2026-07-13",
  ownerName: "山田太郎",
  petName: "ポチ",
  petId: "pet-1",
  medicalRecordId: "mr-1",
  testType: "血液検査",
  testTypeId: "type-1",
  doctor: "佐藤医師",
  doctorId: "staff-1",
  status: "依頼中",
  resultSummary: undefined,
  machine: undefined,
  items: undefined,
};

function LocationProbe() {
  const { pathname } = useLocation();
  return <output data-testid="location">{pathname}</output>;
}

function renderPage() {
  return render(
    <MemoryRouter initialEntries={["/examinations"]}>
      <ExaminationsList />
      <LocationProbe />
    </MemoryRouter>,
  );
}

beforeEach(() => {
  mockUseFilterExaminationRecords.mockReturnValue({
    data: [examination],
    allExaminations: [examination],
    isLoading: false,
    error: null,
    total: 1,
    page: 1,
    limit: 20,
  });
});

describe("ExaminationsList row navigation accessibility", () => {
  it("ペット名・実施日を含む44px以上のnative detail linkを行内に表示する", () => {
    renderPage();

    const detailLink = screen.getByRole("link", { name: /ポチ/ });
    expect(detailLink).toHaveAttribute("href", "/examinations/exam-1");
    expect(detailLink).toHaveAccessibleName(/ポチ/);
    expect(detailLink).toHaveAccessibleName(/2026-07-13/);
    expect(detailLink).toHaveAccessibleName(/exam-1/);
    expect(detailLink).toHaveClass("min-h-11", "min-w-11");
  });

  it("detail link以外のセルclickでは行遷移しない", () => {
    renderPage();

    fireEvent.click(screen.getByText("山田太郎"));

    expect(screen.getByTestId("location")).toHaveTextContent(/^\/examinations$/);
  });

  it("行操作buttonのaccessible nameに検査IDを含める", () => {
    renderPage();

    expect(screen.getByRole("button", { name: /exam-1/ })).toBeInTheDocument();
  });
});
