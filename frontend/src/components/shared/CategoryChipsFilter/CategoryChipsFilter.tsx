import { memo } from "react";
import { X } from "lucide-react";

import { C, ICON } from "@/lib/design-tokens";
import { cn } from "@/lib/utils";

interface CategoryChipsFilterProps {
  categories: string[];
  activeCategory: string | null;
  onSelectCategory: (category: string | null) => void;
  clearLabel?: string;
  className?: string;
  /**
   * チップの角丸。呼び出し元の元実装差を吸収するためのオプション。
   * "md" (既定, 8px) = ReservationTypePickerDialog の元 CategoryChips 相当。
   * "sm" (5px) = TreatmentSearchDialog の元 CategoryFilter (shadcn Badge ベース) 相当。
   */
  chipRounded?: "sm" | "md";
  /**
   * ホバー時の挙動。呼び出し元の元実装差を吸収するためのオプション。
   * "button" (既定) = 背景色トランジションのみ (ReservationTypePickerDialog 相当)。
   * "badge" = shadcn Badge が持っていた hover:opacity-80 のフェード (TreatmentSearchDialog 相当)。
   */
  hoverEffect?: "button" | "badge";
}

/**
 * カテゴリチップ絞り込み行。TreatmentSearchDialog の CategoryFilter (shadcn Badge ベース) と
 * ReservationTypePickerDialog の CategoryChips (素の button ベース) を統合。
 * 両者は角丸とホバー挙動が異なっていたため、chipRounded/hoverEffect で呼び出し元ごとの
 * 見た目を保存する。
 */
export const CategoryChipsFilter = memo(function CategoryChipsFilter({
  categories,
  activeCategory,
  onSelectCategory,
  clearLabel = "解除",
  className,
  chipRounded = "md",
  hoverEffect = "button",
}: CategoryChipsFilterProps) {
  const roundedClass = chipRounded === "sm" ? "rounded-sm" : "rounded-md";
  const hoverTransitionClass = hoverEffect === "badge" ? "hover:opacity-80 transition-all" : "transition-colors";

  return (
    <div
      className={cn(
        "flex items-center gap-1.5 overflow-x-auto rounded-md p-2 [&::-webkit-scrollbar]:hidden",
        C.bgPage30,
        className,
      )}
      style={{ scrollbarWidth: "none" }}
    >
      <div className="flex min-w-max gap-1.5">
        {activeCategory ? (
          <button
            type="button"
            onClick={() => onSelectCategory(null)}
            className={cn(
              "flex h-8 shrink-0 items-center gap-1 border border-transparent bg-transparent px-3 text-sm",
              roundedClass,
              C.text60,
              C.hoverBgMedium,
            )}
          >
            <X aria-hidden="true" className={ICON.action} />
            {clearLabel}
          </button>
        ) : null}
        {categories.map((category) => {
          const isActive = activeCategory === category;
          return (
            <button
              key={category}
              type="button"
              onClick={() => onSelectCategory(isActive ? null : category)}
              aria-pressed={isActive}
              className={cn(
                "h-8 shrink-0 border px-2.5 text-sm whitespace-nowrap",
                roundedClass,
                hoverTransitionClass,
                isActive
                  ? cn(
                      C.bgBrand,
                      "border-transparent",
                      C.textOnBrand,
                      C.hoverBgBrand,
                      C.hoverTextOnBrand,
                    )
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
