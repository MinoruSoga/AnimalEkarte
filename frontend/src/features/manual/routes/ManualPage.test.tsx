import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { act, render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes } from "react-router";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import type { ManualArticleOverride } from "../api/get-manual-articles";
import { ManualPage } from "./ManualPage";

const { manualPermissions, overridesHookMock } = vi.hoisted(() => ({
  manualPermissions: {
    canView: true,
    canCreate: true,
    canEdit: true,
    canDelete: true,
  },
  overridesHookMock: vi.fn<
    (enabled: boolean) => { data: ManualArticleOverride[] }
  >(() => ({ data: [] })),
}));

vi.mock("@/hooks/use-permission", () => ({
  usePermission: () => manualPermissions,
}));

vi.mock("../api/get-manual-articles", () => ({
  useGetManualArticleOverrides: overridesHookMock,
}));

beforeEach(() => {
  manualPermissions.canView = true;
  manualPermissions.canCreate = true;
  manualPermissions.canEdit = true;
  manualPermissions.canDelete = true;
  overridesHookMock.mockClear();
});

afterEach(() => {
  vi.unstubAllGlobals();
});

function mockDesktopBreakpoint(initialMatches = false) {
  let viewport = { matches: initialMatches };
  const listeners = new Set<(event: MediaQueryListEvent) => void>();
  const media = "(min-width: 768px)";
  const mediaQueryList = {
    get matches() {
      return viewport.matches;
    },
    media,
    onchange: null,
    addEventListener: vi.fn(
      (_type: "change", listener: (event: MediaQueryListEvent) => void) => {
        listeners.add(listener);
      },
    ),
    removeEventListener: vi.fn(
      (_type: "change", listener: (event: MediaQueryListEvent) => void) => {
        listeners.delete(listener);
      },
    ),
  } as unknown as MediaQueryList;

  vi.stubGlobal("matchMedia", vi.fn(() => mediaQueryList));

  return {
    enterDesktop() {
      viewport = { matches: true };
      const event = { matches: true, media } as MediaQueryListEvent;
      listeners.forEach((listener) => listener(event));
    },
  };
}

describe("ManualPage touch targets", () => {
  it("manual-edit:viewがなければoverride queryを無効化してbundle版を表示する", () => {
    manualPermissions.canView = false;
    manualPermissions.canEdit = false;
    overridesHookMock.mockReturnValueOnce({
      data: [{
        id: 1,
        category: "screens",
        slug: "00-overview",
        title: "権限外のoverride",
        order_value: 0,
        section: "基本",
        body_markdown: "# 権限外のoverride",
        created_at: "2026-07-21T00:00:00Z",
        updated_at: "2026-07-21T00:00:00Z",
      }],
    });
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });

    render(
      <QueryClientProvider client={queryClient}>
        <MemoryRouter initialEntries={["/manual/screens/00-overview"]}>
          <Routes>
            <Route path="/manual/:category/:slug" element={<ManualPage />} />
          </Routes>
        </MemoryRouter>
      </QueryClientProvider>,
    );

    expect(overridesHookMock).toHaveBeenCalledWith(false);
    expect(screen.getByRole("heading", { name: "Animal Ekarte 取扱説明書" })).toBeInTheDocument();
    expect(screen.queryByRole("heading", { name: "権限外のoverride" })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "このページを編集" })).not.toBeInTheDocument();
  });

  it("manual-edit:viewがあってもeditがなければ編集buttonを表示しない", () => {
    manualPermissions.canEdit = false;
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });

    render(
      <QueryClientProvider client={queryClient}>
        <MemoryRouter initialEntries={["/manual/screens/00-overview"]}>
          <Routes>
            <Route path="/manual/:category/:slug" element={<ManualPage />} />
          </Routes>
        </MemoryRouter>
      </QueryClientProvider>,
    );

    expect(overridesHookMock).toHaveBeenCalledWith(true);
    expect(screen.queryByRole("button", { name: "このページを編集" })).not.toBeInTheDocument();
  });

  it("目次・編集・印刷buttonsを44px以上に保つ", () => {
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    render(
      <QueryClientProvider client={queryClient}>
        <MemoryRouter initialEntries={["/manual/screens/00-overview"]}>
          <Routes>
            <Route path="/manual/:category/:slug" element={<ManualPage />} />
          </Routes>
        </MemoryRouter>
      </QueryClientProvider>,
    );

    expect(screen.getByRole("button", { name: "マニュアル目次を開く" })).toHaveClass("size-11");
    expect(screen.getByRole("button", { name: "このページを編集" })).toHaveClass("size-11");
    expect(screen.getByRole("button", { name: "このページを印刷" })).toHaveClass("size-11");
  });

  it("mobile目次をmodal dialogとして操作し、Escape後にtriggerへfocusを戻す", async () => {
    const user = userEvent.setup();
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    render(
      <QueryClientProvider client={queryClient}>
        <MemoryRouter initialEntries={["/manual/screens/00-overview"]}>
          <Routes>
            <Route path="/manual/:category/:slug" element={<ManualPage />} />
          </Routes>
        </MemoryRouter>
      </QueryClientProvider>,
    );

    const trigger = screen.getByRole("button", { name: "マニュアル目次を開く" });
    const backgroundEditButton = screen.getByRole("button", { name: "このページを編集" });
    expect(trigger).toHaveAttribute("aria-expanded", "false");

    await user.click(trigger);

    const dialog = await screen.findByRole("dialog", { name: "マニュアル目次" });
    expect(dialog).toHaveAttribute("aria-modal", "true");
    expect(trigger).toHaveAttribute("aria-expanded", "true");
    await waitFor(() => {
      expect(dialog).toContainElement(document.activeElement as HTMLElement);
    });

    await user.tab({ shift: true });
    expect(dialog).toContainElement(document.activeElement as HTMLElement);
    expect(backgroundEditButton).not.toHaveFocus();

    await user.keyboard("{Escape}");
    await waitFor(() => expect(screen.queryByRole("dialog", { name: "マニュアル目次" })).not.toBeInTheDocument());
    expect(trigger).toHaveFocus();
    expect(within(document.body).getByRole("button", { name: "マニュアル目次を開く" })).toBe(trigger);
  });

  it("open中にdesktopへ変化するとmodalを閉じ、可視目次へfocusを移す", async () => {
    const breakpoint = mockDesktopBreakpoint();
    const user = userEvent.setup();
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    render(
      <QueryClientProvider client={queryClient}>
        <MemoryRouter initialEntries={["/manual/screens/00-overview"]}>
          <Routes>
            <Route path="/manual/:category/:slug" element={<ManualPage />} />
          </Routes>
        </MemoryRouter>
      </QueryClientProvider>,
    );

    const trigger = screen.getByRole("button", { name: "マニュアル目次を開く" });
    const backgroundEditButton = screen.getByRole("button", { name: "このページを編集" });
    await user.click(trigger);

    expect(await screen.findByRole("dialog", { name: "マニュアル目次" })).toBeInTheDocument();
    expect(document.querySelector('[data-slot="sheet-overlay"]')).toBeInTheDocument();
    expect(backgroundEditButton.closest('[aria-hidden="true"]')).not.toBeNull();

    act(() => breakpoint.enterDesktop());

    await waitFor(() => {
      expect(screen.queryByRole("dialog", { name: "マニュアル目次" })).not.toBeInTheDocument();
      expect(document.querySelector('[data-slot="sheet-overlay"]')).not.toBeInTheDocument();
    });
    expect(backgroundEditButton.closest('[aria-hidden="true"]')).toBeNull();
    expect(document.body).not.toHaveAttribute("data-scroll-locked");
    expect(document.body.style.pointerEvents).not.toBe("none");
    expect(trigger).not.toHaveFocus();
    expect(screen.getByRole("tab", { name: "画面別" })).toHaveFocus();
  });
});
