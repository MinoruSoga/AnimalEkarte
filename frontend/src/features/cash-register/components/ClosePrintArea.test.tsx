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
  it("印刷エリアが DOM に存在し ヘッダ（医院名・対象日・区分）を含む", () => {
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
    // EMR-354: 飼主名/ペット名は独立セルになったため分離して検証する。
    const detailsTable = within(area).getByText("領収No").closest("table")!;
    expect(within(detailsTable).getAllByText("田中太郎")).toHaveLength(2);
    expect(within(detailsTable).getAllByText("ポチ")).toHaveLength(2);
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

  it("normal house_call and goods remain separate in printed summary and detail footers", () => {
    renderPrint({
      categories: { house_call: { 現金: 1000 }, goods: { 現金: 200 } },
      categoryCounts: { house_call: 1, goods: 1 },
      billingDetails: [
        detail({
          billingId: 20,
          ownerName: "往診飼主",
          petName: "往診ペット",
          category: "house_call",
          billingAmount: 1000,
          refundAmount: 0,
          netAmount: 1000,
          categoryAmounts: { house_call: 1000 },
        }),
        detail({
          billingId: 21,
          ownerName: "物販飼主",
          petName: "物販ペット",
          category: "goods",
          billingAmount: 200,
          refundAmount: 0,
          netAmount: 200,
          categoryAmounts: { goods: 200 },
        }),
      ],
      theoreticalCash: 1200,
    });
    const area = screen.getByTestId("close-print-area");
    const summaryTable = within(area).getByText("件数").closest("table")!;
    const summaryHouseCallRow = within(summaryTable).getByText("往診").closest("tr")!;
    const summaryGoodsRow = within(summaryTable).getByText("用品").closest("tr")!;
    expect(within(summaryHouseCallRow).getByText("1件")).toBeInTheDocument();
    expect(within(summaryHouseCallRow).getAllByText("¥1,000")).toHaveLength(2);
    expect(within(summaryGoodsRow).getByText("1件")).toBeInTheDocument();
    expect(within(summaryGoodsRow).getAllByText("¥200")).toHaveLength(2);

    const summaryFooter = within(summaryTable).getAllByRole("row").at(-1)!;
    expect(within(summaryFooter).getByText("2件")).toBeInTheDocument();
    expect(within(summaryFooter).getAllByText("¥1,200")).toHaveLength(2);

    const detailsTable = within(area).getByText("領収No").closest("table")!;
    const headers = within(detailsTable)
      .getAllByRole("columnheader")
      .map((header) => header.textContent);
    const houseCallIndex = headers.indexOf("往診");
    const goodsIndex = headers.indexOf("用品他");
    expect(houseCallIndex).toBe(goodsIndex + 1);

    const houseCallDetail = within(detailsTable).getByText("往診飼主").closest("tr")!;
    const houseCallCells = within(houseCallDetail)
      .getAllByRole("cell")
      .map((cell) => cell.textContent);
    expect(houseCallCells[houseCallIndex]).toBe("¥1,000");
    expect(houseCallCells[goodsIndex]).toBe("-");

    const goodsDetail = within(detailsTable).getByText("物販飼主").closest("tr")!;
    const goodsCells = within(goodsDetail)
      .getAllByRole("cell")
      .map((cell) => cell.textContent);
    expect(goodsCells[houseCallIndex]).toBe("-");
    expect(goodsCells[goodsIndex]).toBe("¥200");

    const detailsFooter = within(detailsTable).getAllByRole("row").at(-1)!;
    expect(within(detailsFooter).getByText("¥1,000")).toBeInTheDocument();
    expect(within(detailsFooter).getByText("¥200")).toBeInTheDocument();
    expect(within(detailsFooter).getAllByText("¥1,200")).toHaveLength(2);
  });

  it("refund-only house_call summary and footer remain printable while detail rows are empty", () => {
    renderPrint({
      categories: { house_call: { 現金: -1000 } },
      categoryCounts: { house_call: 0 },
      billingDetails: [],
      theoreticalCash: -1000,
    });
    const area = screen.getByTestId("close-print-area");
    const summaryTable = within(area).getByText("件数").closest("table")!;
    const houseCallRow = within(summaryTable).getByText("往診").closest("tr")!;
    expect(within(houseCallRow).getByText("0件")).toBeInTheDocument();
    expect(within(houseCallRow).getAllByText("¥-1,000")).toHaveLength(2);

    const summaryFooter = within(summaryTable).getAllByRole("row").at(-1)!;
    expect(within(summaryFooter).getAllByText("¥-1,000")).toHaveLength(2);
    expect(within(area).getByText("会計明細がありません")).toBeInTheDocument();
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
    // EMR-354: 明細テーブルは「行 + フッタ合計」で同額が2回出るため個数で検証する。
    const detailsTable = within(area).getByText("領収No").closest("table")!;
    expect(within(detailsTable).getAllByText("¥8,000")).toHaveLength(2);
    expect(within(detailsTable).getAllByText("-¥1,234")).toHaveLength(2);
    expect(within(detailsTable).getAllByText("¥6,766")).toHaveLength(2);
    expect(within(area).getByText("¥54,321")).toBeInTheDocument();
    expect(within(area).getByText("¥5,432")).toBeInTheDocument();
    expect(within(area).getByText("¥1,000")).toBeInTheDocument();
    expect(within(area).getByText("¥80")).toBeInTheDocument();
    expect(within(area).getByText("¥77,777")).toBeInTheDocument();
  });
});

// EMR-354: 印刷面の個別会計明細も画面と同じ「会計1件=1行」投影を使う。
// RED 時点では領収No/日次バケット列・billingId 集約・フッタが未実装のため失敗する。
describe("ClosePrintArea — EMR-354 個別会計明細（日次バケット投影）", () => {
  it("同一 billingId の split 行は1行に集約される（飼主の出現は1回）", () => {
    renderPrint({
      categories: { examination: { 現金: 4000 } },
      billingDetails: [
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
      ],
    });
    const area = screen.getByTestId("close-print-area");
    // 現行バグ: split 2行 → 飼主表示が2回。集約後は1回。
    expect(within(area).getAllByText("田中太郎")).toHaveLength(1);
  });

  it("明細テーブルに領収Noを含む固定列ヘッダと、1回だけ集計されるフッタがある", () => {
    renderPrint({
      categories: { examination: { 現金: 4000, クレジットカード: 5000 } },
      billingDetails: [
        detail({
          billingId: 7,
          ownerName: "田中太郎",
          paymentMethodName: "現金",
          billingAmount: 1000,
          refundAmount: 1234,
          netAmount: -234,
          categoryAmounts: { examination: 667, food: 333 },
        }),
        detail({
          billingId: 7,
          ownerName: "田中太郎",
          paymentMethodName: "クレジットカード",
          billingAmount: 3000,
          refundAmount: 1234,
          netAmount: 1766,
          categoryAmounts: { examination: 2000, food: 1000 },
        }),
        detail({
          billingId: 9,
          ownerName: "佐藤次郎",
          paymentMethodName: "クレジットカード",
          billingAmount: 5000,
          refundAmount: 0,
          netAmount: 5000,
          categoryAmounts: { surgery: 5000 },
        }),
      ],
    });
    const area = screen.getByTestId("close-print-area");
    // 個別明細テーブルは「領収No」ヘッダで特定する（統合サマリー表は「件数」）
    const detailsTable = within(area).getByText("領収No").closest("table")!;

    // フッタ行（最終行）が可視セルを1回だけ再集計する:
    //   診療 2,667 / 外科 5,000 / フード 1,333 / 合計 9,000 / 返金 -1,234 / 純額 7,766
    const allRows = detailsTable.querySelectorAll<HTMLTableRowElement>("tr");
    const footerRow = allRows[allRows.length - 1];
    const footerCells = Array.from(footerRow.querySelectorAll("td, th")).map((c) => c.textContent);
    expect(footerCells).toContain("¥2,667");
    expect(footerCells).toContain("¥5,000");
    expect(footerCells).toContain("¥1,333");
    expect(footerCells).toContain("¥9,000");
    expect(footerCells).toContain("-¥1,234");
    expect(footerCells).toContain("¥7,766");
  });

  it("categoryAmounts 欠落（旧ペイロード）は部門セルに記録なしを出す", () => {
    renderPrint({
      categories: { examination: { 現金: 5000 } },
      billingDetails: [detail({ billingId: 4 })],
    });
    const area = screen.getByTestId("close-print-area");
    const detailsTable = within(area).getByText("領収No").closest("table")!;
    expect(within(detailsTable).getAllByText("記録なし").length).toBeGreaterThan(0);
  });

  it("mount 中は html に印刷ルートオーバーライド用クラスを付与し、アンマウントで除去する", () => {
    // R2-3: close 印刷が html/body の overflow/height を @media print で解除するための
    // グローバル DOM 副作用（close-print-active クラス）が、unmount 後に残存して
    // 無関係のビュー・印刷面へ漏れないことを保証する。CSS 文字列自体は検証しない。
    const root = document.documentElement;
    const { unmount } = renderPrint();
    expect(root.classList.contains("close-print-active")).toBe(true);
    unmount();
    expect(root.classList.contains("close-print-active")).toBe(false);
  });
});
