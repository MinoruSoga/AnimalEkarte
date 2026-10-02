import { DndContext } from "@dnd-kit/core";
import { SortableContext, verticalListSortingStrategy } from "@dnd-kit/sortable";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { Table, TableBody } from "@/components/ui/table";

import type { ShiftTemplate } from "../types";
import { ShiftTemplateRow, ShiftTemplateSidePanel } from "./ShiftTemplateSettingsParts";

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

  // EMR-241: 時刻表示はカテゴリ名ではなく時刻の有無だけで決まる
  it('off テンプレートに時刻があれば一覧にも "-" ではなく時刻を表示する', () => {
    const item = template({ shift_type: "off", start_time: "10:00", end_time: "15:00" });

    render(
      <DndContext>
        <SortableContext items={[item.id]} strategy={verticalListSortingStrategy}>
          <Table>
            <TableBody>
              <ShiftTemplateRow item={item} canEdit onEdit={vi.fn()} />
            </TableBody>
          </Table>
        </SortableContext>
      </DndContext>,
    );

    expect(screen.getByText("10:00〜15:00")).toBeInTheDocument();
  });
});

// EMR-241: カテゴリ名による時刻フィールド非表示を廃止。全種別で入力を表示し、
// off/paid_leave は空欄保存可・片方だけの入力は全種別で拒否する。
describe("ShiftTemplateSidePanel EMR-241 time fields", () => {
  function renderPanel(item: ShiftTemplate, onSave = vi.fn()) {
    render(
      <ShiftTemplateSidePanel item={item} onClose={vi.fn()} onSave={onSave} isSaving={false} />,
    );
    return { onSave };
  }

  it("off でも開始/終了時刻と休憩入力を表示する", () => {
    renderPanel(template({ shift_type: "off", start_time: "", end_time: "" }));

    expect(screen.getByLabelText("開始時刻")).toBeInTheDocument();
    expect(screen.getByLabelText("終了時刻")).toBeInTheDocument();
    expect(screen.getByText("休憩時間")).toBeInTheDocument();
  });

  it("paid_leave でも開始/終了時刻を表示する", () => {
    renderPanel(template({ shift_type: "paid_leave", start_time: "", end_time: "" }));

    expect(screen.getByLabelText("開始時刻")).toBeInTheDocument();
    expect(screen.getByLabelText("終了時刻")).toBeInTheDocument();
  });

  it("off + 両方空の時刻で保存できる", async () => {
    const user = userEvent.setup();
    const { onSave } = renderPanel(template({ shift_type: "off", start_time: "", end_time: "" }));

    await user.click(screen.getByRole("button", { name: "保存" }));

    expect(onSave).toHaveBeenCalledTimes(1);
    expect(onSave).toHaveBeenCalledWith(
      expect.objectContaining({ shift_type: "off", start_time: "", end_time: "" }),
    );
  });

  it("paid_leave + 両方空の時刻で保存できる", async () => {
    const user = userEvent.setup();
    const { onSave } = renderPanel(
      template({ shift_type: "paid_leave", start_time: "", end_time: "" }),
    );

    await user.click(screen.getByRole("button", { name: "保存" }));

    expect(onSave).toHaveBeenCalledTimes(1);
    expect(onSave).toHaveBeenCalledWith(
      expect.objectContaining({ shift_type: "paid_leave", start_time: "", end_time: "" }),
    );
  });

  it("勤務種別（full）で時刻を空のまま保存するとエラーで保存しない", async () => {
    const user = userEvent.setup();
    const { onSave } = renderPanel(template({ shift_type: "full", start_time: "", end_time: "" }));

    await user.click(screen.getByRole("button", { name: "保存" }));

    expect(
      await screen.findByText("勤務種別では開始時刻と終了時刻を入力してください"),
    ).toBeInTheDocument();
    expect(onSave).not.toHaveBeenCalled();
  });

  it("off で片方だけの時刻入力は拒否される", async () => {
    const user = userEvent.setup();
    const { onSave } = renderPanel(template({ shift_type: "off", start_time: "", end_time: "" }));

    await user.type(screen.getByLabelText("開始時刻"), "09:00");
    await user.click(screen.getByRole("button", { name: "保存" }));

    expect(await screen.findByText("開始時刻と終了時刻を入力してください")).toBeInTheDocument();
    expect(onSave).not.toHaveBeenCalled();
  });
});
