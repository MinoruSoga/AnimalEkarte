import { describe, expect, it } from "vitest";

import type { CheckupRecord } from "../api/transforms";
import {
  FILTER_PROPERTIES,
  buildCheckupListFilters,
  filterCheckupsBySearch,
  filterCheckupsBySpecies,
} from "./checkups-list-model";

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

describe("FILTER_PROPERTIES — 動物種セレクト (EMR-223)", () => {
  it("species キーの select プロパティが 犬/猫/その他 の選択肢を持つ", () => {
    const prop = FILTER_PROPERTIES.find((p) => p.key === "species");
    expect(prop).toBeDefined();
    expect(prop?.label).toBe("動物種");
    expect(prop?.type).toBe("select");
    expect(prop?.options?.map((o) => o.label)).toEqual(["犬", "猫", "その他"]);
  });
});

describe("filterCheckupsBySpecies", () => {
  const speciesByPetId = new Map<string, string>([
    ["p-dog", "犬"],
    ["p-cat", "猫"],
    ["p-rabbit", "うさぎ"],
    ["p-other", "その他"],
  ]);
  const rows = [
    record({ id: "1", petId: "p-dog" }),
    record({ id: "2", petId: "p-cat" }),
    record({ id: "3", petId: "p-rabbit" }),
    record({ id: "4", petId: "p-other" }),
    record({ id: "5", petId: undefined }),
    record({ id: "6", petId: "p-unresolved" }),
  ];

  it("未選択なら全件返す", () => {
    expect(filterCheckupsBySpecies(rows, undefined, speciesByPetId)).toHaveLength(6);
    expect(filterCheckupsBySpecies(rows, "", speciesByPetId)).toHaveLength(6);
  });

  it("犬 を選ぶと犬の行だけ残る", () => {
    expect(filterCheckupsBySpecies(rows, "犬", speciesByPetId).map((c) => c.id)).toEqual(["1"]);
  });

  it("猫 を選ぶと猫の行だけ残る", () => {
    expect(filterCheckupsBySpecies(rows, "猫", speciesByPetId).map((c) => c.id)).toEqual(["2"]);
  });

  it("その他 は 犬/猫 以外の種名を持つ行だけ残る（マスタ種「その他」自身を含む）", () => {
    expect(filterCheckupsBySpecies(rows, "その他", speciesByPetId).map((c) => c.id)).toEqual([
      "3",
      "4",
    ]);
  });

  it("petId 欠落・種名未解決の行はフィルタ適用中に除外される", () => {
    expect(filterCheckupsBySpecies(rows, "犬", speciesByPetId).map((c) => c.id)).toEqual(["1"]);
    expect(filterCheckupsBySpecies(rows, "その他", speciesByPetId).map((c) => c.id)).toEqual([
      "3",
      "4",
    ]);
  });
});

describe("buildCheckupListFilters — species フィルタ非混入 (EMR-223)", () => {
  it("species フィルタが activeFilters にあっても API パラメータは増えない", () => {
    const filters = buildCheckupListFilters([
      { key: "species", condition: "is", value: "犬", displayValue: "犬" },
    ]);
    expect(filters).toEqual({
      startDate: undefined,
      endDate: undefined,
      nextStartDate: undefined,
      nextEndDate: undefined,
    });
  });
});
