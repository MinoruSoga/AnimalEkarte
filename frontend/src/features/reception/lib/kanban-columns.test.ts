import { describe, expect, it } from "vitest";

import type { ColumnData } from "@/types";
import type { ReceptionAppointment } from "../api/types";
import {
  captureCardSnapshot,
  cloneColumns,
  findAppointment,
  mergeCard,
  relocateCard,
  removeCard,
  restoreCardSnapshot,
} from "./kanban-columns";

function makeAppointment(
  overrides: Partial<ReceptionAppointment> & { id: string },
): ReceptionAppointment {
  return {
    time: "09:00",
    visitDate: "2026-06-01",
    end: new Date(2026, 5, 1, 9, 30, 0),
    ownerName: "山田",
    petType: "犬",
    petName: "ポチ",
    visitType: "再診",
    reservationType: "一般診察",
    reservationTypeId: "1",
    reservationCategory: "general",
    isDesignated: false,
    doctor: undefined,
    doctorId: "",
    petId: "10",
    ownerId: "20",
    status: "checked_in",
    notes: undefined,
    source: "manual",
    ...overrides,
  };
}

function makeColumns(): ColumnData[] {
  return [
    { title: "受付予約", appointments: [makeAppointment({ id: "a1" })] },
    {
      title: "受付済",
      appointments: [makeAppointment({ id: "a2" }), makeAppointment({ id: "a3" })],
    },
  ];
}

describe("cloneColumns", () => {
  it("元の columns と appointments 配列を独立させた浅いクローンを返す", () => {
    const original = makeColumns();
    const cloned = cloneColumns(original);

    expect(cloned).toEqual(original);
    expect(cloned).not.toBe(original);
    expect(cloned[0].appointments).not.toBe(original[0].appointments);
  });
});

describe("removeCard", () => {
  it("対象カードを除外した新しい columns を返す", () => {
    const columns = makeColumns();
    const result = removeCard(columns, "a2");

    expect(result?.find((c) => c.title === "受付済")?.appointments.map((a) => a.id)).toEqual([
      "a3",
    ]);
    // 元の columns は変更しない（イミュータブル）
    expect(columns.find((c) => c.title === "受付済")?.appointments.map((a) => a.id)).toEqual([
      "a2",
      "a3",
    ]);
  });

  it("対象カードが存在しなければ null を返す", () => {
    const columns = makeColumns();
    expect(removeCard(columns, "ghost")).toBeNull();
  });
});

describe("mergeCard", () => {
  it("対象カードのフィールドをマージした新しい columns を返す", () => {
    const columns = makeColumns();
    const updated = makeAppointment({ id: "a2", petName: "タマ" });
    const result = mergeCard(columns, updated);

    const merged = result
      ?.find((c) => c.title === "受付済")
      ?.appointments.find((a) => a.id === "a2");
    expect(merged?.petName).toBe("タマ");
  });

  it("対象カードが存在しなければ null を返す", () => {
    const columns = makeColumns();
    expect(mergeCard(columns, makeAppointment({ id: "ghost" }))).toBeNull();
  });
});

describe("relocateCard", () => {
  it("source から target へカードを移動する（挿入位置省略時は末尾）", () => {
    const columns = makeColumns();
    const result = relocateCard(columns, "a1", "受付予約", "受付済");

    expect(result?.find((c) => c.title === "受付予約")?.appointments).toHaveLength(0);
    expect(result?.find((c) => c.title === "受付済")?.appointments.map((a) => a.id)).toEqual([
      "a2",
      "a3",
      "a1",
    ]);
  });

  it("resolveInsertIndex で指定した位置へ挿入する", () => {
    const columns = makeColumns();
    const result = relocateCard(columns, "a1", "受付予約", "受付済", () => 1);

    expect(result?.find((c) => c.title === "受付済")?.appointments.map((a) => a.id)).toEqual([
      "a2",
      "a1",
      "a3",
    ]);
  });

  it("同一カラム内での移動も成立する", () => {
    const columns = makeColumns();
    const result = relocateCard(columns, "a3", "受付済", "受付済", () => 0);

    expect(result?.find((c) => c.title === "受付済")?.appointments.map((a) => a.id)).toEqual([
      "a3",
      "a2",
    ]);
  });

  it("source/target カラムが存在しなければ null を返す", () => {
    const columns = makeColumns();
    expect(relocateCard(columns, "a1", "存在しない", "受付済")).toBeNull();
  });

  it("カードが source カラムに存在しなければ null を返す", () => {
    const columns = makeColumns();
    expect(relocateCard(columns, "ghost", "受付予約", "受付済")).toBeNull();
  });
});

describe("findAppointment", () => {
  it("全カラムから id に一致する appointment を探す", () => {
    const columns = makeColumns();
    expect(findAppointment(columns, "a3")?.id).toBe("a3");
  });

  it("見つからなければ undefined を返す", () => {
    const columns = makeColumns();
    expect(findAppointment(columns, "ghost")).toBeUndefined();
  });
});

describe("captureCardSnapshot / restoreCardSnapshot", () => {
  it("snapshot を取得し、他カードを動かした後でも対象カードだけを元の位置へ復元する", () => {
    const columns = makeColumns();
    const snapshot = captureCardSnapshot(columns, "a2");
    expect(snapshot).toEqual({
      appointment: makeAppointment({ id: "a2" }),
      columnTitle: "受付済",
      index: 0,
    });

    // a2 を移動し、a1 のフィールドも変更された状態を模す
    const moved = relocateCard(columns, "a2", "受付済", "受付予約") ?? columns;
    const withOtherChange =
      mergeCard(moved, makeAppointment({ id: "a1", petName: "別名" })) ?? moved;

    const restored = restoreCardSnapshot(withOtherChange, snapshot!);

    expect(restored.find((c) => c.title === "受付済")?.appointments.map((a) => a.id)).toEqual([
      "a2",
      "a3",
    ]);
    // 他カードの変更は保持される
    expect(
      restored.find((c) => c.title === "受付予約")?.appointments.find((a) => a.id === "a1")
        ?.petName,
    ).toBe("別名");
  });

  it("captureCardSnapshot は対象が存在しなければ null を返す", () => {
    const columns = makeColumns();
    expect(captureCardSnapshot(columns, "ghost")).toBeNull();
  });
});
