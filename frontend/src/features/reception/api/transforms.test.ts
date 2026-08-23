import { describe, it, expect } from "vitest";
import { toJSTWallDate } from "@/lib/jst-date";
import {
  transformReservationToReceptionAppointment,
  transformReservationsToReceptionColumns,
  RECEPTION_COLUMNS,
} from "./transforms";
import {
  DangerLevelHigh,
  PetGenderMale,
  PetStatusDeceased,
  ReservationSourceManual,
  type Reservation as BackendReservation,
} from "@/types/generated/models";

type BackendPetWithDangerReason = NonNullable<BackendReservation["pet"]> & {
  danger_reason?: string;
};

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
  notes: "",
  source: ReservationSourceManual,
  is_staff_delegated: false,
  customer_fields: {},
  created_at: "2026-03-25T00:00:00Z",
  updated_at: "2026-03-25T00:00:00Z",
};

describe("transformReservationToReceptionAppointment", () => {
  it("id を string に変換する", () => {
    expect(transformReservationToReceptionAppointment({ ...minimal, id: 7 }).id).toBe("7");
  });

  it("start_time から JST 基準の HH:mm 形式の time を生成する", () => {
    const result = transformReservationToReceptionAppointment({
      ...minimal,
      start_time: "2026-03-25T10:30:00Z",
    });
    expect(result.time).toBe("19:30");
  });

  it("start_time から JST 基準の visitDate を生成する", () => {
    const result = transformReservationToReceptionAppointment({
      ...minimal,
      start_time: "2026-03-25T15:30:00Z",
    });

    expect(result.visitDate).toBe("2026-03-26");
  });

  it("end_time が start+1h でなくても実際の終了時刻を保持する", () => {
    const result = transformReservationToReceptionAppointment({
      ...minimal,
      start_time: "2026-03-25T10:00:00Z",
      end_time: "2026-03-25T10:20:00Z",
    });

    expect(result.end).toEqual(toJSTWallDate("2026-03-25T10:20:00Z"));
    expect(result.end?.getHours()).toBe(19);
    expect(result.end?.getMinutes()).toBe(20);
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

  it("pet.status と pet.danger_level を reception sentinel にマップする", () => {
    const result = transformReservationToReceptionAppointment({
      ...minimal,
      pet: {
        id: 10,
        clinic_id: 1,
        owner_id: 20,
        animal_species_id: 1,
        pet_number: "P00010",
        name: "ポチ",
        name_kana: "ポチ",
        gender: PetGenderMale,
        status: PetStatusDeceased,
        breed: "",
        color: "",
        danger_level: DangerLevelHigh,
        food: "",
        environment: "",
        phone: "",
        remarks: "",
        created_at: "2026-03-25T00:00:00Z",
        updated_at: "2026-03-25T00:00:00Z",
      } satisfies NonNullable<BackendReservation["pet"]>,
    });

    expect(result.petStatus).toBe(PetStatusDeceased);
    expect(result.petDangerLevel).toBe(DangerLevelHigh);
  });

  it("pet.danger_reason を petDangerReason にマップする", () => {
    const pet = {
      id: 10,
      clinic_id: 1,
      owner_id: 20,
      animal_species_id: 1,
      pet_number: "P00010",
      name: "ポチ",
      name_kana: "ポチ",
      gender: PetGenderMale,
      status: PetStatusDeceased,
      breed: "",
      color: "",
      danger_level: DangerLevelHigh,
      danger_reason: "保定時に噛む",
      food: "",
      environment: "",
      phone: "",
      remarks: "",
      created_at: "2026-03-25T00:00:00Z",
      updated_at: "2026-03-25T00:00:00Z",
    } satisfies BackendPetWithDangerReason;

    const result = transformReservationToReceptionAppointment({
      ...minimal,
      pet,
    });

    expect(result.petDangerReason).toBe("保定時に噛む");
  });

  it("pet が未紐付けの場合 reception sentinel は undefined", () => {
    const result = transformReservationToReceptionAppointment({
      ...minimal,
      pet: undefined,
    });

    expect(result.petStatus).toBeUndefined();
    expect(result.petDangerLevel).toBeUndefined();
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

  it("pet_id / owner_id が未確定の場合は 0 ではなく空文字にする", () => {
    const result = transformReservationToReceptionAppointment({
      ...minimal,
      pet_id: undefined,
      owner_id: undefined,
      pet: undefined,
      owner: undefined,
      customer_fields: { owner_name: "山田太郎", pets: [{ name: "ハナ", type: "柴犬" }] },
    });

    expect(result.petId).toBe("");
    expect(result.ownerId).toBe("");
  });

  it("pet_id / owner_id が null の場合も空文字にする", () => {
    const result = transformReservationToReceptionAppointment({
      ...minimal,
      pet_id: null,
      owner_id: null,
    } as BackendReservation);

    expect(result.petId).toBe("");
    expect(result.ownerId).toBe("");
  });

  it("reservation_type.name を reservationType にマップする", () => {
    const result = transformReservationToReceptionAppointment({
      ...minimal,
      reservation_type: { id: 1, clinic_id: 1, name: "診療" } as BackendReservation["reservation_type"],
    });
    expect(result.reservationType).toBe("診療");
  });

  it("reservation_type.category を reservationCategory にマップする", () => {
    const result = transformReservationToReceptionAppointment({
      ...minimal,
      reservation_type: { id: 1, clinic_id: 1, name: "シャンプー", category: "trimming" } as BackendReservation["reservation_type"],
    });
    expect(result.reservationCategory).toBe("trimming");
  });

  it("reservation_type_id / doctor_id を編集フォーム用IDとして保持する", () => {
    const result = transformReservationToReceptionAppointment({
      ...minimal,
      reservation_type_id: 9,
      doctor_id: 33,
    });

    expect(result.reservationTypeId).toBe("9");
    expect(result.doctorId).toBe("33");
  });

  it("is_designated をそのまま返す", () => {
    expect(transformReservationToReceptionAppointment({ ...minimal, is_designated: true }).isDesignated).toBe(true);
  });

  // 受付ヘッダー テレメトリ Phase 2（change-ui.md）: checked_in_at のマッピング。
  it("checked_in_at を checkedInAt にマップする", () => {
    const withCheckedInAt = { ...minimal, checked_in_at: "2026-07-05T01:00:00Z" };
    expect(transformReservationToReceptionAppointment(withCheckedInAt).checkedInAt).toBe("2026-07-05T01:00:00Z");
  });

  it("checked_in_at が未設定の場合 checkedInAt は undefined", () => {
    expect(transformReservationToReceptionAppointment(minimal).checkedInAt).toBeUndefined();
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

  it("no_show の予約はカンバンに含まれない", () => {
    const reservations: BackendReservation[] = [
      { ...minimal, id: 1, status: "confirmed" },
      { ...minimal, id: 2, status: "no_show" },
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
