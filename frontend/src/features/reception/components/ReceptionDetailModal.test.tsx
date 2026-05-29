import { fireEvent, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";

import { ReceptionDetailModal } from "./ReceptionDetailModal";
import type { ReceptionAppointment } from "../api/types";

const { navigateMock } = vi.hoisted(() => ({
  navigateMock: vi.fn(),
}));

vi.mock("react-router", () => ({
  useNavigate: () => navigateMock,
}));

afterEach(() => {
  navigateMock.mockClear();
});

const baseAppointment: ReceptionAppointment = {
  id: "101",
  time: "09:45",
  visitDate: "2026-05-29",
  ownerName: "山田",
  petType: "犬",
  petName: "ポチ",
  visitType: "再診",
  reservationType: "一般診察",
  reservationCategory: "general",
  isDesignated: false,
  doctor: "担当者A",
  petId: "",
  ownerId: "20",
  status: "checked_in",
  notes: undefined,
  source: "manual",
};

function renderModal(appointment: ReceptionAppointment = baseAppointment) {
  return render(
    <ReceptionDetailModal
      isOpen={true}
      onClose={vi.fn()}
      appointment={appointment}
      currentStatus="受付済"
      canCreateMedicalRecord={true}
      canCreateAccounting={true}
      canCreateHospitalization={true}
    />,
  );
}

describe("ReceptionDetailModal", () => {
  it("petId 未確定の通常予約では appointmentId を保持してペット選択へ遷移する", () => {
    renderModal();

    fireEvent.click(screen.getByRole("button", { name: /カルテ作成/ }));

    expect(navigateMock).toHaveBeenCalledWith(
      "/medical-records/select-pet?appointmentId=101&visitDate=2026-05-29",
      { state: { from: "/", appointmentId: "101", visitDate: "2026-05-29" } },
    );
  });

  it("petId 未確定のトリミング予約では appointmentId を保持してペット選択へ遷移する", () => {
    renderModal({
      ...baseAppointment,
      id: "202",
      reservationType: "シャンプーコース",
      reservationCategory: "trimming",
    });

    fireEvent.click(screen.getByRole("button", { name: /トリミングカルテ作成/ }));

    expect(navigateMock).toHaveBeenCalledWith(
      "/trimming/select-pet?appointmentId=202&visitDate=2026-05-29",
      { state: { from: "/", appointmentId: "202", visitDate: "2026-05-29" } },
    );
  });
});
