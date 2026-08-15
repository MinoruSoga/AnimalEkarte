import { act, renderHook } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";

import { useUnsavedChanges } from "./use-unsaved-changes";

function requireEventListener(
  listener: EventListenerOrEventListenerObject | undefined,
): EventListener {
  if (typeof listener !== "function") {
    throw new Error("beforeunload listener was not registered");
  }
  return listener;
}

afterEach(() => {
  vi.restoreAllMocks();
});

describe("useUnsavedChanges", () => {
  it("公開APIとcallback参照をdirty状態の変更後も維持する", () => {
    const { result } = renderHook(() => useUnsavedChanges());
    const initialMarkDirty = result.current.markDirty;
    const initialMarkClean = result.current.markClean;

    expect(Object.keys(result.current).sort()).toEqual([
      "isDirty",
      "markClean",
      "markDirty",
    ]);
    expect(result.current.isDirty).toBe(false);

    act(() => {
      result.current.markDirty();
    });

    expect(result.current.isDirty).toBe(true);
    expect(result.current.markDirty).toBe(initialMarkDirty);
    expect(result.current.markClean).toBe(initialMarkClean);

    act(() => {
      result.current.markClean();
    });

    expect(result.current.isDirty).toBe(false);
    expect(result.current.markDirty).toBe(initialMarkDirty);
    expect(result.current.markClean).toBe(initialMarkClean);
  });

  it("dirty時だけbeforeunloadを登録し、イベント処理後にclean化で解除する", () => {
    const addEventListenerSpy = vi.spyOn(window, "addEventListener");
    const removeEventListenerSpy = vi.spyOn(window, "removeEventListener");
    const { result } = renderHook(() => useUnsavedChanges());

    expect(
      addEventListenerSpy.mock.calls.filter(
        ([eventName]) => eventName === "beforeunload",
      ),
    ).toHaveLength(0);

    act(() => {
      result.current.markDirty();
    });

    const beforeUnloadCall = addEventListenerSpy.mock.calls.find(
      ([eventName]) => eventName === "beforeunload",
    );
    const listener = requireEventListener(beforeUnloadCall?.[1]);
    const event: BeforeUnloadEvent = new Event("beforeunload", {
      cancelable: true,
    });
    Object.defineProperty(event, "returnValue", {
      configurable: true,
      value: "unchanged",
      writable: true,
    });
    const preventDefaultSpy = vi.spyOn(event, "preventDefault");

    listener(event);

    expect(preventDefaultSpy).toHaveBeenCalledOnce();
    expect(event.defaultPrevented).toBe(true);
    expect(event.returnValue).toBe("");

    act(() => {
      result.current.markClean();
    });

    expect(removeEventListenerSpy).toHaveBeenCalledWith(
      "beforeunload",
      listener,
    );
  });

  it("dirtyのままunmountするとbeforeunloadを解除する", () => {
    const addEventListenerSpy = vi.spyOn(window, "addEventListener");
    const removeEventListenerSpy = vi.spyOn(window, "removeEventListener");
    const { result, unmount } = renderHook(() => useUnsavedChanges());

    act(() => {
      result.current.markDirty();
    });

    const beforeUnloadCall = addEventListenerSpy.mock.calls.find(
      ([eventName]) => eventName === "beforeunload",
    );
    const listener = requireEventListener(beforeUnloadCall?.[1]);

    unmount();

    expect(removeEventListenerSpy).toHaveBeenCalledWith(
      "beforeunload",
      listener,
    );
  });
});
