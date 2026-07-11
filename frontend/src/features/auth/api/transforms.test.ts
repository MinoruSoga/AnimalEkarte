/**
 * FE5-1: BE `MeResponse.Clinics` は `json:"clinics,omitempty"` で省略可能だが、
 * FE スキーマは `clinics: z.array(...)` 必須だったため、所属クリニック 0 件のスタッフで
 * `/me` の parse が throw し認証フローが落ちる回帰バグ（FE-refactor.md §4.1 M1）。
 */
import { describe, it, expect } from "vitest";
import { mapMeToAuthUser } from "./transforms";

function buildRawMe(overrides: Record<string, unknown> = {}): Record<string, unknown> {
  return {
    id: "user-1",
    email: "staff@example.com",
    display_name: "田中 太郎",
    is_system_admin: false,
    main_clinic_id: "clinic-a",
    clinic: null,
    permissions: {},
    ...overrides,
  };
}

describe("FE5-1: mapMeToAuthUser — clinics omitempty 耐性", () => {
  it("clinics フィールドが未定義（BE omitempty）でも throw せず空配列になる", () => {
    const raw = buildRawMe();
    expect("clinics" in raw).toBe(false);

    const user = mapMeToAuthUser(raw);

    expect(user.clinics).toEqual([]);
  });

  it("clinics が通常通り渡された場合は従来通りマップされる", () => {
    const raw = buildRawMe({
      clinics: [{ clinic_id: "clinic-a", clinic_name: "八王子院", is_main: true }],
    });

    const user = mapMeToAuthUser(raw);

    expect(user.clinics).toEqual([
      { clinicId: "clinic-a", clinicName: "八王子院", isMain: true },
    ]);
  });
});
