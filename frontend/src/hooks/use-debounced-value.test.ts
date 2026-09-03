import { act, renderHook } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { useDebouncedValue } from "./use-debounced-value";

describe("useDebouncedValue", () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("初回は遅延せず初期値をそのまま返す", () => {
    const { result } = renderHook(() => useDebouncedValue("初期", 300));

    expect(result.current).toBe("初期");
  });

  it("値の変更直後は反映せず、遅延時間の経過後に反映する", () => {
    const { result, rerender } = renderHook(({ value }) => useDebouncedValue(value, 300), {
      initialProps: { value: "あ" },
    });

    rerender({ value: "あい" });
    expect(result.current).toBe("あ");

    act(() => {
      vi.advanceTimersByTime(299);
    });
    expect(result.current).toBe("あ");

    act(() => {
      vi.advanceTimersByTime(1);
    });
    expect(result.current).toBe("あい");
  });

  // 入力中は1文字ごとにfetchが飛ばないことの担保。
  // 連続入力ではタイマーが張り直され、最後の値だけが1回反映される。
  it("連続入力では中間値を飛ばし、最後の値だけを1回反映する", () => {
    const { result, rerender } = renderHook(({ value }) => useDebouncedValue(value, 300), {
      initialProps: { value: "" },
    });

    rerender({ value: "も" });
    act(() => {
      vi.advanceTimersByTime(200);
    });
    rerender({ value: "もも" });
    act(() => {
      vi.advanceTimersByTime(200);
    });

    // 累計400msだが、最後の入力からは200msしか経っていないので未反映。
    expect(result.current).toBe("");

    act(() => {
      vi.advanceTimersByTime(100);
    });
    expect(result.current).toBe("もも");
  });

  it("オブジェクトなど非プリミティブでも遅延して反映する", () => {
    const first = { search: "", species: "" };
    const second = { search: "もも", species: "3" };
    const { result, rerender } = renderHook(({ value }) => useDebouncedValue(value, 300), {
      initialProps: { value: first },
    });

    rerender({ value: second });
    expect(result.current).toBe(first);

    act(() => {
      vi.advanceTimersByTime(300);
    });
    expect(result.current).toBe(second);
  });
});
