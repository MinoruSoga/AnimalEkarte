import { renderHook } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { useReservationModalState } from "./use-reservation-modal-state";

describe("useReservationModalState", () => {
  it("受付起点の新規作成では checked_in と reception 経路を初期値にする", () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-05-29T01:07:00.000Z"));

    const { result } = renderHook(() =>
      useReservationModalState({ locationSearch: "?reception=1" }),
    );

    expect(result.current.isFormOpen).toBe(true);
    expect(result.current.editingAppointment).toEqual(
      expect.objectContaining({
        status: "checked_in",
        reservationRoute: "reception",
        source: "manual",
        visitType: "first",
      }),
    );
    expect(result.current.editingAppointment?.start?.getMinutes()).toBe(15);
    expect(result.current.editingAppointment?.end?.getTime()).toBe(
      (result.current.editingAppointment?.start?.getTime() ?? 0) + 60 * 60 * 1000,
    );

    vi.useRealTimers();
  });

  it("受付予約ボード起点(newReservation=1)では confirmed・route 未強制で初期化する", () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-05-29T01:07:00.000Z"));

    const { result } = renderHook(() =>
      useReservationModalState({ locationSearch: "?newReservation=1" }),
    );

    expect(result.current.isFormOpen).toBe(true);
    expect(result.current.editingAppointment).toEqual(
      expect.objectContaining({
        status: "confirmed",
        visitType: "first",
      }),
    );
    // 受付 walk-in と異なり reservationRoute は強制しない（モーダルで選択）
    expect(result.current.editingAppointment?.reservationRoute).toBeUndefined();
    expect(result.current.editingAppointment?.start?.getMinutes()).toBe(15);

    vi.useRealTimers();
  });

  // EMR-243: 受付ボードの日付切替。システム時刻 2026-05-29T01:07:00Z = JST 10:07、
  // quarter 丸め後の壁時計は 10:15。JST 当日は 2026-05-29。
  it("newReservation=1&date=<選択日> では選択日 + 現在丸め時刻の stub を作る", () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-05-29T01:07:00.000Z"));

    const { result } = renderHook(() =>
      useReservationModalState({ locationSearch: "?newReservation=1&date=2026-05-31" }),
    );

    expect(result.current.isFormOpen).toBe(true);
    const start = result.current.editingAppointment?.start;
    expect(start?.getFullYear()).toBe(2026);
    expect(start?.getMonth()).toBe(4);
    expect(start?.getDate()).toBe(31);
    expect(start?.getHours()).toBe(10);
    expect(start?.getMinutes()).toBe(15);
    expect(result.current.editingAppointment?.status).toBe("confirmed");

    vi.useRealTimers();
  });

  it("newReservation=1&date=不正値 は当日にフォールバックする", () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-05-29T01:07:00.000Z"));

    const { result } = renderHook(() =>
      useReservationModalState({ locationSearch: "?newReservation=1&date=not-a-date" }),
    );

    const start = result.current.editingAppointment?.start;
    expect(start?.getFullYear()).toBe(2026);
    expect(start?.getMonth()).toBe(4);
    expect(start?.getDate()).toBe(29);

    vi.useRealTimers();
  });

  it("reception=1 は date が指定されても常に当日の checked_in stub に留める", () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-05-29T01:07:00.000Z"));

    const { result } = renderHook(() =>
      useReservationModalState({ locationSearch: "?reception=1&date=2020-01-01" }),
    );

    expect(result.current.isFormOpen).toBe(true);
    expect(result.current.editingAppointment?.status).toBe("checked_in");
    const start = result.current.editingAppointment?.start;
    expect(start?.getFullYear()).toBe(2026);
    expect(start?.getMonth()).toBe(4);
    expect(start?.getDate()).toBe(29);

    vi.useRealTimers();
  });
});
