import { describe, expect, it } from "vitest";

import type { InterviewHistoryItem } from "../types";
import {
  collectPastWeightMeasurements,
  resolvePastWeightReferenceDate,
} from "./interview-weight-model";

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

// EMR356-1: 「過去」は現在カルテの時点より厳密に前の測定のみ。
// 既存 fixture（2026-03-01 測定）が境界をまたがないよう翌日 00:00 JST を参照にする。
const REFERENCE = new Date("2026-03-02T00:00:00+09:00");

describe("collectPastWeightMeasurements（C2, EMR-356 repair1）", () => {
  it("全行が投影未提供（フィールド欠落）なら undefined を返す — unavailable ≠ 測定なし", () => {
    const items: InterviewHistoryItem[] = [
      { ...baseItem, id: "1" },
      { ...baseItem, id: "2" },
    ];

    expect(collectPastWeightMeasurements(items, REFERENCE)).toEqual({
      measurements: [],
      unavailable: true,
    });
  });

  it("既知の空と未提供の医療行が混在すると履歴を不完全として保持する", () => {
    const items: InterviewHistoryItem[] = [
      { ...baseItem, id: "1" }, // 未提供
      { ...baseItem, id: "2", weightMeasurements: [] },
    ];

    expect(collectPastWeightMeasurements(items, REFERENCE)).toEqual({
      measurements: [],
      unavailable: true,
    });
  });

  it("行ローカル測定を行順に flatten し、同一 id の重複は1件に畳む", () => {
    const items: InterviewHistoryItem[] = [
      { ...baseItem, id: "1", weightMeasurements: [measurementA] },
      { ...baseItem, id: "2" }, // 医療行の未提供は不完全な履歴
      { ...baseItem, id: "3", weightMeasurements: [measurementA, measurementB] },
    ];

    expect(collectPastWeightMeasurements(items, REFERENCE)).toEqual({
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

    expect(collectPastWeightMeasurements(items, REFERENCE)).toEqual({
      measurements: [],
      unavailable: false,
    });
  });

  it("正常取得した空履歴に未提供の医療行はない（取得状態は呼出側で保持する）", () => {
    expect(collectPastWeightMeasurements([], REFERENCE)).toEqual({
      measurements: [],
      unavailable: false,
    });
  });
});

describe("collectPastWeightMeasurements — 時系列境界（EMR356-1 review-fix）", () => {
  const later = {
    id: "11",
    value: 4.2,
    unit: "Kg" as const,
    measuredAt: "2026-03-05T09:30:00+09:00",
  };
  const earlier = {
    id: "12",
    value: 4.0,
    unit: "Kg" as const,
    measuredAt: "2026-02-20T10:00:00+09:00",
  };
  const boundary = new Date("2026-03-01T00:00:00+09:00");

  it("基準時刻以後の測定は「過去」に含めない（古いカルテで後日の体重を出さない）", () => {
    const items: InterviewHistoryItem[] = [
      { ...baseItem, id: "1", weightMeasurements: [later, earlier] },
    ];

    expect(collectPastWeightMeasurements(items, boundary)).toEqual({
      measurements: [earlier],
      unavailable: false,
    });
  });

  it("基準時刻と同一時刻の測定は含めない（strictly before）", () => {
    const items: InterviewHistoryItem[] = [
      {
        ...baseItem,
        id: "1",
        weightMeasurements: [
          { id: "13", value: 4.1, unit: "Kg" as const, measuredAt: "2026-03-01T00:00:00+09:00" },
        ],
      },
    ];

    expect(collectPastWeightMeasurements(items, boundary)).toEqual({
      measurements: [],
      unavailable: false,
    });
  });

  it("選択カルテ当日の測定は日付境界より前ではないため「過去」に含めない", () => {
    const items: InterviewHistoryItem[] = [
      {
        ...baseItem,
        id: "1",
        weightMeasurements: [
          { id: "14", value: 3.9, unit: "Kg" as const, measuredAt: "2026-03-01T09:30:00+09:00" },
        ],
      },
    ];

    expect(collectPastWeightMeasurements(items, new Date("2026-03-01T00:00:00+09:00"))).toEqual({
      measurements: [],
      unavailable: false,
    });
  });

  it("境界未確定（undefined）では測定を落とし unavailable（時系列不明で過去断定しない）", () => {
    const items: InterviewHistoryItem[] = [{ ...baseItem, id: "1", weightMeasurements: [earlier] }];

    expect(collectPastWeightMeasurements(items, undefined)).toEqual({
      measurements: [],
      unavailable: true,
    });
  });

  it("measuredAt が不正な測定は落として unavailable（配置不能を空と誤表示しない）", () => {
    const items: InterviewHistoryItem[] = [
      {
        ...baseItem,
        id: "1",
        weightMeasurements: [
          { id: "15", value: 4.4, unit: "Kg" as const, measuredAt: "not-a-date" },
          earlier,
        ],
      },
    ];

    expect(collectPastWeightMeasurements(items, boundary)).toEqual({
      measurements: [earlier],
      unavailable: true,
    });
  });
});

describe("resolvePastWeightReferenceDate（EMR356-1 境界解決）", () => {
  it("PostgreSQL date の UTC シリアライズを訪問時刻とみなさず、同日08:30JSTを過去から除外する", () => {
    const selectedBoundary = resolvePastWeightReferenceDate({
      recordDateIso: "2026-03-01T00:00:00Z",
    });
    expect(selectedBoundary).toEqual(new Date("2026-03-01T00:00:00+09:00"));
    expect(
      collectPastWeightMeasurements(
        [
          {
            ...baseItem,
            id: "older-record",
            weightMeasurements: [{ ...measurementA, measuredAt: "2026-03-01T08:30:00+09:00" }],
          },
        ],
        selectedBoundary,
      ),
    ).toEqual({ measurements: [], unavailable: false });
  });
  it("wire record.date の時刻部分を訪問時刻とみなさずJST日付開始へ正規化する", () => {
    expect(resolvePastWeightReferenceDate({ recordDateIso: "2026-03-01T14:00:00+09:00" })).toEqual(
      new Date("2026-03-01T00:00:00+09:00"),
    );
  });

  it("wire カルテ日付が form の日付より優先され、時刻は採用しない", () => {
    expect(
      resolvePastWeightReferenceDate({
        recordDateIso: "2026-03-01T14:00:00+09:00",
        recordDateJst: "2026-03-01",
      }),
    ).toEqual(new Date("2026-03-01T00:00:00+09:00"));
  });

  it("date-only（YYYY-MM-DD）は JST 00:00 起点にする", () => {
    expect(resolvePastWeightReferenceDate({ recordDateJst: "2026-03-01" })).toEqual(
      new Date("2026-03-01T00:00:00+09:00"),
    );
  });

  it("有効な入力が無ければ undefined（呼出側は unavailable として扱う）", () => {
    expect(resolvePastWeightReferenceDate({})).toBeUndefined();
    expect(resolvePastWeightReferenceDate({ recordDateIso: "garbage" })).toBeUndefined();
    expect(resolvePastWeightReferenceDate({ recordDateJst: "2026/03/01" })).toBeUndefined();
    expect(resolvePastWeightReferenceDate({ recordDateJst: "2026-02-30" })).toBeUndefined();
    expect(
      resolvePastWeightReferenceDate({ recordDateIso: "2026-03-01T12:00:00" }),
    ).toBeUndefined();
  });
});
