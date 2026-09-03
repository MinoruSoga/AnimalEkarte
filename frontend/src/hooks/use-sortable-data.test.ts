import { describe, it, expect } from "vitest";
import { renderHook, act } from "@testing-library/react";
import { useSortableData } from "./use-sortable-data";

interface Item {
  id: number;
  name: string;
  age: number;
}

const items: Item[] = [
  { id: 1, name: "田中", age: 30 },
  { id: 2, name: "佐藤", age: 25 },
  { id: 3, name: "鈴木", age: 35 },
];

describe("useSortableData", () => {
  describe("初期状態", () => {
    it("ソートなしのとき元の順序を維持する", () => {
      const { result } = renderHook(() => useSortableData(items));
      expect(result.current.sortedData.map((i) => i.id)).toEqual([1, 2, 3]);
    });

    it("activeSorts は空配列で始まる", () => {
      const { result } = renderHook(() => useSortableData(items));
      expect(result.current.activeSorts).toHaveLength(0);
    });
  });

  describe("toggleSort", () => {
    it("初回 toggle で昇順ソートになる", () => {
      const { result } = renderHook(() => useSortableData(items));
      act(() => result.current.toggleSort("name"));
      // 日本語ロケール: localeCompare("ja") による実際の順序
      expect(result.current.sortedData.map((i) => i.name)).toEqual(["佐藤", "田中", "鈴木"]);
    });

    it("2回 toggle で降順ソートになる", () => {
      const { result } = renderHook(() => useSortableData(items));
      act(() => result.current.toggleSort("name"));
      act(() => result.current.toggleSort("name"));
      expect(result.current.sortedData.map((i) => i.name)).toEqual(["鈴木", "田中", "佐藤"]);
    });

    it("3回 toggle でソート解除になる（元の順序）", () => {
      const { result } = renderHook(() => useSortableData(items));
      act(() => result.current.toggleSort("name"));
      act(() => result.current.toggleSort("name"));
      act(() => result.current.toggleSort("name"));
      expect(result.current.sortedData.map((i) => i.id)).toEqual([1, 2, 3]);
      expect(result.current.activeSorts).toHaveLength(0);
    });
  });

  describe("数値ソート（numericKeys）", () => {
    it("numericKeys 指定で数値昇順ソートが機能する", () => {
      const { result } = renderHook(() => useSortableData(items, { numericKeys: ["age"] }));
      act(() => result.current.toggleSort("age"));
      expect(result.current.sortedData.map((i) => i.age)).toEqual([25, 30, 35]);
    });

    it("numericKeys 指定で数値降順ソートが機能する", () => {
      const { result } = renderHook(() => useSortableData(items, { numericKeys: ["age"] }));
      act(() => result.current.toggleSort("age"));
      act(() => result.current.toggleSort("age"));
      expect(result.current.sortedData.map((i) => i.age)).toEqual([35, 30, 25]);
    });
  });

  describe("getSortValue（カスタムソート）", () => {
    it("getSortValue を使ったカスタム派生値でソートできる", () => {
      const { result } = renderHook(() =>
        useSortableData(items, {
          getSortValue: (item, key) => {
            if (key === "name_len") return item.name.length;
            return "";
          },
        }),
      );
      act(() => result.current.toggleSort("name_len"));
      // 佐藤(2) = 田中(2) = 鈴木(2) — 同じ長さなので順序維持
      expect(result.current.activeSorts).toHaveLength(1);
    });
  });

  describe("directionFor", () => {
    it("未ソートキーは 'none' を返す", () => {
      const { result } = renderHook(() => useSortableData(items));
      expect(result.current.directionFor("name")).toBe("none");
    });

    it("昇順ソート中は 'ascending' を返す", () => {
      const { result } = renderHook(() => useSortableData(items));
      act(() => result.current.toggleSort("name"));
      expect(result.current.directionFor("name")).toBe("ascending");
    });

    it("降順ソート中は 'descending' を返す", () => {
      const { result } = renderHook(() => useSortableData(items));
      act(() => result.current.toggleSort("name"));
      act(() => result.current.toggleSort("name"));
      expect(result.current.directionFor("name")).toBe("descending");
    });
  });

  describe("データ更新への対応", () => {
    it("ソート中に入力データが変わっても正しくソートされる", () => {
      let data = items;
      const { result, rerender } = renderHook(() => useSortableData(data));
      act(() => result.current.toggleSort("age"));

      data = [...items, { id: 4, name: "高橋", age: 20 }];
      rerender();
      // 昇順: 20 < 25 < 30 < 35
      expect(result.current.sortedData.map((i) => i.age)).toEqual([20, 25, 30, 35]);
    });
  });
});
