import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router";
import type { Estimate } from "../types";
import { EstimateDetail } from "./EstimateDetail";

const { mockEstimate, mockHasPermission } = vi.hoisted(() => ({
  mockEstimate: {
    current: null as Estimate | null,
  },
  mockHasPermission: vi.fn((_resource: string, _action: string) => true),
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

vi.mock("../api/get-estimate", () => ({
  useGetEstimate: () => ({
    data: mockEstimate.current,
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

function makeEstimate(status: Estimate["status"]): Estimate {
  return {
    id: "1",
    clinicId: "1",
    estimateNo: "EST-001",
    title: "テスト見積書",
    status,
    subtotal: 1000,
    taxTotal: 100,
    totalAmount: 1100,
    insuranceAmount: 0,
    discountAmount: 0,
    validUntil: null,
    comment: "",
    notes: "",
    items: [],
    createdAt: "2026-01-01T00:00:00Z",
    updatedAt: "2026-01-01T00:00:00Z",
  };
}

function renderDetail() {
  return render(
    <MemoryRouter initialEntries={["/estimates/1"]}>
      <Routes>
        <Route path="/estimates/:id" element={<EstimateDetail />} />
      </Routes>
    </MemoryRouter>,
  );
}

describe("EstimateDetail locked status actions", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockHasPermission.mockImplementation(() => true);
  });

  it("draft + 権限あり → 編集・削除ボタンを表示する", () => {
    mockEstimate.current = makeEstimate("draft");
    renderDetail();

    expect(screen.getByRole("button", { name: "編集" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "削除" })).toBeInTheDocument();
  });

  it("sent + 権限あり → 編集・削除ボタンを表示する", () => {
    mockEstimate.current = makeEstimate("sent");
    renderDetail();

    expect(screen.getByRole("button", { name: "編集" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "削除" })).toBeInTheDocument();
  });

  it("approved + 権限あり → 編集・削除ボタンを非表示にする", () => {
    mockEstimate.current = makeEstimate("approved");
    renderDetail();

    expect(screen.queryByRole("button", { name: "編集" })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "削除" })).not.toBeInTheDocument();
  });

  it("rejected + 権限あり → 編集・削除ボタンを非表示にする", () => {
    mockEstimate.current = makeEstimate("rejected");
    renderDetail();

    expect(screen.queryByRole("button", { name: "編集" })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "削除" })).not.toBeInTheDocument();
  });
});

describe("EstimateDetail mobile-first layout", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockHasPermission.mockImplementation(() => true);
  });

  it("基本情報は mobile で単一列、sm 以上で2列にする", () => {
    mockEstimate.current = makeEstimate("draft");
    renderDetail();

    const basicInfoGrid = screen.getByText("見積番号").parentElement?.parentElement;
    expect(basicInfoGrid).toHaveClass("grid-cols-1", "sm:grid-cols-2");
  });
});
