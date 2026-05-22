/**
 * ManualPage — Manual ルートレイアウト
 *
 * URL:
 *   /manual                              → 初期表示 (画面別の最初の項目)
 *   /manual/:category/:slug              → 個別記事
 *
 * 認証済みユーザー全員アクセス可。permission gating は無し。
 * モバイル: サイドバーはドロワー化、ハンバーガーボタンで開閉。
 */

import { useState, useMemo, useEffect } from "react";
import { Navigate, useParams } from "react-router";
import { Menu, X, Printer, Edit2 } from "lucide-react";

import { C } from "@/lib/design-tokens";

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

export function ManualPage() {
  const params = useParams<{ category?: string; slug?: string }>();

  const initialMode: ManualCategory =
    params.category === "workflows" ? "workflows" : "screens";
  const [viewMode, setViewMode] = useState<ManualCategory>(initialMode);
  const [query, setQuery] = useState("");
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [editMode, setEditMode] = useState(false);

  // DB に保存されたオーバーライド版を取得（取得失敗時は空配列、MD バンドル版が引き続き使われる）
  const { data: overrides } = useGetManualArticleOverrides();

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

  const filtered = useManualSearch(query, allMergedArticles);
  const isSearching = query.trim().length > 0;

  // React Hooks ルール: 全ての hooks を early return より前に呼び出す
  const article = useMemo(() => {
    if (!params.category || !params.slug) return undefined;
    if (params.category !== "screens" && params.category !== "workflows") return undefined;
    const list = params.category === "screens" ? screenArticles : workflowArticles;
    return list.find((a) => a.slug === params.slug);
  }, [params.category, params.slug, screenArticles, workflowArticles]);

  // URL 遷移時はドロワーを閉じる + 編集モードも解除
  useEffect(() => {
    setDrawerOpen(false);
    setEditMode(false);
  }, [params.category, params.slug]);

  // URL 未指定時は当該ビューモードの先頭記事へリダイレクト
  if (!params.category || !params.slug) {
    const fallback =
      viewMode === "screens" ? screenArticles[0] : workflowArticles[0];
    if (fallback) {
      return <Navigate to={`/manual/${fallback.category}/${fallback.slug}`} replace />;
    }
  }

  return (
    <div className="flex flex-1 overflow-hidden relative manual-root" style={{ backgroundColor: "#FFFFFF" }}>
      {/* モバイル用ハンバーガー (md 未満で表示)
       * アプリのメインサイドバー（collapsed 約50px）と被らないよう left-16 に配置 */}
      <button
        type="button"
        onClick={() => setDrawerOpen(true)}
        aria-label="マニュアル目次を開く"
        className={`md:hidden fixed top-3 left-16 z-30 size-10 flex items-center justify-center rounded-[3px] border ${C.borderDivider} bg-white shadow-sm ${C.hoverBgLight} no-print`}
      >
        <Menu className="size-5" />
      </button>

      {/* 編集ボタン (印刷ボタンの左隣) — 編集モード中は非表示 */}
      {article && !editMode ? (
        <button
          type="button"
          onClick={() => setEditMode(true)}
          aria-label="このページを編集"
          title="このページを編集（編集内容はダウンロードして管理者に共有）"
          className={`fixed top-3 right-16 z-30 size-10 flex items-center justify-center rounded-[3px] border ${C.borderDivider} bg-white shadow-sm ${C.hoverBgLight} no-print`}
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
          className={`fixed top-3 right-3 z-30 size-10 flex items-center justify-center rounded-[3px] border ${C.borderDivider} bg-white shadow-sm ${C.hoverBgLight} no-print`}
        >
          <Printer className="size-5" />
        </button>
      ) : null}

      {/* デスクトップ用サイドバー (md 以上で常時表示) */}
      <div className="hidden md:flex no-print">
        <ManualSidebar
          viewMode={viewMode}
          onChangeViewMode={setViewMode}
          query={query}
          onChangeQuery={setQuery}
          filteredArticles={filtered}
          isSearching={isSearching}
        />
      </div>

      {/* モバイル用ドロワー */}
      {drawerOpen ? (
        <>
          {/* オーバーレイ */}
          <div
            role="button"
            tabIndex={0}
            aria-label="ドロワーを閉じる"
            onClick={() => setDrawerOpen(false)}
            onKeyDown={(e) => {
              if (e.key === "Escape" || e.key === "Enter" || e.key === " ") {
                setDrawerOpen(false);
              }
            }}
            className="md:hidden fixed inset-0 bg-black/30 z-40 no-print"
          />
          {/* ドロワー本体 */}
          <div className="md:hidden fixed inset-y-0 left-0 z-50 flex no-print">
            <ManualSidebar
              viewMode={viewMode}
              onChangeViewMode={setViewMode}
              query={query}
              onChangeQuery={setQuery}
              filteredArticles={filtered}
              isSearching={isSearching}
            />
            <button
              type="button"
              onClick={() => setDrawerOpen(false)}
              aria-label="ドロワーを閉じる"
              className={`size-10 m-2 flex items-center justify-center rounded-[3px] bg-white border ${C.borderDivider} shadow-sm`}
            >
              <X className="size-5" />
            </button>
          </div>
        </>
      ) : null}

      {/* メインコンテンツ（編集モード時は ManualEditor、それ以外は ManualContent） */}
      {article ? (
        editMode ? (
          <ManualEditor article={article} onClose={() => setEditMode(false)} />
        ) : (
          <ManualContent article={article} />
        )
      ) : (
        <div className="flex-1 flex items-center justify-center p-8">
          <p className={C.text50}>マニュアル項目が見つかりません</p>
        </div>
      )}
    </div>
  );
}
