import { describe, it, expect } from "vitest";
import { render, screen, within } from "@testing-library/react";
import { DailyPrintArea } from "./DailyAccountingPrintArea";
import type { Accounting } from "../types";
import type { DetailedBreakdown, RowData, TotalsData } from "../lib/daily-accounting-utils";

const ZERO_BREAKDOWN: DetailedBreakdown = {
  medical: { cash: 0, card: 0 },
  surgery: { cash: 0, card: 0 },
  rv: { cash: 0, card: 0 },
  food: { cash: 0, card: 0 },
  trimming: { cash: 0, card: 0 },
  hotel: { cash: 0, card: 0 },
  goods: { cash: 0, card: 0 },
};

function accounting(
  overrides: Partial<Accounting> & Pick<Accounting, "id" | "ownerName" | "petName">,
): Accounting {
  return {
    clinicId: "1",
    ownerId: "1",
    petId: "1",
    status: "completed",
    scheduledDate: "2026-07-01",
    items: [],
    totalRefundedAmount: 0,
    payment: {
      subtotal: 0,
      taxTotal: 0,
      totalAmount: 0,
      insuranceAmount: 0,
      discountAmount: 0,
      billingAmount: 0,
      receivedAmount: 0,
      changeAmount: 0,
      method: "cash",
    },
    ...overrides,
  };
}

function row(overrides: Partial<RowData> & Pick<RowData, "accounting" | "total">): RowData {
  return {
    breakdown: { medical: 0, surgery: 0, rv: 0, food: 0, trimming: 0, hotel: 0, goods: 0 },
    detailedBreakdown: ZERO_BREAKDOWN,
    subtotal: 0,
    tax: 0,
    discount: 0,
    ...overrides,
  };
}

const ROWS: RowData[] = [
  row({
    accounting: accounting({ id: "1", ownerName: "田中太郎", petName: "ポチ" }),
    total: 12345,
  }),
  row({ accounting: accounting({ id: "2", ownerName: "鈴木花子", petName: "タマ" }), total: 0 }),
];

// 病院合計 = medical(10000) + surgery(0) + rv(5000) + food(0) + goods(2000) = 17000
// トリミング合計 = trimming(3000) + hotel(500) = 3500
const TOTALS: TotalsData = {
  medical: 10000,
  surgery: 0,
  rv: 5000,
  food: 0,
  trimming: 3000,
  hotel: 500,
  goods: 2000,
  subtotal: 0,
  tax: 0,
  discount: 0,
  total: 99999,
};

describe("DailyAccountingPrintArea: 金額セルの印字が固定されている (FE5-14)", () => {
  it("明細行の合計は 0 円でも ¥0 と表示する（ガードなしパターン）", () => {
    render(<DailyPrintArea date="2026-07-01" rows={ROWS} totals={TOTALS} />);
    const area = screen.getByTestId("daily-print-area");
    expect(within(area).getByText("¥12,345")).toBeInTheDocument();
    expect(within(area).getByText("¥0")).toBeInTheDocument();
  });

  it("病院合計行は 0 円科目を「-」にし、合計は ¥ 区切りで表示する", () => {
    render(<DailyPrintArea date="2026-07-01" rows={ROWS} totals={TOTALS} />);
    const area = screen.getByTestId("daily-print-area");
    const hospitalRow = within(area).getByText("病院合計").closest("tr")!;
    expect(within(hospitalRow).getByText("¥10,000")).toBeInTheDocument();
    expect(within(hospitalRow).getByText("¥5,000")).toBeInTheDocument();
    expect(within(hospitalRow).getByText("¥17,000")).toBeInTheDocument();
    expect(within(hospitalRow).queryByText("¥0")).not.toBeInTheDocument();
  });

  it("明細行の診療セルが負でも符号のまま印字する", () => {
    const negativeRow = row({
      accounting: accounting({ id: "3", ownerName: "赤伝", petName: "ハナ" }),
      total: -3000,
      breakdown: { medical: -3000, surgery: 0, rv: 0, food: 0, trimming: 0, hotel: 0, goods: 0 },
      detailedBreakdown: {
        ...ZERO_BREAKDOWN,
        medical: { cash: -3000, card: 0 },
      },
    });
    render(
      <DailyPrintArea
        date="2026-07-01"
        rows={[negativeRow]}
        totals={{ ...TOTALS, medical: -3000, total: -3000 }}
      />,
    );
    const area = screen.getByTestId("daily-print-area");
    const dataRow = within(area).getByText("赤伝").closest("tr")!;
    const cells = dataRow.querySelectorAll("td");
    // EMR-186: 飼主名・ペット名が最右端へ移動したため診療列は index 1
    expect(cells[1]).toHaveTextContent("¥-3,000");
  });

  it("科目合計が負でも符号のまま印字する", () => {
    const negativeTotals: TotalsData = {
      ...TOTALS,
      medical: -3000,
      surgery: 0,
      rv: 0,
      food: 0,
      goods: 0,
      trimming: 0,
      hotel: 0,
      total: -3000,
    };
    render(<DailyPrintArea date="2026-07-01" rows={ROWS} totals={negativeTotals} />);
    const area = screen.getByTestId("daily-print-area");
    const hospitalRow = within(area).getByText("病院合計").closest("tr")!;
    const cells = hospitalRow.querySelectorAll("td");
    expect(cells[1]).toHaveTextContent("¥-3,000");
    expect(within(hospitalRow).queryByText("¥0")).not.toBeInTheDocument();
  });

  it("トリミング合計行は科目金額と合計を ¥ 区切りで表示する", () => {
    render(<DailyPrintArea date="2026-07-01" rows={ROWS} totals={TOTALS} />);
    const area = screen.getByTestId("daily-print-area");
    const trimmingRow = within(area).getByText("トリミング合計").closest("tr")!;
    expect(within(trimmingRow).getByText("¥3,000")).toBeInTheDocument();
    expect(within(trimmingRow).getByText("¥500")).toBeInTheDocument();
    expect(within(trimmingRow).getByText("¥3,500")).toBeInTheDocument();
  });

  it("全体合計行の総合計は ¥ 区切りで表示する", () => {
    render(<DailyPrintArea date="2026-07-01" rows={ROWS} totals={TOTALS} />);
    const area = screen.getByTestId("daily-print-area");
    const grandRow = within(area).getByText("全体合計").closest("tr")!;
    expect(within(grandRow).getByText("¥99,999")).toBeInTheDocument();
  });
});

describe("EMR-205 再発防止: hidden 属性ではなく hidden/print:block クラス + data-print-portal 固定キー", () => {
  it("印刷面ルートは hidden HTML 属性を持たず hidden と print:block クラスを持つ", () => {
    render(<DailyPrintArea date="2026-07-01" rows={ROWS} totals={TOTALS} />);
    const area = screen.getByTestId("daily-print-area");
    // Tailwind v4 preflight の [hidden] display:none !important（@layer base）は
    // unlayered な display:block !important より強いため、hidden 属性が残ると白紙化する。
    expect(area).not.toHaveAttribute("hidden");
    expect(area).toHaveClass("hidden");
    expect(area).toHaveClass("print:block");
  });

  it("印刷面は固定キー data-print-portal を持ち、注入CSSは per-instance な testId を参照しない", () => {
    render(<DailyPrintArea date="2026-07-01" rows={ROWS} totals={TOTALS} />);
    const area = screen.getByTestId("daily-print-area");
    // #187: 除外キーは全ポータル共通の data-print-portal。per-instance な
    // data-testid 除外は同居ポータル同士を隠し合って白紙化するため禁止。
    expect(area).toHaveAttribute("data-print-portal");
    const css = area.querySelector("style")?.textContent ?? "";
    expect(css).not.toContain("data-testid");
    expect(css).toContain(":not([data-print-portal])");
  });

  it("@page は A4 landscape を維持する", () => {
    render(<DailyPrintArea date="2026-07-01" rows={ROWS} totals={TOTALS} />);
    const area = screen.getByTestId("daily-print-area");
    const css = area.querySelector("style")?.textContent ?? "";
    expect(css).toContain("size: A4 landscape");
  });
});

describe("EMR-186: 列順 — 飼主名・ペット名が最右端の2列（この順）", () => {
  it("ヘッダーと明細行の最右端2列が 飼主名→ペット名 の順である", () => {
    render(<DailyPrintArea date="2026-07-01" rows={ROWS} totals={TOTALS} />);
    const area = screen.getByTestId("daily-print-area");
    const headerTexts = Array.from(area.querySelectorAll("thead th")).map((h) => h.textContent);
    // 最右端2列は 飼主名 → ペット名、合計列はその左隣、先頭は領収No
    expect(headerTexts[0]).toBe("領収No");
    expect(headerTexts.slice(-2)).toEqual(["飼主名", "ペット名"]);
    expect(headerTexts[headerTexts.length - 3]).toBe("合計");

    const dataRow = within(area).getByText("田中太郎").closest("tr")!;
    const cellTexts = Array.from(dataRow.querySelectorAll("td")).map((c) => c.textContent);
    expect(cellTexts.slice(-2)).toEqual(["田中太郎", "ポチ"]);
    expect(cellTexts[cellTexts.length - 3]).toBe("¥12,345");
  });

  it("各集計行（病院/トリミング/全体合計）のセル数がヘッダー列数と一致する", () => {
    render(<DailyPrintArea date="2026-07-01" rows={ROWS} totals={TOTALS} />);
    const area = screen.getByTestId("daily-print-area");
    const headerCount = area.querySelectorAll("thead th").length;
    for (const label of ["病院合計", "トリミング合計", "全体合計"]) {
      const tr = within(area).getByText(label).closest("tr")!;
      const span = Array.from(tr.querySelectorAll("td")).reduce(
        (sum, td) => sum + (Number(td.getAttribute("colspan")) || 1),
        0,
      );
      expect(span).toBe(headerCount);
    }
  });
});
