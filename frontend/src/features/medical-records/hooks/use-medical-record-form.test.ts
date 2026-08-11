import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { renderHook, act, waitFor } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { AxiosError, AxiosHeaders, type InternalAxiosRequestConfig } from "axios";
import { useMedicalRecordForm } from "./use-medical-record-form";
import { useGetPet } from "@/hooks/use-pet";
import { useGetOwner } from "@/hooks/use-owner";
import { useGetReservationTypesGrouped } from "@/hooks/use-reservation-types";
import { useCreateReservation } from "@/hooks/use-create-reservation";
import { useGetReservations } from "@/hooks/use-get-reservations";
import { server } from "@/testing/mocks/node";
import { createTestWrapper } from "@/testing/utils";

// ──────────────────────────────────────────────────────────
// モック定義
// ──────────────────────────────────────────────────────────
//
// FE4-18: 821 行(分割前)だったため describe 境界で分割。新規作成 auto-create
// effect / formAction（useActionState）の回帰は
// use-medical-record-form.auto-create.test.ts を参照。vi.mock はファイル
// スコープで hoist されるため、この定義ブロックは両ファイルへ逐語複製している
// （値・ロジックは 1 文字も変えていない）。

const mockNavigate = vi.fn();
let mockSearchParams = new URLSearchParams();
let mockLocationState: Record<string, unknown> | null = null;

vi.mock("react-router", () => ({
  useNavigate: () => mockNavigate,
  useSearchParams: () => [mockSearchParams, vi.fn()],
  useLocation: () => ({ pathname: "/medical-records/10", search: "", state: mockLocationState }),
}));

vi.mock("@tanstack/react-query", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@tanstack/react-query")>();
  return {
    ...actual,
    useQueryClient: () => ({ invalidateQueries: vi.fn() }),
  };
});

vi.mock("sonner", () => ({ toast: { error: vi.fn(), success: vi.fn(), info: vi.fn() } }));
vi.mock("@/lib/handle-api-error", () => ({ handleApiError: vi.fn() }));
vi.mock("@/hooks/use-permission", () => ({
  usePermission: () => ({ canView: true, canCreate: true, canEdit: true, canDelete: true }),
}));

// API フック群をすべてスタブ化（デフォルト: データなし・ローディングなし）
const noData = { data: undefined, isLoading: false, isError: false };
const noMutation = { mutateAsync: vi.fn().mockResolvedValue({}), isPending: false };

vi.mock("@/hooks/use-pet", () => ({
  useGetPet: vi.fn(() => ({ data: undefined, isLoading: false, isError: false })),
  useGetPets: vi.fn(() => ({ data: [], isLoading: false, isError: false })),
}));
vi.mock("@/hooks/use-owner", () => ({ useGetOwner: vi.fn(() => noData) }));
const mockUseGetMedicalRecord = vi.fn(() => noData);
vi.mock("../api/get-medical-record", () => ({ useGetMedicalRecord: (...args: unknown[]) => mockUseGetMedicalRecord(...args) }));
vi.mock("../api/create-medical-record", () => ({ useCreateMedicalRecord: vi.fn(() => noMutation) }));
vi.mock("@/hooks/use-create-reservation", () => ({
  useCreateReservation: vi.fn(() => noMutation),
}));
vi.mock("@/hooks/use-get-reservations", () => ({
  useGetReservations: vi.fn(() => ({ data: [], isLoading: false })),
}));
vi.mock("../api/update-medical-record", () => ({ useUpdateMedicalRecord: () => noMutation }));
vi.mock("../api/inquiries", () => ({ useUpdateInquiry: () => noMutation }));
// BUG-416③: useGetClinicalPlan は clinical_plan の version（楽観ロック用）取得元。
const mockUseGetClinicalPlan = vi.fn(() => noData);
vi.mock("../api/clinical-plan", () => ({
  useUpdateClinicalPlan: () => noMutation,
  useGetClinicalPlan: (...args: unknown[]) => mockUseGetClinicalPlan(...args),
}));
vi.mock("@/hooks/use-reservation-types", () => ({
  useGetReservationTypesGrouped: vi.fn(() => ({
    data: [
      {
        label: "一般診療",
        types: [
          {
            id: 1,
            name: "一般診察",
            color: "#000000",
            is_active: true,
            duration_minutes: 15,
            sort_order: 1,
            is_internal: false,
            category: "general",
            group_id: 1,
            group: { id: 1, name: "一般診療", color: "#000000" },
          },
        ],
      },
    ],
  })),
}));

// ──────────────────────────────────────────────────────────
// テスト
// ──────────────────────────────────────────────────────────

describe("useGetPets enabled guard", () => {
  it.each([
    ["undefined", undefined],
    ["空文字", ""],
  ])("ownerId が%sで enabled=false のときAPIリクエストを発行しない", async (_label, ownerId) => {
    let requestCount = 0;
    server.use(
      http.get("/api/v1/pets", () => {
        requestCount += 1;
        return HttpResponse.json({ data: [] });
      }),
    );
    const { useGetPets: useActualGetPets } = await vi.importActual<
      typeof import("@/hooks/use-pet")
    >("@/hooks/use-pet");

    const { result } = renderHook(
      () => useActualGetPets(ownerId, {}, { enabled: false }),
      { wrapper: createTestWrapper() },
    );

    await waitFor(() => expect(result.current.fetchStatus).toBe("idle"));
    expect(requestCount).toBe(0);
  });
});

describe("useMedicalRecordForm", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockSearchParams = new URLSearchParams();
    mockLocationState = null;
    mockUseGetMedicalRecord.mockReturnValue(noData);
    mockUseGetClinicalPlan.mockReturnValue(noData);
    // デフォルト: pet データなし
    vi.mocked(useGetPet).mockReturnValue({ data: undefined, isLoading: false, isError: false });
    // デフォルト: owner データなし
    vi.mocked(useGetOwner).mockReturnValue(noData as never);
    vi.mocked(useGetReservationTypesGrouped).mockReturnValue({
      data: [
        {
          label: "一般診療",
          types: [
            {
              id: 1,
              name: "一般診察",
              color: "#000000",
              is_active: true,
              duration_minutes: 15,
              sort_order: 1,
              is_internal: false,
              category: "general",
              group_id: 1,
              group: { id: 1, name: "一般診療", color: "#000000" },
            },
          ],
        },
      ],
    } as ReturnType<typeof useGetReservationTypesGrouped>);
    vi.mocked(useCreateReservation).mockReturnValue({
      mutateAsync: vi.fn().mockResolvedValue({ id: "appointment-1" }),
      isPending: false,
    } as ReturnType<typeof useCreateReservation>);
    vi.mocked(useGetReservations).mockReturnValue({ data: [], isLoading: false } as ReturnType<typeof useGetReservations>);
  });

  afterEach(() => {
    vi.clearAllTimers();
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

    it("既存レコードに visitType がある場合、state に反映される", async () => {
      const loadedRecord = {
        data: {
          id: "10",
          visitType: "初診",
          chiefComplaint: "",
          plan: "",
          assessment: "",
          notes: "",
          version: 1,
        },
        isLoading: false,
        isError: false,
      };
      // まずローディング状態で renderHook し、その後データ到着をシミュレートする
      // 初回レンダー時に data が存在すると prevExistingRecord === existingRecord になり
      // useApplyMedicalRecord の差分検出が動かないため、async loading パターンを模倣する
      mockUseGetMedicalRecord.mockReturnValue({ data: undefined, isLoading: true, isError: false });
      const { result, rerender } = renderHook(() => useMedicalRecordForm("10"));

      mockUseGetMedicalRecord.mockReturnValue(loadedRecord as never);
      rerender();

      await waitFor(() => {
        expect(result.current.visitType).toBe("初診");
      });
    });

    it("BUG-406 follow-up: 既存レコードに chiefComplaintTypeId がある場合、state に反映される", async () => {
      const loadedRecord = {
        data: {
          id: "10",
          visitType: "再診",
          chiefComplaint: "",
          chiefComplaintTypeId: 5,
          plan: "",
          assessment: "",
          notes: "",
          version: 1,
        },
        isLoading: false,
        isError: false,
      };
      mockUseGetMedicalRecord.mockReturnValue({ data: undefined, isLoading: true, isError: false });
      const { result, rerender } = renderHook(() => useMedicalRecordForm("10"));

      mockUseGetMedicalRecord.mockReturnValue(loadedRecord as never);
      rerender();

      await waitFor(() => {
        expect(result.current.chiefComplaintTypeId).toBe(5);
      });
    });

    it("BUG-410: 既存レコードに diagnosis1/2 の category/name ID がある場合、state に反映される（未反映だと編集保存で無言クリアされる）", async () => {
      const loadedRecord = {
        data: {
          id: "10",
          visitType: "再診",
          chiefComplaint: "",
          plan: "",
          assessment: "",
          notes: "",
          version: 1,
          diagnosis1CategoryId: 3,
          diagnosis1NameId: 7,
          diagnosis2CategoryId: 4,
          diagnosis2NameId: 9,
        },
        isLoading: false,
        isError: false,
      };
      mockUseGetMedicalRecord.mockReturnValue({ data: undefined, isLoading: true, isError: false });
      const { result, rerender } = renderHook(() => useMedicalRecordForm("10"));

      mockUseGetMedicalRecord.mockReturnValue(loadedRecord as never);
      rerender();

      await waitFor(() => {
        expect(result.current.diagnosis1CategoryId).toBe(3);
        expect(result.current.diagnosis1NameId).toBe(7);
        expect(result.current.diagnosis2CategoryId).toBe(4);
        expect(result.current.diagnosis2NameId).toBe(9);
      });
    });

    // react-reviewer 指摘: TanStack Query のウォームキャッシュ（staleTime=5分、
    // QUERY_STALE_TIMES.MEDIUM）により、同一カルテを短時間内に再訪問すると
    // useGetMedicalRecord は「ローディング→到着」ではなく初回レンダーから
    // 既に data を返す。この場合 useState(existingRecord) で初期化される
    // prevExistingRecord が existingRecord と同一参照になり、hydrate が
    // 一度も発火しない可能性がある（render-phase setState の既知の穴）。
    it("BUG-410: ウォームキャッシュ（初回レンダーから既に既存レコードを保持）でも diagnosis1/2 が state に反映される", async () => {
      const loadedRecord = {
        data: {
          id: "10",
          visitType: "再診",
          chiefComplaint: "",
          plan: "",
          assessment: "",
          notes: "",
          version: 1,
          diagnosis1CategoryId: 3,
          diagnosis1NameId: 7,
          diagnosis2CategoryId: 4,
          diagnosis2NameId: 9,
        },
        isLoading: false,
        isError: false,
      };
      // ローディング状態を経由せず、初回レンダーから既にデータ到着済みにする
      // （TanStack Query のウォームキャッシュ相当）。
      mockUseGetMedicalRecord.mockReturnValue(loadedRecord as never);
      const { result } = renderHook(() => useMedicalRecordForm("10"));

      await waitFor(() => {
        expect(result.current.diagnosis1CategoryId).toBe(3);
        expect(result.current.diagnosis1NameId).toBe(7);
        expect(result.current.diagnosis2CategoryId).toBe(4);
        expect(result.current.diagnosis2NameId).toBe(9);
      });
    });

    // BUG-034: 問診「治療方針」は inquiry.notes → treatmentPolicy。
    // detail wire で notes が落ちると再読込後 DEFAULT「# 治療方針」に戻る。
    it("BUG-034: 既存レコード notes を treatmentPolicy に hydrate する（確定後再読込でも保持）", async () => {
      const loadedRecord = {
        data: {
          id: "1425558",
          visitType: "再診",
          chiefComplaint: "UAT再検証 主訴",
          plan: "",
          assessment: "",
          notes: "UAT再検証 治療方針",
          version: 2,
          status: "確定済",
        },
        isLoading: false,
        isError: false,
      };
      mockUseGetMedicalRecord.mockReturnValue({ data: undefined, isLoading: true, isError: false });
      const { result, rerender } = renderHook(() => useMedicalRecordForm("1425558"));

      mockUseGetMedicalRecord.mockReturnValue(loadedRecord as never);
      rerender();

      await waitFor(() => {
        expect(result.current.chiefComplaint).toBe("UAT再検証 主訴");
        expect(result.current.treatmentPolicy).toBe("UAT再検証 治療方針");
      });
    });

    it("BUG-034: notes が空のとき treatmentPolicy は DEFAULT のまま（clinical_plan.plan と混同しない）", async () => {
      const loadedRecord = {
        data: {
          id: "10",
          chiefComplaint: "主訴のみ",
          plan: "診察タブ側の治療方針テキスト",
          assessment: "",
          notes: "",
          version: 1,
        },
        isLoading: false,
        isError: false,
      };
      mockUseGetMedicalRecord.mockReturnValue({ data: undefined, isLoading: true, isError: false });
      const { result, rerender } = renderHook(() => useMedicalRecordForm("10"));

      mockUseGetMedicalRecord.mockReturnValue(loadedRecord as never);
      rerender();

      await waitFor(() => {
        expect(result.current.chiefComplaint).toBe("主訴のみ");
      });
      // notes が空なので setTreatmentPolicy は呼ばれず DEFAULT のまま
      expect(result.current.treatmentPolicy).toBe("# 治療方針");
      // medical-record wire の plan は clinical_plan 正本ではない（useApply は truthy plan を setPlan）
      expect(result.current.plan).toBe("診察タブ側の治療方針テキスト");
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
      await waitFor(() => {
        expect(noMutation.mutateAsync).toHaveBeenCalledWith(
          expect.objectContaining({
            id: "10",
            req: expect.objectContaining({ doctor_id: 3 }),
          }),
        );
      });
    });
  });

  // ──────────────────────────
  // 飼主変更（requestOwnerChange / confirmOwnerChange / cancelOwnerChange）
  // ──────────────────────────
  describe("飼主変更", () => {
    it("recordId なしの場合、confirmOwnerChange は mutateAsync を呼ばない", async () => {
      const { result } = renderHook(() => useMedicalRecordForm());
      // owner が undefined のため needsConfirm = true → dialog が表示される
      act(() => { result.current.requestOwnerChange({ id: "5", name: "佐藤", discountRate: 0, membershipType: "" }); });
      // pendingOwnerChange がセットされる
      expect(result.current.pendingOwnerChange).toEqual({ id: "5", name: "佐藤" });
      // recordId なしのため confirm しても mutation は呼ばれない
      await act(async () => { result.current.confirmOwnerChange(); });
      expect(noMutation.mutateAsync).not.toHaveBeenCalled();
    });

    it("cancelOwnerChange で pending をリセットできる", () => {
      const { result } = renderHook(() => useMedicalRecordForm("10"));
      act(() => { result.current.requestOwnerChange({ id: "5", name: "佐藤", discountRate: 0, membershipType: "" }); });
      expect(result.current.pendingOwnerChange).not.toBeNull();
      act(() => { result.current.cancelOwnerChange(); });
      expect(result.current.pendingOwnerChange).toBeNull();
    });
  });

  // ──────────────────────────
  // BUG-373: requestOwnerChange 条件付き確認ダイアログ
  // ──────────────────────────
  describe("BUG-373: requestOwnerChange — 飼主変更 条件付き確認", () => {
    const ownerQueryWith = (discountRate: number, membershipType: string) => ({
      data: { discountRate, membershipType } as never,
      isLoading: false,
      isError: false,
    } as never);

    it("owner 未ロード(undefined) → 安全策として dialog を表示する", () => {
      // デフォルト: useGetOwner → noData (data: undefined)
      const { result } = renderHook(() => useMedicalRecordForm("10"));
      act(() => {
        result.current.requestOwnerChange({ id: "5", name: "鈴木", discountRate: 10, membershipType: "会員" });
      });
      expect(result.current.pendingOwnerChange).toEqual({ id: "5", name: "鈴木" });
    });

    it("discount_rate・membershipType が同じ → dialog なし・mutation を即時実行", async () => {
      vi.mocked(useGetOwner).mockReturnValue(ownerQueryWith(10, "会員"));
      const { result } = renderHook(() => useMedicalRecordForm("10"));
      await act(async () => {
        result.current.requestOwnerChange({ id: "5", name: "鈴木", discountRate: 10, membershipType: "会員" });
        await Promise.resolve();
      });
      expect(result.current.pendingOwnerChange).toBeNull();
      expect(noMutation.mutateAsync).toHaveBeenCalled();
    });

    it("discount_rate が異なる → dialog を表示する", () => {
      vi.mocked(useGetOwner).mockReturnValue(ownerQueryWith(10, "会員"));
      const { result } = renderHook(() => useMedicalRecordForm("10"));
      act(() => {
        result.current.requestOwnerChange({ id: "5", name: "鈴木", discountRate: 20, membershipType: "会員" });
      });
      expect(result.current.pendingOwnerChange).toEqual({ id: "5", name: "鈴木" });
    });

    it("membershipType が異なる → dialog を表示する", () => {
      vi.mocked(useGetOwner).mockReturnValue(ownerQueryWith(10, "会員"));
      const { result } = renderHook(() => useMedicalRecordForm("10"));
      act(() => {
        result.current.requestOwnerChange({ id: "5", name: "鈴木", discountRate: 10, membershipType: "非会員" });
      });
      expect(result.current.pendingOwnerChange).toEqual({ id: "5", name: "鈴木" });
    });

    it("confirmOwnerChange → pendingOwnerChange をクリアし mutation を呼ぶ", async () => {
      vi.mocked(useGetOwner).mockReturnValue(ownerQueryWith(10, "会員"));
      const { result } = renderHook(() => useMedicalRecordForm("10"));
      // discount_rate 違い → dialog 表示
      act(() => {
        result.current.requestOwnerChange({ id: "5", name: "鈴木", discountRate: 20, membershipType: "会員" });
      });
      expect(result.current.pendingOwnerChange).not.toBeNull();
      // 続行 → mutation 実行
      await act(async () => {
        result.current.confirmOwnerChange();
        await Promise.resolve();
      });
      expect(result.current.pendingOwnerChange).toBeNull();
      expect(noMutation.mutateAsync).toHaveBeenCalled();
    });
  });

  // ──────────────────────────
  // BUG-017: missing chart ID read gate
  // ──────────────────────────
  describe("BUG-017: 存在しないカルテIDの read gate", () => {
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

    it("loading → isReadLoading、notFound ではない", () => {
      mockUseGetMedicalRecord.mockReturnValue({
        data: undefined,
        isLoading: true,
        isError: false,
        error: null,
        refetch: vi.fn(),
      } as never);
      const { result } = renderHook(() => useMedicalRecordForm("999999999"));
      expect(result.current.isReadLoading).toBe(true);
      expect(result.current.isReadNotFound).toBe(false);
      expect(result.current.notFound).toBe(false);
      expect(result.current.isReadError).toBe(false);
    });

    it("404 → isReadNotFound / notFound", () => {
      mockUseGetMedicalRecord.mockReturnValue({
        data: undefined,
        isLoading: false,
        isError: true,
        error: axiosError(404),
        refetch: vi.fn(),
      } as never);
      const { result } = renderHook(() => useMedicalRecordForm("999999999"));
      expect(result.current.isReadNotFound).toBe(true);
      expect(result.current.notFound).toBe(true);
      expect(result.current.isReadError).toBe(false);
      expect(result.current.isReadLoading).toBe(false);
    });

    it("403 → 404 と同一の非開示 (isReadNotFound)", () => {
      mockUseGetMedicalRecord.mockReturnValue({
        data: undefined,
        isLoading: false,
        isError: true,
        error: axiosError(403),
        refetch: vi.fn(),
      } as never);
      const { result } = renderHook(() => useMedicalRecordForm("42"));
      expect(result.current.isReadNotFound).toBe(true);
      expect(result.current.notFound).toBe(true);
      expect(result.current.isReadError).toBe(false);
    });

    it("network error → isReadError + retry、404 へ偽装しない", () => {
      const refetch = vi.fn();
      mockUseGetMedicalRecord.mockReturnValue({
        data: undefined,
        isLoading: false,
        isError: true,
        error: axiosError(undefined),
        refetch,
      } as never);
      const { result } = renderHook(() => useMedicalRecordForm("999"));
      expect(result.current.isReadError).toBe(true);
      expect(result.current.isReadNotFound).toBe(false);
      expect(result.current.notFound).toBe(false);
      expect(result.current.retryRead).toBeTypeOf("function");
      result.current.retryRead?.();
      expect(refetch).toHaveBeenCalledTimes(1);
    });

    it("settled で data なし（isError=false）→ notFound（空白にしない）", () => {
      mockUseGetMedicalRecord.mockReturnValue({
        data: undefined,
        isLoading: false,
        isError: false,
        error: null,
        refetch: vi.fn(),
      } as never);
      const { result } = renderHook(() => useMedicalRecordForm("999999999"));
      expect(result.current.isReadNotFound).toBe(true);
      expect(result.current.notFound).toBe(true);
    });

    it("found → notFound ではない", () => {
      mockUseGetMedicalRecord.mockReturnValue({
        data: {
          id: "10",
          petId: "pet-1",
          visitType: "再診",
          status: "作成中",
          version: 1,
        },
        isLoading: false,
        isError: false,
        error: null,
        refetch: vi.fn(),
      } as never);
      const { result } = renderHook(() => useMedicalRecordForm("10"));
      expect(result.current.isReadNotFound).toBe(false);
      expect(result.current.notFound).toBe(false);
      expect(result.current.isReadLoading).toBe(false);
      expect(result.current.isReadError).toBe(false);
    });
  });
});
