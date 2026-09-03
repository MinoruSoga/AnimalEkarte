import { describe, it, expect } from "vitest";
import {
  parseLocalDate,
  formatIso,
  formatDisplay,
  formatShort,
  SINGLE_CALENDAR_CLASSES,
  RANGE_CALENDAR_CLASSES,
} from "./DatePickerModel";

// FE4-8 特性テスト（RED→GREEN 先行）: DatePickerModel を formatJapaneseDate /
// formatJSTWallDate へ委譲化する前に、現実装の入出力を固定する。
// 委譲化後もこのテストが GREEN のまま保たれることで出力が 1 文字も変わっていないことを保証する。
describe("DatePickerModel (FE4-8 特性テスト)", () => {
  describe("parseLocalDate", () => {
    it("YYYY-MM-DD をローカル正午の Date に変換する", () => {
      const date = parseLocalDate("2026-07-11");
      expect(date).toBeInstanceOf(Date);
      expect(date?.getFullYear()).toBe(2026);
      expect(date?.getMonth()).toBe(6);
      expect(date?.getDate()).toBe(11);
      expect(date?.getHours()).toBe(12);
      expect(date?.getMinutes()).toBe(0);
      expect(date?.getSeconds()).toBe(0);
    });

    it("空文字は undefined を返す", () => {
      expect(parseLocalDate("")).toBeUndefined();
    });
  });

  describe("formatIso", () => {
    it("Date を YYYY-MM-DD 形式にする", () => {
      expect(formatIso(new Date(2026, 0, 5, 12, 0, 0))).toBe("2026-01-05");
    });

    it("月・日を 2 桁 0 パディングする", () => {
      expect(formatIso(new Date(2026, 11, 31, 12, 0, 0))).toBe("2026-12-31");
    });
  });

  describe("formatDisplay", () => {
    it("Date を「Y年M月D日（曜）」形式にする", () => {
      // 2026-07-11 は土曜日
      expect(formatDisplay(new Date(2026, 6, 11, 12, 0, 0))).toBe("2026年7月11日（土）");
    });

    it("月・日は 0 パディングしない", () => {
      // 2026-01-05 は月曜日
      expect(formatDisplay(new Date(2026, 0, 5, 12, 0, 0))).toBe("2026年1月5日（月）");
    });
  });

  describe("formatShort", () => {
    it("Date を YYYY/M/D 形式にする（0 パディングなし）", () => {
      expect(formatShort(new Date(2026, 0, 5, 12, 0, 0))).toBe("2026/1/5");
    });
  });
});

// FE-RC-106: selected は完成形静的トークンのみ。runtime 合成（hover: + ${）は Tailwind v4 が拾えない。
// 禁止部分文字列は結合して書く（このファイル自体が監査 / rg の対象になるため）。
const RUNTIME_HOVER_SYNTHESIS = ["hover:", "${"].join("");

describe("DatePickerModel calendar selected tokens (FE-RC-106)", () => {
  it.each([
    ["SINGLE_CALENDAR_CLASSES", SINGLE_CALENDAR_CLASSES.selected],
    ["RANGE_CALENDAR_CLASSES", RANGE_CALENDAR_CLASSES.selected],
  ] as const)("%s.selected は完成形 hover/focus brand トークンを使い runtime 合成しない", (_name, selected) => {
    expect(selected).toContain("hover:bg-[#027078]");
    expect(selected).toContain("focus:bg-[#027078]");
    expect(selected).not.toContain(RUNTIME_HOVER_SYNTHESIS);
  });
});
