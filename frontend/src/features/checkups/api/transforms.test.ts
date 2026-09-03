import { describe, it, expect } from "vitest";
import { transformCheckupGlobal } from "./transforms";
import type { BackendCheckupGlobal } from "../types";

const minimal: BackendCheckupGlobal = {
  id: "ck-1",
  medical_record_id: "mr-1",
  checkup_type_id: "ct-1",
  date: "2026-03-25T00:00:00Z",
  result: "異常なし",
  pet_name: "ポチ",
  owner_name: "田中太郎",
};

describe("transformCheckupGlobal", () => {
  it("id をそのまま返す", () => {
    expect(transformCheckupGlobal(minimal).id).toBe("ck-1");
  });

  it("date をスライスして日付部分のみ返す", () => {
    expect(transformCheckupGlobal({ ...minimal, date: "2026-03-25T10:00:00Z" }).date).toBe(
      "2026-03-25",
    );
  });

  it("date が空のとき空文字を返す", () => {
    expect(transformCheckupGlobal({ ...minimal, date: "" }).date).toBe("");
  });

  it("next_date をスライスして日付部分のみ返す", () => {
    const result = transformCheckupGlobal({ ...minimal, next_date: "2027-03-25T00:00:00Z" });
    expect(result.nextDate).toBe("2027-03-25");
  });

  it("next_date が未設定のとき undefined を返す", () => {
    expect(transformCheckupGlobal({ ...minimal, next_date: undefined }).nextDate).toBeUndefined();
  });

  it("checkup_type.name を checkupTypeName にマップする", () => {
    const result = transformCheckupGlobal({
      ...minimal,
      checkup_type: { id: "ct-1", name: "血液検査" },
    });
    expect(result.checkupTypeName).toBe("血液検査");
  });

  it("checkup_type が未設定のとき空文字を返す", () => {
    expect(transformCheckupGlobal({ ...minimal, checkup_type: undefined }).checkupTypeName).toBe(
      "",
    );
  });

  it("doctor.name を doctorName にマップする", () => {
    const result = transformCheckupGlobal({
      ...minimal,
      doctor: { id: "d-1", name: "山田医師" },
    });
    expect(result.doctorName).toBe("山田医師");
  });

  it("pet_name / owner_name をそのままマップする", () => {
    const result = transformCheckupGlobal({ ...minimal, pet_name: "タマ", owner_name: "佐藤" });
    expect(result.petName).toBe("タマ");
    expect(result.ownerName).toBe("佐藤");
  });

  it("pet_id / owner_id をそのままマップする", () => {
    const result = transformCheckupGlobal({ ...minimal, pet_id: "p-5", owner_id: "o-10" });
    expect(result.petId).toBe("p-5");
    expect(result.ownerId).toBe("o-10");
  });
});
