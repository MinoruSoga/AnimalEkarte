import { act, renderHook } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { todayJSTISO } from "@/lib/jst-date";

import { useTodayJst } from "./use-today-jst";

/**
 * Codex PR606 #13: 期限超過 tentative の JST 暦日境界は、ページが開かれたまま
 * 日付を跨いでも更新される必要がある。useTodayJst の interval tick で
 * todayJSTISO() が再評価されることを固定する。
 */
describe("useTodayJst (Codex PR606 #13)", () => {
  beforeEach(() => {
    vi.useFakeTimers();
    // 2026-06-15 23:59:30 JST = 14:59:30 UTC
    vi.setSystemTime(new Date("2026-06-15T14:59:30Z"));
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("初期値は今日の JST 暦日", () => {
    const { result } = renderHook(() => useTodayJst());
    expect(result.current).toBe("2026-06-15");
    expect(result.current).toBe(todayJSTISO());
  });

  it("JST 暦日を跨ぐと tick 後に新しい日付へ更新される", () => {
    const { result } = renderHook(() => useTodayJst());
    expect(result.current).toBe("2026-06-15");

    // 90秒進めて 2026-06-16 00:01 JST を跨ぐ
    act(() => {
      vi.setSystemTime(new Date("2026-06-15T15:01:00Z"));
      vi.advanceTimersByTime(61_000);
    });

    expect(result.current).toBe("2026-06-16");
  });

  it("同一暦日内の tick では値が変わらない（同一値 setState は bailout）", () => {
    // 同日中（JST 11:00）に固定してから tick させる
    vi.setSystemTime(new Date("2026-06-15T02:00:00Z"));
    const { result } = renderHook(() => useTodayJst());
    const before = result.current;

    act(() => {
      vi.advanceTimersByTime(120_000);
    });

    expect(result.current).toBe(before);
  });
});
