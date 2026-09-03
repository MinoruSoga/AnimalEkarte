/**
 * ManualSidebar — Manual 内左サイドバー
 *
 * - ビューモード切替 (画面別 / 業務フロー別)
 * - 検索ボックス (Fuse.js)
 * - セクション別目次
 */

import { Link, useLocation } from "react-router";
import { Search } from "lucide-react";
import { memo, useMemo } from "react";

import { C, PALETTE, STYLE } from "@/lib/design-tokens";
import { paths } from "@/config/paths";

import {
  screenArticles,
  workflowArticles,
  groupBySection,
  type ManualArticle,
  type ManualCategory,
} from "../lib/manual-index";

interface ManualSidebarProps {
  viewMode: ManualCategory;
  onChangeViewMode: (mode: ManualCategory) => void;
  query: string;
  onChangeQuery: (q: string) => void;
  filteredArticles: ManualArticle[];
  isSearching: boolean;
}

interface ManualNavigationProps {
  groups: ReturnType<typeof groupBySection>;
  filteredCount: number;
  isSearching: boolean;
  viewMode: ManualCategory;
}

const ManualNavigation = memo(function ManualNavigation({
  groups,
  filteredCount,
  isSearching,
  viewMode,
}: ManualNavigationProps) {
  const location = useLocation();

  return (
    <nav aria-label="マニュアル目次" className="flex-1 overflow-y-auto px-2 py-2">
      {isSearching && filteredCount === 0 ? (
        <p className={`px-2 py-3 text-sm ${C.text50}`}>該当する項目が見つかりません</p>
      ) : null}

      {groups.map((group) => (
        <div key={group.section} className="mb-3">
          <p className={`px-2 mb-1 text-2xs font-semibold ${C.text40} uppercase`}>
            {group.section}
          </p>
          <ul className="space-y-0.5">
            {group.items.map((article) => {
              const href = paths.manual.article.getHref(article.category, article.slug);
              const isActive = location.pathname === href;
              return (
                <li key={`${article.category}/${article.slug}`}>
                  <Link
                    to={href}
                    className={`flex min-h-11 items-center px-2 py-1.5 rounded-xxs text-sm transition-colors ${
                      isActive
                        ? `${STYLE.sidebarItemActive} font-medium`
                        : `${C.text65} ${C.hoverBgLight} ${C.hoverText}`
                    }`}
                  >
                    {article.title}
                    {isSearching && article.category !== viewMode ? (
                      <span className={`ml-1 text-2xs ${C.text40}`}>
                        ({article.category === "screens" ? "画面" : "フロー"})
                      </span>
                    ) : null}
                  </Link>
                </li>
              );
            })}
          </ul>
        </div>
      ))}
    </nav>
  );
});

export function ManualSidebar({
  viewMode,
  onChangeViewMode,
  query,
  onChangeQuery,
  filteredArticles,
  isSearching,
}: ManualSidebarProps) {
  // 検索中は全カテゴリ横断結果、非検索時はビューモードに応じて分類表示
  const baseList = viewMode === "screens" ? screenArticles : workflowArticles;
  const displayList = isSearching ? filteredArticles : baseList;
  const groups = useMemo(() => groupBySection(displayList), [displayList]);

  return (
    <aside
      className={`w-[280px] md:w-[260px] h-full shrink-0 border-r ${C.borderDivider} flex flex-col overflow-hidden`}
      style={{ backgroundColor: PALETTE.bgMain }}
    >
      {/* View Mode Toggle */}
      <div className={`p-3 border-b ${C.borderDivider} space-y-2`}>
        <div
          role="tablist"
          aria-label="マニュアル表示モード"
          className="grid grid-cols-2 gap-1 p-0.5 rounded-xxs bg-black/5"
        >
          <button
            type="button"
            role="tab"
            aria-selected={viewMode === "screens"}
            onClick={() => onChangeViewMode("screens")}
            className={`min-h-11 min-w-11 px-2 py-1.5 text-sm rounded-xxs transition-colors ${
              viewMode === "screens"
                ? `${C.bgBrand} ${C.textOnBrand} font-medium`
                : `${C.text65} ${C.hoverBgLight}`
            }`}
          >
            画面別
          </button>
          <button
            type="button"
            role="tab"
            aria-selected={viewMode === "workflows"}
            onClick={() => onChangeViewMode("workflows")}
            className={`min-h-11 min-w-11 px-2 py-1.5 text-sm rounded-xxs transition-colors ${
              viewMode === "workflows"
                ? `${C.bgBrand} ${C.textOnBrand} font-medium`
                : `${C.text65} ${C.hoverBgLight}`
            }`}
          >
            業務フロー
          </button>
        </div>

        {/* Search */}
        <div className="relative">
          <Search className={`absolute left-2 top-1/2 -translate-y-1/2 size-4 ${C.text40}`} />
          <input
            type="search"
            placeholder="マニュアル内を検索"
            value={query}
            onChange={(e) => onChangeQuery(e.target.value)}
            aria-label="マニュアル内検索"
            className={`min-h-11 w-full pl-8 pr-2 py-1.5 text-sm rounded-xxs border ${C.borderDivider} bg-white outline-none focus:ring-1 ${C.focusRingBrand40} ${C.text}`}
          />
        </div>
      </div>

      <ManualNavigation
        groups={groups}
        filteredCount={filteredArticles.length}
        isSearching={isSearching}
        viewMode={viewMode}
      />
    </aside>
  );
}
