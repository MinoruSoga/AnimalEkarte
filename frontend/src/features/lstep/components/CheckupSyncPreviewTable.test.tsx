import { fireEvent, render, screen, within } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import type { CheckupSyncPreviewOwner } from "../api/get-checkup-sync-preview";
import { CheckupSyncPreviewTable } from "./CheckupSyncPreviewTable";

const owner: CheckupSyncPreviewOwner = {
  owner_id: "owner-1",
  owner_name: "山田 太郎",
  pet_names: ["ポチ"],
  last_visit_date: "2026-07-01",
  has_line: true,
  is_opt_out: false,
  has_living_pet: true,
  exclusion_reason: null,
  current_tags: [],
  min_pet_age_years: 3,
  max_pet_age_years: 3,
  has_chronic_condition: false,
  cpm_stage: "",
  total_amount: 0,
  annual_visit_count: 1,
  last_checkup_date: null,
};

describe("CheckupSyncPreviewTable DESIGN.md table contract", () => {
  it("header/body cell は typography と 12px 16px padding に一致する", () => {
    render(
      <CheckupSyncPreviewTable
        owners={[owner]}
        selectedIds={new Set()}
        onSelectionChange={vi.fn()}
        eligibleCount={1}
        lineLinkedCount={1}
        optOutCount={0}
        noLivingPetCount={0}
        totalCount={1}
      />,
    );

    const table = screen.getByRole("table");
    for (const header of within(table).getAllByRole("columnheader")) {
      expect(header).toHaveClass("text-2xs", "font-semibold", "px-4", "py-3");
    }
    for (const cell of within(table).getAllByRole("cell")) {
      expect(cell).toHaveClass("px-4", "py-3");
    }
  });
});

describe("CheckupSyncPreviewTable selection limit", () => {
  function createOwners(count: number): CheckupSyncPreviewOwner[] {
    return Array.from({ length: count }, (_, index) => ({
      ...owner,
      owner_id: `owner-${index + 1}`,
      owner_name: `飼主 ${index + 1}`,
    }));
  }

  it("select all chooses at most 100 eligible owners", () => {
    const onSelectionChange = vi.fn();
    render(
      <CheckupSyncPreviewTable
        owners={createOwners(101)}
        selectedIds={new Set()}
        onSelectionChange={onSelectionChange}
        eligibleCount={101}
        lineLinkedCount={101}
        optOutCount={0}
        noLivingPetCount={0}
        totalCount={101}
      />,
    );

    fireEvent.click(screen.getByRole("checkbox", { name: "送信可能対象をすべて選択" }));

    const selected = onSelectionChange.mock.calls[0]?.[0] as Set<string> | undefined;
    expect(selected).toBeInstanceOf(Set);
    expect(selected?.size).toBe(100);
    expect(selected?.has("owner-101")).toBe(false);
  });

  it("disables an unselected owner after 100 owners are selected", () => {
    const owners = createOwners(101);
    const selectedIds = new Set(owners.slice(0, 100).map(({ owner_id }) => owner_id));
    render(
      <CheckupSyncPreviewTable
        owners={owners}
        selectedIds={selectedIds}
        onSelectionChange={vi.fn()}
        eligibleCount={101}
        lineLinkedCount={101}
        optOutCount={0}
        noLivingPetCount={0}
        totalCount={101}
      />,
    );

    expect(screen.getByRole("checkbox", { name: "飼主 101を選択" })).toBeDisabled();
    expect(screen.getByText("一度に選択できるのは最大100名です")).toBeInTheDocument();
  });
});
