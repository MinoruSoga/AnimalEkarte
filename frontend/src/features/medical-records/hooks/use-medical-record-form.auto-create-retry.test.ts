import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { renderHook, act, waitFor, render, screen } from "@testing-library/react";
import { createElement } from "react";
import { useMedicalRecordForm } from "./use-medical-record-form";
import { useGetPet } from "@/hooks/use-pet";
import { useGetOwner } from "@/hooks/use-owner";
import { useGetReservationTypesGrouped } from "@/hooks/use-reservation-types";
import { useCreateReservation } from "@/hooks/use-create-reservation";
import { useGetReservations } from "@/hooks/use-get-reservations";
import { useCreateMedicalRecord } from "../api/create-medical-record";
import { MedicalRecordAutoCreateFailure } from "../components/MedicalRecordAutoCreateFailure";

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

// FE-RC-045: use-medical-record-form.auto-create.test.ts (1219行) を分割した2ファイル目。
// 親 describe("useMedicalRecordForm") の beforeEach/afterEach を複製し、独立 describe として実行する（振る舞いは維持）。

describe("useMedicalRecordForm — 新規作成 auto-create effect (権限剥奪/死亡ペット/再試行)", () => {
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

  it("死亡が明示されたペットでは予約もカルテも自動作成しない", async () => {
    mockSearchParams = new URLSearchParams({ petId: "5" });
    vi.mocked(useGetPet).mockReturnValue({
      data: { ...mockPet, status: "死亡" },
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

    expect(mockCreateReservation).not.toHaveBeenCalled();
    expect(mockCreateRecord).not.toHaveBeenCalled();
  });

  it("appointment検索中にcreate権限を失った場合は予約もカルテも自動作成しない", async () => {
    mockSearchParams = new URLSearchParams({ petId: "5" });
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

    const { rerender } = renderHook(() => useMedicalRecordForm());
    expect(mockCreateReservation).not.toHaveBeenCalled();

    mockUsePermission.mockReturnValue({
      canView: true,
      canCreate: false,
      canEdit: true,
      canDelete: true,
    });
    vi.mocked(useGetReservations).mockReturnValue({
      data: [],
      isLoading: false,
    } as ReturnType<typeof useGetReservations>);
    await act(async () => {
      rerender();
    });

    expect(mockCreateReservation).not.toHaveBeenCalled();
    expect(mockCreateRecord).not.toHaveBeenCalled();
  });

  it("予約作成中にcreate権限を失った場合はカルテを自動作成しない", async () => {
    mockSearchParams = new URLSearchParams({ petId: "5" });
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
    let resolveAppointment: ((appointment: { id: string }) => void) | undefined;
    const appointmentPromise = new Promise<{ id: string }>((resolve) => {
      resolveAppointment = resolve;
    });
    const mockCreateReservation = vi.fn(() => appointmentPromise);
    vi.mocked(useCreateReservation).mockReturnValue({
      mutateAsync: mockCreateReservation,
      isPending: false,
    } as ReturnType<typeof useCreateReservation>);

    const { rerender } = renderHook(() => useMedicalRecordForm());
    await waitFor(() => {
      expect(mockCreateReservation).toHaveBeenCalledTimes(1);
    });

    mockUsePermission.mockReturnValue({
      canView: true,
      canCreate: false,
      canEdit: true,
      canDelete: true,
    });
    rerender();
    await act(async () => {
      resolveAppointment?.({ id: "appointment-1" });
      await appointmentPromise;
    });

    expect(mockCreateRecord).not.toHaveBeenCalled();
  });

  it("予約作成中に選択ペットが死亡へ変わった場合はカルテを自動作成しない", async () => {
    mockSearchParams = new URLSearchParams({ petId: "5" });
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
    let resolveAppointment: ((appointment: { id: string }) => void) | undefined;
    const appointmentPromise = new Promise<{ id: string }>((resolve) => {
      resolveAppointment = resolve;
    });
    const mockCreateReservation = vi.fn(() => appointmentPromise);
    vi.mocked(useCreateReservation).mockReturnValue({
      mutateAsync: mockCreateReservation,
      isPending: false,
    } as ReturnType<typeof useCreateReservation>);

    const { rerender } = renderHook(() => useMedicalRecordForm());
    await waitFor(() => {
      expect(mockCreateReservation).toHaveBeenCalledTimes(1);
    });

    vi.mocked(useGetPet).mockReturnValue({
      data: { ...mockPet, status: "死亡" },
      isLoading: false,
      isError: false,
    });
    rerender();
    await act(async () => {
      resolveAppointment?.({ id: "appointment-1" });
      await appointmentPromise;
    });

    expect(mockCreateRecord).not.toHaveBeenCalled();
  });

  it("appointment 作成失敗後は自動再試行せず、明示的な再試行で appointment phase から再開する", async () => {
    mockSearchParams = new URLSearchParams({ petId: "5" });
    vi.mocked(useGetPet).mockReturnValue({
      data: mockPet,
      isLoading: false,
      isError: false,
    });

    const mockCreateAppointment = vi
      .fn()
      .mockRejectedValueOnce(new Error("Appointment create failed"))
      .mockResolvedValueOnce({ id: "appointment-2" });
    vi.mocked(useCreateReservation).mockReturnValue({
      mutateAsync: mockCreateAppointment,
      isPending: false,
    } as ReturnType<typeof useCreateReservation>);
    const mockCreateRecord = vi.fn().mockResolvedValue({ id: "new-record-1" });
    vi.mocked(useCreateMedicalRecord).mockReturnValue({
      mutateAsync: mockCreateRecord,
      isPending: false,
    } as ReturnType<typeof useCreateMedicalRecord>);

    const { result, rerender } = renderHook(() => useMedicalRecordForm());

    await waitFor(() => {
      expect(result.current.autoCreateFailurePhase).toBe("appointment");
    });
    expect(mockCreateAppointment).toHaveBeenCalledTimes(1);
    expect(mockCreateRecord).not.toHaveBeenCalled();

    mockLocationState = { visitDate: "2026-06-02" };
    rerender();
    await act(async () => {
      await Promise.resolve();
    });
    expect(mockCreateAppointment).toHaveBeenCalledTimes(1);
    expect(mockCreateRecord).not.toHaveBeenCalled();

    render(
      createElement(MedicalRecordAutoCreateFailure, {
        failurePhase: "appointment",
        isRetrying: false,
        onRetry: () => result.current.retryAutoCreate(),
      }),
    );
    const retryButton = screen.getByRole("button", {
      name: "カルテ作成を再試行する",
    });
    expect(retryButton).toBeEnabled();

    act(() => {
      retryButton.click();
      result.current.retryAutoCreate();
    });
    await waitFor(() => {
      expect(mockCreateAppointment).toHaveBeenCalledTimes(2);
      expect(mockCreateRecord).toHaveBeenCalledTimes(1);
      expect(mockCreateRecord).toHaveBeenLastCalledWith(
        expect.objectContaining({ appointment_id: "appointment-2" }),
      );
      expect(result.current.autoCreateFailurePhase).toBeNull();
    });
  });

  it("カルテ作成失敗後の明示的な再試行は作成済み appointment_id を再利用する", async () => {
    mockSearchParams = new URLSearchParams({ petId: "5" });
    vi.mocked(useGetPet).mockReturnValue({
      data: mockPet,
      isLoading: false,
      isError: false,
    });

    const mockCreateAppointment = vi.fn().mockResolvedValue({ id: "appointment-1" });
    vi.mocked(useCreateReservation).mockReturnValue({
      mutateAsync: mockCreateAppointment,
      isPending: false,
    } as ReturnType<typeof useCreateReservation>);
    const mockCreateRecord = vi
      .fn()
      .mockRejectedValueOnce(new Error("Medical record create failed"))
      .mockResolvedValueOnce({ id: "new-record-1" });
    vi.mocked(useCreateMedicalRecord).mockReturnValue({
      mutateAsync: mockCreateRecord,
      isPending: false,
    } as ReturnType<typeof useCreateMedicalRecord>);

    const { result, rerender } = renderHook(() => useMedicalRecordForm());

    await waitFor(() => {
      expect(result.current.autoCreateFailurePhase).toBe("medical-record");
    });
    expect(mockCreateAppointment).toHaveBeenCalledTimes(1);
    expect(mockCreateRecord).toHaveBeenCalledTimes(1);
    expect(mockCreateRecord).toHaveBeenLastCalledWith(
      expect.objectContaining({ appointment_id: "appointment-1" }),
    );

    mockLocationState = { visitDate: "2026-06-02" };
    rerender();
    await act(async () => {
      await Promise.resolve();
    });
    expect(mockCreateAppointment).toHaveBeenCalledTimes(1);
    expect(mockCreateRecord).toHaveBeenCalledTimes(1);

    render(
      createElement(MedicalRecordAutoCreateFailure, {
        failurePhase: "medical-record",
        isRetrying: false,
        onRetry: () => result.current.retryAutoCreate(),
      }),
    );
    const retryButton = screen.getByRole("button", {
      name: "カルテ作成を再試行する",
    });
    expect(retryButton).toBeEnabled();

    act(() => {
      retryButton.click();
      result.current.retryAutoCreate();
    });
    await waitFor(() => {
      expect(mockCreateAppointment).toHaveBeenCalledTimes(1);
      expect(mockCreateRecord).toHaveBeenCalledTimes(2);
      expect(mockCreateRecord).toHaveBeenLastCalledWith(
        expect.objectContaining({ appointment_id: "appointment-1" }),
      );
      expect(result.current.autoCreateFailurePhase).toBeNull();
    });
  });
});
