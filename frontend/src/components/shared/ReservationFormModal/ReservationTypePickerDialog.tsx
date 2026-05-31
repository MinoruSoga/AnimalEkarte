// React/Framework
import { useState, useCallback, useMemo, memo, Fragment } from "react";

// External
import { Search, X, Check } from "lucide-react";

// Internal
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { cn } from "@/lib/utils";
import { C, ICON } from "@/lib/design-tokens";

// --- Types ---
export interface ReservationTypePickerItem {
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

// --- Sub-Components ---

interface CategoryChipsProps {
  categories: string[];
  activeCategory: string | null;
  onSelectCategory: (category: string | null) => void;
}

const CategoryChips = memo(function CategoryChips({
  categories,
  activeCategory,
  onSelectCategory,
}: CategoryChipsProps) {
  return (
    <div
      className={cn(
        "flex items-center gap-1.5 overflow-x-auto rounded-md p-2 [&::-webkit-scrollbar]:hidden",
        C.bgPage30,
      )}
      style={{ scrollbarWidth: "none" }}
    >
      <div className="flex min-w-max gap-1.5">
        {activeCategory ? (
          <button
            type="button"
            onClick={() => onSelectCategory(null)}
            className={cn(
              "flex h-8 shrink-0 items-center gap-1 rounded-md border border-transparent bg-transparent px-3 text-sm",
              C.text60,
              C.hoverBgMedium,
            )}
          >
            <X className={ICON.action} />
            解除
          </button>
        ) : null}
        {categories.map((category) => {
          const isActive = activeCategory === category;
          return (
            <button
              key={category}
              type="button"
              onClick={() => onSelectCategory(isActive ? null : category)}
              className={cn(
                "h-8 shrink-0 rounded-md border px-2.5 text-sm whitespace-nowrap transition-colors",
                isActive
                  ? cn(C.bgAccent, "border-transparent text-white", C.bgAccentHover)
                  : cn("bg-white", C.text, C.borderMedium, C.hoverBgLight),
              )}
            >
              {category}
            </button>
          );
        })}
      </div>
    </div>
  );
});

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
    const term = searchTerm.trim().toLowerCase();
    return groups
      .filter((g) => activeCategory === null || g.label === activeCategory)
      .map((g) => ({
        label: g.label,
        items: term ? g.items.filter((it) => it.name.toLowerCase().includes(term)) : g.items,
      }))
      .filter((g) => g.items.length > 0);
  }, [groups, activeCategory, searchTerm]);

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
        <div className="relative">
          <Search className={cn("absolute top-1/2 left-2.5 -translate-y-1/2", ICON.action, C.text40)} />
          <Input
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            placeholder="予約区分を検索..."
            className={cn("h-11 bg-white pl-9 text-sm", C.borderMedium)}
          />
          {searchTerm ? (
            <button
              type="button"
              onClick={() => setSearchTerm("")}
              className={cn("absolute top-1/2 right-2.5 -translate-y-1/2", C.text40, C.hoverText)}
            >
              <X className={ICON.xs} />
            </button>
          ) : null}
        </div>

        {/* Category chips */}
        <CategoryChips
          categories={categories}
          activeCategory={activeCategory}
          onSelectCategory={setActiveCategory}
        />

        {/* Card list */}
        <div className="max-h-[440px] flex-1 space-y-1 overflow-y-auto overscroll-contain pr-1">
          {!hasResults ? (
            <div className={cn("py-12 text-center text-sm", C.text60)}>
              該当する予約区分が見つかりません。
            </div>
          ) : (
            filteredGroups.map((group) => (
              <Fragment key={group.label}>
                {activeCategory === null ? (
                  <div
                    className={cn(
                      "px-2 py-1.5 text-xs font-semibold tracking-wider uppercase",
                      C.text40,
                    )}
                  >
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
                            ? cn(C.borderAccent, C.bgAccent8)
                            : cn("bg-white", C.borderMedium, C.hoverBorderPrimary30, C.hoverBgPageHalf),
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
                          <Check className={cn(ICON.action, C.textAccentDark)} />
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
