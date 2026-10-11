import { createElement, StrictMode } from "react";
import { renderToString } from "react-dom/server";
import { act, cleanup, renderHook } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { todayJSTISO } from "@/lib/jst-date";

import { useTodayJst } from "./use-today-jst";

/**
 * Codex PR606 #13: 期限超過 tentative の JST 暦日境界は、ページが開かれたまま
 * 日付を跨いでも更新される必要がある。useTodayJst の日界 timer で
 * todayJSTISO() が再評価されることを固定する。
 */
describe("useTodayJst (Codex PR606 #13)", () => {
  beforeEach(() => {
    vi.useFakeTimers();
    // 2026-06-15 23:59:30 JST = 14:59:30 UTC
    vi.setSystemTime(new Date("2026-06-15T14:59:30Z"));
  });

  afterEach(() => {
    cleanup();
    vi.restoreAllMocks();
    vi.useRealTimers();
  });

  it("初期値は今日の JST 暦日", () => {
    const { result } = renderHook(() => useTodayJst());
    expect(result.current).toBe("2026-06-15");
    expect(result.current).toBe(todayJSTISO());
  });

  it("JST 暦日を跨ぐと timer 後に新しい日付へ更新される", () => {
    const { result } = renderHook(() => useTodayJst());
    expect(result.current).toBe("2026-06-15");

    // 90秒進めて 2026-06-16 00:01 JST を跨ぐ
    act(() => {
      vi.setSystemTime(new Date("2026-06-15T15:01:00Z"));
      vi.advanceTimersByTime(61_000);
    });

    expect(result.current).toBe("2026-06-16");
  });

  it("同一暦日内では snapshot の日付が変わらない", () => {
    // 同日中（JST 11:00）に固定してから時刻を進める
    vi.setSystemTime(new Date("2026-06-15T02:00:00Z"));
    const { result } = renderHook(() => useTodayJst());
    const before = result.current;

    act(() => {
      vi.advanceTimersByTime(120_000);
    });

    expect(result.current).toBe(before);
  });
});

describe("shared reactive JST clock", () => {
  beforeEach(() => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-10-31T14:59:30Z"));
  });
  afterEach(() => {
    cleanup();
    vi.restoreAllMocks();
    vi.useRealTimers();
  });
  it.each([
    ["2026-10-31T14:59:30Z", "2026-11-01"],
    ["2026-12-31T14:59:30Z", "2027-01-01"],
    ["2028-02-28T14:59:30Z", "2028-02-29"],
    ["2028-02-29T14:59:30Z", "2028-03-01"],
  ])("updates at the exact JST boundary after %s", (start, expected) => {
    vi.setSystemTime(new Date(start));
    const { result } = renderHook(() => useTodayJst());
    act(() => vi.advanceTimersByTime(29_999));
    expect(result.current).not.toBe(expected);
    act(() => vi.advanceTimersByTime(1));
    expect(result.current).toBe(expected);
  });
  it("uses one timer for multiple consumers and cleans up after the last unsubscribe", () => {
    const first = renderHook(() => useTodayJst());
    const second = renderHook(() => useTodayJst());
    expect(vi.getTimerCount()).toBe(1);
    first.unmount();
    expect(vi.getTimerCount()).toBe(1);
    second.unmount();
    expect(vi.getTimerCount()).toBe(0);
  });
  it.each(["focus", "pageshow"])(
    "refreshes after long sleep on %s without waiting for a timer",
    (event) => {
      const { result } = renderHook(() => useTodayJst());
      act(() => {
        vi.setSystemTime(new Date("2027-01-03T02:00:00Z"));
        window.dispatchEvent(new Event(event));
      });
      expect(result.current).toBe("2027-01-03");
      expect(vi.getTimerCount()).toBe(1);
    },
  );
  it("pauses while hidden and immediately catches up on visibility restoration", () => {
    const visibility = vi.spyOn(document, "visibilityState", "get").mockReturnValue("visible");
    const { result } = renderHook(() => useTodayJst());
    act(() => {
      visibility.mockReturnValue("hidden");
      document.dispatchEvent(new Event("visibilitychange"));
    });
    expect(vi.getTimerCount()).toBe(0);
    act(() => {
      vi.setSystemTime(new Date("2026-11-03T00:00:00Z"));
      visibility.mockReturnValue("visible");
      document.dispatchEvent(new Event("visibilitychange"));
    });
    expect(result.current).toBe("2026-11-03");
    expect(vi.getTimerCount()).toBe(1);
  });
  it("removes hidden-tab listeners after unsubscribe and starts fresh when remounted", () => {
    const visibility = vi.spyOn(document, "visibilityState", "get").mockReturnValue("hidden");
    const first = renderHook(() => useTodayJst());
    expect(vi.getTimerCount()).toBe(0);
    first.unmount();
    act(() => {
      vi.setSystemTime(new Date("2026-11-03T00:00:00Z"));
      visibility.mockReturnValue("visible");
      document.dispatchEvent(new Event("visibilitychange"));
    });
    expect(vi.getTimerCount()).toBe(0);
    const second = renderHook(() => useTodayJst());
    expect(second.result.current).toBe("2026-11-03");
    expect(vi.getTimerCount()).toBe(1);
  });
  it("does not multiply timers under StrictMode subscription replay", () => {
    const { unmount } = renderHook(() => useTodayJst(), {
      wrapper: ({ children }) => createElement(StrictMode, null, children),
    });
    expect(vi.getTimerCount()).toBe(1);
    unmount();
    expect(vi.getTimerCount()).toBe(0);
  });
  it("provides a stable current date for server rendering without browser timers", () => {
    function Clock() {
      return createElement("span", null, useTodayJst());
    }
    expect(renderToString(createElement(Clock))).toContain("2026-10-31");
    expect(vi.getTimerCount()).toBe(0);
  });
});
