import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { StaffNameCell } from "./StaffNameCell";

function renderCell(staff: string, isValidStaff: (name: string) => boolean) {
  return render(
    <table>
      <tbody>
        <tr>
          <StaffNameCell staff={staff} isValidStaff={isValidStaff} />
        </tr>
      </tbody>
    </table>,
  );
}

describe("StaffNameCell 担当者の識別 (EMR-412)", () => {
  it("空欄は検証や警告をせず - を表示する", () => {
    const isValidStaff = vi.fn(() => false);
    renderCell("", isValidStaff);

    expect(screen.getByRole("cell").textContent).toBe("-");
    expect(isValidStaff).not.toHaveBeenCalled();
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
    expect(screen.queryByText(/無効|担当者不明/)).not.toBeInTheDocument();
  });

  it("有効な担当者は氏名を表示する", () => {
    renderCell("田中 一郎", () => true);

    expect(screen.getByRole("cell")).toHaveTextContent(/^田中 一郎$/);
  });

  it("無効になった担当者は氏名と無効の文字を残す", () => {
    renderCell("田中 一郎", () => false);

    expect(screen.getByRole("cell")).toHaveTextContent("田中 一郎（無効）");
  });

  it.each(["123", "担当者不明"])("未解決の担当者 %s は無効扱いと区別する", (staff) => {
    renderCell(staff, () => false);

    expect(screen.getByRole("cell")).toHaveTextContent(/^担当者不明$/);
    expect(screen.queryByText(/無効/)).not.toBeInTheDocument();
  });
});
