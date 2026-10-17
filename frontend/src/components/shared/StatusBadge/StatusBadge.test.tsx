import { describe, expect, it } from "vitest";
import { render, screen } from "@testing-library/react";
import { StatusBadge } from "./StatusBadge";

describe("StatusBadge", () => {
  it("既定は text-2xs / font-semibold / h-7 / px-2 の outline badge を描画する", () => {
    render(<StatusBadge>診察中</StatusBadge>);
    const badge = screen.getByText("診察中");
    expect(badge.className).toContain("text-2xs");
    expect(badge.className).toContain("font-semibold");
    expect(badge.className).not.toContain("text-base");
    expect(badge.className).not.toContain("font-normal");
    expect(badge.className).toContain("h-7");
    expect(badge.className).toContain("px-2");
    expect(badge.className).toContain("border");
  });

  it("colorClass を合成し、呼び出し側 className が最後に適用される", () => {
    render(
      <StatusBadge colorClass="bg-emerald-50 text-emerald-700" className="text-sm">
        完了
      </StatusBadge>,
    );
    const badge = screen.getByText("完了");
    expect(badge.className).toContain("bg-emerald-50");
    expect(badge.className).toContain("text-emerald-700");
    expect(badge.className).toContain("text-sm");
  });

  it("children の内容をそのまま表示する", () => {
    render(<StatusBadge>会計待ち</StatusBadge>);
    expect(screen.getByText("会計待ち")).toBeInTheDocument();
  });
});
