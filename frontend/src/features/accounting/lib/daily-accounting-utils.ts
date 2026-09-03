import type { Accounting, AccountingItem } from "../types";

// ── 定数 ────────────────────────────────────────────────────────
const MEDICAL_CATEGORIES = new Set(["examination", "test", "procedure", "medicine"]);

// ── ヘルパー ─────────────────────────────────────────────────────
export function formatReceiptNo(billingId: string): string {
  return billingId.padStart(8, "0");
}

// ── カテゴリ集計 ─────────────────────────────────────────────────
export interface CategoryBreakdown {
  medical: number;
  surgery: number;
  rv: number; // vaccine
  food: number;
  trimming: number;
  hotel: number;
  goods: number; // goods + training + other
}

export function getCategoryBreakdown(items: AccountingItem[]): CategoryBreakdown {
  let medical = 0;
  let surgery = 0;
  let rv = 0;
  let food = 0;
  let trimming = 0;
  let hotel = 0;
  let goods = 0;
  for (const item of items) {
    const sub = item.subtotal;
    if (MEDICAL_CATEGORIES.has(item.category)) {
      medical += sub;
    } else if (item.category === "surgery") {
      surgery += sub;
    } else if (item.category === "vaccine") {
      rv += sub;
    } else if (item.category === "food") {
      food += sub;
    } else if (item.category === "trimming") {
      trimming += sub;
    } else if (item.category === "hotel") {
      hotel += sub;
    } else {
      goods += sub;
    }
  }
  return { medical, surgery, rv, food, trimming, hotel, goods };
}

// ── 支払按分（印刷用）: RV優先→診療優先で現金を割り当て ──────────
export interface CashCardAmount {
  cash: number;
  card: number;
}

export interface DetailedBreakdown {
  medical: CashCardAmount;
  surgery: CashCardAmount;
  rv: CashCardAmount;
  food: CashCardAmount;
  trimming: CashCardAmount;
  hotel: CashCardAmount;
  goods: CashCardAmount;
}

export function apportionPayment(
  breakdown: CategoryBreakdown,
  cashTotal: number,
): DetailedBreakdown {
  // RV優先、次に診療、次に外科…と現金を割り当てる
  const PRIORITY: Array<keyof CategoryBreakdown> = [
    "rv",
    "medical",
    "surgery",
    "food",
    "trimming",
    "hotel",
    "goods",
  ];
  let cashLeft = cashTotal;
  const result: DetailedBreakdown = {
    medical: { cash: 0, card: 0 },
    surgery: { cash: 0, card: 0 },
    rv: { cash: 0, card: 0 },
    food: { cash: 0, card: 0 },
    trimming: { cash: 0, card: 0 },
    hotel: { cash: 0, card: 0 },
    goods: { cash: 0, card: 0 },
  };
  for (const cat of PRIORITY) {
    const catTotal = breakdown[cat];
    const catCash = Math.min(catTotal, cashLeft);
    result[cat] = { cash: catCash, card: catTotal - catCash };
    cashLeft -= catCash;
  }
  return result;
}

export function getRowCashTotal(accounting: Accounting): number {
  if (accounting.paymentSplits && accounting.paymentSplits.length > 1) {
    return accounting.paymentSplits
      .filter((s) => s.method === "cash")
      .reduce((sum, s) => sum + s.amount, 0);
  }
  const total = accounting.payment?.totalAmount ?? 0;
  return accounting.payment?.method === "cash" ? total : 0;
}

// ── 行・合計データ ───────────────────────────────────────────────
export interface RowData {
  accounting: Accounting;
  breakdown: CategoryBreakdown;
  detailedBreakdown: DetailedBreakdown;
  subtotal: number;
  tax: number;
  discount: number;
  total: number;
}

export interface TotalsData {
  medical: number;
  surgery: number;
  rv: number;
  food: number;
  trimming: number;
  hotel: number;
  goods: number;
  subtotal: number;
  tax: number;
  discount: number;
  total: number;
}
