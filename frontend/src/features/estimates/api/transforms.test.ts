import { describe, it, expect } from "vitest";
import { transformEstimate } from "./transforms";
import type { BackendEstimate, BackendEstimateItem } from "./types";

const item: BackendEstimateItem = {
  id: 1,
  estimate_id: 100,
  name: "診察料",
  category: "treatment" as BackendEstimateItem["category"],
  unit_price: 1000,
  quantity: 2,
  tax_rate: 0.1,
  discount_rate: 0,
  discount_amount: 0,
  is_insurance_applicable: false,
  sort_order: 1,
  created_at: "2026-03-25T00:00:00Z",
  updated_at: "2026-03-25T00:00:00Z",
};

const minimal: BackendEstimate = {
  id: 1,
  clinic_id: 1,
  estimate_no: "EST-001",
  title: "テスト見積",
  status: "draft" as BackendEstimate["status"],
  subtotal: 2000,
  tax_total: 200,
  total_amount: 2200,
  insurance_amount: 0,
  discount_amount: 0,
  comment: "",
  notes: "",
  items: [],
  created_at: "2026-03-25T00:00:00Z",
  updated_at: "2026-03-25T00:00:00Z",
};

describe("transformEstimate", () => {
  it("id を string に変換する", () => {
    expect(transformEstimate({ ...minimal, id: 99 }).id).toBe("99");
  });

  it("id が未設定のとき '0' を返す", () => {
    expect(transformEstimate({ ...minimal, id: undefined as unknown as number }).id).toBe("0");
  });

  it("clinic_id を string に変換して clinicId にマップする", () => {
    expect(transformEstimate({ ...minimal, clinic_id: 3 }).clinicId).toBe("3");
  });

  it("estimate_no をそのまま返す", () => {
    expect(transformEstimate({ ...minimal, estimate_no: "EST-999" }).estimateNo).toBe("EST-999");
  });

  it("estimate_no が未設定のとき空文字を返す", () => {
    expect(transformEstimate({ ...minimal, estimate_no: undefined }).estimateNo).toBe("");
  });

  it("title をそのまま返す", () => {
    expect(transformEstimate({ ...minimal, title: "新見積" }).title).toBe("新見積");
  });

  it("status をそのまま返す", () => {
    expect(transformEstimate({ ...minimal, status: "approved" as BackendEstimate["status"] }).status).toBe("approved");
  });

  it("status が未設定のとき 'draft' にフォールバックする", () => {
    expect(transformEstimate({ ...minimal, status: undefined }).status).toBe("draft");
  });

  it("subtotal / taxTotal / totalAmount を数値のまま返す", () => {
    const result = transformEstimate({ ...minimal, subtotal: 2000, tax_total: 200, total_amount: 2200 });
    expect(result.subtotal).toBe(2000);
    expect(result.taxTotal).toBe(200);
    expect(result.totalAmount).toBe(2200);
  });

  it("owner_id を string に変換して ownerId にマップする", () => {
    expect(transformEstimate({ ...minimal, owner_id: 5 }).ownerId).toBe("5");
  });

  it("owner_id が未設定のとき null を返す", () => {
    expect(transformEstimate({ ...minimal, owner_id: undefined }).ownerId).toBeNull();
  });

  it("pet_id を string に変換して petId にマップする", () => {
    expect(transformEstimate({ ...minimal, pet_id: 8 }).petId).toBe("8");
  });

  it("pet.name を petName にマップする", () => {
    expect(transformEstimate({ ...minimal, pet: { id: 8, name: "ポチ" } }).petName).toBe("ポチ");
  });

  it("medical_record_id を string に変換して medicalRecordId にマップする", () => {
    expect(transformEstimate({ ...minimal, medical_record_id: 7 }).medicalRecordId).toBe("7");
  });

  it("medical_record_id が未設定のとき null を返す", () => {
    expect(transformEstimate({ ...minimal, medical_record_id: undefined }).medicalRecordId).toBeNull();
  });

  it("items を変換して返す", () => {
    const result = transformEstimate({ ...minimal, items: [item] });
    expect(result.items).toHaveLength(1);
    expect(result.items[0].id).toBe("1");
    expect(result.items[0].name).toBe("診察料");
    expect(result.items[0].unitPrice).toBe(1000);
    expect(result.items[0].quantity).toBe(2);
  });

  it("owner.name を ownerName にマップする", () => {
    const result = transformEstimate({
      ...minimal,
      owner: { id: 5, clinic_id: 1, name: "田中太郎" } as BackendEstimate["owner"],
    });
    expect(result.ownerName).toBe("田中太郎");
  });

  it("valid_until / comment / notes をそのまま返す", () => {
    const result = transformEstimate({
      ...minimal,
      valid_until: "2026-04-30",
      comment: "コメント",
      notes: "備考",
    });
    expect(result.validUntil).toBe("2026-04-30");
    expect(result.comment).toBe("コメント");
    expect(result.notes).toBe("備考");
  });
});
