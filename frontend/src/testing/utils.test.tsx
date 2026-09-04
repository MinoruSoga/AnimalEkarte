import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { useLocation } from "react-router";
import { createTestWrapper } from "./TestUtils";

function LocationProbe() {
  const location = useLocation();
  return <span>{location.pathname}</span>;
}

describe("createTestWrapper (FE4-16)", () => {
  it("router なし: QueryClientProvider のみでラップする(useLocationはエラーになる)", () => {
    const Wrapper = createTestWrapper();
    expect(() => render(<span>ok</span>, { wrapper: Wrapper })).not.toThrow();
  });

  it("router: true: MemoryRouter でラップし既定パス '/' を持つ", () => {
    const Wrapper = createTestWrapper({ router: true });
    render(<LocationProbe />, { wrapper: Wrapper });
    expect(screen.getByText("/")).toBeInTheDocument();
  });

  it("initialEntries 指定: 指定パスから開始する", () => {
    const Wrapper = createTestWrapper({ initialEntries: ["/aggregation"] });
    render(<LocationProbe />, { wrapper: Wrapper });
    expect(screen.getByText("/aggregation")).toBeInTheDocument();
  });
});
