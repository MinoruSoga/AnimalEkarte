import { act, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { ReactNode } from "react";

import type { Hospitalization, MasterItem } from "@/types";
import { KeyboardSensor, PointerSensor, useSensor } from "@dnd-kit/core";
import { HospitalizationBoard } from "./HospitalizationBoard";

const dndHarness = vi.hoisted(() => ({
  onDragEnd: undefined as ((event: unknown) => void) | undefined,
  active: null as unknown,
  useDraggable: vi.fn(() => ({
    attributes: { "data-drag-attributes": "enabled" },
    listeners: { "data-drag-listener": "enabled" },
    setNodeRef: vi.fn(),
    isDragging: false,
  })),
  useDroppable: vi.fn(() => ({
    setNodeRef: vi.fn(),
    isOver: false,
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
  useDroppable: dndHarness.useDroppable,
  useDndContext: vi.fn(() => ({ active: dndHarness.active })),
}));

beforeEach(() => {
  dndHarness.useDraggable.mockClear();
  dndHarness.useDroppable.mockClear();
  dndHarness.active = null;
});

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

// ─────────────────────────────────────────────────────────────
// EMR-306: ケージ一時連結 merge mode
// ─────────────────────────────────────────────────────────────

const mergeCages: MasterItem[] = [
  { id: "10", name: "ケージ10", category: "犬舎", price: 0, status: "active" },
  { id: "11", name: "ケージ11", category: "犬舎", price: 0, status: "active" },
  {
    id: "12",
    name: "ケージ12",
    category: "犬舎",
    price: 0,
    status: "active",
    mergedIntoCageId: "10",
  },
  { id: "13", name: "ケージ13", category: "犬舎", price: 0, status: "active" },
  { id: "14", name: "ケージ14", category: "犬舎", price: 0, status: "inactive" },
  { id: "15", name: "ケージ15", category: "犬舎", price: 0, status: "active" },
];

const occupiedOn13: Hospitalization = {
  id: "hospitalization-13",
  cageId: "13",
  petName: "タマ",
  ownerName: "飼主B",
  species: "猫",
  hospitalizationType: "入院",
  petIsDeceased: false,
} as Hospitalization;

function renderMergeBoard(overrides: Partial<Parameters<typeof HospitalizationBoard>[0]> = {}) {
  return render(
    <HospitalizationBoard
      cages={mergeCages}
      hospitalizations={[occupiedOn13]}
      onNavigateToForm={vi.fn()}
      onMovePet={vi.fn()}
      canCreate
      canEdit
      mergeMode
      canMerge
      onRequestMerge={vi.fn()}
      onRequestUnmerge={vi.fn()}
      {...overrides}
    />,
  );
}

describe("HospitalizationBoard EMR-306 merge mode", () => {
  it("merge mode ON: standalone cage は merge-cage- id で drag 可能、patient drag id は登録されない", () => {
    renderMergeBoard();

    // standalone & active & 非占有 の 11・15 が source になる
    for (const id of ["merge-cage-11", "merge-cage-15"]) {
      expect(dndHarness.useDraggable).toHaveBeenCalledWith(
        expect.objectContaining({ id, disabled: false }),
      );
    }
    expect(dndHarness.useDraggable).toHaveBeenCalledWith(
      expect.objectContaining({
        id: "merge-cage-11",
        data: { mergeCageId: "11" },
      }),
    );
    // 代表(10)・member(12)・占有(13)・inactive(14) は source 不可
    for (const id of ["merge-cage-10", "merge-cage-12", "merge-cage-13", "merge-cage-14"]) {
      expect(dndHarness.useDraggable).toHaveBeenCalledWith(
        expect.objectContaining({ id, disabled: true }),
      );
    }
    // pet-move 用の hospitalization id / empty- id は一切登録されない
    expect(dndHarness.useDraggable).not.toHaveBeenCalledWith(
      expect.objectContaining({ id: occupiedOn13.id }),
    );
    expect(dndHarness.useDraggable).not.toHaveBeenCalledWith(
      expect.objectContaining({ id: expect.stringMatching(/^empty-/) }),
    );
    // KeyboardSensor はそのまま（cage drag も keyboard-operable）
    expect(useSensor).toHaveBeenCalledWith(
      KeyboardSensor,
      expect.objectContaining({ coordinateGetter: expect.any(Function) }),
    );
  });

  it("merge mode ON: source cage の card に連結元 aria-label が付く（keyboard 操作の activator）", () => {
    renderMergeBoard();

    const sourceCard = screen.getByText("ケージ11").closest("[data-slot='card']") as HTMLElement;
    expect(sourceCard.getAttribute("aria-label")).toBe(
      "犬舎 ケージ11（ケージID: 11）を連結元としてドラッグ",
    );
    expect(sourceCard).toHaveAttribute("data-drag-listener", "enabled");
  });

  it("merge mode ON: eligible drop で onRequestMerge(dragged, target) を呼ぶ", () => {
    const onRequestMerge = vi.fn();
    renderMergeBoard({ onRequestMerge });

    // drag 11 → drop on 15（standalone cage。代表 cage への追加連結は BE が
    // 単一階層制約で 409 拒否するため target 不適格）
    act(() => {
      dndHarness.onDragEnd?.({
        active: { data: { current: { mergeCageId: "11" } } },
        over: { id: "cage-15" },
      });
    });
    expect(onRequestMerge).toHaveBeenCalledWith("11", "15");
  });

  it("merge mode ON: ineligible drop（member/self/占有 target, source 不適格, over なし）を拒否する", () => {
    const onRequestMerge = vi.fn();
    renderMergeBoard({ onRequestMerge });

    const drops = [
      // 代表(10) への drop（既存グループへの member 追加は BE が 409 拒否）
      {
        active: { data: { current: { mergeCageId: "11" } } },
        over: { id: "cage-10" },
      },
      // member(12) への drop
      {
        active: { data: { current: { mergeCageId: "11" } } },
        over: { id: "cage-12" },
      },
      // self drop
      {
        active: { data: { current: { mergeCageId: "11" } } },
        over: { id: "cage-11" },
      },
      // 占有 cage(13) への drop
      {
        active: { data: { current: { mergeCageId: "11" } } },
        over: { id: "cage-13" },
      },
      // member(12) を source にした drop（source 不適格）
      {
        active: { data: { current: { mergeCageId: "12" } } },
        over: { id: "cage-11" },
      },
      // inactive(14) への drop
      {
        active: { data: { current: { mergeCageId: "11" } } },
        over: { id: "cage-14" },
      },
      // over なし
      { active: { data: { current: { mergeCageId: "11" } } }, over: null },
      // patient drag 形式の active data は無視
      {
        active: { data: { current: { hospitalizationId: "h-1" } } },
        over: { id: "cage-10" },
      },
    ];
    for (const event of drops) {
      act(() => {
        dndHarness.onDragEnd?.(event);
      });
    }
    expect(onRequestMerge).not.toHaveBeenCalled();
  });

  it("merge mode ON: drag 中は eligible target が強調・ineligible がグレー表示/drop 不可になる", () => {
    const { rerender } = renderMergeBoard();

    dndHarness.active = { data: { current: { mergeCageId: "11" } } };
    rerender(
      <HospitalizationBoard
        cages={mergeCages}
        hospitalizations={[occupiedOn13]}
        onNavigateToForm={vi.fn()}
        onMovePet={vi.fn()}
        canCreate
        canEdit
        mergeMode
        canMerge
        onRequestMerge={vi.fn()}
        onRequestUnmerge={vi.fn()}
      />,
    );

    // eligible: 15（standalone・空き）, ineligible: 10(代表),11(self),12(member),13(占有),14(inactive)
    expect(dndHarness.useDroppable).toHaveBeenCalledWith(
      expect.objectContaining({ id: "cage-15", disabled: false }),
    );
    for (const id of ["cage-10", "cage-11", "cage-12", "cage-13", "cage-14"]) {
      expect(dndHarness.useDroppable).toHaveBeenCalledWith(
        expect.objectContaining({ id, disabled: true }),
      );
    }

    const eligibleCard = screen.getByText("ケージ15").closest("[data-slot='card']") as HTMLElement;
    const memberCard = screen.getByText("ケージ12").closest("[data-slot='card']") as HTMLElement;
    expect(eligibleCard.className).toMatch(/ring-2/);
    expect(memberCard.className).toMatch(/opacity-50/);
  });

  it("merge mode ON: static 表示では merge に参加できない cage がグレーアウトする", () => {
    renderMergeBoard();

    // member(12)・代表(10)・占有(13)・inactive(14) は merge に参加できない → mute
    for (const name of ["ケージ10", "ケージ12", "ケージ13", "ケージ14"]) {
      const card = screen.getByText(name).closest("[data-slot='card']") as HTMLElement;
      expect(card.className).toMatch(/opacity-50/);
    }
    // source(11) と standalone target(15) は mute しない
    for (const name of ["ケージ11", "ケージ15"]) {
      const card = screen.getByText(name).closest("[data-slot='card']") as HTMLElement;
      expect(card.className).not.toMatch(/opacity-50/);
    }
  });
});

describe("HospitalizationBoard EMR-306 merged member cards", () => {
  const repAndMember: MasterItem[] = [
    { id: "10", name: "ケージ10", category: "犬舎", price: 0, status: "active" },
    {
      id: "12",
      name: "ケージ12",
      category: "犬舎",
      price: 0,
      status: "active",
      mergedIntoCageId: "10",
    },
  ];

  it("member card は 連結中 + 代表ケージ名を表示し、新規 + button を出さず pet-move drop を拒否する", () => {
    const onMovePet = vi.fn();
    render(
      <HospitalizationBoard
        cages={repAndMember}
        hospitalizations={[]}
        onNavigateToForm={vi.fn()}
        onMovePet={onMovePet}
        canCreate
        canEdit
      />,
    );

    expect(screen.getByText("連結中: ケージ10")).toBeInTheDocument();
    // member(12) の空きカードには新規登録 + button が無い
    expect(
      screen.queryByRole("button", { name: /ケージ12（ケージID: 12）の空き枠に入院/ }),
    ).not.toBeInTheDocument();
    // 非 member(10) には従来通り + button がある
    expect(
      screen.getByRole("button", { name: /ケージ10（ケージID: 10）の空き枠に入院/ }),
    ).toBeInTheDocument();
    // droppable disabled（pet-move drop 拒否）
    expect(dndHarness.useDroppable).toHaveBeenCalledWith(
      expect.objectContaining({ id: "cage-12", disabled: true }),
    );
    // handler 側の二重ガード
    act(() => {
      dndHarness.onDragEnd?.({
        active: { data: { current: { hospitalizationId: "h-1" } } },
        over: { id: "cage-12" },
      });
    });
    expect(onMovePet).not.toHaveBeenCalled();
  });

  it("member card の 連結解除 button は代表 id で onRequestUnmerge を呼ぶ（canMerge 権限必須）", async () => {
    const user = userEvent.setup();
    const onRequestUnmerge = vi.fn();
    render(
      <HospitalizationBoard
        cages={repAndMember}
        hospitalizations={[]}
        onNavigateToForm={vi.fn()}
        onMovePet={vi.fn()}
        canCreate
        canEdit
        canMerge
        onRequestUnmerge={onRequestUnmerge}
      />,
    );

    await user.click(screen.getByRole("button", { name: "ケージ10への連結を解除" }));
    expect(onRequestUnmerge).toHaveBeenCalledWith("10");
  });

  it("rep card は 連結中（代表）バッジを表示し、解除 button は代表自身の id で onRequestUnmerge を呼ぶ", async () => {
    const user = userEvent.setup();
    const onRequestUnmerge = vi.fn();
    render(
      <HospitalizationBoard
        cages={repAndMember}
        hospitalizations={[]}
        onNavigateToForm={vi.fn()}
        onMovePet={vi.fn()}
        canCreate
        canEdit
        canMerge
        onRequestUnmerge={onRequestUnmerge}
      />,
    );

    expect(screen.getByText("連結中（代表）")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "ケージ10の連結を解除" }));
    expect(onRequestUnmerge).toHaveBeenCalledWith("10");
  });

  it("canMerge=false では member card に 連結解除 button を出さない", () => {
    render(
      <HospitalizationBoard
        cages={repAndMember}
        hospitalizations={[]}
        onNavigateToForm={vi.fn()}
        onMovePet={vi.fn()}
        canCreate
        canEdit
        canMerge={false}
        onRequestUnmerge={vi.fn()}
      />,
    );

    expect(screen.getByText("連結中: ケージ10")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /連結を解除/ })).not.toBeInTheDocument();
  });
});
