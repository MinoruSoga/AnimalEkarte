import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { ItemListCard } from "./ItemListCard";

vi.mock("../api/get-merchandise-items", () => ({
  useGetAllMerchandiseItems: () => ({
    data: [
      {
        id: "item-1",
        name: "療法食",
        category: "goods",
        unitPrice: 1200,
        taxRate: 0.1,
        isActive: true,
      },
    ],
  }),
}));

describe("ItemListCard merchandise selection", () => {
  it("行は追加せず、固有名の44px native buttonだけが品目を追加する", async () => {
    const user = userEvent.setup();
    const onAddItem = vi.fn();
    render(
      <ItemListCard
        items={[]}
        subtotal={0}
        taxTotal={0}
        totalAmount={0}
        newItemOpen
        onNewItemOpenChange={vi.fn()}
        onAddItem={onAddItem}
        onDeleteItem={vi.fn()}
        canEdit
        canDelete={false}
      />,
    );

    await user.click(screen.getByText("物販"));
    expect(onAddItem).not.toHaveBeenCalled();

    const addButton = screen.getByRole("button", { name: "追加: 療法食 (ID item-1)" });
    expect(addButton.tagName).toBe("BUTTON");
    expect(addButton).toHaveClass("min-h-11", "min-w-11");
    await user.click(addButton);
    expect(onAddItem).toHaveBeenCalledWith({
      name: "療法食",
      price: "1200",
      category: "goods",
      taxRate: 0.1,
      merchandiseItemId: "item-1",
    });
  });

  function renderEditableItemListCard(onAddItem = vi.fn()) {
    render(
      <ItemListCard
        items={[]}
        subtotal={0}
        taxTotal={0}
        totalAmount={0}
        newItemOpen
        onNewItemOpenChange={vi.fn()}
        onAddItem={onAddItem}
        onDeleteItem={vi.fn()}
        canEdit
        canDelete={false}
      />,
    );
  }

  it("手動入力では選択したカテゴリを渡し、商品マスタ ID を渡さない", async () => {
    const user = userEvent.setup();
    const onAddItem = vi.fn();
    renderEditableItemListCard(onAddItem);

    await user.click(screen.getByRole("button", { name: "手動入力" }));
    await user.type(screen.getByLabelText(/品目名/), "手入力商品");
    await user.type(screen.getByLabelText(/単価/), "500");
    await user.click(screen.getByRole("combobox", { name: "カテゴリ" }));
    expect(screen.getAllByRole("option")).toHaveLength(12);
    await user.click(screen.getByRole("option", { name: "検査" }));
    await user.click(screen.getByRole("button", { name: "追加する" }));

    expect(onAddItem).toHaveBeenCalledWith({
      name: "手入力商品",
      price: "500",
      category: "test",
    });
  });

  it("手動入力ではカテゴリ未選択のまま追加できない", async () => {
    const user = userEvent.setup();
    const onAddItem = vi.fn();
    renderEditableItemListCard(onAddItem);

    await user.click(screen.getByRole("button", { name: "手動入力" }));
    await user.type(screen.getByLabelText(/品目名/), "カテゴリ未選択商品");
    await user.type(screen.getByLabelText(/単価/), "500");

    expect(screen.getByRole("button", { name: "追加する" })).toBeDisabled();
    await user.click(screen.getByRole("button", { name: "追加する" }));
    expect(onAddItem).not.toHaveBeenCalled();
  });

  it("手動入力ではotherを明示選択できる", async () => {
    const user = userEvent.setup();
    const onAddItem = vi.fn();
    renderEditableItemListCard(onAddItem);

    await user.click(screen.getByRole("button", { name: "手動入力" }));
    await user.type(screen.getByLabelText(/品目名/), "その他手入力商品");
    await user.type(screen.getByLabelText(/単価/), "500");
    await user.click(screen.getByRole("combobox", { name: "カテゴリ" }));
    await user.click(screen.getByRole("option", { name: "その他" }));
    await user.click(screen.getByRole("button", { name: "追加する" }));

    expect(onAddItem).toHaveBeenCalledWith({
      name: "その他手入力商品",
      price: "500",
      category: "other",
    });
  });
});
