import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { DataTableRow } from "./DataTableRow";

function renderRow(onLegacyRowClick?: () => void) {
  const legacyInteractionProps = onLegacyRowClick ? { onClick: onLegacyRowClick } : {};
  render(
    <table>
      <tbody>
        <DataTableRow {...legacyInteractionProps}>
          <td>テスト行</td>
        </DataTableRow>
      </tbody>
    </table>,
  );

  return screen.getByRole("row");
}

describe("DataTableRow noninteractive semantics", () => {
  it("常にdefault cursorを表示し interactive semantics を持たない", () => {
    const row = renderRow();

    expect(row).toHaveClass("cursor-default");
    expect(row).not.toHaveClass("cursor-pointer");
    expect(row).not.toHaveAttribute("role");
    expect(row).not.toHaveAttribute("tabindex");
    expect(row).not.toHaveAttribute("onclick");
  });

  it("legacy onClick が渡されても行clickを発火せず非interactiveのままにする", () => {
    const onLegacyRowClick = vi.fn();
    const row = renderRow(onLegacyRowClick);

    fireEvent.click(screen.getByText("テスト行"));

    expect(onLegacyRowClick).not.toHaveBeenCalled();
    expect(row).toHaveClass("cursor-default");
    expect(row).not.toHaveClass("cursor-pointer");
  });
});
