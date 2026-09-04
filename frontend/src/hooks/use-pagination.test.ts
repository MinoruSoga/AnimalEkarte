import { describe, it, expect } from "vitest";
import { renderHook, act } from "@testing-library/react";
import { usePagination } from "./use-pagination";

describe("usePagination", () => {
  const makeData = (n: number) => Array.from({ length: n }, (_, i) => ({ id: i + 1 }));

  describe("初期状態", () => {
    it("デフォルト pageSize 20 でページ 1 から開始する", () => {
      const { result } = renderHook(() => usePagination(makeData(50)));
      expect(result.current.currentPage).toBe(1);
      expect(result.current.pageSize).toBe(20);
      expect(result.current.totalPages).toBe(3);
    });

    it("カスタム pageSize を受け付ける", () => {
      const { result } = renderHook(() => usePagination(makeData(10), { pageSize: 5 }));
      expect(result.current.totalPages).toBe(2);
      expect(result.current.paginatedData).toHaveLength(5);
    });

    it("空配列のとき totalPages は 1 になる", () => {
      const { result } = renderHook(() => usePagination([]));
      expect(result.current.totalPages).toBe(1);
      expect(result.current.paginatedData).toHaveLength(0);
    });
  });

  describe("ページネーション", () => {
    it("nextPage で currentPage が増加する", () => {
      const { result } = renderHook(() => usePagination(makeData(50)));
      act(() => result.current.nextPage());
      expect(result.current.currentPage).toBe(2);
    });

    it("prevPage で currentPage が減少する", () => {
      const { result } = renderHook(() => usePagination(makeData(50)));
      act(() => result.current.nextPage());
      act(() => result.current.prevPage());
      expect(result.current.currentPage).toBe(1);
    });

    it("prevPage はページ 1 を下回らない", () => {
      const { result } = renderHook(() => usePagination(makeData(50)));
      act(() => result.current.prevPage());
      expect(result.current.currentPage).toBe(1);
    });

    it("nextPage は totalPages を超えない", () => {
      const { result } = renderHook(() => usePagination(makeData(50)));
      act(() => result.current.nextPage());
      act(() => result.current.nextPage());
      act(() => result.current.nextPage());
      expect(result.current.currentPage).toBe(3);
    });

    it("goToPage で指定ページに移動できる", () => {
      const { result } = renderHook(() => usePagination(makeData(50)));
      act(() => result.current.goToPage(3));
      expect(result.current.currentPage).toBe(3);
    });

    it("goToPage は範囲外のページをクランプする", () => {
      const { result } = renderHook(() => usePagination(makeData(50)));
      act(() => result.current.goToPage(100));
      expect(result.current.currentPage).toBe(3);
      act(() => result.current.goToPage(-5));
      expect(result.current.currentPage).toBe(1);
    });

    it("resetPage でページ 1 に戻る", () => {
      const { result } = renderHook(() => usePagination(makeData(50)));
      act(() => result.current.goToPage(3));
      act(() => result.current.resetPage());
      expect(result.current.currentPage).toBe(1);
    });
  });

  describe("paginatedData", () => {
    it("先頭ページは最初の 20 件を返す", () => {
      const data = makeData(50);
      const { result } = renderHook(() => usePagination(data));
      expect(result.current.paginatedData).toHaveLength(20);
      expect(result.current.paginatedData[0]).toEqual({ id: 1 });
      expect(result.current.paginatedData[19]).toEqual({ id: 20 });
    });

    it("2ページ目は 21〜40 件目を返す", () => {
      const data = makeData(50);
      const { result } = renderHook(() => usePagination(data));
      act(() => result.current.nextPage());
      expect(result.current.paginatedData[0]).toEqual({ id: 21 });
      expect(result.current.paginatedData[19]).toEqual({ id: 40 });
    });

    it("最終ページは端数件数を返す", () => {
      const data = makeData(45);
      const { result } = renderHook(() => usePagination(data));
      act(() => result.current.goToPage(3));
      expect(result.current.paginatedData).toHaveLength(5);
    });
  });

  describe("startIndex / endIndex", () => {
    it("先頭ページの startIndex=1, endIndex=20", () => {
      const { result } = renderHook(() => usePagination(makeData(50)));
      expect(result.current.startIndex).toBe(1);
      expect(result.current.endIndex).toBe(20);
    });

    it("空データのとき startIndex=0, endIndex=0", () => {
      const { result } = renderHook(() => usePagination([]));
      expect(result.current.startIndex).toBe(0);
      expect(result.current.endIndex).toBe(0);
    });
  });

  describe("resetKey による自動リセット", () => {
    it("resetKey が変わると currentPage が 1 にリセットされる", () => {
      let key = "a";
      const { result, rerender } = renderHook(() => usePagination(makeData(50), { resetKey: key }));
      act(() => result.current.goToPage(3));
      expect(result.current.currentPage).toBe(3);

      key = "b";
      rerender();
      expect(result.current.currentPage).toBe(1);
    });
  });

  describe("フィルタ後ページ超過の自動補正", () => {
    it("データが減少して totalPages を超えた場合は最終ページに自動補正する", () => {
      let data = makeData(50);
      const { result, rerender } = renderHook(() => usePagination(data));
      act(() => result.current.goToPage(3));
      expect(result.current.currentPage).toBe(3);

      // データを 10 件に縮小 → totalPages=1
      data = makeData(10);
      rerender();
      expect(result.current.currentPage).toBe(1);
    });
  });
});
