import { describe, it, expect } from "vitest";
import { computeReceptionTotalCount, computeCheckedInWaitStats } from "./reception-telemetry";
import type { ColumnData } from "@/types";
import type { ReceptionAppointment } from "../api/transforms";

function makeAppointment(overrides: Partial<ReceptionAppointment> = {}): ReceptionAppointment {
  return {
    id: "1",
    time: "10:00",
    visitDate: "2026-07-05",
    end: new Date(2026, 6, 5, 10, 30, 0),
    ownerName: "山田",
    petType: "犬",
    petName: "ポチ",
    visitType: "初診",
    reservationType: "一般診療",
    reservationTypeId: "1",
    reservationCategory: "general",
    isDesignated: false,
    doctorId: "",
    petId: "10",
    ownerId: "20",
    status: "pending",
    source: "manual",
    ...overrides,
  };
}

function makeColumns(byTitle: Record<string, ReceptionAppointment[]>): ColumnData[] {
  return Object.entries(byTitle).map(([title, appointments]) => ({ title, appointments }));
}

describe("computeReceptionTotalCount", () => {
  it("全カラムの appointments 件数を合算する", () => {
    const columns = makeColumns({
      受付予約: [makeAppointment({ id: "1" }), makeAppointment({ id: "2" })],
      受付済: [makeAppointment({ id: "3", status: "checked_in" })],
      会計済: [],
    });

    expect(computeReceptionTotalCount(columns)).toBe(3);
  });

  it("空の columns では 0 を返す", () => {
    expect(computeReceptionTotalCount([])).toBe(0);
  });

  it("カラム数・appointments 配列の中身に関わらず合計のみを見る(フィルタ済み配列を渡した場合の挙動と等価であることの確認)", () => {
    // フィルタ操作で appointments が絞り込まれた columns を渡せば、その絞り込み後の件数を返す。
    // つまり「常に columns（フィルタ非適用）を渡す」という呼び出し側の責務が本質であり、
    // この関数自体はどちらを渡されても素直に合算するだけ。呼び出し側の契約は
    // use-reception-telemetry.test.ts の回帰テストで保証する。
    const allColumns = makeColumns({ 受付予約: [makeAppointment(), makeAppointment()] });
    const filteredSubset = makeColumns({ 受付予約: [makeAppointment()] });

    expect(computeReceptionTotalCount(allColumns)).toBe(2);
    expect(computeReceptionTotalCount(filteredSubset)).toBe(1);
  });
});

describe("computeCheckedInWaitStats", () => {
  it("受付済(checked_in) 0 件では averageMinutes/longest ともに null を返す", () => {
    const columns = makeColumns({
      受付予約: [makeAppointment({ status: "pending" })],
      会計済: [makeAppointment({ status: "completed" })],
    });

    const result = computeCheckedInWaitStats(columns, new Date("2026-07-05T10:00:00Z"));

    expect(result.averageMinutes).toBeNull();
    expect(result.longest).toBeNull();
  });

  it("checkedInAt が未設定の checked_in 患者は集計対象外にする", () => {
    const columns = makeColumns({
      受付済: [makeAppointment({ status: "checked_in", checkedInAt: undefined })],
    });

    const result = computeCheckedInWaitStats(columns, new Date("2026-07-05T10:00:00Z"));

    expect(result.averageMinutes).toBeNull();
    expect(result.longest).toBeNull();
  });

  it("複数の受付済患者の平均待ち時間を算出する", () => {
    const now = new Date("2026-07-05T10:30:00Z");
    const columns = makeColumns({
      受付済: [
        makeAppointment({ id: "1", petName: "ポチ", status: "checked_in", checkedInAt: "2026-07-05T10:20:00Z" }), // 10分待ち
        makeAppointment({ id: "2", petName: "ミルク", status: "checked_in", checkedInAt: "2026-07-05T09:58:00Z" }), // 32分待ち
      ],
    });

    const result = computeCheckedInWaitStats(columns, now);

    expect(result.averageMinutes).toBe(21); // round((10+32)/2) = 21
    expect(result.longest).toEqual({ minutes: 32, petName: "ミルク" });
  });

  it("最長待ちは患者名つきで1名のみ返す", () => {
    const now = new Date("2026-07-05T10:30:00Z");
    const columns = makeColumns({
      受付済: [
        makeAppointment({ id: "1", petName: "ポチ", status: "checked_in", checkedInAt: "2026-07-05T10:25:00Z" }),
        makeAppointment({ id: "2", petName: "ミルク", status: "checked_in", checkedInAt: "2026-07-05T09:58:00Z" }),
        makeAppointment({ id: "3", petName: "タマ", status: "checked_in", checkedInAt: "2026-07-05T10:10:00Z" }),
      ],
    });

    const result = computeCheckedInWaitStats(columns, now);

    expect(result.longest).toEqual({ minutes: 32, petName: "ミルク" });
  });

  it("checked_in 以外のステータスは待ち時間集計から除外する", () => {
    const now = new Date("2026-07-05T10:30:00Z");
    const columns = makeColumns({
      診療中: [makeAppointment({ status: "in_consultation", checkedInAt: "2026-07-05T09:00:00Z" })],
      受付済: [makeAppointment({ status: "checked_in", checkedInAt: "2026-07-05T10:20:00Z" })],
    });

    const result = computeCheckedInWaitStats(columns, now);

    expect(result.averageMinutes).toBe(10);
  });
});
