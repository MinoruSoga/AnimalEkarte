import { describe, it, expect } from "vitest";
import { transformOwner } from "./transforms";
import type { OwnerApiResponse } from "./transforms";

/**
 * OwnerApiResponse の最小スタブ。
 * API ハンドラが返す owner_name を使う（BackendOwner.name ではない）。
 */
const minimalBackend: OwnerApiResponse = {
  id: 1,
  clinic_id: 1,
  owner_name: "田中太郎",
  membership_type: "non_member",
  delivery_excluded: false,
  is_transferred: false,
  lstep_opt_out: false,
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
    const result = transformOwner({ ...minimalBackend, owner_name: undefined as unknown as string });
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

  it("line_user_id を lineUserId にマップする", () => {
    const result = transformOwner({ ...minimalBackend, line_user_id: "Uabc123" });
    expect(result.lineUserId).toBe("Uabc123");
  });

  it("line_user_id が未設定のとき undefined を返す", () => {
    const result = transformOwner({ ...minimalBackend, line_user_id: undefined });
    expect(result.lineUserId).toBeUndefined();
  });

  it("line_id_confirmed_at を lineIdConfirmedAt にマップする", () => {
    const result = transformOwner({
      ...minimalBackend,
      line_id_confirmed_at: "2026-04-01T10:00:00Z",
    });
    expect(result.lineIdConfirmedAt).toBe("2026-04-01T10:00:00Z");
  });

  it("delivery_excluded を deliveryExcluded にマップする (true)", () => {
    const result = transformOwner({ ...minimalBackend, delivery_excluded: true });
    expect(result.deliveryExcluded).toBe(true);
  });

  it("delivery_excluded が未設定のとき false を返す", () => {
    const result = transformOwner({
      ...minimalBackend,
      delivery_excluded: undefined as unknown as boolean,
    });
    expect(result.deliveryExcluded).toBe(false);
  });

  it("delivery_excluded_reason を deliveryExcludedReason にマップする", () => {
    const result = transformOwner({
      ...minimalBackend,
      delivery_excluded: true,
      delivery_excluded_reason: "配信不要希望",
    });
    expect(result.deliveryExcludedReason).toBe("配信不要希望");
  });

  it("is_transferred を isTransferred にマップする (true)", () => {
    const result = transformOwner({ ...minimalBackend, is_transferred: true });
    expect(result.isTransferred).toBe(true);
  });

  it("is_transferred が未設定のとき false を返す", () => {
    const result = transformOwner({
      ...minimalBackend,
      is_transferred: undefined as unknown as boolean,
    });
    expect(result.isTransferred).toBe(false);
  });

  it("transfer_at を transferAt にマップする", () => {
    const result = transformOwner({
      ...minimalBackend,
      is_transferred: true,
      transfer_at: "2026-03-15T00:00:00Z",
    });
    expect(result.transferAt).toBe("2026-03-15T00:00:00Z");
  });

  it("lstep_opt_out を lstepOptOut にマップする", () => {
    const result = transformOwner({ ...minimalBackend, lstep_opt_out: true });
    expect(result.lstepOptOut).toBe(true);
  });

  it("lstep_opt_out が未設定のとき false を返す", () => {
    const result = transformOwner({
      ...minimalBackend,
      lstep_opt_out: undefined as unknown as boolean,
    });
    expect(result.lstepOptOut).toBe(false);
  });

  it("lstep_opt_out_reason を lstepOptOutReason にマップする", () => {
    const result = transformOwner({
      ...minimalBackend,
      lstep_opt_out: true,
      lstep_opt_out_reason: "苦情あり",
    });
    expect(result.lstepOptOutReason).toBe("苦情あり");
  });

  it("dm_preference=true を dmPreference=true にマップする", () => {
    const result = transformOwner({ ...minimalBackend, dm_preference: true });
    expect(result.dmPreference).toBe(true);
  });

  it("dm_preference=false を undefined に潰さず dmPreference=false にマップする", () => {
    const result = transformOwner({ ...minimalBackend, dm_preference: false });
    expect(result.dmPreference).toBe(false);
  });

  it("dm_preference 未設定は undefined のまま", () => {
    const result = transformOwner({ ...minimalBackend, dm_preference: undefined });
    expect(result.dmPreference).toBeUndefined();
  });

  it("dm_preference=null は null のまま（未設定を不要に潰さない）", () => {
    const result = transformOwner({ ...minimalBackend, dm_preference: null });
    expect(result.dmPreference).toBeNull();
  });
});
