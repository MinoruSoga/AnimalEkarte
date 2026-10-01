import { fireEvent, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { PropertyFilter } from "./PropertyFilter";

const baseProps = {
  properties: [],
  activeFilters: [],
  onFilterChange: vi.fn(),
};

describe("PropertyFilter accessibility", () => {
  it("検索確定・クリア操作のhit areaを44px以上に保つ", () => {
    render(<PropertyFilter {...baseProps} searchTerm="ポチ" onSearchChange={vi.fn()} />);

    fireEvent.click(screen.getByRole("button", { name: "検索" }));

    expect(screen.getByRole("button", { name: "検索を実行" })).toHaveClass("min-h-11", "min-w-11");
    expect(screen.getByRole("button", { name: "検索をクリア" })).toHaveClass(
      "min-h-11",
      "min-w-11",
    );
  });

  it("確定・クリア両ボタンの表示中はinputに88px分の右余白を確保する", () => {
    render(<PropertyFilter {...baseProps} searchTerm="ポチ" onSearchChange={vi.fn()} />);

    fireEvent.click(screen.getByRole("button", { name: "検索" }));

    expect(screen.getByRole("textbox", { name: "検索..." })).toHaveClass("pr-22");
  });
});

describe("PropertyFilter count display", () => {
  it("件数はカンマ区切りで表示する", () => {
    render(<PropertyFilter {...baseProps} searchTerm="" onSearchChange={vi.fn()} count={13025} />);

    expect(screen.getByText("13,025 件")).toBeInTheDocument();
  });
});

// EMR-247: 検索語は確定操作（Enter / 検索ボタン）でのみ onSearchChange へ渡す。
// 入力途中は内部 draft のみ更新し、キー毎の発火（リクエスト連打）を防ぐ。
describe("PropertyFilter search confirm-only", () => {
  async function openSearch() {
    const user = userEvent.setup();
    await user.click(screen.getByRole("button", { name: "検索" }));
    return user;
  }

  it("入力中は onSearchChange を呼ばず内部 draft のみ更新する", async () => {
    const onSearchChange = vi.fn();
    render(<PropertyFilter {...baseProps} searchTerm="" onSearchChange={onSearchChange} />);

    const user = await openSearch();
    const input = screen.getByRole("textbox", { name: "検索..." });
    await user.type(input, "田中");

    expect(input).toHaveValue("田中");
    expect(onSearchChange).not.toHaveBeenCalled();
  });

  it("Enter で draft を確定し onSearchChange(draft) を1回だけ呼ぶ", async () => {
    const onSearchChange = vi.fn();
    render(<PropertyFilter {...baseProps} searchTerm="" onSearchChange={onSearchChange} />);

    const user = await openSearch();
    const input = screen.getByRole("textbox", { name: "検索..." });
    await user.type(input, "田中");
    fireEvent.keyDown(input, { key: "Enter" });

    expect(onSearchChange).toHaveBeenCalledTimes(1);
    expect(onSearchChange).toHaveBeenLastCalledWith("田中");
  });

  it("IME 変換中の確定 Enter では onSearchChange を呼ばない", async () => {
    const onSearchChange = vi.fn();
    render(<PropertyFilter {...baseProps} searchTerm="" onSearchChange={onSearchChange} />);

    const user = await openSearch();
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

    const user = await openSearch();
    const input = screen.getByRole("textbox", { name: "検索..." });
    await user.type(input, "佐藤");
    await user.click(screen.getByRole("button", { name: "検索を実行" }));

    expect(onSearchChange).toHaveBeenCalledTimes(1);
    expect(onSearchChange).toHaveBeenLastCalledWith("佐藤");
  });

  it("draft が確定済み searchTerm と一致する確定は no-op", async () => {
    const onSearchChange = vi.fn();
    render(<PropertyFilter {...baseProps} searchTerm="田中" onSearchChange={onSearchChange} />);

    await openSearch();
    const input = screen.getByRole("textbox", { name: "検索..." });
    fireEvent.keyDown(input, { key: "Enter" });

    expect(onSearchChange).not.toHaveBeenCalled();
  });

  it('空の確定は onSearchChange("") でアクティブ検索をクリアする', async () => {
    const onSearchChange = vi.fn();
    render(<PropertyFilter {...baseProps} searchTerm="田中" onSearchChange={onSearchChange} />);

    const user = await openSearch();
    const input = screen.getByRole("textbox", { name: "検索..." });
    await user.clear(input);
    fireEvent.keyDown(input, { key: "Enter" });

    expect(onSearchChange).toHaveBeenCalledTimes(1);
    expect(onSearchChange).toHaveBeenLastCalledWith("");
  });

  it('クリア ✕ は onSearchChange("") を即時呼び draft もクリアする', async () => {
    const onSearchChange = vi.fn();
    render(<PropertyFilter {...baseProps} searchTerm="田中" onSearchChange={onSearchChange} />);

    const user = await openSearch();
    await user.click(screen.getByRole("button", { name: "検索をクリア" }));

    expect(onSearchChange).toHaveBeenCalledTimes(1);
    expect(onSearchChange).toHaveBeenLastCalledWith("");
    expect(screen.getByRole("textbox", { name: "検索..." })).toHaveValue("");
  });

  it('検索バーを閉じると onSearchChange("") を即時呼ぶ（BUG-091）', async () => {
    const onSearchChange = vi.fn();
    render(<PropertyFilter {...baseProps} searchTerm="田中" onSearchChange={onSearchChange} />);

    const user = await openSearch();
    const input = screen.getByRole("textbox", { name: "検索..." });
    await user.type(input, "追加分");
    await user.click(screen.getByRole("button", { name: "検索" }));

    expect(onSearchChange).toHaveBeenCalledWith("");
  });

  it("外部から searchTerm が変わると draft が再同期される", async () => {
    const onSearchChange = vi.fn();
    const { rerender } = render(
      <PropertyFilter {...baseProps} searchTerm="初期" onSearchChange={onSearchChange} />,
    );

    const user = await openSearch();
    const input = screen.getByRole("textbox", { name: "検索..." });
    await user.type(input, "未確定");

    rerender(
      <PropertyFilter {...baseProps} searchTerm="外部更新" onSearchChange={onSearchChange} />,
    );

    expect(input).toHaveValue("外部更新");
  });
});
