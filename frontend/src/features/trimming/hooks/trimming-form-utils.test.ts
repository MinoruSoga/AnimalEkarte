import { describe, it, expect } from "vitest";
import {
  defaultRecordShortcutTimes,
  filterActiveOrSelectedMasterItems,
  findDefaultTrimmingReservationTypeId,
} from "./trimming-form-utils";

describe("defaultRecordShortcutTimes", () => {
  it("uses unique JST clock time instead of a fixed 10:00 slot", () => {
    const now = new Date("2026-06-01T01:23:45.678Z");
    expect(defaultRecordShortcutTimes("2026-06-01", now)).toEqual({
      start: "2026-06-01T10:23:45.678+09:00",
      end: "2026-06-01T11:53:45.678+09:00",
    });
  });
});

describe("findDefaultTrimmingReservationTypeId", () => {
  it("最小sort_orderの公開トリミング種別を返し、入力順を変更しない", () => {
    const types = [
      { id: 3, category: "trimming", is_internal: false, sort_order: 30 },
      { id: 1, category: "medical", is_internal: false, sort_order: 1 },
      { id: 2, category: "trimming", is_internal: false, sort_order: 20 },
      { id: 4, category: "trimming", is_internal: true, sort_order: 10 },
    ];
    const originalOrder = types.map((type) => type.id);

    expect(findDefaultTrimmingReservationTypeId([{ types }])).toBe(2);
    expect(types.map((type) => type.id)).toEqual(originalOrder);
  });
});

describe("filterActiveOrSelectedMasterItems", () => {
  it("active なアイテムのみを返す（未選択の無効アイテムは除外）", () => {
    const items = [
      { id: "1", name: "スタンダードコース", status: "active" },
      { id: "2", name: "廃止コース", status: "inactive" },
    ];

    const result = filterActiveOrSelectedMasterItems(items, []);

    expect(result).toEqual([{ id: "1", name: "スタンダードコース", status: "active" }]);
  });

  it("選択済みの無効アイテムは名前に（無効）を付与して維持する", () => {
    const items = [
      { id: "1", name: "スタンダードコース", status: "active" },
      { id: "2", name: "廃止コース", status: "inactive" },
    ];

    const result = filterActiveOrSelectedMasterItems(items, ["2"]);

    expect(result).toEqual([
      { id: "1", name: "スタンダードコース", status: "active" },
      { id: "2", name: "廃止コース（無効）", status: "inactive" },
    ]);
  });

  it("選択済みでも active なアイテムには（無効）を付与しない（重複しない）", () => {
    const items = [
      { id: "1", name: "スタンダードコース", status: "active" },
    ];

    const result = filterActiveOrSelectedMasterItems(items, ["1"]);

    expect(result).toEqual([{ id: "1", name: "スタンダードコース", status: "active" }]);
  });

  it("選択IDがアイテム一覧に存在しない場合は無視する", () => {
    const items = [{ id: "1", name: "スタンダードコース", status: "active" }];

    const result = filterActiveOrSelectedMasterItems(items, ["999"]);

    expect(result).toEqual([{ id: "1", name: "スタンダードコース", status: "active" }]);
  });

  it("空文字の選択IDは無視する", () => {
    const items = [{ id: "1", name: "スタンダードコース", status: "active" }];

    const result = filterActiveOrSelectedMasterItems(items, [""]);

    expect(result).toEqual([{ id: "1", name: "スタンダードコース", status: "active" }]);
  });
});
