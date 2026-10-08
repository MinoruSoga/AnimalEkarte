import { memo, type ReactNode } from "react";
import { ICON, C } from "@/lib/design-tokens";
import { ChevronRight } from "lucide-react";
import { formatCurrency } from "@/lib/format/number";

interface MasterSelectTriggerProps {
  id?: string;
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
  /** EMR-319: true でトリガーに aria-invalid + 赤枠を付与する */
  ariaInvalid?: boolean;
}

/** EMR-319: エラー表示用の追加クラス（STYLE.formInputError 相当） */
const INVALID_CLASSES = `${C.borderDanger} ring-2 ring-[#C0392B]/30`;

export const MasterSelectTrigger = memo(function MasterSelectTrigger({
  id,
  selectedItem,
  placeholder,
  icon,
  onClick,
  variant = "inline",
  ariaInvalid = false,
}: MasterSelectTriggerProps) {
  if (selectedItem) {
    // --- Selected state ---
    if (variant === "block") {
      return (
        <button
          id={id}
          type="button"
          onClick={onClick}
          aria-invalid={ariaInvalid || undefined}
          className={`w-full p-3 border ${ariaInvalid ? INVALID_CLASSES : C.borderPrimary} ${C.bgPage} rounded-md cursor-pointer ${C.hoverBgPage} transition-colors text-left`}
        >
          <div className="flex items-center justify-between">
            <div>
              <div className={`text-sm ${C.text}`}>{selectedItem.name}</div>
              {selectedItem.price != null ? (
                <div className={`text-xs ${C.text60} mt-0.5`}>
                  {formatCurrency(selectedItem.price)}
                </div>
              ) : null}
            </div>
            <div className={`flex items-center gap-1.5 text-xs ${C.text50}`}>
              <span>変更</span>
              <ChevronRight className={`${ICON.xs}`} />
            </div>
          </div>
        </button>
      );
    }

    // inline variant
    return (
      <button
        id={id}
        type="button"
        onClick={onClick}
        aria-invalid={ariaInvalid || undefined}
        className={`w-full h-11 px-3 border ${ariaInvalid ? INVALID_CLASSES : C.borderPrimary} ${C.bgPage} rounded-md cursor-pointer ${C.hoverBgPage} transition-colors flex items-center justify-between text-left`}
      >
        <div className="flex items-center gap-2 min-w-0">
          <span className={`text-sm ${C.text} truncate`}>{selectedItem.name}</span>
          {selectedItem.price != null ? (
            <span className={`text-xs ${C.text50} shrink-0`}>
              {formatCurrency(selectedItem.price)}
            </span>
          ) : null}
        </div>
        <div className={`flex items-center gap-1 text-xs ${C.text50} shrink-0 ml-2`}>
          <span>変更</span>
          <ChevronRight className={`${ICON.xs}`} />
        </div>
      </button>
    );
  }

  // --- Unselected state ---
  if (variant === "block") {
    return (
      <button
        id={id}
        type="button"
        onClick={onClick}
        aria-invalid={ariaInvalid || undefined}
        className={`w-full p-3 border-2 border-dashed ${ariaInvalid ? INVALID_CLASSES : C.borderMedium} rounded-md ${C.bgPage} ${C.hoverBgMedium} transition-colors text-center cursor-pointer`}
      >
        <div className="flex flex-col items-center">
          <span className={`${C.text30} mb-1`}>{icon}</span>
          <span className={`text-sm ${C.text50}`}>{placeholder}</span>
        </div>
      </button>
    );
  }

  // inline variant
  return (
    <button
      id={id}
      type="button"
      onClick={onClick}
      aria-invalid={ariaInvalid || undefined}
      className={`w-full h-11 px-3 border ${ariaInvalid ? INVALID_CLASSES : C.borderMedium} rounded-md bg-white ${C.hoverBgPageHalf} transition-colors text-left cursor-pointer flex items-center gap-2`}
    >
      <span className={C.text30}>{icon}</span>
      <span className={`text-sm ${C.text40}`}>{placeholder}</span>
    </button>
  );
});
