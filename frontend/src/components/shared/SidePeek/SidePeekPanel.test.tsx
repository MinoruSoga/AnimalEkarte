import { render } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { SidePeekPanel } from "./SidePeekPanel";

/**
 * EMR-411 C-3: viewport <1280px では SidePeek がリストの後ろに隠れないよう、
 * 共有の位置トークン（max-xl:absolute inset-y-0 right-0）をパネルへ適用する。
 * JSDOM では viewport 依存の実レイアウトは検証できないため、ここでは
 * ブレークポイント付きクラスの DOM 適用だけを検証する。
 * （実際のオーバーレイ挙動はブラウザ検証待ち）
 */
describe("SidePeekPanel responsive overlay contract (EMR-411 C-3)", () => {
  it("LAYOUT.sidePeek.position / width トークンの全クラスをパネルに適用する", () => {
    const { container } = render(
      <SidePeekPanel>
        <div>panel body</div>
      </SidePeekPanel>,
    );

    const panel = container.querySelector("[data-side-peek]");
    expect(panel).not.toBeNull();

    // max-xl で absolute overlay（下のリストの上に右寄せ）・xl 以上では通常 flex 配置
    expect(panel).toHaveClass(
      "relative",
      "max-xl:absolute",
      "max-xl:inset-y-0",
      "max-xl:right-0",
      "max-xl:z-20",
    );
    expect(panel).toHaveClass("w-[520px]", "max-w-full", "shrink-0");
  });
});
