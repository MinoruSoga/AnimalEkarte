import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { renderHook, act, waitFor } from "@testing-library/react";
import { startTransition } from "react";
import { useMedicalRecordForm } from "./use-medical-record-form";
import { useGetPet } from "@/hooks/use-pet";
import { useGetOwner } from "@/hooks/use-owner";
import { useGetReservationTypesGrouped } from "@/hooks/use-reservation-types";
import { useCreateReservation } from "@/hooks/use-create-reservation";
import { useGetReservations } from "@/hooks/use-get-reservations";
import { useCreateMedicalRecord } from "../api/create-medical-record";

// ──────────────────────────────────────────────────────────
// モック定義
// ──────────────────────────────────────────────────────────
//
// FE4-18: use-medical-record-form.test.ts（821 行）から describe 境界で分割。
// 新規作成 auto-create effect / formAction（useActionState）の回帰をカバーする。
// vi.mock はファイルスコープで hoist されるため、この定義ブロックは
// use-medical-record-form.test.ts と同一の内容を逐語複製している
// （値・ロジックは 1 文字も変えていない）。基本的な state/handler の回帰は
// use-medical-record-form.test.ts を参照。

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
vi.mock("@/hooks/use-create-reservation", () => ({
  useCreateReservation: vi.fn(() => noMutation),
}));
vi.mock("@/hooks/use-get-reservations", () => ({
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

    // BUG-416 ②: diagnosis1 と同じバリデーションを diagnosis2 にも適用する（FE validation parity）
    it("recordId あり & 診察/治療プランタブ & diagnosis2CategoryId あり & diagnosis2NameId なし → バリデーションエラー", async () => {
      const { result } = renderHook(() => useMedicalRecordForm("10"));

      // タブを診察/治療プランに切り替え
      act(() => { result.current.setActiveTab("診察/治療プラン"); });
      await waitFor(() => {
        expect(result.current.activeTab).toBe("診察/治療プラン");
      });
      // 診断2カテゴリを設定、名前は未設定
      act(() => { result.current.setDiagnosis2CategoryId(5); });
      await waitFor(() => {
        expect(result.current.diagnosis2CategoryId).toBe(5);
      });

      runFormAction(result.current.formAction);
      await waitFor(() => {
        expect(result.current.formState.fieldErrors?.diagnosis2_name_id).toBe("診断名を選択してください");
      });

      // toast.error は使わず fieldErrors でインライン表示する
      expect(result.current.formState.success).toBe(false);
    });

    // BUG-410: hydrate された diagnosis1/2 が実際の保存ペイロードに反映されることを、
    // state のアサーションではなく updateTreatmentPlanMutation.mutateAsync への
    // 実引数で証明する。state レベルのテスト（use-medical-record-form.test.ts）だけでは
    // render中setState(useApplyMedicalRecord)→useActionStateのクロージャ経路で
    // 古い値が送信され続ける可能性を反証できない。
    it("BUG-410: 既存レコードの diagnosis2 が hydrate 済みの状態で診断以外を編集して保存すると、mutateAsync に stale null ではなく hydrate された diagnosis_2_type_id/diagnosis_2_name_id が渡る", async () => {
      const loadedRecord = {
        data: {
          id: "10",
          visitType: "再診",
          chiefComplaint: "",
          plan: "既存の治療方針",
          assessment: "既存の所見",
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

      // hydrate が state に反映されるまで待つ（use-medical-record-form.test.ts と同一の待機点）
      await waitFor(() => {
        expect(result.current.diagnosis2CategoryId).toBe(4);
        expect(result.current.diagnosis2NameId).toBe(9);
      });

      act(() => { result.current.setActiveTab("診察/治療プラン"); });
      await waitFor(() => {
        expect(result.current.activeTab).toBe("診察/治療プラン");
      });

      // 診断とは無関係な治療方針テキストのみ編集する（診断1/2 は一切触らない）
      act(() => { result.current.setPlan("更新後の治療方針"); });

      runFormAction(result.current.formAction);

      await waitFor(() => {
        expect(result.current.formState.success).toBe(true);
      });
      // useUpdateClinicalPlan は noMutation にモックされている（clinical-plan.ts の
      // vi.mock）。診断以外の治療方針テキストしか編集していないのに、hydrate が
      // 効いていなければここに diagnosis_2_type_id: null が渡り保存済み診断2が消える。
      expect(noMutation.mutateAsync).toHaveBeenCalledWith(
        expect.objectContaining({
          diagnosis_2_type_id: 4,
          diagnosis_2_name_id: 9,
        }),
      );
    });
  });
});
