import { describe, expect, it } from "vitest";

import { calculateSaigram, type SaigramResult } from "./saigram";

const MS_PER_DAY = 86_400_000;
const BASE_DATE_MS = Date.UTC(1999, 10, 8);

const EXPECTED_TYPES = [
  5, 4, 8, 6, 3, 11, 5, 4, 8, 6,
  1, 7, 9, 2, 8, 6, 1, 7, 9, 2,
  12, 12, 2, 9, 9, 2, 12, 12, 2, 9,
  7, 1, 6, 8, 2, 9, 7, 1, 6, 8,
  4, 5, 11, 3, 6, 8, 4, 5, 11, 3,
  10, 10, 3, 11, 11, 3, 10, 10, 3, 11,
] as const;

const TYPE_METADATA = [
  { typeCode: "F1", ha: "☆楽観派" },
  { typeCode: "F2", ha: "★慎重派" },
  { typeCode: "F3", ha: "☆楽観派" },
  { typeCode: "F4", ha: "★慎重派" },
  { typeCode: "A5", ha: "☆楽観派" },
  { typeCode: "A6", ha: "★慎重派" },
  { typeCode: "A7", ha: "☆楽観派" },
  { typeCode: "A8", ha: "★慎重派" },
  { typeCode: "M9", ha: "☆楽観派" },
  { typeCode: "M10", ha: "★慎重派" },
  { typeCode: "M11", ha: "☆楽観派" },
  { typeCode: "M12", ha: "★慎重派" },
] as const;

function dateAtOffset(offset: number): string {
  return new Date(BASE_DATE_MS + offset * MS_PER_DAY).toISOString().slice(0, 10);
}

function expectedResult(kosei: number, typeNo: number): SaigramResult {
  const metadata = TYPE_METADATA[typeNo - 1];
  return {
    kosei,
    typeNo,
    ...metadata,
  };
}

describe("calculateSaigram", () => {
  it.each(
    EXPECTED_TYPES.map((typeNo, offset) => ({
      birthDate: dateAtOffset(offset),
      want: expectedResult(offset + 1, typeNo),
    })),
  )("$birthDateを個性番号$want.koseiの完全な分類結果へ対応付ける", ({ birthDate, want }) => {
    expect(calculateSaigram(birthDate)).toEqual(want);
  });

  it.each([
    ["1962-04-10", { kosei: 15, typeNo: 8, typeCode: "A8", ha: "★慎重派" }],
    ["2002-10-04", { kosei: 42, typeNo: 5, typeCode: "A5", ha: "☆楽観派" }],
    ["2003-12-11", { kosei: 55, typeNo: 11, typeCode: "M11", ha: "☆楽観派" }],
    ["2006-02-24", { kosei: 21, typeNo: 12, typeCode: "M12", ha: "★慎重派" }],
    ["2000-02-29", { kosei: 54, typeNo: 11, typeCode: "M11", ha: "☆楽観派" }],
  ] satisfies ReadonlyArray<readonly [string, SaigramResult]>)(
    "仕様書の代表例 %s を完全な分類結果へ対応付ける",
    (birthDate, want) => {
      expect(calculateSaigram(birthDate)).toEqual(want);
    },
  );

  it("基準日前日の負の剰余を個性番号60へ非負化する", () => {
    expect(calculateSaigram("1999-11-07")).toEqual({
      kosei: 60,
      typeNo: 11,
      typeCode: "M11",
      ha: "☆楽観派",
    });
  });

  it("基準日から60日後に個性番号1へ循環する", () => {
    expect(calculateSaigram("2000-01-07")).toEqual({
      kosei: 1,
      typeNo: 5,
      typeCode: "A5",
      ha: "☆楽観派",
    });
  });

  it("うるう日を有効な暦日として扱う", () => {
    expect(calculateSaigram("2000-02-29")).toEqual({
      kosei: 54,
      typeNo: 11,
      typeCode: "M11",
      ha: "☆楽観派",
    });
  });

  it.each([
    "",
    "not-a-date",
    "1999-1-08",
    "1999-11-8",
    "1999/11/08",
    "1999-11-08T00:00:00Z",
  ])("空またはYYYY-MM-DDでない入力 %j はnullを返す", (birthDate) => {
    expect(calculateSaigram(birthDate)).toBeNull();
  });

  it.each([
    "2001-02-29",
    "2000-02-30",
    "1999-04-31",
    "1999-00-10",
    "1999-13-01",
    "1999-11-00",
    "1999-11-31",
  ])("存在しない暦日 %s はnullを返す", (birthDate) => {
    expect(calculateSaigram(birthDate)).toBeNull();
  });
});
