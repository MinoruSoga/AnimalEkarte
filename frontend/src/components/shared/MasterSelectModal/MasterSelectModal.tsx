import { useState, useMemo, useCallback, useDeferredValue, memo } from "react";
import { Check } from "lucide-react";
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { EmptyState } from "@/components/shared/DataStates";
import { ClearableSearchInput } from "@/components/shared/ClearableSearchInput";
import { C, ICON } from "@/lib/design-tokens";
import { normalizeKana } from "@/lib/normalize-kana";
import { formatCurrency } from "@/lib/format/number";

export interface MasterSelectItem {
  id: string | number;
  name: string;
  price?: number;
}

interface MasterSelectModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  title: string;
  items: MasterSelectItem[];
  onSelect: (item: MasterSelectItem) => void;
  selectedValue?: string | number | null;
  searchPlaceholder?: string;

  /**
   * 選択状態の判定基準
   * - "id": id が一致するか
   * - "name": name が一致するか (デフォルト)
   */
  matchBy?: "id" | "name";
}

export const MasterSelectModal = memo(function MasterSelectModal({
  open,
  onOpenChange,
  title,
  items,
  onSelect,
  selectedValue,
  searchPlaceholder = "検索...",
  matchBy = "name",
}: MasterSelectModalProps) {
  const [searchTerm, setSearchTerm] = useState("");
  const deferredSearchTerm = useDeferredValue(searchTerm);

  const filtered = useMemo(() => {
    if (!deferredSearchTerm) return items;
    const normalizedTerm = normalizeKana(deferredSearchTerm).toLowerCase();
    return items.filter((item) => normalizeKana(item.name).toLowerCase().includes(normalizedTerm));
  }, [items, deferredSearchTerm]);

  const handleSelect = useCallback(
    (item: MasterSelectItem) => {
      onSelect(item);
      onOpenChange(false);
      setSearchTerm("");
    },
    [onSelect, onOpenChange]
  );

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[500px]">
        <DialogHeader>
          <DialogTitle className={`text-base font-bold ${C.text}`}>{title}</DialogTitle>
          <DialogDescription className="sr-only">
            マスタ項目を検索して選択します。
          </DialogDescription>
        </DialogHeader>

        {/* Search */}
        <ClearableSearchInput
          value={searchTerm}
          onChange={setSearchTerm}
          placeholder={searchPlaceholder}
        />

        {/* Item List */}
        <div className="space-y-2 max-h-[400px] overflow-y-auto pr-1">
          {filtered.length === 0 ? (
            <EmptyState message="該当する項目がありません" />
          ) : (
            filtered.map((item) => {
              const isSelected =
                matchBy === "id"
                  ? selectedValue === item.id
                  : selectedValue === item.name;

              return (
                <button
                  key={item.id}
                  type="button"
                  onClick={() => handleSelect(item)}
                  className={`
                    w-full p-3 border rounded-lg cursor-pointer transition-all flex items-center justify-between group text-left
                    ${
                      isSelected
                        ? `${C.bgPage} ${C.borderPrimary}`
                        : `bg-white ${C.borderMedium} ${C.hoverBorderPrimary30} ${C.hoverBgPageHalf}`
                    }
                  `}
                >
                  <div className="flex-1 min-w-0">
                    <div className={`text-sm ${C.text}`}>{item.name}</div>
                    {item.price != null ? (
                      <div className={`text-xs ${C.text70} mt-0.5`}>
                        {formatCurrency(item.price)}
                      </div>
                    ) : null}
                  </div>
                  <div className="flex items-center gap-2 shrink-0 ml-3">
                    {isSelected ? (
                      <div className={`size-5 rounded-full ${C.bgBrand} flex items-center justify-center`}>
                        <Check className={`${ICON.xs} ${C.textWhite}`} />
                      </div>
                    ) : (
                      <div className={`size-5 rounded-full border ${C.borderLight} ${C.groupHoverBorderPrimary} transition-colors`} />
                    )}
                  </div>
                </button>
              );
            })
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
});
