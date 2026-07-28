import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { AxiosError, AxiosHeaders, type InternalAxiosRequestConfig } from "axios";
import { MemoryRouter, useLocation } from "react-router";
import { toast } from "sonner";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { C } from "@/lib/design-tokens";
import { resetPassword } from "../api/reset-password";
import { ResetPasswordPage } from "./ResetPasswordPage";

vi.mock("../api/reset-password", () => ({ resetPassword: vi.fn() }));
vi.mock("sonner", () => ({
  toast: { error: vi.fn(), success: vi.fn(), warning: vi.fn() },
}));

function CurrentLocation() {
  const location = useLocation();
  return (
    <span data-testid="current-location">
      {`${location.pathname}${location.search}${location.hash}`}
    </span>
  );
}

function invalidTokenError(): AxiosError {
  const config: InternalAxiosRequestConfig = { headers: new AxiosHeaders() };
  return new AxiosError(
    "invalid or expired token",
    AxiosError.ERR_BAD_REQUEST,
    config,
    undefined,
    {
      config,
      data: { message: "invalid or expired token" },
      headers: new AxiosHeaders(),
      status: 400,
      statusText: "Bad Request",
    },
  );
}

describe("ResetPasswordPage recovery behavior", () => {
  beforeEach(() => {
    vi.mocked(resetPassword).mockReset();
    vi.mocked(toast.error).mockClear();
  });

  it("無効なリンク画面の再申請リンクは44px以上の操作領域を持つ", () => {
    render(
      <MemoryRouter initialEntries={["/reset-password"]}>
        <ResetPasswordPage />
      </MemoryRouter>,
    );

    expect(
      screen.getByRole("link", { name: "パスワードリセットを再申請する" }),
    ).toHaveClass("min-h-11");
    expect(resetPassword).not.toHaveBeenCalled();
  });

  it("有効なリンク画面のログインリンクは44px以上の操作領域を持つ", () => {
    render(
      <MemoryRouter initialEntries={["/reset-password?token=test-token"]}>
        <ResetPasswordPage />
      </MemoryRouter>,
    );

    expect(screen.getByRole("link", { name: "ログインページに戻る" })).toHaveClass("min-h-11");
    expect(screen.getByPlaceholderText("8文字以上で入力")).toHaveClass("pr-12");
    expect(screen.getByPlaceholderText("同じパスワードを入力")).toHaveClass("pr-12");
  });

  it("有効なリンク画面のCTAとロゴは brand identity teal を明示する", () => {
    render(
      <MemoryRouter initialEntries={["/reset-password?token=test-token"]}>
        <ResetPasswordPage />
      </MemoryRouter>,
    );

    const submit = screen.getByRole("button", { name: "パスワードを設定する" });
    expect(submit).toHaveClass(C.bgBrandIdentity, C.hoverBgBrandIdentity);
    expect(screen.getByTestId("reset-password-brand-mark")).toHaveClass(C.bgBrandIdentity);
  });

  it("fragment tokenを取得後にbrowser URLから除去する", async () => {
    render(
      <MemoryRouter initialEntries={["/reset-password/#token=fragment-token"]}>
        <ResetPasswordPage />
        <CurrentLocation />
      </MemoryRouter>,
    );

    expect(
      screen.getByRole("heading", { name: "新しいパスワードの設定" }),
    ).toBeInTheDocument();
    await waitFor(() =>
      expect(screen.getByTestId("current-location")).toHaveTextContent(
        /^\/reset-password$/,
      ),
    );
    expect(screen.getByTestId("current-location")).not.toHaveTextContent("token");
  });

  it.each(["invalid-token", "expired-token"])(
    "%sの400は画面固有エラーを表示してrecovery routeに留まる",
    async (token) => {
      vi.mocked(resetPassword).mockRejectedValue(invalidTokenError());
      const user = userEvent.setup();

      render(
        <MemoryRouter initialEntries={[`/reset-password?token=${token}`]}>
          <ResetPasswordPage />
          <CurrentLocation />
        </MemoryRouter>,
      );

      await user.type(screen.getByLabelText("新しいパスワード"), "password123");
      await user.type(screen.getByLabelText("パスワード（確認）"), "password123");
      await user.click(
        screen.getByRole("button", { name: "パスワードを設定する" }),
      );

      expect(
        await screen.findByText(
          "パスワードのリセットに失敗しました。リンクの有効期限が切れている可能性があります。",
        ),
      ).toBeInTheDocument();
      expect(
        screen.getByRole("heading", { name: "新しいパスワードの設定" }),
      ).toBeInTheDocument();
      expect(screen.getByTestId("current-location")).toHaveTextContent(
        /^\/reset-password$/,
      );
      expect(screen.getByTestId("current-location")).not.toHaveTextContent(token);
      expect(resetPassword).toHaveBeenCalledWith({
        token,
        password: "password123",
      });
      expect(toast.error).not.toHaveBeenCalled();
    },
  );
});
