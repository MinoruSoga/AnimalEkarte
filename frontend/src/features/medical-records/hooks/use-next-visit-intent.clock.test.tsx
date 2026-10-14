import { act, cleanup, render, renderHook, screen } from "@testing-library/react";
import "@testing-library/jest-dom/vitest";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { useGetReservations } from "@/hooks/use-get-reservations";
import { usePermission } from "@/hooks/use-permission";
import { toJSTWallDate } from "@/lib/jst-date";
import type { Reservation } from "@/lib/transforms/reservation";

import { NextVisitDateField } from "../components/NextVisitDateField";
import { useNextVisitIntent } from "./use-next-visit-intent";

vi.mock("sonner", () => ({ toast: { success: vi.fn(), error: vi.fn() } }));
vi.mock("@/hooks/use-get-reservations", () => ({ useGetReservations: vi.fn() }));
vi.mock("@/hooks/use-permission", () => ({ usePermission: vi.fn() }));

function appointment(instant: string, status: Reservation["status"] = "confirmed"): Reservation {
  return {
    id: instant,
    status,
    start: toJSTWallDate(new Date(instant)),
    end: toJSTWallDate(new Date(instant)),
    ownerName: "飼主",
    petName: "患者",
    petType: undefined,
    visitType: "revisit",
    type: "診療",
    category: "general",
    reservationTypeId: undefined,
    doctor: "担当医",
    doctorId: undefined,
    isDesignated: false,
    notes: undefined,
    petId: "10",
    ownerId: "3",
    source: "manual",
    reservationRoute: null,
    clinicId: "2",
  };
}
function tentative(date = "2026-10-12"): Reservation & { intendedDate: string } {
  return {
    ...appointment(`${date}T00:00:00+09:00`),
    id: "intent",
    // The runtime sentinel intentionally remains outside the generated status union.
    status: "tentative" as Reservation["status"],
    intendedDate: date,
  };
}
function mockReservations(items: Reservation[]) {
  vi.mocked(useGetReservations).mockReturnValue({ data: items } as ReturnType<
    typeof useGetReservations
  >);
}
const setIntentMutation = { mutateAsync: vi.fn().mockResolvedValue({}) };
const clearIntentMutation = { mutateAsync: vi.fn().mockResolvedValue(undefined) };
function useIntent() {
  return useNextVisitIntent({
    petId: "10",
    clinicId: "2",
    isPetDeceased: false,
    setIntentMutation,
    clearIntentMutation,
  });
}

beforeEach(() => {
  vi.useFakeTimers();
  vi.setSystemTime(new Date("2026-10-11T01:59:30Z"));
  vi.clearAllMocks();
  mockReservations([]);
  vi.mocked(usePermission).mockReturnValue({
    canView: true,
    canCreate: true,
    canEdit: true,
    canDelete: false,
  });
});
afterEach(() => {
  cleanup();
  vi.restoreAllMocks();
  vi.useRealTimers();
});

describe("next visit reactive clock", () => {
  it("switches from an appointment to tentative confirmation at its exact start", () => {
    mockReservations([appointment("2026-10-11T02:00:00Z"), tentative()]);
    const { result } = renderHook(useIntent);
    expect(result.current.nextVisitLabel).toBe("2026-10-11 11:00");
    expect(result.current.canConfirmTentative).toBe(false);
    act(() => vi.advanceTimersByTime(29_999));
    expect(result.current.canConfirmTentative).toBe(false);
    act(() => vi.advanceTimersByTime(1));
    expect(result.current.nextVisitLabel).toBe("2026-10-12");
    expect(result.current.canConfirmTentative).toBe(true);
  });
  it("rearms for each future pending or confirmed appointment and finally becomes unset", () => {
    mockReservations([
      appointment("2026-10-11T02:00:00Z", "pending"),
      appointment("2026-10-11T03:00:00Z"),
    ]);
    const { result } = renderHook(useIntent);
    act(() => vi.advanceTimersByTime(30_000));
    expect(result.current.nextVisitLabel).toBe("2026-10-11 12:00");
    act(() => vi.advanceTimersByTime(3_600_000));
    expect(result.current.nextVisitLabel).toBe("");
    expect(result.current.canConfirmTentative).toBe(false);
    expect(vi.getTimerCount()).toBe(1);
  });
  it("updates both query dates at midnight even without reservations", () => {
    vi.setSystemTime(new Date("2026-10-11T14:59:30Z"));
    renderHook(useIntent);
    expect(useGetReservations).toHaveBeenLastCalledWith(
      expect.objectContaining({ startDate: "2026-10-11", endDate: "2028-10-11" }),
    );
    act(() => vi.advanceTimersByTime(30_000));
    expect(useGetReservations).toHaveBeenLastCalledWith(
      expect.objectContaining({ startDate: "2026-10-12", endDate: "2028-10-12" }),
    );
  });
  it("revalidates an unchanged same-day intent when it becomes a past date", () => {
    vi.setSystemTime(new Date("2026-10-11T14:59:30Z"));
    mockReservations([tentative("2026-10-11")]);
    function IntentField() {
      const intent = useIntent();
      return (
        <NextVisitDateField
          value={intent.intentFieldValue}
          onChange={intent.handleIntentDateChange}
        />
      );
    }
    render(<IntentField />);
    expect(screen.getByLabelText("次回来院推奨日")).not.toHaveAttribute("aria-invalid");
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
    act(() => vi.advanceTimersByTime(30_000));
    expect(screen.getByLabelText("次回来院推奨日")).toBeInvalid();
    expect(screen.getByLabelText("次回来院推奨日")).toHaveAttribute("aria-invalid", "true");
    expect(screen.getByRole("alert")).toHaveTextContent("今日より前");
  });
  it.each(["focus", "pageshow"])("catches up an expired appointment after sleep on %s", (event) => {
    mockReservations([appointment("2026-10-11T02:00:00Z"), tentative()]);
    const { result } = renderHook(useIntent);
    act(() => {
      vi.setSystemTime(new Date("2026-10-12T02:00:00Z"));
      window.dispatchEvent(new Event(event));
    });
    expect(result.current.canConfirmTentative).toBe(true);
    expect(useGetReservations).toHaveBeenLastCalledWith(
      expect.objectContaining({ startDate: "2026-10-12" }),
    );
  });
  it("pauses in the background and catches up at visibility restoration", () => {
    const visibility = vi.spyOn(document, "visibilityState", "get").mockReturnValue("visible");
    mockReservations([appointment("2026-10-11T02:00:00Z"), tentative()]);
    const { result } = renderHook(useIntent);
    act(() => {
      visibility.mockReturnValue("hidden");
      document.dispatchEvent(new Event("visibilitychange"));
    });
    expect(vi.getTimerCount()).toBe(0);
    act(() => {
      vi.setSystemTime(new Date("2026-10-11T02:01:00Z"));
      visibility.mockReturnValue("visible");
      document.dispatchEvent(new Event("visibilitychange"));
    });
    expect(result.current.canConfirmTentative).toBe(true);
    expect(vi.getTimerCount()).toBe(1);
  });
  it("removes the clock timer and listeners on unmount", () => {
    mockReservations([appointment("2026-10-11T02:00:00Z")]);
    const { unmount } = renderHook(useIntent);
    expect(vi.getTimerCount()).toBe(1);
    unmount();
    expect(vi.getTimerCount()).toBe(0);
    act(() => window.dispatchEvent(new Event("focus")));
    expect(vi.getTimerCount()).toBe(0);
  });
});
