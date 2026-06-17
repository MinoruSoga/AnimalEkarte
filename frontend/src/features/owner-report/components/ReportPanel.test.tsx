import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { ReportPanel } from "./ReportPanel";

describe("ReportPanel", () => {
  it("見出しテキストを accessible name に持つ region ランドマークになる", () => {
    render(
      <ReportPanel title="予防接種履歴">
        <p>本文</p>
      </ReportPanel>,
    );
    // <section aria-labelledby> + <h2> により、名前付き region として公開される。
    expect(screen.getByRole("region", { name: "予防接種履歴" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "予防接種履歴" })).toBeInTheDocument();
  });

  it("本文はキーボードでスクロール可能な境界コンテナ(tabindex=0, overflow-auto)に入る", () => {
    render(
      <ReportPanel title="治療履歴">
        <p data-testid="panel-child">行</p>
      </ReportPanel>,
    );
    const child = screen.getByTestId("panel-child");
    // 直近の祖先スクロールコンテナ = 本文。tabindex=0 でフォーカス可能であること。
    const scrollBody = child.parentElement as HTMLElement;
    expect(scrollBody).toHaveAttribute("tabindex", "0");
    expect(scrollBody.className).toContain("overflow-auto");
    expect(scrollBody.className).toContain("min-h-0");
  });

  it("count を渡すとヘッダーに件数バッジを表示する", () => {
    render(
      <ReportPanel title="投薬履歴" count={12}>
        <p>本文</p>
      </ReportPanel>,
    );
    expect(screen.getByText("12")).toBeInTheDocument();
  });

  it("count 未指定なら件数バッジを表示しない", () => {
    render(
      <ReportPanel title="投薬履歴">
        <p>本文</p>
      </ReportPanel>,
    );
    // ヘッダーには見出ししか存在しない（数値バッジなし）。
    expect(screen.queryByText(/^\d+$/)).not.toBeInTheDocument();
  });
});
