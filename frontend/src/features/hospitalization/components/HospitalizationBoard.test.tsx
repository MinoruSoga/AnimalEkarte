import { act, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import type { ReactNode } from "react";

import type { Hospitalization, MasterItem } from "@/types";
import { KeyboardSensor, PointerSensor, useSensor } from "@dnd-kit/core";
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
  KeyboardSensor: function KeyboardSensor() {},
  closestCorners: vi.fn(),
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
    expect(screen.getByText("犬 ポチ").closest("[data-slot='card']")).not.toHaveAttribute(
      "data-drag-listener",
    );
    expect(screen.getByText("犬 ポチ").closest("[data-slot='card']")).not.toHaveAttribute(
      "data-drag-attributes",
    );
    expect(
      screen
        .getByText("犬 ポチ")
        .closest("[data-slot='card']")
        ?.querySelector(".lucide-grip-vertical"),
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
    expect(screen.getByText("犬 ポチ").closest("[data-slot='card']")).toHaveAttribute(
      "data-drag-listener",
      "enabled",
    );
    expect(screen.getByText("犬 ポチ").closest("[data-slot='card']")).toHaveAttribute(
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
    expect(useSensor).toHaveBeenCalledWith(PointerSensor, expect.anything());
    expect(useSensor).toHaveBeenCalledWith(
      KeyboardSensor,
      expect.objectContaining({ coordinateGetter: expect.any(Function) }),
    );
  });

  // FE-RC-044: Card 自体の onClick は非フォーカス可能・キーボード操作不可のため削除し、
  // 詳細への遷移は aria-label 付き button に一本化する（冗長な操作導線の排除）。
  it("occupied cardの本体クリックでは遷移せず、詳細buttonのクリックでのみ遷移する", async () => {
    const user = userEvent.setup();
    const onNavigateToForm = vi.fn();

    render(
      <HospitalizationBoard
        cages={cages}
        hospitalizations={[occupied]}
        onNavigateToForm={onNavigateToForm}
        onMovePet={vi.fn()}
        canCreate
        canEdit
      />,
    );

    const card = screen.getByText("犬 ポチ").closest("[data-slot='card']") as HTMLElement;
    await user.click(card);
    expect(onNavigateToForm).not.toHaveBeenCalled();

    const detailButton = screen.getByRole("button", { name: "ポチの詳細" });
    await user.click(detailButton);
    expect(onNavigateToForm).toHaveBeenCalledWith(occupied.id);
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
    expect(
      screen.queryByRole("button", { name: /空き枠に入院・ホテルを登録/ }),
    ).not.toBeInTheDocument();
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

// EMR-316: ケージ未設定（cage_id=NULL）の入院を「選択なし」で登録できるようにしたため、
// ボードから見えなくならないよう「ケージ未割当」欄に表示する。
describe("HospitalizationBoard cageless hospitalizations (EMR-316)", () => {
  const cageless: Hospitalization = {
    id: "hospitalization-cageless",
    cageId: undefined,
    petName: "ミケ",
    ownerName: "飼主B",
    species: "猫",
    hospitalizationType: "入院",
    petIsDeceased: false,
  } as unknown as Hospitalization;

  it("ケージ未設定の入院を「ケージ未割当」欄に表示し、ケージカードには紛れ込ませない", () => {
    render(
      <HospitalizationBoard
        cages={cages}
        hospitalizations={[occupied, cageless]}
        onNavigateToForm={vi.fn()}
        onMovePet={vi.fn()}
        canCreate
        canEdit
      />,
    );

    const section = screen.getByRole("region", { name: "ケージ未割当" });
    expect(section).toHaveTextContent("ミケ");
    expect(section).toHaveTextContent("飼主B");
    // ケージ設定済みの入院は従来どおりケージカード側にだけ出る。
    expect(section).not.toHaveTextContent("ポチ");
    expect(screen.getAllByText(/ポチ/)).toHaveLength(1);
    // 未割当の入院は「空き」ケージを増減させない（cage-a-2 だけが空き）。
    expect(screen.getAllByText("空き")).toHaveLength(1);
  });

  it("ケージ未設定の入院が無ければ「ケージ未割当」欄を出さない", () => {
    render(
      <HospitalizationBoard
        cages={cages}
        hospitalizations={[occupied]}
        onNavigateToForm={vi.fn()}
        onMovePet={vi.fn()}
        canCreate
        canEdit
      />,
    );

    expect(screen.queryByRole("region", { name: "ケージ未割当" })).toBeNull();
  });

  it("未割当カードの詳細ボタンで入院詳細へ遷移する", async () => {
    const user = userEvent.setup();
    const onNavigateToForm = vi.fn();
    render(
      <HospitalizationBoard
        cages={cages}
        hospitalizations={[cageless]}
        onNavigateToForm={onNavigateToForm}
        onMovePet={vi.fn()}
        canCreate
        canEdit
      />,
    );

    await user.click(screen.getByRole("button", { name: "ミケの詳細" }));

    expect(onNavigateToForm).toHaveBeenCalledWith("hospitalization-cageless");
  });

  it("未割当カードはドラッグ対象にしない（本票は表示と遷移のみ）", () => {
    dndHarness.useDraggable.mockClear();
    render(
      <HospitalizationBoard
        cages={[]}
        hospitalizations={[cageless]}
        onNavigateToForm={vi.fn()}
        onMovePet={vi.fn()}
        canCreate
        canEdit
      />,
    );

    expect(dndHarness.useDraggable).not.toHaveBeenCalled();
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
