import { describe, it, expect } from "vitest";
import {
  transformReservationToReceptionAppointment,
  transformReservationsToReceptionColumns,
  RECEPTION_COLUMNS,
} from "./transforms";
import type { Appointment as BackendReservation } from "@/types/generated/models";

const minimal: BackendReservation = {
  id: 1,
  clinic_id: 1,
  start_time: "2026-03-25T10:00:00Z",
  end_time: "2026-03-25T10:30:00Z",
  visit_type: "first",
  reservation_type_id: 1,
  status: "confirmed",
  is_designated: false,
  pet_id: 10,
  owner_id: 20,
  created_at: "2026-03-25T00:00:00Z",
  updated_at: "2026-03-25T00:00:00Z",
};

describe("transformReservationToReceptionAppointment", () => {
  it("id を string に変換する", () => {
    expect(transformReservationToReceptionAppointment({ ...minimal, id: 7 }).id).toBe("7");
  });

  it("start_time から HH:mm 形式の time を生成する", () => {
    const result = transformReservationToReceptionAppointment({
      ...minimal,
      start_time: "2026-03-25T10:30:00Z",
    });
    // UTC時刻をそのまま使う（テスト環境はUTC）
    expect(result.time).toMatch(/^\d{2}:\d{2}$/);
  });

  it("visit_type: first → '初診'", () => {
    expect(transformReservationToReceptionAppointment({ ...minimal, visit_type: "first" }).visitType).toBe("初診");
  });

  it("visit_type: revisit → '再診'", () => {
    expect(transformReservationToReceptionAppointment({ ...minimal, visit_type: "revisit" }).visitType).toBe("再診");
  });

  it("status: confirmed → カラム pending にマップ", () => {
    expect(transformReservationToReceptionAppointment({ ...minimal, status: "confirmed" }).status).toBe("pending");
  });

  it("status: checked_in → カラム checked_in にマップ", () => {
    expect(transformReservationToReceptionAppointment({ ...minimal, status: "checked_in" }).status).toBe("checked_in");
  });

  it("status: in_consultation → カラム in_consultation にマップ", () => {
    expect(transformReservationToReceptionAppointment({ ...minimal, status: "in_consultation" }).status).toBe("in_consultation");
  });

  it("status: accounting → カラム accounting にマップ", () => {
    expect(transformReservationToReceptionAppointment({ ...minimal, status: "accounting" }).status).toBe("accounting");
  });

  it("status: completed → カラム completed にマップ", () => {
    expect(transformReservationToReceptionAppointment({ ...minimal, status: "completed" }).status).toBe("completed");
  });

  it("owner.name を ownerName にマップする", () => {
    const result = transformReservationToReceptionAppointment({
      ...minimal,
      owner: { id: 20, clinic_id: 1, name: "田中太郎" } as BackendReservation["owner"],
    });
    expect(result.ownerName).toBe("田中太郎");
  });

  it("pet.name を petName にマップする", () => {
    const result = transformReservationToReceptionAppointment({
      ...minimal,
      pet: { id: 10, clinic_id: 1, name: "ポチ" } as BackendReservation["pet"],
    });
    expect(result.petName).toBe("ポチ");
  });

  it("LINE予約で relation が無い場合 customer_fields.owner_name を ownerName にフォールバックする", () => {
    const result = transformReservationToReceptionAppointment({
      ...minimal,
      owner: undefined,
      // customer_fields は json.RawMessage（Go）→ JSON オブジェクトとして届く
      customer_fields: { owner_name: "山田太郎" },
    });
    expect(result.ownerName).toBe("山田太郎");
  });

  it("LINE予約で relation が無い場合 customer_fields.pets[0].name を petName にフォールバックする", () => {
    const result = transformReservationToReceptionAppointment({
      ...minimal,
      pet: undefined,
      customer_fields: { pets: [{ name: "ハナ", type: "柴犬" }] },
    });
    expect(result.petName).toBe("ハナ");
  });

  it("LINE予約で relation が無い場合 customer_fields.pets[0].type を petType にフォールバックする", () => {
    const result = transformReservationToReceptionAppointment({
      ...minimal,
      pet: undefined,
      customer_fields: { pets: [{ name: "ハナ", type: "柴犬" }] },
    });
    expect(result.petType).toBe("柴犬");
  });

  it("reservation_type.name を reservationType にマップする", () => {
    const result = transformReservationToReceptionAppointment({
      ...minimal,
      reservation_type: { id: 1, clinic_id: 1, name: "診療" } as BackendReservation["reservation_type"],
    });
    expect(result.reservationType).toBe("診療");
  });

  it("is_designated をそのまま返す", () => {
    expect(transformReservationToReceptionAppointment({ ...minimal, is_designated: true }).isDesignated).toBe(true);
  });
});

describe("transformReservationsToReceptionColumns", () => {
  it("cancelled の予約はカンバンに含まれない", () => {
    const reservations: BackendReservation[] = [
      { ...minimal, id: 1, status: "confirmed" },
      { ...minimal, id: 2, status: "cancelled" },
    ];
    const columns = transformReservationsToReceptionColumns(reservations);
    const allAppointments = columns.flatMap((c) => c.appointments);
    expect(allAppointments.every((a) => a.id !== "2")).toBe(true);
  });

  it("RECEPTION_COLUMNS の全カラムが返る", () => {
    const columns = transformReservationsToReceptionColumns([]);
    expect(columns).toHaveLength(RECEPTION_COLUMNS.length);
  });

  it("confirmed の予約は pending カラムに入る", () => {
    const reservations: BackendReservation[] = [
      { ...minimal, id: 1, status: "confirmed" },
    ];
    const columns = transformReservationsToReceptionColumns(reservations);
    const pendingCol = columns.find((c) => c.id === "pending");
    expect(pendingCol?.appointments.some((a) => a.id === "1")).toBe(true);
  });

  it("in_consultation の予約は in_consultation カラムに入る", () => {
    const reservations: BackendReservation[] = [
      { ...minimal, id: 3, status: "in_consultation" },
    ];
    const columns = transformReservationsToReceptionColumns(reservations);
    const col = columns.find((c) => c.id === "in_consultation");
    expect(col?.appointments.some((a) => a.id === "3")).toBe(true);
  });
});

