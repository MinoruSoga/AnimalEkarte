import { act, renderHook, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { formatJSTWallDate } from "@/lib/jst-date";
import { useGetReservations } from "@/hooks/use-get-reservations";
import { usePermission } from "@/hooks/use-permission";
import {
  transformReservationWithIntent,
  type Reservation,
} from "@/hooks/reservation-intent-response";

import { useNextVisitIntent } from "./use-next-visit-intent";

vi.mock("sonner", () => ({ toast: { success: vi.fn(), error: vi.fn() } }));
vi.mock("@/hooks/use-get-reservations", () => ({ useGetReservations: vi.fn() }));
vi.mock("@/hooks/use-permission", () => ({ usePermission: vi.fn() }));

/**
 * R3: clinic+pet キーの FIFO queue が PUT レスポンス由来の CAS guard を保持し、
 * queued clear は「直前の自分の PUT の result guard」で DELETE する契約。
 * 先行 PUT 失敗・ガード欠落・権限剥奪・死亡では unguarded DELETE を絶対に送らない。
 */

function wallDateInDays(days: number): Date {
  const d = new Date();
  d.setDate(d.getDate() + days);
  d.setHours(0, 0, 0, 0);
  return d;
}

function validFutureDate(offsetMonths = 2): string {
  const d = new Date();
  d.setMonth(d.getMonth() + offsetMonths);
  return formatJSTWallDate(d);
}

/** PUT レスポンス（BE は tentative 行を返す）。id/updatedAt が CAS guard になる。 */
function putResult(id: string, updatedAt: string): Reservation {
  return tentativeRow(id, updatedAt);
}

function tentativeRow(id: string, updatedAt: string): Reservation {
  const start = wallDateInDays(45);
  return {
    ...transformReservationWithIntent({
      id: Number(id),
      clinic_id: 7,
      start_time: start.toISOString(),
      end_time: start.toISOString(),
      visit_type: "revisit",
      reservation_type_id: 1,
      is_designated: false,
      source: "manual",
      is_staff_delegated: false,
      customer_fields: null,
      status: "tentative" as Reservation["status"],
      notes: "",
      created_at: updatedAt,
      updated_at: updatedAt,
    }),
    id,
    start,
    end: start,
  };
}

function mockReservations(items: Reservation[]) {
  vi.mocked(useGetReservations, { partial: true }).mockReturnValue({
    data: items,
  });
}

function permissions(over: Partial<ReturnType<typeof usePermission>> = {}) {
  vi.mocked(usePermission).mockReturnValue({
    canView: true,
    canCreate: true,
    canEdit: true,
    canDelete: false,
    ...over,
  } as ReturnType<typeof usePermission>);
}

interface Deferred<T> {
  promise: Promise<T>;
  resolve: (value: T) => void;
  reject: (err: unknown) => void;
}
function deferred<T>(): Deferred<T> {
  let resolve!: (value: T) => void;
  let reject!: (err: unknown) => void;
  const promise = new Promise<T>((res, rej) => {
    resolve = res;
    reject = rej;
  });
  return { promise, resolve, reject };
}

/** 順序制御できる set mutation。呼出ごとに deferred が返る。 */
function deferredSetMutation() {
  const calls: Array<Deferred<Reservation>> = [];
  const mutateAsync = vi.fn<SetIntentMutation["mutateAsync"]>(() => {
    const d = deferred<Reservation>();
    calls.push(d);
    return d.promise;
  });
  return { mutateAsync, calls };
}

type IntentParams = Parameters<typeof useNextVisitIntent>[0];
type SetIntentMutation = IntentParams["setIntentMutation"];
type ClearIntentMutation = IntentParams["clearIntentMutation"];

interface RenderOptions {
  petId?: string;
  clinicId?: string;
  isPetDeceased?: boolean;
  setIntentMutation?: SetIntentMutation;
  clearIntentMutation?: ClearIntentMutation;
}

function renderIntent({
  petId = "10",
  clinicId = "7",
  isPetDeceased = false,
  setIntentMutation = {
    mutateAsync: vi.fn<SetIntentMutation["mutateAsync"]>().mockResolvedValue(putResult("1", "rev")),
  },
  clearIntentMutation = {
    mutateAsync: vi.fn<ClearIntentMutation["mutateAsync"]>().mockResolvedValue(undefined),
  },
}: RenderOptions = {}) {
  return renderHook<ReturnType<typeof useNextVisitIntent>, RenderOptions>(
    (props: RenderOptions = {}) =>
      useNextVisitIntent({
        petId: props.petId ?? petId,
        clinicId: props.clinicId ?? clinicId,
        isPetDeceased: props.isPetDeceased ?? isPetDeceased,
        setIntentMutation: props.setIntentMutation ?? setIntentMutation,
        clearIntentMutation: props.clearIntentMutation ?? clearIntentMutation,
      }),
    { initialProps: { petId, clinicId, isPetDeceased, setIntentMutation, clearIntentMutation } },
  );
}

beforeEach(() => {
  vi.clearAllMocks();
  mockReservations([]);
  permissions();
});

describe("useNextVisitIntent — R3 queued PUT result guard", () => {
  it("initial empty → deferred PUT → clear: DELETE uses the PUT response guard", async () => {
    const setMutation = deferredSetMutation();
    const clearMutation = { mutateAsync: vi.fn().mockResolvedValue(undefined) };
    const { result } = renderIntent({
      setIntentMutation: setMutation,
      clearIntentMutation: clearMutation,
    });

    // 表示は初期 empty（tentative 無し）→ PUT を積み、続けて clear を要求する。
    await act(async () => result.current.handleIntentDateChange(validFutureDate()));
    await act(async () => result.current.handleIntentDateChange(""));
    expect(clearMutation.mutateAsync).not.toHaveBeenCalled();

    await act(async () => setMutation.calls[0].resolve(putResult("77", "put-rev-1")));
    await waitFor(() =>
      expect(clearMutation.mutateAsync).toHaveBeenCalledWith({
        petId: "10",
        clinicId: "7",
        expectedReservationId: "77",
        expectedUpdatedAt: "put-rev-1",
      }),
    );
    await waitFor(() => expect(result.current.intentFieldValue).toBe(""));
  });

  it("failed predecessor PUT → queued clear never issues DELETE", async () => {
    const setMutation = deferredSetMutation();
    const clearMutation = { mutateAsync: vi.fn().mockResolvedValue(undefined) };
    mockReservations([tentativeRow("42", "disp-rev")]);
    const { result } = renderIntent({
      setIntentMutation: setMutation,
      clearIntentMutation: clearMutation,
    });

    await act(async () => result.current.handleIntentDateChange(validFutureDate()));
    await act(async () => result.current.handleIntentDateChange(""));

    await act(async () => setMutation.calls[0].reject(new Error("500")));
    // 先行 PUT が失敗して guard を出さない以上、clear は削除してはいけない。
    await act(async () => Promise.resolve());
    expect(clearMutation.mutateAsync).not.toHaveBeenCalled();
  });

  it("newer PUT failure invalidates the earlier PUT guard — clear must not use it", async () => {
    const setMutation = deferredSetMutation();
    const clearMutation = { mutateAsync: vi.fn().mockResolvedValue(undefined) };
    mockReservations([tentativeRow("42", "disp-rev")]);
    const { result } = renderIntent({
      setIntentMutation: setMutation,
      clearIntentMutation: clearMutation,
    });

    await act(async () => result.current.handleIntentDateChange(validFutureDate()));
    await act(async () => result.current.handleIntentDateChange(validFutureDate(3)));
    await act(async () => result.current.handleIntentDateChange(""));

    await act(async () => setMutation.calls[0].resolve(putResult("42", "put-rev-1")));
    await waitFor(() => expect(setMutation.calls.length).toBe(2));
    await act(async () => setMutation.calls[1].reject(new Error("500")));
    await act(async () => Promise.resolve());
    // PUT#1 の token は PUT#2 失敗で stale — それを使った DELETE は不可。
    expect(clearMutation.mutateAsync).not.toHaveBeenCalled();
  });

  it("multiple queued PUTs: clear uses the latest own PUT guard (FIFO)", async () => {
    const setMutation = deferredSetMutation();
    const clearMutation = { mutateAsync: vi.fn().mockResolvedValue(undefined) };
    mockReservations([tentativeRow("42", "disp-rev")]);
    const { result } = renderIntent({
      setIntentMutation: setMutation,
      clearIntentMutation: clearMutation,
    });

    await act(async () => result.current.handleIntentDateChange(validFutureDate()));
    await act(async () => result.current.handleIntentDateChange(validFutureDate(3)));
    await act(async () => result.current.handleIntentDateChange(""));

    await act(async () => setMutation.calls[0].resolve(putResult("42", "put-rev-1")));
    await waitFor(() => expect(setMutation.calls.length).toBe(2));
    await act(async () => setMutation.calls[1].resolve(putResult("42", "put-rev-2")));
    await waitFor(() =>
      expect(clearMutation.mutateAsync).toHaveBeenCalledWith({
        petId: "10",
        clinicId: "7",
        expectedReservationId: "42",
        expectedUpdatedAt: "put-rev-2",
      }),
    );
  });

  it("no queued PUT: clear falls back to the displayed guard", async () => {
    const clearMutation = { mutateAsync: vi.fn().mockResolvedValue(undefined) };
    mockReservations([tentativeRow("42", "disp-rev")]);
    const { result } = renderIntent({ clearIntentMutation: clearMutation });

    await act(async () => result.current.handleIntentDateChange(""));
    await waitFor(() =>
      expect(clearMutation.mutateAsync).toHaveBeenCalledWith({
        petId: "10",
        clinicId: "7",
        expectedReservationId: "42",
        expectedUpdatedAt: "disp-rev",
      }),
    );
  });

  it("queued clear never adopts a newer GET/refetch guard at execution", async () => {
    const setMutation = deferredSetMutation();
    const clearMutation = { mutateAsync: vi.fn().mockResolvedValue(undefined) };
    mockReservations([tentativeRow("42", "disp-rev")]);
    const { result, rerender } = renderIntent({
      setIntentMutation: setMutation,
      clearIntentMutation: clearMutation,
    });

    await act(async () => result.current.handleIntentDateChange(validFutureDate()));
    await act(async () => result.current.handleIntentDateChange(""));

    // PUT 未完了の間に refetch が「別の新しい revision」を返しても guard として採用しない。
    mockReservations([tentativeRow("42", "newer-refetch-rev")]);
    rerender();
    await act(async () => setMutation.calls[0].resolve(putResult("42", "put-rev-1")));
    await waitFor(() =>
      expect(clearMutation.mutateAsync).toHaveBeenCalledWith({
        petId: "10",
        clinicId: "7",
        expectedReservationId: "42",
        expectedUpdatedAt: "put-rev-1",
      }),
    );
  });

  it("clinic+pet scope A→B→A: pet A の queued write は scope 復帰後も FIFO で完了する", async () => {
    const setMutation = deferredSetMutation();
    const clearMutation = { mutateAsync: vi.fn().mockResolvedValue(undefined) };
    const { result, rerender } = renderIntent({
      petId: "10",
      setIntentMutation: setMutation,
      clearIntentMutation: clearMutation,
    });

    await act(async () => result.current.handleIntentDateChange(validFutureDate()));
    await act(async () => result.current.handleIntentDateChange(""));

    // pet B へ切替 → 別 scope として draft はクリアされる。
    rerender({ petId: "20" });
    expect(result.current.intentFieldValue).toBe("");
    // pet A に戻る。A の queue は継続して動く。
    rerender({ petId: "10" });

    await act(async () => setMutation.calls[0].resolve(putResult("77", "put-rev-A")));
    await waitFor(() =>
      expect(clearMutation.mutateAsync).toHaveBeenCalledWith({
        petId: "10",
        clinicId: "7",
        expectedReservationId: "77",
        expectedUpdatedAt: "put-rev-A",
      }),
    );
  });

  it("permission revoked between enqueue and execution: queued clear is skipped", async () => {
    const setMutation = deferredSetMutation();
    const clearMutation = { mutateAsync: vi.fn().mockResolvedValue(undefined) };
    mockReservations([tentativeRow("42", "disp-rev")]);
    const { result, rerender } = renderIntent({
      setIntentMutation: setMutation,
      clearIntentMutation: clearMutation,
    });

    await act(async () => result.current.handleIntentDateChange(validFutureDate()));
    await act(async () => result.current.handleIntentDateChange(""));

    // 実行前に edit 権限が剥奪される → queued clear は発行されてはいけない。
    permissions({ canEdit: false });
    rerender();
    await act(async () => setMutation.calls[0].resolve(putResult("42", "put-rev-1")));
    await act(async () => Promise.resolve());
    expect(clearMutation.mutateAsync).not.toHaveBeenCalled();
  });

  it("pet death between enqueue and execution: queued clear is skipped", async () => {
    const setMutation = deferredSetMutation();
    const clearMutation = { mutateAsync: vi.fn().mockResolvedValue(undefined) };
    mockReservations([tentativeRow("42", "disp-rev")]);
    const { result, rerender } = renderIntent({
      setIntentMutation: setMutation,
      clearIntentMutation: clearMutation,
    });

    await act(async () => result.current.handleIntentDateChange(validFutureDate()));
    await act(async () => result.current.handleIntentDateChange(""));

    rerender({ isPetDeceased: true });
    await act(async () => setMutation.calls[0].resolve(putResult("42", "put-rev-1")));
    await act(async () => Promise.resolve());
    expect(clearMutation.mutateAsync).not.toHaveBeenCalled();
  });
});
