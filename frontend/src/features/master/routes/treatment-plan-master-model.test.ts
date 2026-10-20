import { describe, expect, it } from "vitest";

import type { TreatmentFormData } from "../components/TreatmentItemSidePanel";
import {
  applyItemInsuranceForCreate,
  applyItemInsuranceForUpdate,
  buildCheckupCreateRequest,
  buildCheckupUpdateRequest,
  buildConsultationCreateRequest,
  buildConsultationUpdateRequest,
  buildExaminationCreateRequest,
  buildExaminationUpdateRequest,
  buildProcedureCreateRequest,
  buildProcedureUpdateRequest,
  buildVaccineCreateRequest,
  buildVaccineUpdateRequest,
} from "./treatment-plan-master-model";

function makeFormData(overrides: Partial<TreatmentFormData> = {}): TreatmentFormData {
  return {
    name: "V04処置テスト",
    price: 500,
    description: "",
    isActive: true,
    taxType: "excluded",
    taxRate: 0.1,
    isNonInsurance: false,
    anesthesia: "none",
    ...overrides,
  };
}

describe("buildConsultationCreateRequest", () => {
  it("buildConsultationCreateRequest persists price 0 and tax_type", () => {
    const request = buildConsultationCreateRequest(
      makeFormData({ name: "再診", price: 0, taxType: "excluded", taxRate: 0.1 }),
    );
    expect(request.price).toBe(0);
    expect(request.tax_type).toBe("excluded");
    expect(request.tax_rate).toBe(0.1);
  });
});

describe("buildConsultationUpdateRequest", () => {
  it("persists edited price 1200→2300 with tax for reread-shaped update", () => {
    const request = buildConsultationUpdateRequest(
      makeFormData({ name: "再診", price: 2300, taxType: "included", taxRate: 0.08 }),
    );
    expect(request.price).toBe(2300);
    expect(request.tax_type).toBe("included");
    expect(request.tax_rate).toBe(0.08);
  });
});

describe("buildExaminationCreateRequest", () => {
  it("persists price 0 and 1200 without tax fields", () => {
    const zero = buildExaminationCreateRequest(
      makeFormData({ name: "血算", price: 0, isNonInsurance: true, taxType: "included" }),
    );
    expect(zero.price).toBe(0);
    expect(zero.is_non_insurance).toBe(true);
    expect(zero).not.toHaveProperty("tax_type");
    expect(zero).not.toHaveProperty("tax_rate");

    const priced = buildExaminationCreateRequest(makeFormData({ name: "血算", price: 1200 }));
    expect(priced.price).toBe(1200);
    expect(priced).not.toHaveProperty("tax_type");
  });
});

describe("buildExaminationUpdateRequest", () => {
  it("persists edited examination price without leaking tax", () => {
    const request = buildExaminationUpdateRequest(
      makeFormData({ name: "血算", price: 2300, taxType: "included", taxRate: 0.08 }),
    );
    expect(request.price).toBe(2300);
    expect(request).not.toHaveProperty("tax_type");
    expect(request).not.toHaveProperty("tax_rate");
  });
});

describe("buildProcedureCreateRequest", () => {
  it.each([
    ["none", "none"],
    ["local", "local"],
    ["sedation", "sedation"],
    ["general", "general"],
  ] as const)("sends anesthesia=%s when form selects %s", (selected, _expected) => {
    const request = buildProcedureCreateRequest(makeFormData({ anesthesia: selected }));
    expect(request.anesthesia).toBe(selected);
    expect(request.name).toBe("V04処置テスト");
    expect(request.price).toBe(500);
  });

  it("includes tax and parent fields alongside anesthesia", () => {
    const request = buildProcedureCreateRequest(
      makeFormData({
        anesthesia: "general",
        taxType: "included",
        taxRate: 0.08,
        parentId: "12",
        description: "備考",
      }),
    );
    expect(request).toMatchObject({
      name: "V04処置テスト",
      price: 500,
      anesthesia: "general",
      tax_type: "included",
      tax_rate: 0.08,
      parent_id: 12,
      description: "備考",
      is_active: true,
    });
  });

  it("persists price 0 with tax without changing anesthesia contract", () => {
    const request = buildProcedureCreateRequest(
      makeFormData({ price: 0, anesthesia: "local", taxType: "excluded", taxRate: 0.1 }),
    );
    expect(request.price).toBe(0);
    expect(request.anesthesia).toBe("local");
    expect(request.tax_type).toBe("excluded");
    expect(request.tax_rate).toBe(0.1);
  });
});

describe("buildProcedureUpdateRequest", () => {
  it("forwards anesthesia change and clears parent when empty string", () => {
    const request = buildProcedureUpdateRequest(
      makeFormData({ anesthesia: "local", parentId: "" }),
    );
    expect(request.anesthesia).toBe("local");
    expect(request.clear_parent_id).toBe(true);
  });

  it("persists edited procedure price 2300 with tax", () => {
    const request = buildProcedureUpdateRequest(
      makeFormData({ price: 2300, taxType: "included", taxRate: 0.08, anesthesia: "none" }),
    );
    expect(request.price).toBe(2300);
    expect(request.tax_type).toBe("included");
    expect(request.tax_rate).toBe(0.08);
  });
});

describe("buildVaccineCreateRequest", () => {
  it("persists vaccine price 0 and 1200 without tax fields", () => {
    const zero = buildVaccineCreateRequest(makeFormData({ name: "混合ワクチン", price: 0 }));
    expect(zero.price).toBe(0);
    expect(zero).not.toHaveProperty("tax_type");

    const priced = buildVaccineCreateRequest(makeFormData({ name: "混合ワクチン", price: 1200 }));
    expect(priced.price).toBe(1200);
  });
});

describe("buildVaccineUpdateRequest", () => {
  it("persists edited vaccine price for reread-shaped update", () => {
    const request = buildVaccineUpdateRequest(makeFormData({ name: "混合ワクチン", price: 2300 }));
    expect(request.price).toBe(2300);
  });
});

describe("buildCheckupCreateRequest", () => {
  it("persists checkup price 0 and 1200 without tax or billing id fields", () => {
    const zero = buildCheckupCreateRequest(makeFormData({ name: "年次健診", price: 0 }));
    expect(zero.price).toBe(0);
    expect(zero).not.toHaveProperty("tax_type");
    expect(zero).not.toHaveProperty("checkup_id");

    const priced = buildCheckupCreateRequest(makeFormData({ name: "年次健診", price: 1200 }));
    expect(priced.price).toBe(1200);
  });
});

// ─────────────────────────────────────────────────
// EMR-454: 品目別保険設定（create=対象外初期値 / update=未指定・null・false・true の区別）
// ─────────────────────────────────────────────────

describe("applyItemInsuranceForCreate", () => {
  it("新規登録は未設定(null)でも保険対象外(true)を明示する", () => {
    const request = applyItemInsuranceForCreate(
      buildConsultationCreateRequest(makeFormData()),
      null,
    );
    expect(request.is_non_insurance).toBe(true);
  });

  it("処置も同じく新規登録は保険対象外(true)を明示する", () => {
    const request = applyItemInsuranceForCreate(buildProcedureCreateRequest(makeFormData()), null);
    expect(request.is_non_insurance).toBe(true);
  });

  it("利用者が選んだ false(保険対象)を上書きしない", () => {
    const request = applyItemInsuranceForCreate(
      buildConsultationCreateRequest(makeFormData()),
      false,
    );
    expect(request.is_non_insurance).toBe(false);
  });
});

describe("applyItemInsuranceForUpdate", () => {
  it("未指定(undefined)は is_non_insurance を送らない", () => {
    const request = applyItemInsuranceForUpdate(
      buildConsultationUpdateRequest(makeFormData()),
      undefined,
    );
    expect(request).not.toHaveProperty("is_non_insurance");
  });

  it("旧欠落(null)は送らず、既存値を反転させない", () => {
    const request = applyItemInsuranceForUpdate(buildProcedureUpdateRequest(makeFormData()), null);
    expect(request).not.toHaveProperty("is_non_insurance");
    expect(request.name).toBe("V04処置テスト");
  });

  it("明示 false を送る", () => {
    const request = applyItemInsuranceForUpdate(
      buildConsultationUpdateRequest(makeFormData()),
      false,
    );
    expect(request.is_non_insurance).toBe(false);
  });

  it("明示 true を送る", () => {
    const request = applyItemInsuranceForUpdate(buildProcedureUpdateRequest(makeFormData()), true);
    expect(request.is_non_insurance).toBe(true);
  });

  it("親クリア指定を保険設定の付与で落とさない", () => {
    const request = applyItemInsuranceForUpdate(
      buildConsultationUpdateRequest(makeFormData({ parentId: "" })),
      true,
    );
    expect(request.clear_parent_id).toBe(true);
    expect(request.is_non_insurance).toBe(true);
  });
});

describe("checkup editing contract", () => {
  it("requires the captured editing version instead of substituting a current version", () => {
    expect(() => buildCheckupUpdateRequest(makeFormData())).toThrow("編集開始時");
    const req = buildCheckupUpdateRequest(makeFormData({ version: 7, isNonInsurance: false }));
    expect(req.version).toBe(7);
    expect(applyItemInsuranceForUpdate(req, false).is_non_insurance).toBe(false);
  });
  it("omits an unchanged or legacy insurance value", () => {
    expect(applyItemInsuranceForUpdate({ name: "A" }, undefined)).not.toHaveProperty(
      "is_non_insurance",
    );
    expect(applyItemInsuranceForUpdate({ name: "A" }, null)).not.toHaveProperty("is_non_insurance");
  });
});
