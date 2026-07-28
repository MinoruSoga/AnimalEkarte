import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, act } from "@testing-library/react";
import { startTransition, useLayoutEffect, useRef } from "react";
import { calculateBillingTotals } from "@/lib/calculations";
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
} = vi.hoisted(() => ({
  mockNavigate: vi.fn(),
  mockToast: { error: vi.fn(), success: vi.fn() },
  mockSelectedPets: [] as unknown[],
  mockSelectedPetsSnapshot: { current: undefined as unknown[] | undefined },
  mockSetSelectedPets: vi.fn(),
  mockSearchParams: new URLSearchParams(),
  mockPetFromQuery: { current: undefined as unknown },
  mockCreateHospitalization: vi.fn().mockResolvedValue({}),
  mockUpdateHospitalization: vi.fn().mockResolvedValue({}),
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

vi.mock("../api/get-hospitalization", () => ({
  useGetHospitalizationRaw: vi.fn(() => ({
    data: undefined,
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
    mockCreateHospitalization.mockResolvedValue({});
    mockUpdateHospitalization.mockResolvedValue({});
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

    it("selectedPets がある & id なし → createHospitalization が呼ばれる", async () => {
      const { result } = renderHospitalizationForm();

      await submitForm(result.current.formAction);

      expect(mockCreateHospitalization).toHaveBeenCalledTimes(1);
      expect(mockUpdateHospitalization).not.toHaveBeenCalled();
    });

    it("createHospitalization に pet_id と owner_id が渡される", async () => {
      const { result } = renderHospitalizationForm();

      await submitForm(result.current.formAction);

      expect(mockCreateHospitalization).toHaveBeenCalledWith(
        expect.objectContaining({
          pet_id: "1",
          owner_id: "2",
        })
      );
    });

    it("成功時 → toast.success が呼ばれる", async () => {
      const { result } = renderHospitalizationForm();

      await submitForm(result.current.formAction);

      expect(mockToast.success).toHaveBeenCalledWith("入院情報を登録しました");
    });

    it("成功時 → formState.success = true", async () => {
      const { result } = renderHospitalizationForm();

      await submitForm(result.current.formAction);

      expect(result.current.formState.success).toBe(true);
    });

    it("createHospitalization が失敗した場合 → formState.success = false", async () => {
      mockCreateHospitalization.mockRejectedValueOnce(new Error("API Error"));

      const { result } = renderHospitalizationForm();

      await submitForm(result.current.formAction);

      expect(result.current.formState.success).toBe(false);
    });

    it("作成権限なし → createHospitalization は呼ばれない", async () => {
      const { result } = renderHospitalizationForm(undefined, false);

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

    it("selectedPets がある & id あり → updateHospitalization が呼ばれる", async () => {
      const { result } = renderHospitalizationForm("42");

      await submitForm(result.current.formAction);

      expect(mockUpdateHospitalization).toHaveBeenCalledTimes(1);
      expect(mockCreateHospitalization).not.toHaveBeenCalled();
    });

    it("updateHospitalization に id が渡される", async () => {
      const { result } = renderHospitalizationForm("42");

      await submitForm(result.current.formAction);

      expect(mockUpdateHospitalization).toHaveBeenCalledWith(
        "42",
        expect.any(Object)
      );
    });

    it("成功時 → toast.success が呼ばれる（更新メッセージ）", async () => {
      const { result } = renderHospitalizationForm("42");

      await submitForm(result.current.formAction);

      expect(mockToast.success).toHaveBeenCalledWith("入院情報を更新しました");
    });

    it("成功時 → formState.success = true", async () => {
      const { result } = renderHospitalizationForm("42");

      await submitForm(result.current.formAction);

      expect(result.current.formState.success).toBe(true);
    });

    it("updateHospitalization が失敗した場合 → formState.success = false", async () => {
      mockUpdateHospitalization.mockRejectedValueOnce(new Error("API Error"));

      const { result } = renderHospitalizationForm("42");

      await submitForm(result.current.formAction);

      expect(result.current.formState.success).toBe(false);
    });

    it("編集権限なし → updateHospitalization は呼ばれない", async () => {
      const { result } = renderHospitalizationForm("42", false);

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
      const { result } = renderHospitalizationForm("42");

      act(() => {
        result.current.handleFormDataChange({
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
      const { result } = renderHospitalizationForm("42");

      act(() => {
        result.current.handleFormDataChange({
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
  });
});
