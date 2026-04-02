// React/Framework
import { C, ICON } from "@/lib/design-tokens";
import * as React from "react";

// External
import { X } from "lucide-react";

// Internal
import { CommandDialog, CommandInput, CommandList, CommandEmpty, CommandGroup, CommandItem, CommandSeparator } from "@/components/ui/command";
import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";

// --- Types ---
export type TreatmentMasterItem = {
  id: string;
  code: string;
  name: string;
  unitPrice: number;
  category: string;
};

interface TreatmentSearchDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSelect: (item: TreatmentMasterItem) => void;
}

// --- Constants ---
const CATEGORY_ORDER = ["診察", "検査", "処置", "予防", "入院", "薬剤"];

const TREATMENT_MASTER: TreatmentMasterItem[] = [
  { id: "1001", code: "1001", name: "再診料(再診)", unitPrice: 800, category: "診察" },
  { id: "1002", code: "1002", name: "初診料", unitPrice: 1500, category: "診察" },
  { id: "1003", code: "1003", name: "時間外診察料", unitPrice: 2000, category: "診察" },
  { id: "2001", code: "2001", name: "混合ワクチン(5種)", unitPrice: 6000, category: "予防" },
  { id: "2002", code: "2002", name: "混合ワクチン(7種)", unitPrice: 8000, category: "予防" },
  { id: "2003", code: "2003", name: "狂犬病予防注射", unitPrice: 3000, category: "予防" },
  { id: "3001", code: "3001", name: "血液検査セットA", unitPrice: 5000, category: "検査" },
  { id: "3002", code: "3002", name: "血液検査セットB(生化学)", unitPrice: 7000, category: "検査" },
  { id: "3003", code: "3003", name: "X線検査(2枚)", unitPrice: 4000, category: "検査" },
  { id: "3004", code: "3004", name: "超音波検査(腹部)", unitPrice: 3000, category: "検査" },
  { id: "4001", code: "4001", name: "爪切り", unitPrice: 500, category: "処置" },
  { id: "4002", code: "4002", name: "耳掃除", unitPrice: 800, category: "処置" },
  { id: "4003", code: "4003", name: "肛門腺絞り", unitPrice: 500, category: "処置" },
  { id: "5001", code: "5001", name: "入院料(小型)", unitPrice: 3000, category: "入院" },
  { id: "5002", code: "5002", name: "入院料(中型)", unitPrice: 4000, category: "入院" },
  { id: "6001", code: "6001", name: "内服薬A(抗生剤)", unitPrice: 100, category: "薬剤" },
  { id: "6002", code: "6002", name: "内服薬B(消炎剤)", unitPrice: 80, category: "薬剤" },
];

// --- Sub-Components ---

interface CategoryFilterProps {
  categories: string[];
  activeCategory: string | null;
  onSelectCategory: (category: string | null) => void;
}

const CategoryFilter = React.memo(function CategoryFilter({
  categories,
  activeCategory,
  onSelectCategory,
}: CategoryFilterProps) {
  return (
    <div className={`flex gap-2 p-2 border-b overflow-x-auto items-center ${C.bgPage30} ${C.borderLight} [&::-webkit-scrollbar]:hidden [-ms-overflow-style:none] [scrollbar-width:none]`}>
      <div className="flex gap-1.5 min-w-max px-1">
        {activeCategory ? (
          <Badge
            variant="outline"
            className={`h-10 px-3 text-sm cursor-pointer ${C.hoverBgMedium} gap-1 ${C.text60} border-transparent bg-transparent`}
            onClick={() => onSelectCategory(null)}
            tabIndex={0}
            role="button"
            onKeyDown={(e) => {
              if (e.key === "Enter" || e.key === " ") {
                e.preventDefault();
                onSelectCategory(null);
              }
            }}
          >
            <X className={ICON.action} />
            解除
          </Badge>
        ) : null}
        {categories.map((category) => {
          const isSelected = activeCategory === category;
          return (
            <Badge
              key={category}
              variant={isSelected ? "default" : "outline"}
              className={cn(
                "h-10 px-2.5 text-sm cursor-pointer hover:opacity-80 transition-all",
                isSelected
                  ? `${C.bgPrimary} text-white ${C.hoverBgPrimaryDark} border-transparent`
                  : `bg-white ${C.text} ${C.hoverBgLight} ${C.borderMedium}`
              )}
              onClick={() => onSelectCategory(isSelected ? null : category)}
              tabIndex={0}
              role="button"
              onKeyDown={(e) => {
                if (e.key === "Enter" || e.key === " ") {
                  e.preventDefault();
                  onSelectCategory(isSelected ? null : category);
                }
              }}
            >
              {category}
            </Badge>
          );
        })}
      </div>
    </div>
  );
});

// --- Main Component ---

export function TreatmentSearchDialog({
  open,
  onOpenChange,
  onSelect,
}: TreatmentSearchDialogProps) {
  const [activeCategory, setActiveCategory] = React.useState<string | null>(null);

  // Reset category when dialog closes
  React.useEffect(() => {
    if (!open) {
      setActiveCategory(null);
    }
  }, [open]);

  // Memoize grouped items calculation
  const groupedItems = React.useMemo(() => {
    return TREATMENT_MASTER.reduce((acc, item) => {
      if (!acc[item.category]) acc[item.category] = [];
      acc[item.category].push(item);
      return acc;
    }, {} as Record<string, TreatmentMasterItem[]>);
  }, []);

  // Calculate all categories once
  const allCategories = React.useMemo(() => {
    return [
      ...CATEGORY_ORDER,
      ...Object.keys(groupedItems).filter((cat) => !CATEGORY_ORDER.includes(cat)),
    ];
  }, [groupedItems]);

  const handleSelect = React.useCallback((item: TreatmentMasterItem) => {
    onSelect(item);
    onOpenChange(false);
  }, [onSelect, onOpenChange]);

  return (
    <CommandDialog
      open={open}
      onOpenChange={onOpenChange}
      title="治療プラン検索"
      description="追加する治療プランを検索・選択してください"
    >
      <CommandInput placeholder="治療プランを検索... (例: 再診、ワクチン、3001)" />

      <CategoryFilter
        categories={allCategories}
        activeCategory={activeCategory}
        onSelectCategory={setActiveCategory}
      />

      <CommandList className="max-h-[500px]">
        <CommandEmpty className={`py-12 text-center text-sm ${C.text60}`}>該当する治療プランが見つかりません。</CommandEmpty>

        {allCategories.map((category) => {
          // Optimization: Skip rendering logic early if category doesn't match active filter
          if (activeCategory && activeCategory !== category) return null;

          const items = groupedItems[category];
          if (!items) return null;

          return (
            <React.Fragment key={category}>
              <CommandGroup heading={category}>
                {items.map((item) => (
                  <CommandItem
                    key={item.code}
                    value={`${item.name} ${item.code} ${item.category}`}
                    onSelect={() => handleSelect(item)}
                    className={`data-[selected=true]:${C.bgPage} cursor-pointer !py-1.5`}
                  >
                    <div className="flex flex-1 items-center justify-between">
                      <div className="flex flex-col gap-0.5">
                        <span className={`font-medium ${C.text} text-sm`}>
                          {item.name}
                        </span>
                        <div className="flex items-center gap-2">
                          <span className={`text-sm ${C.text40} font-mono ${C.bgPage30} px-1 rounded`}>
                            {item.code}
                          </span>
                        </div>
                      </div>
                      <span className={`font-mono font-bold ${C.text} text-sm`}>
                        ¥{item.unitPrice.toLocaleString()}
                      </span>
                    </div>
                  </CommandItem>
                ))}
              </CommandGroup>
              {/* Show separator only when not filtering by category (cleaner look) */}
              {!activeCategory ? <CommandSeparator className={C.bgLight} /> : null}
            </React.Fragment>
          );
        })}
      </CommandList>
    </CommandDialog>
  );
}
