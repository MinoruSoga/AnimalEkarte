import { act, renderHook } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";

import { useSidePeekDirty } from "./use-side-peek-dirty";

const CONFIRM_MESSAGE = "未保存の変更があります。破棄してよろしいですか?";

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

describe("useSidePeekDirty", () => {
  it("公開API・dirty ref・callback参照を状態変更後も維持する", () => {
    const { result } = renderHook(() => useSidePeekDirty());
    const initialDirtyRef = result.current.isDirtyRef;
    const initialMarkDirty = result.current.markDirty;
    const initialMarkClean = result.current.markClean;
    const initialConfirmDiscard = result.current.confirmDiscard;

    expect(Object.keys(result.current).sort()).toEqual([
      "confirmDiscard",
      "isDirty",
      "isDirtyRef",
      "markClean",
      "markDirty",
    ]);
    expect(result.current.isDirty).toBe(false);
    expect(result.current.isDirtyRef.current).toBe(false);

    act(() => {
      result.current.markDirty();
    });

    expect(result.current.isDirty).toBe(true);
    expect(result.current.isDirtyRef).toBe(initialDirtyRef);
    expect(result.current.isDirtyRef.current).toBe(true);
    expect(result.current.markDirty).toBe(initialMarkDirty);
    expect(result.current.markClean).toBe(initialMarkClean);
    expect(result.current.confirmDiscard).toBe(initialConfirmDiscard);

    act(() => {
      result.current.markClean();
    });

    expect(result.current.isDirty).toBe(false);
    expect(result.current.isDirtyRef.current).toBe(false);
    expect(result.current.markDirty).toBe(initialMarkDirty);
    expect(result.current.markClean).toBe(initialMarkClean);
    expect(result.current.confirmDiscard).toBe(initialConfirmDiscard);
  });

  it("dirty化でbeforeunloadを登録し、unmountで解除する", () => {
    const addEventListenerSpy = vi.spyOn(window, "addEventListener");
    const removeEventListenerSpy = vi.spyOn(window, "removeEventListener");
    const { result, unmount } = renderHook(() => useSidePeekDirty());

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

    unmount();

    expect(removeEventListenerSpy).toHaveBeenCalledWith(
      "beforeunload",
      listener,
    );
  });

  it("cleanなら確認なしで破棄を許可する", () => {
    const confirmSpy = vi.spyOn(window, "confirm");
    const { result } = renderHook(() => useSidePeekDirty());

    expect(result.current.confirmDiscard()).toBe(true);
    expect(confirmSpy).not.toHaveBeenCalled();
  });

  it("dirtyで確認をキャンセルするとdirty状態を維持する", () => {
    const confirmSpy = vi.spyOn(window, "confirm").mockReturnValue(false);
    const { result } = renderHook(() => useSidePeekDirty());
    const confirmDiscardBeforeDirty = result.current.confirmDiscard;

    act(() => {
      result.current.markDirty();
    });

    expect(confirmDiscardBeforeDirty()).toBe(false);
    expect(confirmSpy).toHaveBeenCalledWith(CONFIRM_MESSAGE);
    expect(result.current.isDirty).toBe(true);
    expect(result.current.isDirtyRef.current).toBe(true);
  });

  it("dirtyで確認を承認するとclean化して破棄を許可する", () => {
    const confirmSpy = vi.spyOn(window, "confirm").mockReturnValue(true);
    const { result } = renderHook(() => useSidePeekDirty());
    const confirmDiscardBeforeDirty = result.current.confirmDiscard;
    let confirmed = false;

    act(() => {
      result.current.markDirty();
    });

    act(() => {
      confirmed = confirmDiscardBeforeDirty();
    });

    expect(confirmed).toBe(true);
    expect(confirmSpy).toHaveBeenCalledWith(CONFIRM_MESSAGE);
    expect(result.current.isDirty).toBe(false);
    expect(result.current.isDirtyRef.current).toBe(false);
  });
});
