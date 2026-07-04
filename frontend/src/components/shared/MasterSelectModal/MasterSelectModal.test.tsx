import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MasterSelectModal } from "./MasterSelectModal";

const ITEMS = [
  { id: 1, name: "一般診療", price: 3000 },
  { id: 2, name: "再診", price: 1000 },
];

function renderModal(overrides?: { onSelect?: (item: (typeof ITEMS)[number]) => void }) {
  return render(
    <MasterSelectModal
      open
      onOpenChange={() => {}}
      title="診療内容を選択"
      items={ITEMS}
      onSelect={overrides?.onSelect ?? (() => {})}
    />,
  );
}

describe("MasterSelectModal", () => {
  it("開くと全項目が表示される", () => {
    renderModal();
    expect(screen.getByText("一般診療")).toBeInTheDocument();
    expect(screen.getByText("再診")).toBeInTheDocument();
  });

  it("検索語に一致しない場合、空状態メッセージが表示される", async () => {
    const user = userEvent.setup();
    renderModal();

    await user.type(screen.getByPlaceholderText("検索..."), "存在しない項目名ZZZ");

    expect(await screen.findByText("該当する項目がありません")).toBeInTheDocument();
    expect(screen.queryByText("一般診療")).not.toBeInTheDocument();
    expect(screen.queryByText("再診")).not.toBeInTheDocument();
  });

  it("検索語で名前一致の項目のみ表示される", async () => {
    const user = userEvent.setup();
    renderModal();

    await user.type(screen.getByPlaceholderText("検索..."), "再診");

    expect(await screen.findByText("再診")).toBeInTheDocument();
    expect(screen.queryByText("一般診療")).not.toBeInTheDocument();
  });

  it("項目をクリックすると onSelect が呼ばれる", async () => {
    const user = userEvent.setup();
    const onSelect = vi.fn();
    renderModal({ onSelect });

    await user.click(screen.getByText("一般診療"));

    expect(onSelect).toHaveBeenCalledWith(ITEMS[0]);
  });

  it("検索後にクリアボタンで全件表示に戻る", async () => {
    const user = userEvent.setup();
    renderModal();

    const input = screen.getByPlaceholderText("検索...");
    await user.type(input, "再診");
    expect(input).toHaveValue("再診");
    expect(screen.queryByText("一般診療")).not.toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "検索をクリア" }));

    expect(input).toHaveValue("");
    expect(screen.getByText("一般診療")).toBeInTheDocument();
    expect(screen.getByText("再診")).toBeInTheDocument();
  });
});
