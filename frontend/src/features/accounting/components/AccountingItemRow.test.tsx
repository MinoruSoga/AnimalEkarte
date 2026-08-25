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

function renderRow(canEdit = true, item: AccountingItem = ITEM) {
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
