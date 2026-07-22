import { render, screen, within } from "@testing-library/react";
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
