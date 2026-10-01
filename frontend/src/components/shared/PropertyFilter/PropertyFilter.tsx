import { memo, useState, useCallback, useMemo } from "react";
import { Search, ListFilter, X } from "lucide-react";
import { Button } from "@/components/ui/button";
import { C, STYLE, ICON } from "@/lib/design-tokens";
import { FilterRuleRow } from "./FilterRuleRow";
import { FilterAddPopover } from "./FilterAddPopover";
import { SortPopover } from "./SortPopover";
import { SortPill } from "./SortPill";
import type { PropertyFilterProps, ActiveFilter } from "./types";

export const PropertyFilter = memo(function PropertyFilter({
  properties,
  activeFilters,
  onFilterChange,
  filterLogic = "and",
  onFilterLogicChange,
  searchTerm,
  onSearchChange,
  searchPlaceholder = "検索...",
  count,
  sortProperties,
  activeSorts,
  onSortChange,
}: PropertyFilterProps) {
  const [searchOpen, setSearchOpen] = useState(false);

  // EMR-247: 検索語は確定操作（Enter / 検索ボタン）まで内部 draft に保持し、
  // キー入力のたびに onSearchChange を呼ばない（リクエスト連打の防止）。
  const confirmedTerm = searchTerm ?? "";
  const [searchDraft, setSearchDraft] = useState(confirmedTerm);
  // 外部から searchTerm が変わった場合（URL 戻る/進む等）は draft を再同期する。
  // rerender-derived-state-no-effect: OwnersList の searchParamsKey 比較と同型。
  const [prevConfirmedTerm, setPrevConfirmedTerm] = useState(confirmedTerm);
  if (confirmedTerm !== prevConfirmedTerm) {
    setPrevConfirmedTerm(confirmedTerm);
    setSearchDraft(confirmedTerm);
  }

  const handleRemoveFilter = useCallback(
    (key: string) => {
      onFilterChange(activeFilters.filter((f) => f.key !== key));
    },
    [activeFilters, onFilterChange],
  );

  const handleAddFilter = useCallback(
    (filter: ActiveFilter) => {
      // 同じキーのフィルタがあれば置換
      const next = activeFilters.filter((f) => f.key !== filter.key);
      next.push(filter);
      onFilterChange(next);
    },
    [activeFilters, onFilterChange],
  );

  const handleUpdateFilter = useCallback(
    (updated: ActiveFilter) => {
      onFilterChange(activeFilters.map((f) => (f.key === updated.key ? updated : f)));
    },
    [activeFilters, onFilterChange],
  );

  // Property lookup map for rule rows
  const propertyMap = useMemo(() => new Map(properties.map((p) => [p.key, p])), [properties]);

  // ── Sort handlers ──

  const handleSortToggleDirection = useCallback(
    (key: string) => {
      if (!onSortChange || !activeSorts) return;
      onSortChange(
        activeSorts.map((s) =>
          s.key === key ? { ...s, direction: s.direction === "asc" ? "desc" : "asc" } : s,
        ),
      );
    },
    [activeSorts, onSortChange],
  );

  const handleSortChangeProperty = useCallback(
    (oldKey: string, newKey: string) => {
      if (!onSortChange || !activeSorts || oldKey === newKey) return;
      onSortChange(activeSorts.map((s) => (s.key === oldKey ? { ...s, key: newKey } : s)));
    },
    [activeSorts, onSortChange],
  );

  const handleSortRemove = useCallback(
    (key: string) => {
      if (!onSortChange || !activeSorts) return;
      onSortChange(activeSorts.filter((s) => s.key !== key));
    },
    [activeSorts, onSortChange],
  );

  // BUG-091: 検索バーを閉じる際に searchTerm もクリアして全件表示に戻す（即時発火を維持）
  const handleSearchToggle = useCallback(() => {
    if (searchOpen) {
      setSearchDraft("");
      onSearchChange?.("");
    }
    setSearchOpen((prev) => !prev);
  }, [searchOpen, onSearchChange]);

  // EMR-247: 確定操作（Enter / 検索ボタン）。確定済み値と一致する確定は no-op。
  const handleSearchCommit = useCallback(() => {
    if (!onSearchChange) return;
    if (searchDraft !== confirmedTerm) {
      onSearchChange(searchDraft);
    }
  }, [onSearchChange, searchDraft, confirmedTerm]);

  const handleSearchClear = useCallback(() => {
    setSearchDraft("");
    onSearchChange?.("");
  }, [onSearchChange]);

  const hasSortProps = sortProperties && activeSorts && onSortChange;
  const hasActiveSorts = activeSorts && activeSorts.length > 0;

  return (
    <div className="flex flex-col gap-1">
      {/* Toolbar row */}
      <div className="flex flex-wrap items-center gap-2">
        {/* 左側: 件数 + ソートピル + フィルタピル + フィルタ追加 */}
        {count !== undefined ? (
          <span className={STYLE.searchCount}>{count.toLocaleString()} 件</span>
        ) : null}

        {/* Sort pills (orange) */}
        {hasActiveSorts && sortProperties
          ? activeSorts.map((sort) => (
              <SortPill
                key={sort.key}
                sort={sort}
                sortProperties={sortProperties}
                onToggleDirection={handleSortToggleDirection}
                onChangeProperty={handleSortChangeProperty}
                onRemove={handleSortRemove}
              />
            ))
          : null}

        {/* フィルタ追加ボタン */}
        <FilterAddPopover
          properties={properties}
          activeFilters={activeFilters}
          onAdd={handleAddFilter}
        />

        {/* Spacer */}
        <div className="flex-1" />

        {/* 右側: ツールバーアイコンボタン */}
        <div className="flex items-center gap-1">
          {/* Filter icon (visual indicator - same as add popover trigger) */}
          {activeFilters.length > 0 ? (
            <span className={`h-9 w-9 flex items-center justify-center ${C.textBrand}`}>
              <ListFilter className={ICON.lg} />
            </span>
          ) : null}

          {/* Sort popover icon */}
          {hasSortProps ? (
            <SortPopover
              sortProperties={sortProperties}
              activeSorts={activeSorts}
              onSortChange={onSortChange}
            />
          ) : null}

          {/* Search toggle icon */}
          {onSearchChange ? (
            <Button
              variant="ghost"
              size="sm"
              className={`h-11 w-11 p-0 ${C.hoverBgMedium} ${
                searchOpen ? C.textBrand : `${C.text50} ${C.hoverText80}`
              }`}
              onClick={handleSearchToggle}
              aria-label="検索"
            >
              <Search className={ICON.lg} />
            </Button>
          ) : null}
        </div>
      </div>

      {/* Search bar (toggle) */}
      {onSearchChange && searchOpen ? (
        <div className="relative max-w-md">
          <Search className={STYLE.searchIcon} />
          <input
            type="text"
            value={searchDraft}
            onChange={(e) => setSearchDraft(e.target.value)}
            onKeyDown={(e) => {
              // IME 変換中の確定 Enter / キーリピートでは確定しない
              // （TreatmentQuantityCell.tsx と同じ isComposing + keyCode 229 判定）。
              const isComposing = e.nativeEvent.isComposing || e.nativeEvent.keyCode === 229;
              if (e.key === "Enter" && !e.repeat && !isComposing) {
                handleSearchCommit();
              }
            }}
            placeholder={searchPlaceholder}
            aria-label={searchPlaceholder}
            className={`${STYLE.searchInput} pr-22`}
            autoFocus
          />
          {/* BUG-091: 検索クリアボタン - draft が空でないときのみ表示（即時クリアを維持） */}
          {searchDraft ? (
            <button
              type="button"
              onClick={handleSearchClear}
              className={`absolute right-11 top-1/2 flex min-h-11 min-w-11 -translate-y-1/2 items-center justify-center rounded-sm ${C.text40} ${C.hoverText80} ${C.hoverBgMedium} transition-colors`}
              aria-label="検索をクリア"
            >
              <X className={ICON.smXs} />
            </button>
          ) : null}
          {/* EMR-247: 確定ボタン - draft を onSearchChange へ反映する（Enter と同等） */}
          <button
            type="button"
            onClick={handleSearchCommit}
            className={`absolute right-0 top-1/2 flex min-h-11 min-w-11 -translate-y-1/2 items-center justify-center rounded-sm ${C.text40} ${C.hoverText80} ${C.hoverBgMedium} transition-colors`}
            aria-label="検索を実行"
          >
            <Search className={ICON.smXs} />
          </button>
        </div>
      ) : null}

      {/* Filter rule rows */}
      {activeFilters.length > 0 ? (
        <div className="flex flex-col gap-0 pl-1 pt-1">
          {activeFilters.map((filter, i) => (
            <FilterRuleRow
              key={filter.key}
              filter={filter}
              property={propertyMap.get(filter.key)}
              isFirst={i === 0}
              logic={filterLogic}
              onLogicChange={activeFilters.length >= 2 ? onFilterLogicChange : undefined}
              onUpdate={handleUpdateFilter}
              onRemove={() => handleRemoveFilter(filter.key)}
            />
          ))}
        </div>
      ) : null}
    </div>
  );
});
