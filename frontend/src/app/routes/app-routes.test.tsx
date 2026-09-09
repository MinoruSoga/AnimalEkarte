import { isValidElement, type ReactElement, type ReactNode } from "react";
import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
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

function loginSuspenseFallback(loginElement: ReactNode): ReactElement {
  if (!isValidElement(loginElement)) {
    throw new Error("login route element is not a React element");
  }
  const fallback = (loginElement as ReactElement<{ fallback?: ReactNode }>).props.fallback;
  if (!isValidElement(fallback)) {
    throw new Error("login route Suspense fallback is missing");
  }
  return fallback;
}

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

describe("appRoutes login Suspense fallback (PERF-STG-LOGIN-A)", () => {
  it("configures SessionPending as the login chunk fallback", () => {
    const loginRoute = appRoutes.find((route) => route.path === paths.auth.login.path);
    if (loginRoute?.element === undefined) {
      throw new Error("login route is not configured");
    }

    render(loginSuspenseFallback(loginRoute.element));
    expect(screen.getByRole("status")).toHaveTextContent("画面を読み込んでいます");
  });
});
