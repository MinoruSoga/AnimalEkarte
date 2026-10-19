import { fireEvent, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { C } from "@/lib/design-tokens";

import { PropertyFilter } from "./PropertyFilter";

const baseProps = {
  properties: [],
  activeFilters: [],
  onFilterChange: vi.fn(),
};

describe("PropertyFilter accessibility", () => {
  it("検索確定・クリア操作のhit areaを44px以上に保つ", () => {
    render(<PropertyFilter {...baseProps} searchTerm="ポチ" onSearchChange={vi.fn()} />);

    expect(screen.getByRole("button", { name: "検索を実行" })).toHaveClass("min-h-11", "min-w-11");
    expect(screen.getByRole("button", { name: "検索をクリア" })).toHaveClass(
      "min-h-11",
      "min-w-11",
    );
  });

  it("確定・クリア両ボタンの表示中はinputに88px分の右余白を確保する", () => {
    render(<PropertyFilter {...baseProps} searchTerm="ポチ" onSearchChange={vi.fn()} />);

    expect(screen.getByRole("textbox", { name: "検索..." })).toHaveClass("pr-22");
  });
});

// EMR-245: 表示列テキストフィルタ（type:"text"、contains 固定）。
describe("PropertyFilter text filter", () => {
  const textProps = {
    ...baseProps,
    properties: [{ key: "owner_name", label: "飼主名", type: "text" as const }],
  };

  it("text プロパティは条件選択を飛ばしてテキスト入力になる", async () => {
    const onFilterChange = vi.fn();
    render(<PropertyFilter {...textProps} onFilterChange={onFilterChange} />);
    const user = userEvent.setup();

    await user.click(screen.getByRole("button", { name: "フィルタを追加" }));
    await user.click(screen.getByRole("option", { name: "飼主名" }));

    const input = await screen.findByRole("textbox", { name: "飼主名を入力" });
    await user.type(input, "山田");
    fireEvent.keyDown(input, { key: "Enter" });

    expect(onFilterChange).toHaveBeenCalledTimes(1);
    expect(onFilterChange).toHaveBeenLastCalledWith([
      { key: "owner_name", condition: "contains", value: "山田", displayValue: "山田" },
    ]);
  });

  it("空白のみの入力はフィルタを追加しない", async () => {
    const onFilterChange = vi.fn();
    render(<PropertyFilter {...textProps} onFilterChange={onFilterChange} />);
    const user = userEvent.setup();

    await user.click(screen.getByRole("button", { name: "フィルタを追加" }));
    await user.click(screen.getByRole("option", { name: "飼主名" }));

    const input = await screen.findByRole("textbox", { name: "飼主名を入力" });
    await user.type(input, "   ");
    fireEvent.keyDown(input, { key: "Enter" });

    expect(onFilterChange).not.toHaveBeenCalled();
  });

  it("追加済み text フィルタは『含む』固定ラベルと値を表示し、値を編集できる", async () => {
    const onFilterChange = vi.fn();
    render(
      <PropertyFilter
        {...textProps}
        activeFilters={[
          { key: "owner_name", condition: "contains", value: "山田", displayValue: "山田" },
        ]}
        onFilterChange={onFilterChange}
      />,
    );
    const user = userEvent.setup();

    // 条件セレクタではなく静的ラベル
    expect(screen.getByText("含む")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "次と一致" })).not.toBeInTheDocument();

    // 値ピルを開いて再編集
    await user.click(screen.getByRole("button", { name: "山田" }));
    const input = await screen.findByRole("textbox", { name: "飼主名を入力" });
    await user.clear(input);
    await user.type(input, "佐藤");
    await user.click(screen.getByRole("button", { name: "飼主名フィルタを適用" }));

    expect(onFilterChange).toHaveBeenLastCalledWith([
      { key: "owner_name", condition: "contains", value: "佐藤", displayValue: "佐藤" },
    ]);
  });
});

describe("PropertyFilter count display", () => {
  it("件数はカンマ区切りで表示する", () => {
    render(<PropertyFilter {...baseProps} searchTerm="" onSearchChange={vi.fn()} count={13025} />);

    expect(screen.getByText("13,025 件")).toBeInTheDocument();
  });
});

// EMR-453: 検索語は確定操作（検索ボタン）でのみ onSearchChange へ渡す。
// 入力途中・入力欄 Enter は内部 draft のみ更新し、適用済み条件は変更しない
// （ui-contracts.md AC-common-ui-explicit-search-clear / FR-common-ui-explicit-business-search）。
describe("PropertyFilter search confirm-only", () => {
  it("検索入力は常時表示される（トグル操作なしで textbox が存在する）", () => {
    render(<PropertyFilter {...baseProps} searchTerm="田中" onSearchChange={vi.fn()} />);

    const input = screen.getByRole("textbox", { name: "検索..." });
    expect(input).toBeInTheDocument();
    expect(input).toHaveValue("田中");
    expect(screen.queryByRole("button", { name: "検索" })).not.toBeInTheDocument();
  });

  it("入力中は onSearchChange を呼ばず内部 draft のみ更新する", async () => {
    const onSearchChange = vi.fn();
    render(<PropertyFilter {...baseProps} searchTerm="" onSearchChange={onSearchChange} />);

    const user = userEvent.setup();
    const input = screen.getByRole("textbox", { name: "検索..." });
    await user.type(input, "田中");

    expect(input).toHaveValue("田中");
    expect(onSearchChange).not.toHaveBeenCalled();
  });

  it("Enter を押しても draft を確定せず onSearchChange を呼ばない（明示検索のみ適用）", async () => {
    const onSearchChange = vi.fn();
    render(<PropertyFilter {...baseProps} searchTerm="" onSearchChange={onSearchChange} />);

    const user = userEvent.setup();
    const input = screen.getByRole("textbox", { name: "検索..." });
    await user.type(input, "田中");
    fireEvent.keyDown(input, { key: "Enter" });

    expect(onSearchChange).not.toHaveBeenCalled();
    expect(input).toHaveValue("田中");
  });

  it("IME 変換中の確定 Enter では onSearchChange を呼ばない", async () => {
    const onSearchChange = vi.fn();
    render(<PropertyFilter {...baseProps} searchTerm="" onSearchChange={onSearchChange} />);

    const user = userEvent.setup();
    const input = screen.getByRole("textbox", { name: "検索..." });
    await user.type(input, "たなか");

    // 日本語 IME の変換確定 Enter（isComposing / keyCode 229）は検索確定にしない
    const imeEvent = new KeyboardEvent("keydown", { key: "Enter", bubbles: true });
    Object.defineProperty(imeEvent, "isComposing", { value: true });
    fireEvent(input, imeEvent);
    fireEvent.keyDown(input, { key: "Enter", keyCode: 229 });

    expect(onSearchChange).not.toHaveBeenCalled();
  });

  it("検索ボタン押下で draft を確定し onSearchChange(draft) を1回だけ呼ぶ", async () => {
    const onSearchChange = vi.fn();
    render(<PropertyFilter {...baseProps} searchTerm="" onSearchChange={onSearchChange} />);

    const user = userEvent.setup();
    const input = screen.getByRole("textbox", { name: "検索..." });
    await user.type(input, "佐藤");
    await user.click(screen.getByRole("button", { name: "検索を実行" }));

    expect(onSearchChange).toHaveBeenCalledTimes(1);
    expect(onSearchChange).toHaveBeenLastCalledWith("佐藤");
  });

  it("draft が確定済み searchTerm と一致する場合、検索ボタンは no-op", async () => {
    const onSearchChange = vi.fn();
    render(<PropertyFilter {...baseProps} searchTerm="田中" onSearchChange={onSearchChange} />);

    const user = userEvent.setup();
    await user.click(screen.getByRole("button", { name: "検索を実行" }));

    expect(onSearchChange).not.toHaveBeenCalled();
  });

  it('空 draft で検索ボタンを押すと onSearchChange("") で空条件を明示適用する', async () => {
    const onSearchChange = vi.fn();
    render(<PropertyFilter {...baseProps} searchTerm="田中" onSearchChange={onSearchChange} />);

    const user = userEvent.setup();
    const input = screen.getByRole("textbox", { name: "検索..." });
    await user.clear(input);
    await user.click(screen.getByRole("button", { name: "検索を実行" }));

    expect(onSearchChange).toHaveBeenCalledTimes(1);
    expect(onSearchChange).toHaveBeenLastCalledWith("");
  });

  it('クリア ✕ は draft のみ消去し、適用済み条件（onSearchChange 済みの値）は変更しない', async () => {
    const onSearchChange = vi.fn();
    render(<PropertyFilter {...baseProps} searchTerm="田中" onSearchChange={onSearchChange} />);

    const user = userEvent.setup();
    await user.click(screen.getByRole("button", { name: "検索をクリア" }));

    expect(onSearchChange).not.toHaveBeenCalled();
    expect(screen.getByRole("textbox", { name: "検索..." })).toHaveValue("");
  });

  it("外部から searchTerm が変わると draft が再同期される", async () => {
    const onSearchChange = vi.fn();
    const { rerender } = render(
      <PropertyFilter {...baseProps} searchTerm="初期" onSearchChange={onSearchChange} />,
    );

    const user = userEvent.setup();
    const input = screen.getByRole("textbox", { name: "検索..." });
    await user.type(input, "未確定");

    rerender(
      <PropertyFilter {...baseProps} searchTerm="外部更新" onSearchChange={onSearchChange} />,
    );

    expect(input).toHaveValue("外部更新");
  });
});

// EMR-411 C-1b: フィルタ削除ボタンはタッチ端末 (hover:none) で常時表示にする。
// opacity-0 + group-hover/focus 系のマウス契約は維持し、カラーを C.text60 へ。
describe("PropertyFilter フィルタ削除ボタンのタッチ可視性 (EMR-411 C-1b)", () => {
  it("削除ボタンが hover:none 常時表示・マウス hover/focus 表示・text60 を併せ持つ", () => {
    render(
      <PropertyFilter
        properties={[
          {
            key: "owner_name",
            label: "飼主名",
            type: "select",
            options: [{ value: "x", label: "X" }],
          },
        ]}
        activeFilters={[{ key: "owner_name", condition: "is", value: "x", displayValue: "X" }]}
        onFilterChange={vi.fn()}
      />,
    );

    const remove = screen.getByRole("button", { name: "飼主名 フィルタを削除" });
    expect(remove).toHaveClass("[@media(hover:none)]:opacity-100");
    expect(remove).toHaveClass(
      "opacity-0",
      "group-hover:opacity-100",
      "focus-visible:opacity-100",
      "group-focus-within:opacity-100",
    );
    expect(remove).toHaveClass(C.text60);
    expect(remove).toHaveClass("min-h-11", "min-w-11");
  });
});
