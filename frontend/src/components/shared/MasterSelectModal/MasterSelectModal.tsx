import { useState, useCallback } from "react";
import { Input } from "@/components/ui/input";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from "@/components/ui/dialog";
import { Search, X, Check } from "lucide-react";
import { MasterLink } from "@/components/shared/MasterLink";
import type { MasterLinkCategory } from "@/components/shared/MasterLink";
import type { MasterItem } from "@/types";

interface MasterSelectModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  title: string;
  description: string;
  searchPlaceholder?: string;
  items: MasterItem[];
  selectedValue?: string;
  onSelect: (item: MasterItem) => void;
  /** "name" or "id" — determines how `selectedValue` is matched against items. Default: "name" */
  matchBy?: "name" | "id";
  emptySearchMessage?: string;
  emptyMessage?: string;
  /** If provided, shows a link to the master settings page inside the dialog */
  masterCategory?: MasterLinkCategory;
}

export function MasterSelectModal({
  open,
  onOpenChange,
  title,
  description,
  searchPlaceholder = "名前・コードで検索...",
  items,
  selectedValue,
  onSelect,
  matchBy = "name",
  emptySearchMessage = "該当する項目が見つかりません",
  emptyMessage = "項目が登録されていません",
  masterCategory,
}: MasterSelectModalProps) {
  const [searchTerm, setSearchTerm] = useState("");

  const handleOpenChange = useCallback(
    (nextOpen: boolean) => {
      onOpenChange(nextOpen);
      if (!nextOpen) setSearchTerm("");
    },
    [onOpenChange]
  );

  const handleSelect = useCallback(
    (item: MasterItem) => {
      onSelect(item);
      handleOpenChange(false);
    },
    [onSelect, handleOpenChange]
  );

  const filtered = items.filter((item) => {
    if (!searchTerm) return true;
    const term = searchTerm.toLowerCase();
    return item.name.toLowerCase().includes(term);
  });

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle className="text-[#37352F]">{title}</DialogTitle>
          <DialogDescription>{description}</DialogDescription>
        </DialogHeader>

        {/* Search */}
        <div className="relative">
          <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 size-4 text-[#37352F]/40" />
          <Input
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            placeholder={searchPlaceholder}
            className="pl-9 h-11 text-sm bg-white border-[rgba(55,53,47,0.16)]"
          />
          {searchTerm ? (
            <button
              onClick={() => setSearchTerm("")}
              className="absolute right-2.5 top-1/2 -translate-y-1/2 text-[#37352F]/40 hover:text-[#37352F]/70"
            >
              <X className="size-3.5" />
            </button>
          ) : null}
        </div>

        {/* Item List */}
        <div className="space-y-2 max-h-[400px] overflow-y-auto">
          {filtered.map((item) => {
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
                      ? "bg-[#F7F6F3] border-[#37352F] shadow-sm"
                      : "bg-white border-[rgba(55,53,47,0.16)] hover:border-[#37352F]/30 hover:bg-[#F7F6F3]/50"
                  }
                `}
              >
                <div className="flex-1 min-w-0">
                  <div className="text-sm text-[#37352F]">{item.name}</div>
                </div>
                <div className="flex items-center gap-2 shrink-0 ml-3">
                  {item.price != null ? (
                    <div className="text-sm text-[#37352F]/70">
                      ¥{item.price.toLocaleString()}
                    </div>
                  ) : null}
                  {isSelected ? (
                    <div className="size-5 rounded-full bg-[#37352F] flex items-center justify-center">
                      <Check className="size-3 text-white" />
                    </div>
                  ) : null}
                </div>
              </div>
            );
          })}
          {filtered.length === 0 ? (
            <div className="text-center py-8 text-sm text-[#37352F]/40">
              {searchTerm ? emptySearchMessage : emptyMessage}
            </div>
          ) : null}
        </div>

        {/* Master Settings Link */}
        {masterCategory ? (
          <div className="pt-2 border-t border-[rgba(55,53,47,0.09)] flex justify-end">
            <MasterLink category={masterCategory} label="マスタ設定を編集" />
          </div>
        ) : null}
      </DialogContent>
    </Dialog>
  );
}
