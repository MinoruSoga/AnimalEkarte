// React/Framework
import { useState, useCallback, useMemo, useDeferredValue, memo, Fragment } from "react";

// Internal
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { EmptyState } from "@/components/shared/DataStates";
import { ClearableSearchInput } from "@/components/shared/ClearableSearchInput";
import { CategoryChipsFilter } from "@/components/shared/CategoryChipsFilter";
import { FilteringIndicator } from "@/components/shared/FilteringIndicator/FilteringIndicator";
import { C } from "@/lib/design-tokens";
import { normalizeKana } from "@/lib/normalize-kana";
import { formatCurrency } from "@/lib/format/number";
import {
  useGetAllConsultations,
  useGetAllProcedures,
  useGetAllVaccinesMaster,
  useGetAllCheckupTypes,
  useGetAllMedicinesMaster,
} from "@/hooks/use-treatment-master";

// --- Types ---
export type TreatmentMasterItem = {
  id: string;
  name: string;
  unitPrice: number;
  category: string;
  /** #201: category="薬剤" の場合のみ設定。投与量自動計算・medicine_id 紐付けに使う */
  medicineId?: string;
};

interface TreatmentSearchDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSelect: (item: TreatmentMasterItem) => void;
}

// --- Constants ---
const CATEGORY_ORDER = ["診察", "検査", "処置", "予防", "入院", "薬剤"];

// --- Main Component ---

export const TreatmentSearchDialog = memo(function TreatmentSearchDialog({
  open,
  onOpenChange,
  onSelect,
}: TreatmentSearchDialogProps) {
  const [activeCategory, setActiveCategory] = useState<string | null>(null);
  const [searchTerm, setSearchTerm] = useState("");
  const deferredSearchTerm = useDeferredValue(searchTerm);
  const isFiltering = searchTerm !== deferredSearchTerm;

  // Fetch master data from APIs
  const { data: consultations = [] } = useGetAllConsultations();
  const { data: procedures = [] } = useGetAllProcedures();
  const { data: vaccines = [] } = useGetAllVaccinesMaster();
  const { data: checkupTypes = [] } = useGetAllCheckupTypes();
  const { data: medicines = [] } = useGetAllMedicinesMaster();

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

    medicines.forEach((m) => {
      // カテゴリ見出し行のみ除外。未分類でも剤形/単位/単価があれば薬剤として出す (BUG-006)。
      const isCategoryPlaceholder =
        !m.parentId && m.price === 0 && !m.dosageForm && !m.medicineUnit;
      if (m.isActive && !isCategoryPlaceholder) {
        items.push({ id: m.id, name: m.name, unitPrice: m.price, category: "薬剤", medicineId: m.id });
      }
    });

    return items;
  }, [consultations, procedures, vaccines, checkupTypes, medicines]);

  const searchableItems = useMemo(
    () => TREATMENT_MASTER.map((item) => ({
      item,
      normalizedName: normalizeKana(item.name).toLowerCase(),
    })),
    [TREATMENT_MASTER],
  );

  // Filter items by search term and category（カタカナ・ひらがな非区別）
  const filteredItems = useMemo(() => {
    const normalizedTerm = deferredSearchTerm ? normalizeKana(deferredSearchTerm).toLowerCase() : "";
    return searchableItems.flatMap(({ item, normalizedName }) => {
      const matchesSearch = !deferredSearchTerm || normalizedName.includes(normalizedTerm);
      const matchesCategory = !activeCategory || item.category === activeCategory;
      return matchesSearch && matchesCategory ? [item] : [];
    });
  }, [searchableItems, deferredSearchTerm, activeCategory]);

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
        <ClearableSearchInput
          value={searchTerm}
          onChange={setSearchTerm}
          placeholder="治療プランを検索..."
        />

        {/* Category Filter */}
        <CategoryChipsFilter
          categories={allCategories}
          activeCategory={activeCategory}
          onSelectCategory={setActiveCategory}
          chipRounded="sm"
          hoverEffect="badge"
        />

        {/* Item List */}
        <FilteringIndicator isFiltering={isFiltering} className="flex-1 overflow-y-auto space-y-1 pr-1 max-h-[400px]">
          {filteredItems.length === 0 ? (
            <EmptyState message="該当する治療プランが見つかりません。" />
          ) : (
            CATEGORY_ORDER.map((category) => {
              const items = groupedItems[category];
              if (!items?.length) return null;
              return (
                <Fragment key={category}>
                  {/* Category Header */}
                  {!activeCategory ? (
                    <div className={`px-2 py-1.5 text-2xs font-semibold ${C.text40} uppercase`}>
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
                            {formatCurrency(item.unitPrice)}
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
        </FilteringIndicator>
      </DialogContent>
    </Dialog>
  );
});
