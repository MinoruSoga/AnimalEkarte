import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { C, STYLE } from "@/lib/design-tokens";
import { DataTable, DESIGN_TABLE_HEADER_ROW, DESIGN_TABLE_HEADER_CELL } from "./DataTable";

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
    // 置換（併記ではない）ことを保証: 既定の tableHeaderCell 由来のクラスは残らない
    expect(headerCell.className).not.toContain(STYLE.tableHeaderCell.split(" ")[0]);

    const headerRow = headerCell.closest("tr");
    expect(headerRow?.className).toContain("custom-row-class");
  });

  it("DESIGN_TABLE_HEADER_ROW/CELL は ex-data-table-cell（canvas-soft + sectionLabel）を構成する", () => {
    expect(DESIGN_TABLE_HEADER_ROW).toContain(C.bgPage);
    expect(DESIGN_TABLE_HEADER_ROW).toContain(C.borderLight);
    for (const cls of STYLE.sectionLabel.split(" ")) {
      expect(DESIGN_TABLE_HEADER_CELL).toContain(cls);
    }
  });
});
