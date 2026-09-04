import { describe, it, expect } from "vitest";
import type { Reservation } from "../types";
import { filterCalendarAppointments } from "./ReservationManagement";

function makeReservation(
  overrides: { id: string } & Partial<Omit<Reservation, "id">>,
): Reservation {
  return {
    id: overrides.id,
    start: new Date("2026-06-15T09:00:00"),
    end: new Date("2026-06-15T10:00:00"),
    ownerName: "テスト飼主",
    petName: "テストペット",
    petType: undefined,
    visitType: "revisit",
    type: "一般診察",
    category: "general",
    reservationTypeId: undefined,
    doctor: "",
    doctorId: undefined,
    isDesignated: false,
    status: "confirmed",
    notes: undefined,
    petId: undefined,
    ownerId: undefined,
    source: "manual",
    reservationRoute: null,
    clinicId: undefined,
    ...overrides,
  };
}

describe("filterCalendarAppointments (#116)", () => {
  it("cancelled ステータスの予約を除外する", () => {
    const result = filterCalendarAppointments([
      makeReservation({ id: "1", status: "confirmed" }),
      makeReservation({ id: "2", status: "cancelled" }),
      makeReservation({ id: "3", status: "pending" }),
    ]);
    expect(result).toHaveLength(2);
    expect(result.find((a) => a.id === "2")).toBeUndefined();
  });

  it("LINE経由キャンセルも除外する（source に依存しない）", () => {
    const result = filterCalendarAppointments([
      makeReservation({ id: "1", status: "confirmed", source: "manual" }),
      makeReservation({ id: "2", status: "cancelled", source: "line" }),
      makeReservation({ id: "3", status: "cancelled", source: "manual" }),
    ]);
    expect(result).toHaveLength(1);
    expect(result[0].id).toBe("1");
  });

  it("no_show は除外しない（スコープ外）", () => {
    const result = filterCalendarAppointments([
      makeReservation({ id: "1", status: "no_show" }),
      makeReservation({ id: "2", status: "cancelled" }),
    ]);
    expect(result).toHaveLength(1);
    expect(result[0].status).toBe("no_show");
  });

  it("cancelled 以外の全ステータスを維持する", () => {
    const statuses = [
      "confirmed",
      "pending",
      "checked_in",
      "in_consultation",
      "accounting",
      "completed",
      "no_show",
    ];
    const appointments = statuses.map((status, i) => makeReservation({ id: String(i), status }));
    const result = filterCalendarAppointments(appointments);
    expect(result).toHaveLength(statuses.length);
  });

  it("空配列は空配列を返す", () => {
    expect(filterCalendarAppointments([])).toHaveLength(0);
  });
});
