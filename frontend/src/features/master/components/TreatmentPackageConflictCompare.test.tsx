import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { TreatmentPackageConflictCompare } from "./TreatmentPackageConflictCompare";
import { draftFromPackage } from "../lib/treatment-package-form";
import type { TreatmentPackage, TreatmentPackageManualDefinition } from "../api/treatment-packages";

const latest: TreatmentPackage = {
  id: "201",
  clinic_id: "7",
  version: "5",
  name: "最新名",
  description: "最新の説明",
  is_active: true,
  sort_order: 0,
  created_at: "t",
  updated_at: "t",
  components: [
    {
      component_key: "a0000000-0000-4000-8000-000000000011",
      item: { kind: "consultation", consultation_id: "101" },
      quantity: "1",
    },
  ],
};

const mine: TreatmentPackage = {
  ...latest,
  version: "3",
  name: "自分の入力",
  description: "最新の説明",
  is_active: false,
  components: [
    {
      component_key: "a0000000-0000-4000-8000-000000000011",
      item: { kind: "consultation", consultation_id: "102" },
      quantity: "2",
    },
  ],
};

describe("TreatmentPackageConflictCompare（差分比較・明示選択）", () => {
  it("あなたの入力 vs 最新を行単位で表示し差分を強調する", () => {
    render(
      <TreatmentPackageConflictCompare
        draft={draftFromPackage(mine)}
        latest={latest}
        onAdoptLatest={vi.fn()}
        onKeepEditing={vi.fn()}
      />,
    );
    expect(screen.getByRole("alert")).toBeTruthy();
    expect(screen.getByText("あなたの入力")).toBeTruthy();
    expect(screen.getByText("最新")).toBeTruthy();
    expect(screen.getAllByText("自分の入力").length).toBeGreaterThan(0);
    expect(screen.getAllByText("最新名").length).toBeGreaterThan(0);
    expect(screen.getByText(/構成/)).toBeTruthy();
    expect(screen.getAllByText(/×2/).length).toBeGreaterThan(0);
    expect(screen.getAllByText(/×1/).length).toBeGreaterThan(0);
  });

  it("2つの明示選択ボタンがそれぞれのハンドラを呼び自動送信しない", async () => {
    const user = userEvent.setup();
    const adopt = vi.fn();
    const keep = vi.fn();
    render(
      <TreatmentPackageConflictCompare
        draft={draftFromPackage(mine)}
        latest={latest}
        onAdoptLatest={adopt}
        onKeepEditing={keep}
      />,
    );
    await user.click(screen.getByText("最新を採用（入力を破棄）"));
    expect(adopt).toHaveBeenCalledTimes(1);
    await user.click(screen.getByText("入力を保持して最新版を基準に編集を続ける"));
    expect(keep).toHaveBeenCalledTimes(1);
  });

  const baseManual: TreatmentPackageManualDefinition = {
    name: "手入力項目",
    reason: "理由",
    unit_label: null,
    unit_price: "110",
    tax_type: "included",
    tax_rate: "0.1",
    is_non_insurance: false,
  };
  const manualPkg = (manual: TreatmentPackageManualDefinition): TreatmentPackage => ({
    ...latest,
    components: [
      {
        component_key: "a0000000-0000-4000-8000-000000000011",
        item: { kind: "manual_other" as const, manual },
        quantity: "1",
      },
    ],
  });

  it.each([
    ["reason", { reason: "理由B" }, "理由:理由", "理由:理由B"],
    ["unit_label", { unit_label: "1回" }, "単位:なし", "単位:1回"],
    ["unit_price", { unit_price: "120" }, "単価:110円", "単価:120円"],
    ["tax_type", { tax_type: "excluded" as const }, "税込", "税抜"],
    ["tax_rate", { tax_rate: "0.05" }, "税率0.1", "税率0.05"],
    ["is_non_insurance", { is_non_insurance: true }, "保険対象", "保険外"],
  ])(
    "A2-01: manual の %s だけが違う競合は左右で異なる表示になる",
    (_field, patch, yoursText, latestText) => {
      render(
        <TreatmentPackageConflictCompare
          draft={draftFromPackage(manualPkg(baseManual))}
          latest={manualPkg({ ...baseManual, ...patch })}
          onAdoptLatest={vi.fn()}
          onKeepEditing={vi.fn()}
        />,
      );
      const rows = screen.getAllByRole("listitem");
      expect(rows).toHaveLength(2);
      expect(rows[0]!.textContent).toContain(yoursText);
      expect(rows[1]!.textContent).toContain(latestText);
      expect(rows[0]!.textContent).not.toBe(rows[1]!.textContent);
    },
  );
});
