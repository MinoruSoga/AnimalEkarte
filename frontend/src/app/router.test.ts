import { createElement, type ComponentType, type ReactElement, type ReactNode } from "react";
import { act, cleanup, render, screen, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { createMemoryRouter, RouterProvider } from "react-router";

import { RootHydrateFallback } from "./root-hydrate-fallback";
import { router } from "./router";

afterEach(() => {
  cleanup();
  vi.resetModules();
  vi.doUnmock("@/features/auth/provider");
});

describe("BUG-20260906-002 HydrateFallback", () => {
  it("root route に production HydrateFallback を置き、未解決 lazy の間は確認中表示を出す (PERF-STG-LOGIN-A)", async () => {
    const root = router.routes[0] as {
      HydrateFallback?: ComponentType;
      hydrateFallback?: ComponentType;
      hydrateFallbackElement?: ReactNode;
    };
    // createBrowserRouter materializes HydrateFallback as hydrateFallbackElement.
    if (root.hydrateFallbackElement == null) {
      throw new Error(
        "production root hydrateFallbackElement is not configured on router.routes[0]",
      );
    }
    expect(RootHydrateFallback).toBeTypeOf("function");

    function ProductionHydrateFallbackFromRouter() {
      return root.hydrateFallbackElement as ReactElement;
    }

    let releaseLazy!: (value: { Component: () => ReactElement }) => void;
    const lazyGate = new Promise<{ Component: () => ReactElement }>((resolve) => {
      releaseLazy = resolve;
    });

    const memoryRouter = createMemoryRouter(
      [
        {
          id: "root",
          path: "/",
          HydrateFallback: ProductionHydrateFallbackFromRouter,
          lazy: () => lazyGate,
        },
      ],
      { initialEntries: ["/"] },
    );

    render(createElement(RouterProvider, { router: memoryRouter }));

    expect(await screen.findByRole("status")).toHaveTextContent("画面を読み込んでいます");
    expect(screen.queryByTestId("router-hydrated-child")).not.toBeInTheDocument();

    await act(async () => {
      releaseLazy({
        Component: () =>
          createElement("div", { "data-testid": "router-hydrated-child" }, "router-hydrated"),
      });
    });

    expect(await screen.findByTestId("router-hydrated-child")).toHaveTextContent("router-hydrated");
    await waitFor(() => {
      expect(screen.queryByRole("status")).not.toBeInTheDocument();
    });
  });
});

describe("root Suspense pending-to-resolved (PERF-STG-LOGIN-A)", () => {
  it("shows the production root Suspense SessionPending until AuthProvider resolves, then renders children", async () => {
    let releaseProvider!: () => void;
    const providerGate = new Promise<void>((resolve) => {
      releaseProvider = resolve;
    });
    let providerReady = false;

    vi.resetModules();
    vi.doMock("@/features/auth/provider", () => ({
      AuthProvider: ({ children }: { children: ReactNode }) => {
        if (!providerReady) {
          throw providerGate;
        }
        return children;
      },
    }));

    const { router: freshRouter } = await import("./router");
    const memoryRouter = createMemoryRouter(
      [
        {
          path: "/",
          element: freshRouter.routes[0]?.element,
          children: [
            {
              index: true,
              element: createElement("div", { "data-testid": "root-suspense-child" }, "root-ready"),
            },
          ],
        },
      ],
      { initialEntries: ["/"] },
    );

    render(createElement(RouterProvider, { router: memoryRouter }));

    expect(await screen.findByRole("status")).toHaveTextContent("画面を読み込んでいます");
    expect(screen.queryByTestId("root-suspense-child")).not.toBeInTheDocument();

    await act(async () => {
      providerReady = true;
      releaseProvider();
    });

    expect(await screen.findByTestId("root-suspense-child")).toHaveTextContent("root-ready");
    await waitFor(() => {
      expect(screen.queryByRole("status")).not.toBeInTheDocument();
    });
  });
});
