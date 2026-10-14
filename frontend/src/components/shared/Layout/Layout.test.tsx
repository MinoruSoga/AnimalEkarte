import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { createMemoryRouter, RouterProvider } from "react-router";

import { SUPPORT_WIDGET_LAYOUT } from "@/constants/support-widget-layout";
import { C } from "@/lib/design-tokens";

import { Layout } from "./Layout";

vi.mock("./Sidebar", () => ({ Sidebar: () => null }));
vi.mock("@/hooks/use-auth", () => ({
  useAuth: () => ({ isAuthenticated: true, isLoading: false }),
}));

function renderLayout() {
  const router = createMemoryRouter(
    [{ element: <Layout />, children: [{ path: "/", element: <div>page</div> }] }],
    { initialEntries: ["/"] },
  );
  render(<RouterProvider router={router} />);
}

describe("Layout", () => {
  it("main には固定の下余白を持たせない（クリアランスはスクロール領域側が確保する）", () => {
    renderLayout();
    expect(screen.getByRole("main")).not.toHaveClass(SUPPORT_WIDGET_LAYOUT.scrollBottomClearance);
  });

  it("main の下余白はページ背景色で塗り、ページ下端に白帯が透けない", () => {
    renderLayout();
    expect(screen.getByRole("main")).toHaveClass(C.bgPage);
  });

  it("シェル下余白定数がサポートボタンの占有高さをカバーする", () => {
    expect(SUPPORT_WIDGET_LAYOUT.scrollBottomClearancePx).toBeGreaterThanOrEqual(
      SUPPORT_WIDGET_LAYOUT.footprintPx,
    );
    expect(SUPPORT_WIDGET_LAYOUT.scrollBottomClearance).toBe(
      `pb-${SUPPORT_WIDGET_LAYOUT.scrollBottomClearancePx / 4}`,
    );
    expect(SUPPORT_WIDGET_LAYOUT.scrollBottomClearance).toBe("pb-20");
  });

  it("キーボード最初の到達点は 本文へスキップ link で、activation で main へ focus する", async () => {
    renderLayout();

    const main = screen.getByRole("main");
    expect(main).toHaveAttribute("id", "main-content");
    expect(main).toHaveAttribute("tabindex", "-1");
    expect(main).toHaveClass("min-w-0");

    const skipLink = screen.getByRole("link", { name: "本文へスキップ" });
    expect(skipLink).toHaveAttribute("href", "#main-content");

    await userEvent.tab();
    expect(skipLink).toHaveFocus();

    await userEvent.keyboard("{Enter}");
    expect(main).toHaveFocus();
  });

  it("skip link は通常非表示で focus 時だけ可視化される", () => {
    renderLayout();

    const skipLink = screen.getByRole("link", { name: "本文へスキップ" });
    expect(skipLink).toHaveClass("sr-only");
    expect(skipLink.className).toContain("focus:not-sr-only");
  });

  it("footprintPx は position(bottom-N) + buttonSizeClass(size-N) の算術と一致する", () => {
    const bottomOffsetPx = Number(SUPPORT_WIDGET_LAYOUT.position.match(/bottom-(\d+)/)?.[1]) * 4;
    const buttonSizePx = Number(SUPPORT_WIDGET_LAYOUT.buttonSizeClass.match(/size-(\d+)/)?.[1]) * 4;
    expect(bottomOffsetPx).toBeGreaterThan(0);
    expect(buttonSizePx).toBeGreaterThan(0);
    expect(SUPPORT_WIDGET_LAYOUT.footprintPx).toBe(bottomOffsetPx + buttonSizePx);
  });
});
