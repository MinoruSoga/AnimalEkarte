import { describe, expect, it } from "vitest";

import type { Reservation } from "@/hooks/reservation-intent-response";

import { selectReusableGeneralAppointment } from "./use-medical-record-form-helpers";

/**
 * EMR-308: 同日の tentative（次回予定・未予約）は「再利用する実予約」ではない。
 * reusableAppointment に tentative を返すとカルテが sentinel 行に紐づき
 * status=tentative のまま残る（BackfillForMedicalRecord は status を触らない）。
 * 除外すれば record_shortcut 作成経路に回り、BE createOrPromoteTentative が
 * 同日 tentative を同一行昇格する（＝当日 tentative を promotable に扱う）。
 */

function appointmentOf(overrides: Partial<Reservation> & { id: string }): Reservation {
  return {
    id: overrides.id,
    start: new Date("2026-10-06T10:00:00+09:00"),
    end: new Date("2026-10-06T10:30:00+09:00"),
    ownerName: "山田",
    petName: "ポチ",
    petType: "犬",
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
    reservationRoute: "reception",
    clinicId: undefined,
    isTentative: false,
    intendedDate: undefined,
    hadIntendedDate: false,
    sourceMedicalRecordId: undefined,
    ...overrides,
  };
}

describe("selectReusableGeneralAppointment (EMR-308 same-day tentative)", () => {
  it("当日の tentative 行は再利用せず undefined を返す（作成経路→同一行昇格へ回す）", () => {
    const tentative = appointmentOf({
      id: "t1",
      isTentative: true,
      intendedDate: "2026-10-06",
      // sentinel: 意図日の JST 00:00 — 実時間を持たない
      start: new Date("2026-10-06T00:00:00+09:00"),
      end: new Date("2026-10-06T00:30:00+09:00"),
      status: "tentative" as Reservation["status"],
      reservationRoute: "record_intent" as Reservation["reservationRoute"],
    });

    expect(selectReusableGeneralAppointment([tentative])).toBeUndefined();
  });

  it("tentative + 実予約が混在する場合は実予約を返す", () => {
    const tentative = appointmentOf({
      id: "t1",
      isTentative: true,
      status: "tentative" as Reservation["status"],
    });
    const real = appointmentOf({ id: "r1", status: "confirmed" });

    expect(selectReusableGeneralAppointment([tentative, real])?.id).toBe("r1");
  });

  it("一般区分の未終了予約は従来どおり再利用する", () => {
    const real = appointmentOf({ id: "r1", status: "checked_in" });
    expect(selectReusableGeneralAppointment([real])?.id).toBe("r1");
  });

  it("終了/キャンセル/未来院は再利用しない", () => {
    for (const status of ["completed", "cancelled", "no_show"] as const) {
      expect(
        selectReusableGeneralAppointment([appointmentOf({ id: "x", status })]),
      ).toBeUndefined();
    }
  });

  it("trimming 区分の予約は対象外", () => {
    const trimming = appointmentOf({ id: "tr1", category: "trimming" });
    expect(selectReusableGeneralAppointment([trimming])).toBeUndefined();
  });
});
