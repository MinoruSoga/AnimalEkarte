import { describe, expect, it } from "vitest";

import { formatJSTWallDate, todayJSTISO } from "@/lib/jst-date";

import { buildConfirmTentativeBody, buildTentativeConfirmInitialData } from "./confirm-tentative";

/**
 * EMR-308: confirm-tentative の request body 変換と確定モーダル初期値。
 * POST /v1/reservations/:id/confirm-tentative は同一行昇格なので
 * pet_id 等の参照フィールドは送らず、status は pending/confirmed のみ送る。
 */

describe("buildConfirmTentativeBody (EMR-308)", () => {
  const start = new Date(2026, 9, 5, 10, 30); // JST wall-clock 10:30
  const end = new Date(2026, 9, 5, 11, 30);

  it("start/end が無ければ null（fail-closed）", () => {
    expect(buildConfirmTentativeBody({})).toBeNull();
    expect(buildConfirmTentativeBody({ start })).toBeNull();
    expect(buildConfirmTentativeBody({ end })).toBeNull();
  });

  it("JST wall-clock を ISO へ変換し、type/doctor は数値化する", () => {
    const body = buildConfirmTentativeBody({
      start,
      end,
      type: "7",
      doctor: "33",
      isDesignated: true,
      status: "confirmed",
      notes: "メモ",
    });

    expect(body).toEqual({
      start_time: "2026-10-05T01:30:00.000Z",
      end_time: "2026-10-05T02:30:00.000Z",
      reservation_type_id: 7,
      doctor_id: 33,
      is_designated: true,
      status: "confirmed",
      notes: "メモ",
    });
  });

  it("status は pending/confirmed のみ送信し、他値は confirmed に倒す", () => {
    expect(buildConfirmTentativeBody({ start, end, status: "pending" })?.status).toBe("pending");
    expect(buildConfirmTentativeBody({ start, end, status: "checked_in" })?.status).toBe(
      "confirmed",
    );
    expect(buildConfirmTentativeBody({ start, end })?.status).toBe("confirmed");
  });

  it("数値化できない type/doctor は送らず、空 notes も送らない", () => {
    const body = buildConfirmTentativeBody({
      start,
      end,
      type: "一般診察",
      doctor: "",
      notes: "",
    });

    expect(body?.reservation_type_id).toBeUndefined();
    expect(body?.doctor_id).toBeUndefined();
    expect(body?.notes).toBeUndefined();
  });
});

describe("buildTentativeConfirmInitialData (EMR-308)", () => {
  it("意図日の 10:00–11:00 を初期値にし、tentative の区分/担当/メモを引き継ぐ", () => {
    const tentative = {
      // tentative.start は意図日 JST 00:00 sentinel — 未来日を使う
      start: new Date(2999, 5, 20, 0, 0),
      visitType: "first" as const,
      reservationTypeId: "9",
      doctorId: "55",
      notes: "経過観察",
    };

    const initial = buildTentativeConfirmInitialData(tentative, "10");

    expect(initial.petId).toBe("10");
    expect(formatJSTWallDate(initial.start as Date)).toBe("2999-06-20");
    expect(initial.start?.getHours()).toBe(10);
    expect(initial.end?.getHours()).toBe(11);
    expect(initial.visitType).toBe("first");
    expect(initial.type).toBe("9");
    expect(initial.doctor).toBe("55");
    expect(initial.status).toBe("confirmed");
    expect(initial.notes).toBe("経過観察");
    // 確定モーダルは別行作成ではないため id を持たない
    expect(initial.id).toBeUndefined();
  });

  it("意図日が過去（期限超過 tentative）なら初期日付を今日に倒す", () => {
    const initial = buildTentativeConfirmInitialData({ start: new Date(2000, 0, 1, 0, 0) }, "10");

    expect(formatJSTWallDate(initial.start as Date)).toBe(todayJSTISO());
  });
});
