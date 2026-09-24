import { describe, it, expect, vi } from "vitest";
import { fireEvent, render, screen } from "@testing-library/react";
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

describe("ReceptionDialogActionButtons — EMR-74 受付済 trimming", () => {
  it("トリミング予約は onConfirm（in_consultation 遷移）を呼ばず onCreateTrimming のみ呼ぶ", () => {
    const onConfirm = vi.fn();
    const onCreateTrimming = vi.fn();
    render(
      <ActionButtons
        currentStatus="受付済"
        appointment={appointment}
        isTrimming={true}
        isHospitalization={false}
        isMedical={false}
        onConfirm={onConfirm}
        onOpenOwnerDetail={vi.fn()}
        onCreateMedicalRecord={vi.fn()}
        onCreateTrimming={onCreateTrimming}
        onCreateAccounting={vi.fn()}
        onCreateHospitalization={vi.fn()}
      />,
    );

    fireEvent.click(screen.getByRole("button", { name: /トリミングカルテ作成/ }));

    expect(onConfirm).not.toHaveBeenCalled();
    expect(onCreateTrimming).toHaveBeenCalledTimes(1);
  });

  it("トリミング予約の受付済 CTA は「診療中」への遷移を案内しない", () => {
    render(
      <ActionButtons
        currentStatus="受付済"
        appointment={appointment}
        isTrimming={true}
        isHospitalization={false}
        isMedical={false}
        onConfirm={vi.fn()}
        onOpenOwnerDetail={vi.fn()}
        onCreateMedicalRecord={vi.fn()}
        onCreateTrimming={vi.fn()}
        onCreateAccounting={vi.fn()}
        onCreateHospitalization={vi.fn()}
      />,
    );

    expect(screen.queryByText(/診療中/)).not.toBeInTheDocument();
  });

  it("一般予約の受付済カルテ作成は引き続き onConfirm と onCreateMedicalRecord を呼ぶ", () => {
    const onConfirm = vi.fn();
    const onCreateMedicalRecord = vi.fn();
    render(
      <ActionButtons
        currentStatus="受付済"
        appointment={appointment}
        isTrimming={false}
        isHospitalization={false}
        isMedical={true}
        onConfirm={onConfirm}
        onOpenOwnerDetail={vi.fn()}
        onCreateMedicalRecord={onCreateMedicalRecord}
        onCreateTrimming={vi.fn()}
        onCreateAccounting={vi.fn()}
        onCreateHospitalization={vi.fn()}
      />,
    );

    fireEvent.click(screen.getByRole("button", { name: /カルテ作成/ }));

    expect(onConfirm).toHaveBeenCalledTimes(1);
    expect(onCreateMedicalRecord).toHaveBeenCalledTimes(1);
  });
});
