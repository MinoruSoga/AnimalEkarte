import { describe, expect, it } from "vitest";

import { resolvePostCreateOwnerNavigation } from "./post-create-owner-navigation";

describe("resolvePostCreateOwnerNavigation (BUG-010)", () => {
  it("同一クリニックなら SPA navigate", () => {
    expect(
      resolvePostCreateOwnerNavigation({
        ownerId: "42",
        targetClinicId: "1",
        currentClinicId: "1",
      }),
    ).toEqual({ mode: "spa", href: "/owners/42" });
  });

  it("登録先未指定なら SPA navigate（サーバは X-Clinic-ID 既定）", () => {
    expect(
      resolvePostCreateOwnerNavigation({
        ownerId: "42",
        targetClinicId: undefined,
        currentClinicId: "1",
      }),
    ).toEqual({ mode: "spa", href: "/owners/42" });
  });

  it("登録先がグローバル選択と異なるなら hard navigate + 切替 clinicId", () => {
    expect(
      resolvePostCreateOwnerNavigation({
        ownerId: "99",
        targetClinicId: "2",
        currentClinicId: "1",
      }),
    ).toEqual({ mode: "hard", href: "/owners/99", clinicId: "2" });
  });

  it("空白 clinicId は未指定扱い", () => {
    expect(
      resolvePostCreateOwnerNavigation({
        ownerId: "7",
        targetClinicId: "   ",
        currentClinicId: "1",
      }),
    ).toEqual({ mode: "spa", href: "/owners/7" });
  });
});
