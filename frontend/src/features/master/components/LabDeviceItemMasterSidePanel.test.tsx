import { createMemoryRouter, RouterProvider } from "react-router";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { LAYOUT } from "@/lib/design-tokens";

import { LabDeviceItemMasterSidePanel } from "./LabDeviceItemMasterSidePanel";

/**
 * EMR-411 C-3: 検査機器パネルは 640px 幅。viewport <1280px でも 640px 指定を
 * 維持しつつ max-w-full でクランプする（min-w-160 が小画面で
 * max-width を打ち消さないよう xl: スコープ化）。
 * JSDOM では実レイアウトは検証できないため DOM クラス適用のみ確認する。
 */
function renderPanel() {
  // NavigationBlocker (dirty 検知) が useBlocker を呼ぶため data router 配下で描画する
  const router = createMemoryRouter(
    [
      {
        path: "*",
        element: (
          <LabDeviceItemMasterSidePanel
            device={null}
            items={[]}
            examTypes={[]}
            unusedSourceTypes={[]}
            onClose={vi.fn()}
            onSave={vi.fn()}
          />
        ),
      },
    ],
    { initialEntries: ["/"] },
  );
  return render(<RouterProvider router={router} />);
}

describe("LabDeviceItemMasterSidePanel 640px panel (EMR-411 C-3)", () => {
  it("パネルが xl 以上 640px 最小幅・max-xl 640px 幅+min-w-0・共有位置トークンを持つ", () => {
    const { container } = renderPanel();

    const panel = container.querySelector("[data-side-peek]");
    expect(panel).not.toBeNull();
    expect(panel).toHaveClass("xl:min-w-160", "max-xl:w-[640px]", "max-xl:min-w-0");
    // max-xl grid workspace (LabDeviceItemMasterSettings) でヘッダ行(1)を避け body 行(2)へ束縛する
    expect(panel).toHaveClass(
      "max-xl:col-start-1",
      "max-xl:col-end-2",
      "max-xl:row-start-2",
      "max-xl:row-end-3",
    );

    for (const cls of LAYOUT.sidePeek.position.split(" ")) {
      expect(panel).toHaveClass(cls);
    }
  });
});

describe("LabDeviceItemMasterSidePanel real StatusToggleButton (EMR-411 C-2)", () => {
  it("共有 StatusToggleButton の aria-pressed が内部状態で true→false→true に遷移する", async () => {
    const user = userEvent.setup();
    renderPanel();

    const toggle = screen.getByRole("button", { name: "ステータスを切り替え" });
    // 新規機器マスタは isActive=true が初期値
    expect(toggle).toHaveAttribute("aria-pressed", "true");

    await user.click(toggle);
    expect(toggle).toHaveAttribute("aria-pressed", "false");

    await user.click(toggle);
    expect(toggle).toHaveAttribute("aria-pressed", "true");
  });
});
