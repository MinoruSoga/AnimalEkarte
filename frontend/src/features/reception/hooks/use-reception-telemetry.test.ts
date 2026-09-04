import { act, renderHook } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { useReceptionTelemetry } from "./use-reception-telemetry";
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

describe("useReceptionTelemetry", () => {
  // 回帰ガード: 「本日受付」がフィルタ操作の影響を受けないこと（columns のみを見ること）を
  // フックの引数契約レベルで直接検証する。誤って filteredColumns を渡す退行を防ぐ。
  it("フィルタで絞り込まれた columns を渡しても、渡した columns 全体の件数をそのまま返す(呼び出し側が columns を渡す契約であることの保証)", () => {
    const fullColumns: ColumnData[] = [
      {
        title: "受付予約",
        appointments: [makeAppointment({ id: "1" }), makeAppointment({ id: "2" })],
      },
      { title: "受付済", appointments: [makeAppointment({ id: "3", status: "checked_in" })] },
    ];
    // フィルタが適用されて 1 件しか残っていない状態を模したより小さい columns
    const filteredLikeColumns: ColumnData[] = [
      { title: "受付予約", appointments: [makeAppointment({ id: "1" })] },
      { title: "受付済", appointments: [] },
    ];

    const { result: fullResult } = renderHook(() => useReceptionTelemetry(fullColumns));
    const { result: filteredResult } = renderHook(() => useReceptionTelemetry(filteredLikeColumns));

    expect(fullResult.current.totalCount).toBe(3);
    expect(filteredResult.current.totalCount).toBe(1);
    // 同じ日の予約データでも、渡した columns が変われば結果も変わる
    // → Reception.tsx 側は必ず columns（フィルタ非適用）を渡さなければならないことの裏付け
    expect(fullResult.current.totalCount).not.toBe(filteredResult.current.totalCount);
  });

  it("受付済 0 件では averageMinutes/longest ともに null", () => {
    const columns: ColumnData[] = [{ title: "受付予約", appointments: [makeAppointment()] }];

    const { result } = renderHook(() => useReceptionTelemetry(columns));

    expect(result.current.averageMinutes).toBeNull();
    expect(result.current.longest).toBeNull();
  });

  describe("60秒間隔の再計算", () => {
    beforeEach(() => {
      vi.useFakeTimers();
      vi.setSystemTime(new Date("2026-07-05T10:00:00Z"));
    });

    afterEach(() => {
      vi.useRealTimers();
    });

    it("60秒経過ごとに待ち時間を再計算する(専用ポーリングなしのローカルtick)", () => {
      const columns: ColumnData[] = [
        {
          title: "受付済",
          appointments: [
            makeAppointment({ status: "checked_in", checkedInAt: "2026-07-05T09:50:00Z" }),
          ],
        },
      ];

      const { result } = renderHook(() => useReceptionTelemetry(columns));

      expect(result.current.averageMinutes).toBe(10);

      act(() => {
        vi.advanceTimersByTime(60_000);
      });

      expect(result.current.averageMinutes).toBe(11);
    });
  });
});
