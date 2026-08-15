import { fireEvent, render, screen } from "@testing-library/react";
import { useSortable } from "@dnd-kit/sortable";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { SortableDataTableRow } from "./SortableDataTableRow";

vi.mock("@dnd-kit/sortable", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@dnd-kit/sortable")>();
  return {
    ...actual,
    useSortable: vi.fn(),
  };
});

const mockUseSortable = vi.mocked(useSortable);
const onPointerDown = vi.fn();
const onKeyDown = vi.fn();
const setNodeRef = vi.fn();
const setActivatorNodeRef = vi.fn();
const setDroppableNodeRef = vi.fn();
const setDraggableNodeRef = vi.fn();

beforeEach(() => {
  vi.clearAllMocks();
  mockUseSortable.mockReturnValue({
    active: null,
    activeIndex: -1,
    attributes: {
      role: "button",
      tabIndex: 0,
      "aria-disabled": false,
      "aria-pressed": undefined,
      "aria-roledescription": "sortable",
      "aria-describedby": "DndDescribedBy-0",
    },
    data: {
      sortable: {
        containerId: "test-container",
        items: ["row-1"],
        index: 0,
      },
    },
    rect: { current: null },
    index: 0,
    newIndex: 0,
    items: ["row-1"],
    isOver: false,
    isSorting: false,
    isDragging: false,
    listeners: { onPointerDown, onKeyDown },
    node: { current: null },
    overIndex: -1,
    over: null,
    setNodeRef,
    setActivatorNodeRef,
    setDroppableNodeRef,
    setDraggableNodeRef,
    transform: null,
    transition: undefined,
  } satisfies ReturnType<typeof useSortable>);
});

function renderSortableRow(onLegacyRowClick = vi.fn(), dragDisabled = false) {
  const interactionProps = {
    dragLabel: "順番を変更: 第1診察室 ID row-1",
    dragDisabled,
    onClick: onLegacyRowClick,
  };
  const result = render(
    <table>
      <tbody>
        <SortableDataTableRow id="row-1" {...interactionProps}>
          <td>第1診察室</td>
        </SortableDataTableRow>
      </tbody>
    </table>,
  );

  return { ...result, onLegacyRowClick };
}

describe("SortableDataTableRow drag handle accessibility", () => {
  it("useSortable attributes/listeners を固有名付き44px native buttonにだけ渡す", () => {
    renderSortableRow();

    const dragHandle = screen.getByRole("button", {
      name: "順番を変更: 第1診察室 ID row-1",
    });
    expect(dragHandle.tagName).toBe("BUTTON");
    expect(dragHandle).toHaveAttribute("type", "button");
    expect(dragHandle).toHaveClass("min-h-11", "min-w-11");
    expect(dragHandle).toHaveAttribute("tabindex", "0");
    expect(dragHandle).toHaveAttribute("aria-roledescription", "sortable");
    expect(dragHandle).toHaveAttribute("aria-describedby", "DndDescribedBy-0");
    expect(setActivatorNodeRef).toHaveBeenCalledWith(dragHandle);
    expect(setNodeRef).toHaveBeenCalledWith(dragHandle.closest("tr"));

    fireEvent.pointerDown(dragHandle);
    fireEvent.keyDown(dragHandle, { key: "Enter" });

    expect(onPointerDown).toHaveBeenCalledOnce();
    expect(onKeyDown).toHaveBeenCalledOnce();
    expect(mockUseSortable).toHaveBeenCalledWith({ id: "row-1" });
  });

  it("tr は drag/activation semantics と legacy row click を持たない", () => {
    const { container, onLegacyRowClick } = renderSortableRow();
    const row = container.querySelector("tr");

    expect(row).not.toBeNull();
    expect(row).not.toHaveAttribute("role");
    expect(row).not.toHaveAttribute("tabindex");
    expect(row).not.toHaveAttribute("aria-roledescription");
    expect(row).not.toHaveAttribute("onclick");

    fireEvent.click(screen.getByText("第1診察室"));
    expect(onLegacyRowClick).not.toHaveBeenCalled();
  });

  it("dragDisabled 時はnative handleとuseSortableをともに無効化する", () => {
    renderSortableRow(vi.fn(), true);

    const dragHandle = screen.getByRole("button", {
      name: "順番を変更: 第1診察室 ID row-1",
    });
    expect(dragHandle).toBeDisabled();
    expect(mockUseSortable).toHaveBeenCalledWith({ id: "row-1", disabled: true });
  });
});
