import { describe, expect, it } from "vitest";

import { RootHydrateFallback } from "./root-hydrate-fallback";
import { router } from "./router";

describe("BUG-20260906-002 HydrateFallback", () => {
  it("root route に HydrateFallback を置く", () => {
    expect(RootHydrateFallback()).toBeNull();
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
