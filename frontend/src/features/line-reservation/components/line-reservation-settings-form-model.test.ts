import { describe, it, expect } from "vitest";
import {
  parseClosedWeekdays,
  asJsonb,
  isStringArray,
  isBusinessHours,
  isBreakHourArray,
  isBusinessHoursByWeekday,
  toDisplayTime,
  toStorageTime,
} from "../lib/line-reservation-settings-form-model";

describe("isStringArray", () => {
  it("string[] を true と判定する", () => {
    expect(isStringArray(["1", "2"])).toBe(true);
    expect(isStringArray([])).toBe(true);
  });

  it("非配列・非string要素を false と判定する", () => {
    expect(isStringArray("not-an-array")).toBe(false);
    expect(isStringArray([1, 2])).toBe(false);
    expect(isStringArray(null)).toBe(false);
  });
});

describe("isBusinessHours", () => {
  it("{ start, end } 形状を true と判定する", () => {
    expect(isBusinessHours({ start: "0900", end: "1900" })).toBe(true);
  });

  it("必須フィールド欠落・型不一致を false と判定する", () => {
    expect(isBusinessHours({ start: "0900" })).toBe(false);
    expect(isBusinessHours({ start: 900, end: "1900" })).toBe(false);
    expect(isBusinessHours(null)).toBe(false);
    expect(isBusinessHours([])).toBe(false);
    expect(isBusinessHours("0900-1900")).toBe(false);
  });
});

describe("isBreakHourArray", () => {
  it("BreakHour[] を true と判定する", () => {
    expect(isBreakHourArray([{ start: "1200", end: "1300" }])).toBe(true);
    expect(isBreakHourArray([])).toBe(true);
  });

  it("要素形状不一致を false と判定する", () => {
    expect(isBreakHourArray([{ start: "1200" }])).toBe(false);
    expect(isBreakHourArray("not-an-array")).toBe(false);
  });
});

describe("isBusinessHoursByWeekday", () => {
  it("曜日キー→BusinessHours の Record を true と判定する", () => {
    expect(
      isBusinessHoursByWeekday({
        "1": { start: "0900", end: "1900" },
        "2": { start: "1000", end: "1800" },
      }),
    ).toBe(true);
  });

  it("空オブジェクトを true と判定する（未設定状態）", () => {
    expect(isBusinessHoursByWeekday({})).toBe(true);
  });

  it("値の形状不一致・配列・非オブジェクトを false と判定する", () => {
    expect(isBusinessHoursByWeekday({ "1": { start: "0900" } })).toBe(false);
    expect(isBusinessHoursByWeekday([])).toBe(false);
    expect(isBusinessHoursByWeekday(null)).toBe(false);
    expect(isBusinessHoursByWeekday("not-an-object")).toBe(false);
  });
});

describe("asJsonb", () => {
  it("null/undefined は fallback を返す", () => {
    expect(asJsonb(null, ["fallback"], isStringArray)).toEqual(["fallback"]);
    expect(asJsonb(undefined, ["fallback"], isStringArray)).toEqual(["fallback"]);
  });

  it("string値: 正しい JSON かつ型一致なら parse 結果を返す", () => {
    expect(asJsonb(JSON.stringify(["1", "2"]), [], isStringArray)).toEqual(["1", "2"]);
  });

  it("string値: 不正 JSON なら fallback を返す", () => {
    expect(asJsonb("{not-valid-json", ["fallback"], isStringArray)).toEqual(["fallback"]);
  });

  it("string値: parse は成功するが型不一致なら fallback を返す", () => {
    expect(
      asJsonb(JSON.stringify({ start: "0900", end: "1900" }), ["fallback"], isStringArray),
    ).toEqual(["fallback"]);
  });

  it("非string値（既にオブジェクト/配列）: 型一致なら値をそのまま返す", () => {
    expect(asJsonb(["1", "2"], [], isStringArray)).toEqual(["1", "2"]);
    expect(
      asJsonb({ start: "0900", end: "1900" }, { start: "0000", end: "0000" }, isBusinessHours),
    ).toEqual({
      start: "0900",
      end: "1900",
    });
  });

  it("非string値: 型不一致なら fallback を返す", () => {
    expect(asJsonb({ foo: "bar" }, ["fallback"], isStringArray)).toEqual(["fallback"]);
  });
});

describe("parseClosedWeekdays", () => {
  it("normalizes numeric and exact legacy weekday values from arrays and JSON", () => {
    const legacyWeekdays = ["0", "2", "6"];

    expect(parseClosedWeekdays([])).toEqual([]);
    expect(parseClosedWeekdays([0])).toEqual([0]);
    expect(parseClosedWeekdays([2, 6])).toEqual([2, 6]);
    expect(parseClosedWeekdays(legacyWeekdays)).toEqual([0, 2, 6]);
    expect(parseClosedWeekdays(`[0,"2",6]`)).toEqual([0, 2, 6]);
    expect(parseClosedWeekdays(`["0","2","6"]`)).toEqual([0, 2, 6]);
    expect(parseClosedWeekdays([2, "2"])).toEqual([2, 2]);
    expect(legacyWeekdays).toEqual(["0", "2", "6"]);
  });

  it("returns an empty array for omitted or whole-null values", () => {
    expect(parseClosedWeekdays(undefined)).toEqual([]);
    expect(parseClosedWeekdays(null)).toEqual([]);
    expect(parseClosedWeekdays("null")).toEqual([]);
  });

  it("returns null for malformed, invalid, or noncanonical stored values", () => {
    const invalidValues: unknown[] = [
      {},
      "{not-json",
      "{}",
      [null],
      [true],
      [{}],
      [2.5],
      [-1],
      [7],
      ["02"],
      [" 2"],
      ["7"],
    ];

    invalidValues.forEach((value) => expect(parseClosedWeekdays(value)).toBeNull());
  });
});

describe("toDisplayTime / toStorageTime", () => {
  it("4桁文字列を HH:MM 表示に変換する", () => {
    expect(toDisplayTime("0900")).toBe("09:00");
  });

  it("既に HH:MM 形式ならそのまま返す", () => {
    expect(toDisplayTime("09:00")).toBe("09:00");
  });

  it("HH:MM をストレージ形式(コロン除去)に変換する", () => {
    expect(toStorageTime("09:00")).toBe("0900");
  });
});
