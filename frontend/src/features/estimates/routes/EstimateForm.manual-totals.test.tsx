import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { MemoryRouter, Route, Routes } from "react-router";
import type { Estimate } from "../types";
import { EstimateForm } from "./EstimateForm";

// PR669: 実 useEstimateForm + 実 AmountSection + 実 NumberInput を通した
// 明細なし見積の手入力合計ルート回帰。モックは API/navigation/auth のみ。

const {
  mockNavigate,
  mockToast,
  mockEstimateData,
  mockCreateMutateAsync,
  mockUpdateMutateAsync,
  mockPermissions,
} = vi.hoisted(() => ({
  mockNavigate: vi.fn(),
  mockToast: { info: vi.fn(), success: vi.fn(), error: vi.fn() },
  mockEstimateData: { current: undefined as Estimate | undefined },
  mockCreateMutateAsync: vi.fn().mockResolvedValue({}),
  mockUpdateMutateAsync: vi.fn().mockResolvedValue({}),
  mockPermissions: { current: { canCreate: true, canEdit: true } },
}));

vi.mock("@/hooks/use-permission", () => ({
  usePermission: () => ({
    canView: true,
    canCreate: mockPermissions.current.canCreate,
    canEdit: mockPermissions.current.canEdit,
    canDelete: true,
  }),
}));

vi.mock("@/hooks/use-clinic-tax-rates", () => ({
  useClinicTaxRates: () => ({ standardTaxRate: 0.1, reducedTaxRate: 0.08 }),
}));

vi.mock("react-router", async () => {
  const actual = await vi.importActual<typeof import("react-router")>("react-router");
  return { ...actual, useNavigate: () => mockNavigate };
});

vi.mock("sonner", () => ({ toast: mockToast }));

vi.mock("@/hooks/use-pet", () => ({
  useGetPet: () => ({ data: null, isPending: false, isSuccess: false, isError: false }),
}));

vi.mock("../api/get-estimate", () => ({
  useGetEstimate: () => ({
    data: mockEstimateData.current,
    isLoading: false,
    isError: false,
    error: null,
    refetch: vi.fn(),
  }),
}));

vi.mock("../api/create-estimate", () => ({
  useCreateEstimate: vi.fn(() => ({ mutateAsync: mockCreateMutateAsync })),
}));

vi.mock("../api/update-estimate", () => ({
  useUpdateEstimate: vi.fn(() => ({ mutateAsync: mockUpdateMutateAsync })),
}));

vi.mock("@/hooks/use-unsaved-changes", () => ({
  useUnsavedChanges: () => ({
    isDirty: false,
    markDirty: vi.fn(),
    markClean: vi.fn(),
  }),
}));

// 明細エディタは本テストの対象外（金額セクションの実装は実物を使う）
vi.mock("../components/EstimateItemsEditor", () => ({
  EstimateItemsEditor: () => <div data-testid="items-editor" />,
}));

function makeEstimate(overrides: Partial<Estimate> = {}): Estimate {
  return {
    id: "1",
    clinicId: "1",
    estimateNo: "EST-001",
    medicalRecordId: null,
    ownerId: null,
    petId: null,
    createdBy: null,
    title: "既存見積",
    status: "draft",
    subtotal: 9000,
    taxTotal: 900,
    totalAmount: 9900,
    insuranceAmount: 0,
    discountAmount: 0,
    validUntil: null,
    comment: "",
    notes: "",
    items: [],
    createdAt: "2026-01-01T00:00:00Z",
    updatedAt: "2026-01-01T00:00:00Z",
    ...overrides,
  };
}

const itemizedEstimate: Estimate = makeEstimate({
  items: [
    {
      id: "10",
      estimateId: "1",
      name: "診察",
      category: "medicine",
      unitPrice: 1000,
      quantity: 1,
      taxType: "excluded",
      taxRate: 0.1,
      discountRate: 0,
      discountAmount: 0,
      isInsuranceApplicable: false,
      sortOrder: 0,
      createdAt: "2026-01-01T00:00:00Z",
      updatedAt: "2026-01-01T00:00:00Z",
    },
  ],
});

function renderForm(path: string) {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={queryClient}>
      <MemoryRouter initialEntries={[path]}>
        <Routes>
          <Route path="/estimates/new" element={<EstimateForm />} />
          <Route path="/estimates/:id/edit" element={<EstimateForm />} />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>,
  );
}

describe("EstimateForm — PR669 明細なし手入力合計（実 hook/セクション/NumberInput）", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockEstimateData.current = undefined;
    mockPermissions.current = { canCreate: true, canEdit: true };
  });

  it("create: 明細なしで 小計/消費税/合計金額 のラベル付き入力へ値を入れて送信できる", async () => {
    const user = userEvent.setup();
    renderForm("/estimates/new");

    const subtotal = screen.getByLabelText("小計");
    const tax = screen.getByLabelText("消費税");
    const total = screen.getByLabelText("合計金額");
    expect(subtotal).not.toBeDisabled();
    expect(tax).not.toBeDisabled();
    expect(total).not.toBeDisabled();

    await user.clear(subtotal);
    await user.type(subtotal, "9000");
    await user.clear(tax);
    await user.type(tax, "900");
    await user.clear(total);
    await user.type(total, "9900");
    await user.type(screen.getByLabelText(/タイトル/), "手入力見積");
    await user.click(screen.getByRole("button", { name: "保存" }));

    await waitFor(() => expect(mockCreateMutateAsync).toHaveBeenCalledTimes(1));
    expect(mockCreateMutateAsync).toHaveBeenCalledWith(
      expect.objectContaining({
        subtotal: 9000,
        tax_total: 900,
        total_amount: 9900,
      }),
    );
  });

  it("edit: もともと明細なしの見積は保存済み合計をハイドレートし編集して保存できる（items 置換なし）", async () => {
    const user = userEvent.setup();
    mockEstimateData.current = makeEstimate();
    renderForm("/estimates/1/edit");

    const total = await screen.findByLabelText("合計金額");
    expect(total).toHaveValue(9900);
    await user.clear(total);
    await user.type(total, "12000");
    await user.click(screen.getByRole("button", { name: "保存" }));

    await waitFor(() => expect(mockUpdateMutateAsync).toHaveBeenCalledTimes(1));
    const data = mockUpdateMutateAsync.mock.calls[0]![0].data;
    expect(data.total_amount).toBe(12000);
    expect(data).not.toHaveProperty("items");
  });

  it("明細あり見積では手入力欄を出さず導出合計は読み取り専用", async () => {
    mockEstimateData.current = itemizedEstimate;
    renderForm("/estimates/1/edit");

    await screen.findByText("合計金額");
    // 読み取り専用サマリ（input ではない）
    expect(screen.queryByLabelText("小計")).not.toBeInTheDocument();
    expect(screen.queryByLabelText("消費税")).not.toBeInTheDocument();
    expect(screen.queryByLabelText("合計金額")).not.toBeInTheDocument();
    expect(screen.getByText("小計（明細合計）")).toBeInTheDocument();
  });
});
