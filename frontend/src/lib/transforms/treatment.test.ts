import { describe, it, expect } from "vitest";
import {
  transformConsultation,
  transformExaminationType,
  transformProcedure,
  transformVaccine,
  type ConsultationWire,
  type ProcedureWire,
} from "./treatment";
import type { Consultation, ExaminationType, Procedure } from "@/types/generated/models";

const minimalExamType: ExaminationType = {
  id: 1,
  name: "血液検査",
  clinic_id: 1,
  description: "",
  sort_order: 0,
  is_active: true,
  is_non_insurance: false,
  created_at: "",
  updated_at: "",
};

describe("transformExaminationType", () => {
  it("is_non_insurance=false をそのままマップする", () => {
    const result = transformExaminationType(minimalExamType);
    expect(result.isNonInsurance).toBe(false);
  });

  it("is_non_insurance=true をそのままマップする", () => {
    const result = transformExaminationType({ ...minimalExamType, is_non_insurance: true });
    expect(result.isNonInsurance).toBe(true);
  });

  it("id を文字列に変換する", () => {
    expect(transformExaminationType(minimalExamType).id).toBe("1");
  });

  it("name をそのまま返す", () => {
    expect(transformExaminationType(minimalExamType).name).toBe("血液検査");
  });
});

// ─────────────────────────────────────────────────
// EMR-454: 診察・処置の is_non_insurance は nullable（null=旧欠落）
// ─────────────────────────────────────────────────

const minimalConsultation: Consultation = {
  id: 10,
  clinic_id: 1,
  name: "初診",
  price: 1500,
  is_active: true,
  description: "",
  time_condition: "",
  tax_type: "excluded",
  tax_rate: 0.1,
  sort_order: 0,
  created_at: "",
  updated_at: "",
};

const minimalProcedure: Procedure = {
  id: 20,
  clinic_id: 1,
  name: "処置A",
  price: 800,
  is_active: true,
  description: "",
  anesthesia: "none",
  is_surgery: false,
  tax_type: "excluded",
  tax_rate: 0.1,
  sort_order: 0,
  created_at: "",
  updated_at: "",
};

describe("transformConsultation / transformProcedure の is_non_insurance", () => {
  it("明示 false を保険対象として保持する", () => {
    expect(
      transformConsultation({ ...minimalConsultation, is_non_insurance: false }).isNonInsurance,
    ).toBe(false);
    expect(
      transformProcedure({ ...minimalProcedure, is_non_insurance: false }).isNonInsurance,
    ).toBe(false);
  });

  it("明示 true を保険対象外として保持する", () => {
    expect(
      transformConsultation({ ...minimalConsultation, is_non_insurance: true }).isNonInsurance,
    ).toBe(true);
    expect(transformProcedure({ ...minimalProcedure, is_non_insurance: true }).isNonInsurance).toBe(
      true,
    );
  });

  it("旧欠落（カラム自体が無い応答）を false に補完しない", () => {
    expect(transformConsultation(minimalConsultation).isNonInsurance).toBeNull();
    expect(transformProcedure(minimalProcedure).isNonInsurance).toBeNull();
  });

  it("null 応答を false に補完しない", () => {
    expect(
      transformConsultation({ ...minimalConsultation, is_non_insurance: null }).isNonInsurance,
    ).toBeNull();
    expect(
      transformProcedure({ ...minimalProcedure, is_non_insurance: null }).isNonInsurance,
    ).toBeNull();
  });
});

describe("master editing version read roundtrip", () => {
  it("retains consultation/procedure version and leaves missing legacy version unavailable", () => {
    expect(transformConsultation({ ...minimalConsultation, version: 7 }).version).toBe(7);
    expect(transformProcedure({ ...minimalProcedure, version: 8 }).version).toBe(8);
    expect(transformConsultation(minimalConsultation).version).toBeNull();
    expect(transformProcedure(minimalProcedure).version).toBeNull();
  });
  it("retains vaccine version without filling a missing version", () => {
    const wire = {
      id: 1,
      clinic_id: 7,
      name: "vaccine",
      is_active: true,
      description: "",
      interval: "",
      sort_order: 1,
      is_non_insurance: false,
      created_at: "",
      updated_at: "",
    };
    expect(transformVaccine({ ...wire, version: 9 }).version).toBe(9);
    expect(transformVaccine(wire).version).toBeNull();
  });
});

describe("nullable READ wire and display defaults", () => {
  it("accepts consultation null sort order without mutating its wire value", () => {
    const wire: ConsultationWire = { ...minimalConsultation, sort_order: null };
    expect(transformConsultation(wire).sortOrder).toBe(0);
    expect(wire.sort_order).toBeNull();
  });
  it("accepts procedure null sort order/anesthesia without rewriting wire values", () => {
    const wire: ProcedureWire = { ...minimalProcedure, sort_order: null, anesthesia: null };
    expect(transformProcedure(wire)).toMatchObject({ sortOrder: 0, anesthesia: "none" });
    expect(wire.sort_order).toBeNull();
    expect(wire.anesthesia).toBeNull();
  });
});
