import type { ActiveFilter } from "@/components/shared/NotionFilter/types";
import type { TreatmentItem } from "@/lib/transforms/treatment";

export type TreatmentTreeItem = TreatmentItem & { children: TreatmentItem[] };

export type TreatmentVirtualRow =
  | { type: "root"; item: TreatmentTreeItem; isExpanded: boolean }
  | { type: "child"; item: TreatmentItem };

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

  const lower = searchTerm.toLowerCase();
  return items.filter(
    (root) =>
      root.name.toLowerCase().includes(lower) ||
      root.children.some((child) => child.name.toLowerCase().includes(lower)),
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
