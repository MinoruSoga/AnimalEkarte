import { describe, it, expect } from "vitest";
import { transformMedicalRecord } from "./transforms";
import type { BackendMedicalRecord } from "./types";

const minimal: BackendMedicalRecord = {
  id: 1,
  clinic_id: 1,
  pet_id: 10,
  owner_id: 20,
  doctor_id: 3,
  record_no: "MR-001",
  date: "2026-03-25T00:00:00Z",
  status: "draft",
  created_at: "2026-03-25T00:00:00Z",
  updated_at: "2026-03-25T00:00:00Z",
};

describe("transformMedicalRecord", () => {
  it("id を string に変換する", () => {
    expect(transformMedicalRecord({ ...minimal, id: 42 }).id).toBe("42");
  });

  it("id が未設定のとき '0' を返す", () => {
    expect(transformMedicalRecord({ ...minimal, id: undefined as unknown as number }).id).toBe("0");
  });

  it("record_no をそのまま返す", () => {
    expect(transformMedicalRecord({ ...minimal, record_no: "MR-999" }).recordNo).toBe("MR-999");
  });

  it("status: draft → '作成中'", () => {
    expect(transformMedicalRecord({ ...minimal, status: "draft" }).status).toBe("作成中");
  });

  it("status: finalized → '確定済'", () => {
    expect(transformMedicalRecord({ ...minimal, status: "finalized" }).status).toBe("確定済");
  });

  it("未知の status は '作成中' にフォールバックする", () => {
    expect(transformMedicalRecord({ ...minimal, status: "unknown" as "draft" }).status).toBe("作成中");
  });

  it("owner.owner_name を ownerName にマップする", () => {
    const result = transformMedicalRecord({
      ...minimal,
      owner: { id: 20, clinic_id: 1, owner_name: "田中太郎" } as BackendMedicalRecord["owner"],
    });
    expect(result.ownerName).toBe("田中太郎");
  });

  it("pet.name を petName にマップする", () => {
    const result = transformMedicalRecord({
      ...minimal,
      pet: { id: 10, clinic_id: 1, name: "ポチ" } as BackendMedicalRecord["pet"],
    });
    expect(result.petName).toBe("ポチ");
  });

  it("doctor.name を doctor にマップする", () => {
    const result = transformMedicalRecord({
      ...minimal,
      doctor: { id: 3, clinic_id: 1, name: "山田医師" } as BackendMedicalRecord["doctor"],
    });
    expect(result.doctor).toBe("山田医師");
  });

  it("doctor が未設定のとき doctor_id の文字列を使う", () => {
    expect(transformMedicalRecord({ ...minimal, doctor: undefined, doctor_id: 7 }).doctor).toBe("7");
  });

  it("pet_id を string に変換して petId にマップする", () => {
    expect(transformMedicalRecord({ ...minimal, pet_id: 10 }).petId).toBe("10");
  });

  it("pet_id が未設定のとき petId は undefined", () => {
    expect(transformMedicalRecord({ ...minimal, pet_id: undefined }).petId).toBeUndefined();
  });

  it("owner_id を string に変換して ownerId にマップする", () => {
    expect(transformMedicalRecord({ ...minimal, owner_id: 20 }).ownerId).toBe("20");
  });

  it("accounting_id を string に変換して accountingId にマップする", () => {
    expect(transformMedicalRecord({ ...minimal, accounting_id: 5 }).accountingId).toBe("5");
  });

  it("accounting_id が未設定のとき accountingId は undefined", () => {
    expect(transformMedicalRecord({ ...minimal, accounting_id: undefined }).accountingId).toBeUndefined();
  });

  it("clinical_plan の情報を objective/assessment/plan にマップする", () => {
    const result = transformMedicalRecord({
      ...minimal,
      clinical_plan: {
        physical_exam: "体重3kg",
        diagnosis_details: "風邪",
        treatment_policy: "安静",
      } as BackendMedicalRecord["clinical_plan"],
    });
    expect(result.objective).toBe("体重3kg");
    expect(result.assessment).toBe("風邪");
    expect(result.plan).toBe("安静");
  });

  it("inquiry.chief_complaint を chiefComplaint にマップする", () => {
    const result = transformMedicalRecord({
      ...minimal,
      inquiry: { chief_complaint: "元気がない", notes: "" } as BackendMedicalRecord["inquiry"],
    });
    expect(result.chiefComplaint).toBe("元気がない");
  });
});
