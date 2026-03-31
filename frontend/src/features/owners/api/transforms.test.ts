import { describe, it, expect } from "vitest";
import { transformOwner } from "./transforms";
import type { Owner as BackendOwner } from "@/types/generated/models";

/** BackendOwner の最小スタブ */
const minimalBackend: BackendOwner = {
  id: 1,
  clinic_id: 1,
  owner_name: "田中太郎",
  membership_type: "non_member",
  created_at: "2026-01-01T00:00:00Z",
  updated_at: "2026-01-01T00:00:00Z",
};

describe("transformOwner", () => {
  it("id を string に変換する", () => {
    const result = transformOwner({ ...minimalBackend, id: 99 });
    expect(result.id).toBe("99");
  });

  it("id が null/undefined のとき '0' を返す", () => {
    const result = transformOwner({ ...minimalBackend, id: undefined as unknown as number });
    expect(result.id).toBe("0");
  });

  it("owner_name を ownerName にマップする", () => {
    const result = transformOwner({ ...minimalBackend, owner_name: "佐藤花子" });
    expect(result.ownerName).toBe("佐藤花子");
  });

  it("owner_name が未設定のとき空文字を返す", () => {
    const result = transformOwner({ ...minimalBackend, owner_name: undefined });
    expect(result.ownerName).toBe("");
  });

  it("membership_type: non_member → '非会員'", () => {
    const result = transformOwner({ ...minimalBackend, membership_type: "non_member" });
    expect(result.membershipType).toBe("非会員");
  });

  it("membership_type: member → '会員'", () => {
    const result = transformOwner({ ...minimalBackend, membership_type: "member" });
    expect(result.membershipType).toBe("会員");
  });

  it("membership_type: deceased → '退亡者'", () => {
    const result = transformOwner({ ...minimalBackend, membership_type: "deceased" });
    expect(result.membershipType).toBe("退亡者");
  });

  it("membership_type: transferred → '他診/準'", () => {
    const result = transformOwner({ ...minimalBackend, membership_type: "transferred" });
    expect(result.membershipType).toBe("他診/準");
  });

  it("未知の membership_type はそのまま返す", () => {
    const result = transformOwner({ ...minimalBackend, membership_type: "custom_type" });
    expect(result.membershipType).toBe("custom_type");
  });

  it("birth_date を T 以前の日付部分のみに整形する", () => {
    const result = transformOwner({ ...minimalBackend, birth_date: "1990-04-01T00:00:00Z" });
    expect(result.birthDate).toBe("1990-04-01");
  });

  it("birth_date が未設定のとき undefined を返す", () => {
    const result = transformOwner({ ...minimalBackend, birth_date: undefined });
    expect(result.birthDate).toBeUndefined();
  });

  it("is_dangerous を isDangerous にマップする", () => {
    const result = transformOwner({ ...minimalBackend, is_dangerous: true });
    expect(result.isDangerous).toBe(true);
  });

  it("is_dangerous が未設定のとき false を返す", () => {
    const result = transformOwner({ ...minimalBackend, is_dangerous: undefined });
    expect(result.isDangerous).toBe(false);
  });

  it("discount_rate を discountRate にマップする", () => {
    const result = transformOwner({ ...minimalBackend, discount_rate: 10 });
    expect(result.discountRate).toBe(10);
  });

  it("discount_rate が未設定のとき 0 を返す", () => {
    const result = transformOwner({ ...minimalBackend, discount_rate: undefined });
    expect(result.discountRate).toBe(0);
  });

  it("pets が存在するとき変換して返す", () => {
    const result = transformOwner({
      ...minimalBackend,
      pets: [
        {
          id: 10,
          clinic_id: 1,
          owner_id: 1,
          name: "ポチ",
          pet_number: "P-001",
          created_at: "2026-01-01T00:00:00Z",
          updated_at: "2026-01-01T00:00:00Z",
        },
      ],
    });
    expect(result.pets).toHaveLength(1);
    expect(result.pets![0].id).toBe("10");
    expect(result.pets![0].name).toBe("ポチ");
  });

  it("pets が未設定のとき undefined を返す", () => {
    const result = transformOwner({ ...minimalBackend, pets: undefined });
    expect(result.pets).toBeUndefined();
  });

  it("住所フィールドを正しくマップする", () => {
    const result = transformOwner({
      ...minimalBackend,
      postal_code: "123-4567",
      address1: "東京都",
      address2: "新宿区",
    });
    expect(result.postalCode).toBe("123-4567");
    expect(result.address1).toBe("東京都");
    expect(result.address2).toBe("新宿区");
  });
});
