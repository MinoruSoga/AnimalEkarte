import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { PrintPortal } from "./PrintPortal";

/**
 * #187 再発防止テスト。
 *
 * 旧実装は印刷除外ルールに per-instance な testId を埋め込んでいたため、
 * `body > :not([data-testid="A"])` が同居ポータル B を巻き込んで隠し、
 * specificity 競合で全印刷面が `display:none` になり白紙化した。
 *
 * JSDOM は `@media print` のレイアウト評価を行わないため、ここでは
 * 「生成される CSS セレクタの固定キー化」と「DOM 属性」を構造検証し、
 * 白紙化の根本原因（per-instance 除外キーによる相互打ち消し）が
 * 再発しないことを保証する。
 */
function styleTextOf(testId: string): string {
  const area = screen.getByTestId(testId);
  const style = area.querySelector("style");
  return style?.textContent ?? "";
}

describe("PrintPortal (#187 多重ポータル白紙化の再発防止)", () => {
  it("印刷面は固定キー属性 data-print-portal を持つ（除外キーの安定化）", () => {
    render(
      <PrintPortal testId="area-a">
        <p>A</p>
      </PrintPortal>,
    );
    const area = screen.getByTestId("area-a");
    expect(area).toHaveAttribute("data-print-portal");
  });

  it("印刷 CSS は testId を一切参照しない（per-instance セレクタ依存の排除）", () => {
    render(
      <PrintPortal testId="area-a">
        <p>A</p>
      </PrintPortal>,
    );
    const css = styleTextOf("area-a");
    // 旧失敗モード: body > :not([data-testid="A"]) のような per-instance 除外。
    expect(css).not.toContain("data-testid");
    // 除外は全ポータル共通の固定キーで行う。
    expect(css).toContain(":not([data-print-portal])");
  });

  it("2 ポータル同居時、互いの除外ルールが相手を隠さない（白紙化しない）", () => {
    render(
      <>
        <PrintPortal testId="area-a">
          <p>A</p>
        </PrintPortal>
        <PrintPortal testId="area-b">
          <p>B</p>
        </PrintPortal>
      </>,
    );
    const a = screen.getByTestId("area-a");
    const b = screen.getByTestId("area-b");

    // 両ポータルとも固定キーを持つ → 互いの :not([data-print-portal]) から除外される。
    expect(a).toHaveAttribute("data-print-portal");
    expect(b).toHaveAttribute("data-print-portal");

    // どちらの style も相手の testId を参照しない（A!=B でも双方消滅しない）。
    expect(styleTextOf("area-a")).not.toContain("area-b");
    expect(styleTextOf("area-b")).not.toContain("area-a");

    // 既定はアクティブ（最低 1 面が必ず残る＝意図しない全消失の防止）。
    expect(a).toHaveAttribute("data-print-active", "true");
    expect(b).toHaveAttribute("data-print-active", "true");
  });

  it("active={false} のポータルのみ非アクティブ指定され、印刷面を 1 件へ限定できる", () => {
    render(
      <>
        <PrintPortal testId="area-a" active>
          <p>A</p>
        </PrintPortal>
        <PrintPortal testId="area-b" active={false}>
          <p>B</p>
        </PrintPortal>
      </>,
    );
    expect(screen.getByTestId("area-a")).toHaveAttribute("data-print-active", "true");
    expect(screen.getByTestId("area-b")).toHaveAttribute("data-print-active", "false");

    // 非アクティブ専用の隠蔽ルールが存在し、アクティブ側は表示ルールが残る。
    const css = styleTextOf("area-b");
    expect(css).toContain('[data-print-portal][data-print-active="false"]');
  });

  it("orientation を landscape にすると @page が landscape になる", () => {
    render(
      <PrintPortal testId="area-a" orientation="landscape">
        <p>A</p>
      </PrintPortal>,
    );
    expect(styleTextOf("area-a")).toContain("size: A4 landscape");
  });
});
