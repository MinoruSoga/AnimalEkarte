import { useState, useMemo, useCallback } from "react";

type SortDirection = "ascending" | "descending" | "none";

const DIRECTION_CYCLE: Record<SortDirection, SortDirection> = {
  none: "ascending",
  ascending: "descending",
  descending: "none",
};

interface UseTableSortOptions<T, K extends string> {
  accessor: (item: T, key: K) => string;
  comparator?: (a: T, b: T, key: K) => number;
  locale?: string;
  initialKey?: K;
  initialDirection?: SortDirection;
}

interface UseTableSortReturn<T, K extends string> {
  sortedData: T[];
  directionFor: (key: K) => SortDirection;
  toggleSort: (key: K) => void;
  resetSort: () => void;
}

export function useTableSort<T, K extends string>(
  data: readonly T[],
  options: UseTableSortOptions<T, K>,
): UseTableSortReturn<T, K> {
  const { accessor, comparator, locale = "ja" } = options;

  const initialKey = options.initialKey ?? null;
  const initialDir: SortDirection = options.initialKey
    ? (options.initialDirection ?? "ascending")
    : "none";

  const [sortKey, setSortKey] = useState<K | null>(initialKey);
  const [sortDir, setSortDir] = useState<SortDirection>(initialDir);

  const resetSort = useCallback(() => {
    setSortKey(initialKey);
    setSortDir(initialDir);
  }, [initialKey, initialDir]);

  const toggleSort = useCallback(
    (key: K) => {
      if (sortKey === key) {
        const next = DIRECTION_CYCLE[sortDir];
        setSortDir(next);
        if (next === "none") setSortKey(null);
      } else {
        setSortKey(key);
        setSortDir("ascending");
      }
    },
    [sortKey, sortDir],
  );

  const sortedData = useMemo(() => {
    if (!sortKey || sortDir === "none") return [...data];
    const compare = comparator
      ? (a: T, b: T) => comparator(a, b, sortKey)
      : (a: T, b: T) => accessor(a, sortKey).localeCompare(accessor(b, sortKey), locale);
    const sorted = [...data].sort(compare);
    return sortDir === "descending" ? sorted.reverse() : sorted;
  }, [data, sortKey, sortDir, accessor, comparator, locale]);

  const directionFor = useCallback(
    (key: K): SortDirection => (sortKey === key ? sortDir : "none"),
    [sortKey, sortDir],
  );

  return { sortedData, directionFor, toggleSort, resetSort };
}
