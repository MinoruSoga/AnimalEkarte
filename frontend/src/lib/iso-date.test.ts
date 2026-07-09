import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { todayISODate, addDaysISO } from "./iso-date";

beforeEach(() => {
  vi.useFakeTimers();
});

afterEach(() => {
  vi.useRealTimers();
});

describe("todayISODate / addDaysISO", () => {
  it("A: vi.setSystemTime で固定した日付を yyyy-mm-dd で返す", () => {
    vi.setSystemTime(new Date("2026-05-06T03:00:00.000Z")); // 2026-05-06 12:00 JST
    expect(todayISODate()).toBe("2026-05-06");
  });

  it("B: addDaysISO(+30) — 月境界またぎ", () => {
    expect(addDaysISO("2026-05-06", 30)).toBe("2026-06-05");
  });

  it("C: addDaysISO(-1) — 前日", () => {
    expect(addDaysISO("2026-05-06", -1)).toBe("2026-05-05");
  });

  it("D: addDaysISO(+1) — 月末 → 翌月1日", () => {
    expect(addDaysISO("2026-01-31", 1)).toBe("2026-02-01");
  });
});
