import type { ReactElement } from "react";
import { act, cleanup, render, screen, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it } from "vitest";
import { createMemoryRouter, RouterProvider } from "react-router";

import { RootHydrateFallback } from "./root-hydrate-fallback";

afterEach(() => {
  cleanup();
});

describe("RootHydrateFallback", () => {
  it("shows the shared non-sensitive pending shell (PERF-STG-LOGIN-A)", () => {
    render(<RootHydrateFallback />);
    expect(screen.getByRole("status")).toHaveTextContent("画面を読み込んでいます");
  });

  it("keeps production HydrateFallback visible until a deferred lazy route resolves, then shows the child", async () => {
    let releaseLazy!: (value: { Component: () => ReactElement }) => void;
    const lazyGate = new Promise<{ Component: () => ReactElement }>((resolve) => {
      releaseLazy = resolve;
    });

    const memoryRouter = createMemoryRouter(
      [
        {
          id: "hydrate-root",
          path: "/",
          HydrateFallback: RootHydrateFallback,
          lazy: () => lazyGate,
        },
      ],
      { initialEntries: ["/"] },
    );

    render(<RouterProvider router={memoryRouter} />);

    expect(await screen.findByRole("status")).toHaveTextContent("画面を読み込んでいます");
    expect(screen.queryByTestId("hydrated-resolved-child")).not.toBeInTheDocument();

    await act(async () => {
      releaseLazy({
        Component: () => <div data-testid="hydrated-resolved-child">hydrated-ready</div>,
      });
    });

    expect(await screen.findByTestId("hydrated-resolved-child")).toHaveTextContent(
      "hydrated-ready",
    );
    await waitFor(() => {
      expect(screen.queryByRole("status")).not.toBeInTheDocument();
    });
  });
});
