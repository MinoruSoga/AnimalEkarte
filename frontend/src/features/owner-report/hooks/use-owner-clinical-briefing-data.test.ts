import { act, cleanup, renderHook } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { selectTentativeBriefing } from "../lib/clinical-briefing";
import { useOwnerClinicalBriefingData } from "./use-owner-clinical-briefing-data";

const reservationCalls: Array<Record<string, unknown>> = [];

vi.mock("@/hooks/use-get-reservations", () => ({
  useGetReservations: vi.fn((params: Record<string, unknown>) => {
    reservationCalls.push(params);
    return {
      data: [
        {
          id: "t1",
          start: new Date("2026-06-15T00:00:00+09:00"),
          end: new Date("2026-06-15T00:15:00+09:00"),
          status: "tentative",
          isTentative: true,
          intendedDate: "2026-06-15",
        },
      ],
      isLoading: false,
      isError: false,
    };
  }),
}));
vi.mock("@/hooks/use-medical-records", () => ({
  useGetMedicalRecords: vi.fn(() => ({})),
}));
vi.mock("@/hooks/use-pet-vaccinations", () => ({
  useGetPetVaccinations: vi.fn(() => ({})),
}));
vi.mock("@/hooks/use-permission", () => ({
  usePermission: vi.fn(() => ({ canView: true })),
}));
vi.mock("./use-pet-checkup-results", () => ({
  useGetPetCheckupResults: vi.fn(() => ({})),
}));
vi.mock("../api/get-pet-examinations", () => ({
  useGetPetExaminations: vi.fn(() => ({})),
}));
vi.mock("../api/get-pet-treatment-history", () => ({
  useGetPetTreatmentHistory: vi.fn(() => ({})),
}));
vi.mock("../api/get-pet-trimming-history", () => ({
  useGetPetTrimmingHistory: vi.fn(() => ({})),
}));

function lastReservationParams() {
  return reservationCalls[reservationCalls.length - 1];
}

describe("useOwnerClinicalBriefingData (Codex PR606 JST boundary)", () => {
  beforeEach(() => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-06-15T14:59:30Z"));
    reservationCalls.length = 0;
  });

  afterEach(() => {
    cleanup();
    vi.restoreAllMocks();
    vi.useRealTimers();
  });

  it("JST 暦日を跨ぐと today と予約クエリ範囲が更新され、昨日の tentative 次回行動は落ちる", () => {
    const { result } = renderHook(() => useOwnerClinicalBriefingData("pet-1"));

    expect(result.current.today).toBe("2026-06-15");
    expect(lastReservationParams()).toMatchObject({
      startDate: "2026-06-15",
      endDate: "2028-06-15",
      petId: "pet-1",
    });
    const tentative = (result.current.reservationsQuery.data ?? []) as Array<{
      isTentative?: boolean;
      intendedDate?: string;
      start: Date;
      id: string;
      end: Date;
      status: string;
    }>;
    expect(selectTentativeBriefing(tentative, result.current.today)?.id).toBe("t1");

    act(() => {
      vi.setSystemTime(new Date("2026-06-15T15:01:00Z"));
      vi.advanceTimersByTime(61_000);
    });

    expect(result.current.today).toBe("2026-06-16");
    expect(lastReservationParams()).toMatchObject({
      startDate: "2026-06-16",
      endDate: "2028-06-16",
      petId: "pet-1",
    });
    expect(selectTentativeBriefing(tentative, result.current.today)).toBeUndefined();
  });

  it.each(["focus", "pageshow"])(
    "%s resume refreshes today and the reservation query range",
    (event) => {
      const { result } = renderHook(() => useOwnerClinicalBriefingData("pet-1"));
      expect(result.current.today).toBe("2026-06-15");

      act(() => {
        vi.setSystemTime(new Date("2026-06-20T02:00:00Z"));
        window.dispatchEvent(new Event(event));
      });

      expect(result.current.today).toBe("2026-06-20");
      expect(lastReservationParams()).toMatchObject({ startDate: "2026-06-20" });
    },
  );
});
