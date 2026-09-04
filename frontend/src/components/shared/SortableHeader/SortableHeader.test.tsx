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
    render(
      <SortableHeader label="飼主名" direction="none" onToggle={() => {}} variant="eyebrow" />,
    );
    const button = screen.getByRole("button", { name: "飼主名でソート" });
    for (const cls of STYLE.sectionLabel.split(" ")) {
      expect(button.className).toContain(cls);
    }
  });

  it("クリックで onToggle が呼ばれる（variant に関わらず挙動不変）", async () => {
    const onToggle = vi.fn();
    render(
      <SortableHeader label="飼主名" direction="none" onToggle={onToggle} variant="eyebrow" />,
    );
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

  describe("a11y (FE-RC-044)", () => {
    it("direction=ascending のとき aria-label に現在の並び順（昇順）が含まれる", () => {
      render(<SortableHeader label="診療日" direction="ascending" onToggle={() => {}} />);
      expect(screen.getByRole("button", { name: "診療日でソート（昇順）" })).toBeInTheDocument();
    });

    it("direction=descending のとき aria-label に現在の並び順（降順）が含まれる", () => {
      render(<SortableHeader label="診療日" direction="descending" onToggle={() => {}} />);
      expect(screen.getByRole("button", { name: "診療日でソート（降順）" })).toBeInTheDocument();
    });

    it("direction=none のときは既存の accessible name を変えない（呼び出し側テストの互換性）", () => {
      render(<SortableHeader label="診療日" direction="none" onToggle={() => {}} />);
      expect(screen.getByRole("button", { name: "診療日でソート" })).toBeInTheDocument();
    });

    it("方向アイコンは装飾であり aria-hidden を持つ（方向は aria-label 側で伝える）", () => {
      render(<SortableHeader label="診療日" direction="ascending" onToggle={() => {}} />);
      const icon = screen
        .getByRole("button", { name: "診療日でソート（昇順）" })
        .querySelector("svg");
      expect(icon).toHaveAttribute("aria-hidden", "true");
    });
  });
});
