import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import { createMemoryRouter, RouterProvider } from "react-router";

import { SUPPORT_WIDGET_LAYOUT } from "@/constants/support-widget-layout";

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
  it("main にサポートウィジェット分の下余白を確保する", () => {
    renderLayout();
    expect(screen.getByRole("main")).toHaveClass(SUPPORT_WIDGET_LAYOUT.shellBottomClearance);
  });

  it("シェル下余白定数がサポートボタンの占有高さをカバーする", () => {
    expect(SUPPORT_WIDGET_LAYOUT.shellBottomClearancePx).toBeGreaterThanOrEqual(
      SUPPORT_WIDGET_LAYOUT.footprintPx,
    );
    expect(SUPPORT_WIDGET_LAYOUT.shellBottomClearance).toBe(
      `pb-${SUPPORT_WIDGET_LAYOUT.shellBottomClearancePx / 4}`,
    );
    expect(SUPPORT_WIDGET_LAYOUT.shellBottomClearance).toBe("pb-20");
  });
});
