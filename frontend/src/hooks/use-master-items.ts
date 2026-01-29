import { useState, useMemo } from "react";
import type { MasterItem } from "@/types";

const STORAGE_KEY = "med_master_items";

// Initial data merged from various sources
const INITIAL_MASTER_ITEMS: MasterItem[] = [
  // --- Consultation (診察) ---
  { id: "tm_1001", code: "1001", name: "再診料(再診)", price: 800, category: "診察", status: "active" },
  { id: "tm_1002", code: "1002", name: "初診料", price: 1500, category: "診察", status: "active" },
  { id: "tm_1003", code: "1003", name: "時間外診察料", price: 2000, category: "診察", status: "active" },

  // --- Service Type (診療内容) ---
  { id: "sv_001", code: "SV-001", name: "診療", price: 0, category: "診療内容", status: "active" },
  { id: "sv_002", code: "SV-002", name: "トリミング", price: 0, category: "診療内容", status: "active" },
  { id: "sv_003", code: "SV-003", name: "入院", price: 0, category: "診療内容", status: "active" },
  { id: "sv_004", code: "SV-004", name: "ワクチン", price: 0, category: "診療内容", status: "active" },
  { id: "sv_005", code: "SV-005", name: "手術", price: 0, category: "診療内容", status: "active" },
  { id: "sv_006", code: "SV-006", name: "検査", price: 0, category: "診療内容", status: "active" },

  // --- Vaccine/Prevention (予防) ---
  { id: "tm_2001", code: "VC-001", name: "狂犬病ワクチン", price: 3500, category: "予防", status: "active" },
  { id: "tm_2002", code: "VC-002", name: "5種混合ワクチン", price: 6000, category: "予防", status: "active", inventoryId: "1", defaultQuantity: 1 },
  { id: "tm_2003", code: "VC-003", name: "8種混合ワクチン", price: 8000, category: "予防", status: "active" },
  { id: "tm_2004", code: "VC-004", name: "フィラリア予防薬(S)", price: 1200, category: "予防", status: "active" },
  { id: "tm_2005", code: "VC-005", name: "ジステンパーワクチン", price: 4000, category: "予防", status: "active" },
  { id: "tm_2006", code: "VC-006", name: "ノミダニ予防薬", price: 1500, category: "予防", status: "active" },
  { id: "tm_2007", code: "VC-007", name: "混合ワクチン(7種)", price: 8000, category: "予防", status: "active" },
  // From VaccinationForm selection
  { id: "tm_2008", code: "VC-008", name: "食道炎予防措置", price: 2000, category: "予防", status: "active" }, // Renamed for clarity, assumed preventive measure

  // --- Examination (検査) ---
  { id: "tm_3001", code: "EX-001", name: "血液検査(CBC)", price: 3000, category: "検査", status: "active", inventoryId: "3", defaultQuantity: 1 },
  { id: "tm_3002", code: "EX-002", name: "生化学検査(16項目)", price: 8000, category: "検査", status: "active" },
  { id: "tm_3003", code: "EX-003", name: "レントゲン(胸部)", price: 5000, category: "検査", status: "active" },
  { id: "tm_3004", code: "EX-004", name: "超音波検査(腹部)", price: 6000, category: "検査", status: "active" },
  { id: "tm_3005", code: "EX-005", name: "尿検査", price: 1500, category: "検査", status: "active" },
  { id: "tm_3006", code: "EX-006", name: "糞便検査", price: 1000, category: "検査", status: "active" },
  { id: "tm_3007", code: "EX-007", name: "血液検査セットA", price: 5000, category: "検査", status: "active" },
  { id: "tm_3008", code: "EX-008", name: "血液検査セットB(生化学)", price: 7000, category: "検査", status: "active" },

  // --- Procedure (処置) ---
  { id: "tm_4001", code: "4001", name: "爪切り", price: 500, category: "処置", status: "active" },
  { id: "tm_4002", code: "4002", name: "耳掃除", price: 800, category: "処置", status: "active" },
  { id: "tm_4003", code: "4003", name: "肛門腺絞り", price: 500, category: "処置", status: "active" },

  // --- Hospitalization (入院) ---
  { id: "tm_5001", code: "5001", name: "入院料(小型)", price: 3000, category: "入院", status: "active" },
  { id: "tm_5002", code: "5002", name: "入院料(中型)", price: 4000, category: "入院", status: "active" },

  // --- Medicine (薬剤) ---
  { id: "tm_6001", code: "MD-001", name: "アモキシシリン", price: 100, category: "薬剤", status: "active", inventoryId: "4", defaultQuantity: 10, description: "1錠あたり" },
  { id: "tm_6002", code: "MD-002", name: "プレドニゾロン", price: 80, category: "薬剤", status: "active", description: "1錠あたり" },
  { id: "tm_6003", code: "MD-003", name: "メトクロプラミド", price: 120, category: "薬剤", status: "active", description: "1錠あたり" },
  { id: "tm_6004", code: "6001", name: "内服薬A(抗生剤)", price: 100, category: "薬剤", status: "active" },
  { id: "tm_6005", code: "6002", name: "内服薬B(消炎剤)", price: 80, category: "薬剤", status: "active" },

  // --- Staff (スタッフ) ---
  { id: "st_001", code: "ST-001", name: "医師A", price: 0, category: "スタッフ", status: "active" },
  { id: "st_002", code: "ST-002", name: "医師B", price: 0, category: "スタッフ", status: "active" },
  { id: "st_003", code: "ST-003", name: "医師C", price: 0, category: "スタッフ", status: "active" },
  { id: "st_004", code: "ST-004", name: "スタッフA", price: 0, category: "スタッフ", status: "active" },
  { id: "st_005", code: "ST-005", name: "スタッフB", price: 0, category: "スタッフ", status: "active" },

  // --- Insurance (保険) ---
  { id: "in_001", code: "IN-001", name: "アニコム", price: 0, category: "保険", status: "active" },
  { id: "in_002", code: "IN-002", name: "アイペット", price: 0, category: "保険", status: "active" },

  // --- Cage (ケージ) ---
  { id: "c1", code: "ICU-01", name: "ICU-1", price: 0, category: "ケージ", description: "ICU - 酸素室対応", status: "active" },
  { id: "c2", code: "ICU-02", name: "ICU-2", price: 0, category: "ケージ", description: "ICU - 酸素室対応", status: "active" },
  { id: "c3", code: "DOG-01", name: "犬舎1", price: 0, category: "ケージ", description: "犬舎 - 大型", status: "active" },
  { id: "c4", code: "DOG-02", name: "犬舎2", price: 0, category: "ケージ", description: "犬舎 - 中型", status: "active" },
  { id: "c5", code: "DOG-03", name: "犬舎3", price: 0, category: "ケージ", description: "犬舎 - 中型", status: "active" },
  { id: "c6", code: "CAT-01", name: "猫舎1", price: 0, category: "ケージ", description: "猫舎 - 防音", status: "active" },
  { id: "c7", code: "CAT-02", name: "猫舎2", price: 0, category: "ケージ", description: "猫舎 - 防音", status: "active" },

  // --- Trimming Course (トリミングコース) ---
  { id: "tc_001", code: "TC-001", name: "シャンプーコース(小型)", price: 4000, category: "トリミングコース", status: "active" },
  { id: "tc_002", code: "TC-002", name: "シャンプーカットコース(小型)", price: 6000, category: "トリミングコース", status: "active" },
  { id: "tc_003", code: "TC-003", name: "シャンプーコース(中型)", price: 5000, category: "トリミングコース", status: "active" },
  { id: "tc_004", code: "TC-004", name: "シャンプーカットコース(中型)", price: 7500, category: "トリミングコース", status: "active" },

  // --- Trimming Option (トリミングオプション) ---
  { id: "to_001", code: "TO-001", name: "薬用シャンプー", price: 500, category: "トリミングオプション", status: "active" },
  { id: "to_002", code: "TO-002", name: "炭酸泉", price: 1000, category: "トリミングオプション", status: "active" },
  { id: "to_003", code: "TO-003", name: "泥パック", price: 1500, category: "トリミングオプション", status: "active" },
  { id: "to_004", code: "TO-004", name: "足裏・足回りカット", price: 500, category: "トリミングオプション", status: "active" },
  { id: "to_005", code: "TO-005", name: "歯磨き", price: 500, category: "トリミングオプション", status: "active" },
];

const CATEGORY_MAP: Record<string, string> = {
  examination: "検査",
  vaccine: "予防",
  medicine: "薬剤",
  consultation: "診察",
  serviceType: "診療内容",
  procedure: "処置",
  hospitalization: "入院",
  staff: "スタッフ",
  insurance: "保険",
  cage: "ケージ",
  trimming_course: "トリミングコース",
  trimming_option: "トリミングオプション",
};

export function useMasterItems(category?: string, searchTerm?: string) {
  const [items, setItems] = useState<MasterItem[]>(() => {
    const stored = localStorage.getItem(STORAGE_KEY);
    if (stored) {
      try {
        const parsedItems: MasterItem[] = JSON.parse(stored);

        // Merge with INITIAL_MASTER_ITEMS to ensure new categories/items appear
        // This fixes the issue where new master items defined in code don't show up
        // because localStorage persists the old list.
        let hasChanges = false;
        const currentIds = new Set(parsedItems.map((i) => i.id));
        const mergedItems = [...parsedItems];

        INITIAL_MASTER_ITEMS.forEach((initItem) => {
          if (!currentIds.has(initItem.id)) {
            mergedItems.push(initItem);
            hasChanges = true;
          }
        });

        if (hasChanges) {
          localStorage.setItem(STORAGE_KEY, JSON.stringify(mergedItems));
        }
        return hasChanges ? mergedItems : parsedItems;
      } catch {
        // Reset to initial data if parsing fails
        localStorage.setItem(STORAGE_KEY, JSON.stringify(INITIAL_MASTER_ITEMS));
        return INITIAL_MASTER_ITEMS;
      }
    }
    localStorage.setItem(STORAGE_KEY, JSON.stringify(INITIAL_MASTER_ITEMS));
    return INITIAL_MASTER_ITEMS;
  });

  const filteredItems = useMemo(() => {
    let res = items;
    if (category && category !== "all") {
      const targetCat = CATEGORY_MAP[category] || category;
      res = res.filter((i) => i.category === targetCat);
    }

    if (searchTerm) {
      const lower = searchTerm.toLowerCase();
      res = res.filter(
        (i) =>
          i.name.toLowerCase().includes(lower) ||
          i.code.toLowerCase().includes(lower) ||
          (i.category && i.category.toLowerCase().includes(lower))
      );
    }
    return res;
  }, [items, category, searchTerm]);

  const add = (item: Omit<MasterItem, "id">) => {
    const newItem = { ...item, id: `tm_${Date.now()}` };
    const newItems = [...items, newItem];
    setItems(newItems);
    localStorage.setItem(STORAGE_KEY, JSON.stringify(newItems));
    return newItem;
  };

  const update = (id: string, updates: Partial<MasterItem>) => {
    const newItems = items.map((i) =>
      i.id === id ? { ...i, ...updates } : i
    );
    setItems(newItems);
    localStorage.setItem(STORAGE_KEY, JSON.stringify(newItems));
  };

  const remove = (id: string) => {
    const newItems = items.filter((i) => i.id !== id);
    setItems(newItems);
    localStorage.setItem(STORAGE_KEY, JSON.stringify(newItems));
  };

  return { data: filteredItems, add, update, remove, isLoading: false };
}
