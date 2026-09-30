import type { ReactNode } from "react";
import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import type { CheckupTypeFieldRow } from "@/hooks/use-checkup-fields";
import { CheckupTypeFieldsEditor } from "./CheckupTypeFieldsEditor";

type FieldsQueryState = Pick<
  ReturnType<typeof import("@/hooks/use-checkup-fields").useGetCheckupTypeFields>,
  "data" | "error" | "isError" | "isPending"
>;

const mocks = vi.hoisted(() => ({
  create: vi.fn(),
  update: vi.fn(),
  remove: vi.fn(),
  reorder: vi.fn(),
  resetOrder: vi.fn(),
  getFields: vi.fn<() => FieldsQueryState>(),
  reorderCallbacks: [] as Array<(ids: string[]) => void>,
}));

vi.mock("@dnd-kit/core", () => ({
  DndContext: ({ children }: { children: ReactNode }) => <>{children}</>,
  closestCenter: vi.fn(),
}));

vi.mock("@dnd-kit/sortable", () => ({
  SortableContext: ({ children }: { children: ReactNode }) => <>{children}</>,
  verticalListSortingStrategy: vi.fn(),
  useSortable: () => ({
    attributes: {},
    listeners: {},
    setNodeRef: vi.fn(),
    setActivatorNodeRef: vi.fn(),
    transform: null,
    transition: undefined,
    isDragging: false,
  }),
}));

vi.mock("@dnd-kit/utilities", () => ({
  CSS: { Transform: { toString: () => undefined } },
}));

vi.mock("@/hooks/use-sortable-list", () => ({
  useSortableList: ({
    items,
    onReorder,
  }: {
    items: Array<{ id: string }>;
    onReorder: (ids: string[]) => void;
  }) => {
    mocks.reorderCallbacks.push(onReorder);
    return {
      orderedItems: items,
      sensors: [],
      handleDragEnd: vi.fn(),
      resetOrder: mocks.resetOrder,
    };
  },
}));

vi.mock("@/hooks/use-checkup-fields", () => ({
  useGetCheckupTypeFields: () => mocks.getFields(),
}));

vi.mock("../api/checkup-type-fields", async (importOriginal) => {
  const actual = await importOriginal<typeof import("../api/checkup-type-fields")>();
  return {
    ...actual,
    useCreateCheckupTypeField: () => ({ mutateAsync: mocks.create }),
    useUpdateCheckupTypeField: () => ({ mutateAsync: mocks.update }),
    useDeleteCheckupTypeField: () => ({ mutate: mocks.remove }),
    useReorderCheckupTypeFields: () => ({ mutate: mocks.reorder }),
  };
});

const fieldRows: CheckupTypeFieldRow[] = [
  {
    id: 31,
    checkupTypeId: 7,
    name: "体重",
    fieldType: "number",
    unit: "kg",
    minValue: 1,
    maxValue: 80,
    options: [],
    isProvisional: false,
    sortOrder: 1,
  },
  {
    id: 32,
    checkupTypeId: 7,
    name: "総合評価",
    fieldType: "single_select",
    unit: "",
    options: [{ value: "a", label: "良好" }],
    isProvisional: false,
    sortOrder: 2,
  },
];

describe("CheckupTypeFieldsEditor", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mocks.reorderCallbacks.length = 0;
    mocks.create.mockResolvedValue(undefined);
    mocks.update.mockResolvedValue(undefined);
    mocks.getFields.mockReturnValue({
      data: fieldRows,
      error: null,
      isError: false,
      isPending: false,
    });
  });

  it("exposes accessible field create/edit/delete and reorder controls", async () => {
    const user = userEvent.setup();
    render(<CheckupTypeFieldsEditor checkupTypeId="7" canCreate canEdit canDelete />);

    expect(screen.getByRole("heading", { name: "健診項目" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "健診項目を追加" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "編集: 健診項目 体重 (ID 31)" })).toBeInTheDocument();
    expect(screen.getByText("数値")).toBeInTheDocument();
    expect(screen.getByText("kg")).toBeInTheDocument();

    // 削除は ConfirmDialog で確認後にのみ実行される（直行削除なし）。
    await user.click(screen.getByRole("button", { name: "削除: 健診項目 体重 (ID 31)" }));
    expect(mocks.remove).not.toHaveBeenCalled();
    const dialog = screen.getByRole("alertdialog", { name: "健診項目を削除しますか？" });
    expect(dialog).toHaveTextContent("「体重」を削除します");
    await user.click(within(dialog).getByRole("button", { name: "削除" }));
    expect(mocks.remove).toHaveBeenCalledWith({ checkupTypeId: "7", fieldId: "31" });

    mocks.reorderCallbacks.at(-1)?.(["32", "31"]);
    expect(mocks.reorder).toHaveBeenCalledWith(
      { checkupTypeId: "7", ids: [32, 31] },
      expect.objectContaining({ onError: expect.any(Function) }),
    );
    mocks.reorder.mock.calls[0]?.[1]?.onError();
    expect(mocks.resetOrder).toHaveBeenCalledTimes(1);
  });

  it("prevents all mutation entry points in read-only mode", () => {
    render(
      <CheckupTypeFieldsEditor
        checkupTypeId="7"
        canCreate={false}
        canEdit={false}
        canDelete={false}
      />,
    );

    expect(screen.queryByRole("button", { name: "健診項目を追加" })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /編集: 健診項目/ })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /削除: 健診項目/ })).not.toBeInTheDocument();
    mocks.reorderCallbacks.at(-1)?.(["32", "31"]);
    expect(mocks.reorder).not.toHaveBeenCalled();
  });

  it("does not delete a field when the confirm dialog is canceled", async () => {
    const user = userEvent.setup();
    render(<CheckupTypeFieldsEditor checkupTypeId="7" canCreate canEdit canDelete />);

    await user.click(screen.getByRole("button", { name: "削除: 健診項目 体重 (ID 31)" }));
    const dialog = screen.getByRole("alertdialog", { name: "健診項目を削除しますか？" });
    await user.click(within(dialog).getByRole("button", { name: "キャンセル" }));

    expect(mocks.remove).not.toHaveBeenCalled();
    await waitFor(() => {
      expect(screen.queryByRole("alertdialog")).not.toBeInTheDocument();
    });
  });

  it("confirms before removing a draft option row", async () => {
    const user = userEvent.setup();
    render(<CheckupTypeFieldsEditor checkupTypeId="7" canCreate canEdit canDelete />);

    await user.click(screen.getByRole("button", { name: "健診項目を追加" }));
    await user.selectOptions(
      screen.getByRole("combobox", { name: "健診項目の種別" }),
      "single_select",
    );
    await user.click(screen.getByRole("button", { name: "選択肢を追加" }));
    expect(screen.getByRole("textbox", { name: "選択肢1の値" })).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "選択肢1を削除" }));
    const dialog = screen.getByRole("alertdialog", { name: "選択肢を削除しますか？" });
    await user.click(within(dialog).getByRole("button", { name: "削除" }));

    await waitFor(() => {
      expect(screen.queryByRole("alertdialog")).not.toBeInTheDocument();
    });
    expect(screen.queryByRole("textbox", { name: "選択肢1の値" })).not.toBeInTheDocument();
  });

  it("creates a number field with unit and bounds", async () => {
    const user = userEvent.setup();
    render(<CheckupTypeFieldsEditor checkupTypeId="7" canCreate canEdit canDelete />);

    await user.click(screen.getByRole("button", { name: "健診項目を追加" }));
    await user.type(screen.getByRole("textbox", { name: "項目名" }), "体温");
    await user.type(screen.getByRole("textbox", { name: "単位" }), "℃");
    await user.type(screen.getByRole("spinbutton", { name: "基準値下限" }), "37");
    await user.type(screen.getByRole("spinbutton", { name: "基準値上限" }), "40");
    await user.click(screen.getByRole("button", { name: "項目を保存" }));

    expect(mocks.create).toHaveBeenCalledWith({
      checkupTypeId: "7",
      req: {
        name: "体温",
        field_type: "number",
        unit: "℃",
        min_value: 37,
        max_value: 40,
        sort_order: 3,
      },
    });
  });

  it("shows the options editor only for select types and validates them", async () => {
    const user = userEvent.setup();
    render(<CheckupTypeFieldsEditor checkupTypeId="7" canCreate canEdit canDelete />);

    await user.click(screen.getByRole("button", { name: "健診項目を追加" }));
    // number（初期値）: 選択肢エディタなし、基準値入力あり。
    expect(screen.queryByText("選択肢")).not.toBeInTheDocument();
    expect(screen.getByRole("spinbutton", { name: "基準値下限" })).toBeInTheDocument();

    await user.selectOptions(
      screen.getByRole("combobox", { name: "健診項目の種別" }),
      "single_select",
    );
    // 選択式: 選択肢エディタあり、基準値入力なし。
    expect(screen.getByText("選択肢")).toBeInTheDocument();
    expect(screen.queryByRole("spinbutton", { name: "基準値下限" })).not.toBeInTheDocument();

    await user.type(screen.getByRole("textbox", { name: "項目名" }), "総合評価");
    await user.click(screen.getByRole("button", { name: "項目を保存" }));
    expect(screen.getByRole("alert")).toHaveTextContent("選択肢を1件以上登録してください");
    expect(mocks.create).not.toHaveBeenCalled();

    await user.click(screen.getByRole("button", { name: "選択肢を追加" }));
    await user.type(screen.getByRole("textbox", { name: "選択肢1の値" }), "a");
    await user.type(screen.getByRole("textbox", { name: "選択肢1の表示名" }), "良好");
    await user.click(screen.getByRole("button", { name: "項目を保存" }));

    expect(mocks.create).toHaveBeenCalledWith({
      checkupTypeId: "7",
      req: {
        name: "総合評価",
        field_type: "single_select",
        unit: "",
        options: [{ value: "a", label: "良好" }],
        sort_order: 3,
      },
    });
  });

  it("edits an existing field and sends the full update payload with clear flags", async () => {
    const user = userEvent.setup();
    render(<CheckupTypeFieldsEditor checkupTypeId="7" canCreate canEdit canDelete />);

    await user.click(screen.getByRole("button", { name: "編集: 健診項目 体重 (ID 31)" }));
    expect(screen.getByRole("heading", { name: "健診項目を編集" })).toBeInTheDocument();

    const name = screen.getByRole("textbox", { name: "項目名" });
    await user.clear(name);
    await user.type(name, "体重（朝）");
    const min = screen.getByRole("spinbutton", { name: "基準値下限" });
    await user.clear(min);
    await user.click(screen.getByRole("button", { name: "項目を保存" }));

    expect(mocks.update).toHaveBeenCalledWith({
      checkupTypeId: "7",
      fieldId: "31",
      req: {
        name: "体重（朝）",
        field_type: "number",
        unit: "kg",
        clear_min_value: true,
        max_value: 80,
      },
    });
  });

  it("shows accessible loading and error states for the field list fetch", () => {
    mocks.getFields.mockReturnValue({
      data: undefined,
      error: null,
      isError: false,
      isPending: true,
    });
    const { rerender } = render(
      <CheckupTypeFieldsEditor checkupTypeId="7" canCreate canEdit canDelete />,
    );
    const status = screen.getByRole("status");
    expect(status).toHaveTextContent("健診項目を読み込み中です。");
    expect(status).toHaveAttribute("aria-live", "polite");
    expect(screen.queryByRole("button", { name: /削除: 健診項目/ })).not.toBeInTheDocument();

    mocks.getFields.mockReturnValue({
      data: undefined,
      error: new Error("GET /v1/masters/checkup-types/7/fields: timeout"),
      isError: true,
      isPending: false,
    });
    rerender(<CheckupTypeFieldsEditor checkupTypeId="7" canCreate canEdit canDelete />);
    const alert = screen.getByRole("alert");
    expect(alert).toHaveTextContent("健診項目の取得に失敗しました。");
    expect(screen.queryByText(/timeout/)).not.toBeInTheDocument();
    expect(screen.queryByRole("status")).not.toBeInTheDocument();
  });

  it("reports nested dirty state and clears it on cancel/save", async () => {
    const user = userEvent.setup();
    const onDirtyChange = vi.fn();
    render(
      <CheckupTypeFieldsEditor
        checkupTypeId="7"
        canCreate
        canEdit
        canDelete
        onDirtyChange={onDirtyChange}
      />,
    );

    await user.click(screen.getByRole("button", { name: "編集: 健診項目 体重 (ID 31)" }));
    await user.type(screen.getByRole("textbox", { name: "単位" }), "x");
    expect(onDirtyChange).toHaveBeenLastCalledWith(true);
    // dirty 中は追加/編集/削除/並べ替えを抑止。
    expect(screen.getByRole("button", { name: "健診項目を追加" })).toBeDisabled();
    expect(screen.getByRole("button", { name: "削除: 健診項目 体重 (ID 31)" })).toBeDisabled();
    mocks.reorderCallbacks.at(-1)?.(["32", "31"]);
    expect(mocks.reorder).not.toHaveBeenCalled();

    await user.click(screen.getByRole("button", { name: "キャンセル" }));
    expect(onDirtyChange).toHaveBeenLastCalledWith(false);

    await user.click(screen.getByRole("button", { name: "編集: 健診項目 体重 (ID 31)" }));
    await user.type(screen.getByRole("textbox", { name: "単位" }), "y");
    await user.click(screen.getByRole("button", { name: "項目を保存" }));
    expect(onDirtyChange).toHaveBeenLastCalledWith(false);
  });

  it("keeps the draft open when a mutation rejects", async () => {
    const user = userEvent.setup();
    const onDirtyChange = vi.fn();
    mocks.update.mockRejectedValueOnce(new Error("network"));
    render(
      <CheckupTypeFieldsEditor
        checkupTypeId="7"
        canCreate
        canEdit
        canDelete
        onDirtyChange={onDirtyChange}
      />,
    );

    await user.click(screen.getByRole("button", { name: "編集: 健診項目 体重 (ID 31)" }));
    await user.type(screen.getByRole("textbox", { name: "単位" }), "z");
    await user.click(screen.getByRole("button", { name: "項目を保存" }));

    expect(screen.getByRole("heading", { name: "健診項目を編集" })).toBeInTheDocument();
    expect(onDirtyChange).toHaveBeenLastCalledWith(true);
  });

  it("blocks nested text input Enter from submitting its parent form", async () => {
    const user = userEvent.setup();
    const parentSubmit = vi.fn((event: React.FormEvent) => event.preventDefault());
    const parentKeyDown = vi.fn();
    render(
      <form onSubmit={parentSubmit} onKeyDown={parentKeyDown}>
        <CheckupTypeFieldsEditor checkupTypeId="7" canCreate canEdit canDelete />
      </form>,
    );

    await user.click(screen.getByRole("button", { name: "編集: 健診項目 体重 (ID 31)" }));
    await user.type(screen.getByRole("textbox", { name: "項目名" }), "{enter}");

    expect(parentSubmit).not.toHaveBeenCalled();
    expect(parentKeyDown).not.toHaveBeenCalled();
  });
});
