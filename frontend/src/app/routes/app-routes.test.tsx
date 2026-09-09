import { act, cleanup, render, screen, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { paths } from "@/config/paths";
import { C } from "@/lib/design-tokens";
import { appRoutes } from "./app-routes";

function hasClassInAncestry(element: Element | null, className: string): boolean {
  let current = element;
  while (current !== null) {
    if (current.classList.contains(className)) return true;
    current = current.parentElement;
  }
  return false;
}

afterEach(() => {
  cleanup();
  vi.resetModules();
  vi.doUnmock("@/features/auth");
});

describe("appRoutes 404 fallback", () => {
  it("DESIGN.md の canvas-soft shell 上に表示する", () => {
    const layoutRoute = appRoutes.find((route) => route.children !== undefined);
    const notFoundRoute = layoutRoute?.children?.find((route) => route.path === "*");

    if (notFoundRoute?.element === undefined) {
      throw new Error("404 fallback route is not configured");
    }

    render(<>{notFoundRoute.element}</>);

    expect(hasClassInAncestry(screen.getByText("ページが見つかりません"), C.bgPage)).toBe(true);
  });
});

describe("appRoutes password recovery routes", () => {
  it.each([paths.auth.login.path, paths.auth.forgotPassword.path, paths.auth.resetPassword.path])(
    "registers the centralized public route %s",
    (path) => {
      expect(appRoutes.some((route) => route.path === path)).toBe(true);
    },
  );
});

describe("appRoutes login Suspense pending-to-resolved (PERF-STG-LOGIN-A)", () => {
  it("shows SessionPending on the actual login route until the lazy chunk resolves, then shows the login child", async () => {
    let releaseAuthModule!: () => void;
    const authModuleGate = new Promise<void>((resolve) => {
      releaseAuthModule = resolve;
    });

    vi.resetModules();
    vi.doMock("@/features/auth", async () => {
      await authModuleGate;
      return {
        Login: () => <div data-testid="login-resolved-child">login-ready</div>,
      };
    });

    const { paths: freshPaths } = await import("@/config/paths");
    const { appRoutes: freshRoutes } = await import("./app-routes");
    const loginRoute = freshRoutes.find((route) => route.path === freshPaths.auth.login.path);
    if (loginRoute?.element === undefined) {
      throw new Error("login route is not configured");
    }

    render(<>{loginRoute.element}</>);

    expect(await screen.findByRole("status")).toHaveTextContent("画面を読み込んでいます");
    expect(screen.queryByTestId("login-resolved-child")).not.toBeInTheDocument();

    await act(async () => {
      releaseAuthModule();
    });

    expect(await screen.findByTestId("login-resolved-child")).toHaveTextContent("login-ready");
    await waitFor(() => {
      expect(screen.queryByRole("status")).not.toBeInTheDocument();
    });
  });
});
