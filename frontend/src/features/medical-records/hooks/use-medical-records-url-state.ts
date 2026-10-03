// React/Framework
import { useCallback, useLayoutEffect, useRef, useState } from "react";
import { useSearchParams } from "react-router";

// Types
import type { MedicalRecordSortKey } from "../api/get-medical-records";

interface UseMedicalRecordsUrlStateResult {
  searchParams: URLSearchParams;
  currentPage: number;
  sortKey: MedicalRecordSortKey | undefined;
  sortOrder: "asc" | "desc";
  handleSortToggle: (key: MedicalRecordSortKey) => void;
  directionForSort: (key: MedicalRecordSortKey) => "ascending" | "descending" | "none";
  handlePageChange: (page: number) => void;
  resetPage: () => void;
}

// カルテ一覧の URL 同期（page/sort/order）専用フック。
// 検索・フィルタ変更時の1ページ目リセットは、呼出側が search/filters 更新と同じ
// イベント内で resetPage を呼び URL の page パラメータを除去して行う。
// EMR-245: render-phase 調整（prevResetKey）だけでは検出 pass の currentPage=1 が
// React に破棄され、確定描画では URL 上の page が復活する。リセットは URL 更新で
// 確定させる必要があるため、UI 起点の変更は呼出側で resetPage を併用する。
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

  // RouterProvider は router state を React.startTransition 内でコミットするため、
  // setSearchParams の関数型更新が受け取る prev は描画時点の URL に留まり、
  // URL が遷移済みでも古いことがある（遷移コミット前の連続クリックで desc 停滞した実績）。
  // そのため直近に要求・確定したパラメータを ref で保持し、更新は常にそこから組み立てる。
  const latestParamsRef = useRef(searchParams);
  useLayoutEffect(() => {
    latestParamsRef.current = searchParams;
  }, [searchParams]);

  const updateParams = useCallback(
    (mutate: (next: URLSearchParams) => void) => {
      const next = new URLSearchParams(latestParamsRef.current);
      mutate(next);
      latestParamsRef.current = next;
      setSearchParams(next, { replace: true });
    },
    [setSearchParams],
  );

  const handleSortToggle = useCallback(
    (key: MedicalRecordSortKey) => {
      updateParams((next) => {
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
      });
    },
    [updateParams],
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
      updateParams((next) => {
        if (page <= 1) {
          next.delete("page");
        } else {
          next.set("page", String(page));
        }
      });
    },
    [updateParams],
  );

  // EMR-245: 検索・フィルタ変更時の1ページ目リセット。page パラメータを URL から
  // 除去するだけでよい（urlPage は既定 1 に畳まれる）。呼出側の search/filters
  // state 更新と同じイベントで呼ぶことで確定描画から一貫する。
  const resetPage = useCallback(() => {
    updateParams((next) => {
      next.delete("page");
    });
  }, [updateParams]);

  // raw の setSearchParams は latestParamsRef を迂回するため公開しない。
  // URL 更新は必ず updateParams 経由（handleSortToggle/handlePageChange/resetPage）に限定する。
  return {
    searchParams,
    currentPage,
    sortKey,
    sortOrder,
    handleSortToggle,
    directionForSort,
    handlePageChange,
    resetPage,
  };
}
