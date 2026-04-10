import { describe, it, expect } from "vitest";
import { transformReservation, transformToCreateRequest } from "./transforms";
import type { ReservationAppointment as BackendReservation } from "@/types/generated/models";
import type { ReservationAppointment } from "@/types";

/** BackendReservation の最小スタブ */
const minimalBackend: BackendReservation = {
  id: 1,
  clinic_id: 1,
  start_time: "2026-03-25T10:00:00Z",
  end_time: "2026-03-25T10:30:00Z",
  visit_type: "first",
  status: "confirmed",
  is_designated: false,
  created_at: "2026-03-25T00:00:00Z",
  updated_at: "2026-03-25T00:00:00Z",
};

describe("transformReservation", () => {
  it("id を string に変換する", () => {
    const result = transformReservation({ ...minimalBackend, id: 42 });
    expect(result.id).toBe("42");
  });

  it("id が null/undefined のとき '0' を返す", () => {
    const result = transformReservation({ ...minimalBackend, id: undefined as unknown as number });
    expect(result.id).toBe("0");
  });

  it("start_time を Date に変換する", () => {
    const result = transformReservation({ ...minimalBackend, start_time: "2026-03-25T10:00:00Z" });
    expect(result.start).toBeInstanceOf(Date);
    expect(result.start.toISOString()).toBe("2026-03-25T10:00:00.000Z");
  });

  it("end_time を Date に変換する", () => {
    const result = transformReservation({ ...minimalBackend, end_time: "2026-03-25T10:30:00Z" });
    expect(result.end).toBeInstanceOf(Date);
    expect(result.end.toISOString()).toBe("2026-03-25T10:30:00.000Z");
  });

  it("visit_type: first をそのままマップする", () => {
    const result = transformReservation({ ...minimalBackend, visit_type: "first" });
    expect(result.visitType).toBe("first");
  });

  it("visit_type: revisit をそのままマップする", () => {
    const result = transformReservation({ ...minimalBackend, visit_type: "revisit" });
    expect(result.visitType).toBe("revisit");
  });

  it("visit_type が未設定のとき 'first' にフォールバックする", () => {
    const result = transformReservation({ ...minimalBackend, visit_type: undefined });
    expect(result.visitType).toBe("first");
  });

  it("status をそのままマップする", () => {
    const statuses: ReservationAppointment["status"][] = [
      "confirmed", "pending", "checked_in", "in_consultation", "accounting", "completed", "cancelled",
    ];
    for (const status of statuses) {
      const result = transformReservation({ ...minimalBackend, status });
      expect(result.status).toBe(status);
    }
  });

  it("status が未設定のとき 'pending' にフォールバックする", () => {
    const result = transformReservation({ ...minimalBackend, status: undefined });
    expect(result.status).toBe("pending");
  });

  it("owner.owner_name を ownerName にマップする", () => {
    const result = transformReservation({
      ...minimalBackend,
      owner: { id: 1, clinic_id: 1, owner_name: "田中太郎" } as BackendReservation["owner"],
    });
    expect(result.ownerName).toBe("田中太郎");
  });

  it("owner が未設定でも pet.owner.owner_name を ownerName にマップする", () => {
    const result = transformReservation({
      ...minimalBackend,
      owner: undefined,
      pet: {
        id: 1,
        clinic_id: 1,
        owner: { id: 1, clinic_id: 1, owner_name: "佐藤花子" } as BackendReservation["pet"]["owner"],
      } as BackendReservation["pet"],
    });
    expect(result.ownerName).toBe("佐藤花子");
  });

  it("owner も pet.owner も未設定のとき ownerName は空文字", () => {
    const result = transformReservation({ ...minimalBackend, owner: undefined, pet: undefined });
    expect(result.ownerName).toBe("");
  });

  it("pet.name を petName にマップする", () => {
    const result = transformReservation({
      ...minimalBackend,
      pet: { id: 1, clinic_id: 1, name: "ポチ" } as BackendReservation["pet"],
    });
    expect(result.petName).toBe("ポチ");
  });

  it("reservation_category.name を type にマップする", () => {
    const result = transformReservation({
      ...minimalBackend,
      reservation_category: { id: 1, clinic_id: 1, name: "診療" } as BackendReservation["reservation_category"],
    });
    expect(result.type).toBe("診療");
  });

  it("doctor.name を doctor にマップする", () => {
    const result = transformReservation({
      ...minimalBackend,
      doctor: { id: 1, clinic_id: 1, name: "山田医師" } as BackendReservation["doctor"],
    });
    expect(result.doctor).toBe("山田医師");
  });

  it("is_designated をそのままマップする", () => {
    const result = transformReservation({ ...minimalBackend, is_designated: true });
    expect(result.isDesignated).toBe(true);
  });

  it("notes が存在するとき notes をマップする", () => {
    const result = transformReservation({ ...minimalBackend, notes: "備考テスト" });
    expect(result.notes).toBe("備考テスト");
  });

  it("notes が空文字のとき undefined を返す", () => {
    const result = transformReservation({ ...minimalBackend, notes: "" });
    expect(result.notes).toBeUndefined();
  });

  it("pet_id を string に変換して petId にマップする", () => {
    const result = transformReservation({ ...minimalBackend, pet_id: 5 });
    expect(result.petId).toBe("5");
  });

  it("pet_id が未設定のとき petId は undefined", () => {
    const result = transformReservation({ ...minimalBackend, pet_id: undefined });
    expect(result.petId).toBeUndefined();
  });
});

describe("transformToCreateRequest", () => {
  const baseData: Partial<ReservationAppointment> = {
    start: new Date("2026-03-25T10:00:00Z"),
    end: new Date("2026-03-25T10:30:00Z"),
    visitType: "first",
    type: "1",
    isDesignated: false,
  };

  it("pet_id / owner_id を number に変換する", () => {
    const result = transformToCreateRequest(baseData, "10", "20");
    expect(result.pet_id).toBe(10);
    expect(result.owner_id).toBe(20);
  });

  it("start/end を ISO 文字列に変換する", () => {
    const result = transformToCreateRequest(baseData, "1", "1");
    expect(result.start_time).toBe("2026-03-25T10:00:00.000Z");
    expect(result.end_time).toBe("2026-03-25T10:30:00.000Z");
  });

  it("start が未設定のとき start_time は空文字", () => {
    const result = transformToCreateRequest({ ...baseData, start: undefined }, "1", "1");
    expect(result.start_time).toBe("");
  });

  it("visitType が未設定のとき 'first' にフォールバックする", () => {
    const result = transformToCreateRequest({ ...baseData, visitType: undefined }, "1", "1");
    expect(result.visit_type).toBe("first");
  });

  it("type を reservation_category_id (number) に変換する", () => {
    const result = transformToCreateRequest({ ...baseData, type: "3" }, "1", "1");
    expect(result.reservation_category_id).toBe(3);
  });

  it("notes を正しく渡す", () => {
    const result = transformToCreateRequest({ ...baseData, notes: "備考" }, "1", "1");
    expect(result.notes).toBe("備考");
  });

  it("notes が未設定のとき undefined を渡す", () => {
    const result = transformToCreateRequest({ ...baseData, notes: undefined }, "1", "1");
    expect(result.notes).toBeUndefined();
  });
});
