import { fireEvent, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { ExaminationUnconfirmDialog } from "./ExaminationUnconfirmDialog";

describe("ExaminationUnconfirmDialog", () => {
  it("空白理由を拒否し、mutationを呼ばない", async () => {
    const user = userEvent.setup();
    const onUnconfirm = vi.fn().mockResolvedValue(true);
    render(<ExaminationUnconfirmDialog onUnconfirm={onUnconfirm} />);

    await user.click(screen.getByRole("button", { name: "確定解除" }));
    await user.type(
      screen.getByRole("textbox", { name: "確定解除理由" }),
      "   ",
    );
    await user.click(screen.getByRole("button", { name: "確定を解除する" }));

    expect(await screen.findByRole("alert")).toHaveTextContent(
      "確定解除理由は必須です",
    );
    expect(onUnconfirm).not.toHaveBeenCalled();
  });

  it("500文字を超える理由を拒否し、文字数を通知する", async () => {
    const user = userEvent.setup();
    const onUnconfirm = vi.fn().mockResolvedValue(true);
    render(<ExaminationUnconfirmDialog onUnconfirm={onUnconfirm} />);

    await user.click(screen.getByRole("button", { name: "確定解除" }));
    const reason = screen.getByRole("textbox", { name: "確定解除理由" });
    fireEvent.change(reason, { target: { value: "あ".repeat(501) } });
    await user.click(screen.getByRole("button", { name: "確定を解除する" }));

    expect(await screen.findByRole("alert")).toHaveTextContent("500文字以内");
    expect(onUnconfirm).not.toHaveBeenCalled();
  });

  it("trim済みの有効理由を送信し、成功時に閉じる", async () => {
    const user = userEvent.setup();
    const onUnconfirm = vi.fn().mockResolvedValue(true);
    render(<ExaminationUnconfirmDialog onUnconfirm={onUnconfirm} />);

    await user.click(screen.getByRole("button", { name: "確定解除" }));
    const reason = screen.getByRole("textbox", { name: "確定解除理由" });
    expect(reason).toHaveAttribute("maxlength", "500");
    expect(reason).toHaveAttribute("aria-required", "true");
    await user.type(reason, "  再確認のため  ");
    await user.click(screen.getByRole("button", { name: "確定を解除する" }));

    expect(onUnconfirm).toHaveBeenCalledOnce();
    expect(onUnconfirm).toHaveBeenCalledWith("再確認のため");
    expect(
      screen.queryByRole("textbox", { name: "確定解除理由" }),
    ).not.toBeInTheDocument();
  });

  it("mutation失敗時はdialogを開いたままerrorを表示する", async () => {
    const user = userEvent.setup();
    const onUnconfirm = vi.fn().mockResolvedValue(false);
    render(<ExaminationUnconfirmDialog onUnconfirm={onUnconfirm} />);

    await user.click(screen.getByRole("button", { name: "確定解除" }));
    await user.type(
      screen.getByRole("textbox", { name: "確定解除理由" }),
      "再確認のため",
    );
    await user.click(screen.getByRole("button", { name: "確定を解除する" }));

    expect(await screen.findByRole("alert")).toHaveTextContent(
      "確定解除に失敗しました",
    );
    expect(
      screen.getByRole("textbox", { name: "確定解除理由" }),
    ).toBeInTheDocument();
  });
});
