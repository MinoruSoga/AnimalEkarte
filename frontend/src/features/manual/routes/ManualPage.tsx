/**
 * ManualPage — Manual ルートレイアウト
 *
 * URL:
 *   /manual                              → 初期表示 (画面別の最初の項目)
 *   /manual/:category/:slug              → 個別記事
 *
 * 認証済みユーザー全員アクセス可。DB override閲覧・編集だけ権限制御する。
 * モバイル: サイドバーはドロワー化、ハンバーガーボタンで開閉。
 */

import {
  useState,
  useMemo,
  useEffect,
  useCallback,
  useDeferredValue,
  useRef,
} from "react";
import { Navigate, useParams } from "react-router";
import { Menu, Printer, Edit2 } from "lucide-react";

import { C } from "@/lib/design-tokens";
import { EmptyState } from "@/components/shared/DataStates";
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from "@/components/ui/sheet";
import { paths } from "@/config/paths";
import { usePermission } from "@/hooks/use-permission";
import { ResourceManualEdit } from "@/types/generated/models";

import "../manual-print.css";

import { ManualSidebar } from "../components/ManualSidebar";
import { ManualContent } from "../components/ManualContent";
import { ManualEditor } from "../components/ManualEditor";
import { useGetManualArticleOverrides } from "../api/get-manual-articles";
import {
  applyOverrides,
  screenArticles as bundledScreenArticles,
  workflowArticles as bundledWorkflowArticles,
  type ManualArticle,
  type ManualCategory,
} from "../lib/manual-index";
import { useManualSearch } from "../hooks/use-manual-search";

const DESKTOP_BREAKPOINT_QUERY = "(min-width: 768px)";

export function ManualPage() {
  const params = useParams<{ category?: string; slug?: string }>();
  const { canView: canViewOverrides, canEdit: canEditManual } =
    usePermission(ResourceManualEdit);

  const initialMode: ManualCategory =
    params.category === "workflows" ? "workflows" : "screens";
  const [viewMode, setViewMode] = useState<ManualCategory>(initialMode);
  const [query, setQuery] = useState("");
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [editMode, setEditMode] = useState(false);
  const desktopSidebarRef = useRef<HTMLDivElement>(null);

  const handleDrawerCloseAutoFocus = useCallback((event: Event) => {
    if (
      typeof window.matchMedia !== "function" ||
      !window.matchMedia(DESKTOP_BREAKPOINT_QUERY).matches
    ) {
      return;
    }

    event.preventDefault();
    desktopSidebarRef.current?.querySelector<HTMLElement>('[role="tab"]')?.focus();
  }, []);

  // DB に保存されたオーバーライド版を取得（取得失敗時は空配列、MD バンドル版が引き続き使われる）
  const { data: fetchedOverrides } = useGetManualArticleOverrides(canViewOverrides);
  const overrides = canViewOverrides ? fetchedOverrides : undefined;

  // バンドル MD と DB オーバーライドをマージした最終リスト
  const screenArticles = useMemo<ManualArticle[]>(
    () => applyOverrides(bundledScreenArticles, (overrides ?? []).filter((o) => o.category === "screens")),
    [overrides],
  );
  const workflowArticles = useMemo<ManualArticle[]>(
    () => applyOverrides(bundledWorkflowArticles, (overrides ?? []).filter((o) => o.category === "workflows")),
    [overrides],
  );

  // 全記事（検索インデックス用、override 適用済）
  const allMergedArticles = useMemo<ManualArticle[]>(
    () => [...screenArticles, ...workflowArticles],
    [screenArticles, workflowArticles],
  );

  const deferredQuery = useDeferredValue(query);
  const filtered = useManualSearch(deferredQuery, allMergedArticles);
  const isSearching = deferredQuery.trim().length > 0;

  // React Hooks ルール: 全ての hooks を early return より前に呼び出す
  const article = useMemo(() => {
    if (!params.category || !params.slug) return undefined;
    if (params.category !== "screens" && params.category !== "workflows") return undefined;
    const list = params.category === "screens" ? screenArticles : workflowArticles;
    return list.find((a) => a.slug === params.slug);
  }, [params.category, params.slug, screenArticles, workflowArticles]);

  // URL 遷移時はドロワーを閉じる + 編集モードも解除
  useEffect(() => {
    /* eslint-disable react-hooks/set-state-in-effect -- URL changes must synchronously close transient manual UI state. */
    setDrawerOpen(false);
    setEditMode(false);
    /* eslint-enable react-hooks/set-state-in-effect */
  }, [params.category, params.slug]);

  useEffect(() => {
    if (typeof window.matchMedia !== "function") return;

    const desktopBreakpoint = window.matchMedia(DESKTOP_BREAKPOINT_QUERY);
    const handleBreakpointChange = (event: MediaQueryListEvent) => {
      if (event.matches) setDrawerOpen(false);
    };

    desktopBreakpoint.addEventListener("change", handleBreakpointChange);
    return () => {
      desktopBreakpoint.removeEventListener("change", handleBreakpointChange);
    };
  }, []);

  // 取説ページの印刷時だけ、アプリ全体の固定高レイアウトを解除するためのスコープ。
  useEffect(() => {
    document.documentElement.classList.add("manual-print-mode");
    document.body.classList.add("manual-print-mode");
    return () => {
      document.documentElement.classList.remove("manual-print-mode");
      document.body.classList.remove("manual-print-mode");
    };
  }, []);

  // URL 未指定時は当該ビューモードの先頭記事へリダイレクト
  if (!params.category || !params.slug) {
    const fallback =
      viewMode === "screens" ? screenArticles[0] : workflowArticles[0];
    if (fallback) {
      return <Navigate to={paths.manual.article.getHref(fallback.category, fallback.slug)} replace />;
    }
  }

  return (
    <div className={`flex flex-1 overflow-hidden relative manual-root ${C.bgPage}`}>
      {/* モバイル用ハンバーガー (md 未満で表示)
       * アプリのメインサイドバー（collapsed 約50px）と被らないよう left-16 に配置 */}
      <Sheet open={drawerOpen} onOpenChange={setDrawerOpen}>
        <SheetTrigger asChild>
          <button
            type="button"
            aria-label="マニュアル目次を開く"
            className={`md:hidden fixed top-3 left-16 z-30 size-11 flex items-center justify-center rounded-xxs border ${C.borderDivider} ${C.bgWhite} ${C.hoverBgLight} no-print`}
          >
            <Menu className="size-5" />
          </button>
        </SheetTrigger>
        <SheetContent
          side="left"
          aria-modal="true"
          aria-describedby={undefined}
          onCloseAutoFocus={handleDrawerCloseAutoFocus}
          className="w-[280px] gap-0 p-0 md:hidden [&>button]:min-h-11 [&>button]:min-w-11"
        >
          <SheetHeader className="min-h-14 shrink-0 justify-center py-2 pr-16">
            <SheetTitle>マニュアル目次</SheetTitle>
          </SheetHeader>
          <div className="min-h-0 flex-1">
            <ManualSidebar
              viewMode={viewMode}
              onChangeViewMode={setViewMode}
              query={query}
              onChangeQuery={setQuery}
              filteredArticles={filtered}
              isSearching={isSearching}
            />
          </div>
        </SheetContent>
      </Sheet>

      {/* 編集ボタン (印刷ボタンの左隣) — 編集モード中は非表示 */}
      {article && !editMode && canEditManual ? (
        <button
          type="button"
          onClick={() => setEditMode(true)}
          aria-label="このページを編集"
          title="このページを編集（編集内容はダウンロードして管理者に共有）"
          className={`fixed top-3 right-16 z-30 size-11 flex items-center justify-center rounded-xxs border ${C.borderDivider} ${C.bgWhite} ${C.hoverBgLight} no-print`}
        >
          <Edit2 className="size-5" />
        </button>
      ) : null}

      {/* 印刷ボタン (常時右上、編集モード中は非表示) */}
      {article && !editMode ? (
        <button
          type="button"
          onClick={() => window.print()}
          aria-label="このページを印刷"
          title="このページを印刷"
          className={`fixed top-3 right-3 z-30 size-11 flex items-center justify-center rounded-xxs border ${C.borderDivider} ${C.bgWhite} ${C.hoverBgLight} no-print`}
        >
          <Printer className="size-5" />
        </button>
      ) : null}

      {/* デスクトップ用サイドバー (md 以上で常時表示) */}
      <div ref={desktopSidebarRef} className="hidden md:flex no-print">
        <ManualSidebar
          viewMode={viewMode}
          onChangeViewMode={setViewMode}
          query={query}
          onChangeQuery={setQuery}
          filteredArticles={filtered}
          isSearching={isSearching}
        />
      </div>

      {/* メインコンテンツ（編集モード時は ManualEditor、それ以外は ManualContent） */}
      {article ? (
        editMode && canEditManual ? (
          <ManualEditor article={article} onClose={() => setEditMode(false)} />
        ) : (
          <ManualContent article={article} />
        )
      ) : (
        <div className="flex-1 flex items-center justify-center p-8">
          <EmptyState message="マニュアル項目が見つかりません" />
        </div>
      )}
    </div>
  );
}
