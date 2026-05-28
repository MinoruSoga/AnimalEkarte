// React/Framework
import { useState, useCallback, useMemo, memo, Fragment } from "react";

// External
import { Search, X } from "lucide-react";

// Internal
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";
import { C, ICON } from "@/lib/design-tokens";
import {
  useGetAllConsultations,
  useGetAllProcedures,
  useGetAllVaccinesMaster,
  useGetAllCheckupTypes,
} from "@/hooks/use-treatment-master";

// --- Types ---
export type TreatmentMasterItem = {
  id: string;
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

const CategoryFilter = memo(function CategoryFilter({
  categories,
  activeCategory,
  onSelectCategory,
}: CategoryFilterProps) {
  return (
    <div className={`flex gap-2 overflow-x-auto items-center ${C.bgPage30} rounded-md p-2 [&::-webkit-scrollbar]:hidden [-ms-overflow-style:none] [scrollbar-width:none]`}>
      <div className="flex gap-1.5 min-w-max">
        {activeCategory ? (
          <Badge
            asChild
            variant="outline"
            className={`h-8 px-3 text-sm cursor-pointer ${C.hoverBgMedium} gap-1 ${C.text60} border-transparent bg-transparent`}
          >
            <button type="button" onClick={() => onSelectCategory(null)}>
              <X className={ICON.action} />
              解除
            </button>
          </Badge>
        ) : null}
        {categories.map((category) => {
          const isSelected = activeCategory === category;
          return (
            <Badge
              asChild
              key={category}
              variant={isSelected ? "default" : "outline"}
              className={cn(
                "h-8 px-2.5 text-sm cursor-pointer hover:opacity-80 transition-all",
                isSelected
                  ? `${C.bgAccent} ${C.textWhite} ${C.bgAccentHover} border-transparent`
                  : `bg-white ${C.text} ${C.hoverBgLight} ${C.borderMedium}`
              )}
            >
              <button type="button" onClick={() => onSelectCategory(isSelected ? null : category)}>
                {category}
              </button>
            </Badge>
          );
        })}
      </div>
    </div>
  );
});

// --- Main Component ---

export const TreatmentSearchDialog = memo(function TreatmentSearchDialog({
  open,
  onOpenChange,
  onSelect,
}: TreatmentSearchDialogProps) {
  const [activeCategory, setActiveCategory] = useState<string | null>(null);
  const [searchTerm, setSearchTerm] = useState("");

  // Fetch master data from APIs
  const { data: consultations = [] } = useGetAllConsultations();
  const { data: procedures = [] } = useGetAllProcedures();
  const { data: vaccines = [] } = useGetAllVaccinesMaster();
  const { data: checkupTypes = [] } = useGetAllCheckupTypes();

  const handleOpenChange = useCallback((nextOpen: boolean) => {
    if (!nextOpen) {
      setActiveCategory(null);
      setSearchTerm("");
    }
    onOpenChange(nextOpen);
  }, [onOpenChange]);

  // Build treatment master from API data
  const TREATMENT_MASTER = useMemo(() => {
    const items: TreatmentMasterItem[] = [];

    consultations.forEach((c) => {
      if (c.isActive) {
        items.push({ id: c.id, name: c.name, unitPrice: c.price, category: "診察" });
      }
    });

    procedures.forEach((p) => {
      if (p.isActive) {
        items.push({ id: p.id, name: p.name, unitPrice: p.price, category: "処置" });
      }
    });

    vaccines.forEach((v) => {
      if (v.isActive) {
        items.push({ id: v.id, name: v.name, unitPrice: v.price, category: "予防" });
      }
    });

    checkupTypes.forEach((ct) => {
      if (ct.isActive) {
        items.push({ id: ct.id, name: ct.name, unitPrice: ct.price, category: "検査" });
      }
    });

    return items;
  }, [consultations, procedures, vaccines, checkupTypes]);

  // Filter items by search term and category
  const filteredItems = useMemo(() => {
    return TREATMENT_MASTER.filter((item) => {
      const matchesSearch = !searchTerm || item.name.toLowerCase().includes(searchTerm.toLowerCase());
      const matchesCategory = !activeCategory || item.category === activeCategory;
      return matchesSearch && matchesCategory;
    });
  }, [TREATMENT_MASTER, searchTerm, activeCategory]);

  // Group filtered items by category
  const groupedItems = useMemo(() => {
    return filteredItems.reduce((acc, item) => {
      if (!acc[item.category]) acc[item.category] = [];
      acc[item.category].push(item);
      return acc;
    }, {} as Record<string, TreatmentMasterItem[]>);
  }, [filteredItems]);

  // Calculate all categories once
  const allCategories = useMemo(() => {
    return [
      ...CATEGORY_ORDER,
      ...Object.keys(groupedItems).filter((cat) => !CATEGORY_ORDER.includes(cat)),
    ];
  }, [groupedItems]);

  const handleSelect = useCallback((item: TreatmentMasterItem) => {
    onSelect(item);
    onOpenChange(false);
    setSearchTerm("");
    setActiveCategory(null);
  }, [onSelect, onOpenChange]);

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-[560px] max-h-[80vh] flex flex-col gap-3">
        <DialogHeader>
          <DialogTitle className={`text-base font-bold ${C.text}`}>治療プラン検索</DialogTitle>
          <DialogDescription className="sr-only">
            治療プランを検索して選択します。
          </DialogDescription>
        </DialogHeader>

        {/* Search */}
        <div className="relative">
          <Search className={`absolute left-2.5 top-1/2 -translate-y-1/2 ${ICON.action} ${C.text40}`} />
          <Input
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            placeholder="治療プランを検索..."
            className={`pl-9 h-11 text-sm bg-white ${C.borderMedium}`}
          />
          {searchTerm ? (
            <button type="button"
              onClick={() => setSearchTerm("")}
              className={`absolute right-2.5 top-1/2 -translate-y-1/2 ${C.text40} ${C.hoverText}`}
            >
              <X className={ICON.xs} />
            </button>
          ) : null}
        </div>

        {/* Category Filter */}
        <CategoryFilter
          categories={allCategories}
          activeCategory={activeCategory}
          onSelectCategory={setActiveCategory}
        />

        {/* Item List */}
        <div className="flex-1 overflow-y-auto space-y-1 pr-1 max-h-[400px]">
          {filteredItems.length === 0 ? (
            <div className={`py-12 text-center text-sm ${C.text60}`}>
              該当する治療プランが見つかりません。
            </div>
          ) : (
            CATEGORY_ORDER.map((category) => {
              const items = groupedItems[category];
              if (!items?.length) return null;
              return (
                <Fragment key={category}>
                  {/* Category Header */}
                  {!activeCategory ? (
                    <div className={`px-2 py-1.5 text-xs font-semibold ${C.text40} uppercase tracking-wider`}>
                      {category}
                    </div>
                  ) : null}
                  <div className="space-y-1.5">
                    {items.map((item) => (
                      <button
                        key={item.id}
                        type="button"
                        onClick={() => handleSelect(item)}
                        className={`w-full text-left p-3 border rounded-lg cursor-pointer transition-all flex items-center justify-between group bg-white ${C.borderMedium} ${C.hoverBorderPrimary30} ${C.hoverBgPageHalf}`}
                      >
                        <div className="flex-1 min-w-0">
                          <div className={`text-sm font-medium ${C.text}`}>{item.name}</div>
                          <div className={`text-xs ${C.text60} mt-0.5`}>
                            ¥{item.unitPrice.toLocaleString()}
                          </div>
                        </div>
                        <div className={`size-5 rounded-full border ${C.borderLight} group-hover:border-current transition-colors shrink-0 ml-3`} />
                      </button>
                    ))}
                  </div>
                </Fragment>
              );
            })
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
});
