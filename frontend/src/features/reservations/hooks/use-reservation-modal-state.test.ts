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
});
