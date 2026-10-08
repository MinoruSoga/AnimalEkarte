import { describe, expect, it } from "vitest";

import { buildJSTWallDateTime } from "@/lib/jst-date";

import {
  isExpiredTentativeReservation,
  isTentativeReservation,
  tentativeIntendedDate,
} from "./tentative";

/**
 * EMR-308: tentative（次回予定・未予約）判定ユーティリティの契約固定。
 * - isTentative は transformReservation が正規化したフィールドだけを見る
 *   （status の生値 "tentative" は generated 型に無く isTentative 経由で統一）。
 * - 意図日は JST 暦日比較。当日の tentative は期限超過ではない（確定/紐付け対象）。
 */

describe("isTentativeReservation", () => {
  it("isTentative=true の行のみ tentative 判定する", () => {
    expect(isTentativeReservation({ isTentative: true })).toBe(true);
    expect(isTentativeReservation({ isTentative: false })).toBe(false);
    // transform 未経由の stub や旧オブジェクトは fail-closed で非 tentative
    expect(isTentativeReservation({})).toBe(false);
  });
});

describe("tentativeIntendedDate", () => {
  it("intendedDate を優先して返す", () => {
    expect(
      tentativeIntendedDate({
        intendedDate: "2026-10-10",
        start: buildJSTWallDateTime("2026-10-20", "00:00"),
      }),
    ).toBe("2026-10-10");
  });

  it("intendedDate 欠落時は sentinel start の JST 暦日へフォールバックする", () => {
    expect(tentativeIntendedDate({ start: buildJSTWallDateTime("2026-10-20", "00:00") })).toBe(
      "2026-10-20",
    );
  });
});

describe("isExpiredTentativeReservation", () => {
  const TODAY = "2026-10-06";

  it("意図日が昨日以前の tentative は期限超過", () => {
    expect(
      isExpiredTentativeReservation(
        { isTentative: true, intendedDate: "2026-10-05", start: new Date() },
        TODAY,
      ),
    ).toBe(true);
  });

  it("当日の tentative は期限超過ではない（当日こそ確定/紐付け対象）", () => {
    expect(
      isExpiredTentativeReservation(
        { isTentative: true, intendedDate: "2026-10-06", start: new Date() },
        TODAY,
      ),
    ).toBe(false);
  });

  it("未来の tentative は期限超過ではない", () => {
    expect(
      isExpiredTentativeReservation(
        { isTentative: true, intendedDate: "2026-10-07", start: new Date() },
        TODAY,
      ),
    ).toBe(false);
  });

  it("tentative でない行は意図日が過去でも期限超過にしない", () => {
    expect(
      isExpiredTentativeReservation(
        { isTentative: false, intendedDate: "2020-01-01", start: new Date() },
        TODAY,
      ),
    ).toBe(false);
  });

  it("intendedDate 欠落時は sentinel start の暦日で判定する", () => {
    // sentinel は意図日の JST 00:00。当日の sentinel は期限超過ではない。
    expect(
      isExpiredTentativeReservation(
        { isTentative: true, start: buildJSTWallDateTime("2026-10-06", "00:00") },
        TODAY,
      ),
    ).toBe(false);
    expect(
      isExpiredTentativeReservation(
        { isTentative: true, start: buildJSTWallDateTime("2026-10-05", "00:00") },
        TODAY,
      ),
    ).toBe(true);
  });
});
