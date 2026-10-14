import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { RequiredMark } from "./RequiredMark";
import { C } from "@/lib/design-tokens";

describe("RequiredMark", () => {
  it("renders the * marker", () => {
    render(<RequiredMark />);
    expect(screen.getByText("*")).toBeInTheDocument();
  });

  it("is hidden from assistive technology", () => {
    render(<RequiredMark />);
    expect(screen.getByText("*")).toHaveAttribute("aria-hidden", "true");
  });

  it("uses the required-marker token color", () => {
    render(<RequiredMark />);
    expect(screen.getByText("*")).toHaveClass(C.textRequired);
  });
});
