import { describe, expect, it } from "vitest";

import {
  CLINICAL_HISTORY_KINDS,
  buildClinicalHistoryMatrix,
  selectAppointmentBriefing,
} from "./clinical-briefing";

describe("clinical briefing", () => {
  it("履歴を6種類へ分け、全診療日を新しい順の横列にして同日の関連記録を揃える", () => {
    const matrix = buildClinicalHistoryMatrix({
      medicalRecords: [
        {
          id: "mr-1",
          date: "2026/07/12",
          chiefComplaint: "外耳炎",
          doctor: "佐藤",
          assessment: "球菌を確認",
          plan: "2週間後に再評価",
        },
        { id: "mr-2", date: "2026/03/02", chiefComplaint: "右後肢の疼痛" },
      ],
      examinations: [
        {
          id: "exam-1",
          date: "2026-07-12",
          testType: "血液検査",
          status: "確定",
          items: [
            {
              id: "exam-item-1",
              name: "ALT",
              inspectionValue: "186",
              result: "",
              unit: "U/L",
              referenceValue: "17-78",
              isAbnormal: true,
              status: "high",
            },
          ],
        },
      ],
      checkups: [
        {
          id: 1,
          checkupId: 10,
          date: "2026-06-01",
          checkupTypeName: "歯科健診",
          fieldName: "歯石",
          fieldType: "text",
          unit: "",
          valueText: "軽度",
          valueList: [],
          isAbnormal: false,
          status: "normal",
        },
      ],
      treatments: [
        {
          id: "medicine-1",
          date: "26/7/12",
          itemType: "medicine",
          name: "点耳薬",
          adminRoute: "外用",
          quantity: 1,
          medicalRecordId: "mr-1",
        },
        {
          id: "procedure-1",
          date: "26/7/12",
          itemType: "procedure",
          name: "耳洗浄",
          adminRoute: "",
          quantity: 1,
          medicalRecordId: "mr-1",
        },
      ],
      vaccinations: [
        {
          id: 20,
          date: "26/4/12",
          next: "27/4/12",
          nextDate: "2027-04-12",
          name: "狂犬病予防接種",
          vaccineId: 3,
          lot1: "",
          lot2: "",
          lot3: "",
          lot4: "",
          remarks: "",
        },
        {
          id: 21,
          date: "25/8/18",
          next: "26/8/18",
          nextDate: "2026-08-18",
          name: "犬6種混合ワクチン",
          vaccineId: 4,
          lot1: "",
          lot2: "",
          lot3: "",
          lot4: "",
          remarks: "",
        },
      ],
      trimmings: [
        {
          id: "trim-1",
          date: "2026-06-21",
          status: "完了",
          courseName: "シャンプー＆カット",
          staff: "鈴木",
        },
      ],
    });

    expect(matrix.rows.map((row) => row.kind)).toEqual(CLINICAL_HISTORY_KINDS);
    expect(matrix.columns.map((column) => column.dateKey)).toEqual([
      "2026-07-12",
      "2026-06-21",
      "2026-06-01",
      "2026-04-12",
      "2026-03-02",
      "2025-08-18",
    ]);

    const latestColumn = matrix.columns[0];
    expect(latestColumn.entries.map((entry) => entry.kind)).toEqual([
      "診療",
      "検査",
      "薬・処方",
      "処置",
    ]);
    expect(matrix.rows.find((row) => row.kind === "薬・処方")?.count).toBe(1);
    expect(matrix.rows.find((row) => row.kind === "予防接種")?.count).toBe(2);
  });

  it("今日の来院と、キャンセル等を除いた次回予約を入力配列を変えずに選ぶ", () => {
    const reservations = [
      {
        id: "future-cancelled",
        start: new Date(2026, 6, 24, 9, 0),
        end: new Date(2026, 6, 24, 9, 30),
        status: "cancelled",
      },
      {
        id: "next",
        start: new Date(2026, 6, 23, 11, 0),
        end: new Date(2026, 6, 23, 11, 30),
        status: "confirmed",
      },
      {
        id: "today",
        start: new Date(2026, 6, 21, 10, 30),
        end: new Date(2026, 6, 21, 11, 0),
        status: "checked_in",
      },
    ];
    const originalOrder = reservations.map((reservation) => reservation.id);

    const result = selectAppointmentBriefing(reservations, "2026-07-21");

    expect(result.today?.id).toBe("today");
    expect(result.next?.id).toBe("next");
    expect(reservations.map((reservation) => reservation.id)).toEqual(originalOrder);
  });
});
