// React/Framework
import { useState, useCallback, useMemo, useDeferredValue, memo, Fragment } from "react";

// External
import { Check } from "lucide-react";

// Internal
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { EmptyState } from "@/components/shared/DataStates";
import { ClearableSearchInput } from "@/components/shared/ClearableSearchInput";
import { CategoryChipsFilter } from "@/components/shared/CategoryChipsFilter";
import { cn } from "@/lib/utils";
import { C, ICON } from "@/lib/design-tokens";
import { normalizeKana } from "@/lib/normalize-kana";

// --- Types ---
interface ReservationTypePickerItem {
  id: string;
  name: string;
  color: string;
  durationMinutes: number;
}

export interface ReservationTypePickerGroup {
  label: string;
  items: ReservationTypePickerItem[];
}

interface ReservationTypePickerDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  groups: ReservationTypePickerGroup[];
  /** 現在選択中の予約区分 id(未選択は "") */
  selectedId: string;
  onSelect: (id: string) => void;
}

// --- Main Component ---

export const ReservationTypePickerDialog = memo(function ReservationTypePickerDialog({
  open,
  onOpenChange,
  groups,
  selectedId,
  onSelect,
}: ReservationTypePickerDialogProps) {
  const [activeCategory, setActiveCategory] = useState<string | null>(null);
  const [searchTerm, setSearchTerm] = useState("");
  const deferredSearchTerm = useDeferredValue(searchTerm);

  const handleOpenChange = useCallback(
    (next: boolean) => {
      if (!next) {
        setActiveCategory(null);
        setSearchTerm("");
      }
      onOpenChange(next);
    },
    [onOpenChange],
  );

  const handleSelect = useCallback(
    (id: string) => {
      onSelect(id);
      handleOpenChange(false);
    },
    [onSelect, handleOpenChange],
  );

  const categories = useMemo(() => groups.map((g) => g.label), [groups]);

  // 検索語 + カテゴリで絞り込み、グループ構造を保持
  const filteredGroups = useMemo(() => {
    const term = normalizeKana(deferredSearchTerm.trim()).toLowerCase();
    return groups
      .filter((g) => activeCategory === null || g.label === activeCategory)
      .map((g) => ({
        label: g.label,
        items: term
          ? g.items.filter((it) => normalizeKana(it.name).toLowerCase().includes(term))
          : g.items,
      }))
      .filter((g) => g.items.length > 0);
  }, [groups, activeCategory, deferredSearchTerm]);

  const hasResults = filteredGroups.length > 0;

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="flex max-h-[80vh] flex-col gap-3 sm:max-w-[560px]">
        <DialogHeader>
          <DialogTitle className={cn("text-base font-bold", C.text)}>予約区分を選択</DialogTitle>
          <DialogDescription className="sr-only">
            予約区分をカテゴリや検索で絞り込んで選択します。
          </DialogDescription>
        </DialogHeader>

        {/* Search */}
        <ClearableSearchInput
          value={searchTerm}
          onChange={setSearchTerm}
          placeholder="予約区分を検索..."
        />

        {/* Category chips */}
        <CategoryChipsFilter
          categories={categories}
          activeCategory={activeCategory}
          onSelectCategory={setActiveCategory}
        />

        {/* Card list */}
        <div className="max-h-[440px] flex-1 space-y-1 overflow-y-auto overscroll-contain pr-1">
          {!hasResults ? (
            <EmptyState message="該当する予約区分が見つかりません。" />
          ) : (
            filteredGroups.map((group) => (
              <Fragment key={group.label}>
                {activeCategory === null ? (
                  <div className={cn("px-2 py-1.5 text-2xs font-semibold uppercase", C.text40)}>
                    {group.label}
                  </div>
                ) : null}
                <div className="space-y-1.5">
                  {group.items.map((item) => {
                    const isSelected = item.id === selectedId;
                    return (
                      <button
                        key={item.id}
                        type="button"
                        onClick={() => handleSelect(item.id)}
                        data-testid={`res-type-card-${item.id}`}
                        className={cn(
                          "group flex w-full items-center gap-3 rounded-lg border p-3 text-left transition-all",
                          isSelected
                            ? cn(C.borderBrand, C.bgBrand8)
                            : cn(
                                "bg-white",
                                C.borderMedium,
                                C.hoverBorderPrimary30,
                                C.hoverBgPageHalf,
                              ),
                        )}
                      >
                        <span
                          className="size-3 shrink-0 rounded-full"
                          style={{ backgroundColor: item.color }}
                        />
                        <div className="min-w-0 flex-1">
                          <div className={cn("text-sm font-medium", C.text)}>{item.name}</div>
                          <div className={cn("mt-0.5 text-xs", C.text60)}>
                            約{item.durationMinutes}分
                          </div>
                        </div>
                        {isSelected ? (
                          <Check className={cn(ICON.action, C.textBrand)} />
                        ) : (
                          <div
                            className={cn(
                              "size-5 shrink-0 rounded-full border transition-colors group-hover:border-current",
                              C.borderLight,
                            )}
                          />
                        )}
                      </button>
                    );
                  })}
                </div>
              </Fragment>
            ))
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
});
