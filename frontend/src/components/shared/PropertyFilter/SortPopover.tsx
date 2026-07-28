import { C, ICON } from "@/lib/design-tokens";
import { memo, useState, useCallback } from "react";
import { ArrowUpDown, ArrowUp, ArrowDown, X, Plus } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { Command, CommandInput, CommandItem, CommandList, CommandEmpty } from "@/components/ui/command";
import type { SortProperty, ActiveSort } from "./types";

// ─── Sort Rule Row ───────────────────────────────────────────

interface SortRuleRowProps {
  sort: ActiveSort;
  sortProperties: SortProperty[];
  onChangeProperty: (oldKey: string, newKey: string) => void;
  onToggleDirection: (key: string) => void;
  onRemove: (key: string) => void;
}

const SortRuleRow = memo(function SortRuleRow({
  sort,
  sortProperties,
  onChangeProperty,
  onToggleDirection,
  onRemove,
}: SortRuleRowProps) {
  const [propOpen, setPropOpen] = useState(false);
  const property = sortProperties.find((p) => p.key === sort.key);
  const dirLabel = sort.direction === "asc" ? "昇順" : "降順";

  return (
    <div className="flex items-center gap-1.5 py-1 group">
      {/* Property selector */}
      <Popover open={propOpen} onOpenChange={setPropOpen}>
        <PopoverTrigger asChild>
          <button
            type="button"
            className={`flex items-center gap-1 px-2 py-1 text-base ${C.text} ${C.bgMutedBadge} ${C.hoverBgMutedBadge} rounded-xxs transition-colors max-w-[160px] truncate`}
          >
            {property?.icon ? (
              <property.icon className={`${ICON.action} shrink-0 opacity-50`} />
            ) : null}
            <span className="truncate">{property?.label ?? sort.key}</span>
          </button>
        </PopoverTrigger>
        <PopoverContent className="w-[200px] p-0" align="start">
          <Command>
            <CommandInput placeholder="プロパティを検索..." />
            <CommandList>
              <CommandEmpty>プロパティが見つかりません</CommandEmpty>
              {sortProperties.map((prop) => (
                <CommandItem
                  key={prop.key}
                  onSelect={() => {
                    onChangeProperty(sort.key, prop.key);
                    setPropOpen(false);
                  }}
                  className="text-base"
                >
                  {prop.icon ? (
                    <prop.icon className={`mr-2 ${ICON.xs} ${C.text50}`} />
                  ) : null}
                  {prop.label}
                </CommandItem>
              ))}
            </CommandList>
          </Command>
        </PopoverContent>
      </Popover>

      {/* Direction toggle */}
      <button
        type="button"
        onClick={() => onToggleDirection(sort.key)}
        className={`flex items-center gap-1 px-2 py-1 text-base ${C.text} ${C.bgMutedBadge} ${C.hoverBgMutedBadge} rounded-xxs transition-colors whitespace-nowrap`}
      >
        {sort.direction === "asc" ? (
          <ArrowUp className={`${ICON.page} shrink-0`} />
        ) : (
          <ArrowDown className={`${ICON.page} shrink-0`} />
        )}
        {dirLabel}
      </button>

      {/* Remove */}
      <button
        type="button"
        onClick={() => onRemove(sort.key)}
        className={`p-0.5 rounded-xxs ${C.text30} ${C.hoverText60} ${C.hoverBgLight} opacity-0 group-hover:opacity-100 transition-opacity ml-auto`}
        aria-label={`${property?.label ?? sort.key} ソートを削除`}
      >
        <X className={ICON.page} />
      </button>
    </div>
  );
});

// ─── Main Component ──────────────────────────────────────────

interface SortPopoverProps {
  sortProperties: SortProperty[];
  activeSorts: ActiveSort[];
  onSortChange: (sorts: ActiveSort[]) => void;
}

export const SortPopover = memo(function SortPopover({
  sortProperties,
  activeSorts,
  onSortChange,
}: SortPopoverProps) {
  const [open, setOpen] = useState(false);
  const [addingSort, setAddingSort] = useState(false);

  const handleToggleDirection = useCallback(
    (key: string) => {
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

  const handleChangeProperty = useCallback(
    (oldKey: string, newKey: string) => {
      if (oldKey === newKey) return;
      // Replace old sort with new property, keep direction
      onSortChange(
        activeSorts.map((s) =>
          s.key === oldKey ? { ...s, key: newKey } : s,
        ),
      );
    },
    [activeSorts, onSortChange],
  );

  const handleRemove = useCallback(
    (key: string) => {
      onSortChange(activeSorts.filter((s) => s.key !== key));
    },
    [activeSorts, onSortChange],
  );

  const handleAddSort = useCallback(
    (key: string) => {
      onSortChange([...activeSorts, { key, direction: "asc" }]);
      setAddingSort(false);
    },
    [activeSorts, onSortChange],
  );

  // Properties not already used in active sorts
  const availableProperties = sortProperties.filter(
    (p) => !activeSorts.some((s) => s.key === p.key),
  );

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          variant="ghost"
          size="sm"
          className={`h-9 w-9 p-0 ${C.text50} hover:${C.text80} ${C.hoverBgLight}`}
          aria-label="並べ替え"
        >
          <ArrowUpDown className={ICON.lg} />
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-[280px] p-0" align="end">
        <div className="px-3 py-2">
          <p className={`text-base ${C.text50} font-medium mb-1`}>並べ替え</p>

          {/* Active sort rules */}
          {activeSorts.length > 0 ? (
            <div className="space-y-0">
              {activeSorts.map((sort) => (
                <SortRuleRow
                  key={sort.key}
                  sort={sort}
                  sortProperties={sortProperties}
                  onChangeProperty={handleChangeProperty}
                  onToggleDirection={handleToggleDirection}
                  onRemove={handleRemove}
                />
              ))}
            </div>
          ) : (
            <p className={`text-base ${C.text30} py-2`}>
              並べ替えが設定されていません
            </p>
          )}

          {/* Add sort */}
          {availableProperties.length > 0 ? (
            addingSort ? (
              <div className={`mt-1 border-t ${C.borderLight} pt-1`}>
                <Command>
                  <CommandInput placeholder="プロパティを検索..." />
                  <CommandList>
                    <CommandEmpty>プロパティが見つかりません</CommandEmpty>
                    {availableProperties.map((prop) => (
                      <CommandItem
                        key={prop.key}
                        onSelect={() => handleAddSort(prop.key)}
                        className="text-base"
                      >
                        {prop.icon ? (
                          <prop.icon className={`mr-2 ${ICON.xs} ${C.text50}`} />
                        ) : null}
                        {prop.label}
                      </CommandItem>
                    ))}
                  </CommandList>
                </Command>
              </div>
            ) : (
              <button
                type="button"
                onClick={() => setAddingSort(true)}
                className={`flex items-center gap-1 mt-1 px-1 py-1 text-base ${C.text50} ${C.hoverText}/80 ${C.hoverBgLight} rounded-xxs transition-colors w-full`}
              >
                <Plus className={ICON.page} />
                並べ替えを追加
              </button>
            )
          ) : null}
        </div>
      </PopoverContent>
    </Popover>
  );
});
