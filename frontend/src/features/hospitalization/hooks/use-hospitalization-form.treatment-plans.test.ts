import { describe, it, expect, vi, beforeEach } from "vitest";
import { AxiosError, AxiosHeaders, type InternalAxiosRequestConfig } from "axios";
import { renderHook, act, waitFor } from "@testing-library/react";
import { startTransition } from "react";
import { calculateBillingTotals } from "@/lib/calculations";
import { useGetHospitalizationRaw } from "../api/get-hospitalization";
import { useGetTreatmentPlans } from "../api/get-treatment-plans";
import { useHospitalizationForm } from "./use-hospitalization-form";

async function submitForm(action: ReturnType<typeof useHospitalizationForm>["formAction"]) {
  await act(async () => {
    startTransition(() => action(new FormData()));
  });
}

function renderHospitalizationForm(id?: string, canSubmit = true) {
  return renderHook(() => useHospitalizationForm(id, canSubmit));
}

// ──────────────────────────────────────────────────────────
// モック定義
// vi.mock はホイストされるため、参照する変数は vi.hoisted で先に定義する
// ──────────────────────────────────────────────────────────

const {
  mockNavigate,
  mockToast,
  mockSelectedPets,
  mockSelectedPetsSnapshot,
  mockSetSelectedPets,
  mockSearchParams,
  mockPetFromQuery,
  mockCreateHospitalization,
  mockUpdateHospitalization,
  mockCreateTreatmentPlan,
} = vi.hoisted(() => ({
  mockNavigate: vi.fn(),
  mockToast: { error: vi.fn(), success: vi.fn() },
  mockSelectedPets: [] as unknown[],
  mockSelectedPetsSnapshot: { current: undefined as unknown[] | undefined },
  mockSetSelectedPets: vi.fn(),
  mockSearchParams: new URLSearchParams(),
  mockPetFromQuery: { current: undefined as unknown },
  mockCreateHospitalization: vi.fn().mockResolvedValue({ id: "99" }),
  mockUpdateHospitalization: vi.fn().mockResolvedValue({}),
  mockCreateTreatmentPlan: vi.fn().mockResolvedValue({ id: "1" }),
}));

vi.mock("react-router", () => ({
  useNavigate: () => mockNavigate,
  useSearchParams: () => [mockSearchParams, vi.fn()],
}));

vi.mock("sonner", () => ({ toast: mockToast }));

vi.mock("@/lib/handle-api-error", () => ({ handleApiError: vi.fn() }));

vi.mock("@/hooks/use-pet-selection", () => ({
  usePetSelection: vi.fn(() => ({
    selectedPets: mockSelectedPetsSnapshot.current ?? mockSelectedPets,
    setSelectedPets: mockSetSelectedPets,
    togglePetSelection: vi.fn(),
    isPetSelected: vi.fn(() => false),
  })),
}));

vi.mock("@/hooks/use-pet", () => ({
  useGetPet: vi.fn(() => ({ data: mockPetFromQuery.current, isLoading: false })),
}));

vi.mock("../api/create-hospitalization", () => ({
  createHospitalization: mockCreateHospitalization,
}));

vi.mock("../api/update-hospitalization", () => ({
  updateHospitalization: mockUpdateHospitalization,
}));

vi.mock("../api/treatment-plans-write", () => ({
  createTreatmentPlanForHospitalization: mockCreateTreatmentPlan,
}));

vi.mock("../api/get-hospitalization", () => ({
  useGetHospitalizationRaw: vi.fn(() => ({
    data: undefined,
    isLoading: false,
    isError: false,
    error: null,
    refetch: vi.fn(),
  })),
}));

vi.mock("../api/get-treatment-plans", () => ({
  useGetTreatmentPlans: vi.fn(() => ({
    data: undefined,
    isSuccess: false,
    isLoading: false,
    isError: false,
    error: null,
  })),
}));

vi.mock("@/lib/calculations", () => ({
  calculateBillingTotals: vi.fn(() => ({
    subtotal: 0,
    globalDiscountAmount: 0,
    taxableAmount: 0,
    tax: 0,
    total: 0,
  })),
}));

vi.mock("@/config/paths", () => ({
  paths: {
    hospitalization: {
      selectPet: { getHref: () => "/hospitalization/select-pet" },
      list: { getHref: () => "/hospitalization" },
    },
  },
}));

// ──────────────────────────────────────────────────────────
// テスト
// ──────────────────────────────────────────────────────────
// FE-RC-045: use-hospitalization-form.test.ts (982行) から分割した2ファイル目。
// 親 describe の beforeEach を複製し、独立 describe として実行する（振る舞いは維持）。

describe("useHospitalizationForm — treatmentPlans", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    // デフォルト: ペット未選択
    mockSelectedPets.length = 0;
    mockSelectedPetsSnapshot.current = undefined;
    mockSetSelectedPets.mockImplementation((pets: unknown[]) => {
      mockSelectedPets.splice(0, mockSelectedPets.length, ...pets);
    });
    mockSearchParams.delete("petId");
    mockPetFromQuery.current = undefined;
    mockCreateHospitalization.mockResolvedValue({ id: "99" });
    mockUpdateHospitalization.mockResolvedValue({});
    mockCreateTreatmentPlan.mockResolvedValue({ id: "1" });
    vi.mocked(useGetHospitalizationRaw).mockReturnValue({
      data: undefined,
      isLoading: false,
      isError: false,
      error: null,
    } as ReturnType<typeof useGetHospitalizationRaw>);
    vi.mocked(useGetTreatmentPlans).mockReturnValue({
      data: undefined,
      isSuccess: false,
      isLoading: false,
      isError: false,
      error: null,
    } as ReturnType<typeof useGetTreatmentPlans>);
  });

  it("保険対象flagをbilling計算契約へ明示的に変換する", () => {
    const { result } = renderHospitalizationForm();

    act(() => {
      result.current.addTreatmentPlan();
    });
    const planId = result.current.treatmentPlans[0]!.id;
    act(() => {
      result.current.updateTreatmentPlan(planId, "is_insurance", true);
      result.current.updateTreatmentPlan(planId, "treatmentContent", "plan");
    });

    result.current.calculateTotals();

    expect(vi.mocked(calculateBillingTotals)).toHaveBeenCalledWith(
      expect.arrayContaining([
        expect.objectContaining({ is_insurance: true, isInsuranceApplicable: true }),
      ]),
      0,
      0,
    );
  });

  it("addTreatmentPlan で計画を追加できる", () => {
    const { result } = renderHospitalizationForm();
    const initialCount = result.current.treatmentPlans.length;

    act(() => {
      result.current.addTreatmentPlan();
    });

    expect(result.current.treatmentPlans.length).toBe(initialCount + 1);
  });

  it("removeTreatmentPlan で計画を削除できる", () => {
    const { result } = renderHospitalizationForm();
    act(() => {
      result.current.addTreatmentPlan();
    });
    const firstPlanId = result.current.treatmentPlans[0]?.id;

    if (firstPlanId) {
      act(() => {
        result.current.removeTreatmentPlan(firstPlanId);
      });

      expect(result.current.treatmentPlans.find((p) => p.id === firstPlanId)).toBeUndefined();
    }
  });

  it("編集時は GET treatment-plans wire から hydrate し detail の treatment_plans を見ない", async () => {
    vi.mocked(useGetHospitalizationRaw).mockReturnValue({
      data: {
        id: 7,
        clinic_id: 1,
        owner_id: 2,
        pet_id: 3,
        hospitalization_type: "hospitalization",
        start_date: "2026-07-23T00:00:00+09:00",
        end_date: "2026-07-30T00:00:00+09:00",
        status: "admitted",
        memo: "m",
        owner_request: "o",
        staff_notes: "s",
        created_at: "2026-07-23T00:00:00+09:00",
        updated_at: "2026-07-23T00:00:00+09:00",
        // intentionally no treatment_plans — absent on HospitalizationResponse wire
      },
      isLoading: false,
      isError: false,
      error: null,
    } as ReturnType<typeof useGetHospitalizationRaw>);

    vi.mocked(useGetTreatmentPlans).mockReturnValue({
      data: [
        {
          id: "990018",
          hospitalization_id: "7",
          treatment_content: "合成監査輸液",
          memo: "wire由来",
          is_insurance: true,
          unit_price: 3_210,
          quantity: 2,
          discount_rate: 10,
          discount_amount: 642,
          subtotal: 5_778,
          sort_order: 1,
          created_at: "2026-07-23T00:00:00+09:00",
          updated_at: "2026-07-23T00:00:00+09:00",
        },
      ],
      isSuccess: true,
      isLoading: false,
      isError: false,
      error: null,
    } as ReturnType<typeof useGetTreatmentPlans>);

    const { result } = renderHospitalizationForm("7");

    await waitFor(() => {
      expect(result.current.treatmentPlans).toEqual([
        {
          id: "990018",
          treatmentContent: "合成監査輸液",
          memo: "wire由来",
          is_insurance: true,
          unitPrice: 3_210,
          quantity: 2,
          discount: 10,
          discountAmount: 642,
          subtotal: 5_778,
        },
      ]);
    });

    expect(result.current.formData.memo).toBe("m");
    expect(result.current.formData.ownerRequest).toBe("o");
    expect(useGetTreatmentPlans).toHaveBeenCalledWith("7");
  });
});

describe("useHospitalizationForm BUG-016 entity read", () => {
  function axiosError(status: number | undefined) {
    const config = { headers: new AxiosHeaders() } as InternalAxiosRequestConfig;
    if (status === undefined) {
      return new AxiosError("Network Error", AxiosError.ERR_NETWORK, config, undefined, undefined);
    }
    return new AxiosError("request failed", AxiosError.ERR_BAD_RESPONSE, config, undefined, {
      config,
      data: { error: "not found" },
      headers: new AxiosHeaders(),
      status,
      statusText: "Error",
    });
  }

  beforeEach(() => {
    mockUpdateHospitalization.mockClear();
    mockCreateHospitalization.mockClear();
  });

  it("404 → isReadNotFound、formAction で update/create 0 回", async () => {
    vi.mocked(useGetHospitalizationRaw).mockReturnValue({
      data: undefined,
      isLoading: false,
      isError: true,
      error: axiosError(404),
      refetch: vi.fn(),
    } as unknown as ReturnType<typeof useGetHospitalizationRaw>);

    const { result } = renderHospitalizationForm("999999999");
    expect(result.current.isReadNotFound).toBe(true);
    expect(result.current.entityRead.status).toBe("notFound");

    await submitForm(result.current.formAction);
    expect(mockUpdateHospitalization).not.toHaveBeenCalled();
    expect(mockCreateHospitalization).not.toHaveBeenCalled();
  });

  it("403 → forbiddenOrHidden を isReadNotFound として非開示", async () => {
    vi.mocked(useGetHospitalizationRaw).mockReturnValue({
      data: undefined,
      isLoading: false,
      isError: true,
      error: axiosError(403),
      refetch: vi.fn(),
    } as unknown as ReturnType<typeof useGetHospitalizationRaw>);

    const { result } = renderHospitalizationForm("42");
    expect(result.current.isReadNotFound).toBe(true);
    expect(result.current.entityRead.status).toBe("forbiddenOrHidden");
    await submitForm(result.current.formAction);
    expect(mockUpdateHospitalization).not.toHaveBeenCalled();
  });

  it("network error → isReadError と retry、mutation 0 回", async () => {
    const refetch = vi.fn();
    vi.mocked(useGetHospitalizationRaw).mockReturnValue({
      data: undefined,
      isLoading: false,
      isError: true,
      error: axiosError(undefined),
      refetch,
    } as unknown as ReturnType<typeof useGetHospitalizationRaw>);

    const { result } = renderHospitalizationForm("999999999");
    expect(result.current.isReadError).toBe(true);
    expect(result.current.isReadNotFound).toBe(false);
    result.current.retryRead?.();
    expect(refetch).toHaveBeenCalledTimes(1);
    await submitForm(result.current.formAction);
    expect(mockUpdateHospitalization).not.toHaveBeenCalled();
  });

  it("create route: idle", () => {
    vi.mocked(useGetHospitalizationRaw).mockReturnValue({
      data: undefined,
      isLoading: false,
      isError: false,
      error: null,
      refetch: vi.fn(),
    } as unknown as ReturnType<typeof useGetHospitalizationRaw>);
    const { result } = renderHospitalizationForm();
    expect(result.current.entityRead.status).toBe("idle");
    expect(result.current.isEdit).toBe(false);
  });
});
