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

import { useState, useMemo, useEffect, useCallback, useDeferredValue, useRef } from "react";
import { Navigate, useParams } from "react-router";

import { C } from "@/lib/design-tokens";
import { EmptyState } from "@/components/shared/DataStates";
import { paths } from "@/config/paths";
import { usePermission } from "@/hooks/use-permission";
import { ResourceManualEdit } from "@/types/generated/models";

import "../manual-print.css";

import { ManualContent } from "../components/ManualContent";
import { ManualEditor } from "../components/ManualEditor";
import { ManualPageChrome } from "../components/ManualPageChrome";
import { useGetManualArticleOverrides } from "@/hooks/use-manual-article-overrides";
import { useManualSearch } from "@/hooks/use-manual-search";
import {
  applyOverrides,
  screenArticles as bundledScreenArticles,
  workflowArticles as bundledWorkflowArticles,
  type ManualArticle,
  type ManualCategory,
} from "@/lib/manual-index";

const DESKTOP_BREAKPOINT_QUERY = "(min-width: 768px)";

export function ManualPage() {
  const params = useParams<{ category?: string; slug?: string }>();
  const { canEdit: canEditManual } = usePermission(ResourceManualEdit);

  const initialMode: ManualCategory = params.category === "workflows" ? "workflows" : "screens";
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

  // strict 経路: 取得失敗を空配列として bundled 本文へ戻さない。
  // data の有無ではなく query state で判定する（refetch 失敗後も最後の
  // 成功 data が残るため）。公開 READ は認証済みスタッフ全員（編集権限と分離、
  // FR-internal-manual-publication）のため enabled は常に true。
  const overridesQuery = useGetManualArticleOverrides(true, true);
  const overridesBlocked = overridesQuery.isPending || overridesQuery.isError;
  const overrides = overridesQuery.data;

  const screenArticles = useMemo<ManualArticle[]>(
    () =>
      applyOverrides(
        bundledScreenArticles,
        (overrides ?? []).filter((o) => o.category === "screens"),
      ),
    [overrides],
  );
  const workflowArticles = useMemo<ManualArticle[]>(
    () =>
      applyOverrides(
        bundledWorkflowArticles,
        (overrides ?? []).filter((o) => o.category === "workflows"),
      ),
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

  // blocked 中は index も bundled 記事へ遷移させず取得完了を待つ。
  if (!overridesBlocked && (!params.category || !params.slug)) {
    const fallback = viewMode === "screens" ? screenArticles[0] : workflowArticles[0];
    if (fallback) {
      return (
        <Navigate to={paths.manual.article.getHref(fallback.category, fallback.slug)} replace />
      );
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
        filteredArticles={overridesBlocked ? [] : filtered}
        isSearching={isSearching}
        disabled={overridesBlocked}
      />

      {/* 編集中 draft は blocked でも失わない（editor は自前の edit-view 読取り
          と local state を持つ）。block は公開本文・検索・印刷だけ。 */}
      {editMode && canEditManual && article ? (
        <ManualEditor
          article={article}
          onClose={() => setEditMode(false)}
          canEdit={canEditManual === true}
        />
      ) : overridesBlocked ? (
        <div className="flex-1 flex flex-col items-center justify-center gap-4 p-8">
          {overridesQuery.isError ? (
            <>
              <EmptyState
                message="マニュアル記事の取得に失敗しました"
                description="通信状態を確認して再試行してください"
              />
              <button
                type="button"
                onClick={() => overridesQuery.refetch()}
                className={`rounded-xxs border px-4 py-2 ${C.borderDivider} ${C.bgWhite} ${C.hoverBgLight}`}
              >
                再試行
              </button>
            </>
          ) : (
            <EmptyState message="マニュアル記事を読み込んでいます" />
          )}
        </div>
      ) : article ? (
        <ManualContent article={article} />
      ) : (
        <div className="flex-1 flex items-center justify-center p-8">
          <EmptyState message="マニュアル項目が見つかりません" />
        </div>
      )}
    </div>
  );
}
