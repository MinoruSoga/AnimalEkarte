import { describe, it, expect } from "vitest";
import { transformVaccination } from "./transforms";
import type { BackendVaccination } from "./types";

const minimal: BackendVaccination = {
  id: 1,
  clinic_id: 1,
  pet_id: 10,
  vaccine_id: 2,
  date: "2026-03-25T00:00:00Z",
  next_date: "2027-03-25T00:00:00Z",
  created_at: "2026-03-25T00:00:00Z",
  updated_at: "2026-03-25T00:00:00Z",
};

describe("transformVaccination", () => {
  it("id を string に変換する", () => {
    expect(transformVaccination({ ...minimal, id: 99 }).id).toBe("99");
  });

  it("pet_id を string に変換して petId にマップする", () => {
    expect(transformVaccination({ ...minimal, pet_id: 5 }).petId).toBe("5");
  });

  it("pet_id が未設定のとき petId は undefined", () => {
    expect(transformVaccination({ ...minimal, pet_id: undefined }).petId).toBeUndefined();
  });

  it("medical_record_id を string に変換して medicalRecordId にマップする", () => {
    expect(transformVaccination({ ...minimal, medical_record_id: 44 }).medicalRecordId).toBe("44");
  });

  it("vaccine_id を string に変換して vaccineId にマップする", () => {
    expect(transformVaccination({ ...minimal, vaccine_id: 3 }).vaccineId).toBe("3");
  });

  it("date を JST 壁日付 YYYY-MM-DD で返す", () => {
    expect(transformVaccination({ ...minimal, date: "2026-03-25T10:00:00Z" }).date).toBe("2026-03-25");
  });

  it("UTC 境界 instant の date/nextDate は JST 暦日になる", () => {
    // 2026-03-25T16:30:00Z → JST 2026-03-26 01:30（UTC slice だと 03-25 のまま 1 日ずれ）
    const result = transformVaccination({
      ...minimal,
      date: "2026-03-25T16:30:00Z",
      next_date: "2026-04-25T15:00:00Z",
    });
    expect(result.date).toBe("2026-03-26");
    expect(result.nextDate).toBe("2026-04-26");
  });

  it("date が未設定のとき空文字を返す", () => {
    expect(transformVaccination({ ...minimal, date: undefined as unknown as string }).date).toBe("");
  });

  it("next_date を JST 壁日付 YYYY-MM-DD で返す", () => {
    expect(transformVaccination({ ...minimal, next_date: "2027-03-25T00:00:00Z" }).nextDate).toBe("2027-03-25");
  });

  it("next_date が未設定のとき空文字を返す", () => {
    expect(transformVaccination({ ...minimal, next_date: undefined as unknown as string }).nextDate).toBe("");
  });

  it("pet.owner.name を ownerName にマップする", () => {
    const result = transformVaccination({
      ...minimal,
      pet: {
        id: 1, clinic_id: 1,
        owner: { id: 1, clinic_id: 1, name: "田中太郎" } as BackendVaccination["pet"]["owner"],
      } as BackendVaccination["pet"],
    });
    expect(result.ownerName).toBe("田中太郎");
  });

  it("pet.name を petName にマップする", () => {
    const result = transformVaccination({
      ...minimal,
      pet: { id: 1, clinic_id: 1, name: "ポチ" } as BackendVaccination["pet"],
    });
    expect(result.petName).toBe("ポチ");
  });

  it("vaccine.name を vaccineName にマップする", () => {
    const result = transformVaccination({
      ...minimal,
      vaccine: { id: 2, clinic_id: 1, name: "狂犬病ワクチン" } as BackendVaccination["vaccine"],
    });
    expect(result.vaccineName).toBe("狂犬病ワクチン");
  });

  it("doctor.name を doctor にマップする", () => {
    const result = transformVaccination({
      ...minimal,
      doctor: { id: 3, clinic_id: 1, name: "山田医師" } as BackendVaccination["doctor"],
    });
    expect(result.doctor).toBe("山田医師");
  });
});
