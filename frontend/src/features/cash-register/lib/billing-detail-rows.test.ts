import { describe, it, expect } from "vitest";
import type { CloseBillingDetail } from "../api/get-cash-register-preview";
import {
  DETAIL_DEPARTMENT_BUCKETS,
  projectBillingDetailRows,
  sumBillingDetailRowTotals,
} from "./billing-detail-rows";

// EMR-354 R1-5: 投影 helper の回帰（部門バケット網羅・conservation・unknown 伝播・
// 入力不変性・過剰集約なし）。コンポーネントテストとは別に helper 契約を固定する。

function detail(overrides: Partial<CloseBillingDetail>): CloseBillingDetail {
  return {
    billingId: 1,
    paidAt: "2026-05-01T01:00:00Z",
    ownerName: "田中太郎",
    petName: "ポチ",
    isHospitalization: false,
    category: "examination",
    paymentMethodId: 1,
    paymentMethodName: "現金",
    billingAmount: 1000,
    refundAmount: 0,
    netAmount: 1000,
    ...overrides,
  };
}

describe("projectBillingDetailRows — 部門バケット", () => {
  it("日次8バケットへ全カテゴリが折り畳まれる（goods/training/other は用品他）", () => {
    const rows = projectBillingDetailRows([
      detail({
        categoryAmounts: {
          examination: 100,
          test: 110,
          procedure: 120,
          medicine: 130,
          surgery: 200,
          vaccine: 300,
          food: 400,
          trimming: 500,
          hotel: 600,
          goods: 700,
          training: 800,
          other: 900,
          house_call: 1000,
        },
      }),
    ]);
    expect(rows).toHaveLength(1);
    expect(rows[0].departments).toEqual({
      medical: 460, // examination+test+procedure+medicine
      surgery: 200,
      rv: 300, // vaccine
      food: 400,
      trimming: 500,
      hotel: 600,
      goodsOther: 2400, // goods+training+other
      houseCall: 1000,
    });
  });

  it("未知カテゴリは用品他に畳み込まれ金額が消えない", () => {
    const rows = projectBillingDetailRows([
      detail({ categoryAmounts: { future_category_x: 250, examination: 100 } }),
    ]);
    expect(rows[0].departments?.goodsOther).toBe(250);
    expect(rows[0].departments?.medical).toBe(100);
  });

  it("部門合計は行合計（GROSS）と一致する（conservation）", () => {
    const rows = projectBillingDetailRows([
      detail({ billingAmount: 700, categoryAmounts: { examination: 450, food: 250 } }),
      detail({ billingAmount: 300, categoryAmounts: { examination: 150, food: 150 } }),
    ]);
    const deptSum = DETAIL_DEPARTMENT_BUCKETS.reduce(
      (acc, b) => acc + (rows[0].departments?.[b.key] ?? 0),
      0,
    );
    expect(rows[0].totalAmount).toBe(1000);
    expect(deptSum).toBe(rows[0].totalAmount);
  });
});

describe("projectBillingDetailRows — 集約・不変性", () => {
  it("split 行を畳み、合計/返金(1回)/純額が正しく、領収Noは8桁", () => {
    const rows = projectBillingDetailRows([
      detail({
        billingId: 42,
        paymentMethodName: "現金",
        billingAmount: 1000,
        refundAmount: 500,
        categoryAmounts: { examination: 1000 },
      }),
      detail({
        billingId: 42,
        paymentMethodName: "クレジットカード",
        billingAmount: 3000,
        refundAmount: 500,
        categoryAmounts: { food: 3000 },
      }),
    ]);
    expect(rows).toHaveLength(1);
    expect(rows[0].receiptNo).toBe("00000042");
    expect(rows[0].totalAmount).toBe(4000);
    expect(rows[0].refundAmount).toBe(500); // 各行重複のスカラは1回だけ
    expect(rows[0].netAmount).toBe(3500);
    expect(rows[0].paymentMethodLabel).toBe("現金/クレジットカード");
    expect(rows[0].departments).toEqual(expect.objectContaining({ medical: 1000, food: 3000 }));
  });

  it("異なる billingId は別行のまま（過剰集約しない）", () => {
    const rows = projectBillingDetailRows([
      detail({ billingId: 1, categoryAmounts: { examination: 100 } }),
      detail({ billingId: 2, categoryAmounts: { surgery: 200 } }),
    ]);
    expect(rows).toHaveLength(2);
    expect(rows.map((r) => r.billingId)).toEqual([1, 2]);
  });

  it("凍結された入力でも動作し、入力行・カテゴリmapを変更しない", () => {
    const frozenMap = Object.freeze({ examination: 600, food: 400 });
    const frozenDetails = Object.freeze([
      detail({ billingId: 5, billingAmount: 1000, categoryAmounts: frozenMap }),
    ]);
    const rows = projectBillingDetailRows(frozenDetails);
    expect(rows[0].departments).toEqual(expect.objectContaining({ medical: 600, food: 400 }));
    expect(Object.isFrozen(frozenMap)).toBe(true);
    expect(frozenMap).toEqual({ examination: 600, food: 400 });
  });

  it("same invoice splits project house_call separately and coalesce scalar/payment totals", () => {
    const rows = projectBillingDetailRows([
      detail({
        billingId: 73,
        paymentMethodName: "現金",
        billingAmount: 700,
        refundAmount: 0,
        categoryAmounts: { house_call: 600, goods: 100 },
      }),
      detail({
        billingId: 73,
        paymentMethodName: "クレジットカード",
        billingAmount: 500,
        refundAmount: 0,
        categoryAmounts: { house_call: 400, goods: 100 },
      }),
    ]);

    expect(rows).toHaveLength(1);
    expect(rows[0].departments).toEqual(
      expect.objectContaining({ houseCall: 1000, goodsOther: 200 }),
    );
    expect(rows[0].totalAmount).toBe(1200);
    expect(rows[0].refundAmount).toBe(0);
    expect(rows[0].netAmount).toBe(1200);
    expect(rows[0].paymentMethodLabel).toBe("現金/クレジットカード");

    const totals = sumBillingDetailRowTotals(rows);
    expect(totals.departments).toEqual(
      expect.objectContaining({ houseCall: 1000, goodsOther: 200 }),
    );
    expect(totals.totalAmount).toBe(1200);
    expect(totals.refundAmount).toBe(0);
    expect(totals.netAmount).toBe(1200);
  });
});

describe("sumBillingDetailRowTotals — フッタ", () => {
  it("一部行が map 欠落なら行もフッタ部門も unknown、ただしスカラ合計は正しい", () => {
    const rows = projectBillingDetailRows([
      detail({ billingId: 1, billingAmount: 1000, refundAmount: 200 }),
      detail({
        billingId: 2,
        billingAmount: 3000,
        refundAmount: 100,
        categoryAmounts: { surgery: 3000 },
      }),
    ]);
    expect(rows[0].departments).toBeNull();
    const totals = sumBillingDetailRowTotals(rows);
    expect(totals.departments).toBeNull(); // 捏造 0 にしない
    expect(totals.totalAmount).toBe(4000);
    expect(totals.refundAmount).toBe(300);
    expect(totals.netAmount).toBe(3700);
  });

  it("全行 map ありなら部門合計が可視行と一致する", () => {
    const rows = projectBillingDetailRows([
      detail({ billingId: 1, billingAmount: 100, categoryAmounts: { examination: 100 } }),
      detail({ billingId: 2, billingAmount: 200, categoryAmounts: { goods: 50, other: 150 } }),
    ]);
    const totals = sumBillingDetailRowTotals(rows);
    expect(totals.departments).toEqual({
      medical: 100,
      surgery: 0,
      rv: 0,
      food: 0,
      trimming: 0,
      hotel: 0,
      goodsOther: 200,
      houseCall: 0,
    });
  });

  it("空入力は 0 合計・部門既知（0）を返す", () => {
    const totals = sumBillingDetailRowTotals(projectBillingDetailRows([]));
    expect(totals.departments).not.toBeNull();
    expect(totals.totalAmount).toBe(0);
    expect(totals.netAmount).toBe(0);
  });
});
