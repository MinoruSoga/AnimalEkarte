import { describe, it, expect } from "vitest";
import { render, screen, within } from "@testing-library/react";
import { BillingDetailTable } from "./BillingDetailTable";
import type { CloseBillingDetail } from "../api/get-cash-register-preview";

// EMR-354: 個別会計明細を「会計1件=1行」の日次バケット列へ投影する回帰テスト。
// RED 時点では categoryAmounts フィールド・固定列・billingId 集約が未実装のため失敗した。
// GREEN 後は categoryAmounts が型に存在するため fixture は Partial<CloseBillingDetail>。

const EXPECTED_HEADERS = [
  "時刻",
  "領収No",
  "診療",
  "外科",
  "RV",
  "フード",
  "トリミング",
  "ホテル",
  "用品他",
  "支払方法",
  "合計",
  "返金額",
  "純額",
  "飼主名",
  "ペット名",
];

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
    billingAmount: 5000,
    refundAmount: 0,
    netAmount: 5000,
    ...overrides,
  };
}

function bodyRows(table: HTMLElement) {
  return table.querySelectorAll<HTMLTableRowElement>("tbody tr");
}

describe("BillingDetailTable — EMR-354 日次バケット投影", () => {
  it("固定列が日次バケット順で出力される（時刻, 領収No, 診療, 外科, RV, フード, トリミング, ホテル, 用品他, 支払方法, 合計, 返金額, 純額, 飼主名, ペット名）", () => {
    render(<BillingDetailTable details={[detail({ categoryAmounts: { examination: 5000 } })]} />);
    const headers = within(screen.getByRole("table"))
      .getAllByRole("columnheader")
      .map((h) => h.textContent);
    expect(headers).toEqual(EXPECTED_HEADERS);
  });

  it("同一 billingId の split 行は1行に集約され、合計・返金・純額が正しく出る", () => {
    render(
      <BillingDetailTable
        details={[
          detail({
            billingId: 7,
            paymentMethodName: "現金",
            billingAmount: 1000,
            refundAmount: 1234,
            netAmount: -234,
            categoryAmounts: { examination: 667, food: 333 },
          }),
          detail({
            billingId: 7,
            paymentMethodName: "クレジットカード",
            billingAmount: 3000,
            refundAmount: 1234,
            netAmount: 1766,
            categoryAmounts: { examination: 2000, food: 1000 },
          }),
        ]}
      />,
    );
    const rows = bodyRows(screen.getByRole("table"));
    // 現行バグ: billingId 重複キーで split が2行レンダリングされる → 1行であるべき
    expect(rows).toHaveLength(1);
    const row = rows[0];
    // 支払方法は split を順序確定的に結合
    expect(within(row).getByText("現金/クレジットカード")).toBeInTheDocument();
    // 合計 = split billing_amount の和（GROSS）
    expect(within(row).getByText("¥4,000")).toBeInTheDocument();
    // refund_amount は全期間スカラが各行に重複している → 1回だけ取る
    expect(within(row).getByText("-¥1,234")).toBeInTheDocument();
    // 純額 = GROSS 合計 - 返金（netAmount の行和 1532 ではない）
    expect(within(row).getByText("¥2,766")).toBeInTheDocument();
  });

  it("1会計が複数部門を持つとき1行の固定順セルに出て、0部門もセルとして残る", () => {
    render(
      <BillingDetailTable
        details={[
          detail({
            billingId: 9,
            billingAmount: 4000,
            netAmount: 4000,
            categoryAmounts: { examination: 2667, food: 1333 },
          }),
        ]}
      />,
    );
    const row = bodyRows(screen.getByRole("table"))[0];
    const cells = Array.from(row.querySelectorAll("td")).map((c) => c.textContent);
    // 診療 ← examination、フード ← food（バケット折り畳み）
    expect(cells).toContain("¥2,667");
    expect(cells).toContain("¥1,333");
    // セル数 = 固定15列（0部門のセルも残る）
    expect(cells).toHaveLength(EXPECTED_HEADERS.length);
    // 領収No = billingId の 8 桁ゼロ埋め
    expect(within(row).getByText("00000009")).toBeInTheDocument();
  });

  it("入院会計の（入院）表記は集約後も維持される", () => {
    render(
      <BillingDetailTable
        details={[
          detail({ billingId: 3, isHospitalization: true, categoryAmounts: { hotel: 5000 } }),
        ]}
      />,
    );
    expect(screen.getByText(/入院/)).toBeInTheDocument();
  });

  it("明細がない場合は「会計明細がありません」を出す（既存の空ケース維持）", () => {
    render(<BillingDetailTable details={[]} />);
    expect(screen.getByText("会計明細がありません")).toBeInTheDocument();
  });

  it("categoryAmounts 欠落（旧ペイロード）は部門セルに記録なしを出す（MIN(category) へは寄せない）", () => {
    render(<BillingDetailTable details={[detail({ billingId: 4 })]} />);
    // category=examination の MIN(category) 値を全額部門セルへ誤表示しないこと。
    // 代替表示として記録なしが出る。
    expect(screen.getAllByText("記録なし").length).toBeGreaterThan(0);
  });

  it("異なる billingId は別行のまま（過剰集約しない）", () => {
    render(
      <BillingDetailTable
        details={[
          detail({ billingId: 1, categoryAmounts: { examination: 5000 } }),
          detail({ billingId: 2, categoryAmounts: { surgery: 5000 } }),
        ]}
      />,
    );
    expect(bodyRows(screen.getByRole("table"))).toHaveLength(2);
  });
});
