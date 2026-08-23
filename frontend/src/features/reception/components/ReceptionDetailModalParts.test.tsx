import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { ReceptionDialogFooter } from "./ReceptionDetailModalParts";
import type { ReceptionAppointment } from "../api/types";

const trimmingAppointment: ReceptionAppointment = {
  id: "1",
  time: "09:45",
  visitDate: "2026-05-29",
  end: new Date(2026, 4, 29, 10, 15, 0),
  ownerName: "山田",
  petType: "犬",
  petName: "ポチ",
  visitType: "再診",
  reservationType: "トリミング",
  reservationTypeId: "9",
  reservationCategory: "trimming",
  isDesignated: false,
  doctor: "担当者A",
  doctorId: "33",
  petId: "10",
  ownerId: "20",
  status: "checked_in",
  notes: undefined,
  source: "manual",
};

describe("ReceptionDialogFooter", () => {
  it("受付済のトリミング予約ではトリミングカルテ作成から遷移する", async () => {
    const onConfirm = vi.fn();
    const onCreateTrimming = vi.fn();

    render(
      <ReceptionDialogFooter
        currentStatus="受付済"
        appointment={trimmingAppointment}
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

    await userEvent.click(screen.getByRole("button", { name: /トリミングカルテ作成/ }));

    expect(onConfirm).toHaveBeenCalledOnce();
    expect(onCreateTrimming).toHaveBeenCalledOnce();
  });

  it("診療中のトリミング予約では施術記録ボタンを表示しない", () => {
    render(
      <ReceptionDialogFooter
        currentStatus="診療中"
        appointment={{ ...trimmingAppointment, status: "in_consultation" }}
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

    expect(screen.queryByRole("button", { name: /施術記録/ })).not.toBeInTheDocument();
  });
});
