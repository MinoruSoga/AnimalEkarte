import { describe, it, expect } from "vitest";
import { transformOwner } from "./transforms";
import type { OwnerResponse, PetInOwnerResponse } from "@/types/generated/owner-responses";

/**
 * OwnerResponse の最小スタブ。
 * wire 正本は owner domain の OwnerResponse（owner_name / owner_name_kana）。
 * models.ts の Owner（name / name_kana）は使わない（BUG-433）。
 */
function makeOwnerResponse(overrides: Partial<OwnerResponse> = {}): OwnerResponse {
  return {
    id: 1,
    clinic_id: 1,
    owner_name: "田中太郎",
    owner_name_kana: "",
    company: "",
    postal_code: "",
    address1: "",
    address2: "",
    home_postal_code: "",
    home_address1: "",
    home_address2: "",
    phone: "",
    company_phone: "",
    email: "",
    remarks: "",
    is_dangerous: false,
    discount_rate: 0,
    membership_type: "non_member",
    delivery_excluded: false,
    delivery_caution: false,
    is_transferred: false,
    pets: [],
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
    ...overrides,
  };
}

function makePetInOwner(
  overrides: Partial<PetInOwnerResponse> = {},
): PetInOwnerResponse {
  return {
    id: 10,
    owner_id: 1,
    animal_species_id: 1,
    pet_number: "P-001",
    name: "ポチ",
    pet_name_kana: "",
    gender: "unknown",
    status: "alive",
    breed: "",
    color: "",
    danger_level: "low",
    food: "",
    environment: "",
    remarks: "",
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
    ...overrides,
  };
}

describe("transformOwner", () => {
  it("id を string に変換する", () => {
    const result = transformOwner(makeOwnerResponse({ id: 99 }));
    expect(result.id).toBe("99");
  });

  it("id が null/undefined のとき '0' を返す", () => {
    const result = transformOwner(
      makeOwnerResponse({ id: undefined as unknown as number }),
    );
    expect(result.id).toBe("0");
  });

  it("owner_name を ownerName にマップする", () => {
    const result = transformOwner(makeOwnerResponse({ owner_name: "佐藤花子" }));
    expect(result.ownerName).toBe("佐藤花子");
  });

  it("owner_name が未設定のとき空文字を返す", () => {
    const result = transformOwner(
      makeOwnerResponse({ owner_name: undefined as unknown as string }),
    );
    expect(result.ownerName).toBe("");
  });

  it("owner_name_kana を ownerNameKana にマップする", () => {
    const result = transformOwner(
      makeOwnerResponse({ owner_name_kana: "サトウハナコ" }),
    );
    expect(result.ownerNameKana).toBe("サトウハナコ");
  });

  it("membership_type: non_member → '非会員'", () => {
    const result = transformOwner(makeOwnerResponse({ membership_type: "non_member" }));
    expect(result.membershipType).toBe("非会員");
  });

  it("membership_type: member → '会員'", () => {
    const result = transformOwner(makeOwnerResponse({ membership_type: "member" }));
    expect(result.membershipType).toBe("会員");
  });

  it("membership_type: deceased → '退亡者'", () => {
    const result = transformOwner(makeOwnerResponse({ membership_type: "deceased" }));
    expect(result.membershipType).toBe("退亡者");
  });

  it("membership_type: transferred → '他診/準'", () => {
    const result = transformOwner(makeOwnerResponse({ membership_type: "transferred" }));
    expect(result.membershipType).toBe("他診/準");
  });

  it("未知の membership_type はそのまま返す", () => {
    const result = transformOwner(makeOwnerResponse({ membership_type: "custom_type" }));
    expect(result.membershipType).toBe("custom_type");
  });

  it("birth_date を T 以前の日付部分のみに整形する", () => {
    const result = transformOwner(
      makeOwnerResponse({ birth_date: "1990-04-01T00:00:00Z" }),
    );
    expect(result.birthDate).toBe("1990-04-01");
  });

  it("birth_date が未設定のとき undefined を返す", () => {
    const result = transformOwner(makeOwnerResponse({ birth_date: undefined }));
    expect(result.birthDate).toBeUndefined();
  });

  it("is_dangerous を isDangerous にマップする", () => {
    const result = transformOwner(makeOwnerResponse({ is_dangerous: true }));
    expect(result.isDangerous).toBe(true);
  });

  it("is_dangerous が未設定のとき false を返す", () => {
    const result = transformOwner(
      makeOwnerResponse({ is_dangerous: undefined as unknown as boolean }),
    );
    expect(result.isDangerous).toBe(false);
  });

  it("discount_rate を discountRate にマップする", () => {
    const result = transformOwner(makeOwnerResponse({ discount_rate: 10 }));
    expect(result.discountRate).toBe(10);
  });

  it("discount_rate が未設定のとき 0 を返す", () => {
    const result = transformOwner(
      makeOwnerResponse({ discount_rate: undefined as unknown as number }),
    );
    expect(result.discountRate).toBe(0);
  });

  it("pets が存在するとき変換して返す", () => {
    const result = transformOwner(
      makeOwnerResponse({ pets: [makePetInOwner({ id: 10, name: "ポチ" })] }),
    );
    expect(result.pets).toHaveLength(1);
    expect(result.pets![0].id).toBe("10");
    expect(result.pets![0].name).toBe("ポチ");
  });

  it("owner埋込petの死亡statusと死亡日時を保持し、未知statusは不明にする", () => {
    const deceasedAt = "2026-07-10T12:00:00+09:00";
    const result = transformOwner(
      makeOwnerResponse({
        pets: [
          makePetInOwner({ status: "deceased", deceased_at: deceasedAt }),
          makePetInOwner({ id: 11, status: "unexpected" }),
        ],
      }),
    );

    expect(result.pets?.[0]).toEqual(
      expect.objectContaining({ status: "死亡", deceasedAt }),
    );
    expect(result.pets?.[1].status).toBe("不明");
  });

  it("pets が空配列のとき空配列を返す", () => {
    const result = transformOwner(makeOwnerResponse({ pets: [] }));
    expect(result.pets).toEqual([]);
  });

  it("住所フィールドを正しくマップする", () => {
    const result = transformOwner(
      makeOwnerResponse({
        postal_code: "123-4567",
        address1: "東京都",
        address2: "新宿区",
      }),
    );
    expect(result.postalCode).toBe("123-4567");
    expect(result.address1).toBe("東京都");
    expect(result.address2).toBe("新宿区");
  });

  it("OwnerResponse に無い line_user_id は lineUserId を埋めない", () => {
    // LINE user id は owner-line-tags 等の専用 API の正本（BUG-433）。
    const result = transformOwner(makeOwnerResponse());
    expect(result.lineUserId).toBeUndefined();
  });

  it("line_id_confirmed_at を lineIdConfirmedAt にマップする", () => {
    const result = transformOwner(
      makeOwnerResponse({ line_id_confirmed_at: "2026-04-01T10:00:00Z" }),
    );
    expect(result.lineIdConfirmedAt).toBe("2026-04-01T10:00:00Z");
  });

  it("delivery_excluded を deliveryExcluded にマップする (true)", () => {
    const result = transformOwner(makeOwnerResponse({ delivery_excluded: true }));
    expect(result.deliveryExcluded).toBe(true);
  });

  it("delivery_excluded が未設定のとき false を返す", () => {
    const result = transformOwner(
      makeOwnerResponse({ delivery_excluded: undefined as unknown as boolean }),
    );
    expect(result.deliveryExcluded).toBe(false);
  });

  it("delivery_excluded_reason を deliveryExcludedReason にマップする", () => {
    const result = transformOwner(
      makeOwnerResponse({
        delivery_excluded: true,
        delivery_excluded_reason: "配信不要希望",
      }),
    );
    expect(result.deliveryExcludedReason).toBe("配信不要希望");
  });

  it("is_transferred を isTransferred にマップする (true)", () => {
    const result = transformOwner(makeOwnerResponse({ is_transferred: true }));
    expect(result.isTransferred).toBe(true);
  });

  it("is_transferred が未設定のとき false を返す", () => {
    const result = transformOwner(
      makeOwnerResponse({ is_transferred: undefined as unknown as boolean }),
    );
    expect(result.isTransferred).toBe(false);
  });

  it("transfer_at を transferAt にマップする", () => {
    const result = transformOwner(
      makeOwnerResponse({
        is_transferred: true,
        transfer_at: "2026-03-15T00:00:00Z",
      }),
    );
    expect(result.transferAt).toBe("2026-03-15T00:00:00Z");
  });

  it("lstep 系は OwnerResponse に無いため UI 既定値になる", () => {
    const result = transformOwner(makeOwnerResponse());
    expect(result.lstepOptOut).toBe(false);
    expect(result.lstepOptOutAt).toBeUndefined();
    expect(result.lstepOptOutReason).toBeUndefined();
  });

  it("dm_preference=true を dmPreference=true にマップする", () => {
    const result = transformOwner(makeOwnerResponse({ dm_preference: true }));
    expect(result.dmPreference).toBe(true);
  });

  it("dm_preference=false を undefined に潰さず dmPreference=false にマップする", () => {
    const result = transformOwner(makeOwnerResponse({ dm_preference: false }));
    expect(result.dmPreference).toBe(false);
  });

  it("dm_preference 未設定は undefined のまま", () => {
    const result = transformOwner(makeOwnerResponse({ dm_preference: undefined }));
    expect(result.dmPreference).toBeUndefined();
  });
});
