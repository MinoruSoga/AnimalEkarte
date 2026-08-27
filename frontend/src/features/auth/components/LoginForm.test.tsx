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

describe("LoginForm demo accounts (staff-attach)", () => {
  beforeEach(() => {
    loginMock.mockReset().mockResolvedValue(undefined);
  });

  it("DEV ではデモアカウント欄に約10件の staff-attach アカウントを表示する", () => {
    render(
      <MemoryRouter>
        <LoginForm />
      </MemoryRouter>,
    );

    expect(screen.getByText("デモアカウント")).toBeInTheDocument();
    const demoEmails = screen.getAllByText(/stg-staff-\d+@example\.test/);
    expect(demoEmails.length).toBeGreaterThanOrEqual(9);
    expect(demoEmails.length).toBeLessThanOrEqual(12);
    expect(screen.queryByText(/パスワード:\s*password/i)).not.toBeInTheDocument();
    expect(screen.queryByText("hayashi@noah-vet.co.jp")).not.toBeInTheDocument();
    expect(screen.queryByText("admin@example.com")).not.toBeInTheDocument();
  });

  it("デモアカウント行のクリックは email のみ注入し password は空のまま", async () => {
    const user = userEvent.setup();
    render(
      <MemoryRouter>
        <LoginForm />
      </MemoryRouter>,
    );

    const firstDemoEmail = screen.getAllByText(/stg-staff-\d+@example\.test/)[0];
    expect(firstDemoEmail).toBeTruthy();
    const emailText = firstDemoEmail.textContent ?? "";
    const rowButton = firstDemoEmail.closest("button");
    expect(rowButton).not.toBeNull();
    await user.click(rowButton as HTMLButtonElement);

    expect(screen.getByLabelText("メールアドレス")).toHaveValue(emailText);
    expect(screen.getByLabelText("パスワード")).toHaveValue("");
  });
});
