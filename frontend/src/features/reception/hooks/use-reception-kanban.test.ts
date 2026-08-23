import { act, renderHook, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { PetStatusDeceased } from "@/types/generated/models";

import { useReceptionKanban } from "./use-reception-kanban";
import type { ReceptionAppointment, ReceptionColumn } from "../api/types";

/**
 * use-reception-kanban の楽観更新ロジックを保護するテスト。
 *
 * このフックは医療記録のステータス変更 API + 楽観的 UI 更新を担う最重要かつ最複雑な
 * 部位だが従来テストが存在しなかった。リファクタの動作保存を保証する安全網として、
 * 4 つの mutator (moveCard / advanceStatus / cancelAppointment / updateAppointment) の
 * 外部から観測可能な振る舞い (columns の遷移 + mutateAsync 呼び出し) を固定する。
 */

// 本番の react-query は安定参照を返す。テストでも data/Map を毎レンダー新規生成すると
// staffs → staffMap → apiColumnData の identity が変わり派生 state 同期が無限ループする。
// vi.mock factory は import より上に巻き上げられるため、参照する定数も vi.hoisted で固定する。
const {
  mutateAsyncMock,
  toastErrorMock,
  toastSuccessMock,
  handleApiErrorMock,
  stableEmptyStaffs,
  stableEmptyStaffMap,
} = vi.hoisted(() => ({
  mutateAsyncMock: vi.fn(),
  toastErrorMock: vi.fn(),
  toastSuccessMock: vi.fn(),
  handleApiErrorMock: vi.fn(),
  stableEmptyStaffs: [] as never[],
  stableEmptyStaffMap: new Map<string, string>(),
}));

vi.mock("sonner", () => ({
  toast: { error: toastErrorMock, success: toastSuccessMock },
}));

vi.mock("@/lib/handle-api-error", () => ({
  handleApiError: handleApiErrorMock,
}));

// useGetReception の返却を test ごとに差し替えられるよう mutable holder で保持する。
let apiColumnsHolder: ReceptionColumn[] | undefined;

vi.mock("../api/get-reception", () => ({
  todayISO: () => "2026-06-01",
  useGetReception: () => ({ data: apiColumnsHolder, isLoading: false, isError: false }),
}));

vi.mock("../api/get-staffs", () => ({
  useGetStaffs: () => ({ data: stableEmptyStaffs }),
  buildStaffMap: () => stableEmptyStaffMap,
}));

vi.mock("../api/update-appointment-status", () => ({
  useUpdateAppointmentStatus: () => ({ mutateAsync: mutateAsyncMock }),
}));

function makeAppointment(overrides: Partial<ReceptionAppointment> & { id: string }): ReceptionAppointment {
  return {
    time: "09:00",
    visitDate: "2026-06-01",
    end: new Date(2026, 5, 1, 9, 30, 0),
    ownerName: "山田",
    petType: "犬",
    petName: "ポチ",
    visitType: "再診",
    reservationType: "一般診察",
    reservationTypeId: "1",
    reservationCategory: "general",
    isDesignated: false,
    doctor: undefined,
    doctorId: "",
    petId: "10",
    ownerId: "20",
    status: "checked_in",
    notes: undefined,
    source: "manual",
    ...overrides,
  };
}

/** id → 所属カラム title を引く (テスト assertion 用) */
function columnOf(columns: { title: string; appointments: ReceptionAppointment[] }[], id: string): string | undefined {
  return columns.find((c) => c.appointments.some((a) => a.id === id))?.title;
}

function setApiColumns(appointments: ReceptionAppointment[]) {
  const byColumn: Record<string, ReceptionColumn> = {
    受付予約: { id: "pending", title: "受付予約", appointments: [] },
    受付済: { id: "checked_in", title: "受付済", appointments: [] },
    診療中: { id: "in_consultation", title: "診療中", appointments: [] },
    会計待ち: { id: "accounting", title: "会計待ち", appointments: [] },
    会計済: { id: "completed", title: "会計済", appointments: [] },
  };
  const statusToTitle: Record<string, string> = {
    pending: "受付予約",
    confirmed: "受付予約",
    checked_in: "受付済",
    in_consultation: "診療中",
    accounting: "会計待ち",
    completed: "会計済",
  };
  for (const appt of appointments) {
    byColumn[statusToTitle[appt.status] ?? "受付予約"].appointments.push(appt);
  }
  apiColumnsHolder = Object.values(byColumn);
}

function createDeferred<T>() {
  let resolve!: (value: T | PromiseLike<T>) => void;
  let reject!: (reason?: unknown) => void;
  const promise = new Promise<T>((resolvePromise, rejectPromise) => {
    resolve = resolvePromise;
    reject = rejectPromise;
  });
  return { promise, reject, resolve };
}

interface PermissionProps {
  canEditReservation?: boolean;
  canDeleteReservation?: boolean;
}

async function renderKanban(initialPermissions: PermissionProps = {
  canEditReservation: true,
  canDeleteReservation: true,
}) {
  const view = renderHook(
    (permissions: PermissionProps) => useReceptionKanban(permissions),
    { initialProps: initialPermissions },
  );
  // 派生 state の inline 同期 + ref 更新 effect が落ち着くまで待つ
  await waitFor(() => expect(view.result.current.columns.length).toBe(5));
  return view;
}

describe("useReceptionKanban", () => {
  beforeEach(() => {
    mutateAsyncMock.mockReset();
    mutateAsyncMock.mockResolvedValue(undefined);
    toastErrorMock.mockReset();
    toastSuccessMock.mockReset();
    handleApiErrorMock.mockReset();
    apiColumnsHolder = undefined;
  });

  it("API データから 5 カラムを構築し、各 appointment を所属カラムに配置する", async () => {
    setApiColumns([
      makeAppointment({ id: "a1", status: "pending", visitType: "初診" }),
      makeAppointment({ id: "a2", status: "checked_in", visitType: "再診" }),
      makeAppointment({ id: "a3", status: "completed", visitType: "初診" }),
    ]);
    const { result } = await renderKanban();

    expect(result.current.columns.map((c) => c.title)).toEqual([
      "受付予約", "受付済", "診療中", "会計待ち", "会計済",
    ]);
    expect(columnOf(result.current.columns, "a1")).toBe("受付予約");
    expect(columnOf(result.current.columns, "a2")).toBe("受付済");
    expect(columnOf(result.current.columns, "a3")).toBe("会計済");
  });

  describe("moveCard", () => {
    it("カラムをまたぐ移動で楽観的にカードを移し、対応ステータスで API を呼ぶ", async () => {
      setApiColumns([makeAppointment({ id: "a1", status: "pending" })]);
      const { result } = await renderKanban();

      await act(async () => {
        result.current.moveCard(0, "受付予約", "受付済", "a1");
      });

      expect(columnOf(result.current.columns, "a1")).toBe("受付済");
      expect(mutateAsyncMock).toHaveBeenCalledWith({ id: "a1", status: "checked_in" });
    });

    it("受付済 → 診療中 の直接ドラッグはブロックし、API を呼ばず toast.error を出す", async () => {
      setApiColumns([makeAppointment({ id: "a2", status: "checked_in" })]);
      const { result } = await renderKanban();

      let returned: unknown;
      await act(async () => {
        returned = result.current.moveCard(0, "受付済", "診療中", "a2");
      });

      expect(returned).toBe(false);
      expect(columnOf(result.current.columns, "a2")).toBe("受付済");
      expect(mutateAsyncMock).not.toHaveBeenCalled();
      expect(toastErrorMock).toHaveBeenCalled();
    });

    it("同一カラム内の並べ替えでは API を呼ばない", async () => {
      setApiColumns([
        makeAppointment({ id: "a1", status: "pending", time: "09:00" }),
        makeAppointment({ id: "a2", status: "pending", time: "10:00" }),
      ]);
      const { result } = await renderKanban();

      await act(async () => {
        result.current.moveCard(0, "受付予約", "受付予約", "a2");
      });

      expect(mutateAsyncMock).not.toHaveBeenCalled();
      const pending = result.current.columns.find((c) => c.title === "受付予約");
      expect(pending?.appointments.map((a) => a.id)).toEqual(["a2", "a1"]);
    });

    it("クロスカラム移動で hoverIndex の参照カード位置に挿入する", async () => {
      setApiColumns([
        makeAppointment({ id: "a1", status: "pending" }),
        makeAppointment({ id: "b1", status: "checked_in", time: "09:00" }),
        makeAppointment({ id: "b2", status: "checked_in", time: "10:00" }),
      ]);
      const { result } = await renderKanban();

      // a1 を 受付済 の filtered index=1 (= b2 の手前) へ移動
      await act(async () => {
        result.current.moveCard(1, "受付予約", "受付済", "a1");
      });

      const checkedIn = result.current.columns.find((c) => c.title === "受付済");
      expect(checkedIn?.appointments.map((a) => a.id)).toEqual(["b1", "a1", "b2"]);
      expect(mutateAsyncMock).toHaveBeenCalledWith({ id: "a1", status: "checked_in" });
    });

    it("drag 失敗時は対象カードだけを操作前へ戻し、別カードの成功した移動を維持する", async () => {
      const failedMutation = createDeferred<void>();
      mutateAsyncMock
        .mockImplementationOnce(() => failedMutation.promise)
        .mockResolvedValueOnce(undefined);
      setApiColumns([
        makeAppointment({ id: "a1", status: "pending" }),
        makeAppointment({ id: "b1", status: "pending" }),
      ]);
      const { result } = await renderKanban();

      await act(async () => {
        result.current.moveCard(0, "受付予約", "受付済", "a1");
        result.current.moveCard(0, "受付予約", "受付済", "b1");
      });

      expect(columnOf(result.current.columns, "a1")).toBe("受付済");
      expect(columnOf(result.current.columns, "b1")).toBe("受付済");

      await act(async () => {
        failedMutation.reject(new Error("boom"));
        await failedMutation.promise.catch(() => undefined);
      });

      await waitFor(() => {
        expect(columnOf(result.current.columns, "a1")).toBe("受付予約");
      });
      expect(columnOf(result.current.columns, "b1")).toBe("受付済");
      expect(handleApiErrorMock).toHaveBeenCalled();
    });

    it("対象カードが source カラムに無ければ false を返し API を呼ばない", async () => {
      setApiColumns([makeAppointment({ id: "a1", status: "pending" })]);
      const { result } = await renderKanban();

      let returned: unknown;
      await act(async () => {
        returned = result.current.moveCard(0, "受付予約", "受付済", "ghost");
      });

      expect(returned).toBe(false);
      expect(mutateAsyncMock).not.toHaveBeenCalled();
    });

    it("ブロックされた移動の toast は 3 秒間スロットルされ連続では 1 回のみ", async () => {
      setApiColumns([makeAppointment({ id: "a2", status: "checked_in" })]);
      const { result } = await renderKanban();

      await act(async () => {
        result.current.moveCard(0, "受付済", "診療中", "a2");
        result.current.moveCard(0, "受付済", "診療中", "a2");
      });

      expect(toastErrorMock).toHaveBeenCalledTimes(1);
    });

    it("死亡 appointment の drag は拒否し、カラムも API も変更しない", async () => {
      setApiColumns([
        makeAppointment({
          id: "deceased",
          status: "pending",
          petStatus: PetStatusDeceased,
        }),
      ]);
      const { result } = await renderKanban();

      let returned: unknown;
      await act(async () => {
        returned = result.current.moveCard(0, "受付予約", "受付済", "deceased");
      });

      expect(returned).toBe(false);
      expect(columnOf(result.current.columns, "deceased")).toBe("受付予約");
      expect(mutateAsyncMock).not.toHaveBeenCalled();
    });

    it("edit 権限が true でなければ drag を拒否する", async () => {
      setApiColumns([makeAppointment({ id: "a1", status: "pending" })]);
      const { result } = await renderKanban({
        canEditReservation: undefined,
        canDeleteReservation: true,
      });

      let returned: unknown;
      await act(async () => {
        returned = result.current.moveCard(0, "受付予約", "受付済", "a1");
      });

      expect(returned).toBe(false);
      expect(columnOf(result.current.columns, "a1")).toBe("受付予約");
      expect(mutateAsyncMock).not.toHaveBeenCalled();
    });
  });

  describe("advanceStatus", () => {
    it("受付済 → 診療中 へ進め、in_consultation で API を呼ぶ", async () => {
      setApiColumns([makeAppointment({ id: "a2", status: "checked_in" })]);
      const { result } = await renderKanban();

      await act(async () => {
        result.current.advanceStatus(makeAppointment({ id: "a2", status: "checked_in" }));
      });

      expect(columnOf(result.current.columns, "a2")).toBe("診療中");
      expect(mutateAsyncMock).toHaveBeenCalledWith({ id: "a2", status: "in_consultation" });
    });

    it("会計済 のカードは completed で API を呼び、ローカルから除外する", async () => {
      setApiColumns([makeAppointment({ id: "a3", status: "completed" })]);
      const { result } = await renderKanban();

      await act(async () => {
        result.current.advanceStatus(makeAppointment({ id: "a3", status: "completed" }));
      });

      expect(columnOf(result.current.columns, "a3")).toBeUndefined();
      expect(mutateAsyncMock).toHaveBeenCalledWith({ id: "a3", status: "completed" });
    });

    it("advance 失敗時は対象カードだけを操作前へ戻し、別カードの成功した遷移を維持する", async () => {
      const failedMutation = createDeferred<void>();
      mutateAsyncMock
        .mockImplementationOnce(() => failedMutation.promise)
        .mockResolvedValueOnce(undefined);
      const first = makeAppointment({ id: "a1", status: "checked_in" });
      const second = makeAppointment({ id: "b1", status: "checked_in" });
      setApiColumns([first, second]);
      const { result } = await renderKanban();

      await act(async () => {
        result.current.advanceStatus(first);
        result.current.advanceStatus(second);
      });

      expect(columnOf(result.current.columns, "a1")).toBe("診療中");
      expect(columnOf(result.current.columns, "b1")).toBe("診療中");

      await act(async () => {
        failedMutation.reject(new Error("boom"));
        await failedMutation.promise.catch(() => undefined);
      });

      await waitFor(() => {
        expect(columnOf(result.current.columns, "a1")).toBe("受付済");
      });
      expect(columnOf(result.current.columns, "b1")).toBe("診療中");
    });

    it("advance の成功 toast は API 成功後にだけ表示する", async () => {
      const pendingMutation = createDeferred<void>();
      mutateAsyncMock.mockImplementationOnce(() => pendingMutation.promise);
      const appointment = makeAppointment({ id: "a2", status: "checked_in" });
      setApiColumns([appointment]);
      const { result } = await renderKanban();

      act(() => {
        result.current.advanceStatus(appointment);
      });

      expect(columnOf(result.current.columns, "a2")).toBe("診療中");
      expect(toastSuccessMock).not.toHaveBeenCalled();

      await act(async () => {
        pendingMutation.resolve(undefined);
        await pendingMutation.promise;
      });

      await waitFor(() => expect(toastSuccessMock).toHaveBeenCalledTimes(1));
    });

    it("terminal 除外失敗時は対象カードだけを復元し、別カードの成功した遷移を維持する", async () => {
      const failedMutation = createDeferred<void>();
      mutateAsyncMock
        .mockImplementationOnce(() => failedMutation.promise)
        .mockResolvedValueOnce(undefined);
      const terminal = makeAppointment({ id: "a3", status: "completed" });
      const other = makeAppointment({ id: "b1", status: "checked_in" });
      setApiColumns([terminal, other]);
      const { result } = await renderKanban();

      await act(async () => {
        result.current.advanceStatus(terminal);
        result.current.advanceStatus(other);
      });

      expect(columnOf(result.current.columns, "a3")).toBeUndefined();
      expect(columnOf(result.current.columns, "b1")).toBe("診療中");

      await act(async () => {
        failedMutation.reject(new Error("boom"));
        await failedMutation.promise.catch(() => undefined);
      });

      await waitFor(() => {
        expect(columnOf(result.current.columns, "a3")).toBe("会計済");
      });
      expect(columnOf(result.current.columns, "b1")).toBe("診療中");
    });

    it("死亡 appointment の status 変更は拒否する", async () => {
      const deceasedAppointment = makeAppointment({
        id: "deceased",
        status: "checked_in",
        petStatus: PetStatusDeceased,
      });
      setApiColumns([deceasedAppointment]);
      const { result } = await renderKanban();

      await act(async () => {
        result.current.advanceStatus(deceasedAppointment);
      });

      expect(columnOf(result.current.columns, "deceased")).toBe("受付済");
      expect(mutateAsyncMock).not.toHaveBeenCalled();
    });

    it("captured status callback は commit 後の最新 edit 権限を再確認する", async () => {
      const appointment = makeAppointment({ id: "a2", status: "checked_in" });
      setApiColumns([appointment]);
      const { result, rerender } = await renderKanban();
      const capturedAdvanceStatus = result.current.advanceStatus;

      rerender({ canEditReservation: false, canDeleteReservation: true });
      await act(async () => {
        capturedAdvanceStatus(appointment);
      });

      expect(columnOf(result.current.columns, "a2")).toBe("受付済");
      expect(mutateAsyncMock).not.toHaveBeenCalled();
    });
  });

  describe("cancelAppointment", () => {
    it("対象カードをローカルから除外し、cancelled で API を呼ぶ", async () => {
      setApiColumns([makeAppointment({ id: "a2", status: "checked_in" })]);
      const { result } = await renderKanban();

      await act(async () => {
        result.current.cancelAppointment("a2");
      });

      expect(columnOf(result.current.columns, "a2")).toBeUndefined();
      expect(mutateAsyncMock).toHaveBeenCalledWith({ id: "a2", status: "cancelled" });
    });

    it("cancel 失敗時は対象カードだけを復元し、別カードの成功した遷移を維持する", async () => {
      const failedMutation = createDeferred<void>();
      mutateAsyncMock
        .mockImplementationOnce(() => failedMutation.promise)
        .mockResolvedValueOnce(undefined);
      const cancelled = makeAppointment({ id: "a1", status: "checked_in" });
      const other = makeAppointment({ id: "b1", status: "checked_in" });
      setApiColumns([cancelled, other]);
      const { result } = await renderKanban();

      await act(async () => {
        void result.current.cancelAppointment(cancelled.id);
        result.current.advanceStatus(other);
      });

      expect(columnOf(result.current.columns, "a1")).toBeUndefined();
      expect(columnOf(result.current.columns, "b1")).toBe("診療中");

      await act(async () => {
        failedMutation.reject(new Error("boom"));
        await failedMutation.promise.catch(() => undefined);
      });

      await waitFor(() => {
        expect(columnOf(result.current.columns, "a1")).toBe("受付済");
      });
      expect(columnOf(result.current.columns, "b1")).toBe("診療中");
    });

    it("死亡 appointment の cancel は拒否する", async () => {
      setApiColumns([
        makeAppointment({
          id: "deceased",
          status: "checked_in",
          petStatus: PetStatusDeceased,
        }),
      ]);
      const { result } = await renderKanban();

      await act(async () => {
        result.current.cancelAppointment("deceased");
      });

      expect(columnOf(result.current.columns, "deceased")).toBe("受付済");
      expect(mutateAsyncMock).not.toHaveBeenCalled();
    });

    it("captured cancel callback は commit 後の最新 delete 権限を再確認する", async () => {
      setApiColumns([makeAppointment({ id: "a2", status: "checked_in" })]);
      const { result, rerender } = await renderKanban();
      const capturedCancelAppointment = result.current.cancelAppointment;

      rerender({ canEditReservation: true, canDeleteReservation: false });
      await act(async () => {
        capturedCancelAppointment("a2");
      });

      expect(columnOf(result.current.columns, "a2")).toBe("受付済");
      expect(mutateAsyncMock).not.toHaveBeenCalled();
    });
  });

  describe("updateAppointment", () => {
    it("該当カードのフィールドをマージし、API は呼ばない", async () => {
      setApiColumns([makeAppointment({ id: "a2", status: "checked_in", petName: "ポチ" })]);
      const { result } = await renderKanban();

      await act(async () => {
        result.current.updateAppointment(makeAppointment({ id: "a2", status: "checked_in", petName: "タマ" }));
      });

      const card = result.current.columns
        .flatMap((c) => c.appointments)
        .find((a) => a.id === "a2");
      expect(card?.petName).toBe("タマ");
      expect(mutateAsyncMock).not.toHaveBeenCalled();
    });

    it("edit 権限が true でなければローカル appointment も更新しない", async () => {
      setApiColumns([makeAppointment({ id: "a2", status: "checked_in", petName: "ポチ" })]);
      const { result } = await renderKanban({
        canEditReservation: false,
        canDeleteReservation: true,
      });

      act(() => {
        result.current.updateAppointment(
          makeAppointment({ id: "a2", status: "checked_in", petName: "タマ" }),
        );
      });

      const card = result.current.columns
        .flatMap((column) => column.appointments)
        .find((appointment) => appointment.id === "a2");
      expect(card?.petName).toBe("ポチ");
    });
  });

  describe("filters", () => {
    it("visitType フィルタは選択外の診察区分を filteredColumns から除外する", async () => {
      setApiColumns([
        makeAppointment({ id: "a1", status: "checked_in", visitType: "初診" }),
        makeAppointment({ id: "a2", status: "checked_in", visitType: "再診" }),
      ]);
      const { result } = await renderKanban();

      act(() => {
        result.current.filters.toggleVisitType("初診");
      });

      expect(columnOf(result.current.filteredColumns, "a1")).toBeUndefined();
      expect(columnOf(result.current.filteredColumns, "a2")).toBe("受付済");
      // columns (生データ) は影響を受けない
      expect(columnOf(result.current.columns, "a1")).toBe("受付済");
    });

    it("doctor フィルタは指定医師かつ指名ありのカードだけを残す", async () => {
      setApiColumns([
        makeAppointment({ id: "a1", status: "checked_in", doctor: "担当者A", isDesignated: true }),
        makeAppointment({ id: "a2", status: "checked_in", doctor: "担当者B", isDesignated: true }),
        makeAppointment({ id: "a3", status: "checked_in", doctor: "担当者A", isDesignated: false }),
      ]);
      const { result } = await renderKanban();

      act(() => {
        result.current.filters.setSelectedDoctor("担当者A");
      });

      expect(columnOf(result.current.filteredColumns, "a1")).toBe("受付済");
      expect(columnOf(result.current.filteredColumns, "a2")).toBeUndefined(); // 別医師
      expect(columnOf(result.current.filteredColumns, "a3")).toBeUndefined(); // 指名なし
    });

    it("「医師指名なし」フィルタは isDesignated のカードを除外する", async () => {
      setApiColumns([
        makeAppointment({ id: "a1", status: "checked_in", doctor: "担当者A", isDesignated: true }),
        makeAppointment({ id: "a3", status: "checked_in", isDesignated: false }),
      ]);
      const { result } = await renderKanban();

      act(() => {
        result.current.filters.setSelectedDoctor("医師指名なし");
      });

      expect(columnOf(result.current.filteredColumns, "a1")).toBeUndefined();
      expect(columnOf(result.current.filteredColumns, "a3")).toBe("受付済");
    });

    it("trimming フィルタは reservationCategory!=trimming を除外する", async () => {
      setApiColumns([
        makeAppointment({ id: "a1", status: "checked_in", reservationCategory: "general" }),
        makeAppointment({ id: "a2", status: "checked_in", reservationCategory: "trimming" }),
      ]);
      const { result } = await renderKanban();

      act(() => {
        result.current.filters.setIsTrimmingOnly(true);
      });

      expect(columnOf(result.current.filteredColumns, "a1")).toBeUndefined();
      expect(columnOf(result.current.filteredColumns, "a2")).toBe("受付済");
    });
  });
});
