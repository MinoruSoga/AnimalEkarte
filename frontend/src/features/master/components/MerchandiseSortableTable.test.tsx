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
});
