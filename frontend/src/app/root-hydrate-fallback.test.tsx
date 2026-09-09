import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { RootHydrateFallback } from "./root-hydrate-fallback";

describe("RootHydrateFallback", () => {
  it("shows the shared non-sensitive pending shell (PERF-STG-LOGIN-A)", () => {
    render(<RootHydrateFallback />);
    expect(screen.getByRole("status")).toHaveTextContent("画面を読み込んでいます");
  });
});
