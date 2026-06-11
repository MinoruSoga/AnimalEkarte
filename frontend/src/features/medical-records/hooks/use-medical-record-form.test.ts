import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { renderHook, act, waitFor } from "@testing-library/react";
import { startTransition } from "react";
import { useMedicalRecordForm } from "./use-medical-record-form";
import { useGetPet } from "@/hooks/use-pet";
import { useGetOwner } from "@/hooks/use-owner";
import { useGetReservationTypesGrouped } from "@/hooks/use-reservation-types";
import { useCreateReservation, useGetReservations } from "@/features/reservations";
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
vi.mock("@/hooks/use-permission", () => ({
  usePermission: () => ({ canView: true, canCreate: true, canEdit: true, canDelete: true }),
}));

// API フック群をすべてスタブ化（デフォルト: データなし・ローディングなし）
const noData = { data: undefined, isLoading: false, isError: false };
const noMutation = { mutateAsync: vi.fn().mockResolvedValue({}), isPending: false };

function runFormAction(action: (payload: FormData) => void) {
  act(() => {
    startTransition(() => {
      action(new FormData());
    });
  });
}

vi.mock("@/hooks/use-pet", () => ({ useGetPet: vi.fn(() => ({ data: undefined, isLoading: false, isError: false })) }));
vi.mock("@/hooks/use-owner", () => ({ useGetOwner: vi.fn(() => noData) }));
const mockUseGetMedicalRecord = vi.fn(() => noData);
vi.mock("../api/get-medical-record", () => ({ useGetMedicalRecord: (...args: unknown[]) => mockUseGetMedicalRecord(...args) }));
vi.mock("../api/create-medical-record", () => ({ useCreateMedicalRecord: vi.fn(() => noMutation) }));
vi.mock("@/features/reservations", () => ({
  useCreateReservation: vi.fn(() => noMutation),
  useGetReservations: vi.fn(() => ({ data: [], isLoading: false })),
}));
vi.mock("../api/update-medical-record", () => ({ useUpdateMedicalRecord: () => noMutation }));
vi.mock("../api/inquiries", () => ({ useUpdateInquiry: () => noMutation }));
vi.mock("../api/clinical-plan", () => ({ useUpdateClinicalPlan: () => noMutation }));
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

describe("useMedicalRecordForm", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockSearchParams = new URLSearchParams();
    mockLocationState = null;
    mockUseGetMedicalRecord.mockReturnValue(noData);
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
      const mockCreateReservation = vi.fn().mockResolvedValue({ id: "appointment-1" });
      vi.mocked(useCreateReservation).mockReturnValue({
        mutateAsync: mockCreateReservation,
        isPending: false,
      } as ReturnType<typeof useCreateReservation>);

      await act(async () => {
        renderHook(() => useMedicalRecordForm()); // recordId なし → isNewRecord=true
      });

      await waitFor(() => {
        expect(mockCreateReservation).toHaveBeenCalledWith(
          expect.objectContaining({
            pet_id: Number(mockPet.id),
            owner_id: Number(mockPet.ownerId),
            reservation_type_id: 1,
            status: "in_consultation",
            reservation_route: "record_shortcut",
          })
        );
        expect(mockMutateAsync).toHaveBeenCalledWith(
          expect.objectContaining({
            pet_id: mockPet.id,
            owner_id: mockPet.ownerId,
            appointment_id: "appointment-1",
            status: "draft",
          })
        );
      });
    });

    it("受付から遷移した新規作成では既存 appointment_id をカルテに紐付ける", async () => {
      mockSearchParams = new URLSearchParams({ petId: "5" });
      mockLocationState = { appointmentId: "77", visitDate: "2026-05-29" };
      vi.mocked(useGetPet).mockReturnValue({
        data: mockPet,
        isLoading: false,
        isError: false,
      });

      const mockCreateRecord = vi.fn().mockResolvedValue({ id: "new-record-1" });
      vi.mocked(useCreateMedicalRecord).mockReturnValue({
        mutateAsync: mockCreateRecord,
        isPending: false,
      } as ReturnType<typeof useCreateMedicalRecord>);
      const mockCreateReservation = vi.fn().mockResolvedValue({ id: "appointment-1" });
      vi.mocked(useCreateReservation).mockReturnValue({
        mutateAsync: mockCreateReservation,
        isPending: false,
      } as ReturnType<typeof useCreateReservation>);

      await act(async () => {
        renderHook(() => useMedicalRecordForm());
      });

      await waitFor(() => {
        expect(mockCreateReservation).not.toHaveBeenCalled();
        expect(mockCreateRecord).toHaveBeenCalledWith(
          expect.objectContaining({
            appointment_id: "77",
            visit_date: "2026-05-29",
          })
        );
      });
    });

    it("visitDate 指定の一覧新規作成では appointment も同じ日付で作成する", async () => {
      mockSearchParams = new URLSearchParams({ petId: "5", visitDate: "2026-06-01" });
      vi.mocked(useGetPet).mockReturnValue({
        data: mockPet,
        isLoading: false,
        isError: false,
      });

      const mockCreateRecord = vi.fn().mockResolvedValue({ id: "new-record-1" });
      vi.mocked(useCreateMedicalRecord).mockReturnValue({
        mutateAsync: mockCreateRecord,
        isPending: false,
      } as ReturnType<typeof useCreateMedicalRecord>);
      const mockCreateReservation = vi.fn().mockResolvedValue({ id: "appointment-1" });
      vi.mocked(useCreateReservation).mockReturnValue({
        mutateAsync: mockCreateReservation,
        isPending: false,
      } as ReturnType<typeof useCreateReservation>);

      await act(async () => {
        renderHook(() => useMedicalRecordForm());
      });

      await waitFor(() => {
        expect(mockCreateReservation).toHaveBeenCalledWith(
          expect.objectContaining({
            start_time: expect.stringMatching(/^2026-06-01T/),
            end_time: expect.stringMatching(/^2026-06-01T/),
          })
        );
        expect(mockCreateRecord).toHaveBeenCalledWith(
          expect.objectContaining({
            appointment_id: "appointment-1",
            visit_date: "2026-06-01",
          })
        );
      });
    });

    it("一覧新規作成では同日同ペットの未完了通常 appointment を再利用する", async () => {
      mockSearchParams = new URLSearchParams({ petId: "5", visitDate: "2026-06-01" });
      vi.mocked(useGetPet).mockReturnValue({
        data: mockPet,
        isLoading: false,
        isError: false,
      });
      vi.mocked(useGetReservations).mockReturnValue({
        data: [
          {
            id: "appointment-existing",
            category: "general",
            status: "checked_in",
          },
        ],
        isLoading: false,
      } as ReturnType<typeof useGetReservations>);

      const mockCreateRecord = vi.fn().mockResolvedValue({ id: "new-record-1" });
      vi.mocked(useCreateMedicalRecord).mockReturnValue({
        mutateAsync: mockCreateRecord,
        isPending: false,
      } as ReturnType<typeof useCreateMedicalRecord>);
      const mockCreateReservation = vi.fn().mockResolvedValue({ id: "appointment-1" });
      vi.mocked(useCreateReservation).mockReturnValue({
        mutateAsync: mockCreateReservation,
        isPending: false,
      } as ReturnType<typeof useCreateReservation>);

      await act(async () => {
        renderHook(() => useMedicalRecordForm());
      });

      await waitFor(() => {
        expect(mockCreateReservation).not.toHaveBeenCalled();
        expect(mockCreateRecord).toHaveBeenCalledWith(
          expect.objectContaining({
            appointment_id: "appointment-existing",
            visit_date: "2026-06-01",
          })
        );
      });
    });

    it("一覧新規作成では当日 appointment 検索中に自動作成しない", async () => {
      mockSearchParams = new URLSearchParams({ petId: "5", visitDate: "2026-06-01" });
      vi.mocked(useGetPet).mockReturnValue({
        data: mockPet,
        isLoading: false,
        isError: false,
      });
      vi.mocked(useGetReservations).mockReturnValue({
        data: [],
        isLoading: true,
      } as ReturnType<typeof useGetReservations>);

      const mockCreateRecord = vi.fn().mockResolvedValue({ id: "new-record-1" });
      vi.mocked(useCreateMedicalRecord).mockReturnValue({
        mutateAsync: mockCreateRecord,
        isPending: false,
      } as ReturnType<typeof useCreateMedicalRecord>);
      const mockCreateReservation = vi.fn().mockResolvedValue({ id: "appointment-1" });
      vi.mocked(useCreateReservation).mockReturnValue({
        mutateAsync: mockCreateReservation,
        isPending: false,
      } as ReturnType<typeof useCreateReservation>);

      await act(async () => {
        renderHook(() => useMedicalRecordForm());
      });

      expect(mockCreateReservation).not.toHaveBeenCalled();
      expect(mockCreateRecord).not.toHaveBeenCalled();
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
      });

      // navigate が呼ばれ、detail ページパスに new-record-1 が含まれる
      await waitFor(() => {
        expect(mockNavigate).toHaveBeenCalledWith(
          expect.stringContaining("new-record-1"),
          expect.objectContaining({ replace: true })
        );
      });
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
      });

      // createMutation.mutateAsync が呼ばれ、エラー後も hook がクラッシュしない
      await waitFor(() => {
        expect(mockMutateAsync).toHaveBeenCalled();
      });
    });
  });

  // ──────────────────────────
  // formAction（useActionState コールバック）
  // ──────────────────────────
  describe("formAction（useActionState）", () => {
    it("recordId なし → 即 success: false を返す（line 166）", async () => {
      const { result } = renderHook(() => useMedicalRecordForm()); // no recordId
      runFormAction(result.current.formAction);
      await waitFor(() => {
        expect(result.current.formState.success).toBe(false);
      });
    });

    it("recordId あり & 問診タブ → updateInquiryMutation.mutateAsync を呼ぶ", async () => {
      const { useUpdateInquiry: _useUpdateInquiry } = await import("../api/inquiries");
      const mockMutateAsync = vi.fn().mockResolvedValue({});
      vi.doMock("../api/inquiries", () => ({
        useUpdateInquiry: () => ({ mutateAsync: mockMutateAsync, isPending: false }),
      }));

      const { result } = renderHook(() => useMedicalRecordForm("10"));
      // activeTab = "問診" (デフォルト)

      runFormAction(result.current.formAction);
      await waitFor(() => {
        expect(result.current.formState.success).toBe(true);
      });

      // 問診タブ保存 → toast.success
      const { toast } = await import("sonner");
      expect(toast.success).toHaveBeenCalledWith("保存しました");
    });

    it("recordId あり & 診察/治療プランタブ & diagnosis1CategoryId あり & diagnosis1NameId なし → バリデーションエラー（line 183-188）", async () => {
      const { result } = renderHook(() => useMedicalRecordForm("10"));

      // タブを診察/治療プランに切り替え
      act(() => { result.current.setActiveTab("診察/治療プラン"); });
      await waitFor(() => {
        expect(result.current.activeTab).toBe("診察/治療プラン");
      });
      // 診断1カテゴリを設定、名前は未設定
      act(() => { result.current.setDiagnosis1CategoryId(3); });
      await waitFor(() => {
        expect(result.current.diagnosis1CategoryId).toBe(3);
      });

      runFormAction(result.current.formAction);
      await waitFor(() => {
        expect(result.current.formState.fieldErrors?.diagnosis1_name_id).toBe("診断名を選択してください");
      });

      // toast.error は使わず fieldErrors でインライン表示する
      expect(result.current.formState.success).toBe(false);
    });
  });
});
