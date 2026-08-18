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
  version: 1,
  visit_count: 0,
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
    expect(transformMedicalRecord({ ...minimal, status: "unknown" }).status).toBe("作成中");
  });

  it("owner.name を ownerName にマップする", () => {
    const result = transformMedicalRecord({
      ...minimal,
      owner: { id: 20, name: "田中太郎" },
    });
    expect(result.ownerName).toBe("田中太郎");
  });

  it("pet.name を petName にマップする", () => {
    const result = transformMedicalRecord({
      ...minimal,
      pet: { id: 10, name: "ポチ", pet_number: "P-1" },
    });
    expect(result.petName).toBe("ポチ");
  });

  it("pet.status=deceased は明示的な死亡状態へ正規化する", () => {
    const result = transformMedicalRecord({
      ...minimal,
      pet: {
        id: 10,
        name: "ポチ",
        pet_number: "P-1",
        status: "deceased",
      },
    });
    expect(result.petIsDeceased).toBe(true);
  });

  it("pet.status がaliveまたは未取得なら死亡扱いにしない", () => {
    expect(
      transformMedicalRecord({
        ...minimal,
        pet: {
          id: 10,
          name: "ポチ",
          pet_number: "P-1",
          status: "alive",
        },
      }).petIsDeceased
    ).toBe(false);
    expect(transformMedicalRecord({ ...minimal, pet: undefined }).petIsDeceased).toBe(false);
  });

  it("doctor.name を doctor にマップする", () => {
    const result = transformMedicalRecord({
      ...minimal,
      doctor: { id: 3, name: "山田医師" },
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

  it("wire に clinical_plan が無いため objective/assessment/plan は undefined（clinical-plan API が正本）", () => {
    const result = transformMedicalRecord(minimal);
    expect(result.objective).toBeUndefined();
    expect(result.assessment).toBeUndefined();
    expect(result.plan).toBeUndefined();
  });

  it("inquiry.chief_complaint を chiefComplaint にマップする", () => {
    const result = transformMedicalRecord({
      ...minimal,
      inquiry: { id: 1, chief_complaint: "元気がない" },
    });
    expect(result.chiefComplaint).toBe("元気がない");
  });

  it("InquirySummary の chief_complaint_type_id を chiefComplaintTypeId にマップする（BUG-013）", () => {
    const result = transformMedicalRecord({
      ...minimal,
      inquiry: { id: 1, chief_complaint: "", chief_complaint_type_id: 5 },
    });
    expect(result.chiefComplaintTypeId).toBe(5);
  });

  it("visit_type first/revisit を日本語ラベルにマップする（BUG-012）", () => {
    expect(transformMedicalRecord({ ...minimal, visit_type: "first" }).visitType).toBe("初診");
    expect(transformMedicalRecord({ ...minimal, visit_type: "revisit" }).visitType).toBe("再診");
  });

  it("wire に clinical_plan が無いため diagnosis*Id は null（clinical-plan API が正本）", () => {
    const result = transformMedicalRecord(minimal);
    expect(result.diagnosis1CategoryId).toBeNull();
    expect(result.diagnosis1NameId).toBeNull();
    expect(result.diagnosis2CategoryId).toBeNull();
    expect(result.diagnosis2NameId).toBeNull();
  });

  it("visit_count をそのまま返す", () => {
    expect(transformMedicalRecord({ ...minimal, visit_count: 5 }).visitCount).toBe(5);
  });

  it("version を wire からそのまま返す（既定値フォールバック無し）", () => {
    expect(transformMedicalRecord({ ...minimal, version: 3 }).version).toBe(3);
    expect(transformMedicalRecord({ ...minimal, version: 1 }).version).toBe(1);
  });

  it("pet.animal_species.name を species にマップする", () => {
    const result = transformMedicalRecord({
      ...minimal,
      pet: {
        id: 10,
        name: "ポチ",
        pet_number: "P-1",
        animal_species: { id: 1, name: "犬" },
      },
    });
    expect(result.species).toBe("犬");
  });

  it("inquiry.notes を notes にマップする（問診治療方針・BUG-034）", () => {
    const result = transformMedicalRecord({
      ...minimal,
      inquiry: { id: 1, chief_complaint: "x", notes: "UAT再検証 治療方針" },
    });
    expect(result.notes).toBe("UAT再検証 治療方針");
  });

  it("inquiry.notes が空のとき notes は undefined（DEFAULT 表示は form state）", () => {
    const result = transformMedicalRecord({
      ...minimal,
      inquiry: { id: 1, chief_complaint: "x", notes: "" },
    });
    expect(result.notes).toBeUndefined();
  });

  it("inquiry 未設定のとき notes は undefined", () => {
    const result = transformMedicalRecord({
      ...minimal,
      inquiry: undefined,
    });
    expect(result.notes).toBeUndefined();
  });
});

// ─────────────────────────────────────────────────────────────
// transformToHistoryItem
// ─────────────────────────────────────────────────────────────
describe("transformToHistoryItem", () => {
  const minimalHistory: BackendMedicalRecord = {
    id: 1,
    clinic_id: 1,
    pet_id: 10,
    owner_id: 20,
    doctor_id: 3,
    record_no: "MR-001",
    date: "2026-03-25T00:00:00Z",
    status: "draft",
    version: 1,
    visit_count: 0,
    created_at: "2026-03-25T00:00:00Z",
    updated_at: "2026-03-25T00:00:00Z",
  };

  it("id を string に変換する", () => {
    expect(transformToHistoryItem({ ...minimalHistory, id: 42 }).id).toBe("42");
  });

  it("doctor.name を author にマップする", () => {
    const result = transformToHistoryItem({
      ...minimalHistory,
      doctor: { id: 3, name: "山田医師" },
    });
    expect(result.author).toBe("山田医師");
  });

  it("doctor が未設定のとき author は '-'", () => {
    expect(transformToHistoryItem({ ...minimalHistory, doctor: undefined }).author).toBe("-");
  });

  it("status: finalized → type '確定済'", () => {
    expect(transformToHistoryItem({ ...minimalHistory, status: "finalized" }).type).toBe("確定済");
  });

  it("status: draft → type '作成中'", () => {
    expect(transformToHistoryItem({ ...minimalHistory, status: "draft" }).type).toBe("作成中");
  });

  it("chief_complaint を content にする（InquirySummary に notes は無い）", () => {
    const result = transformToHistoryItem({
      ...minimalHistory,
      inquiry: {
        id: 1,
        chief_complaint: "元気がない",
      },
    });
    expect(result.content).toBe("元気がない");
  });

  it("chief_complaint がないとき content は '（記録なし）'", () => {
    const result = transformToHistoryItem({
      ...minimalHistory,
      inquiry: {
        id: 1,
        chief_complaint: "",
      },
    });
    expect(result.content).toBe("（記録なし）");
  });

  it("inquiry が未設定のとき content は '（記録なし）'", () => {
    expect(transformToHistoryItem({ ...minimalHistory, inquiry: undefined }).content).toBe(
      "（記録なし）"
    );
  });

  it("title は chief_complaint、なければ record_no", () => {
    expect(
      transformToHistoryItem({
        ...minimalHistory,
        inquiry: { id: 1, chief_complaint: "発熱" },
      }).title
    ).toBe("発熱");

    expect(
      transformToHistoryItem({
        ...minimalHistory,
        inquiry: { id: 1, chief_complaint: "" },
      }).title
    ).toBe("MR-001");
  });

  it("date を YYYY/MM/DD 形式にフォーマットする", () => {
    expect(transformToHistoryItem({ ...minimalHistory, date: "2026-03-25T00:00:00Z" }).date).toBe(
      "2026/03/25"
    );
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
