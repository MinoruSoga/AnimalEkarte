import { beforeEach, describe, expect, it, vi } from "vitest";
import { QueryClient } from "@tanstack/react-query";

import { axios } from "@/lib/axios";
import { queryKeys } from "@/lib/query-keys";
import {
  createCheckupTypeField,
  deleteCheckupTypeField,
  invalidateCheckupTypeFieldQueries,
  reorderCheckupTypeFields,
  toCheckupTypeFieldEntry,
  updateCheckupTypeField,
} from "./checkup-type-fields";
import type { CheckupTypeFieldRow } from "@/hooks/use-checkup-fields";

vi.mock("@/lib/axios", () => ({
  axios: {
    post: vi.fn(),
    patch: vi.fn(),
    delete: vi.fn(),
  },
}));

// EMR-225: フィールド定義 write 側 API の固定。
// GET（一覧）は use-checkup-fields.ts が正本のためここでは対象外。

const fieldRow: CheckupTypeFieldRow = {
  id: 31,
  checkupTypeId: 7,
  name: "体重",
  fieldType: "number",
  unit: "kg",
  minValue: 1,
  maxValue: 80,
  options: [],
  isProvisional: false,
  sortOrder: 1,
};

describe("checkup type field master API", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("converts numeric row ids to the string-keyed editor entry", () => {
    expect(toCheckupTypeFieldEntry(fieldRow)).toEqual({
      ...fieldRow,
      id: "31",
      checkupTypeId: "7",
    });
  });

  it("uses the fixed nested CRUD and reorder endpoints", async () => {
    vi.mocked(axios.post).mockResolvedValue({
      data: {
        id: 31,
        checkup_type_id: 7,
        name: "体重",
        field_type: "number",
        unit: "kg",
        min_value: 1,
        max_value: 80,
        options: [],
        is_provisional: false,
        sort_order: 1,
      },
    });
    vi.mocked(axios.patch).mockResolvedValue({
      data: {
        id: 31,
        checkup_type_id: 7,
        name: "体重",
        field_type: "number",
        unit: "kg",
        options: null,
        is_provisional: false,
        sort_order: 1,
      },
    });
    vi.mocked(axios.delete).mockResolvedValue({ data: undefined });

    const created = await createCheckupTypeField("7", {
      name: "体重",
      field_type: "number",
      unit: "kg",
      min_value: 1,
      max_value: 80,
      sort_order: 1,
    });
    const updated = await updateCheckupTypeField("7", "31", {
      field_type: "text",
      clear_min_value: true,
      clear_max_value: true,
      options: [],
    });
    await deleteCheckupTypeField("7", "31");
    await reorderCheckupTypeFields("7", [31, 32]);

    expect(axios.post).toHaveBeenCalledWith("/v1/masters/checkup-types/7/fields", {
      name: "体重",
      field_type: "number",
      unit: "kg",
      min_value: 1,
      max_value: 80,
      sort_order: 1,
    });
    // PATCH は clear_* フラグと options 空配列（残存選択肢の明示クリア）をそのまま送る。
    expect(axios.patch).toHaveBeenCalledWith("/v1/masters/checkup-types/7/fields/31", {
      field_type: "text",
      clear_min_value: true,
      clear_max_value: true,
      options: [],
    });
    expect(axios.delete).toHaveBeenCalledWith("/v1/masters/checkup-types/7/fields/31");
    expect(axios.patch).toHaveBeenCalledWith("/v1/masters/checkup-types/7/fields/reorder", {
      ids: [31, 32],
    });
    // レスポンスは共有 transform と同じ写像で Row 形に戻る。
    expect(created).toEqual(fieldRow);
    expect(updated.options).toEqual([]);
    expect(updated.minValue).toBeUndefined();
  });

  it("invalidates only the shared typeFields query key", async () => {
    const queryClient = new QueryClient();
    const invalidate = vi.spyOn(queryClient, "invalidateQueries").mockResolvedValue(undefined);

    await invalidateCheckupTypeFieldQueries(queryClient, "7");

    expect(invalidate).toHaveBeenCalledTimes(1);
    expect(invalidate).toHaveBeenCalledWith({
      queryKey: queryKeys.checkups.typeFields("7"),
    });
    // カルテ画面の動的フォームと同じ key（"checkup-type-fields"）を撃つ。
    expect(queryKeys.checkups.typeFields("7")).toEqual(["checkup-type-fields", "7"]);
  });
});
