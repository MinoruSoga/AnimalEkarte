import { useState, useCallback, useMemo } from "react";
import { Search, ListFilter } from "lucide-react";
import { Button } from "@/components/ui/button";
import { STYLE } from "@/lib/design-tokens";
import { FilterRuleRow } from "./FilterRuleRow";
import { FilterAddPopover } from "./FilterAddPopover";
import { SortPopover } from "./SortPopover";
import { SortPill } from "./SortPill";
import type { NotionFilterProps, ActiveFilter } from "./types";

export function NotionFilter({
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
}: NotionFilterProps) {
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

  const handleSearchToggle = useCallback(() => {
    setSearchOpen((prev) => !prev);
  }, []);

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
            <span className="h-9 w-9 flex items-center justify-center text-[#2383E2]">
              <ListFilter className="h-6 w-6" />
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
              className={`h-9 w-9 p-0 hover:bg-[#F1F1EF] ${
                searchOpen
                  ? "text-[#2383E2]"
                  : "text-[#37352F]/50 hover:text-[#37352F]/80"
              }`}
              onClick={handleSearchToggle}
              aria-label="検索"
            >
              <Search className="h-6 w-6" />
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
            className={STYLE.searchInput}
            autoFocus
          />
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
}
