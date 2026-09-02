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

import { C } from "@/lib/design-tokens";
import { EmptyState } from "@/components/shared/DataStates";
import { paths } from "@/config/paths";
import { usePermission } from "@/hooks/use-permission";
import { ResourceManualEdit } from "@/types/generated/models";

import "../manual-print.css";

import { ManualContent } from "../components/ManualContent";
import { ManualEditor } from "../components/ManualEditor";
import { ManualPageChrome } from "../components/manual-page-chrome";
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

  const { data: fetchedOverrides } = useGetManualArticleOverrides(canViewOverrides);
  const overrides = canViewOverrides ? fetchedOverrides : undefined;

  const screenArticles = useMemo<ManualArticle[]>(
    () => applyOverrides(bundledScreenArticles, (overrides ?? []).filter((o) => o.category === "screens")),
    [overrides],
  );
  const workflowArticles = useMemo<ManualArticle[]>(
    () => applyOverrides(bundledWorkflowArticles, (overrides ?? []).filter((o) => o.category === "workflows")),
    [overrides],
  );

  const allMergedArticles = useMemo<ManualArticle[]>(
    () => [...screenArticles, ...workflowArticles],
    [screenArticles, workflowArticles],
  );

  const deferredQuery = useDeferredValue(query);
  const filtered = useManualSearch(deferredQuery, allMergedArticles);
  const isSearching = deferredQuery.trim().length > 0;

  const article = useMemo(() => {
    if (!params.category || !params.slug) return undefined;
    if (params.category !== "screens" && params.category !== "workflows") return undefined;
    const list = params.category === "screens" ? screenArticles : workflowArticles;
    return list.find((a) => a.slug === params.slug);
  }, [params.category, params.slug, screenArticles, workflowArticles]);

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

  useEffect(() => {
    document.documentElement.classList.add("manual-print-mode");
    document.body.classList.add("manual-print-mode");
    return () => {
      document.documentElement.classList.remove("manual-print-mode");
      document.body.classList.remove("manual-print-mode");
    };
  }, []);

  if (!params.category || !params.slug) {
    const fallback =
      viewMode === "screens" ? screenArticles[0] : workflowArticles[0];
    if (fallback) {
      return <Navigate to={paths.manual.article.getHref(fallback.category, fallback.slug)} replace />;
    }
  }

  return (
    <div className={`flex flex-1 overflow-hidden relative manual-root ${C.bgPage}`}>
      <ManualPageChrome
        drawerOpen={drawerOpen}
        onDrawerOpenChange={setDrawerOpen}
        onDrawerCloseAutoFocus={handleDrawerCloseAutoFocus}
        desktopSidebarRef={desktopSidebarRef}
        article={article}
        editMode={editMode}
        canEditManual={canEditManual === true}
        onStartEdit={() => setEditMode(true)}
        viewMode={viewMode}
        onChangeViewMode={setViewMode}
        query={query}
        onChangeQuery={setQuery}
        filteredArticles={filtered}
        isSearching={isSearching}
      />

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
