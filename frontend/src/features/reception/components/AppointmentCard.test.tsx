import { DndContext } from "@dnd-kit/core";
import { fireEvent, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";

import { AppointmentCard } from "./AppointmentCard";
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
  reservationTypeId: "1",
  reservationCategory: "general",
  isDesignated: false,
  doctor: "担当者A",
  doctorId: "33",
  petId: "10",
  ownerId: "20",
  status: "checked_in",
  notes: undefined,
  source: "manual",
};

function renderCard(appointment: ReceptionAppointment = baseAppointment, columnTitle = "受付済") {
  return render(
    <DndContext>
      <AppointmentCard
        appointment={appointment}
        columnTitle={columnTitle}
        onCardClick={vi.fn()}
      />
    </DndContext>,
  );
}

describe("AppointmentCard", () => {
  it("通常カルテ遷移に appointmentId を query と state の両方で渡す", () => {
    renderCard();

    fireEvent.click(screen.getByRole("button", { name: /ポチのカルテ/ }));

    expect(navigateMock).toHaveBeenCalledWith(
      "/medical-records/new?petId=10&appointmentId=101&visitDate=2026-05-29",
      { state: { from: "/", appointmentId: "101", visitDate: "2026-05-29" } },
    );
  });

  it("petId 未確定の通常予約は appointmentId を保持してペット選択へ遷移する", () => {
    renderCard({
      ...baseAppointment,
      petId: "",
    });

    fireEvent.click(screen.getByRole("button", { name: /ポチのカルテ/ }));

    expect(navigateMock).toHaveBeenCalledWith(
      "/medical-records/select-pet?appointmentId=101&visitDate=2026-05-29",
      { state: { from: "/", appointmentId: "101", visitDate: "2026-05-29" } },
    );
  });

  it("通常予約のカルテボタンは受付予約カラムでは表示しない", () => {
    renderCard(baseAppointment, "受付予約");

    expect(screen.queryByRole("button", { name: /ポチのカルテ/ })).not.toBeInTheDocument();
  });

  it("通常予約のカルテボタンは診療中カラムでは表示する", () => {
    renderCard(baseAppointment, "診療中");

    expect(screen.getByRole("button", { name: /ポチのカルテ/ })).toBeInTheDocument();
  });

  it("トリミング予約はカテゴリで判定し、施術遷移に appointmentId を渡す", () => {
    renderCard({
      ...baseAppointment,
      id: "202",
      reservationType: "シャンプーコース",
      reservationCategory: "trimming",
    });

    fireEvent.click(screen.getByRole("button", { name: /ポチのトリミング記録/ }));

    expect(navigateMock).toHaveBeenCalledWith(
      "/trimming/new?petId=10&appointmentId=202&visitDate=2026-05-29",
      { state: { from: "/", appointmentId: "202", visitDate: "2026-05-29" } },
    );
  });

  it("petId 未確定のトリミング予約は appointmentId を保持してペット選択へ遷移する", () => {
    renderCard({
      ...baseAppointment,
      id: "202",
      petId: "",
      reservationType: "シャンプーコース",
      reservationCategory: "trimming",
    });

    fireEvent.click(screen.getByRole("button", { name: /ポチのトリミング記録/ }));

    expect(navigateMock).toHaveBeenCalledWith(
      "/trimming/select-pet?appointmentId=202&visitDate=2026-05-29",
      { state: { from: "/", appointmentId: "202", visitDate: "2026-05-29" } },
    );
  });

  it("トリミング予約の施術ボタンは受付済カラム以外では表示しない", () => {
    renderCard({
      ...baseAppointment,
      id: "202",
      reservationType: "シャンプーコース",
      reservationCategory: "trimming",
    }, "診療中");

    expect(screen.queryByRole("button", { name: /ポチのトリミング記録/ })).not.toBeInTheDocument();
  });
});
