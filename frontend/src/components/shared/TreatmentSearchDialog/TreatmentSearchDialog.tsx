// React/Framework
import { C, ICON } from "@/lib/design-tokens";
import * as React from "react";
import { useMemo } from "react";

// External
import { X } from "lucide-react";

// Internal
import { CommandDialog, CommandInput, CommandList, CommandEmpty, CommandGroup, CommandItem, CommandSeparator } from "@/components/ui/command";
import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";
import { useGetAllConsultations } from "@/features/master/api/consultations";
import { useGetAllProcedures } from "@/features/master/api/procedures";
import { useGetAllVaccinesMaster } from "@/features/master/api/vaccines-master";
import { useGetAllCheckupTypes } from "@/features/master/api/checkup-types";

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

  // Fetch master data from APIs
  const { data: consultations = [] } = useGetAllConsultations();
  const { data: procedures = [] } = useGetAllProcedures();
  const { data: vaccines = [] } = useGetAllVaccinesMaster();
  const { data: checkupTypes = [] } = useGetAllCheckupTypes();

  // Reset category when dialog closes
  React.useEffect(() => {
    if (!open) {
      setActiveCategory(null);
    }
  }, [open]);

  // Build treatment master from API data (動的・実データ使用)
  const TREATMENT_MASTER = useMemo(() => {
    const items: TreatmentMasterItem[] = [];

    // Consultations → 診察
    consultations.forEach((c, idx) => {
      if (c.isActive) {
        items.push({
          id: c.id,
          code: `1${String(idx + 1).padStart(3, "0")}`,
          name: c.name,
          unitPrice: c.price,
          category: "診察",
        });
      }
    });

    // Procedures → 処置
    procedures.forEach((p, idx) => {
      if (p.isActive) {
        items.push({
          id: p.id,
          code: `4${String(idx + 1).padStart(3, "0")}`,
          name: p.name,
          unitPrice: p.price,
          category: "処置",
        });
      }
    });

    // Vaccines → 予防
    vaccines.forEach((v, idx) => {
      if (v.isActive) {
        items.push({
          id: v.id,
          code: `2${String(idx + 1).padStart(3, "0")}`,
          name: v.name,
          unitPrice: v.price,
          category: "予防",
        });
      }
    });

    // CheckupTypes → 検査
    checkupTypes.forEach((ct, idx) => {
      if (ct.isActive) {
        items.push({
          id: ct.id,
          code: `3${String(idx + 1).padStart(3, "0")}`,
          name: ct.name,
          unitPrice: ct.price,
          category: "検査",
        });
      }
    });

    return items;
  }, [consultations, procedures, vaccines, checkupTypes]);

  // Memoize grouped items calculation
  const groupedItems = React.useMemo(() => {
    return TREATMENT_MASTER.reduce((acc, item) => {
      if (!acc[item.category]) acc[item.category] = [];
      acc[item.category].push(item);
      return acc;
    }, {} as Record<string, TreatmentMasterItem[]>);
  }, [TREATMENT_MASTER]);

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
