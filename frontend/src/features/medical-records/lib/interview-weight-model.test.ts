import { describe, expect, it } from "vitest";

import type { InterviewHistoryItem } from "../types";
import { collectPastWeightMeasurements } from "./interview-weight-model";

// EMR-356 repair1 C2: weightMeasurements 投影の集約は「未提供(undefined)/
// authoritative 空([])/値あり」を区別しなければならない。
// ClinicalTabs の `item.weightMeasurements ?? []` 逐次 flatten は、
// 全行未提供の legacy 投影を [] へ潰して「測定なし」と誤表示させるため RED。

const baseItem = {
  date: "2026-01-01",
  type: "診察",
  title: "カルテ",
  content: "本文",
  author: "田中",
};

const measurementA = {
  id: "11",
  value: 4.2,
  unit: "Kg" as const,
  measuredAt: "2026-03-01T09:30:00+09:00",
};
const measurementB = {
  id: "12",
  value: 4300,
  unit: "g" as const,
  measuredAt: "2026-03-01T15:00:00+09:00",
};

describe("collectPastWeightMeasurements（C2, EMR-356 repair1）", () => {
  it("全行が投影未提供（フィールド欠落）なら undefined を返す — unavailable ≠ 測定なし", () => {
    const items: InterviewHistoryItem[] = [
      { ...baseItem, id: "1" },
      { ...baseItem, id: "2" },
    ];

    expect(collectPastWeightMeasurements(items)).toEqual({ measurements: [], unavailable: true });
  });

  it("既知の空と未提供の医療行が混在すると履歴を不完全として保持する", () => {
    const items: InterviewHistoryItem[] = [
      { ...baseItem, id: "1" }, // 未提供
      { ...baseItem, id: "2", weightMeasurements: [] },
    ];

    expect(collectPastWeightMeasurements(items)).toEqual({ measurements: [], unavailable: true });
  });

  it("行ローカル測定を行順に flatten し、同一 id の重複は1件に畳む", () => {
    const items: InterviewHistoryItem[] = [
      { ...baseItem, id: "1", weightMeasurements: [measurementA] },
      { ...baseItem, id: "2" }, // 医療行の未提供は不完全な履歴
      { ...baseItem, id: "3", weightMeasurements: [measurementA, measurementB] },
    ];

    expect(collectPastWeightMeasurements(items)).toEqual({
      measurements: [measurementA, measurementB],
      unavailable: true,
    });
  });

  it("投影フィールドを持たない行（trimming 等）は集約結果を変えない", () => {
    const items: InterviewHistoryItem[] = [
      { ...baseItem, id: "1", weightMeasurements: [] },
      {
        ...baseItem,
        id: "trimming-7",
        type: "トリミング（完了）",
        href: "/trimming/7",
      },
    ];

    expect(collectPastWeightMeasurements(items)).toEqual({ measurements: [], unavailable: false });
  });

  it("正常取得した空履歴に未提供の医療行はない（取得状態は呼出側で保持する）", () => {
    expect(collectPastWeightMeasurements([])).toEqual({ measurements: [], unavailable: false });
  });
});
