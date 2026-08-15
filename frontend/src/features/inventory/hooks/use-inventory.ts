import { useMemo } from "react";
import type { InventoryItem } from "@/types";
import { useGetInventoryItemsPage } from "../api/inventory";
import { normalizeKana } from "@/lib/normalize-kana";
type CategoryFilter = InventoryItem["category"] | "all";
type StatusFilter = InventoryItem["status"] | "all";

interface UseInventoryParams {
  searchTerm: string;
  category?: CategoryFilter;
  statusFilter?: StatusFilter;
  page: number;
  limit: number;
}

// rerender-stable-empty-array: 未ロード時のフォールバックを毎レンダー新規配列にしない
const EMPTY_ITEMS: InventoryItem[] = [];

export function useInventoryList({
  searchTerm,
  category = "all",
  statusFilter = "all",
  page,
  limit,
}: UseInventoryParams) {
  const categoryParam = category !== "all" ? category : undefined;
  const statusParam = statusFilter !== "all" ? statusFilter : undefined;

  // BUG-412: page/limit をサーバへ転送し、レスポンスの total/page/limit を消費する
  // （旧実装は data のみ返し total/page/limit を捨てていたため 20 件超が不可視だった）。
  const { data: pageResult, isLoading, isError } = useGetInventoryItemsPage({
    category: categoryParam,
    status: statusParam,
    page,
    limit,
  });
  const items = pageResult?.data ?? EMPTY_ITEMS;

  const filteredItems = useMemo(() => {
    if (!searchTerm) return items;
    const normalizedTerm = normalizeKana(searchTerm).toLowerCase();
    return items.filter(
      (item) =>
        normalizeKana(item.name).toLowerCase().includes(normalizedTerm) ||
        (item.location ? normalizeKana(item.location).toLowerCase().includes(normalizedTerm) : false) ||
        (item.supplier ? normalizeKana(item.supplier).toLowerCase().includes(normalizedTerm) : false)
    );
  }, [items, searchTerm]);

  // アラートサマリー(低在庫/欠品件数)はページ送りに左右されないクリニック全体(カテゴリ内)の件数でなければ
  // ならない。現在ページの items から集計すると、ページを送るたびに件数が変動する偽の集計になる
  // (BUG-412のページネーション化に伴い新規混入しうる回帰)。status=low / out_of_stock を明示指定し、
  // total(サーバー実COUNT)のみを参照する軽量クエリ(limit:1, data未使用)で代替する。
  const {
    data: lowStockPage,
    isError: isLowStockError,
  } = useGetInventoryItemsPage({
    category: categoryParam,
    status: "low",
    page: 1,
    limit: 1,
  });
  const {
    data: outOfStockPage,
    isError: isOutOfStockError,
  } = useGetInventoryItemsPage({
    category: categoryParam,
    status: "out_of_stock",
    page: 1,
    limit: 1,
  });

  // code-reviewer指摘(HIGH): 集計クエリが失敗した場合に `?? 0` だけで握り潰すと
  // 「異常なし」であるかのように在庫アラートが黙って非表示になる(サイレント障害)。
  // isError を summary に含め、呼び出し側で明示的なエラー表示に分岐させる。
  const summary = useMemo(
    () => ({
      total: pageResult?.total ?? 0,
      lowStock: lowStockPage?.total ?? 0,
      outOfStock: outOfStockPage?.total ?? 0,
      isError: isLowStockError || isOutOfStockError,
    }),
    [pageResult, lowStockPage, outOfStockPage, isLowStockError, isOutOfStockError]
  );

  return {
    data: filteredItems,
    summary,
    total: pageResult?.total ?? 0,
    page: pageResult?.page ?? page,
    limit: pageResult?.limit ?? limit,
    isLoading,
    isError,
  };
}
