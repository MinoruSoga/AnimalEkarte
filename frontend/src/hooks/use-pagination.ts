import { useMemo, useState, useCallback } from "react";

interface UsePaginationOptions {
  /** Items per page (default: 20) */
  pageSize?: number;
  /** When this value changes, reset to page 1 (e.g., search term) */
  resetKey?: string | number;
}

interface UsePaginationReturn<T> {
  /** Current page (1-based) */
  currentPage: number;
  /** Total number of pages */
  totalPages: number;
  /** Items for the current page */
  paginatedData: T[];
  /** Total item count (before pagination) */
  totalCount: number;
  /** Go to specific page */
  goToPage: (page: number) => void;
  /** Go to next page */
  nextPage: () => void;
  /** Go to previous page */
  prevPage: () => void;
  /** Reset to page 1 */
  resetPage: () => void;
  /** Page size */
  pageSize: number;
  /** Start index (1-based, for display) */
  startIndex: number;
  /** End index (1-based, for display) */
  endIndex: number;
}

export function usePagination<T>(
  data: T[],
  options: UsePaginationOptions = {},
): UsePaginationReturn<T> {
  const { pageSize = 20, resetKey } = options;
  const [currentPage, setCurrentPage] = useState(1);

  // rerender-derived-state-no-effect: useEffect の代わりにレンダー中に derived state で処理
  const [prevResetKey, setPrevResetKey] = useState(resetKey);
  if (prevResetKey !== resetKey) {
    setPrevResetKey(resetKey);
    setCurrentPage(1);
  }

  const totalCount = data.length;
  const totalPages = Math.max(1, Math.ceil(totalCount / pageSize));

  // Auto-correct if currentPage exceeds totalPages (e.g., after filtering)
  const safePage = Math.min(currentPage, totalPages);

  const paginatedData = useMemo(() => {
    const start = (safePage - 1) * pageSize;
    return data.slice(start, start + pageSize);
  }, [data, safePage, pageSize]);

  const startIndex = totalCount === 0 ? 0 : (safePage - 1) * pageSize + 1;
  const endIndex = Math.min(safePage * pageSize, totalCount);

  const goToPage = useCallback(
    (page: number) => {
      setCurrentPage(Math.max(1, Math.min(page, totalPages)));
    },
    [totalPages],
  );

  const nextPage = useCallback(() => {
    setCurrentPage((p) => Math.min(p + 1, totalPages));
  }, [totalPages]);

  const prevPage = useCallback(() => {
    setCurrentPage((p) => Math.max(p - 1, 1));
  }, []);

  const resetPage = useCallback(() => {
    setCurrentPage(1);
  }, []);

  return {
    currentPage: safePage,
    totalPages,
    paginatedData,
    totalCount,
    goToPage,
    nextPage,
    prevPage,
    resetPage,
    pageSize,
    startIndex,
    endIndex,
  };
}
