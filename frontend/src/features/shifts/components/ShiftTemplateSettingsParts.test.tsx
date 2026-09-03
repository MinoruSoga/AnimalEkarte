import { DndContext } from "@dnd-kit/core";
import { SortableContext, verticalListSortingStrategy } from "@dnd-kit/sortable";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { Table, TableBody } from "@/components/ui/table";

import type { ShiftTemplate } from "../types";
import { ShiftTemplateRow } from "./ShiftTemplateSettingsParts";

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

describe("ShiftTemplateRow", () => {
  it("名称クリックで編集を開く", async () => {
    const user = userEvent.setup();
    const onEdit = vi.fn();
    const item = template();

    render(
      <DndContext>
        <SortableContext items={[item.id]} strategy={verticalListSortingStrategy}>
          <Table>
            <TableBody>
              <ShiftTemplateRow item={item} canEdit onEdit={onEdit} />
            </TableBody>
          </Table>
        </SortableContext>
      </DndContext>,
    );

    await user.click(
      screen.getByRole("button", { name: "詳細: シフトテンプレート 午前勤務 (ID 1)" }),
    );

    expect(onEdit).toHaveBeenCalledTimes(1);
  });
});
