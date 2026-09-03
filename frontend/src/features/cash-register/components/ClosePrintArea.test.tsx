import { describe, it, expect } from "vitest";
import { render, screen, within } from "@testing-library/react";
import { ClosePrintArea } from "./ClosePrintArea";
import type { PaymentMethodMaster } from "@/types/generated/models";
import type { CloseBillingDetail } from "../api/get-cash-register-preview";

function pm(id: number, name: string, order: number): PaymentMethodMaster {
  return {
    id,
    clinic_id: 1,
    name,
    display_order: order,
    is_active: true,
    created_at: "",
    updated_at: "",
  };
}

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

const PAYMENT_METHODS = [pm(1, "現金", 1), pm(2, "クレジットカード", 2)];
const TAX = {
  standard: { taxableAmount: 5000, taxAmount: 500 },
  reduced: { taxableAmount: 1000, taxAmount: 80 },
};

function renderPrint(overrides: Partial<React.ComponentProps<typeof ClosePrintArea>> = {}) {
  return render(
    <ClosePrintArea
      date="2026-05-01"
      period="am"
      clinicName="テスト動物病院"
      categories={{ examination: { 現金: 3000, クレジットカード: 2000 }, surgery: { 現金: 5000 } }}
      paymentMethods={PAYMENT_METHODS}
      billingDetails={[
        detail({ billingId: 1, category: "examination" }),
        detail({ billingId: 2, category: "surgery", billingAmount: 5000, netAmount: 5000 }),
      ]}
      taxBreakdown={TAX}
      theoreticalCash={8000}
      actualCash={null}
      {...overrides}
    />,
  );
}

describe("ClosePrintArea (#153 印刷 / PDF出力)", () => {
  it("印刷エリアが DOM に存在し ヘッダ（病院名・対象日・区分）を含む", () => {
    renderPrint();
    const area = screen.getByTestId("close-print-area");
    expect(area).toBeInTheDocument();
    expect(within(area).getByText("テスト動物病院")).toBeInTheDocument();
    expect(area.textContent).toContain("2026-05-01");
    expect(area.textContent).toContain("午前");
  });

  it("統合テーブルの集計値が出力される（件数・合計）", () => {
    renderPrint();
    const area = screen.getByTestId("close-print-area");
    // 「件数」ヘッダを持つのは統合サマリー表のみ（個別明細表と区別する）
    const summaryTable = within(area).getByText("件数").closest("table")!;
    const medicalRow = within(summaryTable).getByText("診療").closest("tr")!;
    expect(within(medicalRow).getByText("1件")).toBeInTheDocument();
    // 個別会計明細（和集合の列）も同一面に含まれる（2 明細行）
    expect(within(area).getAllByText("田中太郎 / ポチ")).toHaveLength(2);
  });

  it("理論現金は常に出力し、実際現金未入力時は差額を出さない", () => {
    renderPrint({ actualCash: null });
    const area = screen.getByTestId("close-print-area");
    expect(within(area).getByText("理論現金")).toBeInTheDocument();
    expect(within(area).queryByText("差額")).not.toBeInTheDocument();
  });

  it("実際現金が入力された場合は差額を出力する", () => {
    renderPrint({ actualCash: 8500 });
    const area = screen.getByTestId("close-print-area");
    expect(within(area).getByText("実際の現金")).toBeInTheDocument();
    expect(within(area).getByText("差額")).toBeInTheDocument();
    // 8500 - 8000 = +500
    expect(within(area).getByText("+500")).toBeInTheDocument();
  });

  // FE5-14: yen() インライン定義を formatCurrency に置換する前の印字固定
  it("統合テーブルの部門合計・支払方法別金額は ¥ 区切りで出力する", () => {
    renderPrint({
      categories: { examination: { 現金: 12345, クレジットカード: 6789 } },
      billingDetails: [
        detail({
          billingId: 1,
          category: "examination",
          billingAmount: 19134,
          refundAmount: 0,
          netAmount: 19134,
        }),
      ],
      theoreticalCash: 19134,
      actualCash: null,
    });
    const area = screen.getByTestId("close-print-area");
    const summaryTable = within(area).getByText("件数").closest("table")!;
    const medicalRow = within(summaryTable).getByText("診療").closest("tr")!;
    expect(within(medicalRow).getByText("¥12,345")).toBeInTheDocument();
    expect(within(medicalRow).getByText("¥6,789")).toBeInTheDocument();
    expect(within(medicalRow).getByText("¥19,134")).toBeInTheDocument();
  });

  it("支払方法データがない部門の金額セルは — を出力する（既定データの外科×クレジットカード）", () => {
    renderPrint();
    const area = screen.getByTestId("close-print-area");
    const summaryTable = within(area).getByText("件数").closest("table")!;
    const surgeryRow = within(summaryTable).getByText("外科").closest("tr")!;
    expect(within(surgeryRow).getByText("—")).toBeInTheDocument();
  });

  it("個別会計明細の返金額は -¥ 表記、0円請求は ¥0、消費税内訳・現金照合も ¥ 区切りで出力する", () => {
    renderPrint({
      categories: { examination: { 現金: 8500 } },
      billingDetails: [
        detail({
          billingId: 1,
          category: "examination",
          billingAmount: 0,
          refundAmount: 0,
          netAmount: 0,
        }),
        detail({
          billingId: 2,
          category: "examination",
          billingAmount: 8000,
          refundAmount: 1234,
          netAmount: 6766,
        }),
      ],
      taxBreakdown: {
        standard: { taxableAmount: 54321, taxAmount: 5432 },
        reduced: { taxableAmount: 1000, taxAmount: 80 },
      },
      theoreticalCash: 77777,
      actualCash: null,
    });
    const area = screen.getByTestId("close-print-area");
    expect(within(area).getAllByText("¥0")).toHaveLength(2);
    expect(within(area).getByText("¥8,000")).toBeInTheDocument();
    expect(within(area).getByText("-¥1,234")).toBeInTheDocument();
    expect(within(area).getByText("¥6,766")).toBeInTheDocument();
    expect(within(area).getByText("¥54,321")).toBeInTheDocument();
    expect(within(area).getByText("¥5,432")).toBeInTheDocument();
    expect(within(area).getByText("¥1,000")).toBeInTheDocument();
    expect(within(area).getByText("¥80")).toBeInTheDocument();
    expect(within(area).getByText("¥77,777")).toBeInTheDocument();
  });
});
