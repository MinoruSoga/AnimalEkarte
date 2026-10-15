import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { StatusPill } from "./StatusPill";

describe("StatusPill", () => {
  it.each([
    { isActive: true, label: "有効" },
    { isActive: false, label: "無効" },
  ])("$labelは狭い列でもラベルを折り返さない", ({ isActive, label }) => {
    render(<StatusPill isActive={isActive} />);
    expect(screen.getByText(label)).toHaveClass("inline-flex", "whitespace-nowrap");
  });
});
