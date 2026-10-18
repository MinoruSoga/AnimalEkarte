import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, act } from "@testing-library/react";
import { startTransition } from "react";
import {
  useEstimateForm,
  type EstimateItemDraft,
  type EstimateMutationPermissions,
} from "./use-estimate-form";
import type { Estimate } from "../types";

// FE-RC-001: このファイルの既存テスト群は「権限あり」の通常フローを検証するため、
// 既定で create/edit を全許可する。権限拒否のfail-closed挙動は専用のdescribeで検証する。
const ALLOW_ALL_PERMISSIONS: Readonly<EstimateMutationPermissions> = {
  canCreate: true,
  canEdit: true,
};

// ──────────────────────────────────────────────────────────
// モック定義
// vi.mock はホイストされるため、参照する変数は vi.hoisted で先に定義する
// ──────────────────────────────────────────────────────────

const { mockNavigate, mockToast, mockCreateMutateAsync, mockUpdateMutateAsync } = vi.hoisted(
  () => ({
    mockNavigate: vi.fn(),
    mockToast: { error: vi.fn(), success: vi.fn(), info: vi.fn() },
    mockCreateMutateAsync: vi.fn().mockResolvedValue({}),
    mockUpdateMutateAsync: vi.fn().mockResolvedValue({}),
  }),
);

vi.mock("react-router", () => ({
  useNavigate: () => mockNavigate,
}));

vi.mock("sonner", () => ({ toast: mockToast }));

vi.mock("../api/create-estimate", () => ({
  useCreateEstimate: vi.fn(() => ({ mutateAsync: mockCreateMutateAsync })),
}));

vi.mock("../api/update-estimate", () => ({
  useUpdateEstimate: vi.fn(() => ({ mutateAsync: mockUpdateMutateAsync })),
}));

vi.mock("@/config/paths", () => ({
  paths: {
    estimates: {
      getHref: () => "/estimates",
      detail: { getHref: (id: string) => `/estimates/${id}` },
    },
  },
}));

// ──────────────────────────────────────────────────────────
// テスト用フィクスチャ
// ──────────────────────────────────────────────────────────

const mockEstimate: Estimate = {
  id: "1",
  clinicId: "1",
  estimateNo: "EST-001",
  medicalRecordId: null,
  ownerId: null,
  ownerName: undefined,
  petId: null,
  petName: undefined,
  createdBy: null,
  title: "テスト見積書",
  status: "draft",
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

// ──────────────────────────────────────────────────────────
// テスト
// ──────────────────────────────────────────────────────────

function makeItem(overrides: Partial<EstimateItemDraft> = {}): EstimateItemDraft {
  return {
    id: "",
    localId: "local-1",
    name: "検査",
    category: "other",
    unitPrice: 1100,
    quantity: 1,
    taxType: "included",
    taxRate: 0.1,
    discountRate: 0,
    discountAmount: 0,
    isInsuranceApplicable: false,
    sortOrder: 0,
    ...overrides,
  };
}

function estimateWithHeaderDiscount(): Estimate {
  return {
    ...mockEstimate,
    discountAmount: 200,
    subtotal: 1000,
    taxTotal: 91,
    totalAmount: 800,
    items: [
      {
        id: "11",
        estimateId: "1",
        name: "Test item",
        category: "other",
        unitPrice: 1100,
        quantity: 1,
        taxType: "included",
        taxRate: 0.1,
        discountRate: 0,
        discountAmount: 100,
        isInsuranceApplicable: false,
        sortOrder: 0,
        createdAt: "2026-01-01T00:00:00Z",
      },
    ],
  };
}

async function submit(formAction: (data: FormData) => void) {
  await act(async () => {
    startTransition(() => formAction(new FormData()));
  });
}

describe("estimate header discount projection", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockCreateMutateAsync.mockResolvedValue({});
    mockUpdateMutateAsync.mockResolvedValue({});
  });

  it("uses the same discounted excluded-tax total for preview and creation", async () => {
    const { result } = renderHook(() =>
      useEstimateForm({ mode: "create", permissions: ALLOW_ALL_PERMISSIONS }),
    );
    act(() => {
      result.current.handleChange("title", "Header discount");
      result.current.handleChange("items", [makeItem({ unitPrice: 1000, taxType: "excluded" })]);
      result.current.handleChange("discountAmount", 300);
    });
    expect(result.current.totals).toEqual({ subtotal: 1000, taxTotal: 100, totalAmount: 800 });
    await submit(result.current.formAction);
    expect(mockCreateMutateAsync).toHaveBeenCalledWith(
      expect.objectContaining({
        subtotal: 1000,
        tax_total: 100,
        total_amount: 800,
        discount_amount: 300,
      }),
    );
  });

  it("subtracts line and header discounts once in included-tax replacement", async () => {
    const estimate = estimateWithHeaderDiscount();
    const { result } = renderHook(() =>
      useEstimateForm({ mode: "edit", estimate, permissions: ALLOW_ALL_PERMISSIONS }),
    );
    act(() => {
      result.current.handleChange(
        "items",
        result.current.form.items.map((item) => ({ ...item, name: "Renamed" })),
      );
    });
    expect(result.current.totals).toEqual({ subtotal: 1000, taxTotal: 91, totalAmount: 800 });
    await submit(result.current.formAction);
    expect(mockUpdateMutateAsync).toHaveBeenCalledWith({
      id: "1",
      data: expect.objectContaining({
        subtotal: 1000,
        tax_total: 91,
        total_amount: 800,
        discount_amount: 200,
        items: [expect.objectContaining({ existing_item_id: 11, discount_amount: 100 })],
      }),
    });
  });

  it("refreshes header-only preview and clearing without replacing rows", async () => {
    const estimate = estimateWithHeaderDiscount();
    const { result } = renderHook(() =>
      useEstimateForm({ mode: "edit", estimate, permissions: ALLOW_ALL_PERMISSIONS }),
    );
    act(() => {
      result.current.handleChange("discountAmount", 300);
    });
    expect(result.current.totals.totalAmount).toBe(700);
    await submit(result.current.formAction);
    expect(mockUpdateMutateAsync.mock.calls[0]?.[0].data).toMatchObject({ discount_amount: 300 });
    expect(mockUpdateMutateAsync.mock.calls[0]?.[0].data).not.toHaveProperty("items");
    // PR669: 控除の実変更は導出合計を再計算して送る（保険0: gross1000-300=700）
    expect(mockUpdateMutateAsync.mock.calls[0]?.[0].data).toMatchObject({ total_amount: 700 });
    act(() => {
      result.current.handleChange("discountAmount", 0);
    });
    expect(result.current.totals.totalAmount).toBe(1000);
  });

  it("preserves stored manual totals for unchanged legacy zero-row estimates", async () => {
    const estimate: Estimate = { ...mockEstimate, discountAmount: 300, totalAmount: 800 };
    const { result } = renderHook(() =>
      useEstimateForm({ mode: "edit", estimate, permissions: ALLOW_ALL_PERMISSIONS }),
    );
    expect(result.current.totals.totalAmount).toBe(800);
    act(() => {
      result.current.handleChange("title", "Metadata only");
    });
    await submit(result.current.formAction);
    expect(mockUpdateMutateAsync.mock.calls[0]?.[0].data).not.toHaveProperty("items");
    expect(mockUpdateMutateAsync.mock.calls[0]?.[0].data).not.toHaveProperty("total_amount");
  });

  it.each([-1, 0.5, 1101])(
    "rejects invalid header discount %s before mutation",
    async (discountAmount) => {
      const { result } = renderHook(() =>
        useEstimateForm({ mode: "create", permissions: ALLOW_ALL_PERMISSIONS }),
      );
      act(() => {
        result.current.handleChange("title", "Invalid header discount");
        result.current.handleChange("items", [makeItem()]);
        result.current.handleChange("discountAmount", discountAmount);
      });
      await submit(result.current.formAction);
      expect(mockCreateMutateAsync).not.toHaveBeenCalled();
      expect(result.current.formState.fieldErrors?.items).toBeTruthy();
    },
  );

  it("accepts a discount equal to the derived gross with a zero final total", async () => {
    const { result } = renderHook(() =>
      useEstimateForm({ mode: "create", permissions: ALLOW_ALL_PERMISSIONS }),
    );
    act(() => {
      result.current.handleChange("title", "Zero final total");
      result.current.handleChange("items", [makeItem()]);
      result.current.handleChange("discountAmount", 1100);
    });
    await submit(result.current.formAction);
    expect(result.current.totals.totalAmount).toBe(0);
    expect(mockCreateMutateAsync).toHaveBeenCalledWith(
      expect.objectContaining({ total_amount: 0, discount_amount: 1100 }),
    );
  });
});
