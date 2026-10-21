import { describe, expect, it } from "vitest";
import type { TreatmentPackagePreview } from "../api/treatment-package-applications";
import {
  TREATMENT_PACKAGE_RUNTIME_AVAILABLE,
  buildApplyCommand,
  canSubmitApply,
  classifyPreview,
  resolveOperationResult,
  scopeKey,
  type TreatmentPackageEditContext,
} from "./treatment-package-operation-state";

const previewBase: TreatmentPackagePreview = {
  record_id: "91",
  clinic_id: "7",
  pet_id: "801",
  package_id: "201",
  package_version: "3",
  plan_collection_version: "17",
  version_state: "ready",
  preview_token: `pkg-preview-v1:${"a".repeat(64)}`,
  applicable: true,
  total: { status: "unresolved", amount: null },
  items: [
    {
      component_key: "a0000000-0000-4000-8000-000000000011",
      item: { kind: "consultation", consultation_id: "101" },
      quantity: "1",
      name: "診察",
      unit_label: null,
      source_stamp: {
        native_version: "2",
        dependency_token: `pkg-source-v1:${"0".repeat(64)}`,
      },
      price: {
        unit_price: null,
        tax_type: null,
        tax_rate: null,
        is_non_insurance: null,
        insurance_setting_state: "legacy_unset",
        price_source: "unresolved",
        tax_source: "unresolved",
        insurance_source: "unresolved",
        amount_basis: "unresolved",
        uncertainty: [{ field: "unit_price", reason: "legacy_unset", required_at: "billing" }],
      },
      classification: "clinical",
      issues: [
        {
          component_key: "a0000000-0000-4000-8000-000000000011",
          severity: "warning",
          required_at: "billing",
          code: "price-unresolved",
          message: "価格未確定",
        },
      ],
    },
  ],
};

const ctx: TreatmentPackageEditContext = {
  editSessionRef: "a0000000-0000-4000-8000-000000000002",
  editEpoch: "9",
  ownerExpectedVersion: "17",
};

describe("classifyPreview（財務unknown=警告・owner未解決=blocking）", () => {
  it("財務 unknown は warning のまま applicable。total null を 0 にしない", () => {
    const r = classifyPreview(previewBase);
    expect(r.applicable).toBe(true);
    expect(r.warnings).toHaveLength(1);
    expect(r.blocking).toHaveLength(0);
    expect(previewBase.total.amount).toBeNull();
  });
  it("owner-contract-unresolved は blocking で applicable:false", () => {
    const p: TreatmentPackagePreview = {
      ...previewBase,
      applicable: false,
      items: [
        {
          ...previewBase.items[0]!,
          classification: "unresolved",
          issues: [
            {
              component_key: previewBase.items[0]!.component_key,
              severity: "blocking",
              required_at: "selection",
              code: "owner-contract-unresolved",
              message: "分類未解決",
            },
          ],
        },
      ],
    };
    const r = classifyPreview(p);
    expect(r.applicable).toBe(false);
    expect(r.blocking).toHaveLength(1);
  });
});

describe("canSubmitApply / buildApplyCommand（明示apply時のみkey生成）", () => {
  it("runtime 未提供・viewのみ・context なし・not applicable は送らない", () => {
    expect(
      canSubmitApply({
        runtimeAvailable: false,
        editContext: ctx,
        preview: previewBase,
        canCreate: true,
      }).ok,
    ).toBe(false);
    expect(
      canSubmitApply({
        runtimeAvailable: true,
        editContext: null,
        preview: previewBase,
        canCreate: true,
      }).ok,
    ).toBe(false);
    expect(
      canSubmitApply({
        runtimeAvailable: true,
        editContext: ctx,
        preview: previewBase,
        canCreate: false,
      }).ok,
    ).toBe(false);
    expect(
      canSubmitApply({
        runtimeAvailable: true,
        editContext: ctx,
        preview: { ...previewBase, applicable: false },
        canCreate: true,
      }).ok,
    ).toBe(false);
    expect(
      canSubmitApply({
        runtimeAvailable: true,
        editContext: ctx,
        preview: previewBase,
        canCreate: true,
      }).ok,
    ).toBe(true);
  });
  it("body は平坦形で preview_token/package版/CASを保持", () => {
    const cmd = buildApplyCommand({
      preview: previewBase,
      editContext: ctx,
      scope: { recordId: "91" },
      generateKey: () => "a0000000-0000-4000-8000-000000000001",
    });
    expect(cmd.body).toEqual({
      operation_key: "a0000000-0000-4000-8000-000000000001",
      edit_session_ref: ctx.editSessionRef,
      edit_epoch: "9",
      owner_expected_version: "17",
      package_id: "201",
      package_version: "3",
      preview_token: previewBase.preview_token,
    });
  });
});

describe("resolveOperationResult（非成功を成功にしない）", () => {
  const ok = {
    status: "succeeded" as const,
    operation_key: "a0000000-0000-4000-8000-000000000001",
    result_refs: { record_id: "91", application_id: "302" },
    committed_versions: { application: "1", plan_items: "18" },
    recorded_at: "t",
  };
  it("succeeded/replayed のみ成功", () => {
    expect(resolveOperationResult(ok).kind).toBe("success");
    expect(resolveOperationResult({ ...ok, status: "replayed" }).kind).toBe("success");
  });
  it("in-progress/result-unavailable は同key照会継続", () => {
    for (const s of ["operation-in-progress", "result-unavailable"] as const) {
      expect(resolveOperationResult({ status: s, operation_key: ok.operation_key }).kind).toBe(
        "retry-lookup",
      );
    }
  });
  it("not-succeeded/conflict群は失敗確定（成功にしない）", () => {
    for (const s of [
      "not-succeeded",
      "version-conflict",
      "edit-right-lost",
      "dependency-changed",
    ] as const) {
      const r = resolveOperationResult({ status: s, operation_key: ok.operation_key });
      expect(r.kind).toBe("failed");
    }
  });
});

describe("scope（切替後の遅延成功を混ぜない）・runtime gate", () => {
  it("scopeKey は clinic/pet/record を含む", () => {
    expect(scopeKey({ recordId: "91", clinicId: "7", petId: "801" })).toBe("7/801/91");
    expect(scopeKey({ recordId: "91" })).not.toBe("7/801/91");
  });
  it("runtime 既定は false（環境変数で自動有効化しない定数）", () => {
    expect(TREATMENT_PACKAGE_RUNTIME_AVAILABLE).toBe(false);
  });
});
