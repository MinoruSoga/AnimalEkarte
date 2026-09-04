import { describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { ReservationTypeTree } from "./ReservationTypeTree";
import type { ReservationType } from "../api/reservation-types";

function makeType(overrides: Partial<ReservationType>): ReservationType {
  return {
    id: "1",
    clinicId: "1",
    name: "区分",
    color: "#3B82F6",
    isActive: true,
    description: "",
    sortOrder: 1,
    groupId: undefined,
    createdAt: "2026-05-29T00:00:00Z",
    updatedAt: "2026-05-29T00:00:00Z",
    reservationDisplayName: "",
    durationMinutes: 15,
    shortName: "",
    showShortName: false,
    reservationVisible: true,
    reservationComment: "",
    reservationImageUrl: "",
    reservationDayOption: "none",
    isInternal: false,
    category: "general",
    parentId: undefined,
    parentName: undefined,
    isLeaf: true,
    depth: 0,
    childIds: [],
    ...overrides,
  };
}

const parentNode = makeType({
  id: "1",
  name: "LINEコース",
  isLeaf: false,
  depth: 0,
  childIds: ["2", "3"],
});

const child1 = makeType({
  id: "2",
  name: "初診コース",
  isLeaf: true,
  depth: 1,
  parentId: "1",
  parentName: "LINEコース",
});

const child2 = makeType({
  id: "3",
  name: "再診コース",
  isLeaf: true,
  depth: 1,
  parentId: "1",
  parentName: "LINEコース",
});

const rootLeaf = makeType({
  id: "10",
  name: "一般診療",
  isLeaf: true,
  depth: 0,
});

const types = [parentNode, child1, child2, rootLeaf];

describe("ReservationTypeTree", () => {
  it("親・leafの予約種別buttonを44px以上に保つ", async () => {
    const user = userEvent.setup();
    render(<ReservationTypeTree types={types} selectedId={null} onSelect={vi.fn()} />);

    const parent = screen.getByRole("button", { name: /LINEコース/ });
    const rootLeafButton = screen.getByRole("button", { name: /一般診療/ });
    expect(parent).toHaveClass("min-h-11");
    expect(rootLeafButton).toHaveClass("min-h-11");

    await user.click(parent);
    expect(screen.getByRole("button", { name: /初診コース/ })).toHaveClass("min-h-11");
  });

  it("親ノードクリックで子が展開/折りたたみされる", async () => {
    const user = userEvent.setup();
    render(<ReservationTypeTree types={types} selectedId={null} onSelect={vi.fn()} />);

    expect(screen.queryByText("初診コース")).not.toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: /LINEコース/ }));
    expect(screen.getByText("初診コース")).toBeInTheDocument();
    expect(screen.getByText("再診コース")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: /LINEコース/ }));
    expect(screen.queryByText("初診コース")).not.toBeInTheDocument();
  });

  it("leaf クリックで onSelect が呼ばれる", async () => {
    const onSelect = vi.fn();
    const user = userEvent.setup();
    render(<ReservationTypeTree types={types} selectedId={null} onSelect={onSelect} />);

    await user.click(screen.getByRole("button", { name: /LINEコース/ }));
    await user.click(screen.getByRole("button", { name: /初診コース/ }));

    expect(onSelect).toHaveBeenCalledWith("2");
    expect(onSelect).toHaveBeenCalledTimes(1);
  });

  it("親ノードクリックでは onSelect が呼ばれない", async () => {
    const onSelect = vi.fn();
    const user = userEvent.setup();
    render(<ReservationTypeTree types={types} selectedId={null} onSelect={onSelect} />);

    await user.click(screen.getByRole("button", { name: /LINEコース/ }));

    expect(onSelect).not.toHaveBeenCalled();
  });

  it("selected leaf の親が初期展開状態になる", () => {
    render(<ReservationTypeTree types={types} selectedId="2" onSelect={vi.fn()} />);

    expect(screen.getByText("初診コース")).toBeInTheDocument();
    expect(screen.getByText("再診コース")).toBeInTheDocument();
  });

  it("inactive leaf に「（無効）」が独立要素として表示される", async () => {
    const user = userEvent.setup();
    const inactiveChild = makeType({
      id: "5",
      name: "休止コース",
      isLeaf: true,
      depth: 1,
      parentId: "1",
      parentName: "LINEコース",
      isActive: false,
    });
    const typesWithInactive = [parentNode, child1, inactiveChild, rootLeaf];

    render(<ReservationTypeTree types={typesWithInactive} selectedId={null} onSelect={vi.fn()} />);

    await user.click(screen.getByRole("button", { name: /LINEコース/ }));
    // exact match: （無効） は名前 span とは別の独立した span に存在する
    expect(screen.getByText("（無効）")).toBeInTheDocument();
  });

  it("root-only leaf が選択可能で onSelect を呼ぶ", async () => {
    const onSelect = vi.fn();
    const user = userEvent.setup();
    render(<ReservationTypeTree types={types} selectedId={null} onSelect={onSelect} />);

    await user.click(screen.getByRole("button", { name: /一般診療/ }));

    expect(onSelect).toHaveBeenCalledWith("10");
  });

  it("root-only leaf は親の展開なしで直接表示される", () => {
    render(<ReservationTypeTree types={types} selectedId={null} onSelect={vi.fn()} />);

    expect(screen.getByText("一般診療")).toBeInTheDocument();
    expect(screen.queryByText("初診コース")).not.toBeInTheDocument();
  });

  it("親ノードに「グループ」badge が表示される", () => {
    render(<ReservationTypeTree types={types} selectedId={null} onSelect={vi.fn()} />);

    expect(screen.getByText("グループ")).toBeInTheDocument();
  });

  it("親ノードの aria-expanded が開閉で切り替わる", async () => {
    const user = userEvent.setup();
    render(<ReservationTypeTree types={types} selectedId={null} onSelect={vi.fn()} />);

    const parentBtn = screen.getByRole("button", { name: /LINEコース/ });
    expect(parentBtn).toHaveAttribute("aria-expanded", "false");

    await user.click(parentBtn);
    expect(parentBtn).toHaveAttribute("aria-expanded", "true");

    await user.click(parentBtn);
    expect(parentBtn).toHaveAttribute("aria-expanded", "false");
  });

  it("子 leaf 選択時に aria-current が true になる", () => {
    // selectedId="2" なら親が自動展開 → 初診コースが表示される
    render(<ReservationTypeTree types={types} selectedId="2" onSelect={vi.fn()} />);

    const selectedBtn = screen.getByRole("button", { name: /初診コース/ });
    expect(selectedBtn).toHaveAttribute("aria-current", "true");

    const otherBtn = screen.getByRole("button", { name: /再診コース/ });
    expect(otherBtn).not.toHaveAttribute("aria-current");
  });

  it("inactive leaf はクリック可能で onSelect が呼ばれる", async () => {
    const onSelect = vi.fn();
    const user = userEvent.setup();
    const inactiveRoot = makeType({
      id: "20",
      name: "停止中の区分",
      isLeaf: true,
      depth: 0,
      isActive: false,
    });

    render(<ReservationTypeTree types={[inactiveRoot]} selectedId={null} onSelect={onSelect} />);

    await user.click(screen.getByRole("button", { name: "停止中の区分（無効）" }));
    expect(onSelect).toHaveBeenCalledWith("20");
  });

  it("子コンテナに role=group が付く", async () => {
    const user = userEvent.setup();
    render(<ReservationTypeTree types={types} selectedId={null} onSelect={vi.fn()} />);

    await user.click(screen.getByRole("button", { name: /LINEコース/ }));

    const group = screen.getByRole("group", { name: "LINEコース" });
    expect(group).toBeInTheDocument();
  });

  it("root-only leaf には role=group コンテナが付かない", () => {
    render(<ReservationTypeTree types={types} selectedId={null} onSelect={vi.fn()} />);

    const rootBtn = screen.getByRole("button", { name: /一般診療/ });
    expect(rootBtn.closest('[role="group"]')).toBeNull();
  });
});
