import { describe, it, expect } from "vitest";
import {
  buildUnifiedClosingRows,
  buildUnifiedClosingTotals,
} from "./closing-summary";
import type { CloseBillingDetail } from "./api/get-cash-register-preview";

function detail(overrides: Partial<CloseBillingDetail>): CloseBillingDetail {
  return {
    billingId: 1,
    paidAt: "2026-05-01T01:00:00Z",
    ownerName: "飼主",
    petName: "ペット",
    isHospitalization: false,
    category: "examination",
    paymentMethodId: 1,
    paymentMethodName: "現金",
    billingAmount: 0,
    refundAmount: 0,
    netAmount: 0,
    ...overrides,
  };
}

describe("buildUnifiedClosingRows (#153 統合テーブル)", () => {
  it("部門ごとに 件数 + 支払方法別金額 + 合計 を統合する", () => {
    const categories = {
      // 同じ表示区分「診療」にまとまる複数の生カテゴリ
      examination: { 現金: 3000, クレジットカード: 2000 },
      medicine: { 現金: 1000 },
      surgery: { 現金: 5000 },
    };
    const billingDetails = [
      detail({ billingId: 1, category: "examination" }),
      detail({ billingId: 2, category: "medicine" }),
      detail({ billingId: 3, category: "surgery" }),
    ];

    const rows = buildUnifiedClosingRows(categories, billingDetails);
    const medical = rows.find((r) => r.label === "診療");
    const surgery = rows.find((r) => r.label === "外科");

    // 診療 = examination + medicine: 件数2 / 現金4000 / カード2000 / 合計6000
    expect(medical).toEqual({
      label: "診療",
      count: 2,
      byMethod: { 現金: 4000, クレジットカード: 2000 },
      rowTotal: 6000,
    });
    expect(surgery).toEqual({
      label: "外科",
      count: 1,
      byMethod: { 現金: 5000 },
      rowTotal: 5000,
    });
    // 表示区分順を維持
    expect(rows.map((r) => r.label)).toEqual(["診療", "外科"]);
  });

  it("金額0でも件数があれば行を残す（取りこぼし防止）", () => {
    const rows = buildUnifiedClosingRows({}, [detail({ category: "trimming" })]);
    expect(rows).toEqual([{ label: "トリミング", count: 1, byMethod: {}, rowTotal: 0 }]);
  });

  it("金額も件数も無い部門は除外する", () => {
    expect(buildUnifiedClosingRows({ surgery: { 現金: 0 } }, [])).toEqual([]);
  });

  it("buildUnifiedClosingTotals は件数・支払方法別・総合計を集計する", () => {
    const categories = {
      examination: { 現金: 3000, クレジットカード: 2000 },
      surgery: { 現金: 5000 },
    };
    const billingDetails = [
      detail({ billingId: 1, category: "examination" }),
      detail({ billingId: 2, category: "surgery" }),
    ];
    const rows = buildUnifiedClosingRows(categories, billingDetails);
    const totals = buildUnifiedClosingTotals(rows);

    expect(totals).toEqual({
      count: 2,
      byMethod: { 現金: 8000, クレジットカード: 2000 },
      grandTotal: 10000,
    });
  });
});
