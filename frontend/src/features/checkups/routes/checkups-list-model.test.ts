import { describe, expect, it } from "vitest";

import type { CheckupRecord } from "../api/transforms";
import { filterCheckupsBySearch } from "./checkups-list-model";

function record(overrides: Partial<CheckupRecord> = {}): CheckupRecord {
  return {
    id: "chk-1",
    medicalRecordId: "mr-1",
    checkupTypeId: "ct-1",
    petId: "p-1",
    date: "2026-01-01",
    nextDate: undefined,
    result: "正常",
    checkupTypeName: "定期健診",
    doctorName: "鈴木",
    petName: "ポチ",
    ownerName: "山田",
    ownerId: "o-1",
    ...overrides,
  };
}

describe("filterCheckupsBySearch", () => {
  const rows = [
    record({
      id: "1",
      petName: "ポチ",
      ownerName: "ヤマダ",
      checkupTypeName: "血液検査",
      result: "イジョウナシ",
    }),
    record({
      id: "2",
      petName: "たろう",
      ownerName: "さとう",
      checkupTypeName: "尿検査",
      result: "けっかなし",
    }),
  ];

  it("空文字なら全件返す", () => {
    expect(filterCheckupsBySearch(rows, "")).toHaveLength(2);
  });

  it("ひらがな入力でカタカナ petName がヒットする", () => {
    const result = filterCheckupsBySearch(rows, "ぽち");
    expect(result.map((c) => c.id)).toEqual(["1"]);
  });

  it("カタカナ入力でひらがな ownerName がヒットする", () => {
    const result = filterCheckupsBySearch(rows, "サトウ");
    expect(result.map((c) => c.id)).toEqual(["2"]);
  });

  it("ひらがな入力でカタカナ result がヒットする（かな正規化漏れの回帰防止）", () => {
    const result = filterCheckupsBySearch(rows, "いじょうなし");
    expect(result.map((c) => c.id)).toEqual(["1"]);
  });

  it("カタカナ入力でひらがな result がヒットする", () => {
    const result = filterCheckupsBySearch(rows, "ケッカナシ");
    expect(result.map((c) => c.id)).toEqual(["2"]);
  });
});
