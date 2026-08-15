import { describe, it, expect, vi, beforeEach } from "vitest";
import { AxiosError, AxiosHeaders, type InternalAxiosRequestConfig } from "axios";
import { renderHook, act, waitFor } from "@testing-library/react";
import { startTransition, useLayoutEffect, useRef } from "react";
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

/** BUG-016: edit mode submit requires a found entity; shared fixture for regression tests. */
function mockFoundHospitalization(id = 42) {
  vi.mocked(useGetHospitalizationRaw).mockReturnValue({
    data: {
      id,
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
    },
    isLoading: false,
    isError: false,
    error: null,
    refetch: vi.fn(),
  } as unknown as ReturnType<typeof useGetHospitalizationRaw>);
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

describe("useHospitalizationForm", () => {
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

  // ──────────────────────────
  // 初期状態
  // ──────────────────────────
  describe("初期状態", () => {
    it("id なし → isEdit = false", () => {
      const { result } = renderHospitalizationForm();
      expect(result.current.isEdit).toBe(false);
    });

    it("id あり → isEdit = true", () => {
      const { result } = renderHospitalizationForm("10");
      expect(result.current.isEdit).toBe(true);
    });

    it("formState の初期値は success: false", () => {
      const { result } = renderHospitalizationForm();
      expect(result.current.formState.success).toBe(false);
    });

    it("isSaving の初期値は false", () => {
      const { result } = renderHospitalizationForm();
      expect(result.current.isSaving).toBe(false);
    });
  });

  // ──────────────────────────
  // バリデーション: ペット未選択
  // ──────────────────────────
  describe("バリデーション: selectedPets が空の場合", () => {
    it("selectedPets が空 → success: false かつ fieldErrors.pet が定義される", async () => {
      // selectedPets は空（デフォルト）
      const { result } = renderHospitalizationForm();

      await submitForm(result.current.formAction);

      expect(result.current.formState.success).toBe(false);
      expect(result.current.formState.fieldErrors?.pet).toBe(
        "ペットを選択してください"
      );
    });

    it("selectedPets が空 → createHospitalization は呼ばれない", async () => {
      const { result } = renderHospitalizationForm();

      await submitForm(result.current.formAction);

      expect(mockCreateHospitalization).not.toHaveBeenCalled();
    });
  });

  // ──────────────────────────
  // 新規作成モード
  // ──────────────────────────
  describe("新規作成モード（id なし）", () => {
    beforeEach(() => {
      // ペットを1件選択した状態にする
      mockSelectedPets.push({
        id: "1",
        ownerId: "2",
        ownerName: "田中太郎",
        name: "ポチ",
        species: "犬",
        breed: "柴犬",
        gender: "male",
      });
    });

    afterEach(() => {
      mockSelectedPets.length = 0;
    });

    async function selectCage(result: { current: { handleFormDataChange: (u: { cageId: string }) => void } }) {
      await act(async () => {
        result.current.handleFormDataChange({ cageId: "10" });
      });
    }

    it("ケージ未選択 → create せず fieldErrors.cage_id（BUG-037）", async () => {
      const { result } = renderHospitalizationForm();

      await submitForm(result.current.formAction);

      expect(mockCreateHospitalization).not.toHaveBeenCalled();
      expect(result.current.formState.success).toBe(false);
      expect(result.current.formState.fieldErrors?.cage_id).toBe(
        "ケージ・個室を選択してください",
      );
    });

    it("selectedPets がある & id なし → createHospitalization が呼ばれる", async () => {
      const { result } = renderHospitalizationForm();
      await selectCage(result);

      await submitForm(result.current.formAction);

      expect(mockCreateHospitalization).toHaveBeenCalledTimes(1);
      expect(mockUpdateHospitalization).not.toHaveBeenCalled();
    });

    it("createHospitalization に pet_id と owner_id が渡される", async () => {
      const { result } = renderHospitalizationForm();
      await selectCage(result);

      await submitForm(result.current.formAction);

      expect(mockCreateHospitalization).toHaveBeenCalledWith(
        expect.objectContaining({
          pet_id: "1",
          owner_id: "2",
          cage_id: "10",
        })
      );
    });

    it("成功時 → toast.success が呼ばれる", async () => {
      const { result } = renderHospitalizationForm();
      await selectCage(result);

      await submitForm(result.current.formAction);

      expect(mockToast.success).toHaveBeenCalledWith("入院情報を登録しました");
    });

    it("成功時 → formState.success = true", async () => {
      const { result } = renderHospitalizationForm();
      await selectCage(result);

      await submitForm(result.current.formAction);

      expect(result.current.formState.success).toBe(true);
    });

    it("createHospitalization が失敗した場合 → formState.success = false", async () => {
      mockCreateHospitalization.mockRejectedValueOnce(new Error("API Error"));

      const { result } = renderHospitalizationForm();
      await selectCage(result);

      await submitForm(result.current.formAction);

      expect(result.current.formState.success).toBe(false);
    });

    it("作成権限なし → createHospitalization は呼ばれない", async () => {
      const { result } = renderHospitalizationForm(undefined, false);
      await selectCage(result);

      await submitForm(result.current.formAction);

      expect(mockCreateHospitalization).not.toHaveBeenCalled();
    });

    it("死亡petは作成mutation境界で拒否する", async () => {
      mockSelectedPets[0] = {
        ...mockSelectedPets[0],
        status: "死亡",
      };
      const { result } = renderHospitalizationForm();

      await submitForm(result.current.formAction);

      expect(mockCreateHospitalization).not.toHaveBeenCalled();
      expect(result.current.formState.fieldErrors?.pet).toBe(
        "死亡したペットは入院登録できません",
      );
    });

    it("新規作成の初期 treatmentPlans は空（偽デフォルト行を持たない）", () => {
      const { result } = renderHospitalizationForm();
      expect(result.current.treatmentPlans).toEqual([]);
    });

    it("治療内容ありの行は create の treatment_plans に同梱され nested POST しない", async () => {
      const { result } = renderHospitalizationForm();
      await selectCage(result);

      await act(async () => {
        result.current.addTreatmentPlan();
      });
      const planId = result.current.treatmentPlans[0]?.id;
      expect(planId).toBeTruthy();
      await act(async () => {
        result.current.updateTreatmentPlan(planId!, "treatmentContent", "adm rate");
        result.current.updateTreatmentPlan(planId!, "unitPrice", 990);
      });

      await submitForm(result.current.formAction);

      expect(mockCreateHospitalization).toHaveBeenCalledTimes(1);
      expect(mockCreateHospitalization).toHaveBeenCalledWith(
        expect.objectContaining({
          treatment_plans: [
            expect.objectContaining({
              treatment_content: "adm rate",
              unit_price: 990,
              quantity: 1,
            }),
          ],
        }),
      );
      expect(mockCreateTreatmentPlan).not.toHaveBeenCalled();
    });

    it("空の治療内容行は treatment_plans に含めない", async () => {
      const { result } = renderHospitalizationForm();
      await selectCage(result);
      await act(async () => {
        result.current.addTreatmentPlan();
      });

      await submitForm(result.current.formAction);

      expect(mockCreateHospitalization).toHaveBeenCalledTimes(1);
      const payload = mockCreateHospitalization.mock.calls[0]?.[0] as {
        treatment_plans?: unknown[];
      };
      expect(payload.treatment_plans).toBeUndefined();
      expect(mockCreateTreatmentPlan).not.toHaveBeenCalled();
    });

    it("選択petが死亡へ変わったcommit直後のlayout phaseでも取得済みformActionはcreate mutationを発行しない", async () => {
      mockSelectedPets.length = 0;
      const livingPet = {
        id: "1",
        ownerId: "2",
        ownerName: "田中太郎",
        name: "ポチ",
        species: "犬",
        status: "生存",
      };
      const deceasedPet = { ...livingPet, status: "死亡" };
      const { result, rerender } = renderHook(
        ({ pet }: { pet: typeof livingPet }) => {
          mockSelectedPetsSnapshot.current = [pet];
          const form = useHospitalizationForm(undefined, true);
          const capturedActionRef = useRef(form.formAction);
          useLayoutEffect(() => {
            if (pet.status === "死亡") {
              startTransition(() => capturedActionRef.current(new FormData()));
            }
          }, [pet.status]);
          return form;
        },
        { initialProps: { pet: livingPet } },
      );

      await act(async () => {
        result.current.handleFormDataChange({ cageId: "10" });
      });

      const initialTimestamp = result.current.formState.timestamp;
      await act(async () => {
        rerender({ pet: deceasedPet });
      });

      expect(result.current.formState.timestamp).not.toBe(initialTimestamp);
      expect(mockCreateHospitalization).not.toHaveBeenCalled();
      expect(result.current.formState.fieldErrors?.pet).toBe(
        "死亡したペットは入院登録できません",
      );
    });
  });

  // ──────────────────────────
  // 編集モード
  // ──────────────────────────
  describe("編集モード（id あり）", () => {
    beforeEach(() => {
      mockFoundHospitalization(42);
      mockSelectedPets.push({
        id: "1",
        ownerId: "2",
        ownerName: "田中太郎",
        name: "ポチ",
        species: "犬",
        breed: "柴犬",
        gender: "male",
      });
    });

    afterEach(() => {
      mockSelectedPets.length = 0;
    });

    async function selectCage(result: { current: { handleFormDataChange: (u: { cageId: string }) => void } }) {
      await act(async () => {
        result.current.handleFormDataChange({ cageId: "10" });
      });
    }

    it("ケージ未選択 → update せず fieldErrors.cage_id（BUG-037）", async () => {
      const { result } = renderHospitalizationForm("42");

      await submitForm(result.current.formAction);

      expect(mockUpdateHospitalization).not.toHaveBeenCalled();
      expect(result.current.formState.fieldErrors?.cage_id).toBe(
        "ケージ・個室を選択してください",
      );
    });

    it("selectedPets がある & id あり → updateHospitalization が呼ばれる", async () => {
      const { result } = renderHospitalizationForm("42");
      await selectCage(result);

      await submitForm(result.current.formAction);

      expect(mockUpdateHospitalization).toHaveBeenCalledTimes(1);
      expect(mockCreateHospitalization).not.toHaveBeenCalled();
    });

    it("updateHospitalization に id が渡される", async () => {
      const { result } = renderHospitalizationForm("42");
      await selectCage(result);

      await submitForm(result.current.formAction);

      expect(mockUpdateHospitalization).toHaveBeenCalledWith(
        "42",
        expect.any(Object)
      );
    });

    it("成功時 → toast.success が呼ばれる（更新メッセージ）", async () => {
      const { result } = renderHospitalizationForm("42");
      await selectCage(result);

      await submitForm(result.current.formAction);

      expect(mockToast.success).toHaveBeenCalledWith("入院情報を更新しました");
    });

    it("成功時 → formState.success = true", async () => {
      const { result } = renderHospitalizationForm("42");
      await selectCage(result);

      await submitForm(result.current.formAction);

      expect(result.current.formState.success).toBe(true);
    });

    it("updateHospitalization が失敗した場合 → formState.success = false", async () => {
      mockUpdateHospitalization.mockRejectedValueOnce(new Error("API Error"));

      const { result } = renderHospitalizationForm("42");
      await selectCage(result);

      await submitForm(result.current.formAction);

      expect(result.current.formState.success).toBe(false);
    });

    it("編集権限なし → updateHospitalization は呼ばれない", async () => {
      const { result } = renderHospitalizationForm("42", false);
      await selectCage(result);

      await submitForm(result.current.formAction);

      expect(mockUpdateHospitalization).not.toHaveBeenCalled();
    });

    it("死亡petは更新mutation境界で拒否する", async () => {
      mockSelectedPets[0] = {
        ...mockSelectedPets[0],
        status: "死亡",
      };
      const { result } = renderHospitalizationForm("42");

      await submitForm(result.current.formAction);

      expect(mockUpdateHospitalization).not.toHaveBeenCalled();
      expect(result.current.formState.fieldErrors?.pet).toBe(
        "死亡したペットは入院情報を更新できません",
      );
    });
  });

  describe("pet status hydration", () => {
    it("petIdの直接指定で取得した死亡statusを保持し作成mutationを拒否する", async () => {
      mockSearchParams.set("petId", "pet-deceased");
      mockPetFromQuery.current = {
        id: "pet-deceased",
        ownerId: "owner-1",
        ownerName: "田中太郎",
        name: "ポチ",
        species: "犬",
        status: "死亡",
      };

      const { result } = renderHospitalizationForm();

      expect(mockSetSelectedPets).toHaveBeenCalledWith([
        expect.objectContaining({
          id: "pet-deceased",
          status: "死亡",
        }),
      ]);

      await submitForm(result.current.formAction);

      expect(mockCreateHospitalization).not.toHaveBeenCalled();
      expect(result.current.formState.fieldErrors?.pet).toBe(
        "死亡したペットは入院登録できません",
      );
    });
  });

  // ──────────────────────────
  // handleFormDataChange
  // ──────────────────────────
  describe("handleFormDataChange", () => {
    it("handleFormDataChange でフォームデータを部分更新できる", () => {
      const { result } = renderHospitalizationForm();

      act(() => {
        result.current.handleFormDataChange({ memo: "テストメモ" });
      });

      expect(result.current.formData.memo).toBe("テストメモ");
    });

    it("hospitalizationType を変更できる", () => {
      const { result } = renderHospitalizationForm();

      act(() => {
        result.current.handleFormDataChange({ hospitalizationType: "ホテル" });
      });

      expect(result.current.formData.hospitalizationType).toBe("ホテル");
    });
  });

  // ──────────────────────────
  // 保険フィールド
  // ──────────────────────────
  describe("保険フィールド", () => {
    beforeEach(() => {
      mockSelectedPets.push({
        id: "1",
        ownerId: "2",
        ownerName: "田中太郎",
        name: "ポチ",
        species: "犬",
        breed: "柴犬",
        gender: "male",
      });
    });

    afterEach(() => {
      mockSelectedPets.length = 0;
    });

    it("isInsurance = false (デフォルト) → create 時 insurance_company_name: null, insurance_number: null", async () => {
      const { result } = renderHospitalizationForm();
      await act(async () => {
        result.current.handleFormDataChange({ cageId: "10" });
      });

      await submitForm(result.current.formAction);

      expect(mockCreateHospitalization).toHaveBeenCalledWith(
        expect.objectContaining({
          insurance_company_name: null,
          insurance_number: null,
        })
      );
    });

    it("isInsurance = true → create 時 保険フィールドが渡される", async () => {
      const { result } = renderHospitalizationForm();

      act(() => {
        result.current.handleFormDataChange({
          cageId: "10",
          isInsurance: true,
          insuranceCompanyName: "アニコム損保",
          insuranceNumber: "INS-001",
        });
      });

      await submitForm(result.current.formAction);

      expect(mockCreateHospitalization).toHaveBeenCalledWith(
        expect.objectContaining({
          is_insurance: true,
          insurance_company_name: "アニコム損保",
          insurance_number: "INS-001",
        })
      );
    });

    it("isInsurance = false → create 時 保険フィールドが null になる（文字列があっても）", async () => {
      const { result } = renderHospitalizationForm();

      act(() => {
        result.current.handleFormDataChange({
          cageId: "10",
          isInsurance: false,
          insuranceCompanyName: "残留データ",
          insuranceNumber: "LEFTOVER",
        });
      });

      await submitForm(result.current.formAction);

      expect(mockCreateHospitalization).toHaveBeenCalledWith(
        expect.objectContaining({
          insurance_company_name: null,
          insurance_number: null,
        })
      );
    });

    it("isInsurance = false → update 時 保険フィールドが null になる", async () => {
      mockFoundHospitalization(42);
      mockSelectedPets.push({
        id: "1",
        ownerId: "2",
        ownerName: "田中太郎",
        name: "ポチ",
        species: "犬",
      });
      const { result } = renderHospitalizationForm("42");

      act(() => {
        result.current.handleFormDataChange({
          cageId: "10",
          isInsurance: false,
          insuranceCompanyName: "消去対象",
          insuranceNumber: "REMOVE",
        });
      });

      await submitForm(result.current.formAction);

      expect(mockUpdateHospitalization).toHaveBeenCalledWith(
        "42",
        expect.objectContaining({
          is_insurance: false,
          insurance_company_name: null,
          insurance_number: null,
        })
      );
    });

    it("isInsurance = true → update 時 保険フィールドが渡される", async () => {
      mockFoundHospitalization(42);
      mockSelectedPets.push({
        id: "1",
        ownerId: "2",
        ownerName: "田中太郎",
        name: "ポチ",
        species: "犬",
      });
      const { result } = renderHospitalizationForm("42");

      act(() => {
        result.current.handleFormDataChange({
          cageId: "10",
          isInsurance: true,
          insuranceCompanyName: "ペット保険",
          insuranceNumber: "P-999",
        });
      });

      await submitForm(result.current.formAction);

      expect(mockUpdateHospitalization).toHaveBeenCalledWith(
        "42",
        expect.objectContaining({
          is_insurance: true,
          insurance_company_name: "ペット保険",
          insurance_number: "P-999",
        })
      );
    });
  });

  // ──────────────────────────
  // treatmentPlans
  // ──────────────────────────
  describe("treatmentPlans", () => {
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

        expect(
          result.current.treatmentPlans.find((p) => p.id === firstPlanId)
        ).toBeUndefined();
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
});


// ──────────────────────────────────────────────────────────
// BUG-016: 不存在 ID / 別 clinic / network error を空 edit に潰さない
// ──────────────────────────────────────────────────────────

describe("useHospitalizationForm BUG-016 entity read", () => {
  function axiosError(status: number | undefined) {
    const config = { headers: new AxiosHeaders() } as InternalAxiosRequestConfig;
    if (status === undefined) {
      return new AxiosError("Network Error", AxiosError.ERR_NETWORK, config, undefined, undefined);
    }
    return new AxiosError(
      "request failed",
      AxiosError.ERR_BAD_RESPONSE,
      config,
      undefined,
      {
        config,
        data: { error: "not found" },
        headers: new AxiosHeaders(),
        status,
        statusText: "Error",
      },
    );
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
