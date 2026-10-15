import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { C } from "@/lib/design-tokens";
import type { HospitalizationTreatmentPlan } from "@/types";

import { HospitalizationTreatmentTable } from "./HospitalizationTreatmentTable";

const TREATMENT_PLAN = {
  id: "synthetic-plan-990018",
  treatmentContent: "合成監査輸液",
  memo: "参照専用",
  is_insurance: true,
  unitPrice: 3_210,
  quantity: 2,
  discount: 10,
  discountAmount: 642,
  subtotal: 5_778,
} satisfies HospitalizationTreatmentPlan;

describe("HospitalizationTreatmentTable", () => {
  it("labels monetary inputs neutrally and formats existing amounts without recalculating", () => {
    render(
      <HospitalizationTreatmentTable
        treatmentPlans={[TREATMENT_PLAN]}
        onAdd={vi.fn()}
        onUpdate={vi.fn()}
        readOnly
      />,
    );
    expect(screen.getByRole("columnheader", { name: "単価（明細入力額）" })).toBeInTheDocument();
    expect(screen.getByRole("columnheader", { name: "明細金額" })).toBeInTheDocument();
    for (const amount of ["¥3,210", "¥642", "¥5,778"])
      expect(screen.getByRole("cell", { name: amount })).toBeInTheDocument();
    expect(screen.getByRole("cell", { name: "2" })).toBeInTheDocument();
    expect(screen.getByRole("cell", { name: "10" })).toBeInTheDocument();
    expect(
      screen.getByText("税額・保険適用・飼主負担額は会計時に確定します。"),
    ).toBeInTheDocument();
  });
  it("edit routeの参照専用状態では治療内容を変更・追加・削除できない", () => {
    render(
      <HospitalizationTreatmentTable
        treatmentPlans={[TREATMENT_PLAN]}
        onAdd={vi.fn()}
        onUpdate={vi.fn()}
        onRemove={vi.fn()}
        readOnly
      />,
    );

    expect(screen.getByRole("button", { name: "追加" })).toBeDisabled();
    expect(screen.getByRole("textbox", { name: "治療内容 1" })).toBeDisabled();
    expect(screen.getByRole("textbox", { name: "治療メモ 1" })).toBeDisabled();
    expect(screen.queryByRole("button", { name: "削除" })).not.toBeInTheDocument();
  });

  it("参照専用状態でも治療プラン表をキーボードでフォーカスして横スクロールできる", async () => {
    const user = userEvent.setup();

    render(
      <HospitalizationTreatmentTable
        treatmentPlans={[TREATMENT_PLAN]}
        onAdd={vi.fn()}
        onUpdate={vi.fn()}
        onRemove={vi.fn()}
        readOnly
      />,
    );

    const scrollRegion = screen.getByRole("region", { name: "治療プラン" });
    expect(scrollRegion).toContainElement(screen.getByRole("table"));
    expect(scrollRegion).toHaveAttribute("tabindex", "0");
    expect(scrollRegion).toHaveClass(
      "overflow-x-auto",
      "outline-none",
      "focus-visible:ring-2",
      "focus-visible:ring-inset",
      C.focusRingAccent40,
    );

    await user.tab();

    expect(scrollRegion).toHaveFocus();
  });
});
