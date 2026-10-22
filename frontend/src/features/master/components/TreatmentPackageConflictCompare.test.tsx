import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { TreatmentPackageConflictCompare } from "./TreatmentPackageConflictCompare";
import { draftFromPackage } from "../lib/treatment-package-form";
import type { TreatmentPackage } from "../api/treatment-packages";

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
});
