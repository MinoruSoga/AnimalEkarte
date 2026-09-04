import { describe, it, expect } from "vitest";
import { formatDate } from "./date";

// ─────────────────────────────────────────────────────────────
// formatDate
// ─────────────────────────────────────────────────────────────
describe("formatDate", () => {
  it("ISO日付文字列を YYYY/MM/DD にフォーマットする", () => {
    expect(formatDate("2026-03-25")).toBe("2026/03/25");
  });

  it("ISO datetime 文字列（T付き）を YYYY/MM/DD にフォーマットする", () => {
    expect(formatDate("2026-01-05T00:00:00Z")).toBe("2026/01/05");
  });

  it("月・日が 1桁のとき 0 パディングする", () => {
    expect(formatDate("2026-01-01")).toBe("2026/01/01");
  });

  it("12月31日を正しくフォーマットする", () => {
    expect(formatDate("2026-12-31")).toBe("2026/12/31");
  });

  it("undefined を渡すと '-' を返す", () => {
    expect(formatDate(undefined)).toBe("-");
  });

  it("null を渡すと '-' を返す", () => {
    expect(formatDate(null)).toBe("-");
  });

  it("空文字を渡すと '-' を返す", () => {
    expect(formatDate("")).toBe("-");
  });

  it("不正な日付文字列を渡すと '-' を返す", () => {
    expect(formatDate("not-a-date")).toBe("-");
  });
});
