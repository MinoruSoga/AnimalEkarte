import { describe, it, expect } from "vitest";

import { transformTrimming, type TrimmingUI } from "@/features/trimming";
import type { BackendTrimming } from "@/types/trimming";

import { selectCompletedTrimmingHistory } from "./get-pet-trimming-history";

// transformTrimming が要求する必須フィールドを埋める最小ファクトリ。
// status / start_time のみテストごとに上書きする。
function bt(overrides: Partial<BackendTrimming>): BackendTrimming {
  return {
    id: 1,
    clinic_id: 1,
    reservation_type_id: 1,
    start_time: "2026-06-10T10:00:00+09:00",
    end_time: "2026-06-10T11:00:00+09:00",
    status: "completed",
    source: "staff",
    has_detail: true,
    style_request: "",
    bw_unit: "Kg",
    used_shampoo: "",
    used_ribbon: "",
    remarks: "",
    style_image: "",
    completed_image: "",
    created_at: "2026-06-10T10:00:00+09:00",
    updated_at: "2026-06-10T10:00:00+09:00",
    options: [],
    ...overrides,
  };
}

describe("selectCompletedTrimmingHistory", () => {
  it("施術実施済み(完了)のみを残し、予約・キャンセル・進行中を除外する", () => {
    const items: TrimmingUI[] = [
      transformTrimming(bt({ id: 1, status: "completed", start_time: "2026-06-10T10:00:00+09:00" })),
      transformTrimming(bt({ id: 2, status: "cancelled", start_time: "2026-06-11T10:00:00+09:00" })),
      transformTrimming(bt({ id: 3, status: "confirmed", start_time: "2026-06-12T10:00:00+09:00" })),
      transformTrimming(bt({ id: 4, status: "in_consultation", start_time: "2026-06-13T10:00:00+09:00" })),
      transformTrimming(bt({ id: 5, status: "no_show", start_time: "2026-06-14T10:00:00+09:00" })),
    ];

    const out = selectCompletedTrimmingHistory(items);

    expect(out).toHaveLength(1);
    expect(out[0].id).toBe("1");
    expect(out[0].status).toBe("完了");
  });

  it("完了を実施日の降順に並べる", () => {
    const items: TrimmingUI[] = [
      transformTrimming(bt({ id: 1, status: "completed", start_time: "2026-02-01T10:00:00+09:00" })),
      transformTrimming(bt({ id: 2, status: "completed", start_time: "2026-05-20T10:00:00+09:00" })),
      transformTrimming(bt({ id: 3, status: "completed", start_time: "2026-03-15T10:00:00+09:00" })),
    ];

    const out = selectCompletedTrimmingHistory(items);

    expect(out.map((t) => t.id)).toEqual(["2", "3", "1"]);
  });

  it("完了が無ければ空配列を返す", () => {
    const items: TrimmingUI[] = [transformTrimming(bt({ id: 1, status: "cancelled" }))];
    expect(selectCompletedTrimmingHistory(items)).toEqual([]);
  });
});
