import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { RootHydrateFallback } from "./root-hydrate-fallback";
import { router } from "./router";

describe("BUG-20260906-002 HydrateFallback", () => {
  it("root route に HydrateFallback を置き、非機密の確認中表示を出す (PERF-STG-LOGIN-A)", () => {
    render(RootHydrateFallback());
    expect(screen.getByRole("status")).toHaveTextContent("画面を読み込んでいます");
    const root = router.routes[0] as {
      HydrateFallback?: unknown;
      hydrateFallback?: unknown;
      hydrateFallbackElement?: unknown;
    };
    expect(
      root.HydrateFallback ?? root.hydrateFallback ?? root.hydrateFallbackElement,
    ).toBeTruthy();
  });
});
