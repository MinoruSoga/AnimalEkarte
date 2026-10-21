import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import type { TreatmentPackagePreview as Preview } from "../api/treatment-package-applications";
import { TreatmentPackagePreview } from "./TreatmentPackagePreview";

const item = {
  component_key: "a0000000-0000-4000-8000-000000000011",
  item: { kind: "consultation", consultation_id: "101" } as const,
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
    insurance_setting_state: "legacy_unset" as const,
    price_source: "unresolved" as const,
    tax_source: "unresolved" as const,
    insurance_source: "unresolved" as const,
    amount_basis: "unresolved" as const,
    uncertainty: [{ field: "unit_price", reason: "legacy_unset", required_at: "billing" } as const],
  },
  issues: [] as never[],
};

const base: Preview = {
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
      ...item,
      classification: "clinical" as const,
      issues: [
        {
          component_key: item.component_key,
          severity: "warning" as const,
          required_at: "billing" as const,
          code: "price-unresolved" as const,
          message: "価格が未確定です",
        },
      ],
    },
  ],
};

const render_ = (preview: Preview, canApply = true) =>
  render(
    <TreatmentPackagePreview
      preview={preview}
      canApply={canApply}
      applyDisabledReason={canApply ? undefined : "不可"}
      applying={false}
      onApply={vi.fn()}
      onClose={vi.fn()}
    />,
  );

describe("TreatmentPackagePreview（行別警告/blocking・total null）", () => {
  it("clinical=未実施表示・財務unknownは警告・total null を 0 にしない", () => {
    render_(base);
    expect(screen.getByText(/未実施として追加/)).toBeTruthy();
    expect(screen.getByText(/警告: 価格が未確定です/)).toBeTruthy();
    expect(screen.getByText(/合計は未確定です/)).toBeTruthy();
    expect(screen.queryByText(/¥0/)).toBeNull();
  });

  it("blocking 行は追加不可と表示し適用を disabled のままにできる", () => {
    const p: Preview = {
      ...base,
      applicable: false,
      items: [
        {
          ...base.items[0]!,
          classification: "unresolved",
          issues: [
            {
              component_key: item.component_key,
              severity: "blocking",
              required_at: "selection",
              code: "owner-contract-unresolved",
              message: "分類が未解決です",
            },
          ],
        },
      ],
    };
    render_(p, false);
    expect(screen.getByText(/追加不可: 分類が未解決です/)).toBeTruthy();
    expect((screen.getByText("この内容で追加") as HTMLButtonElement).disabled).toBe(true);
  });

  it("commercial=selected（選択明細）として表示", () => {
    const p: Preview = {
      ...base,
      items: [{ ...base.items[0]!, classification: "commercial", issues: [] }],
    };
    render_(p);
    expect(screen.getByText(/選択明細として追加/)).toBeTruthy();
  });
});
