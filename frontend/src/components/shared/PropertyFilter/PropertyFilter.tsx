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
      onFilterChange(
        activeFilters.map((f) => (f.key === updated.key ? updated : f)),
      );
    },
    [activeFilters, onFilterChange],
  );

  // Property lookup map for rule rows
  const propertyMap = useMemo(
    () => new Map(properties.map((p) => [p.key, p])),
    [properties],
  );

  // ── Sort handlers ──

  const handleSortToggleDirection = useCallback(
    (key: string) => {
      if (!onSortChange || !activeSorts) return;
      onSortChange(
        activeSorts.map((s) =>
          s.key === key
            ? { ...s, direction: s.direction === "asc" ? "desc" : "asc" }
            : s,
        ),
      );
    },
    [activeSorts, onSortChange],
  );

  const handleSortChangeProperty = useCallback(
    (oldKey: string, newKey: string) => {
      if (!onSortChange || !activeSorts || oldKey === newKey) return;
      onSortChange(
        activeSorts.map((s) =>
          s.key === oldKey ? { ...s, key: newKey } : s,
        ),
      );
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

  // BUG-091: 検索バーを閉じる際に searchTerm もクリアして全件表示に戻す
  const handleSearchToggle = useCallback(() => {
    setSearchOpen((prev) => {
      if (prev && onSearchChange) {
        onSearchChange("");
      }
      return !prev;
    });
  }, [onSearchChange]);

  const hasSortProps = sortProperties && activeSorts && onSortChange;
  const hasActiveSorts = activeSorts && activeSorts.length > 0;

  return (
    <div className="flex flex-col gap-1">
      {/* Toolbar row */}
      <div className="flex flex-wrap items-center gap-2">
        {/* 左側: 件数 + ソートピル + フィルタピル + フィルタ追加 */}
        {count !== undefined ? (
          <span className={STYLE.searchCount}>{count} 件</span>
        ) : null}

        {/* Sort pills (orange) */}
        {hasActiveSorts && sortProperties ? (
          activeSorts.map((sort) => (
            <SortPill
              key={sort.key}
              sort={sort}
              sortProperties={sortProperties}
              onToggleDirection={handleSortToggleDirection}
              onChangeProperty={handleSortChangeProperty}
              onRemove={handleSortRemove}
            />
          ))
        ) : null}

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
              className={`h-9 w-9 p-0 ${C.hoverBgMedium} ${
                searchOpen
                  ? C.textBrand
                  : `${C.text50} hover:${C.text80}`
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
            value={searchTerm ?? ""}
            onChange={(e) => onSearchChange(e.target.value)}
            placeholder={searchPlaceholder}
            aria-label={searchPlaceholder}
            className={`${STYLE.searchInput} pr-11`}
            autoFocus
          />
          {/* BUG-091: 検索クリアボタン - searchTerm が空でないときのみ表示 */}
          {searchTerm ? (
            <button
              type="button"
              onClick={() => onSearchChange("")}
              className={`absolute right-0 top-1/2 flex min-h-11 min-w-11 -translate-y-1/2 items-center justify-center rounded-sm ${C.text40} hover:${C.text80} ${C.hoverBgMedium} transition-colors`}
              aria-label="検索をクリア"
            >
              <X className={ICON.smXs} />
            </button>
          ) : null}
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
              onLogicChange={
                activeFilters.length >= 2 ? onFilterLogicChange : undefined
              }
              onUpdate={handleUpdateFilter}
              onRemove={() => handleRemoveFilter(filter.key)}
            />
          ))}
        </div>
      ) : null}
    </div>
  );
});
