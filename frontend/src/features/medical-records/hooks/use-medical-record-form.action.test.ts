import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { renderHook, act, waitFor } from "@testing-library/react";
import { startTransition } from "react";
import { useMedicalRecordForm } from "./use-medical-record-form";
import { useGetPet } from "@/hooks/use-pet";
import { useGetOwner } from "@/hooks/use-owner";
import { useGetReservationTypesGrouped } from "@/hooks/use-reservation-types";
import { useCreateReservation } from "@/hooks/use-create-reservation";
import { useGetReservations } from "@/hooks/use-get-reservations";

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
const mockUsePermission = vi.hoisted(() => vi.fn());
vi.mock("@/hooks/use-permission", () => ({
  usePermission: () => mockUsePermission(),
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

vi.mock("@/hooks/use-pet", () => ({
  useGetPet: vi.fn(() => ({ data: undefined, isLoading: false, isError: false })),
  useGetPets: vi.fn(() => ({ data: [], isLoading: false, isError: false })),
}));
vi.mock("@/hooks/use-owner", () => ({ useGetOwner: vi.fn(() => noData) }));
const mockUseGetMedicalRecord = vi.fn(() => noData);
vi.mock("../api/get-medical-record", () => ({
  useGetMedicalRecord: (...args: unknown[]) => mockUseGetMedicalRecord(...args),
}));
vi.mock("../api/create-medical-record", () => ({
  useCreateMedicalRecord: vi.fn(() => noMutation),
}));
vi.mock("@/hooks/use-create-reservation", () => ({
  useCreateReservation: vi.fn(() => noMutation),
}));
vi.mock("@/hooks/use-get-reservations", () => ({
  useGetReservations: vi.fn(() => ({ data: [], isLoading: false })),
}));
vi.mock("../api/update-medical-record", () => ({ useUpdateMedicalRecord: () => noMutation }));
vi.mock("../api/inquiries", () => ({ useUpdateInquiry: () => noMutation }));
// BUG-416③: useGetClinicalPlan は clinical_plan の version（楽観ロック用）取得元。
// per-test で version を差し替えられるよう mockUseGetMedicalRecord と同じ間接呼び出しパターンにする。
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

// FE-RC-045: use-medical-record-form.auto-create.test.ts (1219行) を分割した3ファイル目。
// 親 describe("useMedicalRecordForm") の beforeEach/afterEach を複製し、独立 describe として実行する（振る舞いは維持）。

describe("useMedicalRecordForm — formAction（useActionState）", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUsePermission.mockReturnValue({
      canView: true,
      canCreate: true,
      canEdit: true,
      canDelete: true,
    });
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
    vi.mocked(useGetReservations).mockReturnValue({ data: [], isLoading: false } as ReturnType<
      typeof useGetReservations
    >);
  });

  afterEach(() => {
    vi.clearAllTimers();
    vi.useRealTimers();
  });

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

  it("確定済みカルテはprogrammatic formActionでも更新mutationを呼ばない", async () => {
    mockUseGetMedicalRecord.mockReturnValue({
      data: {
        id: "10",
        status: "確定済",
        visitType: "再診",
        chiefComplaint: "",
        plan: "",
        assessment: "",
        notes: "",
        nextVisitRecommendedDate: "",
        version: 1,
      },
      isLoading: false,
      isError: false,
    } as never);
    const { result } = renderHook(() => useMedicalRecordForm("10"));

    runFormAction(result.current.formAction);

    await waitFor(() => {
      expect(result.current.formState.timestamp).toBeGreaterThan(0);
    });
    expect(result.current.formState.success).toBe(false);
    expect(noMutation.mutateAsync).not.toHaveBeenCalled();
    const { toast } = await import("sonner");
    expect(toast.success).not.toHaveBeenCalled();
  });

  it("canEdit=falseではprogrammatic formActionでも更新mutationを呼ばない", async () => {
    mockUsePermission.mockReturnValue({
      canView: true,
      canCreate: false,
      canEdit: false,
      canDelete: false,
    });
    mockUseGetMedicalRecord.mockReturnValue({
      data: {
        id: "10",
        status: "作成中",
        petId: "5",
        visitType: "再診",
        chiefComplaint: "",
        plan: "",
        assessment: "",
        notes: "",
        nextVisitRecommendedDate: "",
        version: 1,
      },
      isLoading: false,
      isError: false,
    } as never);
    const { result } = renderHook(() => useMedicalRecordForm("10"));

    runFormAction(result.current.formAction);

    await waitFor(() => {
      expect(result.current.formState.timestamp).toBeGreaterThan(0);
    });
    expect(result.current.formState.success).toBe(false);
    expect(noMutation.mutateAsync).not.toHaveBeenCalled();
    const { toast } = await import("sonner");
    expect(toast.success).not.toHaveBeenCalled();
  });

  it("死亡した選択ペットの既存カルテはprogrammatic formActionでも更新mutationを呼ばない", async () => {
    mockUseGetMedicalRecord.mockReturnValue({
      data: {
        id: "10",
        status: "作成中",
        petId: "5",
        visitType: "再診",
        chiefComplaint: "",
        plan: "",
        assessment: "",
        notes: "",
        nextVisitRecommendedDate: "",
        version: 1,
      },
      isLoading: false,
      isError: false,
    } as never);
    vi.mocked(useGetPet).mockReturnValue({
      data: {
        id: "5",
        ownerId: "2",
        status: "死亡",
      },
      isLoading: false,
      isError: false,
    } as ReturnType<typeof useGetPet>);
    const { result } = renderHook(() => useMedicalRecordForm("10"));

    runFormAction(result.current.formAction);

    await waitFor(() => {
      expect(result.current.formState.timestamp).toBeGreaterThan(0);
    });
    expect(result.current.formState.success).toBe(false);
    expect(noMutation.mutateAsync).not.toHaveBeenCalled();
  });

  it("recordId あり & 診察/治療プランタブ & diagnosis1CategoryId あり & diagnosis1NameId なし → バリデーションエラー（line 183-188）", async () => {
    // BUG-010: version 未取得時は診断バリデーションより先に hydrate ガードで止まるため version を用意する
    mockUseGetClinicalPlan.mockReturnValue({
      data: {
        id: "cp-1",
        medical_record_id: "10",
        version: 1,
        physical_exam: "",
        diagnosis_details: "",
        treatment_policy: "",
        updated_at: "t0",
      },
      isLoading: false,
      isError: false,
    });
    const { result } = renderHook(() => useMedicalRecordForm("10"));

    // タブを診察/治療プランに切り替え
    act(() => {
      result.current.setActiveTab("診察/治療プラン");
    });
    await waitFor(() => {
      expect(result.current.activeTab).toBe("診察/治療プラン");
    });
    // 診断1カテゴリを設定、名前は未設定
    act(() => {
      result.current.setDiagnosis1CategoryId(3);
    });
    await waitFor(() => {
      expect(result.current.diagnosis1CategoryId).toBe(3);
    });

    runFormAction(result.current.formAction);
    await waitFor(() => {
      expect(result.current.formState.fieldErrors?.diagnosis1_name_id).toBe(
        "診断名を選択してください",
      );
    });

    // toast.error は使わず fieldErrors でインライン表示する
    expect(result.current.formState.success).toBe(false);
  });

  // BUG-416 ②: diagnosis1 と同じバリデーションを diagnosis2 にも適用する（FE validation parity）
  it("recordId あり & 診察/治療プランタブ & diagnosis2CategoryId あり & diagnosis2NameId なし → バリデーションエラー", async () => {
    mockUseGetClinicalPlan.mockReturnValue({
      data: {
        id: "cp-1",
        medical_record_id: "10",
        version: 1,
        physical_exam: "",
        diagnosis_details: "",
        treatment_policy: "",
        updated_at: "t0",
      },
      isLoading: false,
      isError: false,
    });
    const { result } = renderHook(() => useMedicalRecordForm("10"));

    // タブを診察/治療プランに切り替え
    act(() => {
      result.current.setActiveTab("診察/治療プラン");
    });
    await waitFor(() => {
      expect(result.current.activeTab).toBe("診察/治療プラン");
    });
    // 診断2カテゴリを設定、名前は未設定
    act(() => {
      result.current.setDiagnosis2CategoryId(5);
    });
    await waitFor(() => {
      expect(result.current.diagnosis2CategoryId).toBe(5);
    });

    runFormAction(result.current.formAction);
    await waitFor(() => {
      expect(result.current.formState.fieldErrors?.diagnosis2_name_id).toBe(
        "診断名を選択してください",
      );
    });

    // toast.error は使わず fieldErrors でインライン表示する
    expect(result.current.formState.success).toBe(false);
  });

  // BUG-410: hydrate された diagnosis1/2 が実際の保存ペイロードに反映されることを、
  // state のアサーションではなく updateTreatmentPlanMutation.mutateAsync への
  // 実引数で証明する。BUG-010 後は clinical-plan GET が診断マスタの正本。
  it("BUG-410: 既存レコードの diagnosis2 が hydrate 済みの状態で診断以外を編集して保存すると、mutateAsync に stale null ではなく hydrate された diagnosis_2_type_id/diagnosis_2_name_id が渡る", async () => {
    mockUseGetMedicalRecord.mockReturnValue({
      data: { id: "10", visitType: "再診", version: 1 },
      isLoading: false,
      isError: false,
    } as never);
    mockUseGetClinicalPlan.mockReturnValue({
      data: {
        id: "cp-10",
        medical_record_id: "10",
        version: 3,
        physical_exam: "既存所見",
        diagnosis_details: "既存の診断詳細",
        treatment_policy: "既存の治療方針",
        diagnosis_type_id: "3",
        diagnosis_name_id: "7",
        diagnosis_2_type_id: "4",
        diagnosis_2_name_id: "9",
        updated_at: "t1",
      },
      isLoading: false,
      isError: false,
    });
    const { result } = renderHook(() => useMedicalRecordForm("10"));

    // clinical-plan hydrate が state に反映されるまで待つ
    await waitFor(() => {
      expect(result.current.diagnosis2CategoryId).toBe(4);
      expect(result.current.diagnosis2NameId).toBe(9);
    });

    act(() => {
      result.current.setActiveTab("診察/治療プラン");
    });
    await waitFor(() => {
      expect(result.current.activeTab).toBe("診察/治療プラン");
    });

    // 診断とは無関係な治療方針テキストのみ編集する（診断1/2 は一切触らない）
    act(() => {
      result.current.setPlan("更新後の治療方針");
    });

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
        physical_exam: "既存所見",
        treatment_policy: "更新後の治療方針",
        version: 3,
      }),
    );
  });

  // BUG-416③: clinical_plan PATCH の楽観ロック。GET .../clinical-plan から取得した
  // version を保存ペイロードにそのまま渡すことを、updateTreatmentPlanMutation.mutateAsync
  // への実引数で証明する（他のユーザーによる同時編集を BE 側の 409 で検知できるようにする配線）。
  it("BUG-416③: 診察/治療プラン保存時、useGetClinicalPlan の version が updateTreatmentPlanMutation.mutateAsync のペイロードに渡る", async () => {
    mockUseGetClinicalPlan.mockReturnValue({
      data: { version: 5 },
      isLoading: false,
      isError: false,
    });

    const { result } = renderHook(() => useMedicalRecordForm("10"));

    act(() => {
      result.current.setActiveTab("診察/治療プラン");
    });
    await waitFor(() => {
      expect(result.current.activeTab).toBe("診察/治療プラン");
    });

    runFormAction(result.current.formAction);

    await waitFor(() => {
      expect(result.current.formState.success).toBe(true);
    });

    expect(noMutation.mutateAsync).toHaveBeenCalledWith(expect.objectContaining({ version: 5 }));
  });

  it("F2: 編集中に同一カルテの診察プランが v2 へ再取得されても、v1 由来の入力は v1 の CAS version で保存する", async () => {
    mockUseGetMedicalRecord.mockReturnValue({
      data: { id: "10", visitType: "再診", version: 1 },
      isLoading: false,
      isError: false,
    } as never);
    mockUseGetClinicalPlan.mockReturnValue({
      data: {
        id: "cp-10",
        medical_record_id: "10",
        version: 1,
        physical_exam: "v1 身体所見",
        diagnosis_details: "v1 診断詳細",
        treatment_policy: "v1 治療方針",
        diagnosis_type_id: "3",
        diagnosis_name_id: "7",
        diagnosis_2_type_id: "4",
        diagnosis_2_name_id: "9",
        updated_at: "v1",
      },
      isLoading: false,
      isError: false,
    });
    const { result, rerender } = renderHook(() => useMedicalRecordForm("10"));

    await waitFor(() => {
      expect(result.current.plan).toBe("v1 治療方針");
    });
    act(() => {
      result.current.setPlan("A の編集値");
      result.current.setActiveTab("診察/治療プラン");
    });

    mockUseGetClinicalPlan.mockReturnValue({
      data: {
        id: "cp-10",
        medical_record_id: "10",
        version: 2,
        physical_exam: "B の身体所見",
        diagnosis_details: "B の診断詳細",
        treatment_policy: "B の治療方針",
        diagnosis_type_id: "11",
        diagnosis_name_id: "12",
        diagnosis_2_type_id: "13",
        diagnosis_2_name_id: "14",
        updated_at: "v2",
      },
      isLoading: false,
      isError: false,
    });
    rerender();

    await waitFor(() => {
      expect(result.current.plan).toBe("A の編集値");
    });
    runFormAction(result.current.formAction);

    await waitFor(() => {
      expect(result.current.formState.success).toBe(true);
    });
    expect(noMutation.mutateAsync).toHaveBeenCalledWith(
      expect.objectContaining({
        treatment_policy: "A の編集値",
        version: 1,
      }),
    );
  });
});
