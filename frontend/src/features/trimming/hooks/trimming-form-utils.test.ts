import { describe, it, expect } from "vitest";
import type { TrimmingFormData } from "@/types/trimming";
import {
  buildCreateTrimmingRequest,
  buildUpdateTrimmingRequest,
  DEFAULT_TRIMMING_FORM_DATA,
  defaultRecordShortcutTimes,
  filterActiveOrSelectedMasterItems,
  findDefaultTrimmingReservationTypeId,
} from "./trimming-form-utils";

const STYLE_KEY = "trimming-images/1/style.jpg";
const COMPLETED_KEY = "trimming-images/1/done.jpg";

function trimmingForm(overrides: Partial<TrimmingFormData> = {}): TrimmingFormData {
  return { ...DEFAULT_TRIMMING_FORM_DATA, ...overrides };
}

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
    const items = [{ id: "1", name: "スタンダードコース", status: "active" }];

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

describe("trimming image request fields", () => {
  it("create はオブジェクトキーを style_image / completed_image として送る", () => {
    const req = buildCreateTrimmingRequest(
      trimmingForm({ styleImage: STYLE_KEY, completedImage: COMPLETED_KEY }),
      10,
      9,
      undefined,
      undefined,
    );

    expect(req.style_image).toBe(STYLE_KEY);
    expect(req.completed_image).toBe(COMPLETED_KEY);
  });

  it("未選択の画像フィールドは create / update とも省略する", () => {
    const createReq = buildCreateTrimmingRequest(trimmingForm(), 10, 9, undefined, undefined);
    const updateReq = buildUpdateTrimmingRequest(trimmingForm());

    expect(createReq).not.toHaveProperty("style_image");
    expect(createReq).not.toHaveProperty("completed_image");
    expect(updateReq).not.toHaveProperty("style_image");
    expect(updateReq).not.toHaveProperty("completed_image");
  });

  it("明示削除は空文字を送り、触っていない側は省略する", () => {
    const updateReq = buildUpdateTrimmingRequest(
      trimmingForm({ styleImage: "", completedImage: null }),
    );
    const createReq = buildCreateTrimmingRequest(
      trimmingForm({ styleImage: null, completedImage: "" }),
      10,
      9,
      undefined,
      undefined,
    );

    expect(updateReq.style_image).toBe("");
    expect(updateReq).not.toHaveProperty("completed_image");
    expect(createReq).not.toHaveProperty("style_image");
    expect(createReq.completed_image).toBe("");
  });
});
