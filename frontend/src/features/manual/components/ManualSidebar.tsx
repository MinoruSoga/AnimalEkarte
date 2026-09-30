/**
 * ManualSidebar — Manual 内左サイドバー
 *
 * - ビューモード切替 (画面別 / 業務フロー別)
 * - 検索ボックス (Fuse.js)
 * - セクション別目次
 */

import { Link, useLocation } from "react-router";
import { Search } from "lucide-react";
import { memo, useId, useMemo, useRef, type KeyboardEvent } from "react";

import { C, PALETTE, STYLE } from "@/lib/design-tokens";
import { paths } from "@/config/paths";

import {
  screenArticles,
  workflowArticles,
  groupBySection,
  type ManualArticle,
  type ManualCategory,
} from "@/lib/manual-index";

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

/** ビューモード切替タブの定義。tablist のキーボード操作順もこの順に従う。 */
const VIEW_MODES: ReadonlyArray<{ mode: ManualCategory; label: string }> = [
  { mode: "screens", label: "画面別" },
  { mode: "workflows", label: "業務フロー" },
];

const ManualNavigation = memo(function ManualNavigation({
  groups,
  filteredCount,
  isSearching,
  viewMode,
}: ManualNavigationProps) {
  const location = useLocation();

  return (
    <nav aria-label="マニュアル目次" className="flex-1 overflow-y-auto relative px-2 py-2">
      {isSearching && filteredCount === 0 ? (
        <p className={`px-2 py-3 text-sm ${C.text50}`}>該当する項目が見つかりません</p>
      ) : null}

      {groups.map((group) => (
        <div key={group.section} className="mb-3">
          <p className={`px-2 mb-1 text-2xs font-semibold ${C.text60} uppercase`}>
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
                      <span className={`ml-1 text-2xs ${C.text60}`}>
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

  // WAI-ARIA Tabs パターン: roving tabindex + 矢印キー/Home/End 移動。
  // フォーカスと同時にタブを有効化する（automatic activation）。
  const tabBaseId = useId();
  const tabPanelId = `${tabBaseId}-panel`;
  const tabRefs = useRef<Array<HTMLButtonElement | null>>([]);

  const handleTabKeyDown = (event: KeyboardEvent<HTMLButtonElement>) => {
    const focusedIndex = tabRefs.current.findIndex((node) => node === event.currentTarget);
    if (focusedIndex < 0) return;

    let nextIndex: number | null = null;
    if (event.key === "ArrowRight") {
      nextIndex = (focusedIndex + 1) % VIEW_MODES.length;
    } else if (event.key === "ArrowLeft") {
      nextIndex = (focusedIndex + VIEW_MODES.length - 1) % VIEW_MODES.length;
    } else if (event.key === "Home") {
      nextIndex = 0;
    } else if (event.key === "End") {
      nextIndex = VIEW_MODES.length - 1;
    }
    if (nextIndex === null) return;

    event.preventDefault();
    onChangeViewMode(VIEW_MODES[nextIndex].mode);
    tabRefs.current[nextIndex]?.focus();
  };

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
          className={`grid grid-cols-2 gap-1 p-0.5 rounded-xxs ${C.bgPrimary5}`}
        >
          {VIEW_MODES.map((item, index) => (
            <button
              key={item.mode}
              ref={(node) => {
                tabRefs.current[index] = node;
              }}
              type="button"
              role="tab"
              id={`${tabBaseId}-tab-${item.mode}`}
              aria-selected={viewMode === item.mode}
              aria-controls={tabPanelId}
              tabIndex={viewMode === item.mode ? 0 : -1}
              onClick={() => onChangeViewMode(item.mode)}
              onKeyDown={handleTabKeyDown}
              className={`min-h-11 min-w-11 px-2 py-1.5 text-sm rounded-xxs transition-colors ${
                viewMode === item.mode
                  ? `${C.bgActionPrimarySolid} ${C.textOnActionPrimary} ${C.hoverBgActionPrimarySolid} ${C.hoverTextOnActionPrimary} font-medium`
                  : `${C.text65} ${C.hoverBgLight}`
              }`}
            >
              {item.label}
            </button>
          ))}
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

      {/* タブパネル: tablist と aria-controls/labelledby で相互参照 */}
      <div
        role="tabpanel"
        id={tabPanelId}
        aria-labelledby={`${tabBaseId}-tab-${viewMode}`}
        className="flex-1 min-h-0 flex flex-col"
      >
        <ManualNavigation
          groups={groups}
          filteredCount={filteredArticles.length}
          isSearching={isSearching}
          viewMode={viewMode}
        />
      </div>
    </aside>
  );
}
