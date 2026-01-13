import * as React from "react";
import {
  CommandDialog,
  CommandInput,
  CommandList,
  CommandEmpty,
  CommandGroup,
  CommandItem,
  CommandSeparator,
} from "../../../components/ui/command";
import { Badge } from "../../../components/ui/badge";
import { cn } from "../../../components/ui/utils";
import { X } from "lucide-react";

// --- Types ---
export type TreatmentMasterItem = {
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
  { code: "1001", name: "再診料(再診)", unitPrice: 800, category: "診察" },
  { code: "1002", name: "初診料", unitPrice: 1500, category: "診察" },
  { code: "1003", name: "時間外診察料", unitPrice: 2000, category: "診察" },
  { code: "2001", name: "混合ワクチン(5種)", unitPrice: 6000, category: "予防" },
  { code: "2002", name: "混合ワクチン(7種)", unitPrice: 8000, category: "予防" },
  { code: "2003", name: "狂犬病予防注射", unitPrice: 3000, category: "予防" },
  { code: "3001", name: "血液検査セットA", unitPrice: 5000, category: "検査" },
  { code: "3002", name: "血液検査セットB(生化学)", unitPrice: 7000, category: "検査" },
  { code: "3003", name: "X線検査(2枚)", unitPrice: 4000, category: "検査" },
  { code: "3004", name: "超音波検査(腹部)", unitPrice: 3000, category: "検査" },
  { code: "4001", name: "爪切り", unitPrice: 500, category: "処置" },
  { code: "4002", name: "耳掃除", unitPrice: 800, category: "処置" },
  { code: "4003", name: "肛門腺絞り", unitPrice: 500, category: "処置" },
  { code: "5001", name: "入院料(小型)", unitPrice: 3000, category: "入院" },
  { code: "5002", name: "入院料(中型)", unitPrice: 4000, category: "入院" },
  { code: "6001", name: "内服薬A(抗生剤)", unitPrice: 100, category: "薬剤" },
  { code: "6002", name: "内服薬B(消炎剤)", unitPrice: 80, category: "薬剤" },
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
    <div className="flex gap-2 p-2 border-b overflow-x-auto items-center bg-gray-50/50 [&::-webkit-scrollbar]:hidden [-ms-overflow-style:none] [scrollbar-width:none]">
      <div className="flex gap-1.5 min-w-max px-1">
        {activeCategory && (
          <Badge
            variant="outline"
            className="h-10 px-3 text-sm cursor-pointer hover:bg-gray-200 gap-1 text-muted-foreground border-transparent bg-transparent"
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
            <X className="h-3.5 w-3.5" />
            解除
          </Badge>
        )}
        {categories.map((category) => {
          const isSelected = activeCategory === category;
          return (
            <Badge
              key={category}
              variant={isSelected ? "default" : "outline"}
              className={cn(
                "h-10 px-2.5 text-sm cursor-pointer hover:opacity-80 transition-all",
                isSelected
                  ? "bg-[#37352F] text-white hover:bg-[#37352F]/90 border-transparent"
                  : "bg-white text-[#37352F] hover:bg-gray-100 border-gray-200"
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
        <CommandEmpty>該当する治療プランが見つかりません。</CommandEmpty>

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
                    className="data-[selected=true]:bg-[#F7F6F3] cursor-pointer !py-1.5"
                  >
                    <div className="flex flex-1 items-center justify-between">
                      <div className="flex flex-col gap-0.5">
                        <span className="font-medium text-[#37352F] text-sm">
                          {item.name}
                        </span>
                        <div className="flex items-center gap-2">
                          <span className="text-sm text-[#37352F]/60 font-mono bg-gray-100 px-1 rounded">
                            {item.code}
                          </span>
                        </div>
                      </div>
                      <span className="font-mono font-bold text-[#37352F] text-sm">
                        ¥{item.unitPrice.toLocaleString()}
                      </span>
                    </div>
                  </CommandItem>
                ))}
              </CommandGroup>
              {/* Show separator only when not filtering by category (cleaner look) */}
              {!activeCategory && <CommandSeparator />}
            </React.Fragment>
          );
        })}
      </CommandList>
    </CommandDialog>
  );
}
