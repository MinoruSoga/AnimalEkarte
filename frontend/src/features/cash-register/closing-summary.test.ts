import { describe, it, expect } from "vitest";
import {
  buildUnifiedClosingRows,
  buildUnifiedClosingTotals,
  formatClosingCount,
} from "./closing-summary";
import type { CloseBillingDetail } from "./api/get-cash-register-preview";
import { UNCLASSIFIED_OTHER_LABEL } from "./constants";

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

describe("buildUnifiedClosingRows DEC-40 未分類・要確認", () => {
  it("other 行ラベルは「未分類・要確認」", () => {
    const rows = buildUnifiedClosingRows({ other: { 現金: 1000 } }, [], 1);
    expect(rows.map((r) => r.label)).toEqual([UNCLASSIFIED_OTHER_LABEL]);
  });

  it("独立集計件数を使い、billingDetails の MIN/split 行数に依存しない", () => {
    // mixed 会計が MIN(category) で examination に落ち、pure other が 2 split で 2 行になるケース
    const billingDetails = [
      detail({ billingId: 1, category: "examination" }), // mixed だが other を含む会計（detail 上は exam）
      detail({ billingId: 2, category: "other" }), // pure other split 1
      detail({ billingId: 2, category: "other" }), // pure other split 2
    ];
    const categories = {
      examination: { 現金: 3000 },
      other: { 現金: 1500 },
    };
    // サーバ独立集計: mixed 1 + pure other 1 = 2（detail の other 行数 2 や items 数ではない）
    const rows = buildUnifiedClosingRows(categories, billingDetails, 2);
    const other = rows.find((r) => r.label === UNCLASSIFIED_OTHER_LABEL);
    expect(other?.count).toBe(2);
    // billingDetails の other 行は 2 だが、独立件数も 2 で偶然一致し得るので
    // mixed を含めた「detail other 行 ≠ distinct」を明示: detail other 行 = 2、独立件数を 3 にすると乖離
    const rowsDistinct = buildUnifiedClosingRows(categories, billingDetails, 3);
    expect(rowsDistinct.find((r) => r.label === UNCLASSIFIED_OTHER_LABEL)?.count).toBe(3);
    expect(billingDetails.filter((d) => d.category === "other").length).toBe(2);
  });

  it("旧 snapshot（件数 null）は「記録なし」として件数セル用に null を返す", () => {
    const rows = buildUnifiedClosingRows({ other: { 現金: 500 } }, [], null);
    const other = rows.find((r) => r.label === UNCLASSIFIED_OTHER_LABEL);
    expect(other).toEqual({
      label: UNCLASSIFIED_OTHER_LABEL,
      count: null,
      byMethod: { 現金: 500 },
      rowTotal: 500,
    });
    expect(formatClosingCount(null)).toBe("記録なし");
  });

  it("件数 0・金額 0 の other 行は除外する", () => {
    expect(buildUnifiedClosingRows({}, [], 0)).toEqual([]);
  });

  it("件数未記録かつ金額 0 の other 行は除外する（記録なしを 0 行として出さない）", () => {
    expect(buildUnifiedClosingRows({}, [], null)).toEqual([]);
  });

  it("totals は記録なし (null) 件数を加算しない", () => {
    const rows = buildUnifiedClosingRows(
      {
        examination: { 現金: 1000 },
        other: { 現金: 500 },
      },
      [detail({ billingId: 1, category: "examination" })],
      null,
    );
    const totals = buildUnifiedClosingTotals(rows);
    expect(totals.count).toBe(1);
    expect(totals.grandTotal).toBe(1500);
  });
});
