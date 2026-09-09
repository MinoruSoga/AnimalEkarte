import { StrictMode, Suspense } from "react";
import { act, cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter, useLocation, useNavigate } from "react-router";
import { afterAll, afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { useAuth } from "@/hooks/use-auth";
import type { AuthUser } from "@/types/auth";
import { CURRENT_CLINIC_STORAGE_KEY } from "@/lib/current-clinic";
import {
  areClinicWritesPaused,
  clearClinicSelectionRecovery,
  CLINIC_SELECTION_UNAVAILABLE,
  recoverClinicSelectionOnce,
} from "@/lib/clinic-selection-recovery";
import type { SessionRestoreResult } from "../api/restore-session";

const { loginMock, logoutMock, queryClientMock, refreshTokenMock, restoreSessionMock } = vi.hoisted(
  () => ({
    loginMock: vi.fn(),
    logoutMock: vi.fn(),
    queryClientMock: { clear: vi.fn(), setQueryData: vi.fn() },
    refreshTokenMock: vi.fn().mockResolvedValue(null),
    restoreSessionMock: vi.fn().mockResolvedValue({ kind: "anonymous401" }),
  }),
);

const AUTH_USER: AuthUser = {
  id: "staff-1",
  email: "staff@example.com",
  displayName: "Test Staff",
  isSystemAdmin: false,
  mainClinicId: "clinic-1",
  clinic: null,
  clinics: [
    {
      clinicId: "clinic-1",
      clinicName: "Test Clinic",
      isMain: true,
    },
  ],
  permissions: {},
};

vi.mock("../api/login", () => ({
  login: loginMock,
}));

vi.mock("../api/logout", () => ({
  logout: logoutMock,
}));

vi.mock("../api/refresh-token", () => ({
  refreshToken: refreshTokenMock,
}));

vi.mock("../api/restore-session", () => ({
  restoreSession: restoreSessionMock,
}));

vi.mock("../api/get-me", () => ({
  useGetMe: vi.fn().mockReturnValue({ data: undefined }),
}));

vi.mock("@tanstack/react-query", async (importActual) => {
  const actual = await importActual<typeof import("@tanstack/react-query")>();
  return { ...actual, useQueryClient: () => queryClientMock };
});

const originalLocation = window.location;

function setWindowLocation(path: string): void {
  const url = new URL(path, "http://localhost");
  Object.defineProperty(window, "location", {
    configurable: true,
    writable: true,
    value: {
      ...originalLocation,
      href: url.href,
      pathname: url.pathname,
      search: url.search,
    },
  });
}

function RouteControls() {
  const location = useLocation();
  const navigate = useNavigate();
  const auth = useAuth();

  return (
    <div>
      <span data-testid="pathname">{location.pathname}</span>
      <span data-testid="auth-state">{auth.isAuthenticated ? auth.user?.id : "anonymous"}</span>
      <button type="button" onClick={() => void navigate("/reset-password/?token=test-token")}>
        go-reset
      </button>
      <button type="button" onClick={() => void navigate("/login")}>
        go-login
      </button>
      <button type="button" onClick={() => void navigate("/")}>
        go-protected
      </button>
      <button type="button" onClick={() => void auth.login("staff@example.com", "password")}>
        authenticate
      </button>
      <button type="button" onClick={() => void auth.logout()}>
        sign-out
      </button>
    </div>
  );
}

describe("AuthProvider initial session restoration", () => {
  beforeEach(() => {
    clearClinicSelectionRecovery();
    loginMock.mockReset().mockResolvedValue({ user: AUTH_USER });
    logoutMock.mockReset().mockResolvedValue(undefined);
    refreshTokenMock.mockReset().mockResolvedValue(null);
    restoreSessionMock.mockReset().mockResolvedValue({ kind: "anonymous401" });
    queryClientMock.clear.mockReset();
    queryClientMock.setQueryData.mockReset();
    vi.useRealTimers();
  });

  afterEach(() => {
    cleanup();
    clearClinicSelectionRecovery();
    vi.unstubAllGlobals();
  });

  afterAll(() => {
    Object.defineProperty(window, "location", {
      configurable: true,
      writable: true,
      value: originalLocation,
    });
  });

  it("shows non-sensitive pending UI while restore Promise is unresolved and keeps protected children unmounted (PERF-STG-LOGIN-A)", async () => {
    setWindowLocation("/login");
    let resolveRestore: (value: SessionRestoreResult) => void = () => undefined;
    restoreSessionMock.mockImplementation(
      () =>
        new Promise<SessionRestoreResult>((resolve) => {
          resolveRestore = resolve;
        }),
    );

    const protectedMount = vi.fn();
    const businessFetch = vi.fn();
    function ProtectedChild() {
      protectedMount();
      businessFetch();
      return <div data-testid="protected-child">protected</div>;
    }

    const { AuthProvider } = await import("../components/AuthProvider");
    render(
      <MemoryRouter initialEntries={["/login"]}>
        <Suspense fallback={<div>loading</div>}>
          <AuthProvider>
            <ProtectedChild />
            <RouteControls />
          </AuthProvider>
        </Suspense>
      </MemoryRouter>,
    );

    expect(await screen.findByRole("status")).toHaveTextContent("ログイン状態を確認しています");
    expect(screen.queryByTestId("protected-child")).not.toBeInTheDocument();
    expect(screen.queryByTestId("auth-state")).not.toBeInTheDocument();
    expect(protectedMount).not.toHaveBeenCalled();
    expect(businessFetch).not.toHaveBeenCalled();
    expect(restoreSessionMock).toHaveBeenCalledOnce();
    expect(refreshTokenMock).not.toHaveBeenCalled();

    await act(async () => {
      resolveRestore({ kind: "anonymous401" });
    });

    expect(await screen.findByTestId("auth-state")).toHaveTextContent("anonymous");
    expect(screen.queryByRole("status")).not.toBeInTheDocument();
    expect(screen.getByTestId("protected-child")).toBeInTheDocument();
    expect(protectedMount).toHaveBeenCalled();
  });

  it("skips password-recovery public routes and restores once on login (BUG-031)", async () => {
    setWindowLocation("/forgot-password/");
    const { AuthProvider } = await import("../components/AuthProvider");

    await act(async () => {
      render(
        <MemoryRouter initialEntries={["/forgot-password/"]}>
          <Suspense fallback={<div>loading</div>}>
            <AuthProvider>
              <RouteControls />
            </AuthProvider>
          </Suspense>
        </MemoryRouter>,
      );
    });

    await screen.findByTestId("pathname");
    expect(restoreSessionMock).not.toHaveBeenCalled();
    expect(refreshTokenMock).not.toHaveBeenCalled();

    fireEvent.click(screen.getByRole("button", { name: "go-reset" }));
    expect(screen.getByTestId("pathname")).toHaveTextContent("/reset-password/");
    expect(restoreSessionMock).not.toHaveBeenCalled();

    await act(async () => {
      fireEvent.click(screen.getByRole("button", { name: "go-login" }));
    });

    await waitFor(() => expect(screen.getByTestId("pathname")).toHaveTextContent("/login"));
    // BUG-031: /login hydrates session so authenticated cookie users redirect.
    await waitFor(() => expect(restoreSessionMock).toHaveBeenCalledOnce());
    expect(refreshTokenMock).not.toHaveBeenCalled();
  });

  it("hydrates valid session on cold /login and exposes authenticated state (BUG-031)", async () => {
    setWindowLocation("/login");
    restoreSessionMock.mockResolvedValueOnce({ kind: "verified200", user: AUTH_USER });
    const { AuthProvider } = await import("../components/AuthProvider");

    render(
      <MemoryRouter initialEntries={["/login"]}>
        <Suspense fallback={<div>loading</div>}>
          <AuthProvider>
            <RouteControls />
          </AuthProvider>
        </Suspense>
      </MemoryRouter>,
    );

    await waitFor(() => expect(restoreSessionMock).toHaveBeenCalledOnce());
    expect(await screen.findByTestId("auth-state")).toHaveTextContent(AUTH_USER.id);
    expect(queryClientMock.setQueryData).toHaveBeenCalledWith(["me"], AUTH_USER);
    expect(refreshTokenMock).not.toHaveBeenCalled();
  });

  it("takes a fresh session snapshot after login when returning from recovery to a protected route", async () => {
    setWindowLocation("/login");
    restoreSessionMock.mockResolvedValue({ kind: "anonymous401" });
    const { AuthProvider } = await import("../components/AuthProvider");

    render(
      <MemoryRouter initialEntries={["/login"]}>
        <Suspense fallback={<div>loading</div>}>
          <AuthProvider>
            <RouteControls />
          </AuthProvider>
        </Suspense>
      </MemoryRouter>,
    );

    await waitFor(() => expect(restoreSessionMock).toHaveBeenCalledTimes(1));
    expect(await screen.findByTestId("auth-state")).toHaveTextContent("anonymous");

    fireEvent.click(screen.getByRole("button", { name: "authenticate" }));
    await waitFor(() => expect(screen.getByTestId("auth-state")).toHaveTextContent(AUTH_USER.id));

    fireEvent.click(screen.getByRole("button", { name: "go-reset" }));
    expect(await screen.findByTestId("auth-state")).toHaveTextContent("anonymous");

    restoreSessionMock.mockResolvedValueOnce({ kind: "verified200", user: AUTH_USER });
    fireEvent.click(screen.getByRole("button", { name: "go-protected" }));

    await waitFor(() => expect(restoreSessionMock).toHaveBeenCalledTimes(2));
    await waitFor(() => expect(screen.getByTestId("auth-state")).toHaveTextContent(AUTH_USER.id));
  });

  it("takes a fresh anonymous snapshot after logout when returning from recovery to a protected route", async () => {
    setWindowLocation("/");
    restoreSessionMock
      .mockResolvedValueOnce({ kind: "verified200", user: AUTH_USER })
      .mockResolvedValueOnce({ kind: "anonymous401" });
    const { AuthProvider } = await import("../components/AuthProvider");

    render(
      <MemoryRouter initialEntries={["/"]}>
        <Suspense fallback={<div>loading</div>}>
          <AuthProvider>
            <RouteControls />
          </AuthProvider>
        </Suspense>
      </MemoryRouter>,
    );

    await waitFor(() => expect(restoreSessionMock).toHaveBeenCalledTimes(1));
    expect(await screen.findByTestId("auth-state")).toHaveTextContent(AUTH_USER.id);

    fireEvent.click(screen.getByRole("button", { name: "sign-out" }));
    await waitFor(() => expect(screen.getByTestId("auth-state")).toHaveTextContent("anonymous"));

    fireEvent.click(screen.getByRole("button", { name: "go-reset" }));
    expect(await screen.findByTestId("auth-state")).toHaveTextContent("anonymous");

    fireEvent.click(screen.getByRole("button", { name: "go-protected" }));

    await waitFor(() => expect(restoreSessionMock).toHaveBeenCalledTimes(2));
    expect(await screen.findByTestId("auth-state")).toHaveTextContent("anonymous");
  });

  it("cancels clinic recovery before awaiting the logout response", async () => {
    setWindowLocation("/");
    const reload = vi.fn();
    Object.defineProperty(window.location, "reload", { configurable: true, value: reload });
    localStorage.setItem(CURRENT_CLINIC_STORAGE_KEY, "1");
    restoreSessionMock.mockResolvedValueOnce({ kind: "verified200", user: AUTH_USER });
    let finishLogout: () => void = () => undefined;
    logoutMock.mockImplementation(
      () =>
        new Promise<void>((resolve) => {
          finishLogout = resolve;
        }),
    );
    let finishRecovery: (response: Response) => void = () => undefined;
    const request = vi.fn(
      () =>
        new Promise<Response>((resolve) => {
          finishRecovery = resolve;
        }),
    );
    vi.stubGlobal("fetch", request);
    const { AuthProvider } = await import("../components/AuthProvider");
    render(
      <MemoryRouter>
        <AuthProvider>
          <RouteControls />
        </AuthProvider>
      </MemoryRouter>,
    );
    await screen.findByTestId("auth-state");

    const recovery = recoverClinicSelectionOnce();
    fireEvent.click(screen.getByRole("button", { name: "sign-out" }));
    expect(logoutMock).toHaveBeenCalledOnce();
    await act(async () => {
      finishRecovery(
        new Response(JSON.stringify({ main_clinic_id: "2", clinics: [{ clinic_id: "2" }] })),
      );
      await recovery;
    });
    expect(reload).not.toHaveBeenCalled();
    expect(localStorage.getItem(CURRENT_CLINIC_STORAGE_KEY)).toBe("1");
    expect(areClinicWritesPaused()).toBe(true);
    await recoverClinicSelectionOnce();
    expect(request).toHaveBeenCalledOnce();
    await act(async () => {
      finishLogout();
    });
    expect(screen.getByTestId("auth-state")).toHaveTextContent("anonymous");
    expect(areClinicWritesPaused()).toBe(false);
  });

  it("keeps pending through 7999ms and shows restore error at 8000ms without mounting children", async () => {
    vi.useFakeTimers();
    setWindowLocation("/login");
    restoreSessionMock.mockImplementation(() => new Promise<SessionRestoreResult>(() => undefined));
    const protectedMount = vi.fn();
    function ProtectedChild() {
      protectedMount();
      return <div data-testid="protected-child">protected</div>;
    }
    const { AuthProvider } = await import("../components/AuthProvider");
    render(
      <MemoryRouter initialEntries={["/login"]}>
        <AuthProvider>
          <ProtectedChild />
          <RouteControls />
        </AuthProvider>
      </MemoryRouter>,
    );

    expect(screen.getByRole("status")).toHaveTextContent("ログイン状態を確認しています");
    await act(async () => {
      await vi.advanceTimersByTimeAsync(7999);
    });
    expect(screen.getByRole("status")).toBeInTheDocument();
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
    expect(screen.queryByTestId("protected-child")).not.toBeInTheDocument();

    await act(async () => {
      await vi.advanceTimersByTimeAsync(1);
    });
    expect(screen.getByRole("alert")).toHaveTextContent("ログイン状態を確認できませんでした");
    expect(screen.getByRole("button", { name: "再試行" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "ログイン切替" })).toBeInTheDocument();
    expect(screen.queryByRole("status")).not.toBeInTheDocument();
    expect(screen.queryByTestId("protected-child")).not.toBeInTheDocument();
    expect(protectedMount).not.toHaveBeenCalled();
    expect(logoutMock).not.toHaveBeenCalled();
    expect(queryClientMock.clear).not.toHaveBeenCalled();
  });

  it("starts exactly one new restore attempt when retry is clicked", async () => {
    vi.useFakeTimers();
    setWindowLocation("/login");
    restoreSessionMock.mockImplementation(() => new Promise<SessionRestoreResult>(() => undefined));
    const { AuthProvider } = await import("../components/AuthProvider");
    render(
      <MemoryRouter initialEntries={["/login"]}>
        <AuthProvider>
          <RouteControls />
        </AuthProvider>
      </MemoryRouter>,
    );
    await act(async () => {
      await vi.advanceTimersByTimeAsync(8000);
    });
    expect(restoreSessionMock).toHaveBeenCalledOnce();
    fireEvent.click(screen.getByRole("button", { name: "再試行" }));
    expect(screen.getByRole("status")).toHaveTextContent("ログイン状態を確認しています");
    expect(restoreSessionMock).toHaveBeenCalledTimes(2);
  });

  it("ignores a stale verified200 after timeout", async () => {
    vi.useFakeTimers();
    setWindowLocation("/login");
    let resolveRestore: (value: SessionRestoreResult) => void = () => undefined;
    restoreSessionMock.mockImplementation(
      () =>
        new Promise<SessionRestoreResult>((resolve) => {
          resolveRestore = resolve;
        }),
    );
    const { AuthProvider } = await import("../components/AuthProvider");
    render(
      <MemoryRouter initialEntries={["/login"]}>
        <AuthProvider>
          <RouteControls />
        </AuthProvider>
      </MemoryRouter>,
    );
    await act(async () => {
      await vi.advanceTimersByTimeAsync(8000);
    });
    expect(screen.getByRole("alert")).toBeInTheDocument();
    await act(async () => {
      resolveRestore({ kind: "verified200", user: AUTH_USER });
    });
    expect(screen.queryByTestId("auth-state")).not.toBeInTheDocument();
    expect(queryClientMock.setQueryData).not.toHaveBeenCalled();
    expect(screen.getByRole("alert")).toBeInTheDocument();
  });

  it("ignores a stale anonymous401 after login-switch and successful login", async () => {
    vi.useFakeTimers();
    setWindowLocation("/login");
    let resolveRestore: (value: SessionRestoreResult) => void = () => undefined;
    restoreSessionMock.mockImplementation(
      () =>
        new Promise<SessionRestoreResult>((resolve) => {
          resolveRestore = resolve;
        }),
    );
    const { AuthProvider } = await import("../components/AuthProvider");
    render(
      <MemoryRouter initialEntries={["/login"]}>
        <AuthProvider>
          <RouteControls />
        </AuthProvider>
      </MemoryRouter>,
    );
    await act(async () => {
      await vi.advanceTimersByTimeAsync(8000);
    });
    fireEvent.click(screen.getByRole("button", { name: "ログイン切替" }));
    expect(screen.getByLabelText("メールアドレス")).toBeInTheDocument();
    vi.useRealTimers();
    fireEvent.change(screen.getByLabelText("メールアドレス"), {
      target: { value: "staff@example.com" },
    });
    fireEvent.change(screen.getByLabelText("パスワード"), {
      target: { value: "password123" },
    });
    fireEvent.click(screen.getByRole("button", { name: "ログイン" }));
    await waitFor(() => expect(loginMock).toHaveBeenCalledOnce());
    await act(async () => {
      resolveRestore({ kind: "anonymous401" });
    });
    expect(queryClientMock.setQueryData).toHaveBeenCalledWith(["me"], AUTH_USER);
    expect(queryClientMock.clear).not.toHaveBeenCalled();
    expect(screen.queryByTestId("auth-state")?.textContent ?? AUTH_USER.id).not.toBe("anonymous");
  });

  it("ignores a late verified200 after logout", async () => {
    vi.useFakeTimers();
    setWindowLocation("/login");
    let resolveRestore: (value: SessionRestoreResult) => void = () => undefined;
    restoreSessionMock.mockImplementation(
      () =>
        new Promise<SessionRestoreResult>((resolve) => {
          resolveRestore = resolve;
        }),
    );
    const { AuthProvider } = await import("../components/AuthProvider");
    render(
      <MemoryRouter initialEntries={["/login"]}>
        <AuthProvider>
          <RouteControls />
        </AuthProvider>
      </MemoryRouter>,
    );
    await act(async () => {
      await vi.advanceTimersByTimeAsync(8000);
    });
    fireEvent.click(screen.getByRole("button", { name: "ログイン切替" }));
    vi.useRealTimers();
    fireEvent.change(screen.getByLabelText("メールアドレス"), {
      target: { value: "staff@example.com" },
    });
    fireEvent.change(screen.getByLabelText("パスワード"), {
      target: { value: "password123" },
    });
    fireEvent.click(screen.getByRole("button", { name: "ログイン" }));
    await waitFor(() => expect(screen.getByTestId("auth-state")).toHaveTextContent(AUTH_USER.id));
    fireEvent.click(screen.getByRole("button", { name: "sign-out" }));
    await waitFor(() => expect(screen.getByTestId("auth-state")).toHaveTextContent("anonymous"));
    queryClientMock.setQueryData.mockClear();
    await act(async () => {
      resolveRestore({ kind: "verified200", user: AUTH_USER });
    });
    expect(screen.getByTestId("auth-state")).toHaveTextContent("anonymous");
    expect(queryClientMock.setQueryData).not.toHaveBeenCalled();
  });

  it("reuses a single restore flight across StrictMode remount", async () => {
    setWindowLocation("/login");
    restoreSessionMock.mockResolvedValue({ kind: "anonymous401" });
    const { AuthProvider } = await import("../components/AuthProvider");
    render(
      <StrictMode>
        <MemoryRouter initialEntries={["/login"]}>
          <AuthProvider>
            <RouteControls />
          </AuthProvider>
        </MemoryRouter>
      </StrictMode>,
    );
    await waitFor(() => expect(screen.getByTestId("auth-state")).toHaveTextContent("anonymous"));
    expect(restoreSessionMock).toHaveBeenCalledOnce();
  });

  it("login-switch leaves clinic storage intact and does not logout or clear queries", async () => {
    vi.useFakeTimers();
    setWindowLocation("/login");
    localStorage.setItem(CURRENT_CLINIC_STORAGE_KEY, "clinic-1");
    restoreSessionMock.mockImplementation(() => new Promise<SessionRestoreResult>(() => undefined));
    const { AuthProvider } = await import("../components/AuthProvider");
    render(
      <MemoryRouter initialEntries={["/login"]}>
        <AuthProvider>
          <RouteControls />
        </AuthProvider>
      </MemoryRouter>,
    );
    await act(async () => {
      await vi.advanceTimersByTimeAsync(8000);
    });
    fireEvent.click(screen.getByRole("button", { name: "ログイン切替" }));
    expect(screen.getByLabelText("メールアドレス")).toBeInTheDocument();
    expect(localStorage.getItem(CURRENT_CLINIC_STORAGE_KEY)).toBe("clinic-1");
    expect(logoutMock).not.toHaveBeenCalled();
    expect(queryClientMock.clear).not.toHaveBeenCalled();
    expect(screen.queryByRole("button", { name: "ログアウト" })).not.toBeInTheDocument();
  });

  it("does not render clinic 403 as anonymous Login-only", async () => {
    setWindowLocation("/login");
    restoreSessionMock.mockResolvedValue({
      kind: "restricted403",
      reason: "clinic_selection_unavailable",
    });
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue({
        ok: false,
        status: 403,
        json: async () => ({ error_code: CLINIC_SELECTION_UNAVAILABLE }),
      }),
    );
    const { AuthProvider } = await import("../components/AuthProvider");
    render(
      <MemoryRouter initialEntries={["/login"]}>
        <AuthProvider>
          <RouteControls />
        </AuthProvider>
      </MemoryRouter>,
    );
    expect(await screen.findByRole("alertdialog")).toHaveTextContent("利用できる医院がありません");
    expect(screen.queryByLabelText("メールアドレス")).not.toBeInTheDocument();
    expect(screen.queryByTestId("auth-state")).not.toBeInTheDocument();
  });

  it("ignores a late restore after true unmount to password-recovery", async () => {
    setWindowLocation("/login");
    let resolveRestore: (value: SessionRestoreResult) => void = () => undefined;
    restoreSessionMock.mockImplementation(
      () =>
        new Promise<SessionRestoreResult>((resolve) => {
          resolveRestore = resolve;
        }),
    );
    function OutsideNav() {
      const navigate = useNavigate();
      return (
        <button type="button" onClick={() => void navigate("/forgot-password/")}>
          to-recovery
        </button>
      );
    }
    const { AuthProvider } = await import("../components/AuthProvider");
    render(
      <MemoryRouter initialEntries={["/login"]}>
        <Suspense fallback={<div>loading</div>}>
          <OutsideNav />
          <AuthProvider>
            <div data-testid="protected-child">protected</div>
          </AuthProvider>
        </Suspense>
      </MemoryRouter>,
    );
    expect(await screen.findByRole("status")).toBeInTheDocument();
    expect(restoreSessionMock).toHaveBeenCalledOnce();

    await act(async () => {
      fireEvent.click(screen.getByRole("button", { name: "to-recovery" }));
    });
    await waitFor(() => expect(screen.queryByRole("status")).not.toBeInTheDocument());
    queryClientMock.setQueryData.mockClear();

    await act(async () => {
      resolveRestore({ kind: "verified200", user: AUTH_USER });
      await Promise.resolve();
      await Promise.resolve();
    });

    // Password-recovery may mount children; the critical invariant is no late ME hydrate.
    expect(queryClientMock.setQueryData).not.toHaveBeenCalled();
  });

  it("retry after timed-out clinic recovery can start recovery again", async () => {
    vi.useFakeTimers();
    setWindowLocation("/login");
    restoreSessionMock.mockResolvedValue({
      kind: "restricted403",
      reason: "clinic_selection_unavailable",
    });
    const fetchMock = vi.fn().mockImplementation(
      (_input: RequestInfo | URL, init?: RequestInit) =>
        new Promise<Response>((_resolve, reject) => {
          init?.signal?.addEventListener("abort", () => {
            reject(new DOMException("Aborted", "AbortError"));
          });
        }),
    );
    vi.stubGlobal("fetch", fetchMock);
    const { AuthProvider } = await import("../components/AuthProvider");
    render(
      <MemoryRouter initialEntries={["/login"]}>
        <AuthProvider>
          <RouteControls />
        </AuthProvider>
      </MemoryRouter>,
    );
    await act(async () => {
      await Promise.resolve();
      await Promise.resolve();
    });
    expect(fetchMock).toHaveBeenCalledOnce();
    await act(async () => {
      await vi.advanceTimersByTimeAsync(8000);
    });
    expect(screen.getByRole("alert")).toHaveTextContent("ログイン状態を確認できませんでした");
    fireEvent.click(screen.getByRole("button", { name: "再試行" }));
    await act(async () => {
      await Promise.resolve();
      await Promise.resolve();
    });
    expect(restoreSessionMock).toHaveBeenCalledTimes(2);
    expect(fetchMock).toHaveBeenCalledTimes(2);
    vi.useRealTimers();
  });
});
