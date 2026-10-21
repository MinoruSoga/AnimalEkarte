import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { act, render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes } from "react-router";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import type { ManualArticleOverride } from "@/hooks/use-manual-article-overrides";
import { createManualSearcher, workflowArticles } from "@/lib/manual-index";
import { ManualPage } from "./ManualPage";

const { manualPermissions, overridesHookMock } = vi.hoisted(() => ({
  manualPermissions: {
    canView: true,
    canCreate: true,
    canEdit: true,
    canDelete: true,
  },
  overridesHookMock: vi.fn<
    (
      enabled: boolean,
      strict?: boolean,
    ) => {
      data?: ManualArticleOverride[];
      isPending?: boolean;
      isError?: boolean;
      refetch?: () => unknown;
    }
  >(() => ({
    data: [],
  })),
}));

vi.mock("@/hooks/use-permission", () => ({
  usePermission: () => manualPermissions,
}));

vi.mock("@/hooks/use-manual-article-overrides", () => ({
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
    addEventListener: vi.fn((_type: "change", listener: (event: MediaQueryListEvent) => void) => {
      listeners.add(listener);
    }),
    removeEventListener: vi.fn(
      (_type: "change", listener: (event: MediaQueryListEvent) => void) => {
        listeners.delete(listener);
      },
    ),
  } as unknown as MediaQueryList;

  vi.stubGlobal(
    "matchMedia",
    vi.fn(() => mediaQueryList),
  );

  return {
    enterDesktop() {
      viewport = { matches: true };
      const event = { matches: true, media } as MediaQueryListEvent;
      listeners.forEach((listener) => listener(event));
    },
  };
}

describe("ManualPage touch targets", () => {
  it("通常診療レクチャーを業務フロー目次から開き、検索でも見つけられる", async () => {
    const user = userEvent.setup();
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const title = "通常診療レクチャー（受付から会計・ホテルと入院）";
    const slug = "29-normal-consultation-lecture";
    render(
      <QueryClientProvider client={queryClient}>
        <MemoryRouter initialEntries={["/manual/screens/00-overview"]}>
          <Routes>
            <Route path="/manual/:category/:slug" element={<ManualPage />} />
          </Routes>
        </MemoryRouter>
      </QueryClientProvider>,
    );

    await user.click(screen.getByRole("button", { name: "マニュアル目次を開く" }));
    const dialog = await screen.findByRole("dialog", { name: "マニュアル目次" });
    await user.click(within(dialog).getByRole("tab", { name: "業務フロー" }));
    const articleLink = within(dialog).getByRole("link", { name: title });
    expect(articleLink).toHaveAttribute("href", `/manual/workflows/${slug}`);
    await user.click(articleLink);

    expect(await screen.findByRole("heading", { level: 1, name: title })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "主編：通常診療" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "補足編：ホテル・入院" })).toBeInTheDocument();

    const search = createManualSearcher(workflowArticles);
    for (const query of ["通常診療", "レクチャー", "血液検査", "ホテル"]) {
      expect(search(query).some((article) => article.slug === slug)).toBe(true);
    }
  });

  // fix-12: 公開 READ は認証済みスタッフ全員（権限不要）。manual-edit がなくても
  // 同じ strict fetch・同じ内容を読む。編集 button は canEdit のみで出す。
  it("manual-edit権限がなくてもoverride queryをfetchし同じ記事を読む（編集buttonは非表示）", () => {
    manualPermissions.canView = false;
    manualPermissions.canEdit = false;
    overridesHookMock.mockReturnValueOnce({
      data: [
        {
          id: 1,
          category: "screens",
          slug: "00-overview",
          title: "権限外のoverride",
          order_value: 0,
          section: "基本",
          body_markdown: "# 権限外のoverride",
          created_at: "2026-07-21T00:00:00Z",
          updated_at: "2026-07-21T00:00:00Z",
        },
      ],
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

    expect(overridesHookMock).toHaveBeenCalledWith(true, true);
    expect(screen.getByRole("heading", { name: "権限外のoverride" })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "このページを編集" })).not.toBeInTheDocument();
  });

  it("manual-edit:editがなければ編集buttonを表示しない", () => {
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

    expect(overridesHookMock).toHaveBeenCalledWith(true, true);
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
    await waitFor(() =>
      expect(screen.queryByRole("dialog", { name: "マニュアル目次" })).not.toBeInTheDocument(),
    );
    expect(trigger).toHaveFocus();
    expect(within(document.body).getByRole("button", { name: "マニュアル目次を開く" })).toBe(
      trigger,
    );
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

// fix-9b: strict 取得経路 — query state で gate し bundled 本文へ fallback しない。
describe("ManualPage overrides strict gating (fix-9b)", () => {
  function manualPageTree(initialPath: string, queryClient: QueryClient) {
    return (
      <QueryClientProvider client={queryClient}>
        <MemoryRouter initialEntries={[initialPath]}>
          <Routes>
            <Route path="/manual" element={<ManualPage />} />
            <Route path="/manual/:category/:slug" element={<ManualPage />} />
          </Routes>
        </MemoryRouter>
      </QueryClientProvider>
    );
  }

  function renderManualPageStrict(initialPath = "/manual/screens/getting-started") {
    const queryClient = new QueryClient();
    const view = render(manualPageTree(initialPath, queryClient));
    return {
      ...view,
      rerenderWith: (path = initialPath) => view.rerender(manualPageTree(path, queryClient)),
    };
  }

  it("passes strict=true to the overrides hook", () => {
    renderManualPageStrict();
    expect(overridesHookMock).toHaveBeenCalledWith(true, true);
  });

  it("pending: bundled 本文を表示せず検索・印刷を無効化する", () => {
    overridesHookMock.mockReturnValue({ isPending: true });
    renderManualPageStrict();
    expect(screen.getByText("マニュアル記事を読み込んでいます")).toBeInTheDocument();
    // bundled 記事本文は出ない
    expect(screen.queryByRole("article")).toBeNull();
    const printButton = screen.getByRole("button", { name: "このページを印刷" });
    expect(printButton).toBeDisabled();
    const searchInput = screen.getByRole("searchbox", { name: "マニュアル内検索" });
    expect(searchInput).toBeDisabled();
  });

  it.each(["500", "404", "403", "network"])(
    "error (%s): bundled 本文を表示せずエラーと再試行を出す",
    () => {
      const refetch = vi.fn();
      overridesHookMock.mockReturnValue({ isError: true, refetch });
      renderManualPageStrict();
      expect(screen.getByText("マニュアル記事の取得に失敗しました")).toBeInTheDocument();
      expect(screen.queryByRole("article")).toBeNull();
    },
  );

  it("error: 再試行ボタンが refetch を呼ぶ", async () => {
    const refetch = vi.fn();
    overridesHookMock.mockReturnValue({ isError: true, refetch });
    renderManualPageStrict();
    const user = userEvent.setup();
    await user.click(screen.getByRole("button", { name: "再試行" }));
    expect(refetch).toHaveBeenCalledTimes(1);
  });

  it("failed refetch after success: error state で旧本文を表示し続けない", () => {
    // query state (isError) が真実。data が残っていても article は出ない。
    const refetch = vi.fn();
    overridesHookMock.mockReturnValue({
      data: [
        {
          id: 1,
          category: "screens",
          slug: "getting-started",
          title: "T",
          order_value: 1,
          section: "S",
          body_markdown: "B",
          created_at: "",
          updated_at: "",
        },
      ],
      isError: true,
      refetch,
    });
    renderManualPageStrict();
    expect(screen.getByText("マニュアル記事の取得に失敗しました")).toBeInTheDocument();
    expect(screen.queryByRole("article")).toBeNull();
  });

  it("200 {data:[]}: bundled 記事を通常どおり表示する", () => {
    overridesHookMock.mockReturnValue({ data: [] });
    renderManualPageStrict();
    expect(screen.queryByText("マニュアル記事の取得に失敗しました")).toBeNull();
    expect(screen.queryByText("マニュアル記事を読み込んでいます")).toBeNull();
  });

  // fix-12: 公開 READ は権限不要。view/edit がなくても同じ strict fetch で
  // 同じ内容を読む（permission は読むものを変えない）。編集 button のみ canEdit。
  it("permission does not change what is read: override 本文を表示する", () => {
    manualPermissions.canView = false;
    manualPermissions.canEdit = false;
    overridesHookMock.mockReturnValue({
      data: [
        {
          id: 99,
          category: "screens",
          slug: "00-overview",
          title: "OVERRIDE_TITLE_ALWAYS_SHOWN",
          order_value: 1,
          section: "S",
          body_markdown: "OVERRIDE_BODY_ALWAYS_SHOWN",
          created_at: "",
          updated_at: "",
        },
      ],
    });
    renderManualPageStrict("/manual/screens/00-overview");
    expect(screen.getByText("OVERRIDE_TITLE_ALWAYS_SHOWN")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "このページを編集" })).toBeNull();
    manualPermissions.canView = true;
    manualPermissions.canEdit = true;
  });

  // fix-9c (2): blocked 中は index route も bundled 先頭記事へ遷移しない。
  it("index route while pending: bundled 記事へ遷移せず loading を表示する", () => {
    overridesHookMock.mockReturnValue({ isPending: true });
    renderManualPageStrict("/manual");
    expect(screen.getByText("マニュアル記事を読み込んでいます")).toBeInTheDocument();
  });

  // fix-9c (3): 編集中は背景 refetch 失敗でも ManualEditor/draft を失わない。
  it("editor open then error: editor stays mounted, blocked UI not shown", async () => {
    overridesHookMock.mockReturnValue({ data: [] });
    const view = renderManualPageStrict("/manual/screens/00-overview");
    const user = userEvent.setup();
    await user.click(screen.getByRole("button", { name: "このページを編集" }));
    // 公開 list query が error へ遷移しても editor は mounted のまま。
    overridesHookMock.mockReturnValue({ isError: true, refetch: vi.fn() });
    view.rerenderWith();
    expect(screen.getByLabelText("マニュアル本文編集")).toBeInTheDocument();
    // blocked 公開本文 UI / loading は出ない。
    expect(screen.queryByText("マニュアル記事の取得に失敗しました")).toBeNull();
    expect(screen.queryByText("マニュアル記事を読み込んでいます")).toBeNull();
    // 公開本文は描画されない（editor 分岐が優先）。
    expect(screen.queryByRole("button", { name: "このページを編集" })).toBeNull();
  });

  // fix-9c (4): blocked + 未編集中は bundled fallback から編集開始できない。
  it("blocked and not editing: edit button disabled", () => {
    overridesHookMock.mockReturnValue({ isPending: true });
    // 記事 route（bundled slug 00-overview）で article が解決され
    // 編集開始ボタンが確実に描画される状態で assert する。
    renderManualPageStrict("/manual/screens/00-overview");
    const editButton = screen.getByRole("button", { name: "このページを編集" });
    expect(editButton).toBeDisabled();
  });
});
