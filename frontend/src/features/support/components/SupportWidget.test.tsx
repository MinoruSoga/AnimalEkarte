import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { SupportWidget } from "./SupportWidget";

// BugReportTab が動的 import する html2canvas-pro への依存を切る
vi.mock("../lib/capture-screenshot", () => ({
  capturePageScreenshot: vi.fn().mockResolvedValue(null),
}));

describe("SupportWidget", () => {
  it("起動ボタンを表示する", () => {
    render(<SupportWidget />);
    expect(screen.getByRole("button", { name: "サポート・ヘルプを開く" })).toBeInTheDocument();
  });

  it("コンテナは pointer-events-auto を持ち、モーダル Dialog の body ロック中も操作できる", () => {
    render(<SupportWidget />);
    const button = screen.getByRole("button", { name: "サポート・ヘルプを開く" });
    // Radix modal Dialog は body を pointer-events:none にする。ダイアログ外の
    // ウィジェットは見えているのにクリックが背面へ透過するため、明示的に復帰させる。
    const container = button.parentElement;
    expect(container).toHaveClass("pointer-events-auto");
    expect(container).toHaveClass("fixed");
  });
});
