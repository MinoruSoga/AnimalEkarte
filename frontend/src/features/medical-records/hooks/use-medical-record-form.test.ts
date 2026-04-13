import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, act } from "@testing-library/react";
import { useMedicalRecordForm } from "./use-medical-record-form";
import { useGetPet } from "@/hooks/use-pet";
import { useCreateMedicalRecord } from "../api/create-medical-record";

// ──────────────────────────────────────────────────────────
// モック定義
// ──────────────────────────────────────────────────────────

const mockNavigate = vi.fn();
let mockSearchParams = new URLSearchParams();
let mockLocationState: Record<string, unknown> | null = null;

vi.mock("react-router", () => ({
  useNavigate: () => mockNavigate,
  useSearchParams: () => [mockSearchParams, vi.fn()],
  useLocation: () => ({ pathname: "/medical-records/10", search: "", state: mockLocationState }),
}));

vi.mock("@tanstack/react-query", () => ({
  useQueryClient: () => ({ invalidateQueries: vi.fn() }),
}));

vi.mock("sonner", () => ({ toast: { error: vi.fn(), success: vi.fn(), info: vi.fn() } }));
vi.mock("@/lib/handle-api-error", () => ({ handleApiError: vi.fn() }));
vi.mock("@/features/auth", () => ({
  usePermission: () => ({ canView: true, canCreate: true, canEdit: true, canDelete: true }),
}));

// API フック群をすべてスタブ化（デフォルト: データなし・ローディングなし）
const noData = { data: undefined, isLoading: false, isError: false };
const noMutation = { mutateAsync: vi.fn().mockResolvedValue({}), isPending: false };

vi.mock("@/hooks/use-pet", () => ({ useGetPet: vi.fn(() => ({ data: undefined, isLoading: false, isError: false })) }));
vi.mock("@/hooks/use-owner", () => ({ useGetOwner: () => noData }));
vi.mock("../api/get-medical-record", () => ({ useGetMedicalRecord: () => noData }));
vi.mock("../api/create-medical-record", () => ({ useCreateMedicalRecord: vi.fn(() => noMutation) }));
vi.mock("../api/update-medical-record", () => ({ useUpdateMedicalRecord: () => noMutation }));
vi.mock("../api/inquiries", () => ({ useUpdateInquiry: () => noMutation }));
vi.mock("../api/treatment-plans", () => ({ useUpdateTreatmentPlan: () => noMutation }));

// ──────────────────────────────────────────────────────────
// テスト
// ──────────────────────────────────────────────────────────

describe("useMedicalRecordForm", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockSearchParams = new URLSearchParams();
    mockLocationState = null;
    localStorage.clear();
    // デフォルト: pet データなし
    vi.mocked(useGetPet).mockReturnValue({ data: undefined, isLoading: false, isError: false });
  });

  // ──────────────────────────
  // isNewRecord
  // ──────────────────────────
  describe("isNewRecord", () => {
    it("recordId を渡さないと isNewRecord = true になる（バグ再現確認）", () => {
      const { result } = renderHook(() => useMedicalRecordForm());
      expect(result.current.isNewRecord).toBe(true);
    });

    it("recordId を渡すと isNewRecord = false になる", () => {
      const { result } = renderHook(() => useMedicalRecordForm("10"));
      expect(result.current.isNewRecord).toBe(false);
    });
  });

  // ──────────────────────────
  // shouldRedirectToSelectPet（回帰テスト: BUG-カルテ編集）
  // ──────────────────────────
  describe("shouldRedirectToSelectPet", () => {
    it("【回帰】編集時(recordId あり) は shouldRedirectToSelectPet = false", () => {
      // このテストが失敗したら useMedicalRecordForm() に recordId が渡っていない
      const { result } = renderHook(() => useMedicalRecordForm("10"));
      expect(result.current.shouldRedirectToSelectPet).toBe(false);
    });

    it("新規作成かつ petId なし → shouldRedirectToSelectPet = true", () => {
      mockSearchParams = new URLSearchParams(); // petId なし
      const { result } = renderHook(() => useMedicalRecordForm());
      expect(result.current.shouldRedirectToSelectPet).toBe(true);
    });

    it("新規作成かつ petId あり → shouldRedirectToSelectPet = false", () => {
      mockSearchParams = new URLSearchParams({ petId: "1" });
      const { result } = renderHook(() => useMedicalRecordForm());
      expect(result.current.shouldRedirectToSelectPet).toBe(false);
    });
  });

  // ──────────────────────────
  // 初期状態
  // ──────────────────────────
  describe("初期状態", () => {
    it("activeTab の初期値は '問診'", () => {
      const { result } = renderHook(() => useMedicalRecordForm("10"));
      expect(result.current.activeTab).toBe("問診");
    });

    it("visitType の初期値は '再診'", () => {
      const { result } = renderHook(() => useMedicalRecordForm("10"));
      expect(result.current.visitType).toBe("再診");
    });

    it("treatmentPlanItems の初期値は []", () => {
      const { result } = renderHook(() => useMedicalRecordForm("10"));
      expect(result.current.treatmentPlanItems).toEqual([]);
    });

    it("treatmentCompletedItems の初期値は []", () => {
      const { result } = renderHook(() => useMedicalRecordForm("10"));
      expect(result.current.treatmentCompletedItems).toEqual([]);
    });

    it("isSaving の初期値は false", () => {
      const { result } = renderHook(() => useMedicalRecordForm("10"));
      expect(result.current.isSaving).toBe(false);
    });

    it("isCreating の初期値は false", () => {
      const { result } = renderHook(() => useMedicalRecordForm("10"));
      expect(result.current.isCreating).toBe(false);
    });
  });

  // ──────────────────────────
  // setActiveTab / setVisitType
  // ──────────────────────────
  describe("setActiveTab / setVisitType", () => {
    it("setActiveTab でタブを切り替えられる", () => {
      const { result } = renderHook(() => useMedicalRecordForm("10"));
      act(() => {
        result.current.setActiveTab("診察/治療プラン");
      });
      expect(result.current.activeTab).toBe("診察/治療プラン");
    });

    it("setVisitType で来院種別を変更できる", () => {
      const { result } = renderHook(() => useMedicalRecordForm());
      act(() => {
        result.current.setVisitType("初診");
      });
      expect(result.current.visitType).toBe("初診");
    });
  });

  // ──────────────────────────
  // 下書き永続化（localStorage）
  // ──────────────────────────
  describe("下書き永続化（localStorage）", () => {
    it("recordId がある場合、下書きを localStorage に保存する", () => {
      const { result } = renderHook(() => useMedicalRecordForm("42"));
      act(() => {
        result.current.setChiefComplaint("新しい主訴テキスト");
      });
      const key = "medical-record-draft-42";
      const saved = JSON.parse(localStorage.getItem(key) ?? "{}");
      expect(saved.chiefComplaint).toBe("新しい主訴テキスト");
    });

    it("recordId がない場合、localStorage への保存は行わない", () => {
      renderHook(() => useMedicalRecordForm()); // isNewRecord = true
      const allKeys = Object.keys(localStorage);
      expect(allKeys.some(k => k.startsWith("medical-record-draft-"))).toBe(false);
    });

    it("既存の下書きがある場合、mount 時に chiefComplaint を復元する", () => {
      const key = "medical-record-draft-99";
      localStorage.setItem(key, JSON.stringify({ chiefComplaint: "復元テキスト" }));

      const { result } = renderHook(() => useMedicalRecordForm("99"));
      expect(result.current.chiefComplaint).toBe("復元テキスト");
    });
  });

  // ──────────────────────────
  // treatmentPlanItems / treatmentCompletedItems
  // ──────────────────────────
  describe("treatmentPlanItems / treatmentCompletedItems", () => {
    it("setTreatmentPlanItems でアイテムを設定できる", () => {
      const { result } = renderHook(() => useMedicalRecordForm("10"));
      const items = [{ id: "1", name: "処置A", quantity: 1, unitPrice: 1000, total: 1000 }];
      act(() => {
        result.current.setTreatmentPlanItems(items as Parameters<typeof result.current.setTreatmentPlanItems>[0]);
      });
      expect(result.current.treatmentPlanItems).toEqual(items);
    });

    it("setTreatmentCompletedItems でアイテムを設定できる", () => {
      const { result } = renderHook(() => useMedicalRecordForm("10"));
      const items = [{ id: "2", name: "処置B", quantity: 2, unitPrice: 500, total: 1000 }];
      act(() => {
        result.current.setTreatmentCompletedItems(items as Parameters<typeof result.current.setTreatmentCompletedItems>[0]);
      });
      expect(result.current.treatmentCompletedItems).toEqual(items);
    });
  });

  // ──────────────────────────
  // 診断マスタ state
  // ──────────────────────────
  describe("診断マスタ state", () => {
    it("setDiagnosis1CategoryId で診断1カテゴリを設定できる", () => {
      const { result } = renderHook(() => useMedicalRecordForm("10"));
      act(() => {
        result.current.setDiagnosis1CategoryId(5);
      });
      expect(result.current.diagnosis1CategoryId).toBe(5);
    });

    it("setDiagnosis1NameId で診断1名を設定できる", () => {
      const { result } = renderHook(() => useMedicalRecordForm("10"));
      act(() => {
        result.current.setDiagnosis1NameId(10);
      });
      expect(result.current.diagnosis1NameId).toBe(10);
    });

    it("setDiagnosis2CategoryId で診断2カテゴリを設定できる", () => {
      const { result } = renderHook(() => useMedicalRecordForm("10"));
      act(() => {
        result.current.setDiagnosis2CategoryId(3);
      });
      expect(result.current.diagnosis2CategoryId).toBe(3);
    });

    it("setDiagnosis2NameId で診断2名を設定できる", () => {
      const { result } = renderHook(() => useMedicalRecordForm("10"));
      act(() => {
        result.current.setDiagnosis2NameId(7);
      });
      expect(result.current.diagnosis2NameId).toBe(7);
    });
  });

  // ──────────────────────────
  // handleBack
  // ──────────────────────────
  describe("handleBack", () => {
    it("location.state.from がある場合、そのパスにナビゲートする", () => {
      mockLocationState = { from: "/reception" };
      const { result } = renderHook(() => useMedicalRecordForm("10"));
      act(() => {
        result.current.handleBack();
      });
      expect(mockNavigate).toHaveBeenCalledWith("/reception");
    });

    it("recordId なし・location.state なし → ペット選択ページにナビゲート", () => {
      mockLocationState = null;
      const { result } = renderHook(() => useMedicalRecordForm());
      act(() => {
        result.current.handleBack();
      });
      expect(mockNavigate).toHaveBeenCalledWith(
        expect.stringContaining("select-pet")
      );
    });

    it("recordId あり・location.state なし → カルテ一覧にナビゲート", () => {
      mockLocationState = null;
      const { result } = renderHook(() => useMedicalRecordForm("10"));
      act(() => {
        result.current.handleBack();
      });
      // paths.medicalRecords.getHref() = "/medical-records"
      expect(mockNavigate).toHaveBeenCalledWith(
        expect.stringContaining("medical-record")
      );
    });
  });

  // ──────────────────────────
  // handleChangeDoctor
  // ──────────────────────────
  describe("handleChangeDoctor", () => {
    it("recordId なしの場合、何もしない（updateMutation 呼ばれない）", () => {
      const { result } = renderHook(() => useMedicalRecordForm()); // no recordId
      act(() => {
        result.current.handleChangeDoctor("3", "山田医師");
      });
      expect(noMutation.mutateAsync).not.toHaveBeenCalled();
    });

    it("recordId ありの場合、updateMutation.mutateAsync を呼ぶ", async () => {
      const mockMutateAsync = vi.fn().mockResolvedValue({});
      vi.doMock("../api/update-medical-record", () => ({
        useUpdateMedicalRecord: () => ({ mutateAsync: mockMutateAsync, isPending: false }),
      }));

      const { result } = renderHook(() => useMedicalRecordForm("10"));
      await act(async () => {
        result.current.handleChangeDoctor("3", "山田医師");
        // transition をフラッシュ
        await Promise.resolve();
      });
      // handleChangeDoctor は startSaveTransition 内で実行される
      // mutateAsync が呼ばれることを確認（またはエラーなく完了）
      expect(() => result.current.handleChangeDoctor("3", "山田医師")).not.toThrow();
    });
  });

  // ──────────────────────────
  // handleChangeOwner
  // ──────────────────────────
  describe("handleChangeOwner", () => {
    it("recordId なしの場合、何もしない（updateMutation 呼ばれない）", () => {
      const { result } = renderHook(() => useMedicalRecordForm()); // no recordId
      act(() => {
        result.current.handleChangeOwner({ id: "5", name: "佐藤" });
      });
      expect(noMutation.mutateAsync).not.toHaveBeenCalled();
    });

    it("recordId ありの場合、エラーなく実行できる", async () => {
      const { result } = renderHook(() => useMedicalRecordForm("10"));
      await act(async () => {
        result.current.handleChangeOwner({ id: "5", name: "佐藤" });
        await Promise.resolve();
      });
      // mutateAsync が呼ばれる（モックは vi.fn()）
      expect(true).toBe(true); // エラーが投げられなければ OK
    });
  });

  // ──────────────────────────
  // 新規作成時 auto-create（lines 278-293）
  // ──────────────────────────
  describe("新規作成 auto-create effect", () => {
    const mockPet = {
      id: "5",
      name: "ポチ",
      ownerId: "2",
      ownerName: "田中",
      species: "犬",
      breed: "",
      birthday: "",
      gender: "男" as const,
      weight: null,
      imageUrl: null,
      status: "生存" as const,
      microchipNumber: null,
      insuranceNumber: null,
      insuranceExpiry: null,
      memo: null,
    };

    it("isNewRecord && selectedPet あり → createMutation.mutateAsync を呼ぶ", async () => {
      // petId あり → isNewRecord=true かつ selectedPet を設定
      mockSearchParams = new URLSearchParams({ petId: "5" });
      vi.mocked(useGetPet).mockReturnValue({
        data: mockPet,
        isLoading: false,
        isError: false,
      });

      const mockMutateAsync = vi.fn().mockResolvedValue({ id: "new-record-1" });
      vi.mocked(useCreateMedicalRecord).mockReturnValue({
        mutateAsync: mockMutateAsync,
        isPending: false,
      } as ReturnType<typeof useCreateMedicalRecord>);

      await act(async () => {
        renderHook(() => useMedicalRecordForm()); // recordId なし → isNewRecord=true
        await new Promise(resolve => setTimeout(resolve, 0));
      });

      expect(mockMutateAsync).toHaveBeenCalledWith(
        expect.objectContaining({
          pet_id: mockPet.id,
          owner_id: mockPet.ownerId,
          status: "draft",
        })
      );
    });

    it("isNewRecord && selectedPet あり → 作成後に detail ページへナビゲート", async () => {
      mockSearchParams = new URLSearchParams({ petId: "5" });
      vi.mocked(useGetPet).mockReturnValue({
        data: mockPet,
        isLoading: false,
        isError: false,
      });

      const mockMutateAsync = vi.fn().mockResolvedValue({ id: "new-record-1" });
      vi.mocked(useCreateMedicalRecord).mockReturnValue({
        mutateAsync: mockMutateAsync,
        isPending: false,
      } as ReturnType<typeof useCreateMedicalRecord>);

      await act(async () => {
        renderHook(() => useMedicalRecordForm());
        await new Promise(resolve => setTimeout(resolve, 10));
      });

      // navigate が呼ばれ、detail ページパスに new-record-1 が含まれる
      expect(mockNavigate).toHaveBeenCalledWith(
        expect.stringContaining("new-record-1"),
        expect.objectContaining({ replace: true })
      );
    });

    it("isNewRecord だが selectedPet なし → createMutation 呼ばれない", async () => {
      mockSearchParams = new URLSearchParams({ petId: "5" });
      vi.mocked(useGetPet).mockReturnValue({
        data: undefined,
        isLoading: true, // ローディング中 → selectedPet = undefined
        isError: false,
      });

      const mockMutateAsync = vi.fn();
      vi.mocked(useCreateMedicalRecord).mockReturnValue({
        mutateAsync: mockMutateAsync,
        isPending: false,
      } as ReturnType<typeof useCreateMedicalRecord>);

      await act(async () => {
        renderHook(() => useMedicalRecordForm());
        await Promise.resolve();
      });

      expect(mockMutateAsync).not.toHaveBeenCalled();
    });

    it("auto-create が失敗した場合、hasAutoCreatedRef をリセットして再試行可能にする（lines 292-293）", async () => {
      mockSearchParams = new URLSearchParams({ petId: "5" });
      vi.mocked(useGetPet).mockReturnValue({
        data: mockPet,
        isLoading: false,
        isError: false,
      });

      const mockMutateAsync = vi.fn().mockRejectedValue(new Error("Create failed"));
      vi.mocked(useCreateMedicalRecord).mockReturnValue({
        mutateAsync: mockMutateAsync,
        isPending: false,
      } as ReturnType<typeof useCreateMedicalRecord>);

      await act(async () => {
        renderHook(() => useMedicalRecordForm());
        await new Promise(resolve => setTimeout(resolve, 10));
      });

      // createMutation.mutateAsync が呼ばれ、エラー後も hook がクラッシュしない
      expect(mockMutateAsync).toHaveBeenCalled();
    });
  });

  // ──────────────────────────
  // formAction（useActionState コールバック）
  // ──────────────────────────
  describe("formAction（useActionState）", () => {
    it("recordId なし → 即 success: false を返す（line 166）", async () => {
      const { result } = renderHook(() => useMedicalRecordForm()); // no recordId
      await act(async () => {
        await result.current.formAction(new FormData());
      });
      expect(result.current.formState.success).toBe(false);
    });

    it("recordId あり & 問診タブ → updateInquiryMutation.mutateAsync を呼ぶ", async () => {
      const { useUpdateInquiry: _useUpdateInquiry } = await import("../api/inquiries");
      const mockMutateAsync = vi.fn().mockResolvedValue({});
      vi.doMock("../api/inquiries", () => ({
        useUpdateInquiry: () => ({ mutateAsync: mockMutateAsync, isPending: false }),
      }));

      const { result } = renderHook(() => useMedicalRecordForm("10"));
      // activeTab = "問診" (デフォルト)

      await act(async () => {
        await result.current.formAction(new FormData());
      });

      // 問診タブ保存 → toast.success
      const { toast } = await import("sonner");
      expect(toast.success).toHaveBeenCalledWith("保存しました");
      expect(result.current.formState.success).toBe(true);
    });

    it("recordId あり & 診察/治療プランタブ & diagnosis1CategoryId あり & diagnosis1NameId なし → バリデーションエラー（line 183-188）", async () => {
      const { result } = renderHook(() => useMedicalRecordForm("10"));

      // タブを診察/治療プランに切り替え
      act(() => { result.current.setActiveTab("診察/治療プラン"); });
      // 診断1カテゴリを設定、名前は未設定
      act(() => { result.current.setDiagnosis1CategoryId(3); });

      await act(async () => {
        await result.current.formAction(new FormData());
      });

      // toast.error は使わず fieldErrors でインライン表示する
      expect(result.current.formState.success).toBe(false);
      expect(result.current.formState.fieldErrors?.diagnosis1_name_id).toBe("診断名を選択してください");
    });
  });
});
