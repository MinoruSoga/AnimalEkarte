import { describe, it, expect } from "vitest";
import { transformToAccounting } from "./transforms";
import type { BackendAccounting, BackendAccountingItem } from "./types";

const item: BackendAccountingItem = {
  id: 1,
  billing_id: 100,
  category: "treatment" as BackendAccountingItem["category"],
  name: "診察料",
  unit_price: 1000,
  quantity: 1,
  tax_rate: 0.1,
  is_insurance_applicable: false,
  source: "medical_record" as BackendAccountingItem["source"],
  created_at: "2026-03-25T00:00:00Z",
  updated_at: "2026-03-25T00:00:00Z",
};

const minimal: BackendAccounting = {
  id: 1,
  clinic_id: 1,
  owner_id: 10,
  pet_id: 20,
  status: "waiting" as BackendAccounting["status"],
  scheduled_date: "2026-03-25T00:00:00Z",
  memo: "",
  items: [],
  created_at: "2026-03-25T00:00:00Z",
  updated_at: "2026-03-25T00:00:00Z",
};

describe("transformToAccounting", () => {
  it("id を string に変換する", () => {
    expect(transformToAccounting({ ...minimal, id: 42 }).id).toBe("42");
  });

  it("id が未設定のとき '0' を返す", () => {
    expect(transformToAccounting({ ...minimal, id: undefined as unknown as number }).id).toBe("0");
  });

  it("owner_id を string に変換して ownerId にマップする", () => {
    expect(transformToAccounting({ ...minimal, owner_id: 10 }).ownerId).toBe("10");
  });

  it("pet_id を string に変換して petId にマップする", () => {
    expect(transformToAccounting({ ...minimal, pet_id: 20 }).petId).toBe("20");
  });

  it("owner.owner_name を ownerName にマップする", () => {
    const result = transformToAccounting({
      ...minimal,
      owner: { id: 10, clinic_id: 1, owner_name: "田中太郎" } as BackendAccounting["owner"],
    });
    expect(result.ownerName).toBe("田中太郎");
  });

  it("pet.name を petName にマップする", () => {
    const result = transformToAccounting({
      ...minimal,
      pet: { id: 20, clinic_id: 1, name: "ポチ" } as BackendAccounting["pet"],
    });
    expect(result.petName).toBe("ポチ");
  });

  it("scheduled_date を先頭 10 文字に整形する", () => {
    expect(transformToAccounting({ ...minimal, scheduled_date: "2026-03-25T10:00:00Z" }).scheduledDate).toBe("2026-03-25");
  });

  it("items を変換して返す", () => {
    const result = transformToAccounting({ ...minimal, items: [item] });
    expect(result.items).toHaveLength(1);
    expect(result.items[0].id).toBe("1");
    expect(result.items[0].name).toBe("診察料");
    expect(result.items[0].unitPrice).toBe(1000);
    expect(result.items[0].taxRate).toBe(0.1);
  });

  it("tax_rate 0.08 は 0.08 として保持する", () => {
    const result = transformToAccounting({
      ...minimal,
      items: [{ ...item, tax_rate: 0.08 }],
    });
    expect(result.items[0].taxRate).toBe(0.08);
  });

  it("status が waiting の場合 payment は undefined", () => {
    expect(transformToAccounting({ ...minimal, status: "waiting" as BackendAccounting["status"] }).payment).toBeUndefined();
  });

  it("status が completed かつ payment が存在する場合 payment を返す", () => {
    const result = transformToAccounting({
      ...minimal,
      status: "completed" as BackendAccounting["status"],
      subtotal: 1000,
      tax_total: 100,
      total_amount: 1100,
      payments: [{
        id: 1,
        billing_id: 1,
        method: "cash" as BackendAccounting["payments"][0]["method"],
        billing_amount: 1100,
        received_amount: 2000,
        change_amount: 900,
        insurance_amount: 0,
        discount_amount: 0,
        created_at: "2026-03-25T00:00:00Z",
        updated_at: "2026-03-25T00:00:00Z",
      }],
    });
    expect(result.payment).toBeDefined();
    expect(result.payment!.totalAmount).toBe(1100);
    expect(result.payment!.method).toBe("cash");
  });

  it("medical_record_id が存在するとき string に変換して medicalRecordId にマップする", () => {
    expect(transformToAccounting({ ...minimal, medical_record_id: 5 }).medicalRecordId).toBe("5");
  });

  it("medical_record_id が未設定のとき medicalRecordId は undefined", () => {
    expect(transformToAccounting({ ...minimal, medical_record_id: undefined }).medicalRecordId).toBeUndefined();
  });

  it("memo が空のとき memo は undefined", () => {
    expect(transformToAccounting({ ...minimal, memo: "" }).memo).toBeUndefined();
  });

  it("memo が存在するとき返す", () => {
    expect(transformToAccounting({ ...minimal, memo: "注意事項" }).memo).toBe("注意事項");
  });
});
