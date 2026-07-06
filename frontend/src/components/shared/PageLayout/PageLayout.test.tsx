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
});
