import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { AxiosError } from "axios";

import { axios } from "@/lib/axios";
import { ChangePasswordDialog } from "./ChangePasswordDialog";

vi.mock("@/lib/axios", () => ({
  axios: { put: vi.fn() },
}));

/**
 * 表示/非表示トグルの仕様確認用テスト。
 * - 初期状態は 3 フィールドとも `type="password"`
 * - Eye をクリックするとそのフィールドだけ `type="text"` に切替
 * - 再度クリックで `type="password"` に戻る
 * - 各フィールドのトグルは独立（他のフィールドに波及しない）
 *
 * input は placeholder で識別する（label に hint span が含まれるため
 * getByLabelText がマッチしづらい）。
 */
describe("ChangePasswordDialog — password visibility toggle", () => {
  function renderDialog() {
    return render(
      <ChangePasswordDialog open={true} onOpenChange={() => {}} />,
    );
  }

  function getInputs() {
    return {
      current: screen.getByPlaceholderText("現在のパスワードを入力") as HTMLInputElement,
      next: screen.getByPlaceholderText("新しいパスワードを入力") as HTMLInputElement,
      confirm: screen.getByPlaceholderText("もう一度入力") as HTMLInputElement,
    };
  }

  it("初期状態で 3 つのパスワード入力はマスキングされている", () => {
    renderDialog();
    const { current, next, confirm } = getInputs();

    expect(current.type).toBe("password");
    expect(next.type).toBe("password");
    expect(confirm.type).toBe("password");
  });

  it("各フィールドに対応する Eye トグルが配置されている", () => {
    renderDialog();
    expect(screen.getByRole("button", { name: "現在のパスワードを表示" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "新しいパスワードを表示" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "確認用パスワードを表示" })).toBeInTheDocument();
    for (const input of Object.values(getInputs())) {
      expect(input).toHaveClass("pr-12");
    }
  });

  it("現在のパスワードのトグルは current_password だけに作用する", async () => {
    const user = userEvent.setup();
    renderDialog();
    const { current, next, confirm } = getInputs();

    await user.click(screen.getByRole("button", { name: "現在のパスワードを表示" }));

    expect(current.type).toBe("text");
    expect(next.type).toBe("password");
    expect(confirm.type).toBe("password");

    // aria-label が「非表示」へ更新され、再クリックで戻る
    await user.click(screen.getByRole("button", { name: "現在のパスワードを非表示" }));
    expect(current.type).toBe("password");
  });

  it("新しいパスワードと確認用パスワードのトグルも独立している", async () => {
    const user = userEvent.setup();
    renderDialog();
    const { next, confirm } = getInputs();

    await user.click(screen.getByRole("button", { name: "新しいパスワードを表示" }));
    expect(next.type).toBe("text");
    expect(confirm.type).toBe("password");

    await user.click(screen.getByRole("button", { name: "確認用パスワードを表示" }));
    expect(next.type).toBe("text");
    expect(confirm.type).toBe("text");
  });

  it("aria-pressed がトグル状態を反映する", async () => {
    const user = userEvent.setup();
    renderDialog();

    const toggle = screen.getByRole("button", { name: "現在のパスワードを表示" });
    expect(toggle).toHaveAttribute("aria-pressed", "false");

    await user.click(toggle);
    expect(
      screen.getByRole("button", { name: "現在のパスワードを非表示" }),
    ).toHaveAttribute("aria-pressed", "true");
  });

  /**
   * Radix の DialogContent は aria-describedby を要求するため、
   * DialogDescription が存在しない場合は warning を出す。
   * description のテキストが描画され、かつ dialog ロールに
   * accessible description として紐付いていることを検証する。
   */
  it("401 でも axios.isAxiosError でクラッシュせず現在パスワード誤りを表示する (BUG-016)", async () => {
    const user = userEvent.setup();
    vi.mocked(axios.put).mockRejectedValueOnce(
      new AxiosError("unauthorized", "ERR_BAD_REQUEST", undefined, undefined, {
        status: 401,
        statusText: "Unauthorized",
        data: { error: "invalid current password" },
        headers: {},
        config: { headers: {} },
      } as never),
    );
    renderDialog();

    await user.type(screen.getByPlaceholderText("現在のパスワードを入力"), "wrong-pass");
    await user.type(screen.getByPlaceholderText("新しいパスワードを入力"), "newpassword1");
    await user.type(screen.getByPlaceholderText("もう一度入力"), "newpassword1");
    await user.click(screen.getByRole("button", { name: "変更する" }));

    expect(await screen.findByRole("alert")).toHaveTextContent(
      "現在のパスワードが正しくありません",
    );
    expect(screen.getByRole("dialog")).toBeInTheDocument();
  });

  it("DialogDescription がダイアログに紐付いている (a11y)", () => {
    renderDialog();

    const description = screen.getByText(
      "現在のパスワードと新しいパスワードを入力してください。",
    );
    expect(description).toBeInTheDocument();

    const dialog = screen.getByRole("dialog");
    expect(dialog).toHaveAccessibleDescription(
      "現在のパスワードと新しいパスワードを入力してください。",
    );
  });
});
