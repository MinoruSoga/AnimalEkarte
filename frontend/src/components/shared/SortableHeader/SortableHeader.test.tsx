import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import { C, STYLE } from "@/lib/design-tokens";
import { SortableHeader } from "./SortableHeader";

describe("SortableHeader", () => {
  it("既定（variant 未指定）は本文相当の濃い色を使う", () => {
    render(<SortableHeader label="飼主名" direction="none" onToggle={() => {}} />);
    const button = screen.getByRole("button", { name: "飼主名でソート" });
    expect(button.className).toContain(C.text);
    expect(button.className).not.toContain("uppercase");
  });

  it('variant="eyebrow" は DESIGN.md ex-data-table-cell の header 相当（STYLE.sectionLabel）を使う', () => {
    render(<SortableHeader label="飼主名" direction="none" onToggle={() => {}} variant="eyebrow" />);
    const button = screen.getByRole("button", { name: "飼主名でソート" });
    for (const cls of STYLE.sectionLabel.split(" ")) {
      expect(button.className).toContain(cls);
    }
  });

  it("クリックで onToggle が呼ばれる（variant に関わらず挙動不変）", async () => {
    const onToggle = vi.fn();
    render(<SortableHeader label="飼主名" direction="none" onToggle={onToggle} variant="eyebrow" />);
    screen.getByRole("button", { name: "飼主名でソート" }).click();
    expect(onToggle).toHaveBeenCalledTimes(1);
  });

  it("C18のtable cell padding内で44x44px以上の操作領域を持つ", () => {
    render(<SortableHeader label="種" direction="none" onToggle={() => {}} />);

    expect(screen.getByRole("button", { name: "種でソート" })).toHaveClass(
      "min-h-11",
      "min-w-11",
      "-my-3",
    );
  });
});
