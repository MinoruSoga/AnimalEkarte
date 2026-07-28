import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, useLocation } from "react-router";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { C } from "@/lib/design-tokens";
import { LoginForm } from "./LoginForm";

const { loginMock } = vi.hoisted(() => ({ loginMock: vi.fn() }));

vi.mock("../hooks/use-auth", () => ({
  useAuth: () => ({
    login: loginMock,
    isAuthenticated: false,
  }),
}));

function CurrentLocation() {
  const location = useLocation();
  return <span data-testid="current-location">{location.pathname}</span>;
}

describe("LoginForm touch targets", () => {
  beforeEach(() => {
    loginMock.mockReset().mockResolvedValue(undefined);
  });

  it("パスワード再設定リンクは44px以上の操作領域を持つ", () => {
    render(
      <MemoryRouter>
        <LoginForm />
      </MemoryRouter>,
    );

    expect(screen.getByRole("link", { name: "パスワードをお忘れですか？" })).toHaveAttribute(
      "href",
      "/forgot-password",
    );
    expect(screen.getByRole("link", { name: "パスワードをお忘れですか？" })).toHaveClass("min-h-11");
    expect(screen.getByPlaceholderText("パスワードを入力")).toHaveClass("pr-12");
  });

  it("ログインCTAは brand teal と認証向け文字スタイルを明示する", () => {
    render(
      <MemoryRouter>
        <LoginForm />
      </MemoryRouter>,
    );

    const loginButton = screen.getByRole("button", { name: "ログイン" });
    expect(loginButton).toHaveClass(C.bgBrandIdentity, C.hoverBgBrandIdentity);
    expect(loginButton).toHaveClass("text-white", "text-xl", "font-bold");
    expect(loginButton).not.toHaveClass("font-semibold");
    expect(loginButton).not.toHaveClass("text-black", "text-base", "font-medium");
  });

  it("backslashを使うcross-origin fromをhomeへ縮退する", async () => {
    const user = userEvent.setup();
    render(
      <MemoryRouter
        initialEntries={[
          { pathname: "/login", state: { from: "/\\evil.example" } },
        ]}
      >
        <LoginForm />
        <CurrentLocation />
      </MemoryRouter>,
    );

    await user.type(screen.getByLabelText("メールアドレス"), "staff@example.com");
    await user.type(screen.getByLabelText("パスワード"), "password123");
    await user.click(screen.getByRole("button", { name: "ログイン" }));

    await waitFor(() =>
      expect(screen.getByTestId("current-location")).toHaveTextContent(/^\/$/),
    );
  });
});
