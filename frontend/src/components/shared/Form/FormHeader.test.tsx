import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { FormHeader } from "./FormHeader";

describe("FormHeader", () => {
  it("ページ見出しは DESIGN.md title role（20px/600/1.4/-0.125px）を使う", () => {
    render(<FormHeader title="飼主登録" />);

    const heading = screen.getByRole("heading", { level: 1, name: "飼主登録" });

    // globals.css の text-xl が 20px/1.4/-0.125px、font-semibold が 600 を担う。
    expect(heading).toHaveClass("text-xl", "font-semibold");
    expect(heading).not.toHaveClass("text-base", "leading-tight");
  });

  it("長い説明が mobile で折り返しても固定高でクリップしない", () => {
    const { container } = render(
      <FormHeader
        title="健診リマインダー抽出"
        description="Lステップタグを一括付与して健診リマインダーをLINE送信します。"
      />,
    );

    const header = container.firstElementChild;
    expect(header).toHaveClass("min-h-[53px]", "py-1");
    expect(header).not.toHaveClass("h-[53px]");
  });

  it("狭幅で title/action が折り返せる flex-wrap 契約を持つ (BUG-458)", () => {
    const { container } = render(
      <FormHeader
        title="当日の受付"
        description="受付状況をリアルタイムで確認"
        action={<button type="button">新規予約登録</button>}
      />,
    );

    const header = container.firstElementChild;
    expect(header).toHaveClass("flex-wrap", "min-w-0");
    expect(screen.getByRole("heading", { name: "当日の受付" })).toHaveClass("break-words");
  });
});
