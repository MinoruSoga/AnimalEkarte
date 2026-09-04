import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import {
  buildJSTWallDateTime,
  currentJSTMonthDateRange,
  currentJSTYearMonth,
  daysSince,
  formatJSTDate,
  formatJSTDateTimeLocal,
  formatJSTTime,
  jstDateStartISOString,
  jstDateTimeLocalToISOString,
  jstNowISOString,
  jstWallDateToISOString,
  isPastJSTDate,
  toJSTWallDate,
} from "./jst-date";

describe("jst-date", () => {
  it("UTC instant を JST の日付と時刻に変換する", () => {
    expect(formatJSTDate("2026-03-25T15:30:00Z")).toBe("2026-03-26");
    expect(formatJSTTime("2026-03-25T15:30:00Z")).toBe("00:30");
  });

  it("API の instant を画面用の JST 壁時計 Date に変換する", () => {
    const date = toJSTWallDate("2026-03-25T10:00:00Z");

    expect(date.getFullYear()).toBe(2026);
    expect(date.getMonth()).toBe(2);
    expect(date.getDate()).toBe(25);
    expect(date.getHours()).toBe(19);
    expect(date.getMinutes()).toBe(0);
  });

  it("JST 壁時計 Date を API 送信用 ISO instant に戻す", () => {
    const wallDate = buildJSTWallDateTime("2026-06-01", "09:45");

    expect(jstWallDateToISOString(wallDate)).toBe("2026-06-01T00:45:00.000Z");
  });

  it("日付専用フィールドを JST 00:00 の ISO 文字列にする", () => {
    expect(jstDateStartISOString("2026-06-01")).toBe("2026-06-01T00:00:00+09:00");
  });

  it("datetime-local の値を JST オフセット付き ISO 文字列にする", () => {
    expect(jstDateTimeLocalToISOString("2026-06-01T09:45")).toBe("2026-06-01T09:45:00+09:00");
  });

  it("instant を datetime-local 用の JST 壁時計文字列にする", () => {
    expect(formatJSTDateTimeLocal("2026-06-01T00:45:00Z")).toBe("2026-06-01T09:45");
  });

  it("現在時刻も JST オフセット付き ISO 文字列を返す", () => {
    expect(jstNowISOString()).toMatch(/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{3}\+09:00$/);
  });

  describe("JST 日付跨ぎ境界", () => {
    beforeEach(() => {
      vi.useFakeTimers();
    });

    afterEach(() => {
      vi.useRealTimers();
    });

    it("daysSinceは日付跨ぎをJST基準で数える", () => {
      // JST では 2026-03-31 と 2026-04-01 は 1 日差（月跨ぎ）
      expect(daysSince("2026-03-31", "2026-04-01")).toBe(1);
      // refDate がより過去の場合は負値を 0 にクランプする
      expect(daysSince("2026-04-01", "2026-03-31")).toBe(0);
    });

    it("currentJSTMonthDateRangeはJST当月の初日と末日を返す", () => {
      vi.setSystemTime(new Date("2026-05-31T14:59:00Z"));
      expect(currentJSTMonthDateRange()).toEqual({
        start: "2026-05-01",
        end: "2026-05-31",
      });

      vi.setSystemTime(new Date("2026-05-31T15:00:00Z"));
      expect(currentJSTMonthDateRange()).toEqual({
        start: "2026-06-01",
        end: "2026-06-30",
      });
    });

    it("currentJSTYearMonthはJSTの年月を返す", () => {
      // UTC 15:00 未満はまだ JST 当日（5/31）
      vi.setSystemTime(new Date("2026-05-31T14:59:00Z"));
      expect(currentJSTYearMonth()).toBe("2026-05");

      // UTC 15:00 到達で JST は日付・月跨ぎ（6/1 00:00 JST）
      vi.setSystemTime(new Date("2026-05-31T15:00:00Z"));
      expect(currentJSTYearMonth()).toBe("2026-06");
    });
  });

  describe("日付専用の期限超過判定", () => {
    beforeEach(() => {
      vi.useFakeTimers();
      vi.setSystemTime(new Date("2026-07-23T15:30:00Z"));
    });

    afterEach(() => {
      vi.useRealTimers();
    });

    it.each([
      ["過去日", "2026-07-23", true],
      ["JST当日", "2026-07-24", false],
      ["未来日", "2026-07-25", false],
      ["空文字", "", false],
      ["非形式", "2026/07/23", false],
    ])("%sを厳密な過去日として判定する", (_label, date, expected) => {
      expect(isPastJSTDate(date)).toBe(expected);
    });
  });
});
