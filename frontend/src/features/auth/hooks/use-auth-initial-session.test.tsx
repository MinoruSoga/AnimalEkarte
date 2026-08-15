import { Suspense } from "react";
import { act, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter, useLocation, useNavigate } from "react-router";
import { afterAll, beforeEach, describe, expect, it, vi } from "vitest";
import { useAuth } from "@/hooks/use-auth";
import type { AuthUser } from "@/types/auth";

const { loginMock, logoutMock, queryClientMock, refreshTokenMock } = vi.hoisted(() => ({
  loginMock: vi.fn(),
  logoutMock: vi.fn(),
  queryClientMock: { clear: vi.fn() },
  refreshTokenMock: vi.fn().mockResolvedValue(null),
}));

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
      <span data-testid="auth-state">
        {auth.isAuthenticated ? auth.user?.id : "anonymous"}
      </span>
      <button type="button" onClick={() => void navigate("/reset-password/?token=test-token")}>
        go-reset
      </button>
      <button type="button" onClick={() => void navigate("/login")}>
        go-login
      </button>
      <button type="button" onClick={() => void navigate("/")}>
        go-protected
      </button>
      <button
        type="button"
        onClick={() => void auth.login("staff@example.com", "password")}
      >
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
    loginMock.mockReset().mockResolvedValue({ user: AUTH_USER });
    logoutMock.mockReset().mockResolvedValue(undefined);
    refreshTokenMock.mockReset().mockResolvedValue(null);
    queryClientMock.clear.mockReset();
  });

  afterAll(() => {
    Object.defineProperty(window, "location", {
      configurable: true,
      writable: true,
      value: originalLocation,
    });
  });

  it("skips password-recovery public routes and restores once on login (BUG-031)", async () => {
    setWindowLocation("/forgot-password/");
    const { AuthProvider } = await import("./use-auth");

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
    expect(refreshTokenMock).not.toHaveBeenCalled();

    fireEvent.click(screen.getByRole("button", { name: "go-reset" }));
    expect(screen.getByTestId("pathname")).toHaveTextContent("/reset-password/");
    expect(refreshTokenMock).not.toHaveBeenCalled();

    await act(async () => {
      fireEvent.click(screen.getByRole("button", { name: "go-login" }));
    });

    await waitFor(() =>
      expect(screen.getByTestId("pathname")).toHaveTextContent("/login"),
    );
    // BUG-031: /login hydrates session so authenticated cookie users redirect.
    await waitFor(() => expect(refreshTokenMock).toHaveBeenCalledOnce());
  });

  it("hydrates valid session on cold /login and exposes authenticated state (BUG-031)", async () => {
    setWindowLocation("/login");
    refreshTokenMock.mockResolvedValueOnce({ user: AUTH_USER });
    const { AuthProvider } = await import("./use-auth");

    render(
      <MemoryRouter initialEntries={["/login"]}>
        <Suspense fallback={<div>loading</div>}>
          <AuthProvider>
            <RouteControls />
          </AuthProvider>
        </Suspense>
      </MemoryRouter>,
    );

    await waitFor(() => expect(refreshTokenMock).toHaveBeenCalledOnce());
    expect(await screen.findByTestId("auth-state")).toHaveTextContent(AUTH_USER.id);
  });

  it("takes a fresh session snapshot after login when returning from recovery to a protected route", async () => {
    setWindowLocation("/login");
    refreshTokenMock.mockResolvedValue(null);
    const { AuthProvider } = await import("./use-auth");

    render(
      <MemoryRouter initialEntries={["/login"]}>
        <Suspense fallback={<div>loading</div>}>
          <AuthProvider>
            <RouteControls />
          </AuthProvider>
        </Suspense>
      </MemoryRouter>,
    );

    await waitFor(() => expect(refreshTokenMock).toHaveBeenCalledTimes(1));
    expect(await screen.findByTestId("auth-state")).toHaveTextContent("anonymous");

    fireEvent.click(screen.getByRole("button", { name: "authenticate" }));
    await waitFor(() =>
      expect(screen.getByTestId("auth-state")).toHaveTextContent(AUTH_USER.id),
    );

    fireEvent.click(screen.getByRole("button", { name: "go-reset" }));
    expect(await screen.findByTestId("auth-state")).toHaveTextContent("anonymous");

    refreshTokenMock.mockResolvedValueOnce({ user: AUTH_USER });
    fireEvent.click(screen.getByRole("button", { name: "go-protected" }));

    await waitFor(() => expect(refreshTokenMock).toHaveBeenCalledTimes(2));
    await waitFor(() =>
      expect(screen.getByTestId("auth-state")).toHaveTextContent(AUTH_USER.id),
    );
  });

  it("takes a fresh anonymous snapshot after logout when returning from recovery to a protected route", async () => {
    setWindowLocation("/");
    refreshTokenMock
      .mockResolvedValueOnce({ user: AUTH_USER })
      .mockResolvedValueOnce(null);
    const { AuthProvider } = await import("./use-auth");

    render(
      <MemoryRouter initialEntries={["/"]}>
        <Suspense fallback={<div>loading</div>}>
          <AuthProvider>
            <RouteControls />
          </AuthProvider>
        </Suspense>
      </MemoryRouter>,
    );

    await waitFor(() => expect(refreshTokenMock).toHaveBeenCalledTimes(1));
    expect(await screen.findByTestId("auth-state")).toHaveTextContent(AUTH_USER.id);

    fireEvent.click(screen.getByRole("button", { name: "sign-out" }));
    await waitFor(() =>
      expect(screen.getByTestId("auth-state")).toHaveTextContent("anonymous"),
    );

    fireEvent.click(screen.getByRole("button", { name: "go-reset" }));
    expect(await screen.findByTestId("auth-state")).toHaveTextContent("anonymous");

    fireEvent.click(screen.getByRole("button", { name: "go-protected" }));

    await waitFor(() => expect(refreshTokenMock).toHaveBeenCalledTimes(2));
    expect(await screen.findByTestId("auth-state")).toHaveTextContent("anonymous");
  });
});
