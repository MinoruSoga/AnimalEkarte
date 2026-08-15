import { describe, it, expect } from "vitest";
import { render, screen, within } from "@testing-library/react";
import { UnifiedClosingSummaryTable } from "./UnifiedClosingSummaryTable";
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

const PAYMENT_METHODS = [pm(1, "現金", 1), pm(2, "クレジットカード", 2)];

describe("UnifiedClosingSummaryTable (#153 統合テーブル)", () => {
  it("部門 / 件数 / 支払方法別金額 / 合計 を 1 表で統合表示する", () => {
    const categories = {
      examination: { 現金: 3000, クレジットカード: 2000 },
      medicine: { 現金: 1000 },
      surgery: { 現金: 5000 },
    };
    const billingDetails = [
      detail({ billingId: 1, category: "examination" }),
      detail({ billingId: 2, category: "medicine" }),
      detail({ billingId: 3, category: "surgery" }),
    ];

    render(
      <UnifiedClosingSummaryTable
        categories={categories}
        paymentMethods={PAYMENT_METHODS}
        billingDetails={billingDetails}
      />,
    );

    // 統合された列ヘッダ（件数 を含む）
    expect(screen.getByText("件数")).toBeInTheDocument();

    // 診療行: 件数2件 / 合計¥6,000
    const medicalRow = screen.getByText("診療").closest("tr")!;
    expect(within(medicalRow).getByText("2件")).toBeInTheDocument();
    expect(within(medicalRow).getByText("¥6,000")).toBeInTheDocument();

    // 外科行: 件数1件 / 全額現金のため 現金列・合計列の双方に ¥5,000
    const surgeryRow = screen.getByText("外科").closest("tr")!;
    expect(within(surgeryRow).getByText("1件")).toBeInTheDocument();
    expect(within(surgeryRow).getAllByText("¥5,000")).toHaveLength(2);

    // フッタ総合計: 件数3件 / ¥11,000
    const footer = screen.getByText("診療").closest("table")!.querySelector("tfoot")!;
    expect(within(footer).getByText("3件")).toBeInTheDocument();
    expect(within(footer).getByText("¥11,000")).toBeInTheDocument();
  });

  it("データが無い場合は空メッセージを表示する", () => {
    render(
      <UnifiedClosingSummaryTable categories={{}} paymentMethods={PAYMENT_METHODS} billingDetails={[]} />,
    );
    expect(screen.getByText("対象期間の会計データがありません")).toBeInTheDocument();
  });

  it("DEC-40: 未分類・要確認は独立件数を表示し、旧 snapshot は記録なし", () => {
    const { rerender } = render(
      <UnifiedClosingSummaryTable
        categories={{ other: { 現金: 1000 } }}
        paymentMethods={PAYMENT_METHODS}
        billingDetails={[detail({ billingId: 1, category: "other" }), detail({ billingId: 1, category: "other" })]}
        unclassifiedOtherCount={1}
      />,
    );
    const otherRow = screen.getByText("未分類・要確認").closest("tr")!;
    // billingDetails は 2 行でも独立件数 1
    expect(within(otherRow).getByText("1件")).toBeInTheDocument();
    expect(screen.queryByText("その他")).not.toBeInTheDocument();

    rerender(
      <UnifiedClosingSummaryTable
        categories={{ other: { 現金: 1000 } }}
        paymentMethods={PAYMENT_METHODS}
        billingDetails={[]}
        unclassifiedOtherCount={null}
      />,
    );
    const unrecordedRow = screen.getByText("未分類・要確認").closest("tr")!;
    expect(within(unrecordedRow).getByText("記録なし")).toBeInTheDocument();
  });
});

