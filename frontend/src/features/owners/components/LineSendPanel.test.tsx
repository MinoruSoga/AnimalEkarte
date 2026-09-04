import { render, screen, within } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { LineSendPanel } from "./LineSendPanel";

vi.mock("../api/get-owner-line-tags", () => ({
  useGetOwnerLineTags: () => ({
    data: { is_linked: false, lstep_opt_out: false },
  }),
}));

vi.mock("../api/send-line-message", () => ({
  useSendLineMessage: () => ({ mutateAsync: vi.fn() }),
}));

vi.mock("./LineSendHistory", () => ({
  LineSendHistory: () => <div>送信履歴</div>,
}));

describe("LineSendPanel", () => {
  it("狭幅でviewportを越えず、タイトルと44px closeの領域を分ける", () => {
    render(<LineSendPanel ownerId="owner-1" ownerName="山田 太郎" open onOpenChange={vi.fn()} />);

    const panel = screen.getByRole("dialog");
    expect(panel).toHaveAccessibleDescription("山田 太郎さんへLINEメッセージを送信します");
    expect(panel).toHaveClass("w-full", "max-w-full", "sm:max-w-[480px]");
    expect(panel).not.toHaveClass("w-[480px]");
    expect(
      within(panel).getByText("LINE送信 — 山田 太郎").closest('[data-slot="sheet-header"]'),
    ).toHaveClass("pr-16");
    expect(within(panel).getByRole("button", { name: "閉じる" })).toHaveClass(
      "min-h-11",
      "min-w-11",
    );
  });
});
