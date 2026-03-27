import { ICON } from "@/lib/design-tokens";
import { ReactNode } from "react";
import { ChevronRight } from "lucide-react";

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
              {selectedItem.price != null ? (
                <div className="text-xs text-[#37352F]/60 mt-0.5">
                  ¥{selectedItem.price.toLocaleString()}
                </div>
              ) : null}
            </div>
            <div className="flex items-center gap-1.5 text-xs text-[#37352F]/50">
              <span>変更</span>
              <ChevronRight className={`${ICON.xs}.5`} />
            </div>
          </div>
        </div>
      );
    }

    // inline variant
    return (
      <div
        onClick={onClick}
        className="h-11 px-3 border border-[#37352F] bg-[#F7F6F3] rounded-md cursor-pointer hover:bg-[#F7F6F3]/80 transition-colors flex items-center justify-between"
      >
        <div className="flex items-center gap-2 min-w-0">
          <span className="text-sm text-[#37352F] truncate">
            {selectedItem.name}
          </span>
          {selectedItem.price != null ? (
            <span className="text-xs text-[#37352F]/50 shrink-0">
              ¥{selectedItem.price.toLocaleString()}
            </span>
          ) : null}
        </div>
        <div className="flex items-center gap-1 text-xs text-[#37352F]/50 shrink-0 ml-2">
          <span>変更</span>
          <ChevronRight className={`${ICON.xs}.5`} />
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
      className="w-full h-11 px-3 border border-[rgba(55,53,47,0.16)] rounded-md bg-white hover:bg-[#F7F6F3]/50 transition-colors text-left cursor-pointer flex items-center gap-2"
    >
      <span className="text-[#37352F]/30">{icon}</span>
      <span className="text-sm text-[#37352F]/40">{placeholder}</span>
    </button>
  );
}
