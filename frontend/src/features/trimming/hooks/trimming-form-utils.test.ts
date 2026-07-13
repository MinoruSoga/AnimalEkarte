import { describe, it, expect } from "vitest";
import { filterActiveOrSelectedMasterItems } from "./trimming-form-utils";

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
