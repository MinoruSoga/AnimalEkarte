import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { C, STYLE } from "@/lib/design-tokens";
import { DataTable, DESIGN_TABLE_HEADER_ROW, DESIGN_TABLE_HEADER_CELL } from "./DataTable";
import { LIST_TABLE_COL } from "./list-table-col";

interface Row {
  id: string;
  name: string;
}

const rows: Row[] = [{ id: "1", name: "テスト行" }];

describe("DataTable", () => {
  it("headerRowClassName/headerCellClassName 未指定時は既定の STYLE トークンを使う", () => {
    render(
      <DataTable
        columns={[{ header: "名前" }]}
        data={rows}
        renderRow={(row) => (
          <tr key={row.id}>
            <td>{row.name}</td>
          </tr>
        )}
      />
    );

    const headerCell = screen.getByRole("columnheader", { name: "名前" });
    for (const cls of STYLE.tableHeaderCell.split(" ")) {
      expect(headerCell.className).toContain(cls);
    }
  });

  it("headerRowClassName/headerCellClassName 指定時は既定を置き換える（DESIGN.md ex-data-table-cell 準拠の上書き）", () => {
    render(
      <DataTable
        columns={[{ header: "名前" }]}
        data={rows}
        renderRow={(row) => (
          <tr key={row.id}>
            <td>{row.name}</td>
          </tr>
        )}
        headerRowClassName="custom-row-class"
        headerCellClassName="custom-cell-class"
      />
    );

    const headerCell = screen.getByRole("columnheader", { name: "名前" });
    expect(headerCell.className).toContain("custom-cell-class");
    // 置換（併記ではない）ことを保証: 既定の DESIGN_TABLE_HEADER_CELL（sectionLabel）由来の uppercase は残らない
    // （FE10: TableHead 基底が text-2xs を持つため、基底クラスでなくトークン固有クラスで判定する）
    expect(headerCell.className).not.toContain("uppercase");

    const headerRow = headerCell.closest("tr");
    expect(headerRow?.className).toContain("custom-row-class");
  });

  it("table は固定 min-w-[640px] を持たず min-w-0 で狭幅スクロール可能 (BUG-458)", () => {
    const { container } = render(
      <DataTable
        columns={[{ header: "名前" }]}
        data={rows}
        renderRow={(row) => (
          <tr key={row.id}>
            <td>{row.name}</td>
          </tr>
        )}
      />,
    );
    const scroll = container.querySelector(".overflow-auto");
    expect(scroll).toHaveClass("min-w-0");
    const table = container.querySelector("table");
    expect(table?.className ?? "").not.toContain("min-w-[640px]");
    expect(table).toHaveClass("min-w-0");
  });

  it("DESIGN_TABLE_HEADER_ROW/CELL は ex-data-table-cell（canvas-soft + sectionLabel）を構成する", () => {
    expect(DESIGN_TABLE_HEADER_ROW).toContain(C.bgPage);
    expect(DESIGN_TABLE_HEADER_ROW).toContain(C.borderLight);
    expect(DESIGN_TABLE_HEADER_ROW).toContain("h-11");
    expect(DESIGN_TABLE_HEADER_CELL).not.toContain("h-11");
    for (const cls of STYLE.sectionLabel.split(" ")) {
      expect(DESIGN_TABLE_HEADER_CELL).toContain(cls);
    }
  });

  it("LIST_TABLE_COL.status はステータス列の1行表示用に min-width と nowrap を持つ (BUG-020)", () => {
    expect(LIST_TABLE_COL.status).toContain("w-[100px]");
    expect(LIST_TABLE_COL.status).toContain("min-w-[100px]");
    expect(LIST_TABLE_COL.status).toContain("whitespace-nowrap");
  });

  it("column className に LIST_TABLE_COL.status を渡すとヘッダへ nowrap/min-width が載る (BUG-020)", () => {
    render(
      <DataTable
        columns={[{ header: "ステータス", className: LIST_TABLE_COL.status }]}
        data={rows}
        renderRow={(row) => (
          <tr key={row.id}>
            <td>{row.name}</td>
          </tr>
        )}
      />,
    );

    const headerCell = screen.getByRole("columnheader", { name: "ステータス" });
    expect(headerCell.className).toContain("w-[100px]");
    expect(headerCell.className).toContain("min-w-[100px]");
    expect(headerCell.className).toContain("whitespace-nowrap");
  });
});
