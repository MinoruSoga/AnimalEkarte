import { act, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import type { ReactNode } from "react";

import type { Hospitalization, MasterItem } from "@/types";
import { HospitalizationBoard } from "./HospitalizationBoard";

const dndHarness = vi.hoisted(() => ({
  onDragEnd: undefined as ((event: unknown) => void) | undefined,
  useDraggable: vi.fn(() => ({
    attributes: { "data-drag-attributes": "enabled" },
    listeners: { "data-drag-listener": "enabled" },
    setNodeRef: vi.fn(),
    isDragging: false,
  })),
}));

vi.mock("@dnd-kit/core", () => ({
  DndContext: ({
    children,
    onDragEnd,
  }: {
    children: ReactNode;
    onDragEnd: (event: unknown) => void;
  }) => {
    dndHarness.onDragEnd = onDragEnd;
    return children;
  },
  PointerSensor: function PointerSensor() {},
  useSensor: vi.fn(() => ({})),
  useSensors: vi.fn(() => []),
  useDraggable: dndHarness.useDraggable,
  useDroppable: vi.fn(() => ({
    setNodeRef: vi.fn(),
    isOver: false,
  })),
}));

const cages: MasterItem[] = [
  {
    id: "cage-a-1",
    name: "ケージ1",
    category: "入院室A",
    price: 0,
    status: "active",
  },
  {
    id: "cage-a-2",
    name: "ケージ1",
    category: "入院室A",
    price: 0,
    status: "active",
  },
];

const occupied: Hospitalization = {
  id: "hospitalization-1",
  cageId: "cage-a-1",
  petName: "ポチ",
  ownerName: "飼主A",
  species: "犬",
  hospitalizationType: "入院",
  petIsDeceased: false,
} as Hospitalization;

describe("HospitalizationBoard empty cage actions", () => {
  it("同一エリアの同名ケージでも安定ID文脈を含む一意なaccessible nameを付ける", async () => {
    const user = userEvent.setup();
    const onNavigateToForm = vi.fn();

    render(
      <HospitalizationBoard
        cages={cages}
        hospitalizations={[]}
        onNavigateToForm={onNavigateToForm}
        onMovePet={vi.fn()}
        canCreate
        canEdit
      />,
    );

    const areaAButton = screen.getByRole("button", {
      name: "入院室A ケージ1（ケージID: cage-a-1）の空き枠に入院・ホテルを登録",
    });
    const secondAreaAButton = screen.getByRole("button", {
      name: "入院室A ケージ1（ケージID: cage-a-2）の空き枠に入院・ホテルを登録",
    });

    expect(areaAButton).toHaveAttribute(
      "title",
      "入院室A ケージ1（ケージID: cage-a-1）の空き枠に入院・ホテルを登録",
    );
    expect(secondAreaAButton).toHaveAttribute(
      "title",
      "入院室A ケージ1（ケージID: cage-a-2）の空き枠に入院・ホテルを登録",
    );

    await user.click(areaAButton);
    expect(onNavigateToForm).toHaveBeenCalledOnce();
    expect(onNavigateToForm).toHaveBeenCalledWith();
  });

  it("canEdit=falseではoccupied cardのdrag listenerとprogrammatic drag mutationを無効化する", () => {
    const onMovePet = vi.fn();

    render(
      <HospitalizationBoard
        cages={cages}
        hospitalizations={[occupied]}
        onNavigateToForm={vi.fn()}
        onMovePet={onMovePet}
        canCreate
        canEdit={false}
      />,
    );

    expect(dndHarness.useDraggable).toHaveBeenCalledWith(
      expect.objectContaining({ id: occupied.id, disabled: true }),
    );
    expect(screen.getByText("ポチ").closest("[data-slot='card']")).not.toHaveAttribute(
      "data-drag-listener",
    );
    expect(screen.getByText("ポチ").closest("[data-slot='card']")).not.toHaveAttribute(
      "data-drag-attributes",
    );
    expect(
      screen.getByText("ポチ").closest("[data-slot='card']")?.querySelector(".lucide-grip-vertical"),
    ).not.toBeInTheDocument();

    act(() => {
      dndHarness.onDragEnd?.({
        active: { data: { current: { hospitalizationId: occupied.id } } },
        over: { id: "cage-cage-a-2" },
      });
    });

    expect(onMovePet).not.toHaveBeenCalled();
  });

  it("canEdit=trueではoccupied cardをdragでき、drop先ケージへ移動を通知する", () => {
    const onMovePet = vi.fn();

    render(
      <HospitalizationBoard
        cages={cages}
        hospitalizations={[occupied]}
        onNavigateToForm={vi.fn()}
        onMovePet={onMovePet}
        canCreate
        canEdit
      />,
    );

    expect(dndHarness.useDraggable).toHaveBeenCalledWith(
      expect.objectContaining({ id: occupied.id, disabled: false }),
    );
    expect(screen.getByText("ポチ").closest("[data-slot='card']")).toHaveAttribute(
      "data-drag-listener",
      "enabled",
    );
    expect(screen.getByText("ポチ").closest("[data-slot='card']")).toHaveAttribute(
      "data-drag-attributes",
      "enabled",
    );

    act(() => {
      dndHarness.onDragEnd?.({
        active: { data: { current: { hospitalizationId: occupied.id } } },
        over: { id: "cage-cage-a-2" },
      });
    });

    expect(onMovePet).toHaveBeenCalledWith(occupied.id, "cage-a-2");
  });

  it("死亡ペットのoccupied cardは色に依存しない死亡文言を表示する", () => {
    render(
      <HospitalizationBoard
        cages={cages}
        hospitalizations={[{ ...occupied, petIsDeceased: true }]}
        onNavigateToForm={vi.fn()}
        onMovePet={vi.fn()}
        canCreate
        canEdit
      />,
    );

    expect(screen.getByText("死亡", { exact: true })).toBeVisible();
  });

  it("canCreate=falseでは空カード全体から新規導線へ遷移しない", async () => {
    const user = userEvent.setup();
    const onNavigateToForm = vi.fn();

    render(
      <HospitalizationBoard
        cages={[cages[0]]}
        hospitalizations={[]}
        onNavigateToForm={onNavigateToForm}
        onMovePet={vi.fn()}
        canCreate={false}
        canEdit
      />,
    );

    await user.click(screen.getByText("空き").closest("[data-slot='card']") as HTMLElement);

    expect(onNavigateToForm).not.toHaveBeenCalled();
    expect(screen.queryByRole("button", { name: /空き枠に入院・ホテルを登録/ })).not.toBeInTheDocument();
  });

  it("canCreate=trueなら空カード自体は非interactiveのまま、追加buttonをkeyboard操作できる", async () => {
    const user = userEvent.setup();
    const onNavigateToForm = vi.fn();

    render(
      <HospitalizationBoard
        cages={[cages[0]]}
        hospitalizations={[]}
        onNavigateToForm={onNavigateToForm}
        onMovePet={vi.fn()}
        canCreate
        canEdit={false}
      />,
    );

    const emptyCard = screen.getByText("空き").closest("[data-slot='card']") as HTMLElement;
    await user.click(emptyCard);

    expect(onNavigateToForm).not.toHaveBeenCalled();
    expect(emptyCard).not.toHaveAttribute("role");
    expect(emptyCard).not.toHaveAttribute("tabindex");

    const addButton = screen.getByRole("button", {
      name: "入院室A ケージ1（ケージID: cage-a-1）の空き枠に入院・ホテルを登録",
    });
    addButton.focus();
    await user.keyboard("{Enter}");

    expect(onNavigateToForm).toHaveBeenCalledOnce();
    expect(onNavigateToForm).toHaveBeenCalledWith();
  });
});

describe("HospitalizationBoard narrow layout (BUG-458)", () => {
  it("area container に min-w-[800px] を付けない", () => {
    const { container } = render(
      <HospitalizationBoard
        cages={cages}
        hospitalizations={[]}
        onNavigateToForm={vi.fn()}
        onMovePet={vi.fn()}
        canCreate
        canEdit
      />,
    );
    const hasMinWidthConstraint = Array.from(container.querySelectorAll("*")).some(
      (el) => typeof el.className === "string" && el.className.includes("min-w-[800px]"),
    );
    expect(hasMinWidthConstraint).toBe(false);
    expect(container.querySelector(".grid-cols-1")).not.toBeNull();
  });
});
