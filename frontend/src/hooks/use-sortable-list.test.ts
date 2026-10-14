import { act, renderHook } from "@testing-library/react";
import type { DragEndEvent } from "@dnd-kit/core";
import { describe, expect, it, vi } from "vitest";
import { useSortableList } from "./use-sortable-list";

const items = Array.from({ length: 101 }, (_, index) => ({ id: String(index + 1) }));
const event: DragEndEvent = {
  activatorEvent: new Event("pointerup"),
  active: {
    id: "51",
    data: { current: {} },
    rect: { current: { initial: null, translated: null } },
  },
  over: {
    id: "52",
    data: { current: {} },
    disabled: false,
    rect: { top: 0, left: 0, bottom: 1, right: 1, width: 1, height: 1 },
  },
  delta: { x: 0, y: 0 },
  collisions: null,
};

describe("useSortableList global page-independent order", () => {
  it("moves global 51 over 52 and returns all 101 IDs and ordered items", () => {
    const onReorder = vi.fn();
    const { result } = renderHook(() => useSortableList({ items, onReorder }));
    act(() => result.current.handleDragEnd(event));
    const expected = [
      ...items.slice(0, 50).map((item) => item.id),
      "52",
      "51",
      ...items.slice(52).map((item) => item.id),
    ];
    expect(onReorder).toHaveBeenCalledExactlyOnceWith(expected);
    expect(result.current.orderedItems.map((item) => item.id)).toEqual(expected);
    expect(expected).toHaveLength(101);
  });
  it("cancel clears active state without a reorder callback", () => {
    const onReorder = vi.fn();
    const { result } = renderHook(() => useSortableList({ items, onReorder }));
    act(() => result.current.handleDragStart(event));
    expect(result.current.activeId).toBe("51");
    act(() => result.current.handleDragCancel());
    expect(result.current.activeId).toBeNull();
    expect(onReorder).not.toHaveBeenCalled();
    expect(result.current.orderedItems).toEqual(items);
  });
});
