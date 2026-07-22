import { describe, expect, it } from "vitest";

import {
  countLatestExaminationAbnormalResults,
  selectLatestOverdueVaccination,
  selectUpcomingVaccination,
} from "./report-summary";

describe("owner report summary", () => {
  it("最新の検査だけから異常項目数を算出し、入力配列を変更しない", () => {
    const examinations = [
      {
        date: "2026-04-01",
        items: [{ isAbnormal: true }, { isAbnormal: true }],
      },
      {
        date: "2026-06-01",
        items: [{ isAbnormal: true }, { isAbnormal: false }],
      },
    ];
    const originalOrder = examinations.map((item) => item.date);

    expect(countLatestExaminationAbnormalResults(examinations)).toBe(1);
    expect(examinations.map((item) => item.date)).toEqual(originalOrder);
  });

  it("未来の次回接種予定から最も近い1件を選び、過去日と不正日付を除外する", () => {
    const vaccinations = [
      { id: 1, nextDate: "2026-04-01" },
      { id: 2, nextDate: "invalid" },
      { id: 3, nextDate: "2027-08-01" },
      { id: 4, nextDate: "2027-05-01" },
    ];

    expect(selectUpcomingVaccination(vaccinations, "2026-07-21")?.id).toBe(4);
  });

  it("検査・次回接種予定が無い場合は安全な空値を返す", () => {
    expect(countLatestExaminationAbnormalResults([])).toBe(0);
    expect(selectUpcomingVaccination([], "2026-07-21")).toBeUndefined();
  });

  it("未来予定が無い場合に最も近い期限超過予定を選べる", () => {
    const vaccinations = [
      { id: 1, nextDate: "2025-04-01" },
      { id: 2, nextDate: "2026-06-01" },
      { id: 3, nextDate: "invalid" },
    ];

    expect(selectLatestOverdueVaccination(vaccinations, "2026-07-21")?.id).toBe(2);
  });
});
