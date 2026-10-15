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
