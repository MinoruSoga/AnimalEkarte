import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import type { AccountingItem } from "../types";
import { AccountingItemRow } from "./AccountingItemRow";

vi.mock("../api/get-discount-suggestions", () => ({
  useGetBillingItemDiscountSuggestions: () => ({
    data: [
      {
        type: "campaign",
        campaign_id: 7,
        name: "夏季割引",
        discount_type: "amount",
        discount_value: 100,
        amount: 100,
      },
    ],
    isFetching: false,
  }),
}));

const ITEM = {
  id: "item-1",
  category: "other",
  name: "療法食",
  unitPrice: 1200,
  quantity: 1,
  discountRate: 0,
  discountAmount: 0,
  taxType: "excluded",
  taxRate: 0.1,
  taxAmount: 120,
  subtotal: 1200,
  isInsuranceApplicable: false,
  source: "manual",
} satisfies AccountingItem;

interface RenderRowHandlers {
  onUpdateItemName?: (itemId: string, name: string) => void;
  onUpdateItemQuantity?: (itemId: string, quantity: number) => void;
  onUpdateItemAmount?: (item: AccountingItem, amount: number) => void;
}

function renderRow(canEdit = true, item: AccountingItem = ITEM, handlers: RenderRowHandlers = {}) {
  const onUpdateItemDiscount = vi.fn();
  render(
    <table>
      <tbody>
        <AccountingItemRow
          item={item}
          accountingId="accounting-1"
          canEdit={canEdit}
          canDelete={canEdit}
          onDeleteItem={vi.fn()}
          onUpdateItemTax={vi.fn()}
          onUpdateItemDiscount={onUpdateItemDiscount}
          onUpdateItemName={handlers.onUpdateItemName}
          onUpdateItemQuantity={handlers.onUpdateItemQuantity}
          onUpdateItemAmount={handlers.onUpdateItemAmount}
        />
      </tbody>
    </table>,
  );
  return { onUpdateItemDiscount };
}

describe("AccountingItemRow accessibility", () => {
  it("割引・税操作は品目固有名と44px操作領域を持つ", async () => {
    const user = userEvent.setup();
    renderRow();

    const discountInput = screen.getByRole("spinbutton", {
      name: "割引額: 療法食 (ID item-1)",
    });
    expect(discountInput).toHaveClass("min-h-11");

    const suggestionTrigger = screen.getByRole("button", {
      name: "割引候補: 療法食 (ID item-1)",
    });
    expect(suggestionTrigger).toHaveClass("min-h-11", "min-w-11");
    await user.click(suggestionTrigger);

    const suggestion = await screen.findByRole("button", {
      name: "割引を適用: 夏季割引 -100円 (品目ID item-1)",
    });
    expect(suggestion).toHaveClass("min-h-11", "min-w-11");

    const taxType = screen.getByRole("combobox", {
      name: "課税区分: 療法食 (ID item-1)",
    });
    const taxRate = screen.getByRole("combobox", {
      name: "税率: 療法食 (ID item-1)",
    });
    expect(taxType).toHaveClass("min-h-11", "min-w-11");
    expect(taxRate).toHaveClass("min-h-11", "min-w-11");
  });

  it("閲覧専用では割引入力を表示せず値をテキスト表示する", () => {
    renderRow(false, { ...ITEM, discountAmount: 100 });

    expect(
      screen.queryByRole("spinbutton", { name: "割引額: 療法食 (ID item-1)" }),
    ).not.toBeInTheDocument();
    expect(screen.getByText("¥100")).toBeInTheDocument();
  });

  it("項目名・区分バッジ・税額セルは折り返さず横スクロールさせる", () => {
    renderRow(true, {
      ...ITEM,
      name: "R7QAコース",
      category: "trimming",
      source: "trimming",
    });

    const nameCell = screen.getByText("R7QAコース").closest("td");
    expect(nameCell?.className).toContain("whitespace-nowrap");

    const trimmingNodes = screen.getAllByText("トリミング");
    expect(trimmingNodes.length).toBeGreaterThanOrEqual(2);

    const badge = trimmingNodes.find((node) => node.className.includes("font-normal"));
    expect(badge?.className).toContain("whitespace-nowrap");

    const sourceTag = trimmingNodes.find((node) => node.className.includes("ml-2"));
    expect(sourceTag?.closest("td")?.className).toContain("whitespace-nowrap");

    const taxCell = screen.getByText("¥120").closest("td");
    expect(taxCell?.className).toContain("whitespace-nowrap");
  });
});

// EMR-229/230: 項目名・数量・金額（税抜小計）の編集セル。
describe("AccountingItemRow EMR-229/230 editable cells", () => {
  it("canEdit + onUpdateItemName 指定で項目名入力を表示し、blur で trim 済みの値を渡す", async () => {
    const user = userEvent.setup();
    const onUpdateItemName = vi.fn();
    renderRow(true, ITEM, { onUpdateItemName });

    const nameInput = screen.getByRole("textbox", { name: "項目名: 療法食 (ID item-1)" });
    await user.clear(nameInput);
    await user.type(nameInput, "  療法食 R7.5  ");
    await user.tab();

    expect(onUpdateItemName).toHaveBeenCalledWith("item-1", "療法食 R7.5");
  });

  it("項目名を変更せず blur しても handler は呼ばない", async () => {
    const user = userEvent.setup();
    const onUpdateItemName = vi.fn();
    renderRow(true, ITEM, { onUpdateItemName });

    const nameInput = screen.getByRole("textbox", { name: "項目名: 療法食 (ID item-1)" });
    await user.click(nameInput);
    await user.tab();

    expect(onUpdateItemName).not.toHaveBeenCalled();
  });

  it("閲覧専用（canEdit=false）では項目名をテキスト表示する", () => {
    renderRow(false, ITEM, { onUpdateItemName: vi.fn() });

    expect(
      screen.queryByRole("textbox", { name: "項目名: 療法食 (ID item-1)" }),
    ).not.toBeInTheDocument();
    expect(screen.getByText("療法食")).toBeInTheDocument();
  });

  it("manual 行は数量・金額の入力を表示し blur で更新を呼ぶ (EMR-230)", async () => {
    const user = userEvent.setup();
    const onUpdateItemQuantity = vi.fn();
    const onUpdateItemAmount = vi.fn();
    const item: AccountingItem = {
      ...ITEM,
      unitPrice: 800,
      quantity: 2,
      subtotal: 1600,
      taxAmount: 160,
    };
    renderRow(true, item, { onUpdateItemQuantity, onUpdateItemAmount });

    const quantityInput = screen.getByRole("spinbutton", { name: "数量: 療法食 (ID item-1)" });
    expect(quantityInput).toHaveClass("min-h-11");
    // 金額入力の初期値は税抜小計（単価×数量−割引 = 800×2 = 1600。領収書の行金額と同式）
    const amountInput = screen.getByRole("spinbutton", { name: "金額: 療法食 (ID item-1)" });
    expect(amountInput).toHaveValue(1600);

    await user.clear(quantityInput);
    await user.type(quantityInput, "3");
    await user.tab();
    expect(onUpdateItemQuantity).toHaveBeenCalledWith("item-1", 3);
  });

  it("金額入力の blur は行全体（unitPrice/quantity/discountAmount 含む）と金額を渡す (EMR-230)", async () => {
    const user = userEvent.setup();
    const onUpdateItemAmount = vi.fn();
    const item: AccountingItem = {
      ...ITEM,
      unitPrice: 800,
      quantity: 2,
      discountAmount: 100,
      subtotal: 1500,
      taxAmount: 150,
    };
    renderRow(true, item, { onUpdateItemAmount });

    const amountInput = screen.getByRole("spinbutton", { name: "金額: 療法食 (ID item-1)" });
    await user.clear(amountInput);
    await user.type(amountInput, "2000");
    await user.tab();

    expect(onUpdateItemAmount).toHaveBeenCalledWith(item, 2000);
  });

  it("medical_record 行は数量・金額をテキスト表示のみ（カルテ由来の金額を直接編集させない）", () => {
    renderRow(
      true,
      { ...ITEM, source: "medical_record", unitPrice: 800, quantity: 2, subtotal: 1600 },
      { onUpdateItemQuantity: vi.fn(), onUpdateItemAmount: vi.fn() },
    );

    expect(
      screen.queryByRole("spinbutton", { name: "数量: 療法食 (ID item-1)" }),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByRole("spinbutton", { name: "金額: 療法食 (ID item-1)" }),
    ).not.toBeInTheDocument();
    // 金額セルは税抜小計（800×2）を表示する
    expect(screen.getByText("¥1,600")).toBeInTheDocument();
  });

  it("割引率があるとき割引セルに率ヒントを表示する (EMR-229)", () => {
    renderRow(true, { ...ITEM, discountRate: 10, discountAmount: 120 });

    expect(screen.getByText("(10%)")).toBeInTheDocument();
  });

  it("閲覧専用でも割引率ヒントを表示する (EMR-229)", () => {
    renderRow(false, { ...ITEM, discountRate: 5, discountAmount: 60 });

    expect(screen.getByText("(5%)")).toBeInTheDocument();
    expect(screen.getByText("¥60")).toBeInTheDocument();
  });
});
