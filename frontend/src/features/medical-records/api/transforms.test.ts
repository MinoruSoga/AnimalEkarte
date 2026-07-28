import { describe, it, expect } from "vitest";
import { transformMedicalRecord, transformToHistoryItem, toBackendMedicalRecordStatus } from "./transforms";
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

  it("owner.name を ownerName にマップする", () => {
    const result = transformMedicalRecord({
      ...minimal,
      owner: { id: 20, clinic_id: 1, name: "田中太郎" } as BackendMedicalRecord["owner"],
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

  it("pet.status=deceased は明示的な死亡状態へ正規化する", () => {
    const result = transformMedicalRecord({
      ...minimal,
      pet: {
        id: 10,
        clinic_id: 1,
        name: "ポチ",
        status: "deceased",
      } as BackendMedicalRecord["pet"],
    });
    expect(result.petIsDeceased).toBe(true);
  });

  it("pet.status がaliveまたは未取得なら死亡扱いにしない", () => {
    expect(transformMedicalRecord({
      ...minimal,
      pet: {
        id: 10,
        clinic_id: 1,
        name: "ポチ",
        status: "alive",
      } as BackendMedicalRecord["pet"],
    }).petIsDeceased).toBe(false);
    expect(transformMedicalRecord({ ...minimal, pet: undefined }).petIsDeceased).toBe(false);
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

  it("BUG-406 follow-up: inquiry.chief_complaint_type_id を chiefComplaintTypeId にマップする", () => {
    const result = transformMedicalRecord({
      ...minimal,
      inquiry: { chief_complaint: "", chief_complaint_type_id: 5, notes: "" } as BackendMedicalRecord["inquiry"],
    });
    expect(result.chiefComplaintTypeId).toBe(5);
  });

  it("BUG-406 follow-up: inquiry.chief_complaint_type_id が未設定のとき chiefComplaintTypeId は null", () => {
    const result = transformMedicalRecord({
      ...minimal,
      inquiry: { chief_complaint: "", notes: "" } as BackendMedicalRecord["inquiry"],
    });
    expect(result.chiefComplaintTypeId).toBeNull();
  });

  it("BUG-410: clinical_plan の diagnosis_type_id/diagnosis_name_id/diagnosis_2_type_id/diagnosis_2_name_id を diagnosis1CategoryId/diagnosis1NameId/diagnosis2CategoryId/diagnosis2NameId にマップする", () => {
    const result = transformMedicalRecord({
      ...minimal,
      clinical_plan: {
        physical_exam: "",
        diagnosis_details: "",
        treatment_policy: "",
        diagnosis_type_id: 3,
        diagnosis_name_id: 7,
        diagnosis_2_type_id: 4,
        diagnosis_2_name_id: 9,
      } as BackendMedicalRecord["clinical_plan"],
    });
    expect(result.diagnosis1CategoryId).toBe(3);
    expect(result.diagnosis1NameId).toBe(7);
    expect(result.diagnosis2CategoryId).toBe(4);
    expect(result.diagnosis2NameId).toBe(9);
  });

  it("BUG-410: diagnosis1/2 の ID が未設定のとき null（undefined ではなく明示的に null）", () => {
    const result = transformMedicalRecord({
      ...minimal,
      clinical_plan: {
        physical_exam: "",
        diagnosis_details: "",
        treatment_policy: "",
      } as BackendMedicalRecord["clinical_plan"],
    });
    expect(result.diagnosis1CategoryId).toBeNull();
    expect(result.diagnosis1NameId).toBeNull();
    expect(result.diagnosis2CategoryId).toBeNull();
    expect(result.diagnosis2NameId).toBeNull();
  });

  it("visit_count をそのまま返す", () => {
    expect(transformMedicalRecord({ ...minimal, visit_count: 5 }).visitCount).toBe(5);
  });

  it("version が未設定のとき 1 を返す", () => {
    expect(transformMedicalRecord({ ...minimal, version: undefined }).version).toBe(1);
  });

  it("version が設定済みのときその値を返す", () => {
    expect(transformMedicalRecord({ ...minimal, version: 3 }).version).toBe(3);
  });

  it("pet.animal_species.name を species にマップする", () => {
    const result = transformMedicalRecord({
      ...minimal,
      pet: {
        id: 10, clinic_id: 1, name: "ポチ",
        animal_species: { id: 1, clinic_id: 1, name: "犬" },
      } as BackendMedicalRecord["pet"],
    });
    expect(result.species).toBe("犬");
  });

  it("inquiry.notes を notes にマップする", () => {
    const result = transformMedicalRecord({
      ...minimal,
      inquiry: { chief_complaint: "", notes: "備考テキスト" } as BackendMedicalRecord["inquiry"],
    });
    expect(result.notes).toBe("備考テキスト");
  });
});

// ─────────────────────────────────────────────────────────────
// transformToHistoryItem
// ─────────────────────────────────────────────────────────────
describe("transformToHistoryItem", () => {
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

  it("id を string に変換する", () => {
    expect(transformToHistoryItem({ ...minimal, id: 42 }).id).toBe("42");
  });

  it("doctor.name を author にマップする", () => {
    const result = transformToHistoryItem({
      ...minimal,
      doctor: { id: 3, clinic_id: 1, name: "山田医師" } as BackendMedicalRecord["doctor"],
    });
    expect(result.author).toBe("山田医師");
  });

  it("doctor が未設定のとき author は '-'", () => {
    expect(transformToHistoryItem({ ...minimal, doctor: undefined }).author).toBe("-");
  });

  it("status: finalized → type '確定済'", () => {
    expect(transformToHistoryItem({ ...minimal, status: "finalized" }).type).toBe("確定済");
  });

  it("status: draft → type '作成中'", () => {
    expect(transformToHistoryItem({ ...minimal, status: "draft" }).type).toBe("作成中");
  });

  it("chief_complaint と notes を改行で結合して content にする", () => {
    const result = transformToHistoryItem({
      ...minimal,
      inquiry: {
        chief_complaint: "元気がない",
        notes: "食欲なし",
      } as BackendMedicalRecord["inquiry"],
    });
    expect(result.content).toBe("元気がない\n食欲なし");
  });

  it("chief_complaint のみのとき content は chief_complaint", () => {
    const result = transformToHistoryItem({
      ...minimal,
      inquiry: {
        chief_complaint: "嘔吐",
        notes: "",
      } as BackendMedicalRecord["inquiry"],
    });
    expect(result.content).toBe("嘔吐");
  });

  it("chief_complaint も notes もないとき content は '（記録なし）'", () => {
    const result = transformToHistoryItem({
      ...minimal,
      inquiry: { chief_complaint: "", notes: "" } as BackendMedicalRecord["inquiry"],
    });
    expect(result.content).toBe("（記録なし）");
  });

  it("inquiry が未設定のとき content は '（記録なし）'", () => {
    expect(transformToHistoryItem({ ...minimal, inquiry: undefined }).content).toBe("（記録なし）");
  });

  it("title は chief_complaint、なければ record_no", () => {
    expect(
      transformToHistoryItem({
        ...minimal,
        inquiry: { chief_complaint: "発熱", notes: "" } as BackendMedicalRecord["inquiry"],
      }).title
    ).toBe("発熱");

    expect(
      transformToHistoryItem({
        ...minimal,
        inquiry: { chief_complaint: "", notes: "" } as BackendMedicalRecord["inquiry"],
      }).title
    ).toBe("MR-001");
  });

  it("date を YYYY/MM/DD 形式にフォーマットする", () => {
    expect(transformToHistoryItem({ ...minimal, date: "2026-03-25T00:00:00Z" }).date).toBe("2026/03/25");
  });
});

// ─────────────────────────────────────────────────────────────
// toBackendMedicalRecordStatus（BUG-B1: server-side status フィルタ変換）
// ─────────────────────────────────────────────────────────────
describe("toBackendMedicalRecordStatus", () => {
  it("'作成中' → 'draft'", () => {
    expect(toBackendMedicalRecordStatus("作成中")).toBe("draft");
  });

  it("'確定済' → 'finalized'", () => {
    expect(toBackendMedicalRecordStatus("確定済")).toBe("finalized");
  });

  it("未知のラベルは undefined を返す", () => {
    expect(toBackendMedicalRecordStatus("不明")).toBeUndefined();
  });
});
