import { describe, expect, it } from "vitest";

import type { ShiftTemplate } from "../types";
import { filterShiftTemplates } from "./shift-template-table-model";

function template(overrides: Partial<ShiftTemplate> = {}): ShiftTemplate {
  return {
    id: "1",
    clinic_id: "10",
    name: "午前勤務",
    shift_type: "morning",
    start_time: "09:00",
    end_time: "13:00",
    notes: "",
    sort_order: 1,
    is_active: true,
    breaks: [],
    created_at: "",
    updated_at: "",
    ...overrides,
  };
}

describe("filterShiftTemplates", () => {
  it("名称と種別で検索し、無効だけを残せる", () => {
    const items = [
      template({ id: "1", name: "午前勤務", shift_type: "morning", is_active: true }),
      template({ id: "2", name: "休日", shift_type: "off", is_active: false }),
    ];

    expect(filterShiftTemplates(items, "午前", []).map((item) => item.id)).toEqual(["1"]);
    expect(filterShiftTemplates(items, "休日", []).map((item) => item.id)).toEqual(["2"]);
    expect(
      filterShiftTemplates(items, "", [
        { key: "status", condition: "is", value: "inactive", displayValue: "無効" },
      ]).map((item) => item.id),
    ).toEqual(["2"]);
  });
});
