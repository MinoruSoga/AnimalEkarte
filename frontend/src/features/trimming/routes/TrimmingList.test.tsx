import { render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter, useLocation } from "react-router";
import { describe, expect, it, vi, beforeEach } from "vitest";
import type { TrimmingUI } from "@/types";
import { TrimmingList } from "./TrimmingList";

const mockUseFilterTrimmingRecords = vi.hoisted(() => vi.fn());

vi.mock("../hooks/use-trimming-records", () => ({
  useFilterTrimmingRecords: mockUseFilterTrimmingRecords,
}));

vi.mock("@/hooks/use-permission", () => ({
  usePermission: vi.fn(() => ({
    canView: true,
    canCreate: true,
    canEdit: true,
    canDelete: true,
  })),
}));

vi.mock("@/hooks/use-staff-validation", () => ({
  useStaffValidation: vi.fn(() => ({ isValidStaff: () => true })),
}));

const trimming: TrimmingUI = {
  id: "trim-1",
  reservationTypeId: "9",
  hasDetail: true,
  date: "2026-07-13",
  petId: "10",
  ownerId: "20",
  petNumber: "P-1",
  petName: "ポチ",
  ownerName: "ヤマダタロウ",
  species: "犬",
  breed: "",
  weight: "5.2",
  styleRequest: "",
  staff: "田中",
  status: "予約",
  staffId: "3",
  courseId: "4",
  courseName: "スタンダードコース",
  optionIds: [],
  bw: "",
  bwUnit: "Kg",
  bt: "",
  usedShampoo: "",
  usedRibbon: "",
  remarks: "",
  styleImage: undefined,
  completedImage: undefined,
};

function LocationProbe() {
  const { pathname, search } = useLocation();
  return <output data-testid="location">{`${pathname}${search}`}</output>;
}

function renderPage(initialEntry = "/trimming") {
  return render(
    <MemoryRouter initialEntries={[initialEntry]}>
      <TrimmingList />
      <LocationProbe />
    </MemoryRouter>,
  );
}

beforeEach(() => {
  mockUseFilterTrimmingRecords.mockReturnValue({
    data: [trimming],
    allTrimmings: [trimming],
    isTruncated: false,
    isLoading: false,
    error: null,
    deleteRecord: vi.fn(),
  });
});

describe("TrimmingList URL page 同期（FE-RC-028: useUrlPageSync）", () => {
  it("totalPagesを超えるURL pageは読み込み後にクランプされる", async () => {
    renderPage("/trimming?page=99");

    await waitFor(() => {
      expect(screen.getByTestId("location")).toHaveTextContent(/^\/trimming$/);
    });
  });

  it("range内のURL pageはそのまま維持される", async () => {
    renderPage("/trimming");

    await waitFor(() => {
      expect(screen.getByText("ポチ")).toBeInTheDocument();
    });
    expect(screen.getByTestId("location")).toHaveTextContent(/^\/trimming$/);
  });
});
