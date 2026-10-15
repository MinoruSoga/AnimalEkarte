import { describe, it, expect } from "vitest";

import { getCategoryBreakdown, apportionPayment } from "./daily-accounting-utils";
import type { AccountingItem } from "../types";

const makeItem = (category: AccountingItem["category"], subtotal: number): AccountingItem => ({
  id: "1",
  category,
  name: "item",
  unitPrice: subtotal,
  quantity: 1,
  discountRate: 0,
  discountAmount: 0,
  taxType: "excluded",
  taxRate: 0.1,
  taxAmount: 0,
  subtotal,
  isInsuranceApplicable: false,
  source: "manual",
});

describe("getCategoryBreakdown: house_call", () => {
  it("house_call を goods ではなく独立した houseCall へ集計する", () => {
    const items = [makeItem("house_call", 1000), makeItem("goods", 300)];
    const b = getCategoryBreakdown(items);
    expect(b.houseCall).toBe(1000);
    expect(b.goods).toBe(300);
  });

  it("複数 house_call と負 subtotal は符号込みで合算する", () => {
    const items = [
      makeItem("house_call", 1000),
      makeItem("house_call", 500),
      makeItem("house_call", -200),
    ];
    expect(getCategoryBreakdown(items).houseCall).toBe(1300);
  });

  it("house_call が無い既存カテゴリの集計は変わらない", () => {
    const b = getCategoryBreakdown([makeItem("goods", 300), makeItem("other", 50)]);
    expect(b.goods).toBe(350);
    expect(b.houseCall).toBe(0);
  });
});

describe("apportionPayment: house_call", () => {
  it("houseCall は goods より後（末尾）で現金を割り当てる", () => {
    const d = apportionPayment(
      {
        medical: 1000,
        surgery: 0,
        rv: 0,
        food: 0,
        trimming: 0,
        hotel: 0,
        goods: 500,
        houseCall: 2000,
      },
      1750,
    );
    expect(d.medical).toEqual({ cash: 1000, card: 0 });
    expect(d.goods).toEqual({ cash: 500, card: 0 });
    expect(d.houseCall).toEqual({ cash: 250, card: 1750 });
    const cashSum = Object.values(d).reduce((s, v) => s + v.cash, 0);
    const cardSum = Object.values(d).reduce((s, v) => s + v.card, 0);
    expect(cashSum).toBe(1750);
    expect(cardSum).toBe(1750);
  });

  it("houseCall だけの内訳は全額現金になる", () => {
    const d = apportionPayment(
      { medical: 0, surgery: 0, rv: 0, food: 0, trimming: 0, hotel: 0, goods: 0, houseCall: 2000 },
      2000,
    );
    expect(d.houseCall).toEqual({ cash: 2000, card: 0 });
  });
});
