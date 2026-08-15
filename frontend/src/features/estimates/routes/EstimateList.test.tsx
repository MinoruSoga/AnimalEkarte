import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes } from "react-router";
import type { Estimate } from "../types";
import { EstimateList } from "./EstimateList";

const { mockEstimates, mockHasPermission, mockNavigate } = vi.hoisted(() => ({
  mockEstimates: {
    current: [] as Estimate[],
  },
  mockHasPermission: vi.fn((_resource: string, _action: string) => true),
  mockNavigate: vi.fn(),
}));

vi.mock("@/hooks/use-auth", () => ({
  useAuth: () => ({
    user: {
      clinics: [{ clinicId: "1", clinicName: "テスト動物病院", isMain: true }],
      clinic: { name: "テスト動物病院" },
    },
    currentClinicId: "1",
    hasPermission: mockHasPermission,
  }),
}));

vi.mock("react-router", async () => {
  const actual = await vi.importActual<typeof import("react-router")>("react-router");
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  };
});

vi.mock("../api/get-estimates", () => ({
  useGetEstimates: () => ({
    data: {
      data: mockEstimates.current,
      total: mockEstimates.current.length,
      page: 1,
      limit: 50,
    },
    isLoading: false,
    isError: false,
  }),
}));

vi.mock("../api/delete-estimate", () => ({
  useDeleteEstimate: () => ({
    mutate: vi.fn(),
    isPending: false,
  }),
}));

vi.mock("@/components/shared/PropertyFilter/PropertyFilter", () => ({
  PropertyFilter: () => null,
}));

vi.mock("@/components/shared/Pagination/Pagination", () => ({
  Pagination: () => null,
}));

function makeEstimate(status: Estimate["status"], id = "1"): Estimate {
  return {
    id,
    clinicId: "1",
    estimateNo: `EST-00${id}`,
    medicalRecordId: null,
    title: "テスト見積書",
    ownerId: null,
    ownerName: undefined,
    status,
    subtotal: 1000,
    taxTotal: 100,
    totalAmount: 1100,
    insuranceAmount: 0,
    discountAmount: 0,
    validUntil: null,
    comment: "",
    notes: "",
    createdBy: null,
    items: [],
    createdAt: "2026-01-01T00:00:00Z",
    updatedAt: "2026-01-01T00:00:00Z",
  };
}

function renderList() {
  return render(
    <MemoryRouter initialEntries={["/estimates"]}>
      <Routes>
        <Route path="/estimates" element={<EstimateList />} />
      </Routes>
    </MemoryRouter>,
  );
}

async function openRowActions() {
  const user = userEvent.setup();
  await user.click(screen.getByRole("button", {
    name: /見積書.*EST-001.*テスト見積書.*ID: 1.*操作/,
  }));
  return user;
}

describe("EstimateList locked status row actions", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockHasPermission.mockImplementation(() => true);
  });

  it("draft + 権限あり → 編集・削除メニューを表示する", async () => {
    mockEstimates.current = [makeEstimate("draft")];
    renderList();

    await openRowActions();

    expect(await screen.findByRole("menuitem", { name: /詳細/ })).toBeInTheDocument();
    expect(screen.getByRole("menuitem", { name: /編集/ })).toBeInTheDocument();
    expect(screen.getByRole("menuitem", { name: /削除/ })).toBeInTheDocument();
  });

  it("sent + 権限あり → 編集・削除メニューを表示する", async () => {
    mockEstimates.current = [makeEstimate("sent")];
    renderList();

    await openRowActions();

    expect(await screen.findByRole("menuitem", { name: /詳細/ })).toBeInTheDocument();
    expect(screen.getByRole("menuitem", { name: /編集/ })).toBeInTheDocument();
    expect(screen.getByRole("menuitem", { name: /削除/ })).toBeInTheDocument();
  });

  it("approved + 権限あり → 編集・削除を非表示にし詳細は残す", async () => {
    mockEstimates.current = [makeEstimate("approved")];
    renderList();

    await openRowActions();

    expect(await screen.findByRole("menuitem", { name: /詳細/ })).toBeInTheDocument();
    expect(screen.queryByRole("menuitem", { name: /編集/ })).not.toBeInTheDocument();
    expect(screen.queryByRole("menuitem", { name: /削除/ })).not.toBeInTheDocument();
  });

  it("rejected + 権限あり → 編集・削除を非表示にし詳細は残す", async () => {
    mockEstimates.current = [makeEstimate("rejected")];
    renderList();

    await openRowActions();

    expect(await screen.findByRole("menuitem", { name: /詳細/ })).toBeInTheDocument();
    expect(screen.queryByRole("menuitem", { name: /編集/ })).not.toBeInTheDocument();
    expect(screen.queryByRole("menuitem", { name: /削除/ })).not.toBeInTheDocument();
  });

  it("approved 行に見積番号・タイトル・IDを含む44px native detail linkを表示する", () => {
    mockEstimates.current = [makeEstimate("approved", "42")];
    renderList();

    const detailLink = screen.getByRole("link", { name: /EST-0042/ });
    expect(detailLink).toHaveAttribute("href", "/estimates/42");
    expect(detailLink).toHaveAccessibleName(/EST-0042/);
    expect(detailLink).toHaveAccessibleName(/テスト見積書/);
    expect(detailLink).toHaveAccessibleName(/ID:? 42/);
    expect(detailLink).toHaveClass("min-h-11", "min-w-11");
  });

  it("detail link以外のセルclickでは行遷移しない", async () => {
    mockEstimates.current = [makeEstimate("approved", "42")];
    renderList();

    const user = userEvent.setup();
    await user.click(screen.getByText("テスト見積書"));

    expect(mockNavigate).not.toHaveBeenCalled();
  });
});
