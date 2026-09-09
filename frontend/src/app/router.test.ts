import { isValidElement, type ReactElement, type ReactNode } from "react";
import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { RootHydrateFallback } from "./root-hydrate-fallback";
import { router } from "./router";

function suspenseFallback(element: ReactNode): ReactElement {
  if (!isValidElement(element)) {
    throw new Error("root route element is not a React element");
  }
  const fallback = (element as ReactElement<{ fallback?: ReactNode }>).props.fallback;
  if (!isValidElement(fallback)) {
    throw new Error("root Suspense fallback is missing");
  }
  return fallback;
}

describe("BUG-20260906-002 HydrateFallback", () => {
  it("root route に HydrateFallback を置き、非機密の確認中表示を出す (PERF-STG-LOGIN-A)", () => {
    render(RootHydrateFallback());
    expect(screen.getByRole("status")).toHaveTextContent("画面を読み込んでいます");
    const root = router.routes[0] as {
      HydrateFallback?: unknown;
      hydrateFallback?: unknown;
      hydrateFallbackElement?: unknown;
      element?: ReactNode;
    };
    expect(
      root.HydrateFallback ?? root.hydrateFallback ?? root.hydrateFallbackElement,
    ).toBeTruthy();
  });

  it("root Suspense fallback uses SessionPending (PERF-STG-LOGIN-A)", () => {
    const root = router.routes[0] as { element?: ReactNode };
    if (root.element === undefined) {
      throw new Error("root route element is missing");
    }
    render(suspenseFallback(root.element));
    expect(screen.getByRole("status")).toHaveTextContent("画面を読み込んでいます");
  });
});
