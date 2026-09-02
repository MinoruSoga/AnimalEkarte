import { act, render, renderHook, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";

import { useSidePeekDirty } from "./use-side-peek-dirty";

const CONFIRM_TITLE = "未保存の変更があります";

function requireEventListener(
  listener: EventListenerOrEventListenerObject | undefined,
): EventListener {
  if (typeof listener !== "function") {
    throw new Error("beforeunload listener was not registered");
  }
  return listener;
}

function renderDiscardDialog(result: { current: ReturnType<typeof useSidePeekDirty> }) {
  const view = render(<>{result.current.discardDialog}</>);
  return {
    sync() {
      view.rerender(<>{result.current.discardDialog}</>);
    },
  };
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
    const initialRunWithDiscardCheck = result.current.runWithDiscardCheck;

    expect(Object.keys(result.current).sort()).toEqual([
      "confirmDiscard",
      "discardDialog",
      "isDirty",
      "isDirtyRef",
      "markClean",
      "markDirty",
      "runWithDiscardCheck",
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
    expect(result.current.runWithDiscardCheck).toBe(initialRunWithDiscardCheck);

    act(() => {
      result.current.markClean();
    });

    expect(result.current.isDirty).toBe(false);
    expect(result.current.isDirtyRef.current).toBe(false);
    expect(result.current.markDirty).toBe(initialMarkDirty);
    expect(result.current.markClean).toBe(initialMarkClean);
    expect(result.current.confirmDiscard).toBe(initialConfirmDiscard);
    expect(result.current.runWithDiscardCheck).toBe(initialRunWithDiscardCheck);
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
    const pending = vi.fn();
    const { result } = renderHook(() => useSidePeekDirty());
    const dialog = renderDiscardDialog(result);

    expect(result.current.confirmDiscard()).toBe(true);
    act(() => {
      result.current.runWithDiscardCheck(pending);
    });
    dialog.sync();

    expect(pending).toHaveBeenCalledTimes(1);
    expect(screen.queryByText(CONFIRM_TITLE)).not.toBeInTheDocument();
  });

  it("dirtyで確認をキャンセルするとdirty状態を維持し継続処理を捨てる", async () => {
    const user = userEvent.setup();
    const pending = vi.fn();
    const { result } = renderHook(() => useSidePeekDirty());
    const dialog = renderDiscardDialog(result);

    act(() => {
      result.current.markDirty();
    });
    act(() => {
      result.current.runWithDiscardCheck(pending);
    });
    dialog.sync();

    expect(screen.getByText(CONFIRM_TITLE)).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "キャンセル" }));
    dialog.sync();

    expect(pending).not.toHaveBeenCalled();
    expect(result.current.isDirty).toBe(true);
    expect(result.current.isDirtyRef.current).toBe(true);
    expect(screen.queryByText(CONFIRM_TITLE)).not.toBeInTheDocument();
  });

  it("dirtyで確認を承認するとclean化して継続処理を実行する", async () => {
    const user = userEvent.setup();
    const pending = vi.fn();
    const { result } = renderHook(() => useSidePeekDirty());
    const dialog = renderDiscardDialog(result);

    act(() => {
      result.current.markDirty();
    });
    act(() => {
      result.current.runWithDiscardCheck(pending);
    });
    dialog.sync();

    expect(screen.getByText(CONFIRM_TITLE)).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "確認" }));
    dialog.sync();

    expect(pending).toHaveBeenCalledTimes(1);
    expect(result.current.isDirty).toBe(false);
    expect(result.current.isDirtyRef.current).toBe(false);
  });

  it("dirtyなconfirmDiscardはダイアログを開いてfalseを返し継続処理を走らせない", async () => {
    const user = userEvent.setup();
    const pending = vi.fn();
    const { result } = renderHook(() => useSidePeekDirty());
    const confirmDiscardBeforeDirty = result.current.confirmDiscard;
    const dialog = renderDiscardDialog(result);
    let confirmed = true;

    act(() => {
      result.current.markDirty();
    });
    act(() => {
      confirmed = confirmDiscardBeforeDirty();
    });
    dialog.sync();

    expect(confirmed).toBe(false);
    expect(screen.getByText(CONFIRM_TITLE)).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "確認" }));
    dialog.sync();

    expect(pending).not.toHaveBeenCalled();
    expect(result.current.isDirty).toBe(false);
  });
});
