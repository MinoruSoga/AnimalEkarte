import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { LAYOUT } from "@/lib/design-tokens";
import { PageLayout } from "./PageLayout";

describe("PageLayout", () => {
  it("maxWidth 未指定時は設定系フォーム向けの既定値（1440px + 中央寄せ）を使う", () => {
    render(
      <PageLayout title="Test">
        <div data-testid="content">content</div>
      </PageLayout>,
    );
    const contentWrapper = screen.getByTestId("content").parentElement;
    expect(contentWrapper?.className).toContain(LAYOUT.pageContentMaxWidth.default);
    expect(contentWrapper?.className).toContain("mx-auto");
  });

  it('maxWidth="max-w-full"（一覧・カルテ詳細パターン）を指定すると全幅になり中央寄せの余白が付かない', () => {
    render(
      <PageLayout title="Test" maxWidth={LAYOUT.pageContentMaxWidth.full}>
        <div data-testid="content">content</div>
      </PageLayout>,
    );
    const contentWrapper = screen.getByTestId("content").parentElement;
    expect(contentWrapper?.className).toContain(LAYOUT.pageContentMaxWidth.full);
    expect(contentWrapper?.className).not.toContain(LAYOUT.pageContentMaxWidth.default);
  });

  it("コンテンツの縦余白は DESIGN.md spacing.lg（24px）を使い、仕様外の20pxを使わない", () => {
    render(
      <PageLayout title="Test">
        <div data-testid="content">content</div>
      </PageLayout>,
    );

    const contentWrapper = screen.getByTestId("content").parentElement;
    expect(contentWrapper).toHaveClass("py-6");
    expect(contentWrapper).not.toHaveClass("py-5");
  });

  it("ページ説明を FormHeader の補助本文として表示する", () => {
    render(
      <PageLayout title="Test" description="ページの説明">
        content
      </PageLayout>,
    );

    expect(screen.getByText("ページの説明")).toBeInTheDocument();
  });

  it("狭い画面で子要素のmin-content幅に押し広げられないflex境界を持つ", () => {
    render(
      <PageLayout title="Test">
        <div data-testid="wide-content">content</div>
      </PageLayout>,
    );

    const contentWrapper = screen.getByTestId("wide-content").parentElement;
    const scrollContainer = contentWrapper?.parentElement;
    const pageRoot = scrollContainer?.parentElement;

    expect(contentWrapper).toHaveClass("min-w-0");
    expect(scrollContainer).toHaveClass("min-w-0");
    expect(pageRoot).toHaveClass("min-w-0");
  });
});
