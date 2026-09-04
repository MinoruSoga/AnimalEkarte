import type { ActiveFilter } from "@/components/shared/PropertyFilter/types";
import type { TreatmentItem } from "@/lib/transforms/treatment";
import { normalizeKana } from "@/lib/normalize-kana";

export interface TreatmentTabConfig {
  data: TreatmentItem[] | undefined;
  entityLabel: string;
  emptyMessage: string;
  searchPlaceholder: string;
  onReorder: (ids: number[]) => void;
}

export type TreatmentTreeItem = TreatmentItem & { children: TreatmentItem[] };

export type TreatmentVirtualRow =
  | { type: "root"; item: TreatmentTreeItem; isExpanded: boolean }
  | { type: "child"; item: TreatmentItem };

export const TREATMENT_COLUMNS = [
  { header: "", className: "w-11 px-0" },
  { header: "名称" },
  { header: "単価(税込)", className: "w-[120px]", align: "right" as const },
  { header: "ステータス", className: "w-[100px]", align: "center" as const },
  { header: "操作", className: "w-[80px]", align: "right" as const },
];

export function buildTreatmentTree(items: TreatmentItem[]): TreatmentTreeItem[] {
  const map = new Map<string, TreatmentTreeItem>(
    items.map((item) => [item.id, { ...item, children: [] }]),
  );
  const roots: TreatmentTreeItem[] = [];
  for (const item of map.values()) {
    if (item.parentId) {
      map.get(item.parentId)?.children.push(item);
    } else {
      roots.push(item);
    }
  }
  return roots;
}

export function filterTreatmentRoots({
  roots,
  activeFilters,
  searchTerm,
}: {
  roots: TreatmentTreeItem[];
  activeFilters: ActiveFilter[];
  searchTerm: string;
}): TreatmentTreeItem[] {
  let items = roots;
  for (const filter of activeFilters) {
    if (filter.key === "status" && typeof filter.value === "string") {
      const want = filter.value === "active";
      items = items.filter((root) =>
        filter.condition === "is" ? root.isActive === want : root.isActive !== want,
      );
    }
  }

  if (!searchTerm) return items;

  const lower = normalizeKana(searchTerm).toLowerCase();
  return items.filter(
    (root) =>
      normalizeKana(root.name).toLowerCase().includes(lower) ||
      root.children.some((child) => normalizeKana(child.name).toLowerCase().includes(lower)),
  );
}

export function buildTreatmentRows({
  roots,
  expandedIds,
}: {
  roots: TreatmentTreeItem[];
  expandedIds: ReadonlySet<string>;
}): TreatmentVirtualRow[] {
  return roots.flatMap((root) => {
    const isExpanded = expandedIds.has(root.id);
    const rows: TreatmentVirtualRow[] = [{ type: "root", item: root, isExpanded }];
    if (isExpanded && root.children.length > 0) {
      for (const child of root.children) {
        rows.push({ type: "child", item: child });
      }
    }
    return rows;
  });
}
