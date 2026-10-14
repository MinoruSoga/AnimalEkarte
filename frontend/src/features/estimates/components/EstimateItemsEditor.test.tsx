import { describe, it, expect, vi } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { EstimateItemsEditor, type EstimateItemsEditorProps } from "./EstimateItemsEditor";
import type { EstimateItemDraft } from "../hooks/use-estimate-form";

function makeItem(overrides: Partial<EstimateItemDraft> = {}): EstimateItemDraft {
  return {
    id: "",
    localId: "local-1",
    name: "検査",
    category: "other",
    unitPrice: 1100,
    quantity: 2,
    taxType: "included",
    taxRate: 0.1,
    discountRate: 0,
    discountAmount: 0,
    isInsuranceApplicable: false,
    sortOrder: 0,
    ...overrides,
  };
}

function renderEditor(props: Partial<EstimateItemsEditorProps> = {}) {
  const onChange = vi.fn();
  const createItem = vi.fn(() =>
    makeItem({ localId: "local-new", name: "", unitPrice: 0, quantity: 1 }),
  );
  const utils = render(
    <EstimateItemsEditor
      items={[makeItem()]}
      createItem={createItem}
      canEditDiscount
      standardTaxRate={0.1}
      reducedTaxRate={0.08}
      onChange={onChange}
      {...props}
    />,
  );
  return { onChange, createItem, ...utils };
}

describe("EstimateItemsEditor", () => {
  it.each([
    [true, false],
    [false, true],
  ])("既存行と新規行の値引権限を分ける edit=%s create=%s", (edit, create) => {
    renderEditor({
      items: [makeItem({ id: "1", localId: "saved" }), makeItem({ localId: "new" })],
      canEditDiscount: edit,
      canCreateDiscount: create,
    });
    expect(document.getElementById("item-discount-saved") !== null).toBe(edit);
    expect(document.getElementById("item-discount-new") !== null).toBe(create);
  });

  it("各行にネイティブなアクセシブルラベルを持つ入力を描画する", () => {
    renderEditor();
    expect(screen.getByLabelText("品目名")).toHaveValue("検査");
    expect(screen.getByLabelText("単価（税込）")).toHaveValue(1100);
    expect(screen.getByLabelText("数量")).toHaveValue(2);
    expect(screen.getByLabelText("課税区分")).toBeInTheDocument();
    expect(screen.getByLabelText("税率")).toBeInTheDocument();
    expect(screen.getByLabelText("カテゴリ")).toBeInTheDocument();
  });

  it("外税行は単価（税抜）ラベルになる", () => {
    renderEditor({ items: [makeItem({ taxType: "excluded" })] });
    expect(screen.getByLabelText("単価（税抜）")).toBeInTheDocument();
    expect(screen.queryByLabelText("単価（税込）")).not.toBeInTheDocument();
  });

  it("明細行を追加 → createItem の結果を末尾へ追加して onChange", () => {
    const { onChange, createItem } = renderEditor();
    fireEvent.click(screen.getByRole("button", { name: "明細行を追加" }));
    expect(createItem).toHaveBeenCalledTimes(1);
    expect(onChange).toHaveBeenCalledWith([
      expect.objectContaining({ localId: "local-1" }),
      expect.objectContaining({ localId: "local-new" }),
    ]);
  });

  it("行を削除 → 対象行を除いた配列で onChange", () => {
    const { onChange } = renderEditor({
      items: [makeItem(), makeItem({ localId: "local-2", name: "薬" })],
    });
    const removeButtons = screen.getAllByRole("button", { name: "行を削除" });
    fireEvent.click(removeButtons[1]!);
    expect(onChange).toHaveBeenCalledWith([expect.objectContaining({ localId: "local-1" })]);
  });

  it("品目名を編集 → 更新済み行で onChange", () => {
    const { onChange } = renderEditor();
    fireEvent.change(screen.getByLabelText("品目名"), {
      target: { value: "血液検査" },
    });
    expect(onChange).toHaveBeenCalledWith([
      expect.objectContaining({ localId: "local-1", name: "血液検査" }),
    ]);
  });

  it("単価を編集 → 他の行フィールドを保った更新行で onChange", () => {
    const { onChange } = renderEditor();
    fireEvent.change(screen.getByLabelText("単価（税込）"), {
      target: { value: "2200" },
    });
    expect(onChange).toHaveBeenCalledWith([
      expect.objectContaining({
        localId: "local-1",
        unitPrice: 2200,
        name: "検査",
        category: "other",
        quantity: 2,
        taxType: "included",
        taxRate: 0.1,
        discountAmount: 0,
      }),
    ]);
  });

  it("数量を編集 → 他の行フィールドを保った更新行で onChange", () => {
    const { onChange } = renderEditor();
    fireEvent.change(screen.getByLabelText("数量"), {
      target: { value: "1.5" },
    });
    expect(onChange).toHaveBeenCalledWith([
      expect.objectContaining({
        localId: "local-1",
        quantity: 1.5,
        name: "検査",
        category: "other",
        unitPrice: 1100,
        taxType: "included",
        taxRate: 0.1,
        discountAmount: 0,
      }),
    ]);
  });

  it("値引額を編集 → 他の行フィールドを保った更新行で onChange", () => {
    const { onChange } = renderEditor();
    fireEvent.change(screen.getByLabelText("値引（円）"), {
      target: { value: "100" },
    });
    expect(onChange).toHaveBeenCalledWith([
      expect.objectContaining({
        localId: "local-1",
        discountAmount: 100,
        name: "検査",
        category: "other",
        unitPrice: 1100,
        quantity: 2,
        taxType: "included",
        taxRate: 0.1,
      }),
    ]);
  });

  it("カテゴリを処方へ変更 → category: medicine で onChange", async () => {
    const user = userEvent.setup();
    const { onChange } = renderEditor();
    await user.click(screen.getByLabelText("カテゴリ"));
    await user.click(await screen.findByRole("option", { name: "処方" }));
    expect(onChange).toHaveBeenCalledWith([
      expect.objectContaining({
        localId: "local-1",
        category: "medicine",
        name: "検査",
        unitPrice: 1100,
        quantity: 2,
        taxType: "included",
        taxRate: 0.1,
        discountAmount: 0,
      }),
    ]);
  });

  it("税率を8%へ変更 → taxRate: 0.08 で onChange", async () => {
    const user = userEvent.setup();
    const { onChange } = renderEditor();
    await user.click(screen.getByLabelText("税率"));
    await user.click(await screen.findByRole("option", { name: "8%" }));
    expect(onChange).toHaveBeenCalledWith([
      expect.objectContaining({
        localId: "local-1",
        taxRate: 0.08,
        name: "検査",
        category: "other",
        unitPrice: 1100,
        quantity: 2,
        taxType: "included",
        discountAmount: 0,
      }),
    ]);
  });

  it("課税区分を外税へ変更 → taxType: excluded で onChange", async () => {
    const user = userEvent.setup();
    const { onChange } = renderEditor();
    await user.click(screen.getByLabelText("課税区分"));
    await user.click(await screen.findByRole("option", { name: "外税" }));
    expect(onChange).toHaveBeenCalledWith([
      expect.objectContaining({ localId: "local-1", taxType: "excluded" }),
    ]);
  });

  it("行ごとの金額（税込）をカンマ区切り通貨で表示する", () => {
    renderEditor();
    expect(screen.getByText("¥2,200")).toBeInTheDocument();
  });

  it("disabled=true → 入力無効化かつ行追加/削除ボタンを出さない", () => {
    renderEditor({ disabled: true });
    expect(screen.getByLabelText("品目名")).toBeDisabled();
    expect(screen.getByLabelText("数量")).toBeDisabled();
    expect(screen.queryByRole("button", { name: "明細行を追加" })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "行を削除" })).not.toBeInTheDocument();
  });

  it("canEditDiscount=false → 値引入力を表示しない（読込値は維持）", () => {
    renderEditor({
      canEditDiscount: false,
      items: [makeItem({ discountAmount: 300 })],
    });
    expect(screen.queryByLabelText("値引（円）")).not.toBeInTheDocument();
    expect(screen.queryByLabelText("割引額")).not.toBeInTheDocument();
  });

  it("canEditDiscount=true → 値引入力を表示する", () => {
    renderEditor({ canEditDiscount: true });
    expect(screen.getByLabelText("値引（円）")).toHaveValue(0);
  });

  it("error があるとき role=alert を表示しエディタへ関連付ける", () => {
    renderEditor({ error: "1行目: 品目名を入力してください" });
    const alert = screen.getByRole("alert");
    expect(alert).toHaveTextContent("1行目: 品目名を入力してください");
    expect(alert).toHaveAttribute("id", "estimate-items-error");
  });
});
