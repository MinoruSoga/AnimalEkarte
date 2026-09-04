import { useCallback, useMemo, useState } from "react";

interface ActiveSortItem {
  key: string;
  direction: "asc" | "desc";
}

interface UseSortableDataOptions<T> {
  /** 数値ソートを適用するキー（デフォルト: 全キーを文字列ソート） */
  numericKeys?: string[];
  /**
   * キーごとのソート値を返すカスタム関数。
   * 派生値（例: calculateTotal）でソートする場合に使用。
   * useCallback でラップして参照安定性を保つこと。
   */
  getSortValue?: (item: T, key: string) => string | number;
}

export function useSortableData<T extends object>(data: T[], options?: UseSortableDataOptions<T>) {
  const [activeSorts, setActiveSorts] = useState<ActiveSortItem[]>([]);

  const toggleSort = useCallback((key: string) => {
    setActiveSorts((prev) => {
      const existing = prev.find((s) => s.key === key);
      if (!existing) return [{ key, direction: "asc" as const }];
      if (existing.direction === "asc") {
        return prev.map((s) => (s.key === key ? { ...s, direction: "desc" as const } : s));
      }
      return prev.filter((s) => s.key !== key);
    });
  }, []);

  const directionFor = useCallback(
    (key: string): "ascending" | "descending" | "none" => {
      const sort = activeSorts.find((s) => s.key === key);
      if (!sort) return "none";
      return sort.direction === "asc" ? "ascending" : "descending";
    },
    [activeSorts],
  );

  const numericKeys = options?.numericKeys;
  const getSortValue = options?.getSortValue;

  const sortedData = useMemo(() => {
    if (activeSorts.length === 0) return [...data];
    const sorted = [...data];
    sorted.sort((a, b) => {
      for (const sort of activeSorts) {
        const key = sort.key;

        if (getSortValue) {
          const aVal = getSortValue(a, key);
          const bVal = getSortValue(b, key);
          if (typeof aVal === "number" && typeof bVal === "number") {
            const cmp = aVal - bVal;
            if (cmp !== 0) return sort.direction === "asc" ? cmp : -cmp;
          } else {
            const cmp = String(aVal).localeCompare(String(bVal), "ja");
            if (cmp !== 0) return sort.direction === "asc" ? cmp : -cmp;
          }
        } else if (numericKeys?.includes(key)) {
          const cmp =
            Number((a as Record<string, unknown>)[key] ?? 0) -
            Number((b as Record<string, unknown>)[key] ?? 0);
          if (cmp !== 0) return sort.direction === "asc" ? cmp : -cmp;
        } else {
          const aVal = String((a as Record<string, unknown>)[key] ?? "");
          const bVal = String((b as Record<string, unknown>)[key] ?? "");
          const cmp = aVal.localeCompare(bVal, "ja");
          if (cmp !== 0) return sort.direction === "asc" ? cmp : -cmp;
        }
      }
      return 0;
    });
    return sorted;
  }, [data, activeSorts, numericKeys, getSortValue]);

  return { activeSorts, setActiveSorts, toggleSort, directionFor, sortedData };
}
