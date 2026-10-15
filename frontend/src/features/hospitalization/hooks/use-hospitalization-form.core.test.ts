import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { renderHook, act } from "@testing-library/react";
import { startTransition, useLayoutEffect, useRef } from "react";
import { calculateNextDate } from "@/components/shared/NextScheduleField";
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
function mockFoundHospitalization(id = 42, extra: Record<string, unknown> = {}) {
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
      ...extra,
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
  mockSelectedPets: [] as Record<string, unknown>[],
  mockSelectedPetsSnapshot: { current: undefined as Record<string, unknown>[] | undefined },
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
// FE-RC-045: use-hospitalization-form.test.ts (982行) を分割した1ファイル目。
// 「treatmentPlans」「BUG-016 entity read」は use-hospitalization-form.treatment-plans.test.ts へ移動。

describe("useHospitalizationForm", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    // デフォルト: ペット未選択
    mockSelectedPets.length = 0;
    mockSelectedPetsSnapshot.current = undefined;
    mockSetSelectedPets.mockImplementation((pets: Record<string, unknown>[]) => {
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

  describe("初期状態", () => {
    it("start date is JST today and explicit changes survive rerender", () => {
      vi.useFakeTimers();
      try {
        vi.setSystemTime(new Date("2026-10-14T15:01:00Z"));
        const { result, rerender } = renderHospitalizationForm();
        expect(result.current.formData.displayDate).toBe("2026-10-15");
        expect(result.current.formData.nextVisit).toBe("");
        act(() => result.current.handleFormDataChange({ displayDate: "2026-10-13" }));
        rerender();
        expect(result.current.formData.displayDate).toBe("2026-10-13");
      } finally {
        vi.useRealTimers();
      }
    });
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
      expect(result.current.formState.fieldErrors?.pet).toBe("ペットを選択してください");
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

    async function selectCage(result: {
      current: { handleFormDataChange: (u: { cageId: string }) => void };
    }) {
      await act(async () => {
        result.current.handleFormDataChange({ cageId: "10" });
      });
    }

    // EMR-316: BUG-037 の必須化を撤回。「選択なし」のまま登録でき、cage_id は送らない（NULL で保存）。
    it("ケージ未選択（選択なし）でも create され、cage_id を送らない（EMR-316）", async () => {
      const { result } = renderHospitalizationForm();

      await submitForm(result.current.formAction);

      expect(mockCreateHospitalization).toHaveBeenCalledTimes(1);
      const request = mockCreateHospitalization.mock.calls[0]?.[0] as Record<string, unknown>;
      expect(request).not.toHaveProperty("cage_id");
      expect(request).not.toHaveProperty("clear_cage_id");
      expect(result.current.formState.success).toBe(true);
      expect(result.current.formState.fieldErrors?.cage_id).toBeUndefined();
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
        }),
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
      expect(result.current.formState.fieldErrors?.pet).toBe("死亡したペットは入院登録できません");
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
      expect(result.current.formState.fieldErrors?.pet).toBe("死亡したペットは入院登録できません");
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

    async function selectCage(result: {
      current: { handleFormDataChange: (u: { cageId: string }) => void };
    }) {
      await act(async () => {
        result.current.handleFormDataChange({ cageId: "10" });
      });
    }

    // EMR-316: 編集で「選択なし」にして保存すると clear_cage_id: true を送り、cage_id は送らない。
    it("未割当で開いた後に別スタッフが割り当ててもメモ保存でケージを解除しない", async () => {
      const { result, rerender } = renderHospitalizationForm("42");
      const query = vi.mocked(useGetHospitalizationRaw)("42");
      vi.mocked(useGetHospitalizationRaw).mockReturnValue({
        ...query,
        data: { ...query.data!, cage_id: 10 },
      } as ReturnType<typeof useGetHospitalizationRaw>);
      rerender();
      await act(async () => result.current.handleFormDataChange({ memo: "updated" }));

      await submitForm(result.current.formAction);

      expect(mockUpdateHospitalization).toHaveBeenCalledTimes(1);
      const [, request] = mockUpdateHospitalization.mock.calls[0] as [
        string,
        Record<string, unknown>,
      ];
      expect(request.memo).toBe("updated");
      expect(request).not.toHaveProperty("clear_cage_id");
      expect(request).not.toHaveProperty("cage_id");
      expect(result.current.formState.fieldErrors?.cage_id).toBeUndefined();
    });

    it("割当済みケージを明示的に選択なしへ変更すると解除指示を送る", async () => {
      const query = vi.mocked(useGetHospitalizationRaw)("42");
      vi.mocked(useGetHospitalizationRaw).mockReturnValue({
        ...query,
        data: { ...query.data!, cage_id: 10 },
      } as ReturnType<typeof useGetHospitalizationRaw>);
      const { result } = renderHospitalizationForm("42");
      await act(async () => result.current.handleFormDataChange({ cageId: "" }));
      await submitForm(result.current.formAction);
      expect(mockUpdateHospitalization).toHaveBeenCalledWith(
        "42",
        expect.objectContaining({ clear_cage_id: true, expected_cage_id: "10" }),
      );
    });

    it("ケージ選択あり → update は cage_id のみ送り clear_cage_id は送らない（EMR-316）", async () => {
      const { result } = renderHospitalizationForm("42");
      await selectCage(result);

      await submitForm(result.current.formAction);

      const [, request] = mockUpdateHospitalization.mock.calls[0] as [
        string,
        Record<string, unknown>,
      ];
      expect(request.cage_id).toBe("10");
      expect(request).not.toHaveProperty("clear_cage_id");
      expect(request.expected_cage_is_null).toBe(true);
    });

    it("refetch でケージが変わっても表示時点の expected_cage_id を保持する", async () => {
      const query = vi.mocked(useGetHospitalizationRaw)("42");
      vi.mocked(useGetHospitalizationRaw).mockReturnValue({
        ...query,
        data: { ...query.data!, cage_id: 10 },
      } as ReturnType<typeof useGetHospitalizationRaw>);
      const { result, rerender } = renderHospitalizationForm("42");
      const hydrated = vi.mocked(useGetHospitalizationRaw).mock.results.at(-1)?.value;
      vi.mocked(useGetHospitalizationRaw).mockReturnValue({
        ...hydrated,
        data: { ...hydrated.data, cage_id: 20 },
      } as ReturnType<typeof useGetHospitalizationRaw>);
      rerender();
      await act(async () => result.current.handleFormDataChange({ cageId: "30" }));
      await submitForm(result.current.formAction);
      expect(mockUpdateHospitalization).toHaveBeenCalledWith(
        "42",
        expect.objectContaining({ cage_id: "30", expected_cage_id: "10" }),
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

      expect(mockUpdateHospitalization).toHaveBeenCalledWith("42", expect.any(Object));
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
      expect(result.current.formState.fieldErrors?.pet).toBe("死亡したペットは入院登録できません");
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
        }),
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
        }),
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
        }),
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
        }),
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
        }),
      );
    });
  });

  // ──────────────────────────
  // 次回予定日 (EMR-353)
  // ──────────────────────────
  describe("次回予定日の保存 (EMR-353)", () => {
    const alivePet = {
      id: "1",
      ownerId: "2",
      ownerName: "田中太郎",
      name: "ポチ",
      species: "犬",
    };

    it("create 時 nextVisit があれば next_visit として送信する", async () => {
      mockSelectedPets.push({ ...alivePet });
      const { result } = renderHospitalizationForm();

      act(() => {
        result.current.handleFormDataChange({ nextVisit: "2026-11-20" });
      });
      await submitForm(result.current.formAction);

      expect(mockCreateHospitalization).toHaveBeenCalledWith(
        expect.objectContaining({ next_visit: "2026-11-20" }),
      );
    });

    it("create 時 nextVisit 未入力なら next_visit を送らない", async () => {
      mockSelectedPets.push({ ...alivePet });
      const { result } = renderHospitalizationForm();

      await submitForm(result.current.formAction);

      const payload = mockCreateHospitalization.mock.calls[0][0];
      expect(payload).not.toHaveProperty("next_visit");
    });

    it("create 保存失敗時も nextVisit 入力値を保持し成功トーストを出さない", async () => {
      mockSelectedPets.push({ ...alivePet });
      mockCreateHospitalization.mockRejectedValueOnce(new Error("network"));
      const { result } = renderHospitalizationForm();

      act(() => {
        result.current.handleFormDataChange({ nextVisit: "2026-11-20" });
      });
      await submitForm(result.current.formAction);

      expect(result.current.formState.success).toBe(false);
      expect(result.current.formData.nextVisit).toBe("2026-11-20");
      expect(mockToast.success).not.toHaveBeenCalled();
    });

    it("編集: GET の next_visit を form へ復元し timestamp start_date は date-only に正規化する", async () => {
      mockFoundHospitalization(42, { next_visit: "2026-08-15" });
      const { result } = renderHospitalizationForm("42");

      expect(result.current.formData.nextVisit).toBe("2026-08-15");
      expect(result.current.formData.nextScheduleType).toBe("other");
      expect(result.current.formData.displayDate).toBe("2026-07-23");
    });

    it("編集: nextVisit が保存値と同じなら PATCH へ含めない", async () => {
      mockFoundHospitalization(42, { next_visit: "2026-08-15" });
      mockSelectedPets.push({ ...alivePet });
      const { result } = renderHospitalizationForm("42");

      await submitForm(result.current.formAction);

      const payload = mockUpdateHospitalization.mock.calls[0][1];
      expect(payload).not.toHaveProperty("next_visit");
      expect(payload).not.toHaveProperty("clear_next_visit");
    });

    it("編集: 保存値から日付を変更すると next_visit を送る", async () => {
      mockFoundHospitalization(42, { next_visit: "2026-08-15" });
      mockSelectedPets.push({ ...alivePet });
      const { result } = renderHospitalizationForm("42");

      act(() => {
        result.current.handleFormDataChange({ nextVisit: "2026-09-01" });
      });
      await submitForm(result.current.formAction);

      expect(mockUpdateHospitalization).toHaveBeenCalledWith(
        "42",
        expect.objectContaining({ next_visit: "2026-09-01" }),
      );
      const payload = mockUpdateHospitalization.mock.calls[0][1];
      expect(payload).not.toHaveProperty("clear_next_visit");
    });

    it("編集: 保存済みの日付を消去すると clear_next_visit: true を送る", async () => {
      mockFoundHospitalization(42, { next_visit: "2026-08-15" });
      mockSelectedPets.push({ ...alivePet });
      const { result } = renderHospitalizationForm("42");

      act(() => {
        result.current.handleFormDataChange({ nextVisit: "" });
      });
      await submitForm(result.current.formAction);

      const payload = mockUpdateHospitalization.mock.calls[0][1];
      expect(payload.clear_next_visit).toBe(true);
      expect(payload).not.toHaveProperty("next_visit");
    });

    it("編集: 保存失敗時は変更済み nextVisit を保持し成功トーストを出さない", async () => {
      mockFoundHospitalization(42, { next_visit: "2026-08-15" });
      mockSelectedPets.push({ ...alivePet });
      mockUpdateHospitalization.mockRejectedValueOnce(new Error("network"));
      const { result } = renderHospitalizationForm("42");

      act(() => {
        result.current.handleFormDataChange({ nextVisit: "2026-09-01" });
      });
      await submitForm(result.current.formAction);

      expect(result.current.formState.success).toBe(false);
      expect(result.current.formData.nextVisit).toBe("2026-09-01");
      expect(mockToast.success).not.toHaveBeenCalled();
    });
  });

  describe("開始日 start_date の保存", () => {
    const alivePet = {
      id: "1",
      ownerId: "2",
      ownerName: "田中太郎",
      name: "ポチ",
      species: "犬",
    };

    it("編集: 変更した期間を JST 0時の start_date/end_date としてプリセット再計算の next_visit と同送する", async () => {
      mockFoundHospitalization(42);
      mockSelectedPets.push({ ...alivePet });
      const { result } = renderHospitalizationForm("42");

      expect(result.current.formData.displayDate).toBe("2026-07-23");
      expect(result.current.formData.endDate).toBe("2026-07-30");

      act(() => {
        result.current.handleFormDataChange({
          displayDate: "2026-08-01",
          endDate: "2026-08-08",
          nextScheduleType: "4weeks",
          nextVisit: calculateNextDate("2026-08-01", "4weeks"),
        });
      });
      await submitForm(result.current.formAction);

      expect(mockUpdateHospitalization).toHaveBeenCalledWith(
        "42",
        expect.objectContaining({
          start_date: "2026-08-01T00:00:00+09:00",
          end_date: "2026-08-08T00:00:00+09:00",
          next_visit: "2026-08-29",
        }),
      );
      const payload = mockUpdateHospitalization.mock.calls[0][1];
      expect(payload).not.toHaveProperty("clear_next_visit");
    });

    it("編集: 保存失敗後のリトライも同じ start_date/end_date/next_visit とケージ guard を送る", async () => {
      mockFoundHospitalization(42, { cage_id: 10 });
      mockSelectedPets.push({ ...alivePet });
      mockUpdateHospitalization.mockRejectedValueOnce(new Error("network"));
      const { result } = renderHospitalizationForm("42");

      act(() => {
        result.current.handleFormDataChange({
          displayDate: "2026-08-01",
          endDate: "2026-08-08",
          nextScheduleType: "4weeks",
          nextVisit: calculateNextDate("2026-08-01", "4weeks"),
          cageId: "20",
        });
      });
      await submitForm(result.current.formAction);

      expect(result.current.formState.success).toBe(false);
      expect(mockToast.success).not.toHaveBeenCalled();

      await submitForm(result.current.formAction);

      const expectedPayload = expect.objectContaining({
        start_date: "2026-08-01T00:00:00+09:00",
        end_date: "2026-08-08T00:00:00+09:00",
        next_visit: "2026-08-29",
        cage_id: "20",
        expected_cage_id: "10",
      });
      expect(mockUpdateHospitalization).toHaveBeenNthCalledWith(1, "42", expectedPayload);
      expect(mockUpdateHospitalization).toHaveBeenNthCalledWith(2, "42", expectedPayload);
      expect(result.current.formState.success).toBe(true);
    });

    it("編集: 期間を変えずにメモだけ保存すると start_date/end_date を送らない", async () => {
      mockFoundHospitalization(42);
      mockSelectedPets.push({ ...alivePet });
      const { result } = renderHospitalizationForm("42");

      expect(result.current.formData.displayDate).toBe("2026-07-23");
      expect(result.current.formData.endDate).toBe("2026-07-30");

      await act(async () => result.current.handleFormDataChange({ memo: "updated" }));
      await submitForm(result.current.formAction);

      const payload = mockUpdateHospitalization.mock.calls[0][1];
      expect(payload.memo).toBe("updated");
      expect(payload).not.toHaveProperty("start_date");
      expect(payload).not.toHaveProperty("end_date");
    });

    it("編集: 同一idのrefetchで他スタッフが期間を更新しても stale form のメモ保存は日付を送らない", async () => {
      mockFoundHospitalization(42);
      mockSelectedPets.push({ ...alivePet });
      const { result, rerender } = renderHospitalizationForm("42");

      const hydrated = vi.mocked(useGetHospitalizationRaw).mock.results.at(-1)?.value;
      vi.mocked(useGetHospitalizationRaw).mockReturnValue({
        ...hydrated,
        data: {
          ...hydrated.data,
          start_date: "2026-08-01T00:00:00+09:00",
          end_date: "2026-08-08T00:00:00+09:00",
        },
      } as ReturnType<typeof useGetHospitalizationRaw>);
      rerender();

      expect(result.current.formData.displayDate).toBe("2026-07-23");
      expect(result.current.formData.endDate).toBe("2026-07-30");

      await act(async () => result.current.handleFormDataChange({ memo: "updated" }));
      await submitForm(result.current.formAction);

      const payload = mockUpdateHospitalization.mock.calls[0][1];
      expect(payload.memo).toBe("updated");
      expect(payload).not.toHaveProperty("start_date");
      expect(payload).not.toHaveProperty("end_date");
    });

    it("編集: 開始日だけを保存済み期間内へ変更すると start_date のみ送る", async () => {
      mockFoundHospitalization(42);
      mockSelectedPets.push({ ...alivePet });
      const { result } = renderHospitalizationForm("42");

      await act(async () => result.current.handleFormDataChange({ displayDate: "2026-07-24" }));
      await submitForm(result.current.formAction);

      const payload = mockUpdateHospitalization.mock.calls[0][1];
      expect(payload.start_date).toBe("2026-07-24T00:00:00+09:00");
      expect(payload).not.toHaveProperty("end_date");
    });

    it("編集: 終了日だけを変更すると end_date のみ送る", async () => {
      mockFoundHospitalization(42);
      mockSelectedPets.push({ ...alivePet });
      const { result } = renderHospitalizationForm("42");

      await act(async () => result.current.handleFormDataChange({ endDate: "2026-07-31" }));
      await submitForm(result.current.formAction);

      const payload = mockUpdateHospitalization.mock.calls[0][1];
      expect(payload).not.toHaveProperty("start_date");
      expect(payload.end_date).toBe("2026-07-31T00:00:00+09:00");
    });

    it("編集: 期間変更の保存成功後はメモだけの再保存で日付と次回予定日を再送しない", async () => {
      mockFoundHospitalization(42);
      mockSelectedPets.push({ ...alivePet });
      const { result } = renderHospitalizationForm("42");

      act(() => {
        result.current.handleFormDataChange({
          displayDate: "2026-08-01",
          endDate: "2026-08-08",
          nextScheduleType: "4weeks",
          nextVisit: calculateNextDate("2026-08-01", "4weeks"),
        });
      });
      await submitForm(result.current.formAction);

      expect(result.current.formState.success).toBe(true);
      expect(mockUpdateHospitalization).toHaveBeenCalledWith(
        "42",
        expect.objectContaining({
          start_date: "2026-08-01T00:00:00+09:00",
          end_date: "2026-08-08T00:00:00+09:00",
          next_visit: "2026-08-29",
        }),
      );

      await act(async () => result.current.handleFormDataChange({ memo: "updated again" }));
      await submitForm(result.current.formAction);

      expect(mockUpdateHospitalization).toHaveBeenCalledTimes(2);
      const resave = mockUpdateHospitalization.mock.calls[1][1];
      expect(resave.memo).toBe("updated again");
      expect(resave).not.toHaveProperty("start_date");
      expect(resave).not.toHaveProperty("end_date");
      expect(resave).not.toHaveProperty("next_visit");
      expect(resave).not.toHaveProperty("clear_next_visit");
      expect(resave).not.toHaveProperty("cage_id");
      expect(resave).not.toHaveProperty("clear_cage_id");
      expect(resave).not.toHaveProperty("expected_cage_id");
      expect(resave).not.toHaveProperty("expected_cage_is_null");
    });
  });
});
