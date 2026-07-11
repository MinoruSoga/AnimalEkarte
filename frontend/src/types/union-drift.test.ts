import { describe, it, expect } from "vitest";
import * as gen from "@/types/generated/models";
import type { VisitType, ReservationStatus } from "./index";
import { RESERVATION_STATUS_VALUES } from "./index";

// FE4-4: 既存 literal union は型安全性は健在だが、backend が値を追加すると黙って古くなる。
// drift テストで生成定数の値集合との一致を追随漏れ検知として機械固定する（実装は無変更）。
describe("src/types union drift", () => {
  it("VisitType の値集合が VisitType* 生成定数と一致する", () => {
    const values: VisitType[] = ["first", "revisit"];
    expect(new Set<string>(values)).toEqual(
      new Set([gen.VisitTypeFirst, gen.VisitTypeRevisit]),
    );
  });

  it("ReservationStatus の値集合が ReservationStatus* 生成定数と一致する", () => {
    const values: ReservationStatus[] = [...RESERVATION_STATUS_VALUES];
    expect(new Set<string>(values)).toEqual(
      new Set([
        gen.ReservationStatusConfirmed,
        gen.ReservationStatusPending,
        gen.ReservationStatusCheckedIn,
        gen.ReservationStatusInConsultation,
        gen.ReservationStatusAccounting,
        gen.ReservationStatusCompleted,
        gen.ReservationStatusCancelled,
        gen.ReservationStatusNoShow,
      ]),
    );
  });
});
