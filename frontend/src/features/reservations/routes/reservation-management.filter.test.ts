import { describe, it, expect } from "vitest";
import { formatJSTWallDate } from "@/lib/jst-date";
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
    isTentative: false,
    intendedDate: undefined,
    hadIntendedDate: false,
    sourceMedicalRecordId: undefined,
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

// EMR-308: tentative（次回予定・未予約）はグリッドに「次回予定・未予約」付きで出すが、
// 意図日が過ぎた行は ExpiredTentativePanel で処置するためカレンダーからは外す。
describe("filterCalendarAppointments — tentative (EMR-308)", () => {
  // 今日の JST 暦日。テストでは実日付に依存させず意図日だけを動かす。
  const today = formatJSTWallDate(new Date());

  it("当日・未来の tentative はグリッドに残す", () => {
    const result = filterCalendarAppointments([
      makeReservation({ id: "t-today", isTentative: true, intendedDate: today }),
      makeReservation({ id: "t-future", isTentative: true, intendedDate: "2999-12-31" }),
      makeReservation({ id: "normal", status: "confirmed" }),
    ]);
    expect(result.map((a) => a.id).sort()).toEqual(["normal", "t-future", "t-today"]);
  });

  it("意図日が昨日以前の tentative はカレンダーから除外する（期限超過パネル側で扱う）", () => {
    const result = filterCalendarAppointments([
      makeReservation({ id: "t-expired", isTentative: true, intendedDate: "2000-01-01" }),
      makeReservation({ id: "t-today", isTentative: true, intendedDate: today }),
    ]);
    expect(result.map((a) => a.id)).toEqual(["t-today"]);
  });

  it("intendedDate 欠落の tentative は sentinel start の暦日で期限判定する", () => {
    const result = filterCalendarAppointments([
      makeReservation({
        id: "t-sentinel-past",
        isTentative: true,
        start: new Date("2000-01-01T00:00:00+09:00"),
        end: new Date("2000-01-01T00:30:00+09:00"),
      }),
    ]);
    expect(result).toHaveLength(0);
  });

  it("hadIntendedDate だけの非 tentative 行（昇格元 provenance）は通常ルールで扱う", () => {
    const result = filterCalendarAppointments([
      makeReservation({ id: "promoted", status: "confirmed", hadIntendedDate: true }),
    ]);
    expect(result.map((a) => a.id)).toEqual(["promoted"]);
  });
});
