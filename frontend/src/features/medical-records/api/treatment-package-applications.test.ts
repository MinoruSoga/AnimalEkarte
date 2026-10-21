import { describe, expect, it } from "vitest";
import {
  applicationSchema,
  operationResultSchema,
  planListSchema,
  treatmentPackageApplySchema,
  treatmentPackagePreviewSchema,
} from "./treatment-package-applications";

const uuid = "a0000000-0000-4000-8000-000000000001";
const token = `pkg-preview-v1:${"a".repeat(64)}`;

describe("apply body（平坦EMR494形・clinic/price/items override なし）", () => {
  const body = {
    operation_key: uuid,
    edit_session_ref: "a0000000-0000-4000-8000-000000000002",
    edit_epoch: "9",
    owner_expected_version: "17",
    package_id: "201",
    package_version: "3",
    preview_token: token,
  };
  it("契約形を受理し、clinic_id/items/price の混入を拒否", () => {
    expect(treatmentPackageApplySchema.safeParse(body).success).toBe(true);
    for (const extra of [{ clinic_id: "7" }, { items: [] }, { price: "0" }, { actor_id: "1" }]) {
      expect(treatmentPackageApplySchema.safeParse({ ...body, ...extra }).success).toBe(false);
    }
  });
  it("非canonical token/version を拒否", () => {
    expect(treatmentPackageApplySchema.safeParse({ ...body, preview_token: "bogus" }).success).toBe(
      false,
    );
    expect(treatmentPackageApplySchema.safeParse({ ...body, edit_epoch: "0" }).success).toBe(false);
  });
});

describe("operation result（C06 wire・非成功を成功にしない）", () => {
  it("succeeded は refs/versions/time を保持", () => {
    const ok = {
      status: "succeeded",
      operation_key: uuid,
      result_refs: { record_id: "91", application_id: "302" },
      committed_versions: { application: "1", plan_items: "18" },
      recorded_at: "t",
    };
    expect(operationResultSchema.safeParse(ok).success).toBe(true);
  });
  it("not-succeeded 系は status+key のみ（refs混入は拒否）", () => {
    for (const s of ["not-succeeded", "operation-in-progress", "result-unavailable"]) {
      const r = operationResultSchema.safeParse({ status: s, operation_key: uuid });
      expect(r.success).toBe(true);
      if (r.success) expect("result_refs" in r.data).toBe(false);
    }
    expect(
      operationResultSchema.safeParse({
        status: "not-succeeded",
        operation_key: uuid,
        result_refs: {},
      }).success,
    ).toBe(false);
  });
});

describe("preview/plan read（fail-closed・文字列契約）", () => {
  const price = {
    unit_price: null,
    tax_type: null,
    tax_rate: null,
    is_non_insurance: null,
    insurance_setting_state: "legacy_unset",
    price_source: "unresolved",
    tax_source: "unresolved",
    insurance_source: "unresolved",
    amount_basis: "unresolved",
    uncertainty: [],
  };
  it("version_state=uninitialized は collection_version=null", () => {
    const p = {
      record_id: "91",
      clinic_id: "7",
      pet_id: "801",
      package_id: "201",
      package_version: "3",
      plan_collection_version: null,
      version_state: "uninitialized",
      preview_token: token,
      applicable: false,
      items: [
        {
          component_key: uuid,
          item: {
            kind: "manual_other",
            manual: {
              name: "n",
              reason: "r",
              unit_label: null,
              unit_price: "110",
              tax_type: "included",
              tax_rate: "0.1",
              is_non_insurance: true,
            },
          },
          quantity: "1",
          name: "n",
          unit_label: null,
          source_stamp: {
            native_version: null,
            dependency_token: `pkg-source-v1:${"0".repeat(64)}`,
          },
          price,
          classification: "unresolved",
          issues: [],
        },
      ],
      total: { status: "unresolved", amount: null },
    };
    expect(treatmentPackagePreviewSchema.safeParse(p).success).toBe(true);
    expect(treatmentPackagePreviewSchema.safeParse({ ...p, pet_id: 801 }).success).toBe(false);
  });
  it("plan list: commercial=selected・clinical=not_performed・owner_links を保持", () => {
    const list = {
      collection_version: "18",
      version_state: "ready",
      items: [
        {
          id: "401",
          version: "1",
          record_id: "91",
          clinic_id: "7",
          pet_id: "801",
          item: { kind: "consultation", consultation_id: "101" },
          name_snapshot: "診察",
          unit_label_snapshot: null,
          quantity: "1",
          price_snapshot: price,
          origin: {
            kind: "package",
            application_id: "302",
            package_id: "201",
            package_version: "3",
            component_key: uuid,
          },
          execution: {
            kind: "clinical",
            state: "not_performed",
            revision_id: "501",
            result_state: "none",
          },
          owner_links: [],
        },
        {
          id: "402",
          version: "1",
          record_id: "91",
          clinic_id: "7",
          pet_id: "801",
          item: { kind: "merchandise_item", merchandise_item_id: "107" },
          name_snapshot: "物販",
          unit_label_snapshot: null,
          quantity: "2",
          price_snapshot: price,
          origin: { kind: "single" },
          execution: { kind: "commercial", state: "selected", revision_id: "502" },
          owner_links: [
            {
              id: "601",
              owner_kind: "billing_item",
              owner_id: "701",
              owner_version: null,
              fixed_revision_id: null,
              status: "active",
            },
          ],
        },
      ],
    };
    expect(planListSchema.safeParse(list).success).toBe(true);
  });
  it("application は snapshot items + operation_key を保持", () => {
    const app = {
      id: "302",
      record_id: "91",
      package_id: "201",
      package_version: "3",
      package_name_snapshot: "セット",
      applied_at: "t",
      item_count: 1,
      clinic_id: "7",
      pet_id: "801",
      version: "1",
      operation_key: uuid,
      items: [
        {
          component_key: uuid,
          plan_item_id: "401",
          applied_item: {
            component_key: uuid,
            item: { kind: "consultation", consultation_id: "101" },
            quantity: "1",
            name: "診察",
            unit_label: null,
            source_stamp: {
              native_version: "2",
              dependency_token: `pkg-source-v1:${"0".repeat(64)}`,
            },
            price,
            classification: "clinical",
          },
        },
      ],
    };
    expect(applicationSchema.safeParse(app).success).toBe(true);
  });
});
