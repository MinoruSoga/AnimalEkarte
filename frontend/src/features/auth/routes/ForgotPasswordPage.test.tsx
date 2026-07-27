/**
 * FE6-6: anti-enumeration と実装の矛盾是正。
 * catch 内で { status: "sent" } を返し「アドレス不存在でも成功として扱う」と
 * コメントしているにもかかわらず、handleApiError が toast を出すため、
 * ネットワーク断など catch に入るケースで意図せず失敗が可視化されていた
 * （列挙防止の意図と実装が矛盾）。
 */
import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, act } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router";
import { toast } from "sonner";
import { C } from "@/lib/design-tokens";
import { ForgotPasswordPage } from "./ForgotPasswordPage";
import { forgotPassword } from "../api/forgot-password";

vi.mock("sonner", () => ({ toast: { success: vi.fn(), error: vi.fn(), warning: vi.fn() } }));
vi.mock("../api/forgot-password", () => ({ forgotPassword: vi.fn() }));

describe("FE6-6: ForgotPasswordPage anti-enumeration", () => {
  beforeEach(() => {
    vi.mocked(toast.error).mockClear();
    vi.mocked(forgotPassword).mockReset();
  });

  it("forgotPassword が失敗しても toast を出さず成功表示にする（列挙防止を破らない）", async () => {
    vi.mocked(forgotPassword).mockRejectedValue(new Error("network error"));
    const user = userEvent.setup();

    render(
      <MemoryRouter>
        <ForgotPasswordPage />
      </MemoryRouter>,
    );

    await user.type(screen.getByLabelText("メールアドレス"), "notfound@example.com");
    await act(async () => {
      await user.click(screen.getByRole("button", { name: /送信/ }));
    });

    expect(
      await screen.findByText("パスワードリセットのリンクをメールに送信しました。メールをご確認ください。"),
    ).toBeInTheDocument();
    expect(toast.error).not.toHaveBeenCalled();
  });

  it("forgotPassword が成功した場合も同じ成功表示になる", async () => {
    vi.mocked(forgotPassword).mockResolvedValue(undefined);
    const user = userEvent.setup();

    render(
      <MemoryRouter>
        <ForgotPasswordPage />
      </MemoryRouter>,
    );

    await user.type(screen.getByLabelText("メールアドレス"), "found@example.com");
    await act(async () => {
      await user.click(screen.getByRole("button", { name: /送信/ }));
    });

    expect(
      await screen.findByText("パスワードリセットのリンクをメールに送信しました。メールをご確認ください。"),
    ).toBeInTheDocument();
    expect(toast.error).not.toHaveBeenCalled();
  });

  it("ログインへ戻るリンクは44px以上の操作領域を持つ", () => {
    render(
      <MemoryRouter>
        <ForgotPasswordPage />
      </MemoryRouter>,
    );

    expect(screen.getByRole("link", { name: "ログインページに戻る" })).toHaveClass("min-h-11");
  });

  it("認証CTAとロゴは brand identity teal を明示する", () => {
    render(
      <MemoryRouter>
        <ForgotPasswordPage />
      </MemoryRouter>,
    );

    expect(screen.getByRole("button", { name: "リセットリンクを送信" })).toHaveClass(
      C.bgBrandIdentity,
      C.hoverBgBrandIdentity,
    );
    expect(screen.getByTestId("forgot-password-brand-mark")).toHaveClass(C.bgBrandIdentity);
  });
});
