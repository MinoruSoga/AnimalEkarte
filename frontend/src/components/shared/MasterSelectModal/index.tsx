import { useState, useCallback, ReactNode } from "react";
import { Input } from "@/components/ui/input";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from "@/components/ui/dialog";
import { Search, X, Check, ChevronRight } from "lucide-react";
import { MasterLink } from "@/components/shared/MasterLink";
import type { MasterCategory, MasterItem } from "@/types";

// --- MasterSelectModal ---

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
  masterCategory?: MasterCategory;
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
    return (
      item.name.toLowerCase().includes(term) ||
      (item.code && item.code.toLowerCase().includes(term))
    );
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
            className="pl-9 h-10 text-sm bg-white border-[rgba(55,53,47,0.16)]"
          />
          {searchTerm && (
            <button
              onClick={() => setSearchTerm("")}
              className="absolute right-2.5 top-1/2 -translate-y-1/2 text-[#37352F]/40 hover:text-[#37352F]/70"
            >
              <X className="size-3.5" />
            </button>
          )}
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
                  {item.code && (
                    <div className="text-xs text-[#37352F]/40 mt-0.5">
                      {item.code}
                    </div>
                  )}
                </div>
                <div className="flex items-center gap-2 shrink-0 ml-3">
                  {item.price != null && (
                    <div className="text-sm text-[#37352F]/70">
                      ¥{item.price.toLocaleString()}
                    </div>
                  )}
                  {isSelected && (
                    <div className="size-5 rounded-full bg-[#37352F] flex items-center justify-center">
                      <Check className="size-3 text-white" />
                    </div>
                  )}
                </div>
              </div>
            );
          })}
          {filtered.length === 0 && (
            <div className="text-center py-8 text-sm text-[#37352F]/40">
              {searchTerm ? emptySearchMessage : emptyMessage}
            </div>
          )}
        </div>

        {/* Master Settings Link */}
        {masterCategory && (
          <div className="pt-2 border-t border-[rgba(55,53,47,0.09)] flex justify-end">
            <MasterLink category={masterCategory} label="マスタ設定を編集" />
          </div>
        )}
      </DialogContent>
    </Dialog>
  );
}

// --- MasterSelectTrigger ---

interface MasterSelectTriggerProps {
  /** Currently selected item info (null/undefined if nothing selected) */
  selectedItem?: { name: string; price?: number } | null;
  /** Placeholder text when nothing is selected */
  placeholder: string;
  /** Icon shown in the unselected placeholder */
  icon: ReactNode;
  /** Called when the trigger is clicked */
  onClick: () => void;
  /**
   * "inline" — single-line h-10 trigger (used by ExaminationForm, VaccinationForm)
   * "block"  — taller dashed-border card-style trigger (used by TrimmingForm)
   */
  variant?: "inline" | "block";
}

export function MasterSelectTrigger({
  selectedItem,
  placeholder,
  icon,
  onClick,
  variant = "inline",
}: MasterSelectTriggerProps) {
  if (selectedItem) {
    // --- Selected state ---
    if (variant === "block") {
      return (
        <div
          onClick={onClick}
          className="p-3 border border-[#37352F] bg-[#F7F6F3] rounded-md cursor-pointer hover:bg-[#F7F6F3]/80 transition-colors"
        >
          <div className="flex items-center justify-between">
            <div>
              <div className="text-sm text-[#37352F]">
                {selectedItem.name}
              </div>
              {selectedItem.price != null && (
                <div className="text-xs text-[#37352F]/60 mt-0.5">
                  ¥{selectedItem.price.toLocaleString()}
                </div>
              )}
            </div>
            <div className="flex items-center gap-1.5 text-xs text-[#37352F]/50">
              <span>変更</span>
              <ChevronRight className="size-3.5" />
            </div>
          </div>
        </div>
      );
    }

    // inline variant
    return (
      <div
        onClick={onClick}
        className="h-10 px-3 border border-[#37352F] bg-[#F7F6F3] rounded-md cursor-pointer hover:bg-[#F7F6F3]/80 transition-colors flex items-center justify-between"
      >
        <div className="flex items-center gap-2 min-w-0">
          <span className="text-sm text-[#37352F] truncate">
            {selectedItem.name}
          </span>
          {selectedItem.price != null && (
            <span className="text-xs text-[#37352F]/50 shrink-0">
              ¥{selectedItem.price.toLocaleString()}
            </span>
          )}
        </div>
        <div className="flex items-center gap-1 text-xs text-[#37352F]/50 shrink-0 ml-2">
          <span>変更</span>
          <ChevronRight className="size-3.5" />
        </div>
      </div>
    );
  }

  // --- Unselected state ---
  if (variant === "block") {
    return (
      <button
        type="button"
        onClick={onClick}
        className="w-full p-3 border-2 border-dashed border-[rgba(55,53,47,0.16)] rounded-md bg-[#F7F6F3] hover:bg-[rgba(55,53,47,0.08)] transition-colors text-center cursor-pointer"
      >
        <div className="flex flex-col items-center">
          <span className="text-[#37352F]/30 mb-1">{icon}</span>
          <span className="text-sm text-[#37352F]/50">{placeholder}</span>
        </div>
      </button>
    );
  }

  // inline variant
  return (
    <button
      type="button"
      onClick={onClick}
      className="w-full h-10 px-3 border border-[rgba(55,53,47,0.16)] rounded-md bg-white hover:bg-[#F7F6F3]/50 transition-colors text-left cursor-pointer flex items-center gap-2"
    >
      <span className="text-[#37352F]/30">{icon}</span>
      <span className="text-sm text-[#37352F]/40">{placeholder}</span>
    </button>
  );
}
