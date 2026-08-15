import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import { C } from "@/lib/design-tokens";
import { ActionButtons } from "./ReceptionDialogActionButtons";
import type { ReceptionAppointment } from "../api/types";

const appointment = { id: 1 } as ReceptionAppointment;

function renderActionButtons(currentStatus: string) {
  return render(
    <ActionButtons
      currentStatus={currentStatus}
      appointment={appointment}
      isTrimming={false}
      isHospitalization={false}
      isMedical={true}
      onConfirm={vi.fn()}
      onOpenOwnerDetail={vi.fn()}
      onCreateMedicalRecord={vi.fn()}
      onCreateTrimming={vi.fn()}
      onCreateAccounting={vi.fn()}
      onCreateHospitalization={vi.fn()}
    />,
  );
}

describe("ReceptionDialogActionButtons — DESIGN.md brand CTA", () => {
  it.each([
    ["受付予約", "受付済にする"],
    ["受付済", "カルテ作成"],
    ["診療中", "カルテ入力"],
    ["会計待ち", "会計へ進む"],
    ["会計済", "完了/リストから削除"],
  ])("%s の primary CTA は brand pill（accent 不使用）", (status, label) => {
    renderActionButtons(status);
    const btn = screen.getByRole("button", { name: new RegExp(label) });
    expect(btn.className).toContain(C.bgBrand);
    expect(btn.className).toContain("rounded-full");
  });

  it("破壊的アクション（取消）は danger のまま維持される", () => {
    render(
      <ActionButtons
        currentStatus="受付予約"
        appointment={appointment}
        isTrimming={false}
        isHospitalization={false}
        isMedical={true}
        onConfirm={vi.fn()}
        onCancel={vi.fn()}
        onOpenOwnerDetail={vi.fn()}
        onCreateMedicalRecord={vi.fn()}
        onCreateTrimming={vi.fn()}
        onCreateAccounting={vi.fn()}
        onCreateHospitalization={vi.fn()}
      />,
    );
    const cancelBtn = screen.getByRole("button", { name: /取消/ });
    expect(cancelBtn.className).toContain(C.danger);
    expect(cancelBtn.className).not.toContain(C.bgBrand);
  });
});
