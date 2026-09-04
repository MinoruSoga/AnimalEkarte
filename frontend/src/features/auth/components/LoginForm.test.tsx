import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, useLocation } from "react-router";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { C } from "@/lib/design-tokens";
import { LoginForm } from "./LoginForm";

const { loginMock } = vi.hoisted(() => ({ loginMock: vi.fn() }));

/** Test-only fake — must never equal production staff-attach secrets. */
const TEST_DEMO_LOGIN_PASSWORD = "test-demo-pass";

vi.mock("@/hooks/use-auth", () => ({
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
    vi.stubEnv("VITE_DEMO_LOGIN_PASSWORD", TEST_DEMO_LOGIN_PASSWORD);
  });

  afterEach(() => {
    vi.unstubAllEnvs();
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
    expect(screen.getByRole("link", { name: "パスワードをお忘れですか？" })).toHaveClass(
      "min-h-11",
    );
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
      <MemoryRouter initialEntries={[{ pathname: "/login", state: { from: "/\\evil.example" } }]}>
        <LoginForm />
        <CurrentLocation />
      </MemoryRouter>,
    );

    await user.type(screen.getByLabelText("メールアドレス"), "staff@example.com");
    await user.type(screen.getByLabelText("パスワード"), "password123");
    await user.click(screen.getByRole("button", { name: "ログイン" }));

    await waitFor(() => expect(screen.getByTestId("current-location")).toHaveTextContent(/^\/$/));
  });
});

describe("LoginForm demo accounts (staff-attach)", () => {
  beforeEach(() => {
    loginMock.mockReset().mockResolvedValue(undefined);
    vi.stubEnv("VITE_DEMO_LOGIN_PASSWORD", TEST_DEMO_LOGIN_PASSWORD);
  });

  afterEach(() => {
    vi.unstubAllEnvs();
  });

  it("DEV ではデモアカウント欄に八王子・城東・敷島・猫の staff-attach アカウントを表示する", () => {
    render(
      <MemoryRouter>
        <LoginForm />
      </MemoryRouter>,
    );

    expect(screen.getByText("デモアカウント")).toBeInTheDocument();
    const demoEmails = screen.getAllByText(/stg-staff-\d+@example\.test/);
    expect(demoEmails.length).toBeGreaterThanOrEqual(39);
    expect(demoEmails.length).toBeLessThanOrEqual(42);
    // Clinic seed names use U+3000; Testing Library string matchers collapse it.
    expect(screen.getAllByText("八王子病院").length).toBeGreaterThanOrEqual(9);
    expect(screen.getAllByText(/城東センター病院/).length).toBeGreaterThanOrEqual(9);
    expect(screen.getAllByText(/敷島病院/).length).toBeGreaterThanOrEqual(9);
    expect(screen.getAllByText(/Hako bu neco/).length).toBeGreaterThanOrEqual(9);
    expect(screen.getByTestId("demo-accounts")).toHaveClass("max-w-[760px]");
    expect(
      screen.getByText("パスワードは自動入力されます（staff-attach と同一）"),
    ).toBeInTheDocument();
    expect(screen.queryByText(/パスワード:\s*password/i)).not.toBeInTheDocument();
    expect(screen.queryByText(TEST_DEMO_LOGIN_PASSWORD)).not.toBeInTheDocument();
    expect(screen.queryByText("hayashi@noah-vet.co.jp")).not.toBeInTheDocument();
    expect(screen.queryByText("admin@example.com")).not.toBeInTheDocument();
  });

  it("デモアカウント行のクリックは email と VITE_DEMO_LOGIN_PASSWORD を注入する", async () => {
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
    expect(screen.getByLabelText("パスワード")).toHaveValue(TEST_DEMO_LOGIN_PASSWORD);
    expect(screen.getByLabelText("パスワード")).not.toHaveValue("password");
  });

  it("VITE_DEMO_LOGIN_PASSWORD 未設定時は password を入れずヘルパーで案内する", async () => {
    vi.stubEnv("VITE_DEMO_LOGIN_PASSWORD", "");
    const user = userEvent.setup();
    render(
      <MemoryRouter>
        <LoginForm />
      </MemoryRouter>,
    );

    expect(
      screen.getByText(
        "デモ用パスワード未設定 — ローカルは .env.local、STG は GitHub secret を staff-attach と揃えてください",
      ),
    ).toBeInTheDocument();

    const firstDemoEmail = screen.getAllByText(/stg-staff-\d+@example\.test/)[0];
    const rowButton = firstDemoEmail.closest("button");
    expect(rowButton).not.toBeNull();
    await user.click(rowButton as HTMLButtonElement);

    expect(screen.getByLabelText("パスワード")).toHaveValue("");
  });

  it("デモアカウント一覧は縦スクロール可能な領域に収める", () => {
    render(
      <MemoryRouter>
        <LoginForm />
      </MemoryRouter>,
    );

    const firstDemoEmail = screen.getAllByText(/stg-staff-\d+@example\.test/)[0];
    const listRegion = firstDemoEmail.closest("div.overflow-y-auto");
    expect(listRegion).not.toBeNull();
    expect(listRegion).toHaveClass("overflow-y-auto");
    expect(listRegion?.className).toMatch(/max-h-/);
  });
});
