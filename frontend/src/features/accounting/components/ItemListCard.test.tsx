import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { ItemListCard } from "./ItemListCard";
import type { AccountingItem } from "../types";
import { todayJSTISO } from "@/lib/jst-date";

vi.mock("../api/get-merchandise-items", () => ({
  useGetAllMerchandiseItems: () => ({
    data: [
      {
        id: "item-1",
        name: "療法食",
        category: "goods",
        unitPrice: 1200,
        taxType: "included",
        taxRate: 0.08,
        isActive: true,
      },
    ],
  }),
}));

describe("ItemListCard merchandise selection", () => {
  it("行は追加せず、固有名の44px native buttonだけが品目を追加する", async () => {
    const user = userEvent.setup();
    const onAddItem = vi.fn();
    render(
      <ItemListCard
        items={[]}
        subtotal={0}
        taxTotal={0}
        totalAmount={0}
        newItemOpen
        onNewItemOpenChange={vi.fn()}
        onAddItem={onAddItem}
        onDeleteItem={vi.fn()}
        canEdit
        canDelete={false}
      />,
    );

    await user.click(screen.getByText("物販"));
    expect(onAddItem).not.toHaveBeenCalled();

    const addButton = screen.getByRole("button", { name: "追加: 療法食 (ID item-1)" });
    expect(addButton.tagName).toBe("BUTTON");
    expect(addButton).toHaveClass("min-h-11", "min-w-11");
    await user.click(addButton);
    // EMR-65: マスタの税区分・税率を欠落なく onAddItem へ伝播する（一律 excluded 潰しの回帰防止）
    expect(onAddItem).toHaveBeenCalledWith({
      name: "療法食",
      price: "1200",
      category: "goods",
      taxType: "included",
      taxRate: 0.08,
      merchandiseItemId: "item-1",
    });
  });

  it("項目名ヘッダは最小幅を持ち税額ヘッダは折り返さない", () => {
    render(
      <ItemListCard
        items={[]}
        subtotal={0}
        taxTotal={0}
        totalAmount={0}
        newItemOpen={false}
        onNewItemOpenChange={vi.fn()}
        onAddItem={vi.fn()}
        onDeleteItem={vi.fn()}
        canEdit
        canDelete={false}
      />,
    );

    const nameHeader = screen.getByRole("columnheader", { name: "項目名" });
    expect(nameHeader.className).toMatch(/min-w-\[/);
    expect(nameHeader.closest(".overflow-auto")).not.toBeNull();

    const taxHeader = screen.getByRole("columnheader", { name: "税額" });
    expect(taxHeader.className).toContain("whitespace-nowrap");
  });

  function renderEditableItemListCard(onAddItem = vi.fn()) {
    render(
      <ItemListCard
        items={[]}
        subtotal={0}
        taxTotal={0}
        totalAmount={0}
        newItemOpen
        onNewItemOpenChange={vi.fn()}
        onAddItem={onAddItem}
        onDeleteItem={vi.fn()}
        canEdit
        canDelete={false}
      />,
    );
  }

  it("手動入力では選択したカテゴリを渡し、商品マスタ ID を渡さない", async () => {
    const user = userEvent.setup();
    const onAddItem = vi.fn();
    renderEditableItemListCard(onAddItem);

    await user.click(screen.getByRole("button", { name: "手動入力" }));
    await user.type(screen.getByLabelText(/品目名/), "手入力商品");
    await user.type(screen.getByLabelText(/単価/), "500");
    await user.click(screen.getByRole("combobox", { name: "カテゴリ" }));
    expect(screen.getAllByRole("option")).toHaveLength(12);
    await user.click(screen.getByRole("option", { name: "検査" }));
    await user.click(screen.getByRole("button", { name: "追加する" }));

    expect(onAddItem).toHaveBeenCalledWith({
      name: "手入力商品",
      price: "500",
      category: "test",
    });
  });

  // EMR-319: 必須未入力でもボタンは押せ、フィールドエラーで案内する
  // （以前は disabled が必須条件を含み、エラー経路に到達できなかった）。
  it("手動入力で品目名が空のまま追加を押すと品目名のエラーが出る", async () => {
    const user = userEvent.setup();
    const onAddItem = vi.fn();
    renderEditableItemListCard(onAddItem);

    await user.click(screen.getByRole("button", { name: "手動入力" }));
    await user.type(screen.getByLabelText(/単価/), "500");
    await user.click(screen.getByRole("combobox", { name: "カテゴリ" }));
    await user.click(screen.getByRole("option", { name: "検査" }));

    const addButton = screen.getByRole("button", { name: "追加する" });
    expect(addButton).toBeEnabled();
    await user.click(addButton);

    expect(onAddItem).not.toHaveBeenCalled();
    const nameInput = document.getElementById("manual-name");
    expect(nameInput).toHaveAttribute("aria-invalid", "true");
    expect(nameInput).toHaveAttribute("aria-describedby", "manual-name-error");
    expect(document.getElementById("manual-name-error")).toHaveTextContent(
      "品目名を入力してください",
    );
  });

  it("手動入力ではカテゴリ未選択のまま追加を押すとカテゴリのエラーが出る", async () => {
    const user = userEvent.setup();
    const onAddItem = vi.fn();
    renderEditableItemListCard(onAddItem);

    await user.click(screen.getByRole("button", { name: "手動入力" }));
    await user.type(screen.getByLabelText(/品目名/), "カテゴリ未選択商品");
    await user.type(screen.getByLabelText(/単価/), "500");

    const addButton = screen.getByRole("button", { name: "追加する" });
    expect(addButton).toBeEnabled();
    await user.click(addButton);

    expect(onAddItem).not.toHaveBeenCalled();
    const categoryTrigger = document.getElementById("manual-category");
    expect(categoryTrigger).toHaveAttribute("aria-invalid", "true");
    expect(document.getElementById("manual-category-error")).toHaveTextContent(
      "カテゴリを選択してください",
    );
  });

  it("手動入力ではother選択時だけ理由を必須にしてpayloadへ渡す", async () => {
    const user = userEvent.setup();
    const onAddItem = vi.fn();
    renderEditableItemListCard(onAddItem);

    await user.click(screen.getByRole("button", { name: "手動入力" }));
    await user.type(screen.getByLabelText(/品目名/), "その他手入力商品");
    await user.type(screen.getByLabelText(/単価/), "500");
    expect(screen.queryByLabelText(/その他理由/)).not.toBeInTheDocument();
    await user.click(screen.getByRole("combobox", { name: "カテゴリ" }));
    await user.click(screen.getByRole("option", { name: "その他" }));
    const addButton = screen.getByRole("button", { name: "追加する" });
    expect(addButton).toBeEnabled();
    const otherReasonInput = screen.getByLabelText(/その他理由/);
    expect(otherReasonInput).toHaveAttribute("maxlength", "500");
    await user.type(otherReasonInput, "   ");
    // 空白のみの理由で追加を押すと理由欄のエラーになる（ボタンは disabled にしない）
    await user.click(addButton);
    expect(onAddItem).not.toHaveBeenCalled();
    expect(otherReasonInput).toHaveAttribute("aria-invalid", "true");
    expect(document.getElementById("manual-other-reason-error")).toHaveTextContent(
      "その他理由を入力してください",
    );
    await user.type(otherReasonInput, "締め時に確認する分類  ");
    await user.click(screen.getByRole("button", { name: "追加する" }));

    expect(onAddItem).toHaveBeenCalledWith({
      name: "その他手入力商品",
      price: "500",
      category: "other",
      otherReason: "締め時に確認する分類",
    });
    expect(screen.queryByLabelText(/その他理由/)).not.toBeInTheDocument();
  });

  // EMR-319: 不正な単価で追加すると単価入力に aria-invalid+エラー文が出る
  it("単価が不正な場合、追加せず単価入力に aria-invalid+エラー文が出る (EMR-319)", async () => {
    const user = userEvent.setup();
    const onAddItem = vi.fn();
    renderEditableItemListCard(onAddItem);

    await user.click(screen.getByRole("button", { name: "手動入力" }));
    await user.type(screen.getByLabelText(/品目名/), "不正単価商品");
    await user.type(screen.getByLabelText(/単価/), "-1");
    await user.click(screen.getByRole("combobox", { name: "カテゴリ" }));
    await user.click(screen.getByRole("option", { name: "検査" }));
    await user.click(screen.getByRole("button", { name: "追加する" }));

    expect(onAddItem).not.toHaveBeenCalled();
    const priceInput = document.getElementById("manual-price");
    expect(priceInput).toHaveAttribute("aria-invalid", "true");
    expect(priceInput).toHaveAttribute("aria-describedby", "manual-price-error");
    expect(document.getElementById("manual-price-error")).toHaveTextContent(
      "単価は0以上の整数で入力してください",
    );

    // 値を直すとエラー表示が解除される
    await user.clear(priceInput!);
    await user.type(priceInput!, "500");
    expect(priceInput).not.toHaveAttribute("aria-invalid");
    expect(document.getElementById("manual-price-error")).toBeNull();
  });

  // PR#614 指摘: 以前は名称等の early return が単価評価より先で、複数項目が未入力の
  // 初回クリックで単価だけ未エラー表示のままだった。全フィールド同時表示を固定する。
  it("品目名も単価も未入力のまま初回クリックすると両方のエラーが同時に出る", async () => {
    const user = userEvent.setup();
    const onAddItem = vi.fn();
    renderEditableItemListCard(onAddItem);

    await user.click(screen.getByRole("button", { name: "手動入力" }));
    await user.click(screen.getByRole("button", { name: "追加する" }));

    expect(onAddItem).not.toHaveBeenCalled();
    expect(document.getElementById("manual-name-error")).toHaveTextContent(
      "品目名を入力してください",
    );
    expect(document.getElementById("manual-category-error")).toHaveTextContent(
      "カテゴリを選択してください",
    );
    const priceInput = document.getElementById("manual-price");
    expect(priceInput).toHaveAttribute("aria-invalid", "true");
    expect(document.getElementById("manual-price-error")).toHaveTextContent(
      "単価は0以上の整数で入力してください",
    );
  });

  it("手動入力でotherから別カテゴリへ戻すと理由欄とpayloadを消す", async () => {
    const user = userEvent.setup();
    const onAddItem = vi.fn();
    renderEditableItemListCard(onAddItem);

    await user.click(screen.getByRole("button", { name: "手動入力" }));
    await user.type(screen.getByLabelText(/品目名/), "カテゴリ変更商品");
    await user.type(screen.getByLabelText(/単価/), "500");
    await user.click(screen.getByRole("combobox", { name: "カテゴリ" }));
    await user.click(screen.getByRole("option", { name: "その他" }));
    await user.type(screen.getByLabelText(/その他理由/), "分類保留");
    await user.click(screen.getByRole("combobox", { name: "カテゴリ" }));
    await user.click(screen.getByRole("option", { name: "検査" }));

    expect(screen.queryByLabelText(/その他理由/)).not.toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "追加する" }));
    expect(onAddItem).toHaveBeenCalledWith({
      name: "カテゴリ変更商品",
      price: "500",
      category: "test",
    });
  });
});

// EMR-246: 未請求候補の当日/過去セクション区分
function makeItem(overrides: Partial<AccountingItem> = {}): AccountingItem {
  return {
    id: "item-1",
    category: "other",
    name: "項目",
    unitPrice: 100,
    quantity: 1,
    discountRate: 0,
    discountAmount: 0,
    taxType: "excluded",
    taxRate: 0.1,
    taxAmount: 10,
    subtotal: 100,
    isInsuranceApplicable: false,
    source: "manual",
    ...overrides,
  };
}

function renderWithItems(items: AccountingItem[]) {
  // AccountingItemRow → DiscountCell が useQuery を使うため Provider が必要。
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  render(
    <QueryClientProvider client={queryClient}>
      <ItemListCard
        items={items}
        subtotal={0}
        taxTotal={0}
        totalAmount={0}
        newItemOpen={false}
        onNewItemOpenChange={vi.fn()}
        onAddItem={vi.fn()}
        onDeleteItem={vi.fn()}
        canEdit={false}
        canDelete={false}
      />
    </QueryClientProvider>,
  );
}

describe("ItemListCard 当日/過去の区分", () => {
  it("過去の発生日を持つ明細があると当日分/過去分の見出しが出る", () => {
    renderWithItems([
      makeItem({ id: "today-1", name: "当日項目", serviceDate: todayJSTISO() }),
      makeItem({ id: "past-1", name: "過去項目", serviceDate: "2020-01-15" }),
      makeItem({ id: "manual-1", name: "手入力項目" }),
    ]);

    expect(screen.getByText("当日分")).toBeInTheDocument();
    expect(screen.getByText("過去の未請求分")).toBeInTheDocument();
    expect(screen.getByLabelText("発生日: 2020-01-15")).toBeInTheDocument();
    // 当日項目には発生日バッジが出ない
    expect(screen.queryByLabelText(`発生日: ${todayJSTISO()}`)).not.toBeInTheDocument();
  });

  it("過去分がなければセクション見出しを出さない", () => {
    renderWithItems([
      makeItem({ id: "today-1", name: "当日項目", serviceDate: todayJSTISO() }),
      makeItem({ id: "manual-1", name: "手入力項目" }),
    ]);

    expect(screen.queryByText("当日分")).not.toBeInTheDocument();
    expect(screen.queryByText("過去の未請求分")).not.toBeInTheDocument();
  });
});
