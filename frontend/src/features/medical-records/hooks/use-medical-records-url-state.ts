// React/Framework
import { useCallback, useState } from "react";
import { useSearchParams } from "react-router";

// Types
import type { MedicalRecordSortKey } from "../api/get-medical-records";

interface UseMedicalRecordsUrlStateResult {
  searchParams: URLSearchParams;
  setSearchParams: ReturnType<typeof useSearchParams>[1];
  currentPage: number;
  sortKey: MedicalRecordSortKey | undefined;
  sortOrder: "asc" | "desc";
  handleSortToggle: (key: MedicalRecordSortKey) => void;
  directionForSort: (key: MedicalRecordSortKey) => "ascending" | "descending" | "none";
  handlePageChange: (page: number) => void;
}

// カルテ一覧の URL 同期（page/sort/order）専用フック。
// 検索・フィルタが変わったら1ページ目へリセットする（useEffect不使用、rerender-derived-state-no-effect）。
export function useMedicalRecordsUrlState(resetKey: string): UseMedicalRecordsUrlStateResult {
  const [searchParams, setSearchParams] = useSearchParams();

  const [prevResetKey, setPrevResetKey] = useState(resetKey);
  const rawUrlPage = Number(searchParams.get("page"));
  const urlPage = Number.isFinite(rawUrlPage) && rawUrlPage > 0 ? Math.floor(rawUrlPage) : 1;
  let currentPage = urlPage;
  if (prevResetKey !== resetKey) {
    setPrevResetKey(resetKey);
    currentPage = 1;
  }

  // B-1 follow-up: 列ソート server 化。URL ?sort=&order= で状態を保持する（page と同様に共有可能にする）。
  const rawSort = searchParams.get("sort");
  const sortKey: MedicalRecordSortKey | undefined =
    rawSort === "date" || rawSort === "owner_name" || rawSort === "pet_name" || rawSort === "status"
      ? rawSort
      : undefined;
  const sortOrder: "asc" | "desc" = searchParams.get("order") === "asc" ? "asc" : "desc";

  const handleSortToggle = useCallback(
    (key: MedicalRecordSortKey) => {
      setSearchParams(
        (prev) => {
          const next = new URLSearchParams(prev);
          const currentSort = next.get("sort");
          const currentOrder = next.get("order") === "asc" ? "asc" : "desc";
          if (currentSort !== key) {
            next.set("sort", key);
            next.set("order", "desc");
          } else if (currentOrder === "desc") {
            next.set("order", "asc");
          } else {
            next.delete("sort");
            next.delete("order");
          }
          next.delete("page");
          return next;
        },
        { replace: true },
      );
    },
    [setSearchParams],
  );

  const directionForSort = useCallback(
    (key: MedicalRecordSortKey): "ascending" | "descending" | "none" => {
      if (sortKey !== key) return "none";
      return sortOrder === "asc" ? "ascending" : "descending";
    },
    [sortKey, sortOrder],
  );

  const handlePageChange = useCallback(
    (page: number) => {
      setSearchParams(
        (prev) => {
          const next = new URLSearchParams(prev);
          if (page <= 1) {
            next.delete("page");
          } else {
            next.set("page", String(page));
          }
          return next;
        },
        { replace: true },
      );
    },
    [setSearchParams],
  );

  return {
    searchParams,
    setSearchParams,
    currentPage,
    sortKey,
    sortOrder,
    handleSortToggle,
    directionForSort,
    handlePageChange,
  };
}
