import { describe, expect, it } from "vitest";
import {
  parseISODate,
  addDaysISO,
  getJSTToday,
  formatJapaneseDate,
  formatJSTApplicationDate,
  formatTimeHHMM,
} from "./jst-date";

// FE4-10: line-reserve/src/lib/jst-date.ts から逐語移動。日付系で唯一テストが無かったため
// 各関数の入出力をここで固定する。UTC 深夜 parse 契約（main 系のローカル正午 parse とは異なる）。
describe("shared-liff jst-date", () => {
  it("parseISODate: YYYY-MM-DD を UTC 深夜の Date に変換する", () => {
    const date = parseISODate("2026-07-11");
    expect(date.getUTCFullYear()).toBe(2026);
    expect(date.getUTCMonth()).toBe(6);
    expect(date.getUTCDate()).toBe(11);
    expect(date.getUTCHours()).toBe(0);
  });

  it("addDaysISO: 日数を加算した YYYY-MM-DD を返す", () => {
    expect(addDaysISO("2026-07-11", 1)).toBe("2026-07-12");
    expect(addDaysISO("2026-07-11", -1)).toBe("2026-07-10");
  });

  it("addDaysISO: 月末を跨ぐ加算を正しく処理する", () => {
    expect(addDaysISO("2026-07-31", 1)).toBe("2026-08-01");
    expect(addDaysISO("2026-08-01", -1)).toBe("2026-07-31");
  });

  it("addDaysISO: 年末を跨ぐ加算を正しく処理する", () => {
    expect(addDaysISO("2026-12-31", 1)).toBe("2027-01-01");
  });

  it("getJSTToday: JST 日付境界の UTC 深夜 Date を返す", () => {
    const today = getJSTToday();
    expect(today.getUTCHours()).toBe(0);
    expect(today.getUTCMinutes()).toBe(0);
  });

  it("formatJapaneseDate: 「Y年M月D日（曜）」形式にする（padded=false）", () => {
    // 2026-07-11 は土曜日
    expect(formatJapaneseDate("2026-07-11")).toBe("2026年7月11日（土）");
  });

  it("formatJapaneseDate: padded=true で月日を 2 桁 0 パディングする", () => {
    // 2026-07-05 は日曜日
    expect(formatJapaneseDate("2026-07-05", true)).toBe("2026年07月05日(日)");
  });

  it("formatJapaneseDate: 空文字は空文字を返す", () => {
    expect(formatJapaneseDate("")).toBe("");
  });

  it("formatJSTApplicationDate: instant を「申し込み」日付文字列にする", () => {
    expect(formatJSTApplicationDate("2026-07-10T15:00:00Z")).toBe("2026年7月11日 申し込み");
  });

  it("formatJSTApplicationDate: 空文字は空文字を返す", () => {
    expect(formatJSTApplicationDate("")).toBe("");
  });

  it("formatTimeHHMM: HHMM を HH:MM に変換する", () => {
    expect(formatTimeHHMM("1000")).toBe("10:00");
  });

  it("formatTimeHHMM: 空文字と 4 文字未満はそのまま返す", () => {
    expect(formatTimeHHMM("")).toBe("");
    expect(formatTimeHHMM("123")).toBe("123");
  });
});
