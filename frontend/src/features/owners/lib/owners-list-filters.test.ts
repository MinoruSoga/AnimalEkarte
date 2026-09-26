import { describe, expect, it } from "vitest";

import type { ActiveFilter } from "@/components/shared/PropertyFilter/types";
import {
  activeFiltersToParams,
  buildOwnerFilterProperties,
  paramsToActiveFilters,
} from "./owners-list-filters";

// EMR-197-01: 健診受診履歴フィルタ — GET /v1/pets?checkup_history= の
// enum (within_1y|within_2y|within_3y|not_within_1y|not_within_2y|not_within_3y|none) を
// /owners の動的フィルタとして公開し、URL パラメータと双方向に同期する。
describe("buildOwnerFilterProperties — 健診受診履歴", () => {
  it("condition=is のみの select プロパティとして 7 enum 選択肢を公開する", () => {
    const prop = buildOwnerFilterProperties([]).find((p) => p.key === "checkup_history");

    expect(prop).toBeDefined();
    expect(prop?.label).toBe("健診受診履歴");
    expect(prop?.type).toBe("select");
    expect(prop?.conditions).toEqual(["is"]);
    expect(prop?.options?.map((o) => o.value)).toEqual([
      "within_1y",
      "within_2y",
      "within_3y",
      "not_within_1y",
      "not_within_2y",
      "not_within_3y",
      "none",
    ]);
    // 各選択肢は日本語ラベルを持つ（生の enum 値をそのまま表示しない）
    for (const option of prop?.options ?? []) {
      expect(option.label).not.toBe(option.value);
      expect(option.label.length).toBeGreaterThan(0);
    }
  });
});

describe("activeFiltersToParams — checkup_history", () => {
  it("condition=is の checkup_history を URL パラメータへ変換する", () => {
    const filters: ActiveFilter[] = [
      {
        key: "checkup_history",
        condition: "is",
        value: "within_2y",
        displayValue: "2年以内に受診",
      },
    ];

    expect(activeFiltersToParams(filters)).toEqual({ checkup_history: "within_2y" });
  });

  it("is_not 等の非対応 condition は checkup_history へ変換しない（黙って is に倒さない）", () => {
    const filters: ActiveFilter[] = [
      { key: "checkup_history", condition: "is_not", value: "within_2y" },
      { key: "checkup_history", condition: "is", value: "" },
    ];

    expect(activeFiltersToParams(filters)).toEqual({});
  });

  it("species/include_deceased との併存を保つ", () => {
    const filters: ActiveFilter[] = [
      { key: "species", condition: "is", value: "1", displayValue: "犬" },
      {
        key: "include_deceased",
        condition: "is",
        value: "true",
        displayValue: "死亡ペットも含める",
      },
      {
        key: "checkup_history",
        condition: "is",
        value: "none",
        displayValue: "健診受診履歴なし",
      },
    ];

    expect(activeFiltersToParams(filters)).toEqual({
      species: "1",
      include_deceased: "true",
      checkup_history: "none",
    });
  });
});

describe("paramsToActiveFilters — checkup_history", () => {
  it("URL の checkup_history を表示ラベル付き chip として復元する", () => {
    const filters = paramsToActiveFilters(new URLSearchParams("checkup_history=within_1y"), []);

    expect(filters).toEqual([
      {
        key: "checkup_history",
        condition: "is",
        value: "within_1y",
        displayValue: "1年以内に受診",
      },
    ]);
  });

  it("enum 外の値は chip として復元しない（サーバが 400 にする値を UI で正当化しない）", () => {
    const filters = paramsToActiveFilters(new URLSearchParams("checkup_history=within_4y"), []);

    expect(filters).toEqual([]);
  });

  it("未指定では chip を作らない", () => {
    const filters = paramsToActiveFilters(new URLSearchParams(""), []);

    expect(filters).toEqual([]);
  });
});
