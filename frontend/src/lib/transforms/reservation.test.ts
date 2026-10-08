import { describe, expect, it } from "vitest";

import { transformReservation } from "./reservation";

// TASK-444-S1: 新規ファイルは @/types/generated/models を import できない
// （generated-models-import-allowlist.json は frozen）。wire enum は literal で記述し、
// Reservation 型は transformReservation の引数から派生させる。
type BackendReservation = Parameters<typeof transformReservation>[0];

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
  source: "manual",
  is_staff_delegated: false,
  customer_fields: {},
  created_at: "2026-03-25T00:00:00Z",
  updated_at: "2026-03-25T00:00:00Z",
};

const pet = {
  id: 10,
  clinic_id: 1,
  owner_id: 20,
  animal_species_id: 1,
  pet_number: "P00010",
  name: "ポチ",
  name_kana: "ポチ",
  gender: "male",
  status: "alive",
  birth_date: "2018-10-05",
  breed: "",
  color: "",
  food: "",
  environment: "",
  phone: "",
  remarks: "",
  created_at: "2026-03-25T00:00:00Z",
  updated_at: "2026-03-25T00:00:00Z",
} satisfies NonNullable<BackendReservation["pet"]>;

describe("transformReservation", () => {
  it("id を string に変換する", () => {
    expect(transformReservation({ ...minimal, id: 7 }).id).toBe("7");
  });

  it("pet.name を petName にマップする", () => {
    const result = transformReservation({ ...minimal, pet });
    expect(result.petName).toBe("ポチ");
  });

  it("pet.birth_date を petBirthDate にマップする (EMR-301)", () => {
    const result = transformReservation({ ...minimal, pet });
    expect(result.petBirthDate).toBe("2018-10-05");
  });

  it("pet.status を petStatus にマップする (EMR-301・死亡 suppress 判定用)", () => {
    const alive = transformReservation({ ...minimal, pet });
    expect(alive.petStatus).toBe("alive");

    const deceased = transformReservation({
      ...minimal,
      pet: { ...pet, status: "deceased" },
    });
    expect(deceased.petStatus).toBe("deceased");
    expect(deceased.petBirthDate).toBe("2018-10-05");
  });

  it("pet が未紐付けの場合 pet sentinel は undefined", () => {
    const result = transformReservation({ ...minimal, pet: undefined });
    expect(result.petStatus).toBeUndefined();
    expect(result.petBirthDate).toBeUndefined();
  });

  it("LINE予約で relation が無い場合 customer_fields.pets[0].name を petName にフォールバックする", () => {
    const result = transformReservation({
      ...minimal,
      pet: undefined,
      customer_fields: { pets: [{ name: "ハナ", type: "柴犬" }] },
    });
    expect(result.petName).toBe("ハナ");
  });
});
