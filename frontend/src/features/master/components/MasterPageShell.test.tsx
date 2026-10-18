import { render } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import type { ReactNode } from "react";
import { MasterPageShell } from "./MasterPageShell";

const pageLayoutPropsSpy = vi.fn();
vi.mock("@/components/shared/PageLayout/PageLayout", () => ({
  PageLayout: (props: { maxWidth?: string; children: ReactNode }) => {
    pageLayoutPropsSpy(props);
    return <div data-testid="page-layout-mock">{props.children}</div>;
  },
}));

vi.mock("react-router", () => ({
  useNavigate: () => vi.fn(),
}));

vi.mock("@/hooks/use-permission", () => ({
  usePermission: () => ({ canCreate: true }),
}));

describe("MasterPageShell maxWidth forwarding (EMR-333)", () => {
  it("maxWidth 未指定時はデフォルトで max-w-full を PageLayout に渡す", () => {
    pageLayoutPropsSpy.mockClear();
    render(
      <MasterPageShell title="テストマスタ" onNew={vi.fn()} sidePanel={null}>
        <div>コンテンツ</div>
      </MasterPageShell>,
    );

    expect(pageLayoutPropsSpy).toHaveBeenCalledWith(
      expect.objectContaining({
        maxWidth: "max-w-full",
      }),
    );
  });

  it("maxWidth 指定時はその値が PageLayout に渡される", () => {
    pageLayoutPropsSpy.mockClear();
    render(
      <MasterPageShell title="テストマスタ" onNew={vi.fn()} sidePanel={null} maxWidth="max-w-5xl">
        <div>コンテンツ</div>
      </MasterPageShell>,
    );

    expect(pageLayoutPropsSpy).toHaveBeenCalledWith(
      expect.objectContaining({
        maxWidth: "max-w-5xl",
      }),
    );
  });
});

// EMR-411 C-3: side peek が viewport <1280px で absolute overlay になるため、
// その定位コンテキストとしてルート flex wrapper は relative を持つ必要がある。
// 既存の flex h-full / リスト側 flex-1 min-w-0 の契約は維持する。
describe("MasterPageShell workspace positioning (EMR-411 C-3)", () => {
  it("ルートが relative flex h-full、リスト側が flex-1 min-w-0、sidePanel が兄弟要素として残る", () => {
    const { container } = render(
      <MasterPageShell
        title="テストマスタ"
        onNew={vi.fn()}
        sidePanel={<div data-testid="side-panel-marker" />}
      >
        <div>コンテンツ</div>
      </MasterPageShell>,
    );

    const root = container.firstElementChild;
    expect(root).toHaveClass("relative", "flex", "h-full");

    const listArea = container.querySelector('[data-testid="page-layout-mock"]')?.parentElement;
    expect(listArea).toHaveClass("flex-1", "min-w-0");

    const sidePanel = container.querySelector('[data-testid="side-panel-marker"]');
    expect(sidePanel).not.toBeNull();
    expect(sidePanel?.parentElement).toBe(root);
  });
});
