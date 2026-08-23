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
  end: new Date(2026, 4, 29, 10, 15, 0),
  ownerName: "山田",
  petType: "犬",
  petName: "ポチ",
  visitType: "再診",
  reservationType: "一般診察",
  reservationTypeId: "1",
  reservationCategory: "general",
  isDesignated: false,
  doctor: "担当者A",
  doctorId: "33",
  petId: "",
  ownerId: "20",
  status: "checked_in",
  notes: undefined,
  source: "manual",
};

interface RenderModalOptions {
  appointment?: ReceptionAppointment;
  currentStatus?: string;
}

function renderModal({
  appointment = baseAppointment,
  currentStatus = "受付済",
}: RenderModalOptions = {}) {
  return render(
    <ReceptionDetailModal
      isOpen={true}
      onClose={vi.fn()}
      appointment={appointment}
      currentStatus={currentStatus}
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
      appointment: {
        ...baseAppointment,
        id: "202",
        reservationType: "シャンプーコース",
        reservationCategory: "trimming",
      },
    });

    fireEvent.click(screen.getByRole("button", { name: /トリミングカルテ作成/ }));

    expect(navigateMock).toHaveBeenCalledWith(
      "/trimming/select-pet?appointmentId=202&visitDate=2026-05-29",
      { state: { from: "/", appointmentId: "202", visitDate: "2026-05-29" } },
    );
  });

  it("受付済の通常予約では関連ページのカルテ導線を表示しない", () => {
    renderModal();

    expect(screen.queryByRole("button", { name: /^カルテ$/ })).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: /カルテ作成/ })).toBeInTheDocument();
  });

  it("診療中の通常予約では関連ページのカルテ導線から appointmentId と visitDate を保持して遷移する", () => {
    renderModal({
      currentStatus: "診療中",
      appointment: {
        ...baseAppointment,
        petId: "10",
      },
    });

    fireEvent.click(screen.getByRole("button", { name: /^カルテ$/ }));

    expect(navigateMock).toHaveBeenCalledWith(
      "/medical-records/new?petId=10&appointmentId=101&visitDate=2026-05-29",
      { state: { from: "/", appointmentId: "101", visitDate: "2026-05-29" } },
    );
  });

  it("診療中のホテル予約では関連ページのカルテ導線を表示しない", () => {
    renderModal({
      currentStatus: "診療中",
      appointment: {
        ...baseAppointment,
        id: "404",
        reservationType: "ホテル",
        reservationCategory: "general",
        petId: "10",
      },
    });

    expect(screen.queryByRole("button", { name: /^カルテ$/ })).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: /^入院$/ })).toBeInTheDocument();
  });
});
