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

  it("手動入力では商品マスタ ID を渡さない", async () => {
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

    await user.click(screen.getByRole("button", { name: "手動入力" }));
    await user.type(screen.getByLabelText(/品目名/), "手入力商品");
    await user.type(screen.getByLabelText(/単価/), "500");
    await user.click(screen.getByRole("button", { name: "追加する" }));

    expect(onAddItem).toHaveBeenCalledWith({
      name: "手入力商品",
      price: "500",
      category: "other",
    });
  });
});
