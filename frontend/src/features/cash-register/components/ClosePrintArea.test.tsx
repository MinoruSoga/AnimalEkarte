import { describe, it, expect } from "vitest";
import { render, screen, within } from "@testing-library/react";
import { ClosePrintArea } from "./ClosePrintArea";
import type { PaymentMethodMaster } from "@/types/generated/models";
import type { CloseBillingDetail } from "../api/get-cash-register-preview";

function pm(id: number, name: string, order: number): PaymentMethodMaster {
  return { id, clinic_id: 1, name, display_order: order, is_active: true, created_at: "", updated_at: "" };
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
});
