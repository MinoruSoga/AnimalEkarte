import { act, renderHook } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { formatJSTWallDate } from "@/lib/jst-date";
import type { Reservation } from "../types";

import { useReservationManagement } from "./use-reservation-management";

/**
 * EMR-308 A3/A4: use-reservation-management の tentative 回りを固定する。
 * - 期限超過 tentative 専用クエリ（status=tentative / 2000-01-01〜昨日）
 * - confirm-tentative / 再設定 / 終了 の fail-closed ハンドラ
 * - useReservationActions への onTentativeCancelled 配線
 */

const {
  getReservationsMock,
  confirmMutateAsync,
  setIntentMutateAsync,
  clearIntentMutateAsync,
  toastErrorMock,
  toastSuccessMock,
  actionsArgsHolder,
} = vi.hoisted(() => ({
  getReservationsMock: vi.fn(),
  confirmMutateAsync: vi.fn(),
  setIntentMutateAsync: vi.fn(),
  clearIntentMutateAsync: vi.fn(),
  toastErrorMock: vi.fn(),
  toastSuccessMock: vi.fn(),
  actionsArgsHolder: { current: null as Record<string, unknown> | null },
}));

vi.mock("react-router", () => ({
  useNavigate: () => vi.fn(),
  useLocation: () => ({ state: null, search: "" }),
}));

vi.mock("sonner", () => ({
  toast: { error: toastErrorMock, success: toastSuccessMock },
}));

vi.mock("@/hooks/use-get-reservations", () => ({
  useGetReservations: (filters: Record<string, unknown>) => getReservationsMock(filters),
}));

vi.mock("@/hooks/use-confirm-tentative", () => ({
  useConfirmTentative: () => ({ mutateAsync: confirmMutateAsync, isPending: false }),
}));

vi.mock("@/hooks/use-next-visit-intent", () => ({
  useSetNextVisitIntent: () => ({ mutateAsync: setIntentMutateAsync, isPending: false }),
  useClearNextVisitIntent: () => ({ mutateAsync: clearIntentMutateAsync, isPending: false }),
}));

vi.mock("./use-reservation-actions", () => ({
  useReservationActions: (args: Record<string, unknown>) => {
    actionsArgsHolder.current = args;
    return {
      handleSave: vi.fn(),
      handleReservationUpdate: vi.fn(),
      handleStatusChange: vi.fn(),
      executeStatusChange: vi.fn(),
      executeDelete: vi.fn(),
      handleCloseCreateForm: vi.fn(),
    };
  },
}));

vi.mock("./use-reservation-modal-state", () => ({
  useReservationModalState: () => ({
    isFormOpen: false,
    editingAppointment: null,
    editingAppointmentRef: { current: null },
    handleOpenForm: vi.fn(),
    handleCloseForm: vi.fn(),
    isDetailOpen: false,
    detailAppointment: null,
    setDetailAppointment: vi.fn(),
    handleOpenDetail: vi.fn(),
    handleCloseDetail: vi.fn(),
  }),
}));

vi.mock("./use-reservation-record-navigation", () => ({
  useReservationRecordNavigation: () => ({
    petSelectConfirmOpen: false,
    setPetSelectConfirmOpen: vi.fn(),
    handleCreateRecord: vi.fn(),
    handlePetSelectConfirm: vi.fn(),
  }),
}));

function makeReservation(
  overrides: { id: string } & Partial<Omit<Reservation, "id">>,
): Reservation {
  return {
    start: new Date("2026-06-15T09:00:00"),
    end: new Date("2026-06-15T10:00:00"),
    ownerName: "山田",
    petName: "ポチ",
    petType: undefined,
    visitType: "revisit",
    type: "一般診察",
    category: "general",
    reservationTypeId: "1",
    doctor: "",
    doctorId: undefined,
    isDesignated: false,
    status: "confirmed",
    notes: undefined,
    petId: "10",
    ownerId: "20",
    source: "manual",
    reservationRoute: "record_shortcut",
    clinicId: undefined,
    isTentative: false,
    intendedDate: undefined,
    hadIntendedDate: false,
    sourceMedicalRecordId: undefined,
    updatedAt: "2026-10-11T00:00:00.123456Z",
    ...overrides,
  };
}

const ALL_PERMISSIONS = { canCreate: true, canEdit: true, canDelete: true };

function renderManagement(permissions: typeof ALL_PERMISSIONS | undefined = ALL_PERMISSIONS) {
  return renderHook(() =>
    useReservationManagement({
      currentDate: new Date(2026, 5, 15),
      view: "week",
      days: 7,
      createMutations: {
        createOwnerFn: vi.fn().mockRejectedValue(new Error("unexpected owner creation")),
        createPetFn: vi.fn().mockRejectedValue(new Error("unexpected pet creation")),
      },
      permissions,
    }),
  );
}

describe("useReservationManagement — 期限超過 tentative クエリ (EMR-308 A3)", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    actionsArgsHolder.current = null;
    // 既定: メイン範囲は空、tentative 範囲クエリは status フィルタで識別して返す
    getReservationsMock.mockImplementation((filters: Record<string, unknown>) => ({
      data:
        filters.status === "tentative"
          ? [
              makeReservation({
                id: "expired",
                isTentative: true,
                intendedDate: "2000-01-01",
                start: new Date("2000-01-01T00:00:00+09:00"),
                end: new Date("2000-01-01T00:15:00+09:00"),
              }),
              makeReservation({
                id: "future-tentative-in-range",
                isTentative: true,
                intendedDate: "2999-12-31",
              }),
            ]
          : [],
      isLoading: false,
    }));
  });

  it("status=tentative / 2000-01-01〜昨日（JST）の範囲で第2クエリを発行する", () => {
    renderManagement();

    const calls = getReservationsMock.mock.calls.map(([f]) => f);
    const expiredQuery = calls.find((f) => f.status === "tentative");
    expect(expiredQuery).toBeDefined();
    expect(expiredQuery).toMatchObject({
      status: "tentative",
      startDate: "2000-01-01",
    });
    // endDate は JST 暦日の昨日（当日 tentative は期限超過に含めない）
    const yesterday = new Date();
    yesterday.setDate(yesterday.getDate() - 1);
    expect(expiredQuery?.endDate).toBe(formatJSTWallDate(yesterday));
  });

  it("expiredTentativeAppointments は意図日が過去の行だけを残す（fail-safe 再判定）", () => {
    const { result } = renderManagement();

    expect(result.current.expiredTentativeAppointments.map((r) => r.id)).toEqual(["expired"]);
  });
});

describe("useReservationManagement — tentative ハンドラ (EMR-308)", () => {
  const tentative = () =>
    makeReservation({
      id: "t-1",
      isTentative: true,
      intendedDate: "2999-12-31",
      sourceMedicalRecordId: "99",
    });

  beforeEach(() => {
    vi.clearAllMocks();
    actionsArgsHolder.current = null;
    getReservationsMock.mockReturnValue({ data: [], isLoading: false });
    confirmMutateAsync.mockResolvedValue({});
    setIntentMutateAsync.mockResolvedValue({});
    clearIntentMutateAsync.mockResolvedValue(undefined);
  });

  it("handleOpenConfirmTentative: canEdit なしは fail-closed で target を開かない", () => {
    const { result } = renderManagement({ ...ALL_PERMISSIONS, canEdit: false });

    act(() => result.current.handleOpenConfirmTentative(tentative()));

    expect(result.current.confirmTentativeTarget).toBeNull();
    expect(toastErrorMock).toHaveBeenCalledWith("予約を確定する権限がありません");
  });

  it("handleOpenConfirmTentative: tentative 行のみ target を開く（非 tentative は無視）", () => {
    const { result } = renderManagement();

    act(() => result.current.handleOpenConfirmTentative(makeReservation({ id: "real" })));
    expect(result.current.confirmTentativeTarget).toBeNull();

    act(() => result.current.handleOpenConfirmTentative(tentative()));
    expect(result.current.confirmTentativeTarget?.id).toBe("t-1");
  });

  it("handleConfirmTentative は confirm-tentative mutation を呼び成功で閉じる", async () => {
    const { result } = renderManagement();
    act(() => result.current.handleOpenConfirmTentative(tentative()));

    const formData = { start: new Date(2999, 11, 31, 10, 0), end: new Date(2999, 11, 31, 11, 0) };
    await act(async () => {
      const inline = await result.current.handleConfirmTentative("t-1", formData);
      expect(inline).toBeNull();
    });

    expect(confirmMutateAsync).toHaveBeenCalledWith({
      reservationId: "t-1",
      formData,
    });
    expect(result.current.confirmTentativeTarget).toBeNull();
  });

  it("handleConfirmTentative 失敗時はインラインエラーを返しモーダルを開いたまま", async () => {
    confirmMutateAsync.mockRejectedValue(new Error("boom"));
    const { result } = renderManagement();
    act(() => result.current.handleOpenConfirmTentative(tentative()));

    await act(async () => {
      const inline = await result.current.handleConfirmTentative("t-1", {
        start: new Date(2999, 11, 31, 10, 0),
        end: new Date(2999, 11, 31, 11, 0),
      });
      expect(inline).toBe("予約の確定に失敗しました。入力内容を確認して再度お試しください。");
    });
    expect(result.current.confirmTentativeTarget?.id).toBe("t-1");
  });

  it("handleResetNextVisitRequest: active tentative は canEdit なしで fail-closed（Codex #18）", () => {
    // 既存 tentative 行の再設定は PUT 上書き（BE: ルート reservations:create +
    // repo reservations:edit の両方を要求）。canEdit 側が fail-closed の判定。
    const { result } = renderManagement({ ...ALL_PERMISSIONS, canEdit: false });

    act(() => result.current.handleResetNextVisitRequest(tentative()));

    expect(result.current.resetNextVisitTarget).toBeNull();
    expect(toastErrorMock).toHaveBeenCalledWith("次回予定を再設定する権限がありません");
  });

  it("handleResetNextVisitRequest: active tentative は canCreate なしでも fail-closed（Codex #18 再指摘）", () => {
    // 上書き PUT はルートの reservations:create も通るため、edit 単独では
    // BE が 403 を返す。片方欠けたらダイアログを開かせない。
    const { result } = renderManagement({ ...ALL_PERMISSIONS, canCreate: false });

    act(() => result.current.handleResetNextVisitRequest(tentative()));

    expect(result.current.resetNextVisitTarget).toBeNull();
    expect(toastErrorMock).toHaveBeenCalledWith("次回予定を再設定する権限がありません");
  });

  it("handleResetNextVisitRequest: 非 tentative 行（キャンセル後再作成）は canCreate で判定", () => {
    // status=cancelled の行は isTentative=false → 新規 intent 扱い（BE: reservations:create）。
    const cancelled = makeReservation({
      id: "cancelled-1",
      status: "cancelled",
      hadIntendedDate: true,
      intendedDate: "2026-06-20",
    });

    const { result: denied } = renderManagement({ ...ALL_PERMISSIONS, canCreate: false });
    act(() => denied.current.handleResetNextVisitRequest(cancelled));
    expect(denied.current.resetNextVisitTarget).toBeNull();
    expect(toastErrorMock).toHaveBeenCalledWith("次回予定を再設定する権限がありません");

    toastErrorMock.mockClear();
    const { result: allowed } = renderManagement({ ...ALL_PERMISSIONS, canEdit: false });
    act(() => allowed.current.handleResetNextVisitRequest(cancelled));
    expect(allowed.current.resetNextVisitTarget).not.toBeNull();
    expect(toastErrorMock).not.toHaveBeenCalled();
  });

  it("handleResetNextVisitRequest → submit で PUT intent（source カルテ id 引継ぎ）", async () => {
    const { result } = renderManagement();

    act(() => result.current.handleResetNextVisitRequest(tentative()));
    expect(result.current.resetNextVisitTarget).toEqual({
      petId: "10",
      petName: "ポチ",
      previousIntendedDate: "2999-12-31",
      sourceMedicalRecordId: "99",
    });

    await act(async () => {
      await result.current.handleResetNextVisitSubmit("2027-01-15");
    });
    expect(setIntentMutateAsync).toHaveBeenCalledWith({
      petId: "10",
      intendedDate: "2027-01-15",
      medicalRecordId: "99",
    });
    // 再設定ダイアログの送信中フラグは setIntentMutation.isPending をそのまま露出する
    // （手動 useState ではなく React Query の mutation 状態が正本）。
    expect(result.current.isResetNextVisitSubmitting).toBe(false);
  });

  it("拠点横断行は行の clinicId で tentative mutation を送る（Codex #5）", async () => {
    const { result } = renderManagement();
    const crossClinicTentative = makeReservation({
      id: "t-x",
      isTentative: true,
      intendedDate: "2999-12-31",
      sourceMedicalRecordId: "99",
      clinicId: "clinic-77",
    });

    // 再設定 submit → PUT intent は行の clinicId
    act(() => result.current.handleResetNextVisitRequest(crossClinicTentative));
    await act(async () => {
      await result.current.handleResetNextVisitSubmit("2027-01-15");
    });
    expect(setIntentMutateAsync).toHaveBeenCalledWith(
      expect.objectContaining({ petId: "10", clinicId: "clinic-77" }),
    );

    // 終了 → DELETE intent も行の clinicId
    act(() => result.current.handleEndTentativeRequest(crossClinicTentative));
    await act(async () => {
      await result.current.executeEndTentative();
    });
    expect(clearIntentMutateAsync).toHaveBeenCalledWith({
      petId: "10",
      clinicId: "clinic-77",
      expectedReservationId: "t-x",
      expectedUpdatedAt: "2026-10-11T00:00:00.123456Z",
    });

    // 確定 → confirm-tentative も行の clinicId
    act(() => result.current.handleOpenConfirmTentative(crossClinicTentative));
    await act(async () => {
      await result.current.handleConfirmTentative("t-x", {
        start: new Date(2999, 11, 31, 10, 0),
        end: new Date(2999, 11, 31, 11, 0),
      });
    });
    expect(confirmMutateAsync).toHaveBeenCalledWith(
      expect.objectContaining({ reservationId: "t-x", clinicId: "clinic-77" }),
    );
  });

  it("handleEndTentativeRequest: canEdit なしは fail-closed", () => {
    const { result } = renderManagement({ ...ALL_PERMISSIONS, canEdit: false });

    act(() => result.current.handleEndTentativeRequest(tentative()));

    expect(result.current.endTentativeTarget).toBeNull();
    expect(toastErrorMock).toHaveBeenCalledWith("次回予定を終了する権限がありません");
  });

  it("executeEndTentative は DELETE intent（petId スコープ）を呼び target を閉じる", async () => {
    const { result } = renderManagement();

    act(() => result.current.handleEndTentativeRequest(tentative()));
    expect(result.current.endTentativeTarget?.id).toBe("t-1");

    await act(async () => {
      await result.current.executeEndTentative();
    });
    // Codex PR606 #5: DELETE も行の clinicId を含む variables 形で送る
    expect(clearIntentMutateAsync).toHaveBeenCalledWith({
      petId: "10",
      clinicId: undefined,
      expectedReservationId: "t-1",
      expectedUpdatedAt: "2026-10-11T00:00:00.123456Z",
    });
    expect(result.current.endTentativeTarget).toBeNull();
  });

  it("useReservationActions の onTentativeCancelled は再設定ダイアログ入口に配線される", () => {
    const { result } = renderManagement();
    const onTentativeCancelled = actionsArgsHolder.current?.onTentativeCancelled as (
      reservation: Reservation,
    ) => void;
    expect(typeof onTentativeCancelled).toBe("function");

    // tentative 由来のキャンセル完了 Reservation → 再設定 target が開く（自動再作成しない）
    act(() =>
      onTentativeCancelled(
        makeReservation({
          id: "cancelled-1",
          status: "cancelled",
          hadIntendedDate: true,
          intendedDate: "2026-06-20",
          sourceMedicalRecordId: "99",
        }),
      ),
    );
    expect(result.current.resetNextVisitTarget).toEqual({
      petId: "10",
      petName: "ポチ",
      previousIntendedDate: "2026-06-20",
      sourceMedicalRecordId: "99",
    });
    expect(setIntentMutateAsync).not.toHaveBeenCalled();
  });
});

// PR606: a confirmation must identify exactly the intent displayed to the user.
describe("guarded intent termination", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    getReservationsMock.mockReturnValue({ data: [] });
    clearIntentMutateAsync.mockResolvedValue(undefined);
  });
  it("sends displayed ID and original nanosecond timestamp", async () => {
    const { result } = renderManagement();
    const intent = {
      ...makeReservation({ id: "123", isTentative: true, clinicId: "1" }),
      updatedAt: "2026-10-11T00:00:00.123456789Z",
    };
    act(() => result.current.handleEndTentativeRequest(intent));
    await act(async () => result.current.executeEndTentative());
    expect(clearIntentMutateAsync).toHaveBeenCalledWith({
      petId: "10",
      clinicId: "1",
      expectedReservationId: "123",
      expectedUpdatedAt: "2026-10-11T00:00:00.123456789Z",
    });
  });
  it("does not clear when the read response lacks a version", async () => {
    const { result } = renderManagement();
    act(() =>
      result.current.handleEndTentativeRequest(
        makeReservation({ id: "123", isTentative: true, updatedAt: undefined }),
      ),
    );
    await act(async () => result.current.executeEndTentative());
    expect(clearIntentMutateAsync).not.toHaveBeenCalled();
  });
});
