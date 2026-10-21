import { describe, expect, it } from "vitest";
import { LIFF_REVISIT_INTAKE_ACTIVE } from "./revisit-intake-activation";

describe("LIFF_REVISIT_INTAKE_ACTIVE (EMR-474)", () => {
  it("BE の 3 route と認可・receipt 処理が揃い配信先で使えるようになるまで false", () => {
    // true への変更は別 PR（必要 BE が揃い、配信先でも利用可能になった後）。
    expect(LIFF_REVISIT_INTAKE_ACTIVE).toBe(false);
  });
});
