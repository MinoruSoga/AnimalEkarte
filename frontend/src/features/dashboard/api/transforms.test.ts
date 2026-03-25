import { describe, it, expect } from "vitest";
import {
  transformReservationToDashboardAppointment,
  transformReservationsToDashboardColumns,
  COLUMN_ID_TO_TITLE,
  DASHBOARD_COLUMNS,
} from "./transforms";
import type { ReservationAppointment as BackendReservation } from "@/types/generated/models";

const minimal: BackendReservation = {
  id: 1,
  clinic_id: 1,
  start_time: "2026-03-25T10:00:00Z",
  end_time: "2026-03-25T10:30:00Z",
  visit_type: "first",
  status: "confirmed",
  is_designated: false,
  pet_id: 10,
  owner_id: 20,
  created_at: "2026-03-25T00:00:00Z",
  updated_at: "2026-03-25T00:00:00Z",
};

describe("transformReservationToDashboardAppointment", () => {
  it("id を string に変換する", () => {
    expect(transformReservationToDashboardAppointment({ ...minimal, id: 7 }).id).toBe("7");
  });

  it("start_time から HH:mm 形式の time を生成する", () => {
    const result = transformReservationToDashboardAppointment({
      ...minimal,
      start_time: "2026-03-25T10:30:00Z",
    });
    // UTC時刻をそのまま使う（テスト環境はUTC）
    expect(result.time).toMatch(/^\d{2}:\d{2}$/);
  });

  it("visit_type: first → '初診'", () => {
    expect(transformReservationToDashboardAppointment({ ...minimal, visit_type: "first" }).visitType).toBe("初診");
  });

  it("visit_type: revisit → '再診'", () => {
    expect(transformReservationToDashboardAppointment({ ...minimal, visit_type: "revisit" }).visitType).toBe("再診");
  });

  it("status: confirmed → カラム pending にマップ", () => {
    expect(transformReservationToDashboardAppointment({ ...minimal, status: "confirmed" }).status).toBe("pending");
  });

  it("status: checked_in → カラム checked_in にマップ", () => {
    expect(transformReservationToDashboardAppointment({ ...minimal, status: "checked_in" }).status).toBe("checked_in");
  });

  it("status: in_consultation → カラム in_consultation にマップ", () => {
    expect(transformReservationToDashboardAppointment({ ...minimal, status: "in_consultation" }).status).toBe("in_consultation");
  });

  it("status: accounting → カラム accounting にマップ", () => {
    expect(transformReservationToDashboardAppointment({ ...minimal, status: "accounting" }).status).toBe("accounting");
  });

  it("status: completed → カラム completed にマップ", () => {
    expect(transformReservationToDashboardAppointment({ ...minimal, status: "completed" }).status).toBe("completed");
  });

  it("owner.owner_name を ownerName にマップする", () => {
    const result = transformReservationToDashboardAppointment({
      ...minimal,
      owner: { id: 20, clinic_id: 1, owner_name: "田中太郎" } as BackendReservation["owner"],
    });
    expect(result.ownerName).toBe("田中太郎");
  });

  it("pet.name を petName にマップする", () => {
    const result = transformReservationToDashboardAppointment({
      ...minimal,
      pet: { id: 10, clinic_id: 1, name: "ポチ" } as BackendReservation["pet"],
    });
    expect(result.petName).toBe("ポチ");
  });

  it("service_type.name を serviceType にマップする", () => {
    const result = transformReservationToDashboardAppointment({
      ...minimal,
      service_type: { id: 1, clinic_id: 1, name: "診療" } as BackendReservation["service_type"],
    });
    expect(result.serviceType).toBe("診療");
  });

  it("is_designated をそのまま返す", () => {
    expect(transformReservationToDashboardAppointment({ ...minimal, is_designated: true }).isDesignated).toBe(true);
  });
});

describe("transformReservationsToDashboardColumns", () => {
  it("cancelled の予約はカンバンに含まれない", () => {
    const reservations: BackendReservation[] = [
      { ...minimal, id: 1, status: "confirmed" },
      { ...minimal, id: 2, status: "cancelled" },
    ];
    const columns = transformReservationsToDashboardColumns(reservations);
    const allAppointments = columns.flatMap((c) => c.appointments);
    expect(allAppointments.every((a) => a.id !== "2")).toBe(true);
  });

  it("DASHBOARD_COLUMNS の全カラムが返る", () => {
    const columns = transformReservationsToDashboardColumns([]);
    expect(columns).toHaveLength(DASHBOARD_COLUMNS.length);
  });

  it("confirmed の予約は pending カラムに入る", () => {
    const reservations: BackendReservation[] = [
      { ...minimal, id: 1, status: "confirmed" },
    ];
    const columns = transformReservationsToDashboardColumns(reservations);
    const pendingCol = columns.find((c) => c.id === "pending");
    expect(pendingCol?.appointments.some((a) => a.id === "1")).toBe(true);
  });

  it("in_consultation の予約は in_consultation カラムに入る", () => {
    const reservations: BackendReservation[] = [
      { ...minimal, id: 3, status: "in_consultation" },
    ];
    const columns = transformReservationsToDashboardColumns(reservations);
    const col = columns.find((c) => c.id === "in_consultation");
    expect(col?.appointments.some((a) => a.id === "3")).toBe(true);
  });
});

describe("COLUMN_ID_TO_TITLE", () => {
  it("pending → '受付予約'", () => {
    expect(COLUMN_ID_TO_TITLE.pending).toBe("受付予約");
  });

  it("checked_in → '受付済'", () => {
    expect(COLUMN_ID_TO_TITLE.checked_in).toBe("受付済");
  });

  it("completed → '会計済'", () => {
    expect(COLUMN_ID_TO_TITLE.completed).toBe("会計済");
  });
});
