import { describe, expect, it } from "vitest";

import { todayJSTISO } from "@/lib/jst-date";

import {
  deriveNextVisitState,
  nextVisitFetchEndDate,
  validateNextVisitIntentDate,
  type NextVisitSourceReservation,
} from "./next-visit";

// JST wall-clock の "now"（transformReservation の start と同じ規約の Date）
const NOW = new Date(2026, 9, 1, 9, 0); // 2026-10-01 09:00

function reservation(
  status: NextVisitSourceReservation["status"],
  start: Date,
  id = "r-1",
): NextVisitSourceReservation {
  return { id, status, start };
}

describe("deriveNextVisitState (EMR-308)", () => {
  it("未来の pending/confirmed 実予約は最も近いものを日時（YYYY-MM-DD HH:mm）で表示する", () => {
    const state = deriveNextVisitState(
      [
        reservation("confirmed", new Date(2026, 9, 12, 11, 0)),
        reservation("pending", new Date(2026, 9, 5, 10, 30)),
      ],
      NOW,
    );

    expect(state).toEqual({
      kind: "appointment",
      label: "2026-10-05 10:30",
      intentDate: "",
      tentativeId: null,
    });
  });

  it("開始時刻が過去の実予約は無視する", () => {
    const state = deriveNextVisitState(
      [reservation("confirmed", new Date(2026, 8, 20, 10, 0))],
      NOW,
    );

    expect(state).toEqual({ kind: "unset", label: "", intentDate: "", tentativeId: null });
  });

  it("pending/confirmed 以外（cancelled/completed 等）は実予約として扱わない", () => {
    const state = deriveNextVisitState(
      [
        reservation("cancelled", new Date(2026, 9, 5, 10, 30)),
        reservation("completed", new Date(2026, 9, 6, 10, 30)),
      ],
      NOW,
    );

    expect(state).toEqual({ kind: "unset", label: "", intentDate: "", tentativeId: null });
  });

  it("tentative のみの場合は意図日を日付（YYYY-MM-DD）のみで表示し tentativeId を返す", () => {
    const state = deriveNextVisitState(
      [reservation("tentative", new Date(2026, 9, 20, 0, 0), "t-1")],
      NOW,
    );

    expect(state).toEqual({
      kind: "tentative",
      label: "2026-10-20",
      intentDate: "2026-10-20",
      tentativeId: "t-1",
    });
  });

  it("実予約と tentative が両方あれば実予約優先。intentDate は tentative を保持する", () => {
    const state = deriveNextVisitState(
      [
        reservation("tentative", new Date(2026, 9, 20, 0, 0)),
        reservation("pending", new Date(2026, 9, 5, 10, 30)),
      ],
      NOW,
    );

    expect(state.kind).toBe("appointment");
    expect(state.label).toBe("2026-10-05 10:30");
    // Popover の日付入力には tentative の意図日を引き続き入れる
    expect(state.intentDate).toBe("2026-10-20");
  });

  it("tentative は intendedDate を sentinel start より優先する", () => {
    // intendedDate がある行は sentinel 導出より正本を使う
    // （reservations feature の tentativeIntendedDate と同一規約）。
    const state = deriveNextVisitState(
      [
        {
          id: "t-1",
          status: "tentative",
          start: new Date(2026, 9, 20, 0, 0),
          intendedDate: "2026-11-15",
        },
      ],
      NOW,
    );

    expect(state).toEqual({
      kind: "tentative",
      label: "2026-11-15",
      intentDate: "2026-11-15",
      tentativeId: "t-1",
    });
  });

  it("tentative で intendedDate 欠落時は sentinel start の JST 暦日にフォールバックする", () => {
    const state = deriveNextVisitState(
      [{ id: "t-1", status: "tentative", start: new Date(2026, 9, 20, 0, 0) }],
      NOW,
    );

    expect(state.intentDate).toBe("2026-10-20");
  });

  it("予約が無ければ unset", () => {
    expect(deriveNextVisitState([], NOW)).toEqual({
      kind: "unset",
      label: "",
      intentDate: "",
      tentativeId: null,
    });
  });
});

describe("validateNextVisitIntentDate", () => {
  it("空文字は有効（クリア操作は別経路で DELETE する）", () => {
    expect(validateNextVisitIntentDate("")).toBeNull();
  });

  it("過去日はエラー", () => {
    expect(validateNextVisitIntentDate("2000-01-01")).toBe("今日より前の日付は設定できません");
  });

  it("今日は有効", () => {
    expect(validateNextVisitIntentDate(todayJSTISO())).toBeNull();
  });

  it("今日から2年を超える日付はエラー", () => {
    expect(validateNextVisitIntentDate("2999-01-01")).toBe(
      "今日から2年以内の日付を設定してください",
    );
  });
});

describe("nextVisitFetchEndDate", () => {
  it("YYYY-MM-DD 形式で今日より未来の日付を返す", () => {
    const end = nextVisitFetchEndDate();

    expect(end).toMatch(/^\d{4}-\d{2}-\d{2}$/);
    expect(end > todayJSTISO()).toBe(true);
  });
});
