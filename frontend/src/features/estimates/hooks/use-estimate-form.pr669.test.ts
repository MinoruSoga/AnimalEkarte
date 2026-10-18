import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, act } from "@testing-library/react";
import { startTransition } from "react";
import {
  useEstimateForm,
  type EstimateItemDraft,
  type EstimateMutationPermissions,
} from "./use-estimate-form";
import type { Estimate } from "../types";

// PR669: 明細なし見積の手入力合計、保険+割引の実効控除、
// legacy 保存済み合計の保持、数量 0.1 刻みの FP 誤判定の回帰テスト。

const ALLOW_ALL_PERMISSIONS: Readonly<EstimateMutationPermissions> = {
  canCreate: true,
  canEdit: true,
};

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

const baseEstimate: Estimate = {
  id: "1",
  clinicId: "1",
  estimateNo: "EST-001",
  medicalRecordId: null,
  ownerId: null,
  petId: null,
  createdBy: null,
  title: "既存見積",
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

const itemizedEstimate: Estimate = {
  ...baseEstimate,
  items: [
    {
      id: "10",
      estimateId: "1",
      name: "外税処置",
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
};

function makeItem(overrides: Partial<EstimateItemDraft> = {}): EstimateItemDraft {
  return {
    id: "",
    localId: "local-1",
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
    ...overrides,
  };
}

async function submit(formAction: (formData: FormData) => void) {
  await act(async () => {
    await startTransition(async () => {
      formAction(new FormData());
    });
  });
}

describe("useEstimateForm — PR669 manual totals (明細なし見積)", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("create: 明細なしで手入力した小計/税/合計をそのまま送る", async () => {
    const { result } = renderHook(() =>
      useEstimateForm({ mode: "create", permissions: ALLOW_ALL_PERMISSIONS }),
    );
    act(() => {
      result.current.handleChange("title", "手入力見積");
      result.current.handleChange("subtotal", 9000);
      result.current.handleChange("taxTotal", 900);
      result.current.handleChange("totalAmount", 9900);
    });
    expect(result.current.manualTotalsAllowed).toBe(true);
    expect(result.current.totals).toEqual({
      subtotal: 9000,
      taxTotal: 900,
      totalAmount: 9900,
    });
    await submit(result.current.formAction);
    expect(mockCreateMutateAsync).toHaveBeenCalledWith(
      expect.objectContaining({ subtotal: 9000, tax_total: 900, total_amount: 9900 }),
    );
  });

  it("create: 手入力合計が負または安全整数外なら送信しない", async () => {
    const { result } = renderHook(() =>
      useEstimateForm({ mode: "create", permissions: ALLOW_ALL_PERMISSIONS }),
    );
    act(() => {
      result.current.handleChange("title", "不正合計");
      result.current.handleChange("totalAmount", -1);
    });
    await submit(result.current.formAction);
    expect(mockCreateMutateAsync).not.toHaveBeenCalled();
    expect(result.current.formState.fieldErrors?.items).toBeTruthy();
  });

  it("edit: もともと明細なしの見積は手入力合計の変更のみ送る（items置換なし）", async () => {
    const { result } = renderHook(() =>
      useEstimateForm({
        mode: "edit",
        estimate: baseEstimate,
        permissions: ALLOW_ALL_PERMISSIONS,
      }),
    );
    expect(result.current.manualTotalsAllowed).toBe(true);
    act(() => {
      result.current.handleChange("totalAmount", 1200);
    });
    await submit(result.current.formAction);
    const data = mockUpdateMutateAsync.mock.calls[0]![0].data;
    expect(data.total_amount).toBe(1200);
    expect(data).not.toHaveProperty("items");
  });

  it("edit: もともと明細なしで合計未変更なら total 系を送らない", async () => {
    const { result } = renderHook(() =>
      useEstimateForm({
        mode: "edit",
        estimate: baseEstimate,
        permissions: ALLOW_ALL_PERMISSIONS,
      }),
    );
    act(() => {
      result.current.handleChange("title", "タイトルのみ");
    });
    await submit(result.current.formAction);
    const data = mockUpdateMutateAsync.mock.calls[0]![0].data;
    expect(data).not.toHaveProperty("subtotal");
    expect(data).not.toHaveProperty("tax_total");
    expect(data).not.toHaveProperty("total_amount");
    expect(data).not.toHaveProperty("items");
  });

  it("edit: もともと明細ありで全行クリアは手入力不可・導出0のまま", () => {
    const { result } = renderHook(() =>
      useEstimateForm({
        mode: "edit",
        estimate: itemizedEstimate,
        permissions: ALLOW_ALL_PERMISSIONS,
      }),
    );
    act(() => {
      result.current.handleChange("items", []);
    });
    expect(result.current.manualTotalsAllowed).toBe(false);
    expect(result.current.totals).toEqual({ subtotal: 0, taxTotal: 0, totalAmount: 0 });
  });
});

describe("useEstimateForm — PR669 itemized adjustments (保険+割引)", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("create: 明細合計から保険と割引の両方を控除した合計を送る", async () => {
    const { result } = renderHook(() =>
      useEstimateForm({ mode: "create", permissions: ALLOW_ALL_PERMISSIONS }),
    );
    act(() => {
      result.current.handleChange("title", "調整付き");
      result.current.handleChange("items", [makeItem()]); // gross 1100
      result.current.handleChange("discountAmount", 100);
      result.current.handleChange("insuranceAmount", 200);
    });
    expect(result.current.totals.totalAmount).toBe(800);
    await submit(result.current.formAction);
    expect(mockCreateMutateAsync).toHaveBeenCalledWith(
      expect.objectContaining({
        subtotal: 1000,
        tax_total: 100,
        total_amount: 800,
        discount_amount: 100,
        insurance_amount: 200,
      }),
    );
  });

  it("create: 保険が gross-discount を超えると送信しない", async () => {
    const { result } = renderHook(() =>
      useEstimateForm({ mode: "create", permissions: ALLOW_ALL_PERMISSIONS }),
    );
    act(() => {
      result.current.handleChange("title", "超過保険");
      result.current.handleChange("items", [makeItem()]);
      result.current.handleChange("discountAmount", 100);
      result.current.handleChange("insuranceAmount", 1001);
    });
    await submit(result.current.formAction);
    expect(mockCreateMutateAsync).not.toHaveBeenCalled();
    expect(result.current.formState.fieldErrors?.items).toBeTruthy();
  });

  it("edit: 明細・控除とも据え置きなら保存済み合計を表示し total 系を送らない", async () => {
    const legacy: Estimate = {
      ...itemizedEstimate,
      // 旧データ: 保存済み超過割引（gross1100 超）
      discountAmount: 2000,
      totalAmount: 1100,
    };
    const { result } = renderHook(() =>
      useEstimateForm({
        mode: "edit",
        estimate: legacy,
        permissions: ALLOW_ALL_PERMISSIONS,
      }),
    );
    // 据え置きでは保存済み合計を表示（再計算しない）
    expect(result.current.totals.totalAmount).toBe(1100);
    act(() => {
      result.current.handleChange("title", "タイトルのみ変更");
    });
    await submit(result.current.formAction);
    const data = mockUpdateMutateAsync.mock.calls[0]![0].data;
    expect(data).not.toHaveProperty("items");
    expect(data).not.toHaveProperty("subtotal");
    expect(data).not.toHaveProperty("total_amount");
  });

  it("edit: 控除を実変更した場合は導出合計を再計算して送る", async () => {
    const { result } = renderHook(() =>
      useEstimateForm({
        mode: "edit",
        estimate: itemizedEstimate,
        permissions: ALLOW_ALL_PERMISSIONS,
      }),
    );
    act(() => {
      result.current.handleChange("insuranceAmount", 300);
    });
    await submit(result.current.formAction);
    const data = mockUpdateMutateAsync.mock.calls[0]![0].data;
    expect(data).not.toHaveProperty("items");
    // gross1100 - discount0 - insurance300 = 800
    expect(data.total_amount).toBe(800);
    expect(data.insurance_amount).toBe(300);
  });

  it("edit: 超過する控除変更はクライアント側で拒否する", async () => {
    const { result } = renderHook(() =>
      useEstimateForm({
        mode: "edit",
        estimate: itemizedEstimate,
        permissions: ALLOW_ALL_PERMISSIONS,
      }),
    );
    act(() => {
      result.current.handleChange("insuranceAmount", 1101);
    });
    await submit(result.current.formAction);
    expect(mockUpdateMutateAsync).not.toHaveBeenCalled();
    expect(result.current.formState.fieldErrors?.items).toBeTruthy();
  });
});

describe("useEstimateForm — PR669 manual totals unsafe inputs", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it.each([
    ["subtotal", -1],
    ["taxTotal", -1],
    ["totalAmount", -1],
    ["subtotal", 0.5],
    ["taxTotal", 0.5],
    ["totalAmount", 0.5],
    ["subtotal", Number.MAX_SAFE_INTEGER + 1],
    ["totalAmount", Number.MAX_SAFE_INTEGER + 1],
  ] as const)("rejects unsafe manual %s=%s before API", async (field, value) => {
    const { result } = renderHook(() =>
      useEstimateForm({ mode: "create", permissions: ALLOW_ALL_PERMISSIONS }),
    );
    act(() => {
      result.current.handleChange("title", "不正合計");
      result.current.handleChange(field, value);
    });
    await submit(result.current.formAction);
    expect(mockCreateMutateAsync).not.toHaveBeenCalled();
    expect(result.current.formState.fieldErrors?.items).toBeTruthy();
  });
});

describe("useEstimateForm — PR669 included-tax adjustments", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("create: 内税明細 gross1100 tax100 に割引100+保険200 → total800", async () => {
    const { result } = renderHook(() =>
      useEstimateForm({ mode: "create", permissions: ALLOW_ALL_PERMISSIONS }),
    );
    act(() => {
      result.current.handleChange("title", "内税調整");
      result.current.handleChange("items", [makeItem({ unitPrice: 1100, taxType: "included" })]);
      result.current.handleChange("discountAmount", 100);
      result.current.handleChange("insuranceAmount", 200);
    });
    expect(result.current.totals).toEqual({
      subtotal: 1100,
      taxTotal: 100,
      totalAmount: 800,
    });
    await submit(result.current.formAction);
    expect(mockCreateMutateAsync).toHaveBeenCalledWith(
      expect.objectContaining({
        subtotal: 1100,
        tax_total: 100,
        total_amount: 800,
      }),
    );
  });

  it("create: 保険 = gross-discount なら合計0で送信できる", async () => {
    const { result } = renderHook(() =>
      useEstimateForm({ mode: "create", permissions: ALLOW_ALL_PERMISSIONS }),
    );
    act(() => {
      result.current.handleChange("title", "全額保険");
      result.current.handleChange("items", [makeItem()]);
      result.current.handleChange("discountAmount", 100);
      result.current.handleChange("insuranceAmount", 1000);
    });
    expect(result.current.totals.totalAmount).toBe(0);
    await submit(result.current.formAction);
    expect(mockCreateMutateAsync).toHaveBeenCalledWith(
      expect.objectContaining({ total_amount: 0 }),
    );
  });
});

describe("useEstimateForm — PR669 quantity 0.1 step", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  const validQuantities = [0.1, 0.3, 1000000.1, 999999999.9, 999999999.8];
  for (const q of validQuantities) {
    it(`accepts canonical 0.1-step quantity ${q}`, async () => {
      const { result } = renderHook(() =>
        useEstimateForm({ mode: "create", permissions: ALLOW_ALL_PERMISSIONS }),
      );
      act(() => {
        result.current.handleChange("title", `q=${q}`);
        result.current.handleChange("items", [makeItem({ quantity: q, unitPrice: 1 })]);
      });
      await submit(result.current.formAction);
      expect(result.current.formState.fieldErrors?.items).toBeFalsy();
      expect(mockCreateMutateAsync).toHaveBeenCalled();
    });
  }

  const invalidQuantities = [0.15, 1000000.15, 1.01, 0.05];
  for (const q of invalidQuantities) {
    it(`rejects non-0.1-step quantity ${q}`, async () => {
      const { result } = renderHook(() =>
        useEstimateForm({ mode: "create", permissions: ALLOW_ALL_PERMISSIONS }),
      );
      act(() => {
        result.current.handleChange("title", `q=${q}`);
        result.current.handleChange("items", [makeItem({ quantity: q, unitPrice: 1 })]);
      });
      await submit(result.current.formAction);
      expect(result.current.formState.fieldErrors?.items).toBeTruthy();
      expect(mockCreateMutateAsync).not.toHaveBeenCalled();
    });
  }
});
