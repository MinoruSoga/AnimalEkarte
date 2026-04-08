import { useState, useMemo, useCallback } from "react";
import { Search, X, Check } from "lucide-react";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { C, ICON } from "@/lib/design-tokens";

interface MasterItem {
  id: string | number;
  name: string;
  price?: number;
}

interface MasterSelectModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  title: string;
  items: MasterItem[];
  onSelect: (item: MasterItem) => void;
  selectedValue?: string | number | null;
  searchPlaceholder?: string;

  /**
   * 選択状態の判定基準
   * - "id": id が一致するか
   * - "name": name が一致するか (デフォルト)
   */
  matchBy?: "id" | "name";
}

export function MasterSelectModal({
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

  const filtered = useMemo(() => {
    if (!searchTerm) return items;
    const lower = searchTerm.toLowerCase();
    return items.filter((item) => item.name.toLowerCase().includes(lower));
  }, [items, searchTerm]);

  const handleSelect = useCallback(
    (item: MasterItem) => {
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
        </DialogHeader>

        {/* Search */}
        <div className="relative">
          <Search className={`absolute left-2.5 top-1/2 -translate-y-1/2 ${ICON.action} ${C.text40}`} />
          <Input
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            placeholder={searchPlaceholder}
            className={`pl-9 h-11 text-sm bg-white ${C.borderMedium}`}
          />
          {searchTerm ? (
            <button
              onClick={() => setSearchTerm("")}
              className={`absolute right-2.5 top-1/2 -translate-y-1/2 ${C.text40} ${C.hoverText}`}
            >
              <X className={`${ICON.xs}`} />
            </button>
          ) : null}
        </div>

        {/* Item List */}
        <div className="space-y-2 max-h-[400px] overflow-y-auto pr-1">
          {filtered.length === 0 ? (
            <div className={`py-12 text-center text-sm ${C.text40}`}>
              該当する項目がありません
            </div>
          ) : (
            filtered.map((item) => {
              const isSelected =
                matchBy === "id"
                  ? selectedValue === item.id
                  : selectedValue === item.name;

              return (
                <div
                  key={item.id}
                  onClick={() => handleSelect(item)}
                  className={`
                    p-3 border rounded-lg cursor-pointer transition-all flex items-center justify-between group
                    ${
                      isSelected
                        ? `${C.bgPage} ${C.borderPrimary} shadow-sm`
                        : `bg-white ${C.borderMedium} ${C.hoverBorderPrimary30} ${C.hoverBgPageHalf}`
                    }
                  `}
                >
                  <div className="flex-1 min-w-0">
                    <div className={`text-sm ${C.text}`}>{item.name}</div>
                    {item.price != null ? (
                      <div className={`text-xs ${C.text70} mt-0.5`}>
                        ¥{item.price.toLocaleString()}
                      </div>
                    ) : null}
                  </div>
                  <div className="flex items-center gap-2 shrink-0 ml-3">
                    {isSelected ? (
                      <div className={`size-5 rounded-full ${C.bgPrimary} flex items-center justify-center`}>
                        <Check className={`${ICON.xs} ${C.textWhite}`} />
                      </div>
                    ) : (
                      <div className={`size-5 rounded-full border ${C.borderLight} group-hover:${C.borderPrimary} transition-colors`} />
                    )}
                  </div>
                </div>
              );
            })
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
}
