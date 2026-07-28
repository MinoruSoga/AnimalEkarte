import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { ReservationTypeSettingsContent } from "./ReservationTypeSettingsContent";

vi.mock("@/components/shared/PropertyFilter/PropertyFilter", () => ({
  PropertyFilter: () => <div data-testid="property-filter" />,
}));

vi.mock("./ReservationTypeGroupedTable", () => ({
  ReservationTypeGroupedTable: () => <div data-testid="grouped-table" />,
}));

describe("ReservationTypeSettingsContent", () => {
  it("グループ追加ボタンは44px以上の操作領域を保ち、追加操作を通知する", async () => {
    const user = userEvent.setup();
    const onGroupAdd = vi.fn();

    render(
      <ReservationTypeSettingsContent
        groups={[]}
        categories={[]}
        activeFilters={[]}
        onFilterChange={vi.fn()}
        searchTerm=""
        onSearchChange={vi.fn()}
        count={0}
        canCreate
        canEdit
        onCategoryEdit={vi.fn()}
        onGroupEdit={vi.fn()}
        onCategoryAddInGroup={vi.fn()}
        onGroupAdd={onGroupAdd}
      />,
    );

    const addButton = screen.getByRole("button", { name: "グループを追加" });
    expect(addButton).toHaveClass("min-h-11", "min-w-11");

    await user.click(addButton);
    expect(onGroupAdd).toHaveBeenCalledTimes(1);
  });
});
