import { renderHook, waitFor, act } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { useFetchState } from "./use-fetch-state";

describe("useFetchState（R-F23: liff/line-reserve 共通の GET-on-mount フェッチ状態管理フック）", () => {
  it("成功時は loading→data の順で状態が遷移する", async () => {
    const fetcher = vi.fn().mockResolvedValue({ id: 1 });

    const { result } = renderHook(() => useFetchState(fetcher, "テスト取得"));

    expect(result.current.loading).toBe(true);
    expect(result.current.data).toBeNull();

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.data).toEqual({ id: 1 });
    expect(result.current.error).toBeNull();
    expect(fetcher).toHaveBeenCalledTimes(1);
  });

  it("失敗時は resolveFetchError の結果を error にセットする", async () => {
    const err = Object.assign(new Error("boom"), { status: 500 });
    const fetcher = vi.fn().mockRejectedValue(err);

    const { result } = renderHook(() => useFetchState(fetcher, "テスト取得"));

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.data).toBeNull();
    expect(result.current.error).not.toBeNull();
    expect(result.current.error?.status).toBe(500);
    expect(result.current.error?.canRetry).toBe(true);
  });

  it("retry() を呼ぶと fetcher を再実行し、成功すれば data が更新される", async () => {
    const fetcher = vi
      .fn()
      .mockRejectedValueOnce(Object.assign(new Error("boom"), { status: 500 }))
      .mockResolvedValueOnce({ id: 2 });

    const { result } = renderHook(() => useFetchState(fetcher, "テスト取得"));

    await waitFor(() => {
      expect(result.current.error).not.toBeNull();
    });

    act(() => {
      result.current.retry();
    });

    expect(result.current.loading).toBe(true);

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.data).toEqual({ id: 2 });
    expect(result.current.error).toBeNull();
    expect(fetcher).toHaveBeenCalledTimes(2);
  });

  it("fetcher の identity が変わると再フェッチする（依存値変更の代替として使う）", async () => {
    const fetcher1 = vi.fn().mockResolvedValue({ id: 1 });
    const fetcher2 = vi.fn().mockResolvedValue({ id: 2 });

    const { result, rerender } = renderHook(
      ({ fetcher }: { fetcher: () => Promise<{ id: number }> }) =>
        useFetchState(fetcher, "テスト取得"),
      { initialProps: { fetcher: fetcher1 } },
    );

    await waitFor(() => {
      expect(result.current.data).toEqual({ id: 1 });
    });

    rerender({ fetcher: fetcher2 });

    await waitFor(() => {
      expect(result.current.data).toEqual({ id: 2 });
    });

    expect(fetcher1).toHaveBeenCalledTimes(1);
    expect(fetcher2).toHaveBeenCalledTimes(1);
  });

  it("setData でローカルな楽観的更新ができる（一覧の一部書き換え用途）", async () => {
    const fetcher = vi.fn().mockResolvedValue([{ id: 1, status: "confirmed" }]);

    const { result } = renderHook(() => useFetchState(fetcher, "テスト取得"));

    await waitFor(() => {
      expect(result.current.data).not.toBeNull();
    });

    act(() => {
      result.current.setData((prev) =>
        (prev ?? []).map((item) => ({ ...item, status: "cancelled" })),
      );
    });

    expect(result.current.data).toEqual([{ id: 1, status: "cancelled" }]);
  });
});
