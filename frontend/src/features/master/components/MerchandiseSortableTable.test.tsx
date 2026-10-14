import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import type { FrontendMerchandiseItem } from "../api/merchandise-items";
import { MerchandiseSortableTable } from "./MerchandiseSortableTable";

const items: FrontendMerchandiseItem[] = Array.from({ length: 101 }, (_, index) => ({
  id: String(index + 1),
  name: `合成品目${index + 1}`,
  category: "goods",
  unitPrice: 1100,
  taxType: "included",
  taxRate: 0.1,
  isActive: true,
  sortOrder: index,
}));
const handlers = {
  sensors: [],
  activeId: null,
  onDragStart: vi.fn(),
  onDragEnd: vi.fn(),
  onDragCancel: vi.fn(),
  canEdit: true,
  onEdit: vi.fn(),
};

describe("Merchandise visible pagination", () => {
  it("renders 50 rows at most, changes pages and clamps after data shrinks", async () => {
    const user = userEvent.setup();
    const view = render(<MerchandiseSortableTable items={items} {...handlers} />);
    expect(screen.getAllByRole("button", { name: /^詳細: 品目/ })).toHaveLength(50);
    expect(
      screen.queryByRole("button", { name: /詳細: 品目 合成品目51 / }),
    ).not.toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "2" }));
    expect(screen.getByRole("button", { name: /詳細: 品目 合成品目51 / })).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "最後のページ" }));
    expect(screen.getAllByRole("button", { name: /^詳細: 品目/ })).toHaveLength(1);
    view.rerender(<MerchandiseSortableTable items={items.slice(0, 2)} {...handlers} />);
    expect(screen.getAllByRole("button", { name: /^詳細: 品目/ })).toHaveLength(2);
    expect(screen.getByRole("button", { name: "1" })).toHaveAttribute("aria-current", "page");
  });
  it("a reordered full list retains off-page items and global positions", async () => {
    const user = userEvent.setup();
    const reordered = [items[1]!, items[0]!, ...items.slice(2)];
    render(<MerchandiseSortableTable items={reordered} {...handlers} />);
    const rows = screen.getAllByRole("button", { name: /^詳細: 品目/ });
    expect(rows[0]).toHaveTextContent("合成品目2");
    expect(rows[1]).toHaveTextContent("合成品目1");
    await user.click(screen.getByRole("button", { name: "最後のページ" }));
    expect(screen.getByRole("button", { name: /詳細: 品目 合成品目101 / })).toBeInTheDocument();
    expect(screen.getByText("101件中 101-101件")).toBeInTheDocument();
  });

  it("PR665: ページ境界をまたぐ移動（50件目を後へ・51件目を前へ）が明示ボタンから出せる", async () => {
    const user = userEvent.setup();
    const onMoveItem = vi.fn();
    render(<MerchandiseSortableTable items={items} {...handlers} onMoveItem={onMoveItem} />);

    // 1 ページ目末尾（50件目）を後へ移動 → 51件目の直後（2ページ目先頭）へ
    await user.click(screen.getByRole("button", { name: "後へ移動: 品目 合成品目50 (ID 50)" }));
    expect(onMoveItem).toHaveBeenCalledWith(items[49], "down");

    // 2 ページ目先頭（51件目）を前へ移動 → 50件目の直前（1ページ目末尾）へ
    await user.click(screen.getByRole("button", { name: "2" }));
    await user.click(screen.getByRole("button", { name: "前へ移動: 品目 合成品目51 (ID 51)" }));
    expect(onMoveItem).toHaveBeenCalledWith(items[50], "up");
  });

  it("PR665: onMoveItem 未指定または閲覧専用では移動ボタンを出さない", () => {
    const { unmount } = render(<MerchandiseSortableTable items={items} {...handlers} />);
    expect(screen.queryByRole("button", { name: /前へ移動:/ })).not.toBeInTheDocument();
    unmount();
    render(
      <MerchandiseSortableTable items={items} {...handlers} canEdit={false} onMoveItem={vi.fn()} />,
    );
    expect(screen.queryByRole("button", { name: /前へ移動:/ })).not.toBeInTheDocument();
  });

  it("PR665: 並べ替え中は移動ボタンが無効化されクリックが伝わらない", async () => {
    const user = userEvent.setup();
    const onMoveItem = vi.fn();
    render(
      <MerchandiseSortableTable items={items} {...handlers} onMoveItem={onMoveItem} isReordering />,
    );
    const up = screen.getByRole("button", { name: "前へ移動: 品目 合成品目1 (ID 1)" });
    expect(up).toBeDisabled();
    await user.click(up);
    expect(onMoveItem).not.toHaveBeenCalled();
  });

  it("並べ替え中は全ドラッグハンドルを無効化し、完了後に再開できる", () => {
    const view = render(<MerchandiseSortableTable items={items} {...handlers} isReordering />);
    const handles = () => screen.getAllByRole("button", { name: /^並べ替え: 品目/ });
    expect(handles()).toHaveLength(50);
    for (const handle of handles()) expect(handle).toBeDisabled();
    view.rerender(<MerchandiseSortableTable items={items} {...handlers} isReordering={false} />);
    for (const handle of handles()) expect(handle).not.toBeDisabled();
  });

  it("PR665: 操作列は移動2+編集1の44pxボタンが1行に収まる160px幅を保つ", () => {
    render(<MerchandiseSortableTable items={items} {...handlers} onMoveItem={vi.fn()} />);
    const headers = screen.getAllByRole("columnheader");
    const actionHeader = headers[headers.length - 1]!;
    expect(actionHeader).toHaveClass("w-[160px]", "min-w-[160px]");
    const rows = screen.getAllByRole("row");
    const actionCell = rows[1]!.querySelectorAll("td");
    const lastCell = actionCell[actionCell.length - 1]!;
    expect(lastCell).toHaveClass("whitespace-nowrap");
    const moveUp = screen.getByRole("button", { name: "前へ移動: 品目 合成品目1 (ID 1)" });
    expect(moveUp).toHaveClass("h-11", "w-11");
    const edit = screen.getByRole("button", { name: "編集: 品目 合成品目1 (ID 1)" });
    expect(moveUp.parentElement).toContainElement(edit);
    expect(moveUp.parentElement?.children).toHaveLength(3);
    expect(moveUp.parentElement).toHaveClass("flex", "items-center");
  });
});
