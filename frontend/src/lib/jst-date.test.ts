import { describe, expect, it } from "vitest";

import {
  buildJSTWallDateTime,
  formatJSTDate,
  formatJSTDateTimeLocal,
  formatJSTOffsetDateTime,
  formatJSTTime,
  jstDateStartISOString,
  jstDateTimeLocalToISOString,
  jstNowISOString,
  jstWallDateToISOString,
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

  it("instant を JST オフセット付き ISO 文字列にする", () => {
    expect(formatJSTOffsetDateTime("2026-06-01T00:45:00.123Z")).toBe("2026-06-01T09:45:00.123+09:00");
  });

  it("instant を datetime-local 用の JST 壁時計文字列にする", () => {
    expect(formatJSTDateTimeLocal("2026-06-01T00:45:00Z")).toBe("2026-06-01T09:45");
  });

  it("現在時刻も JST オフセット付き ISO 文字列を返す", () => {
    expect(jstNowISOString()).toMatch(/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{3}\+09:00$/);
  });
});
