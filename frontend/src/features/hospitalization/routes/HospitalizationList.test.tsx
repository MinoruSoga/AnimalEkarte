import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { MemoryRouter } from "react-router";
import { AuthContext } from "@/hooks/auth-context";
import type { ResourceAction } from "@/types/auth";
import { HospitalizationList } from "./HospitalizationList";

// ---- mocks ----

vi.mock("../api/get-hospitalizations", () => ({
  useGetHospitalizations: vi.fn(),
}));

vi.mock("@/hooks/use-permission", () => ({
  usePermission: vi.fn(() => ({ canCreate: true, canEdit: true })),
}));

vi.mock("@/hooks/use-master-items", () => ({
  useMasterItems: vi.fn(() => ({ data: [], isLoading: false, error: null })),
}));

import { useGetHospitalizations } from "../api/get-hospitalizations";

// ---- helpers ----

function makeAuthCtx() {
  return {
    user: null,
    currentClinicId: "clinic-1",
    isAuthenticated: true,
    isLoading: false,
    login: async () => {},
    logout: async () => {},
    switchClinic: () => {},
    hasPermission: (_resource: string, _action: ResourceAction): boolean => true,
    refreshPermissions: async () => {},
  };
}

function makeHosp(overrides: Partial<{
  id: string;
  petName: string;
  ownerName: string;
  hospitalizationNo: string;
  status: string;
  hospitalizationType: string;
  cageId: string;
  startDate: string;
  endDate: string | undefined;
  species: string;
  petIsDeceased: boolean;
}> = {}) {
  return {
    id: "h-1",
    petName: "ポチ",
    ownerName: "ヤマダ",
    hospitalizationNo: "H-001",
    status: "入院中",
    hospitalizationType: "short_term",
    cageId: "cage-1",
    startDate: "2026-01-01",
    endDate: undefined,
    species: "犬",
    petIsDeceased: false,
    ...overrides,
  };
}

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return ({ children }: { children: React.ReactNode }) => (
    <AuthContext.Provider value={makeAuthCtx()}>
      <QueryClientProvider client={queryClient}>
        <MemoryRouter>{children}</MemoryRouter>
      </QueryClientProvider>
    </AuthContext.Provider>
  );
}

beforeEach(() => {
  localStorage.setItem("auth_current_clinic:v1", "clinic-1");
  vi.mocked(useGetHospitalizations).mockReturnValue({
    data: [],
    isLoading: false,
    isError: false,
  } as ReturnType<typeof useGetHospitalizations>);
});

// ─────────────────────────────────────────────────────────────
// かな正規化テキスト検索 (list view で検証)
// ─────────────────────────────────────────────────────────────

describe("HospitalizationList — かな正規化テキスト検索", () => {
  it("ひらがな入力でカタカナ petName がヒットする", async () => {
    vi.mocked(useGetHospitalizations).mockReturnValue({
      data: [
        makeHosp({ id: "1", petName: "ポチ", ownerName: "ヤマダ", hospitalizationNo: "H-001" }),
        makeHosp({ id: "2", petName: "たろう", ownerName: "さとう", hospitalizationNo: "H-002" }),
      ],
      isLoading: false,
      isError: false,
    } as ReturnType<typeof useGetHospitalizations>);

    const user = userEvent.setup();
    render(<HospitalizationList />, { wrapper: createWrapper() });

    // リストビューに切り替え (ToggleGroup type="single" は role="radio")
    await user.click(screen.getByRole("radio", { name: "List View" }));

    // 検索入力を開いて入力
    await user.click(screen.getByRole("button", { name: "検索" }));
    await user.type(screen.getByPlaceholderText("飼主名、ペット名、入院No..."), "ぽち");

    expect(await screen.findByText("ポチ")).toBeInTheDocument();
    expect(screen.queryByText("たろう")).not.toBeInTheDocument();
  });

  it("カタカナ入力でひらがな ownerName がヒットする", async () => {
    vi.mocked(useGetHospitalizations).mockReturnValue({
      data: [
        makeHosp({ id: "1", petName: "ポチ", ownerName: "ヤマダ", hospitalizationNo: "H-001" }),
        makeHosp({ id: "2", petName: "たろう", ownerName: "さとう", hospitalizationNo: "H-002" }),
      ],
      isLoading: false,
      isError: false,
    } as ReturnType<typeof useGetHospitalizations>);

    const user = userEvent.setup();
    render(<HospitalizationList />, { wrapper: createWrapper() });

    await user.click(screen.getByRole("radio", { name: "List View" }));
    await user.click(screen.getByRole("button", { name: "検索" }));
    await user.type(screen.getByPlaceholderText("飼主名、ペット名、入院No..."), "サトウ");

    expect(await screen.findByText("たろう")).toBeInTheDocument();
    expect(screen.queryByText("ポチ")).not.toBeInTheDocument();
  });
});
